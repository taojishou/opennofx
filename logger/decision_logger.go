package logger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"nofx/database"
	"os"
	"path/filepath"
	"time"
)

// DecisionRecord 决策记录
type DecisionRecord struct {
	Timestamp      time.Time          `json:"timestamp"`       // 决策时间
	CycleNumber    int                `json:"cycle_number"`    // 周期编号
	SystemPrompt   string             `json:"system_prompt"`   // System Prompt（规则）
	InputPrompt    string             `json:"input_prompt"`    // User Prompt（市场数据）
	CoTTrace       string             `json:"cot_trace"`       // AI思维链（输出）
	DecisionJSON   string             `json:"decision_json"`   // 决策JSON
	AccountState   AccountSnapshot    `json:"account_state"`   // 账户状态快照
	Positions      []PositionSnapshot `json:"positions"`       // 持仓快照
	CandidateCoins []string           `json:"candidate_coins"` // 候选币种列表
	Decisions      []DecisionAction   `json:"decisions"`       // 执行的决策
	ExecutionLog   []string           `json:"execution_log"`   // 执行日志
	Success        bool               `json:"success"`         // 是否成功
	ErrorMessage   string             `json:"error_message"`   // 错误信息（如果有）
}

// AccountSnapshot 账户状态快照
type AccountSnapshot struct {
	TotalBalance          float64 `json:"total_balance"`
	AvailableBalance      float64 `json:"available_balance"`
	TotalUnrealizedProfit float64 `json:"total_unrealized_profit"`
	PositionCount         int     `json:"position_count"`
	MarginUsedPct         float64 `json:"margin_used_pct"`
}

// PositionSnapshot 持仓快照
type PositionSnapshot struct {
	Symbol           string  `json:"symbol"`
	Side             string  `json:"side"`
	PositionAmt      float64 `json:"position_amt"`
	EntryPrice       float64 `json:"entry_price"`
	MarkPrice        float64 `json:"mark_price"`
	UnrealizedProfit float64 `json:"unrealized_profit"`
	Leverage         float64 `json:"leverage"`
	LiquidationPrice float64 `json:"liquidation_price"`
}

// DecisionAction 决策动作
type DecisionAction struct {
	Action      string    `json:"action"`        // open_long, open_short, close_long, close_short
	Symbol      string    `json:"symbol"`        // 币种
	Quantity    float64   `json:"quantity"`      // 数量
	Leverage    int       `json:"leverage"`      // 杠杆（开仓时）
	Price       float64   `json:"price"`         // 执行价格
	OrderID     int64     `json:"order_id"`      // 订单ID
	Timestamp   time.Time `json:"timestamp"`     // 执行时间
	Success     bool      `json:"success"`       // 是否成功
	Error       string    `json:"error"`         // 错误信息
	WasStopLoss bool      `json:"was_stop_loss"` // 是否因止损触发（平仓时）
}

// DecisionLogger 决策日志记录器
type DecisionLogger struct {
	logDir      string
	cycleNumber int
	db          *database.DB // SQLite数据库连接
	traderID    string       // Trader ID
}

// NewDecisionLogger 创建决策日志记录器
func NewDecisionLogger(logDir string) *DecisionLogger {
	if logDir == "" {
		logDir = "decision_logs"
	}

	// 确保日志目录存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("⚠ 创建日志目录失败: %v\n", err)
	}

	// 从目录路径提取 trader ID (decision_logs/trader_id)
	traderID := filepath.Base(logDir)

	// 初始化SQLite数据库
	db, err := database.New(traderID)
	if err != nil {
		fmt.Printf("⚠ 初始化SQLite数据库失败: %v\n", err)
		// 继续运行，只是没有数据库支持
		db = nil
	}

	return &DecisionLogger{
		logDir:      logDir,
		cycleNumber: 0,
		db:          db,
		traderID:    traderID,
	}
}

// GetDB 获取数据库连接
func (l *DecisionLogger) GetDB() *database.DB {
	return l.db
}

// LogDecision 记录决策（只保存到数据库）
func (l *DecisionLogger) LogDecision(record *DecisionRecord) error {
	l.cycleNumber++
	record.CycleNumber = l.cycleNumber
	record.Timestamp = time.Now()

	// 保存到 SQLite 数据库
	if l.db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	if err := l.saveToDatabase(record); err != nil {
		return fmt.Errorf("保存到数据库失败: %w", err)
	}

	fmt.Printf("📝 决策记录已保存到数据库: cycle %d\n", record.CycleNumber)
	return nil
}

