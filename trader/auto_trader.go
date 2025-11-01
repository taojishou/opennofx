package trader

import (
	"encoding/json"
	"fmt"
	"log"
	"nofx/database"
	"nofx/decision"
	"nofx/logger"
	"nofx/market"
	"nofx/mcp"
	"nofx/pool"
	"strings"
	"sync"
	"time"
)

// AutoTraderConfig 自动交易配置（简化版 - AI全权决策）
type AutoTraderConfig struct {
	// Trader标识
	ID      string // Trader唯一标识（用于日志目录等）
	Name    string // Trader显示名称
	AIModel string // AI模型: "qwen" 或 "deepseek"

	// 交易平台选择
	Exchange string // "binance", "hyperliquid" 或 "aster"

	// 币安API配置
	BinanceAPIKey    string
	BinanceSecretKey string

	// Hyperliquid配置
	HyperliquidPrivateKey string
	HyperliquidWalletAddr string
	HyperliquidTestnet    bool

	// Aster配置
	AsterUser       string // Aster主钱包地址
	AsterSigner     string // Aster API钱包地址
	AsterPrivateKey string // Aster API钱包私钥

	CoinPoolAPIURL string

	// AI配置
	UseQwen     bool
	DeepSeekKey string
	QwenKey     string

	// 自定义AI API配置
	CustomAPIURL    string
	CustomAPIKey    string
	CustomModelName string

	// 扫描配置
	ScanInterval time.Duration // 扫描间隔（建议3分钟）

	// 账户配置
	InitialBalance float64 // 初始金额（用于计算盈亏，需手动设置）

	// 杠杆配置
	BTCETHLeverage  int // BTC和ETH的杠杆倍数
	AltcoinLeverage int // 山寨币的杠杆倍数

	// 持仓控制
	MaxPositions int // 最大持仓数限制（默认3）

	// AI学习配置
	EnableAILearning bool // 是否启用AI自动学习总结
	AILearnInterval  int  // AI学习触发间隔（几个周期一次）

	// 风险控制（仅作为提示，AI可自主决定）
	MaxDailyLoss    float64       // 最大日亏损百分比（提示）
	MaxDrawdown     float64       // 最大回撤百分比（提示）
	StopTradingTime time.Duration // 触发风控后暂停时长
}

// AutoTrader 自动交易器
type AutoTrader struct {
	id                    string // Trader唯一标识
	name                  string // Trader显示名称
	aiModel               string // AI模型名称
	exchange              string // 交易平台名称
	config                AutoTraderConfig
	trader                Trader // 使用Trader接口（支持多平台）
	mcpClient             *mcp.Client
	decisionLogger        *logger.DecisionLogger // 决策日志记录器
	initialBalance        float64
	dailyPnL              float64
	lastResetTime         time.Time
	stopUntil             time.Time
	isRunning             bool
	isPaused              bool                   // 是否暂停
	startTime             time.Time              // 系统启动时间
	callCount             int                    // AI调用次数
	positionFirstSeenTime map[string]int64       // 持仓首次出现时间 (symbol_side -> timestamp毫秒)
	lastKnownPositions    map[string]bool        // 上次已知的持仓 (symbol_side -> true)，用于检测自动平仓
	enableAILearning      bool                   // 是否启用AI学习
	aiLearnInterval       int                    // AI学习间隔（周期数）
	mu                    sync.RWMutex           // 保护并发访问
}

// NewAutoTrader 创建自动交易器
func NewAutoTrader(config AutoTraderConfig) (*AutoTrader, error) {
	// 设置默认值
	if config.ID == "" {
		config.ID = "default_trader"
	}
	if config.Name == "" {
		config.Name = "Default Trader"
	}
	if config.AIModel == "" {
		if config.UseQwen {
			config.AIModel = "qwen"
		} else {
			config.AIModel = "deepseek"
		}
	}

	mcpClient := mcp.New()

	// 初始化AI
	if config.AIModel == "custom" {
		// 使用自定义API
		mcpClient.SetCustomAPI(config.CustomAPIURL, config.CustomAPIKey, config.CustomModelName)
		log.Printf("🤖 [%s] 使用自定义AI API: %s (模型: %s)", config.Name, config.CustomAPIURL, config.CustomModelName)
	} else if config.UseQwen || config.AIModel == "qwen" {
		// 使用Qwen
		mcpClient.SetQwenAPIKey(config.QwenKey, "")
		log.Printf("🤖 [%s] 使用阿里云Qwen AI", config.Name)
	} else {
		// 默认使用DeepSeek
		mcpClient.SetDeepSeekAPIKey(config.DeepSeekKey)
		log.Printf("🤖 [%s] 使用DeepSeek AI", config.Name)
	}

	// 初始化币种池API
	if config.CoinPoolAPIURL != "" {
		pool.SetCoinPoolAPI(config.CoinPoolAPIURL)
	}

	// 设置默认交易平台
	if config.Exchange == "" {
		config.Exchange = "binance"
	}

	// 根据配置创建对应的交易器
	var trader Trader
	var err error

	switch config.Exchange {
	case "binance":
		log.Printf("🏦 [%s] 使用币安合约交易", config.Name)
		trader = NewFuturesTrader(config.BinanceAPIKey, config.BinanceSecretKey)
	case "hyperliquid":
		log.Printf("🏦 [%s] 使用Hyperliquid交易", config.Name)
		trader, err = NewHyperliquidTrader(config.HyperliquidPrivateKey, config.HyperliquidWalletAddr, config.HyperliquidTestnet)
		if err != nil {
			return nil, fmt.Errorf("初始化Hyperliquid交易器失败: %w", err)
		}
	case "aster":
		log.Printf("🏦 [%s] 使用Aster交易", config.Name)
		trader, err = NewAsterTrader(config.AsterUser, config.AsterSigner, config.AsterPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("初始化Aster交易器失败: %w", err)
		}
	default:
		return nil, fmt.Errorf("不支持的交易平台: %s", config.Exchange)
	}

	// 验证初始金额配置
	if config.InitialBalance <= 0 {
		return nil, fmt.Errorf("初始金额必须大于0，请在配置中设置InitialBalance")
	}

	// 初始化决策日志记录器（使用trader ID创建独立目录）
	logDir := fmt.Sprintf("decision_logs/%s", config.ID)
	decisionLogger := logger.NewDecisionLogger(logDir)

	// 设置默认最大持仓数
	if config.MaxPositions <= 0 {
		config.MaxPositions = 3
	}

	at := &AutoTrader{
		id:                    config.ID,
		name:                  config.Name,
		aiModel:               config.AIModel,
		exchange:              config.Exchange,
		config:                config,
		trader:                trader,
		mcpClient:             mcpClient,
		decisionLogger:        decisionLogger,
		initialBalance:        config.InitialBalance,
		lastResetTime:         time.Now(),
		startTime:             time.Now(),
		callCount:             0,
		isRunning:             false,
		positionFirstSeenTime: make(map[string]int64),
		lastKnownPositions:    make(map[string]bool),
		enableAILearning:      config.EnableAILearning,
		aiLearnInterval:       config.AILearnInterval,
	}

	// 从数据库恢复持仓开仓时间和运行状态
	if db := decisionLogger.GetDB(); db != nil {
		// 恢复持仓开仓时间
		if savedTimes, err := db.GetAllPositionOpenTimes(); err == nil && len(savedTimes) > 0 {
			at.positionFirstSeenTime = savedTimes
			log.Printf("✓ 从数据库恢复了 %d 个持仓的开仓时间", len(savedTimes))
		}
		
		// 恢复运行状态
		if isPaused, exists := db.GetTraderState(); exists {
			at.isPaused = isPaused
			if isPaused {
				log.Printf("✓ 从数据库恢复状态: 暂停中")
			} else {
				log.Printf("✓ 从数据库恢复状态: 运行中")
			}
		} else {
			// 没有保存的状态，默认为运行（不暂停）
			log.Printf("✓ 首次启动，默认状态: 运行中")
		}
	}

	return at, nil
}

// Run 运行自动交易主循环
func (at *AutoTrader) Run() error {
	at.isRunning = true
	log.Println("🚀 AI驱动自动交易系统启动")
	log.Printf("💰 初始余额: %.2f USDT", at.initialBalance)
	log.Printf("⚙️  扫描间隔: %v", at.config.ScanInterval)
	log.Println("🤖 AI将全权决定杠杆、仓位大小、止损止盈等参数")

	ticker := time.NewTicker(at.config.ScanInterval)
	defer ticker.Stop()

	// 首次立即执行（检查暂停状态）
	if !at.IsPaused() {
		if err := at.runCycle(); err != nil {
			log.Printf("❌ 执行失败: %v", err)
		}
	} else {
		log.Printf("[%s] ⏸️  Trader已暂停，跳过首次执行", at.name)
	}

	for at.isRunning {
		select {
		case <-ticker.C:
			// 检查是否暂停
			if at.IsPaused() {
				log.Printf("[%s] ⏸️  Trader已暂停，跳过本次交易循环", at.name)
				continue
			}
			
			if err := at.runCycle(); err != nil {
				log.Printf("❌ 执行失败: %v", err)
			}
		}
	}

	return nil
}

// Stop 停止自动交易
func (at *AutoTrader) Stop() {
	at.isRunning = false
	log.Println("⏹ 自动交易系统停止")
}

// runCycle 运行一个交易周期（使用AI全权决策）
func (at *AutoTrader) runCycle() error {
	// ⚠️ 关键检查：如果暂停，完全不执行任何操作
	// 不收集数据、不调用AI、不记录日志、不增加callCount
	if at.IsPaused() {
		return nil
	}
	
	at.callCount++

	log.Printf("\n" + strings.Repeat("=", 70))
	log.Printf("[%s] ⏰ %s - AI决策周期 #%d", at.name, time.Now().Format("2006-01-02 15:04:05"), at.callCount)
	log.Printf(strings.Repeat("=", 70))

	// 创建决策记录
	record := &logger.DecisionRecord{
		ExecutionLog: []string{},
		Success:      true,
	}

	// 1. 检查是否需要停止交易（风险控制暂停）
	if time.Now().Before(at.stopUntil) {
		remaining := at.stopUntil.Sub(time.Now())
		log.Printf("⏸ 风险控制：暂停交易中，剩余 %.0f 分钟", remaining.Minutes())
		record.Success = false
		record.ErrorMessage = fmt.Sprintf("风险控制暂停中，剩余 %.0f 分钟", remaining.Minutes())
		at.decisionLogger.LogDecision(record)
		return nil
	}

	// 2. 重置日盈亏（每天重置）
	if time.Since(at.lastResetTime) > 24*time.Hour {
		at.dailyPnL = 0
		at.lastResetTime = time.Now()
		log.Println("📅 日盈亏已重置")
	}

	// 3. 收集交易上下文（同时检测自动平仓）
	ctx, autoClosedPositions, err := at.buildTradingContext()
	if err != nil {
		record.Success = false
		record.ErrorMessage = fmt.Sprintf("构建交易上下文失败: %v", err)
		at.decisionLogger.LogDecision(record)
		return fmt.Errorf("构建交易上下文失败: %w", err)
	}
	
	// 记录自动平仓事件（如果有）
	for _, autoCloseAction := range autoClosedPositions {
		record.Decisions = append(record.Decisions, autoCloseAction)
		record.ExecutionLog = append(record.ExecutionLog, 
			fmt.Sprintf("🤖 %s %s 自动平仓（止损/止盈触发）", autoCloseAction.Symbol, autoCloseAction.Action))
	}

	// 保存账户状态快照
	record.AccountState = logger.AccountSnapshot{
		TotalBalance:          ctx.Account.TotalEquity,
		AvailableBalance:      ctx.Account.AvailableBalance,
		TotalUnrealizedProfit: ctx.Account.TotalPnL,
		PositionCount:         ctx.Account.PositionCount,
		MarginUsedPct:         ctx.Account.MarginUsedPct,
	}

	// 保存持仓快照
	for _, pos := range ctx.Positions {
		record.Positions = append(record.Positions, logger.PositionSnapshot{
			Symbol:           pos.Symbol,
			Side:             pos.Side,
			PositionAmt:      pos.Quantity,
			EntryPrice:       pos.EntryPrice,
			MarkPrice:        pos.MarkPrice,
			UnrealizedProfit: pos.UnrealizedPnL,
			Leverage:         float64(pos.Leverage),
			LiquidationPrice: pos.LiquidationPrice,
		})
	}

	// 保存候选币种列表
	for _, coin := range ctx.CandidateCoins {
		record.CandidateCoins = append(record.CandidateCoins, coin.Symbol)
	}

	log.Printf("📊 账户净值: %.2f USDT | 可用: %.2f USDT | 持仓: %d",
		ctx.Account.TotalEquity, ctx.Account.AvailableBalance, ctx.Account.PositionCount)

	// 4. 调用AI获取完整决策
	log.Println("🤖 正在请求AI分析并决策...")
	decision, err := decision.GetFullDecision(ctx, at.mcpClient)

	// 即使有错误，也保存思维链、决策和输入prompt（用于debug）
	if decision != nil {
		record.SystemPrompt = decision.SystemPrompt
		record.InputPrompt = decision.UserPrompt
		record.CoTTrace = decision.CoTTrace
		if len(decision.Decisions) > 0 {
			decisionJSON, _ := json.MarshalIndent(decision.Decisions, "", "  ")
			record.DecisionJSON = string(decisionJSON)
		}
	}

	if err != nil {
		record.Success = false
		record.ErrorMessage = fmt.Sprintf("获取AI决策失败: %v", err)

		// 打印AI思维链（即使有错误）
		if decision != nil && decision.CoTTrace != "" {
			log.Printf("\n" + strings.Repeat("-", 70))
			log.Println("💭 AI思维链分析（错误情况）:")
			log.Println(strings.Repeat("-", 70))
			log.Println(decision.CoTTrace)
			log.Printf(strings.Repeat("-", 70) + "\n")
		}

		at.decisionLogger.LogDecision(record)
		return fmt.Errorf("获取AI决策失败: %w", err)
	}

	// 5. 打印AI思维链
	log.Printf("\n" + strings.Repeat("-", 70))
	log.Println("💭 AI思维链分析:")
	log.Println(strings.Repeat("-", 70))
	log.Println(decision.CoTTrace)
	log.Printf(strings.Repeat("-", 70) + "\n")

	// 6. 打印AI决策
	log.Printf("📋 AI决策列表 (%d 个):\n", len(decision.Decisions))
	for i, d := range decision.Decisions {
		log.Printf("  [%d] %s: %s - %s", i+1, d.Symbol, d.Action, d.Reasoning)
		if d.Action == "open_long" || d.Action == "open_short" {
			log.Printf("      杠杆: %dx | 仓位: %.2f USDT | 止损: %.4f | 止盈: %.4f",
				d.Leverage, d.PositionSizeUSD, d.StopLoss, d.TakeProfit)
		}
	}
	log.Println()

	// 7. 对决策排序：确保先平仓后开仓（防止仓位叠加超限）
	sortedDecisions := sortDecisionsByPriority(decision.Decisions)

	log.Println("🔄 执行顺序（已优化）: 先平仓→后开仓")
	for i, d := range sortedDecisions {
		log.Printf("  [%d] %s %s", i+1, d.Symbol, d.Action)
	}
	log.Println()

	// 执行决策并记录结果
	for _, d := range sortedDecisions {

		actionRecord := logger.DecisionAction{
			Action:    d.Action,
			Symbol:    d.Symbol,
			Quantity:  0,
			Leverage:  d.Leverage,
			Price:     0,
			Timestamp: time.Now(),
			Success:   false,
		}

		if err := at.executeDecisionWithRecord(&d, &actionRecord); err != nil {
			log.Printf("❌ 执行决策失败 (%s %s): %v", d.Symbol, d.Action, err)
			actionRecord.Error = err.Error()
			record.ExecutionLog = append(record.ExecutionLog, fmt.Sprintf("❌ %s %s 失败: %v", d.Symbol, d.Action, err))
		} else {
			actionRecord.Success = true
			record.ExecutionLog = append(record.ExecutionLog, fmt.Sprintf("✓ %s %s 成功", d.Symbol, d.Action))
			// 成功执行后短暂延迟
			time.Sleep(1 * time.Second)
		}

		record.Decisions = append(record.Decisions, actionRecord)
	}

	// 8. 保存决策记录
	if err := at.decisionLogger.LogDecision(record); err != nil {
		log.Printf("⚠ 保存决策记录失败: %v", err)
	}

	// 9. 自动生成AI学习总结（根据配置间隔）
	if at.enableAILearning && at.aiLearnInterval > 0 && at.callCount%at.aiLearnInterval == 0 {
		go at.maybeGenerateAILearningSummary()
	}

	return nil
}