// saveToDatabase 保存决策记录到数据库
func (l *DecisionLogger) saveToDatabase(record *DecisionRecord) error {
	// 转换 DecisionJSON 为字符串
	decisionJSON := ""
	if record.DecisionJSON != "" {
		decisionJSON = record.DecisionJSON
	}

	// 插入主记录
	dbRecord := &database.DecisionRecord{
		TraderID:              l.traderID,
		CycleNumber:           record.CycleNumber,
		Timestamp:             record.Timestamp,
		SystemPrompt:          record.SystemPrompt,
		InputPrompt:           record.InputPrompt,
		CoTTrace:              record.CoTTrace,
		DecisionJSON:          decisionJSON,
		Success:               record.Success,
		ErrorMessage:          record.ErrorMessage,
		TotalBalance:          record.AccountState.TotalBalance,
		AvailableBalance:      record.AccountState.AvailableBalance,
		TotalUnrealizedProfit: record.AccountState.TotalUnrealizedProfit,
		PositionCount:         record.AccountState.PositionCount,
		MarginUsedPct:         record.AccountState.MarginUsedPct,
	}

	recordID, err := l.db.InsertDecisionRecord(dbRecord)
	if err != nil {
		return fmt.Errorf("插入决策记录失败: %w", err)
	}

	// 插入决策动作
	for _, action := range record.Decisions {
		dbAction := &database.DecisionAction{
			RecordID:    recordID,
			Action:      action.Action,
			Symbol:      action.Symbol,
			Quantity:    action.Quantity,
			Leverage:    action.Leverage,
			Price:       action.Price,
			OrderID:     action.OrderID,
			Timestamp:   action.Timestamp,
			Success:     action.Success,
			Error:       action.Error,
			WasStopLoss: action.WasStopLoss,
		}
		if err := l.db.InsertDecisionAction(dbAction); err != nil {
			return fmt.Errorf("插入决策动作失败: %w", err)
		}
	}

	// 插入持仓快照
	for _, pos := range record.Positions {
		dbPos := &database.PositionSnapshot{
			RecordID:         recordID,
			Symbol:           pos.Symbol,
			Side:             pos.Side,
			PositionAmt:      pos.PositionAmt,
			EntryPrice:       pos.EntryPrice,
			MarkPrice:        pos.MarkPrice,
			UnrealizedProfit: pos.UnrealizedProfit,
			Leverage:         pos.Leverage,
			LiquidationPrice: pos.LiquidationPrice,
		}
		if err := l.db.InsertPositionSnapshot(dbPos); err != nil {
			return fmt.Errorf("插入持仓快照失败: %w", err)
		}
	}

	// 插入候选币种
	for _, symbol := range record.CandidateCoins {
		if err := l.db.InsertCandidateCoin(recordID, symbol); err != nil {
			return fmt.Errorf("插入候选币种失败: %w", err)
		}
	}

	return nil
}

// GetLatestRecords 获取最近N条记录（按时间正序：从旧到新）
func (l *DecisionLogger) GetLatestRecords(n int) ([]*DecisionRecord, error) {
	if l.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	dbRecords, err := l.db.GetLatestRecords(n)
	if err != nil {
		return nil, err
	}
	
	// 转换类型：database.DecisionRecord -> logger.DecisionRecord
	records := make([]*DecisionRecord, len(dbRecords))
	for i, dbRec := range dbRecords {
		// 从数据库加载该记录的所有决策动作
		actions, err := l.db.QueryActions(dbRec.ID)
		if err != nil {
			log.Printf("⚠️ 加载record %d 的决策动作失败: %v", dbRec.ID, err)
			actions = []*database.DecisionAction{} // 使用空数组
		}
		
		// 转换decision actions
		var loggerActions []DecisionAction
		for _, act := range actions {
			loggerActions = append(loggerActions, DecisionAction{
				Action:      act.Action,
				Symbol:      act.Symbol,
				Quantity:    act.Quantity,
				Leverage:    act.Leverage,
				Price:       act.Price,
				OrderID:     act.OrderID,
				Timestamp:   act.Timestamp,
				Success:     act.Success,
				Error:       act.Error,
				WasStopLoss: act.WasStopLoss,
			})
		}
		
		records[i] = &DecisionRecord{
			Timestamp:    dbRec.Timestamp,
			CycleNumber:  dbRec.CycleNumber,
			InputPrompt:  dbRec.InputPrompt,
			CoTTrace:     dbRec.CoTTrace,
			DecisionJSON: dbRec.DecisionJSON,
			Success:      dbRec.Success,
			ErrorMessage: dbRec.ErrorMessage,
			Decisions:    loggerActions, // 加载关联的决策动作
			AccountState: AccountSnapshot{
				TotalBalance:          dbRec.TotalBalance,
				AvailableBalance:      dbRec.AvailableBalance,
				TotalUnrealizedProfit: dbRec.TotalUnrealizedProfit,
				PositionCount:         dbRec.PositionCount,
				MarginUsedPct:         dbRec.MarginUsedPct,
			},
		}
	}
	return records, nil
}

// GetRecordByDate 获取指定日期的所有记录
func (l *DecisionLogger) GetRecordByDate(date time.Time) ([]*DecisionRecord, error) {
	dateStr := date.Format("20060102")
	pattern := filepath.Join(l.logDir, fmt.Sprintf("decision_%s_*.json", dateStr))

	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("查找日志文件失败: %w", err)
	}

	var records []*DecisionRecord
	for _, filepath := range files {
		data, err := ioutil.ReadFile(filepath)
		if err != nil {
			continue
		}

		var record DecisionRecord
		if err := json.Unmarshal(data, &record); err != nil {
			continue
		}

		records = append(records, &record)
	}

	return records, nil
}

// CleanOldRecords 清理N天前的旧记录
func (l *DecisionLogger) CleanOldRecords(days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days)

	files, err := ioutil.ReadDir(l.logDir)
	if err != nil {
		return fmt.Errorf("读取日志目录失败: %w", err)
	}

	removedCount := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if file.ModTime().Before(cutoffTime) {
			filepath := filepath.Join(l.logDir, file.Name())
			if err := os.Remove(filepath); err != nil {
				fmt.Printf("⚠ 删除旧记录失败 %s: %v\n", file.Name(), err)
				continue
			}
			removedCount++
		}
	}

	if removedCount > 0 {
		fmt.Printf("🗑️ 已清理 %d 条旧记录（%d天前）\n", removedCount, days)
	}

	return nil
}

// GetStatistics 获取统计信息
func (l *DecisionLogger) GetStatistics() (*Statistics, error) {
	files, err := ioutil.ReadDir(l.logDir)
	if err != nil {
		return nil, fmt.Errorf("读取日志目录失败: %w", err)
	}

	stats := &Statistics{}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filepath := filepath.Join(l.logDir, file.Name())
		data, err := ioutil.ReadFile(filepath)
		if err != nil {
			continue
		}

		var record DecisionRecord
		if err := json.Unmarshal(data, &record); err != nil {
			continue
		}

		stats.TotalCycles++

		for _, action := range record.Decisions {
			if action.Success {
				switch action.Action {
				case "open_long", "open_short":
					stats.TotalOpenPositions++
				case "close_long", "close_short":
					stats.TotalClosePositions++
				}
			}
		}

		if record.Success {
			stats.SuccessfulCycles++
		} else {
			stats.FailedCycles++
		}
	}

	return stats, nil
}

// Statistics 统计信息
type Statistics struct {
	TotalCycles         int `json:"total_cycles"`
	SuccessfulCycles    int `json:"successful_cycles"`
	FailedCycles        int `json:"failed_cycles"`
	TotalOpenPositions  int `json:"total_open_positions"`
	TotalClosePositions int `json:"total_close_positions"`
}