// buildTradingContext 构建交易上下文（同时检测自动平仓）
func (at *AutoTrader) buildTradingContext() (*decision.Context, []logger.DecisionAction, error) {
	// 1. 获取账户信息
	balance, err := at.trader.GetBalance()
	if err != nil {
		return nil, nil, fmt.Errorf("获取账户余额失败: %w", err)
	}

	// 获取账户字段
	totalWalletBalance := 0.0
	totalUnrealizedProfit := 0.0
	availableBalance := 0.0

	if wallet, ok := balance["totalWalletBalance"].(float64); ok {
		totalWalletBalance = wallet
	}
	if unrealized, ok := balance["totalUnrealizedProfit"].(float64); ok {
		totalUnrealizedProfit = unrealized
	}
	if avail, ok := balance["availableBalance"].(float64); ok {
		availableBalance = avail
	}

	// Total Equity = 钱包余额 + 未实现盈亏
	totalEquity := totalWalletBalance + totalUnrealizedProfit

	// 2. 获取持仓信息并检测自动平仓
	positions, err := at.trader.GetPositions()
	if err != nil {
		return nil, nil, fmt.Errorf("获取持仓失败: %w", err)
	}

	var positionInfos []decision.PositionInfo
	totalMarginUsed := 0.0

	// 当前持仓的key集合（用于清理已平仓的记录）
	currentPositionKeys := make(map[string]bool)
	
	// 检测自动平仓事件（持仓消失但不是AI主动平仓）
	// 这些自动平仓事件会被记录到决策日志中
	var autoClosedPositions []logger.DecisionAction

	for _, pos := range positions {
		symbol := pos["symbol"].(string)
		side := pos["side"].(string)
		entryPrice := pos["entryPrice"].(float64)
		markPrice := pos["markPrice"].(float64)
		quantity := pos["positionAmt"].(float64)
		if quantity < 0 {
			quantity = -quantity // 空仓数量为负，转为正数
		}
		unrealizedPnl := pos["unRealizedProfit"].(float64)
		liquidationPrice := pos["liquidationPrice"].(float64)

		// 计算占用保证金（估算）
		leverage := 10 // 默认值，实际应该从持仓信息获取
		if lev, ok := pos["leverage"].(float64); ok {
			leverage = int(lev)
		}
		marginUsed := (quantity * markPrice) / float64(leverage)
		totalMarginUsed += marginUsed

		// 计算盈亏百分比
		pnlPct := 0.0
		if side == "long" {
			pnlPct = ((markPrice - entryPrice) / entryPrice) * float64(leverage) * 100
		} else {
			pnlPct = ((entryPrice - markPrice) / entryPrice) * float64(leverage) * 100
		}

		// 跟踪持仓首次出现时间
		posKey := symbol + "_" + side
		currentPositionKeys[posKey] = true
		if _, exists := at.positionFirstSeenTime[posKey]; !exists {
			// 新持仓，先尝试从数据库恢复
			if db := at.decisionLogger.GetDB(); db != nil {
				if savedTime, ok := db.GetPositionOpenTime(symbol, side); ok {
					at.positionFirstSeenTime[posKey] = savedTime
					log.Printf("  📅 从数据库恢复 %s %s 的开仓时间", symbol, side)
				} else {
					// 数据库中没有，记录当前时间（可能是系统重启前的持仓）
					at.positionFirstSeenTime[posKey] = time.Now().UnixMilli()
				}
			} else {
				// 没有数据库，使用当前时间
				at.positionFirstSeenTime[posKey] = time.Now().UnixMilli()
			}
		}
		updateTime := at.positionFirstSeenTime[posKey]

		positionInfos = append(positionInfos, decision.PositionInfo{
			Symbol:           symbol,
			Side:             side,
			EntryPrice:       entryPrice,
			MarkPrice:        markPrice,
			Quantity:         quantity,
			Leverage:         leverage,
			UnrealizedPnL:    unrealizedPnl,
			UnrealizedPnLPct: pnlPct,
			LiquidationPrice: liquidationPrice,
			MarginUsed:       marginUsed,
			UpdateTime:       updateTime,
		})
	}

	// 检测自动平仓（上次存在但这次不存在的持仓）
	for key := range at.lastKnownPositions {
		if !currentPositionKeys[key] {
			// 这个持仓消失了，可能是止损或止盈触发
			// 解析 key (symbol_side)
			parts := strings.Split(key, "_")
			if len(parts) == 2 {
				symbol := parts[0]
				side := parts[1]
				
				// 记录自动平仓事件
				action := "close_long"
				if side == "short" {
					action = "close_short"
				}
				
				// 获取当前价格作为平仓价
				marketData, _ := market.Get(symbol)
				closePrice := 0.0
				if marketData != nil {
					closePrice = marketData.CurrentPrice
				}
				
				autoClosedPositions = append(autoClosedPositions, logger.DecisionAction{
					Action:      action,
					Symbol:      symbol,
					Quantity:    0, // 无法获取数量
					Price:       closePrice,
					Timestamp:   time.Now(),
					Success:     true,
					WasStopLoss: true, // 标记为可能的止损/止盈
				})
				
				log.Printf("  📍 检测到自动平仓: %s %s (可能触发止损/止盈)", symbol, strings.ToUpper(side))
				
				// 保存交易记录到trade_outcomes表
				at.saveAutoClosedTradeOutcome(symbol, side, closePrice)
				
				// 从数据库删除（在 if 块内部，symbol 和 side 变量可用）
				if db := at.decisionLogger.GetDB(); db != nil {
					if err := db.DeletePositionOpenTime(symbol, side); err != nil {
						log.Printf("  ⚠️  从数据库删除开仓时间失败: %v", err)
					}
				}
			}
			
			// 清理内存记录
			delete(at.positionFirstSeenTime, key)
		}
	}
	
	// 更新已知持仓列表
	at.lastKnownPositions = currentPositionKeys

	// 3. 获取合并的候选币种池（AI500 + OI Top，去重）
	// 无论有没有持仓，都分析相同数量的币种（让AI看到所有好机会）
	// AI会根据保证金使用率和现有持仓情况，自己决定是否要换仓
	const ai500Limit = 20 // AI500取前20个评分最高的币种

	// 获取合并后的币种池（AI500 + OI Top）
	mergedPool, err := pool.GetMergedCoinPool(ai500Limit)
	if err != nil {
		return nil, nil, fmt.Errorf("获取合并币种池失败: %w", err)
	}

	// 构建候选币种列表（包含来源信息）
	var candidateCoins []decision.CandidateCoin
	for _, symbol := range mergedPool.AllSymbols {
		sources := mergedPool.SymbolSources[symbol]
		candidateCoins = append(candidateCoins, decision.CandidateCoin{
			Symbol:  symbol,
			Sources: sources, // "ai500" 和/或 "oi_top"
		})
	}

	log.Printf("📋 合并币种池: AI500前%d + OI_Top20 = 总计%d个候选币种",
		ai500Limit, len(candidateCoins))

	// 4. 计算总盈亏
	totalPnL := totalEquity - at.initialBalance
	totalPnLPct := 0.0
	if at.initialBalance > 0 {
		totalPnLPct = (totalPnL / at.initialBalance) * 100
	}

	marginUsedPct := 0.0
	if totalEquity > 0 {
		marginUsedPct = (totalMarginUsed / totalEquity) * 100
	}

	// 5. 分析历史表现（最近100个周期，避免长期持仓的交易记录丢失）
	// 假设每3分钟一个周期，100个周期 = 5小时，足够覆盖大部分交易
	performance, err := at.decisionLogger.AnalyzePerformance(100)
	if err != nil {
		log.Printf("⚠️  分析历史表现失败: %v", err)
		// 不影响主流程，继续执行（但设置performance为nil以避免传递错误数据）
		performance = nil
	}

	// 6. 加载AI学习总结（如果有）
	var aiLearningSummary string
	if db := at.decisionLogger.GetDB(); db != nil {
		summary, err := db.GetActiveAILearningSummary()
		if err != nil {
			log.Printf("⚠️ 加载AI学习总结失败: %v", err)
		} else if summary != nil {
			aiLearningSummary = summary.SummaryContent
			log.Printf("📚 已加载AI学习总结（分析%d笔交易，胜率%.1f%%）", summary.TradesCount, summary.WinRate*100)
		}
	}

	// 7. 构建上下文
	ctx := &decision.Context{
		CurrentTime:       time.Now().Format("2006-01-02 15:04:05"),
		RuntimeMinutes:    int(time.Since(at.startTime).Minutes()),
		CallCount:         at.callCount,
		BTCETHLeverage:    at.config.BTCETHLeverage,  // 使用配置的杠杆倍数
		AltcoinLeverage:   at.config.AltcoinLeverage, // 使用配置的杠杆倍数
		MaxPositions:      at.config.MaxPositions,    // 使用配置的最大持仓数
		AILearningSummary: aiLearningSummary, // 添加AI学习总结
		DecisionLogger:    at.decisionLogger, // 传递DecisionLogger用于访问数据库
		Account: decision.AccountInfo{
			TotalEquity:      totalEquity,
			AvailableBalance: availableBalance,
			TotalPnL:         totalPnL,
			TotalPnLPct:      totalPnLPct,
			MarginUsed:       totalMarginUsed,
			MarginUsedPct:    marginUsedPct,
			PositionCount:    len(positionInfos),
		},
		Positions:      positionInfos,
		CandidateCoins: candidateCoins,
		Performance:    performance, // 添加历史表现分析
	}

	return ctx, autoClosedPositions, nil
}

// executeDecisionWithRecord 执行AI决策并记录详细信息
func (at *AutoTrader) executeDecisionWithRecord(decision *decision.Decision, actionRecord *logger.DecisionAction) error {
	switch decision.Action {
	case "open_long":
		return at.executeOpenLongWithRecord(decision, actionRecord)
	case "open_short":
		return at.executeOpenShortWithRecord(decision, actionRecord)
	case "close_long":
		return at.executeCloseLongWithRecord(decision, actionRecord)
	case "close_short":
		return at.executeCloseShortWithRecord(decision, actionRecord)
	case "hold", "wait":
		// 无需执行，仅记录
		return nil
	default:
		return fmt.Errorf("未知的action: %s", decision.Action)
	}
}

// executeOpenLongWithRecord 执行开多仓并记录详细信息
func (at *AutoTrader) executeOpenLongWithRecord(decision *decision.Decision, actionRecord *logger.DecisionAction) error {
	log.Printf("  📈 开多仓: %s", decision.Symbol)

	// ⚠️ 关键：检查是否已有同币种同方向持仓，如果有则拒绝开仓（防止仓位叠加超限）
	positions, err := at.trader.GetPositions()
	if err == nil {
		for _, pos := range positions {
			if pos["symbol"] == decision.Symbol && pos["side"] == "long" {
				return fmt.Errorf("❌ %s 已有多仓，拒绝开仓以防止仓位叠加超限。如需换仓，请先给出 close_long 决策", decision.Symbol)
			}
		}
	}

	// 获取当前价格
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		return err
	}

	// 计算数量
	quantity := decision.PositionSizeUSD / marketData.CurrentPrice
	actionRecord.Quantity = quantity
	actionRecord.Price = marketData.CurrentPrice

	// 开仓
	order, err := at.trader.OpenLong(decision.Symbol, quantity, decision.Leverage)
	if err != nil {
		return err
	}

	// 记录订单ID
	if orderID, ok := order["orderId"].(int64); ok {
		actionRecord.OrderID = orderID
	}

	log.Printf("  ✓ 开仓成功，订单ID: %v, 数量: %.4f", order["orderId"], quantity)

	// 记录开仓时间（内存 + 数据库）
	posKey := decision.Symbol + "_long"
	openTimeMs := time.Now().UnixMilli()
	at.positionFirstSeenTime[posKey] = openTimeMs
	
	// 保存到数据库（持久化）
	if db := at.decisionLogger.GetDB(); db != nil {
		if err := db.SavePositionOpenTime(decision.Symbol, "long", openTimeMs); err != nil {
			log.Printf("  ⚠️  保存开仓时间到数据库失败: %v", err)
		}
	}

	// 设置止损止盈
	if err := at.trader.SetStopLoss(decision.Symbol, "LONG", quantity, decision.StopLoss); err != nil {
		log.Printf("  ⚠ 设置止损失败: %v", err)
	}
	if err := at.trader.SetTakeProfit(decision.Symbol, "LONG", quantity, decision.TakeProfit); err != nil {
		log.Printf("  ⚠ 设置止盈失败: %v", err)
	}

	return nil
}

// executeOpenShortWithRecord 执行开空仓并记录详细信息
func (at *AutoTrader) executeOpenShortWithRecord(decision *decision.Decision, actionRecord *logger.DecisionAction) error {
	log.Printf("  📉 开空仓: %s", decision.Symbol)

	// ⚠️ 关键：检查是否已有同币种同方向持仓，如果有则拒绝开仓（防止仓位叠加超限）
	positions, err := at.trader.GetPositions()
	if err == nil {
		for _, pos := range positions {
			if pos["symbol"] == decision.Symbol && pos["side"] == "short" {
				return fmt.Errorf("❌ %s 已有空仓，拒绝开仓以防止仓位叠加超限。如需换仓，请先给出 close_short 决策", decision.Symbol)
			}
		}
	}

	// 获取当前价格
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		return err
	}

	// 计算数量
	quantity := decision.PositionSizeUSD / marketData.CurrentPrice
	actionRecord.Quantity = quantity
	actionRecord.Price = marketData.CurrentPrice

	// 开仓
	order, err := at.trader.OpenShort(decision.Symbol, quantity, decision.Leverage)
	if err != nil {
		return err
	}

	// 记录订单ID
	if orderID, ok := order["orderId"].(int64); ok {
		actionRecord.OrderID = orderID
	}

	log.Printf("  ✓ 开仓成功，订单ID: %v, 数量: %.4f", order["orderId"], quantity)

	// 记录开仓时间（内存 + 数据库）
	posKey := decision.Symbol + "_short"
	openTimeMs := time.Now().UnixMilli()
	at.positionFirstSeenTime[posKey] = openTimeMs
	
	// 保存到数据库（持久化）
	if db := at.decisionLogger.GetDB(); db != nil {
		if err := db.SavePositionOpenTime(decision.Symbol, "short", openTimeMs); err != nil {
			log.Printf("  ⚠️  保存开仓时间到数据库失败: %v", err)
		}
	}

	// 设置止损止盈
	if err := at.trader.SetStopLoss(decision.Symbol, "SHORT", quantity, decision.StopLoss); err != nil {
		log.Printf("  ⚠ 设置止损失败: %v", err)
	}
	if err := at.trader.SetTakeProfit(decision.Symbol, "SHORT", quantity, decision.TakeProfit); err != nil {
		log.Printf("  ⚠ 设置止盈失败: %v", err)
	}

	return nil
}