// TradeOutcome 单笔交易结果
type TradeOutcome struct {
	Symbol        string    `json:"symbol"`         // 币种
	Side          string    `json:"side"`           // long/short
	Quantity      float64   `json:"quantity"`       // 仓位数量
	Leverage      int       `json:"leverage"`       // 杠杆倍数
	OpenPrice     float64   `json:"open_price"`     // 开仓价
	ClosePrice    float64   `json:"close_price"`    // 平仓价
	PositionValue float64   `json:"position_value"` // 仓位价值（quantity × openPrice）
	MarginUsed    float64   `json:"margin_used"`    // 保证金使用（positionValue / leverage）
	PnL           float64   `json:"pn_l"`           // 盈亏（USDT）
	PnLPct        float64   `json:"pn_l_pct"`       // 盈亏百分比（相对保证金）
	Duration      string    `json:"duration"`       // 持仓时长
	DurationMinutes int64   `json:"duration_minutes"` // 持仓时长（分钟）
	OpenTime      time.Time `json:"open_time"`      // 开仓时间
	CloseTime     time.Time `json:"close_time"`     // 平仓时间
	WasStopLoss   bool      `json:"was_stop_loss"`  // 是否止损
	
	// 新增：开仓时的市场状态（用于失败分析）
	EntryMACD     float64 `json:"entry_macd"`      // 开仓时MACD
	EntryRSI      float64 `json:"entry_rsi"`       // 开仓时RSI
	EntryVolRatio float64 `json:"entry_vol_ratio"` // 开仓时成交量比率
	EntryReason   string  `json:"entry_reason"`    // 开仓依据
	
	// 新增：失败原因分析
	ExitReason    string  `json:"exit_reason"`     // 退出原因: "止损" / "止盈" / "手动平仓"
	IsPremature   bool    `json:"is_premature"`    // 是否过早平仓（<30分钟）
	FailureType   string  `json:"failure_type"`    // 失败类型（如果亏损）
}

// PerformanceAnalysis 交易表现分析
type PerformanceAnalysis struct {
	TotalTrades   int                           `json:"total_trades"`   // 总交易数
	WinningTrades int                           `json:"winning_trades"` // 盈利交易数
	LosingTrades  int                           `json:"losing_trades"`  // 亏损交易数
	WinRate       float64                       `json:"win_rate"`       // 胜率
	AvgWin        float64                       `json:"avg_win"`        // 平均盈利
	AvgLoss       float64                       `json:"avg_loss"`       // 平均亏损
	ProfitFactor  float64                       `json:"profit_factor"`  // 盈亏比
	SharpeRatio   float64                       `json:"sharpe_ratio"`   // 夏普比率（风险调整后收益）
	// 新增：多空统计
	LongTrades    int     `json:"long_trades"`     // 做多交易数
	ShortTrades   int     `json:"short_trades"`    // 做空交易数
	LongWinRate   float64 `json:"long_win_rate"`   // 做多胜率
	ShortWinRate  float64 `json:"short_win_rate"`  // 做空胜率
	LongAvgPnL    float64 `json:"long_avg_pnl"`    // 做多平均盈亏
	ShortAvgPnL   float64 `json:"short_avg_pnl"`   // 做空平均盈亏
	RecentTrades  []TradeOutcome                `json:"recent_trades"`  // 最近N笔交易
	SymbolStats   map[string]*SymbolPerformance `json:"symbol_stats"`   // 各币种表现
	BestSymbol    string                        `json:"best_symbol"`    // 表现最好的币种
	WorstSymbol   string                        `json:"worst_symbol"`   // 表现最差的币种
}

// SymbolPerformance 币种表现统计
type SymbolPerformance struct {
	Symbol        string  `json:"symbol"`         // 币种
	TotalTrades   int     `json:"total_trades"`   // 交易次数
	WinningTrades int     `json:"winning_trades"` // 盈利次数
	LosingTrades  int     `json:"losing_trades"`  // 亏损次数
	WinRate       float64 `json:"win_rate"`       // 胜率
	TotalPnL      float64 `json:"total_pn_l"`     // 总盈亏
	AvgPnL        float64 `json:"avg_pn_l"`       // 平均盈亏
}

// AnalyzePerformance 分析最近N个周期的交易表现（从数据库）
func (l *DecisionLogger) AnalyzePerformance(lookbackCycles int) (*PerformanceAnalysis, error) {
	if l.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	return l.analyzePerformanceFromDB(lookbackCycles)
}