// executeCloseLongWithRecord 执行平多仓并记录详细信息（修复版：记录TradeOutcome + 防止重复平仓）
func (at *AutoTrader) executeCloseLongWithRecord(decision *decision.Decision, actionRecord *logger.DecisionAction) error {
	log.Printf("  🔄 平多仓: %s", decision.Symbol)

	// ===== 修复1: 获取平仓前的持仓信息 =====
	positions, err := at.trader.GetPositions()
	if err != nil {
		return fmt.Errorf("获取持仓失败: %w", err)
	}

	var openPrice, quantity, entryPrice float64
	var leverage int
	var openTime time.Time
	positionExists := false

	for _, pos := range positions {
		if pos["symbol"] == decision.Symbol && pos["side"] == "long" {
			entryPrice = pos["entryPrice"].(float64)
			if qty, ok := pos["positionAmt"].(float64); ok && qty > 0 {
				quantity = qty
			} else if qty, ok := pos["positionAmt"].(string); ok {
				fmt.Sscanf(qty, "%f", &quantity)
			}
			
			if lev, ok := pos["leverage"].(int); ok {
				leverage = lev
			} else if lev, ok := pos["leverage"].(float64); ok {
				leverage = int(lev)
			}
			
			openPrice = entryPrice
			
			// 从positionFirstSeenTime获取开仓时间
			posKey := decision.Symbol + "_long"
			if ts, exists := at.positionFirstSeenTime[posKey]; exists {
				openTime = time.Unix(ts/1000, (ts%1000)*1000000)
			} else {
				openTime = time.Now().Add(-30 * time.Minute) // 默认30分钟前
			}
			
			positionExists = true
			break
		}
	}

	// ===== 修复2: 检查持仓是否存在，防止重复平仓 =====
	if !positionExists {
		log.Printf("  ⚠️  %s 多仓不存在，可能已被止损/止盈自动平仓，跳过", decision.Symbol)
		actionRecord.Success = false
		actionRecord.Error = "持仓不存在（可能已自动平仓）"
		return nil // 不返回错误，避免中断流程
	}

	// 获取当前价格
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		return fmt.Errorf("获取市场数据失败: %w", err)
	}
	closePrice := marketData.CurrentPrice
	actionRecord.Price = closePrice

	// 平仓
	order, err := at.trader.CloseLong(decision.Symbol, 0) // 0 = 全部平仓
	if err != nil {
		return fmt.Errorf("平仓失败: %w", err)
	}

	// 记录订单ID
	if orderID, ok := order["orderId"].(int64); ok {
		actionRecord.OrderID = orderID
	}

	log.Printf("  ✓ 平仓成功")

	// ===== 修复3: 立即记录TradeOutcome =====
	log.Printf("  📊 持仓信息: openPrice=%.4f, quantity=%.4f, leverage=%d", openPrice, quantity, leverage)
	if openPrice > 0 && quantity > 0 {
		closeTime := time.Now()
		durationMinutes := int64(closeTime.Sub(openTime).Minutes())
		if durationMinutes < 0 {
			durationMinutes = 0
		}

		// 计算盈亏
		pnl := quantity * (closePrice - openPrice)
		positionValue := quantity * openPrice
		if leverage == 0 {
			leverage = 1
		}
		marginUsed := positionValue / float64(leverage)
		pnlPct := 0.0
		if marginUsed > 0 {
			pnlPct = (pnl / marginUsed) * 100
		}

		// 判断退出原因
		exitReason := "主动平仓"
		if actionRecord.WasStopLoss {
			exitReason = "止损/止盈触发"
		} else if pnl > 0 {
			exitReason = "主动止盈"
		} else {
			exitReason = "主动止损"
		}

		// 失败原因分析
		failureType := ""
		isPremature := durationMinutes < 30
		if pnl < 0 {
			if isPremature {
				failureType = "过早平仓（<30分钟）+ 亏损"
			} else {
				failureType = "信号判断错误或止损设置不当"
			}
		}

		trade := &logger.TradeOutcome{
			Symbol:          decision.Symbol,
			Side:            "long",
			Quantity:        quantity,
			Leverage:        leverage,
			OpenPrice:       openPrice,
			ClosePrice:      closePrice,
			PositionValue:   positionValue,
			MarginUsed:      marginUsed,
			PnL:             pnl,
			PnLPct:          pnlPct,
			DurationMinutes: durationMinutes,
			OpenTime:        openTime,
			CloseTime:       closeTime,
			WasStopLoss:     actionRecord.WasStopLoss,
			EntryReason:     decision.Reasoning,
			ExitReason:      exitReason,
			IsPremature:     isPremature,
			FailureType:     failureType,
		}

		// 保存到数据库
		if err := at.decisionLogger.SaveTradeOutcome(trade); err != nil {
			log.Printf("  ⚠️  保存交易记录失败: %v", err)
		} else {
			log.Printf("  💾 交易记录已保存: PnL=%+.2f USDT (%.2f%%), 持仓%d分钟", pnl, pnlPct, durationMinutes)
		}
	} else {
		log.Printf("  ⚠️  无法保存交易记录: openPrice=%.4f, quantity=%.4f (条件不满足)", openPrice, quantity)
	}

	// 清理持仓时间记录（内存 + 数据库）
	posKey := decision.Symbol + "_long"
	delete(at.positionFirstSeenTime, posKey)
	
	// 从数据库删除
	if db := at.decisionLogger.GetDB(); db != nil {
		if err := db.DeletePositionOpenTime(decision.Symbol, "long"); err != nil {
			log.Printf("  ⚠️  从数据库删除开仓时间失败: %v", err)
		}
	}

	return nil
}

// executeCloseShortWithRecord 执行平空仓并记录详细信息（修复版：记录TradeOutcome + 防止重复平仓）
func (at *AutoTrader) executeCloseShortWithRecord(decision *decision.Decision, actionRecord *logger.DecisionAction) error {
	log.Printf("  🔄 平空仓: %s", decision.Symbol)

	// ===== 修复1: 获取平仓前的持仓信息 =====
	positions, err := at.trader.GetPositions()
	if err != nil {
		return fmt.Errorf("获取持仓失败: %w", err)
	}

	var openPrice, quantity, entryPrice float64
	var leverage int
	var openTime time.Time
	positionExists := false

	for _, pos := range positions {
		if pos["symbol"] == decision.Symbol && pos["side"] == "short" {
			entryPrice = pos["entryPrice"].(float64)
			if qty, ok := pos["positionAmt"].(float64); ok {
				// 空仓的positionAmt是负数，取绝对值
				if qty < 0 {
					quantity = -qty
				} else {
					quantity = qty
				}
			} else if qty, ok := pos["positionAmt"].(string); ok {
				var tempQty float64
				fmt.Sscanf(qty, "%f", &tempQty)
				if tempQty < 0 {
					quantity = -tempQty
				} else {
					quantity = tempQty
				}
			}
			
			if lev, ok := pos["leverage"].(int); ok {
				leverage = lev
			} else if lev, ok := pos["leverage"].(float64); ok {
				leverage = int(lev)
			}
			
			openPrice = entryPrice
			
			// 从positionFirstSeenTime获取开仓时间
			posKey := decision.Symbol + "_short"
			if ts, exists := at.positionFirstSeenTime[posKey]; exists {
				openTime = time.Unix(ts/1000, (ts%1000)*1000000)
			} else {
				openTime = time.Now().Add(-30 * time.Minute) // 默认30分钟前
			}
			
			positionExists = true
			break
		}
	}

	// ===== 修复2: 检查持仓是否存在，防止重复平仓 =====
	if !positionExists {
		log.Printf("  ⚠️  %s 空仓不存在，可能已被止损/止盈自动平仓，跳过", decision.Symbol)
		actionRecord.Success = false
		actionRecord.Error = "持仓不存在（可能已自动平仓）"
		return nil // 不返回错误，避免中断流程
	}

	// 获取当前价格
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		return fmt.Errorf("获取市场数据失败: %w", err)
	}
	closePrice := marketData.CurrentPrice
	actionRecord.Price = closePrice

	// 平仓
	order, err := at.trader.CloseShort(decision.Symbol, 0) // 0 = 全部平仓
	if err != nil {
		return fmt.Errorf("平仓失败: %w", err)
	}

	// 记录订单ID
	if orderID, ok := order["orderId"].(int64); ok {
		actionRecord.OrderID = orderID
	}

	log.Printf("  ✓ 平仓成功")

	// ===== 修复3: 立即记录TradeOutcome =====
	log.Printf("  📊 持仓信息: openPrice=%.4f, quantity=%.4f, leverage=%d", openPrice, quantity, leverage)
	if openPrice > 0 && quantity > 0 {
		closeTime := time.Now()
		durationMinutes := int64(closeTime.Sub(openTime).Minutes())
		if durationMinutes < 0 {
			durationMinutes = 0
		}

		// 计算盈亏（做空盈亏计算）
		pnl := quantity * (openPrice - closePrice)
		positionValue := quantity * openPrice
		if leverage == 0 {
			leverage = 1
		}
		marginUsed := positionValue / float64(leverage)
		pnlPct := 0.0
		if marginUsed > 0 {
			pnlPct = (pnl / marginUsed) * 100
		}

		// 判断退出原因
		exitReason := "主动平仓"
		if actionRecord.WasStopLoss {
			exitReason = "止损/止盈触发"
		} else if pnl > 0 {
			exitReason = "主动止盈"
		} else {
			exitReason = "主动止损"
		}

		// 失败原因分析
		failureType := ""
		isPremature := durationMinutes < 30
		if pnl < 0 {
			if isPremature {
				failureType = "过早平仓（<30分钟）+ 亏损"
			} else {
				failureType = "信号判断错误或止损设置不当"
			}
		}

		trade := &logger.TradeOutcome{
			Symbol:          decision.Symbol,
			Side:            "short",
			Quantity:        quantity,
			Leverage:        leverage,
			OpenPrice:       openPrice,
			ClosePrice:      closePrice,
			PositionValue:   positionValue,
			MarginUsed:      marginUsed,
			PnL:             pnl,
			PnLPct:          pnlPct,
			DurationMinutes: durationMinutes,
			OpenTime:        openTime,
			CloseTime:       closeTime,
			WasStopLoss:     actionRecord.WasStopLoss,
			EntryReason:     decision.Reasoning,
			ExitReason:      exitReason,
			IsPremature:     isPremature,
			FailureType:     failureType,
		}

		// 保存到数据库
		if err := at.decisionLogger.SaveTradeOutcome(trade); err != nil {
			log.Printf("  ⚠️  保存交易记录失败: %v", err)
		} else {
			log.Printf("  💾 交易记录已保存: PnL=%+.2f USDT (%.2f%%), 持仓%d分钟", pnl, pnlPct, durationMinutes)
		}
	} else {
		log.Printf("  ⚠️  无法保存交易记录: openPrice=%.4f, quantity=%.4f (条件不满足)", openPrice, quantity)
	}

	// 清理持仓时间记录（内存 + 数据库）
	posKey := decision.Symbol + "_short"
	delete(at.positionFirstSeenTime, posKey)
	
	// 从数据库删除
	if db := at.decisionLogger.GetDB(); db != nil {
		if err := db.DeletePositionOpenTime(decision.Symbol, "short"); err != nil {
			log.Printf("  ⚠️  从数据库删除开仓时间失败: %v", err)
		}
	}

	return nil
}