// analyzePerformanceFromDB 从数据库分析交易表现
func (l *DecisionLogger) analyzePerformanceFromDB(lookbackCycles int) (*PerformanceAnalysis, error) {
	analysis := &PerformanceAnalysis{
		RecentTrades: []TradeOutcome{},
		SymbolStats:  make(map[string]*SymbolPerformance),
	}

	// 优先从 trade_outcomes 表读取（如果有数据）
	dbTrades, err := l.db.GetTradeOutcomes(lookbackCycles * 10)
	if err != nil {
		return nil, fmt.Errorf("从数据库读取交易记录失败: %w", err)
	}

	// 如果 trade_outcomes 表为空，尝试从 decision_actions 表分析生成
	if len(dbTrades) == 0 {
		log.Printf("📊 trade_outcomes表为空，尝试从decision_actions分析...")
		return l.analyzeFromDecisionActions(lookbackCycles)
	}

	// 转换数据库记录为分析格式
	for _, dbTrade := range dbTrades {
		trade := TradeOutcome{
			Symbol:          dbTrade.Symbol,
			Side:            dbTrade.Side,
			Quantity:        dbTrade.Quantity,
			Leverage:        dbTrade.Leverage,
			OpenPrice:       dbTrade.OpenPrice,
			ClosePrice:      dbTrade.ClosePrice,
			PositionValue:   dbTrade.PositionValue,
			MarginUsed:      dbTrade.MarginUsed,
			PnL:             dbTrade.PnL,
			PnLPct:          dbTrade.PnLPct,
			Duration:        fmt.Sprintf("%d分钟", dbTrade.DurationMinutes),
			DurationMinutes: dbTrade.DurationMinutes,
			OpenTime:        dbTrade.OpenTime,
			CloseTime:       dbTrade.CloseTime,
			WasStopLoss:     dbTrade.WasStopLoss,
			EntryReason:     dbTrade.EntryReason,
			ExitReason:      dbTrade.ExitReason,
			IsPremature:     dbTrade.IsPremature,
			FailureType:     dbTrade.FailureType,
		}

		analysis.RecentTrades = append(analysis.RecentTrades, trade)
		analysis.TotalTrades++

		// 统计盈亏
		if trade.PnL > 0 {
			analysis.WinningTrades++
			analysis.AvgWin += trade.PnL
		} else if trade.PnL < 0 {
			analysis.LosingTrades++
			analysis.AvgLoss += trade.PnL
		}

		// 多空统计
		if trade.Side == "long" {
			analysis.LongTrades++
			analysis.LongAvgPnL += trade.PnL
			if trade.PnL > 0 {
				analysis.LongWinRate++
			}
		} else if trade.Side == "short" {
			analysis.ShortTrades++
			analysis.ShortAvgPnL += trade.PnL
			if trade.PnL > 0 {
				analysis.ShortWinRate++
			}
		}

		// 币种统计
		if _, exists := analysis.SymbolStats[trade.Symbol]; !exists {
			analysis.SymbolStats[trade.Symbol] = &SymbolPerformance{
				Symbol: trade.Symbol,
			}
		}
		stats := analysis.SymbolStats[trade.Symbol]
		stats.TotalTrades++
		stats.TotalPnL += trade.PnL
		if trade.PnL > 0 {
			stats.WinningTrades++
		} else if trade.PnL < 0 {
			stats.LosingTrades++
		}
	}

	// 计算统计指标
	if analysis.TotalTrades > 0 {
		analysis.WinRate = (float64(analysis.WinningTrades) / float64(analysis.TotalTrades)) * 100

		totalWinAmount := analysis.AvgWin
		totalLossAmount := analysis.AvgLoss

		if analysis.WinningTrades > 0 {
			analysis.AvgWin /= float64(analysis.WinningTrades)
		}
		if analysis.LosingTrades > 0 {
			analysis.AvgLoss /= float64(analysis.LosingTrades)
		}

		if totalLossAmount != 0 {
			analysis.ProfitFactor = totalWinAmount / (-totalLossAmount)
		} else if totalWinAmount > 0 {
			analysis.ProfitFactor = 999.0
		}
	}

	// 计算多空胜率
	if analysis.LongTrades > 0 {
		analysis.LongWinRate = (analysis.LongWinRate / float64(analysis.LongTrades)) * 100
		analysis.LongAvgPnL /= float64(analysis.LongTrades)
	}
	if analysis.ShortTrades > 0 {
		analysis.ShortWinRate = (analysis.ShortWinRate / float64(analysis.ShortTrades)) * 100
		analysis.ShortAvgPnL /= float64(analysis.ShortTrades)
	}

	// 计算各币种胜率和平均盈亏
	bestPnL := -999999.0
	worstPnL := 999999.0
	for symbol, stats := range analysis.SymbolStats {
		if stats.TotalTrades > 0 {
			stats.WinRate = (float64(stats.WinningTrades) / float64(stats.TotalTrades)) * 100
			stats.AvgPnL = stats.TotalPnL / float64(stats.TotalTrades)

			if stats.TotalPnL > bestPnL {
				bestPnL = stats.TotalPnL
				analysis.BestSymbol = symbol
			}
			if stats.TotalPnL < worstPnL {
				worstPnL = stats.TotalPnL
				analysis.WorstSymbol = symbol
			}
		}
	}

	// 只保留最近10笔交易（数据库已DESC排序，前10条就是最新的）
	if len(analysis.RecentTrades) > 10 {
		analysis.RecentTrades = analysis.RecentTrades[:10]
	}
	
	// 确保最新的在最前面（虽然数据库已经DESC，但这里再确认一下）
	log.Printf("✓ 返回最近%d笔交易（最新ID: %d）", len(analysis.RecentTrades), func() int64 {
		if len(dbTrades) > 0 {
			return dbTrades[0].ID
		}
		return 0
	}())

	// 从数据库获取最近的决策记录，计算夏普比率
	records, err := l.db.GetLatestRecords(lookbackCycles)
	if err == nil && len(records) > 0 {
		analysis.SharpeRatio = l.calculateSharpeRatioFromDB(records)
	}

	return analysis, nil
}

// calculateSharpeRatioFromDB 从数据库记录计算夏普比率
func (l *DecisionLogger) calculateSharpeRatioFromDB(records []*database.DecisionRecord) float64 {
	if len(records) < 2 {
		return 0.0
	}

	var equities []float64
	for _, record := range records {
		if record.TotalBalance > 0 {
			equities = append(equities, record.TotalBalance)
		}
	}

	if len(equities) < 2 {
		return 0.0
	}

	// 计算周期收益率
	var returns []float64
	for i := 1; i < len(equities); i++ {
		if equities[i-1] > 0 {
			periodReturn := (equities[i] - equities[i-1]) / equities[i-1]
			returns = append(returns, periodReturn)
		}
	}

	if len(returns) == 0 {
		return 0.0
	}

	// 计算平均收益率
	sumReturns := 0.0
	for _, r := range returns {
		sumReturns += r
	}
	meanReturn := sumReturns / float64(len(returns))

	// 计算标准差
	sumSquaredDiff := 0.0
	for _, r := range returns {
		diff := r - meanReturn
		sumSquaredDiff += diff * diff
	}
	variance := sumSquaredDiff / float64(len(returns))
	stdDev := math.Sqrt(variance)

	if stdDev == 0 {
		if meanReturn > 0 {
			return 999.0
		} else if meanReturn < 0 {
			return -999.0
		}
		return 0.0
	}

	return meanReturn / stdDev
}