// saveAutoClosedTradeOutcome 保存自动平仓的交易记录（从Binance历史订单获取完整信息）
func (at *AutoTrader) saveAutoClosedTradeOutcome(symbol string, side string, closePrice float64) {
	// 尝试从positionFirstSeenTime获取开仓时间
	posKey := symbol + "_" + side
	openTime := time.Now().Add(-30 * time.Minute) // 默认30分钟前
	if ts, exists := at.positionFirstSeenTime[posKey]; exists {
		openTime = time.Unix(ts/1000, (ts%1000)*1000000)
	}
	
	closeTime := time.Now()
	durationMinutes := int64(closeTime.Sub(openTime).Minutes())
	if durationMinutes < 0 {
		durationMinutes = 0
	}
	
	// 尝试从Binance历史订单获取完整信息
	var quantity, openPrice, leverage float64
	var realizedPnl float64
	
	trades, err := at.trader.GetAccountTrades(symbol, 20) // 获取最近20条成交记录
	if err == nil && len(trades) > 0 {
		// 找到最近的平仓成交（根据positionSide和side判断）
		for _, trade := range trades {
			tradeSide, _ := trade["side"].(string)
			positionSide, _ := trade["positionSide"].(string)
			tradeTime, _ := trade["time"].(int64)
			
			// 匹配平仓订单：时间在5分钟内 + 方向匹配
			if time.Since(time.UnixMilli(tradeTime)) < 5*time.Minute {
				// Binance BOTH模式：平多是SELL，平空是BUY
				if (side == "long" && positionSide == "BOTH" && tradeSide == "SELL") ||
				   (side == "short" && positionSide == "BOTH" && tradeSide == "BUY") ||
				   (side == "long" && positionSide == "LONG") ||
				   (side == "short" && positionSide == "SHORT") {
					
					// 找到平仓订单
					if price, ok := trade["price"].(float64); ok {
						closePrice = price
					}
					if qty, ok := trade["qty"].(float64); ok {
						quantity = qty
					}
					if pnl, ok := trade["realizedPnl"].(float64); ok {
						realizedPnl = pnl
					}
					
					log.Printf("  📊 从历史订单获取平仓信息: price=%.4f, qty=%.4f, pnl=%.2f", closePrice, quantity, realizedPnl)
					break
				}
			}
		}
		
		// 查找对应的开仓订单（从后往前找，因为开仓在前）
		for i := len(trades) - 1; i >= 0; i-- {
			trade := trades[i]
			tradeSide, _ := trade["side"].(string)
			positionSide, _ := trade["positionSide"].(string)
			tradeTime, _ := trade["time"].(int64)
			tradeTimestamp := time.UnixMilli(tradeTime)
			
			// 开仓订单必须在openTime附近（±5分钟）
			if tradeTimestamp.After(openTime.Add(-5*time.Minute)) && tradeTimestamp.Before(openTime.Add(5*time.Minute)) {
				if (side == "long" && positionSide == "BOTH" && tradeSide == "BUY") ||
				   (side == "short" && positionSide == "BOTH" && tradeSide == "SELL") ||
				   (side == "long" && positionSide == "LONG") ||
				   (side == "short" && positionSide == "SHORT") {
					
					if price, ok := trade["price"].(float64); ok {
						openPrice = price
						log.Printf("  📊 从历史订单获取开仓信息: openPrice=%.4f", openPrice)
					}
					break
				}
			}
		}
	}
	
	// 如果获取不到数量，尝试估算（使用realizedPnl反推）
	if quantity == 0 && realizedPnl != 0 && openPrice > 0 && closePrice > 0 {
		priceDiff := closePrice - openPrice
		if side == "short" {
			priceDiff = openPrice - closePrice
		}
		if priceDiff != 0 {
			quantity = realizedPnl / priceDiff
			log.Printf("  📊 根据盈亏反推数量: %.4f", quantity)
		}
	}
	
	// 计算leverage（如果有数量和价格）
	if quantity > 0 && openPrice > 0 {
		leverage = 10 // 默认杠杆
	}
	
	// 计算盈亏
	pnl := realizedPnl
	if pnl == 0 && quantity > 0 && openPrice > 0 {
		if side == "long" {
			pnl = quantity * (closePrice - openPrice)
		} else {
			pnl = quantity * (openPrice - closePrice)
		}
	}
	
	positionValue := quantity * openPrice
	marginUsed := positionValue / float64(leverage)
	pnlPct := 0.0
	if marginUsed > 0 {
		pnlPct = (pnl / marginUsed) * 100
	}
	
	// 构建交易记录
	trade := &logger.TradeOutcome{
		Symbol:          symbol,
		Side:            side,
		Quantity:        quantity,
		Leverage:        int(leverage),
		OpenPrice:       openPrice,
		ClosePrice:      closePrice,
		PositionValue:   positionValue,
		MarginUsed:      marginUsed,
		PnL:             pnl,
		PnLPct:          pnlPct,
		DurationMinutes: durationMinutes,
		OpenTime:        openTime,
		CloseTime:       closeTime,
		WasStopLoss:     true,
		EntryReason:     "AI自动开仓",
		ExitReason:      "止损/止盈自动触发",
		IsPremature:     durationMinutes < 30,
		FailureType:     func() string {
			if pnl < 0 && durationMinutes < 30 {
				return "止损触发+过早平仓"
			} else if pnl < 0 {
				return "止损触发"
			}
			return ""
		}(),
	}
	
	// 保存到数据库
	if err := at.decisionLogger.SaveTradeOutcome(trade); err != nil {
		log.Printf("  ⚠️  保存自动平仓记录失败: %v", err)
	} else {
		log.Printf("  💾 已记录自动平仓: %s %s, PnL=%+.2f USDT (%.2f%%), 持仓%d分钟", 
			symbol, side, pnl, pnlPct, durationMinutes)
	}
}