// analyzeFromDecisionActions 从 decision_actions 表分析并生成交易记录
func (l *DecisionLogger) analyzeFromDecisionActions(lookbackCycles int) (*PerformanceAnalysis, error) {
	analysis := &PerformanceAnalysis{
		RecentTrades: []TradeOutcome{},
		SymbolStats:  make(map[string]*SymbolPerformance),
	}

	// 获取最近的决策记录
	records, err := l.db.GetLatestRecords(lookbackCycles * 3) // 扩大窗口确保捕获完整交易
	if err != nil {
		return nil, fmt.Errorf("读取决策记录失败: %w", err)
	}

	if len(records) == 0 {
		return analysis, nil
	}

	// 追踪开仓状态：symbol_side -> 开仓信息
	type OpenPosition struct {
		Side      string
		OpenPrice float64
		OpenTime  time.Time
		Quantity  float64
		Leverage  int
	}
	openPositions := make(map[string]*OpenPosition)

	// 遍历所有决策记录，匹配开仓-平仓
	for _, record := range records {
		// 从数据库获取该记录的所有动作
		actions, err := l.getActionsForRecord(record.ID)
		if err != nil {
			continue
		}

		for _, action := range actions {
			if !action.Success {
				continue
			}

			symbol := action.Symbol
			var side string
			if action.Action == "open_long" || action.Action == "close_long" {
				side = "long"
			} else if action.Action == "open_short" || action.Action == "close_short" {
				side = "short"
			} else {
				continue
			}
			posKey := symbol + "_" + side

			switch action.Action {
			case "open_long", "open_short":
				// 记录开仓
				openPositions[posKey] = &OpenPosition{
					Side:      side,
					OpenPrice: action.Price,
					OpenTime:  action.Timestamp,
					Quantity:  action.Quantity,
					Leverage:  action.Leverage,
				}

			case "close_long", "close_short":
				// 查找对应的开仓记录
				if openPos, exists := openPositions[posKey]; exists {
					// 计算盈亏
					var pnl float64
					if side == "long" {
						pnl = openPos.Quantity * (action.Price - openPos.OpenPrice)
					} else {
						pnl = openPos.Quantity * (openPos.OpenPrice - action.Price)
					}

					// 计算盈亏百分比
					positionValue := openPos.Quantity * openPos.OpenPrice
					marginUsed := positionValue / float64(openPos.Leverage)
					pnlPct := 0.0
					if marginUsed > 0 {
						pnlPct = (pnl / marginUsed) * 100
					}

					// 计算持仓时长
					durationMinutes := int64(action.Timestamp.Sub(openPos.OpenTime).Minutes())
					isPremature := durationMinutes < 30

					// 判断退出原因
					exitReason := "平仓"
					if action.WasStopLoss {
						exitReason = "止损/止盈"
					} else if pnl > 0 {
						exitReason = "主动止盈"
					}

					// 失败原因
					failureType := ""
					if pnl < 0 {
						if isPremature {
							failureType = "过早平仓（<30分钟）"
						} else {
							failureType = "信号判断错误"
						}
					}

					// 创建交易结果
					outcome := TradeOutcome{
						Symbol:          symbol,
						Side:            side,
						Quantity:        openPos.Quantity,
						Leverage:        openPos.Leverage,
						OpenPrice:       openPos.OpenPrice,
						ClosePrice:      action.Price,
						PositionValue:   positionValue,
						MarginUsed:      marginUsed,
						PnL:             pnl,
						PnLPct:          pnlPct,
						Duration:        fmt.Sprintf("%d分钟", durationMinutes),
						DurationMinutes: durationMinutes,
						OpenTime:        openPos.OpenTime,
						CloseTime:       action.Timestamp,
						WasStopLoss:     action.WasStopLoss,
						EntryReason:     "历史交易",
						ExitReason:      exitReason,
						IsPremature:     isPremature,
						FailureType:     failureType,
					}

					analysis.RecentTrades = append(analysis.RecentTrades, outcome)
					analysis.TotalTrades++

					// 统计盈亏
					if pnl > 0 {
						analysis.WinningTrades++
						analysis.AvgWin += pnl
					} else if pnl < 0 {
						analysis.LosingTrades++
						analysis.AvgLoss += pnl
					}

					// 多空统计
					if side == "long" {
						analysis.LongTrades++
						analysis.LongAvgPnL += pnl
						if pnl > 0 {
							analysis.LongWinRate++
						}
					} else {
						analysis.ShortTrades++
						analysis.ShortAvgPnL += pnl
						if pnl > 0 {
							analysis.ShortWinRate++
						}
					}

					// 币种统计
					if _, exists := analysis.SymbolStats[symbol]; !exists {
						analysis.SymbolStats[symbol] = &SymbolPerformance{Symbol: symbol}
					}
					stats := analysis.SymbolStats[symbol]
					stats.TotalTrades++
					stats.TotalPnL += pnl
					if pnl > 0 {
						stats.WinningTrades++
					} else if pnl < 0 {
						stats.LosingTrades++
					}

					// 保存到数据库（供下次直接读取）
					l.SaveTradeOutcome(&outcome)

					// 移除已平仓记录
					delete(openPositions, posKey)
				}
			}
		}
	}

	// 计算统计指标
	if analysis.TotalTrades > 0 {
		analysis.WinRate = (float64(analysis.WinningTrades) / float64(analysis.TotalTrades)) * 100

		totalWinAmount := analysis.AvgWin
		totalLossAmount := analysis.AvgLoss

		if analysis.WinningTrades > 0 {
			analysis.AvgWin /= float64(analysis.WinningTrades)
		}
		if analysis.LosingTrades > 0 {
			analysis.AvgLoss /= float64(analysis.LosingTrades)
		}

		if totalLossAmount != 0 {
			analysis.ProfitFactor = totalWinAmount / (-totalLossAmount)
		} else if totalWinAmount > 0 {
			analysis.ProfitFactor = 999.0
		}
	}

	// 计算多空胜率
	if analysis.LongTrades > 0 {
		analysis.LongWinRate = (analysis.LongWinRate / float64(analysis.LongTrades)) * 100
		analysis.LongAvgPnL /= float64(analysis.LongTrades)
	}
	if analysis.ShortTrades > 0 {
		analysis.ShortWinRate = (analysis.ShortWinRate / float64(analysis.ShortTrades)) * 100
		analysis.ShortAvgPnL /= float64(analysis.ShortTrades)
	}

	// 计算币种统计
	bestPnL := -999999.0
	worstPnL := 999999.0
	for symbol, stats := range analysis.SymbolStats {
		if stats.TotalTrades > 0 {
			stats.WinRate = (float64(stats.WinningTrades) / float64(stats.TotalTrades)) * 100
			stats.AvgPnL = stats.TotalPnL / float64(stats.TotalTrades)

			if stats.TotalPnL > bestPnL {
				bestPnL = stats.TotalPnL
				analysis.BestSymbol = symbol
			}
			if stats.TotalPnL < worstPnL {
				worstPnL = stats.TotalPnL
				analysis.WorstSymbol = symbol
			}
		}
	}

	// 只保留最近10笔
	if len(analysis.RecentTrades) > 10 {
		analysis.RecentTrades = analysis.RecentTrades[len(analysis.RecentTrades)-10:]
	}

	// 计算夏普比率
	if len(records) > 0 {
		analysis.SharpeRatio = l.calculateSharpeRatioFromDB(records)
	}

	log.Printf("✓ 从decision_actions分析出 %d 笔完整交易", analysis.TotalTrades)
	return analysis, nil
}