// GetID 获取trader ID
func (at *AutoTrader) GetID() string {
	return at.id
}

// GetName 获取trader名称
func (at *AutoTrader) GetName() string {
	return at.name
}

// GetAIModel 获取AI模型
func (at *AutoTrader) GetAIModel() string {
	return at.aiModel
}

// GetDecisionLogger 获取决策日志记录器
func (at *AutoTrader) GetDecisionLogger() *logger.DecisionLogger {
	return at.decisionLogger
}

// GetStatus 获取系统状态（用于API）
func (at *AutoTrader) GetStatus() map[string]interface{} {
	aiProvider := "DeepSeek"
	if at.config.UseQwen {
		aiProvider = "Qwen"
	}

	at.mu.RLock()
	defer at.mu.RUnlock()
	
	return map[string]interface{}{
		"trader_id":       at.id,
		"trader_name":     at.name,
		"ai_model":        at.aiModel,
		"exchange":        at.exchange,
		"is_running":      at.isRunning && !at.isPaused,
		"is_paused":       at.isPaused,
		"start_time":      at.startTime.Format(time.RFC3339),
		"runtime_minutes": int(time.Since(at.startTime).Minutes()),
		"call_count":      at.callCount,
		"initial_balance": at.initialBalance,
		"scan_interval":   at.config.ScanInterval.String(),
		"stop_until":      at.stopUntil.Format(time.RFC3339),
		"last_reset_time": at.lastResetTime.Format(time.RFC3339),
		"ai_provider":     aiProvider,
	}
}

// Pause 暂停trader
func (at *AutoTrader) Pause() {
	at.mu.Lock()
	defer at.mu.Unlock()
	
	at.isPaused = true
	
	// 保存状态到数据库
	if db := at.decisionLogger.GetDB(); db != nil {
		if err := db.SaveTraderState(true); err != nil {
			log.Printf("[%s] ⚠️  保存暂停状态到数据库失败: %v", at.name, err)
		}
	}
	
	log.Printf("[%s] ⏸️  Trader已暂停", at.name)
}

// Resume 恢复trader
func (at *AutoTrader) Resume() {
	at.mu.Lock()
	defer at.mu.Unlock()
	
	at.isPaused = false
	
	// 保存状态到数据库
	if db := at.decisionLogger.GetDB(); db != nil {
		if err := db.SaveTraderState(false); err != nil {
			log.Printf("[%s] ⚠️  保存运行状态到数据库失败: %v", at.name, err)
		}
	}
	
	log.Printf("[%s] ▶️  Trader已恢复", at.name)
}

// IsPaused 检查是否暂停
func (at *AutoTrader) IsPaused() bool {
	at.mu.RLock()
	defer at.mu.RUnlock()
	
	return at.isPaused
}

// GetPositionOpenTime 获取持仓的开仓时间
func (at *AutoTrader) GetPositionOpenTime(symbol string, side string) (time.Time, bool) {
	at.mu.RLock()
	defer at.mu.RUnlock()
	
	posKey := symbol + "_" + side
	if ts, exists := at.positionFirstSeenTime[posKey]; exists {
		return time.Unix(ts/1000, (ts%1000)*1000000), true
	}
	return time.Time{}, false
}

// ManualClosePosition 手动平仓
func (at *AutoTrader) ManualClosePosition(symbol string, side string) error {
	log.Printf("[%s] 📤 手动平仓请求: %s %s", at.name, symbol, side)
	
	// 获取当前持仓
	positions, err := at.trader.GetPositions()
	if err != nil {
		return fmt.Errorf("获取持仓失败: %w", err)
	}
	
	// 查找对应的持仓
	var targetPosition map[string]interface{}
	for _, pos := range positions {
		if pos["symbol"] == symbol && pos["side"] == side {
			targetPosition = pos
			break
		}
	}
	
	if targetPosition == nil {
		return fmt.Errorf("未找到持仓: %s %s", symbol, side)
	}
	
	// 获取持仓数量
	quantity := 0.0
	if positionAmt, ok := targetPosition["positionAmt"].(float64); ok {
		quantity = positionAmt
		if quantity < 0 {
			quantity = -quantity // 转为正数
		}
	} else {
		return fmt.Errorf("无法获取持仓数量")
	}
	
	// 执行平仓
	var result map[string]interface{}
	var closeErr error
	if side == "long" {
		result, closeErr = at.trader.CloseLong(symbol, quantity)
	} else if side == "short" {
		result, closeErr = at.trader.CloseShort(symbol, quantity)
	} else {
		return fmt.Errorf("无效的持仓方向: %s", side)
	}
	
	if closeErr != nil {
		return fmt.Errorf("平仓失败: %w", closeErr)
	}
	
	// 记录订单ID（如果有）
	if orderId, ok := result["order_id"].(string); ok {
		log.Printf("[%s] 📝 平仓订单ID: %s", at.name, orderId)
	}
	
	// 清理持仓时间记录（内存 + 数据库）
	at.mu.Lock()
	posKey := symbol + "_" + side
	delete(at.positionFirstSeenTime, posKey)
	at.mu.Unlock()
	
	// 从数据库删除
	if db := at.decisionLogger.GetDB(); db != nil {
		if err := db.DeletePositionOpenTime(symbol, side); err != nil {
			log.Printf("[%s] ⚠️  从数据库删除开仓时间失败: %v", at.name, err)
		}
	}
	
	log.Printf("[%s] ✅ 手动平仓成功: %s %s", at.name, symbol, side)
	return nil
}

// GetAccountInfo 获取账户信息（用于API）
func (at *AutoTrader) GetAccountInfo() (map[string]interface{}, error) {
	balance, err := at.trader.GetBalance()
	if err != nil {
		return nil, fmt.Errorf("获取余额失败: %w", err)
	}

	// 获取账户字段
	totalWalletBalance := 0.0
	totalUnrealizedProfit := 0.0
	availableBalance := 0.0

	if wallet, ok := balance["totalWalletBalance"].(float64); ok {
		totalWalletBalance = wallet
	}
	if unrealized, ok := balance["totalUnrealizedProfit"].(float64); ok {
		totalUnrealizedProfit = unrealized
	}
	if avail, ok := balance["availableBalance"].(float64); ok {
		availableBalance = avail
	}

	// Total Equity = 钱包余额 + 未实现盈亏
	totalEquity := totalWalletBalance + totalUnrealizedProfit

	// 获取持仓计算总保证金
	positions, err := at.trader.GetPositions()
	if err != nil {
		return nil, fmt.Errorf("获取持仓失败: %w", err)
	}

	totalMarginUsed := 0.0
	totalUnrealizedPnL := 0.0
	for _, pos := range positions {
		markPrice := pos["markPrice"].(float64)
		quantity := pos["positionAmt"].(float64)
		if quantity < 0 {
			quantity = -quantity
		}
		unrealizedPnl := pos["unRealizedProfit"].(float64)
		totalUnrealizedPnL += unrealizedPnl

		leverage := 10
		if lev, ok := pos["leverage"].(float64); ok {
			leverage = int(lev)
		}
		marginUsed := (quantity * markPrice) / float64(leverage)
		totalMarginUsed += marginUsed
	}

	totalPnL := totalEquity - at.initialBalance
	totalPnLPct := 0.0
	if at.initialBalance > 0 {
		totalPnLPct = (totalPnL / at.initialBalance) * 100
	}

	marginUsedPct := 0.0
	if totalEquity > 0 {
		marginUsedPct = (totalMarginUsed / totalEquity) * 100
	}

	return map[string]interface{}{
		// 核心字段
		"total_equity":      totalEquity,           // 账户净值 = wallet + unrealized
		"wallet_balance":    totalWalletBalance,    // 钱包余额（不含未实现盈亏）
		"unrealized_profit": totalUnrealizedProfit, // 未实现盈亏（从API）
		"available_balance": availableBalance,      // 可用余额

		// 盈亏统计
		"total_pnl":            totalPnL,           // 总盈亏 = equity - initial
		"total_pnl_pct":        totalPnLPct,        // 总盈亏百分比
		"total_unrealized_pnl": totalUnrealizedPnL, // 未实现盈亏（从持仓计算）
		"initial_balance":      at.initialBalance,  // 初始余额
		"daily_pnl":            at.dailyPnL,        // 日盈亏

		// 持仓信息
		"position_count":  len(positions),  // 持仓数量
		"margin_used":     totalMarginUsed, // 保证金占用
		"margin_used_pct": marginUsedPct,   // 保证金使用率
	}, nil
}