// getActionsForRecord 获取指定记录的所有决策动作
func (l *DecisionLogger) getActionsForRecord(recordID int64) ([]*database.DecisionAction, error) {
	if l.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	return l.db.QueryActions(recordID)
}

// SaveTradeOutcome 保存交易结果到数据库
func (l *DecisionLogger) SaveTradeOutcome(trade *TradeOutcome) error {
	if l.db == nil {
		return nil // 数据库不可用，跳过
	}

	dbTrade := &database.TradeOutcome{
		TraderID:        l.traderID,
		Symbol:          trade.Symbol,
		Side:            trade.Side,
		Quantity:        trade.Quantity,
		Leverage:        trade.Leverage,
		OpenPrice:       trade.OpenPrice,
		ClosePrice:      trade.ClosePrice,
		PositionValue:   trade.PositionValue,
		MarginUsed:      trade.MarginUsed,
		PnL:             trade.PnL,
		PnLPct:          trade.PnLPct,
		DurationMinutes: trade.DurationMinutes,
		OpenTime:        trade.OpenTime,
		CloseTime:       trade.CloseTime,
		WasStopLoss:     trade.WasStopLoss,
		EntryReason:     trade.EntryReason,
		ExitReason:      trade.ExitReason,
		IsPremature:     trade.IsPremature,
		FailureType:     trade.FailureType,
	}

	return l.db.InsertTradeOutcome(dbTrade)
}

// calculateSharpeRatio 计算夏普比率
// 基于账户净值的变化计算风险调整后收益
func (l *DecisionLogger) calculateSharpeRatio(records []*DecisionRecord) float64 {
	if len(records) < 2 {
		return 0.0
	}

	// 提取每个周期的账户净值
	// 注意：TotalBalance字段实际存储的是TotalEquity（账户总净值）
	// TotalUnrealizedProfit字段实际存储的是TotalPnL（相对初始余额的盈亏）
	var equities []float64
	for _, record := range records {
		// 直接使用TotalBalance，因为它已经是完整的账户净值
		equity := record.AccountState.TotalBalance
		if equity > 0 {
			equities = append(equities, equity)
		}
	}

	if len(equities) < 2 {
		return 0.0
	}

	// 计算周期收益率（period returns）
	var returns []float64
	for i := 1; i < len(equities); i++ {
		if equities[i-1] > 0 {
			periodReturn := (equities[i] - equities[i-1]) / equities[i-1]
			returns = append(returns, periodReturn)
		}
	}

	if len(returns) == 0 {
		return 0.0
	}

	// 计算平均收益率
	sumReturns := 0.0
	for _, r := range returns {
		sumReturns += r
	}
	meanReturn := sumReturns / float64(len(returns))

	// 计算收益率标准差
	sumSquaredDiff := 0.0
	for _, r := range returns {
		diff := r - meanReturn
		sumSquaredDiff += diff * diff
	}
	variance := sumSquaredDiff / float64(len(returns))
	stdDev := math.Sqrt(variance)

	// 避免除以零
	if stdDev == 0 {
		if meanReturn > 0 {
			return 999.0 // 无波动的正收益
		} else if meanReturn < 0 {
			return -999.0 // 无波动的负收益
		}
		return 0.0
	}

	// 计算夏普比率（假设无风险利率为0）
	// 注：直接返回周期级别的夏普比率（非年化），正常范围 -2 到 +2
	sharpeRatio := meanReturn / stdDev
	return sharpeRatio
}