// CallAI 调用AI（供外部使用，如生成学习总结）
func (at *AutoTrader) CallAI(systemPrompt, userPrompt string) (string, error) {
	if at.mcpClient == nil {
		return "", fmt.Errorf("MCP客户端未初始化")
	}
	return at.mcpClient.CallWithMessages(systemPrompt, userPrompt)
}

// maybeGenerateAILearningSummary 检查是否需要生成AI学习总结
func (at *AutoTrader) maybeGenerateAILearningSummary() {
	db := at.decisionLogger.GetDB()
	if db == nil {
		log.Printf("⚠️  [%s] 数据库未初始化，跳过AI学习总结生成", at.name)
		return
	}

	// 获取最近的交易记录
	trades, err := db.GetTradeOutcomes(20)
	if err != nil {
		log.Printf("⚠️  [%s] 获取交易记录失败: %v，跳过AI学习总结生成", at.name, err)
		return
	}
	if len(trades) < 5 {
		log.Printf("⚠️  [%s] 交易记录不足（%d笔 < 5笔），跳过AI学习总结生成", at.name, len(trades))
		return // 交易太少，跳过
	}

	log.Printf("🤖 [%s] 正在生成AI学习总结（分析最近%d笔交易）...", at.name, len(trades))

	// 构建分析prompt
	systemPrompt := `你是一个专业的加密货币交易分析师。请分析这些历史交易记录，用简洁的Markdown格式输出总结。

要求：
1. 找出3个最关键的失败模式（什么总是导致亏损）
2. 找出2个成功模式（什么策略有效）
3. 提出3条具体的改进建议

格式：
## ❌ 避免这些错误
1. [具体错误模式，1句话]
2. ...

## ✅ 复制这些成功策略
1. [具体成功模式，1句话]
2. ...

## 💡 改进建议
1. [具体建议，1句话]
2. ...

保持简洁，每个要点不超过15个字。`

	userPrompt := at.buildTradeAnalysisPrompt(trades)

	// 调用AI
	summary, err := at.mcpClient.CallWithMessages(systemPrompt, userPrompt)
	if err != nil {
		log.Printf("❌ [%s] AI分析失败: %v", at.name, err)
		return
	}

	// 计算统计数据
	winCount := 0
	totalPnL := 0.0
	for _, trade := range trades {
		if trade.PnL > 0 {
			winCount++
		}
		totalPnL += trade.PnL
	}
	winRate := float64(winCount) / float64(len(trades))
	avgPnL := totalPnL / float64(len(trades))

	dateStart := trades[len(trades)-1].OpenTime.Format("2006-01-02")
	dateEnd := trades[0].CloseTime.Format("2006-01-02")

	// 保存到数据库
	aiSummary := &database.AILearningSummary{
		TraderID:       at.id,
		SummaryContent: summary,
		TradesCount:    len(trades),
		DateRangeStart: dateStart,
		DateRangeEnd:   dateEnd,
		WinRate:        winRate,
		AvgPnL:         avgPnL,
		IsActive:       true,
	}

	if err := db.SaveAILearningSummary(aiSummary); err != nil {
		log.Printf("❌ [%s] 保存AI总结失败: %v", at.name, err)
		return
	}

	log.Printf("✅ [%s] AI学习总结已生成并保存（分析%d笔，胜率%.1f%%）", 
		at.name, len(trades), winRate*100)
	log.Printf("📚 总结内容：\n%s", summary)
}

// buildTradeAnalysisPrompt 构建交易分析prompt
func (at *AutoTrader) buildTradeAnalysisPrompt(trades []*database.TradeOutcome) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# 最近%d笔交易记录\n\n", len(trades)))

	for i, trade := range trades {
		emoji := "✅"
		if trade.PnL < 0 {
			emoji = "❌"
		}

		sb.WriteString(fmt.Sprintf("%d. %s %s %s\n", i+1, emoji, trade.Symbol, strings.ToUpper(trade.Side)))
		sb.WriteString(fmt.Sprintf("   盈亏: %.2f USDT (%.1f%%) | 持仓: %d分钟\n", 
			trade.PnL, trade.PnLPct, trade.DurationMinutes))
		
		if trade.FailureType != "" {
			sb.WriteString(fmt.Sprintf("   失败: %s\n", trade.FailureType))
		}
		if trade.IsPremature {
			sb.WriteString("   ⚠️ 过早平仓\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// GetPositions 获取持仓列表（用于API）
func (at *AutoTrader) GetPositions() ([]map[string]interface{}, error) {
	positions, err := at.trader.GetPositions()
	if err != nil {
		return nil, fmt.Errorf("获取持仓失败: %w", err)
	}

	var result []map[string]interface{}
	for _, pos := range positions {
		symbol := pos["symbol"].(string)
		side := pos["side"].(string)
		entryPrice := pos["entryPrice"].(float64)
		markPrice := pos["markPrice"].(float64)
		quantity := pos["positionAmt"].(float64)
		if quantity < 0 {
			quantity = -quantity
		}
		unrealizedPnl := pos["unRealizedProfit"].(float64)
		liquidationPrice := pos["liquidationPrice"].(float64)

		leverage := 10
		if lev, ok := pos["leverage"].(float64); ok {
			leverage = int(lev)
		}

		pnlPct := 0.0
		if side == "long" {
			pnlPct = ((markPrice - entryPrice) / entryPrice) * float64(leverage) * 100
		} else {
			pnlPct = ((entryPrice - markPrice) / entryPrice) * float64(leverage) * 100
		}

		marginUsed := (quantity * markPrice) / float64(leverage)

		// 获取开仓时间和持仓时长
		posKey := symbol + "_" + side
		openTime := ""
		holdingMinutes := int64(0)
		at.mu.RLock()
		if openTimeMs, exists := at.positionFirstSeenTime[posKey]; exists {
			openTimeObj := time.Unix(openTimeMs/1000, (openTimeMs%1000)*1000000)
			openTime = openTimeObj.Format(time.RFC3339)
			holdingMinutes = int64(time.Now().Sub(openTimeObj).Minutes())
		}
		at.mu.RUnlock()

		result = append(result, map[string]interface{}{
			"symbol":             symbol,
			"side":               side,
			"entry_price":        entryPrice,
			"mark_price":         markPrice,
			"quantity":           quantity,
			"leverage":           leverage,
			"unrealized_pnl":     unrealizedPnl,
			"unrealized_pnl_pct": pnlPct,
			"liquidation_price":  liquidationPrice,
			"margin_used":        marginUsed,
			"open_time":          openTime,
			"holding_minutes":    holdingMinutes,
		})
	}

	return result, nil
}

// sortDecisionsByPriority 对决策排序：先平仓，再开仓，最后hold/wait
// 这样可以避免换仓时仓位叠加超限
func sortDecisionsByPriority(decisions []decision.Decision) []decision.Decision {
	if len(decisions) <= 1 {
		return decisions
	}

	// 定义优先级
	getActionPriority := func(action string) int {
		switch action {
		case "close_long", "close_short":
			return 1 // 最高优先级：先平仓
		case "open_long", "open_short":
			return 2 // 次优先级：后开仓
		case "hold", "wait":
			return 3 // 最低优先级：观望
		default:
			return 999 // 未知动作放最后
		}
	}

	// 复制决策列表
	sorted := make([]decision.Decision, len(decisions))
	copy(sorted, decisions)

	// 按优先级排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if getActionPriority(sorted[i].Action) > getActionPriority(sorted[j].Action) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}
