package decision

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"nofx/database"
	"nofx/database/models"
	"nofx/market"
	"nofx/mcp"
	"nofx/pool"
	"sort"
	"strings"
	"time"
)

// PositionInfo 持仓信息
type PositionInfo struct {
	Symbol           string  `json:"symbol"`
	Side             string  `json:"side"` // "long" or "short"
	EntryPrice       float64 `json:"entry_price"`
	MarkPrice        float64 `json:"mark_price"`
	Quantity         float64 `json:"quantity"`
	Leverage         int     `json:"leverage"`
	UnrealizedPnL    float64 `json:"unrealized_pnl"`
	UnrealizedPnLPct float64 `json:"unrealized_pnl_pct"`
	LiquidationPrice float64 `json:"liquidation_price"`
	MarginUsed       float64 `json:"margin_used"`
	UpdateTime       int64   `json:"update_time"` // 持仓更新时间戳（毫秒）
}

// AccountInfo 账户信息
type AccountInfo struct {
	TotalEquity      float64 `json:"total_equity"`      // 账户净值
	AvailableBalance float64 `json:"available_balance"` // 可用余额
	TotalPnL         float64 `json:"total_pnl"`         // 总盈亏
	TotalPnLPct      float64 `json:"total_pnl_pct"`     // 总盈亏百分比
	MarginUsed       float64 `json:"margin_used"`       // 已用保证金
	MarginUsedPct    float64 `json:"margin_used_pct"`   // 保证金使用率
	PositionCount    int     `json:"position_count"`    // 持仓数量
	// 风险管理相关字段
	RiskCapacityUSD  float64 `json:"risk_capacity_usd"`  // 剩余风险容量（USD）
	MaxRiskPerTrade  float64 `json:"max_risk_per_trade"` // 单笔最大风险（USD）
	DailyRiskBudget  float64 `json:"daily_risk_budget"`  // 日风险预算（USD）
	UsedRiskBudget   float64 `json:"used_risk_budget"`   // 已使用风险预算（USD）
}

// CandidateCoin 候选币种（来自币种池）
type CandidateCoin struct {
	Symbol  string   `json:"symbol"`
	Sources []string `json:"sources"` // 来源: "ai500" 和/或 "oi_top"
}

// OITopData 持仓量增长Top数据（用于AI决策参考）
type OITopData struct {
	Rank              int     // OI Top排名
	OIDeltaPercent    float64 // 持仓量变化百分比（1小时）
	OIDeltaValue      float64 // 持仓量变化价值
	PriceDeltaPercent float64 // 价格变化百分比
	NetLong           float64 // 净多仓
	NetShort          float64 // 净空仓
}

// RiskMetrics 风险管理指标
type RiskMetrics struct {
	VaR95             float64 `json:"var_95"`              // 95%置信度风险价值（USD）
	VaR99             float64 `json:"var_99"`              // 99%置信度风险价值（USD）
	MaxDrawdown       float64 `json:"max_drawdown"`        // 最大回撤（%）
	MaxDrawdownUSD    float64 `json:"max_drawdown_usd"`    // 最大回撤（USD）
	SharpeRatio       float64 `json:"sharpe_ratio"`        // 夏普比率
	TotalRiskExposure float64 `json:"total_risk_exposure"` // 总风险敞口（USD）
	LeverageRisk      float64 `json:"leverage_risk"`       // 杠杆风险评分（0-100）
	ConcentrationRisk float64 `json:"concentration_risk"`  // 集中度风险评分（0-100）
	LiquidationRisk   float64 `json:"liquidation_risk"`    // 强平风险评分（0-100）
	VolatilityRisk    float64 `json:"volatility_risk"`     // 波动率风险评分（0-100）
}

// Context 交易上下文（传递给AI的完整信息）
type Context struct {
	CurrentTime       string                  `json:"current_time"`
	RuntimeMinutes    int                     `json:"runtime_minutes"`
	CallCount         int                     `json:"call_count"`
	Account           AccountInfo             `json:"account"`
	Positions         []PositionInfo          `json:"positions"`
	CandidateCoins    []CandidateCoin         `json:"candidate_coins"`
	RiskMetrics       RiskMetrics             `json:"risk_metrics"`       // 风险管理指标
	MarketDataMap     map[string]*market.Data `json:"-"` // 不序列化，但内部使用
	OITopDataMap      map[string]*OITopData   `json:"-"` // OI Top数据映射
	Performance       interface{}             `json:"-"` // 历史表现分析（logger.PerformanceAnalysis）
	BTCETHLeverage    int                     `json:"-"` // BTC/ETH杠杆倍数（从配置读取）
	AltcoinLeverage   int                     `json:"-"` // 山寨币杠杆倍数（从配置读取）
	MaxPositions      int                     `json:"-"` // 最大持仓数限制（从配置读取）
	AILearningSummary string                  `json:"-"` // AI学习总结（从数据库加载）
	DecisionLogger    interface{ GetDB() *database.DB } `json:"-"` // 决策日志记录器（用于获取数据库连接）
	AIAutonomyMode    bool                    `json:"-"` // AI自主模式（true=完全自主，false=限制模式）
}

// Decision AI的交易决策
type Decision struct {
	Symbol          string  `json:"symbol"`
	Action          string  `json:"action"` // "open_long", "open_short", "close_long", "close_short", "hold", "wait"
	Leverage        int     `json:"leverage,omitempty"`
	PositionSizeUSD float64 `json:"position_size_usd,omitempty"`
	StopLoss        float64 `json:"stop_loss,omitempty"`
	TakeProfit      float64 `json:"take_profit,omitempty"`
	Confidence      int     `json:"confidence,omitempty"` // 信心度 (0-100)
	RiskUSD         float64 `json:"risk_usd,omitempty"`   // 最大美元风险
	Reasoning       string  `json:"reasoning"`
}

// FullDecision AI的完整决策（包含思维链）
type FullDecision struct {
	SystemPrompt string     `json:"system_prompt"` // System Prompt（规则，从数据库加载）
	UserPrompt   string     `json:"user_prompt"`   // User Prompt（市场数据）
	CoTTrace     string     `json:"cot_trace"`     // 思维链分析（AI输出）
	Decisions    []Decision `json:"decisions"`     // 具体决策列表
	Timestamp    time.Time  `json:"timestamp"`
}

// GetFullDecision 获取AI的完整交易决策（批量分析所有币种和持仓）
func GetFullDecision(ctx *Context, mcpClient *mcp.Client) (*FullDecision, error) {
	// 1. 为所有币种获取市场数据
	if err := fetchMarketDataForContext(ctx); err != nil {
		return nil, fmt.Errorf("获取市场数据失败: %w", err)
	}

	// 2. 计算智能风控参数和实际仓位限制
	smartRisk := CalculateSmartRiskParams(ctx)
	
	// 计算实际最大仓位（与验证逻辑完全一致）
	baseMaxBTC := ctx.Account.TotalEquity * 30.0
	baseMaxAlt := ctx.Account.TotalEquity * 20.0
	actualMaxBTC := CalculateSmartPositionSize(baseMaxBTC, smartRisk, "BTCUSDT", 85)
	actualMaxAlt := CalculateSmartPositionSize(baseMaxAlt, smartRisk, "OTHER", 85)
	
	// 3. 构建 System Prompt（从数据库加载）和 User Prompt（动态数据）
	db := ctx.DecisionLogger.GetDB()
	if db == nil {
		return nil, fmt.Errorf("数据库连接不可用，无法构建提示词")
	}
	
	systemPrompt := db.BuildSystemPromptFromDB(ctx.Account.TotalEquity, ctx.BTCETHLeverage, ctx.AltcoinLeverage, actualMaxBTC, actualMaxAlt, ctx.AIAutonomyMode)
	userPrompt, err := buildUserPrompt(ctx)
	if err != nil {
		return nil, fmt.Errorf("构建用户提示词失败: %w", err)
	}
	
	log.Printf("[Prompt] 实际仓位限制: BTC=%.0f USDT, 其他=%.0f USDT (账户净值%.2f, 盈亏%.1f%%, 保证金%.1f%%)", 
		actualMaxBTC, actualMaxAlt, ctx.Account.TotalEquity, smartRisk.TotalPnLPct, smartRisk.MarginUsedPct)

	// 4. 调用AI API（使用 system + user prompt）
	aiResponse, err := mcpClient.CallWithMessages(systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("调用AI API失败: %w", err)
	}

	// 4. 解析AI响应
	decision, err := parseFullDecisionResponse(aiResponse, ctx.Account.TotalEquity, ctx.BTCETHLeverage, ctx.AltcoinLeverage)
	if err != nil {
		return nil, fmt.Errorf("解析AI响应失败: %w", err)
	}
	
	// 4.5 使用真实ctx验证决策（确保使用正确的AIAutonomyMode）
	if err := validateDecisions(decision.Decisions, ctx); err != nil {
		return nil, fmt.Errorf("决策验证失败: %w", err)
	}

	// 5. 智能市场分析
	marketAnalyzer := NewSmartMarketAnalyzer(ctx)
	marketCondition := marketAnalyzer.AnalyzeMarketCondition()

	// 6. 决策质量评估
	qualityAnalyzer := NewDecisionQualityAnalyzer(ctx, marketCondition)
	
	// 为每个决策评估质量并记录
	for i := range decision.Decisions {
		quality := qualityAnalyzer.EvaluateDecisionQuality(&decision.Decisions[i])
		
		// 记录决策质量信息
		log.Printf("决策 %d 质量评估: 分数=%.1f, 等级=%s", i+1, quality.Score, quality.Grade)
		if len(quality.Issues) > 0 {
			log.Printf("决策 %d 风险提示: %v", i+1, quality.Issues)
		}
		
		// 如果决策质量过低，降低信心度
		if quality.Grade == "poor" {
			if decision.Decisions[i].Confidence > 30 {
				decision.Decisions[i].Confidence = 30
			}
			log.Printf("决策 %d 质量较差，信心度调整为 %d", i+1, decision.Decisions[i].Confidence)
		} else if quality.Grade == "fair" {
			if decision.Decisions[i].Confidence > 60 {
				decision.Decisions[i].Confidence = 60
			}
			log.Printf("决策 %d 质量一般，信心度调整为 %d", i+1, decision.Decisions[i].Confidence)
		}
	}

	// 记录市场状况
	log.Printf("市场状况分析: 趋势=%s, 波动率=%s, 情绪=%s, 风险=%s", 
		marketCondition.Trend, marketCondition.Volatility, 
		marketCondition.Sentiment, marketCondition.Risk)

	decision.Timestamp = time.Now()
	decision.SystemPrompt = systemPrompt // 保存system prompt
	decision.UserPrompt = userPrompt     // 保存user prompt
	return decision, nil
}

// fetchMarketDataForContext 为上下文中的所有币种获取市场数据和OI数据
func fetchMarketDataForContext(ctx *Context) error {
	ctx.MarketDataMap = make(map[string]*market.Data)
	ctx.OITopDataMap = make(map[string]*OITopData)

	// 收集所有需要获取数据的币种
	symbolSet := make(map[string]bool)

	// 1. 优先获取持仓币种的数据（这是必须的）
	for _, pos := range ctx.Positions {
		symbolSet[pos.Symbol] = true
	}

	// 2. 候选币种数量根据账户状态动态调整
	maxCandidates := calculateMaxCandidates(ctx)
	for i, coin := range ctx.CandidateCoins {
		if i >= maxCandidates {
			break
		}
		symbolSet[coin.Symbol] = true
	}

	// 并发获取市场数据
	// 持仓币种集合（用于判断是否跳过OI检查）
	positionSymbols := make(map[string]bool)
	for _, pos := range ctx.Positions {
		positionSymbols[pos.Symbol] = true
	}

	for symbol := range symbolSet {
		data, err := market.Get(symbol)
		if err != nil {
			// 单个币种失败不影响整体，只记录错误
			continue
		}

		// ⚠️ 流动性过滤：持仓价值低于15M USD的币种不做（多空都不做）
		// 持仓价值 = 持仓量 × 当前价格
		// 但现有持仓必须保留（需要决策是否平仓）
		isExistingPosition := positionSymbols[symbol]
		if !isExistingPosition && data.OpenInterest != nil && data.CurrentPrice > 0 {
			// 计算持仓价值（USD）= 持仓量 × 当前价格
			oiValue := data.OpenInterest.Latest * data.CurrentPrice
			oiValueInMillions := oiValue / 1_000_000 // 转换为百万美元单位
			if oiValueInMillions < 15 {
				log.Printf("⚠️  %s 持仓价值过低(%.2fM USD < 15M)，跳过此币种 [持仓量:%.0f × 价格:%.4f]",
					symbol, oiValueInMillions, data.OpenInterest.Latest, data.CurrentPrice)
				continue
			}
		}

		ctx.MarketDataMap[symbol] = data
	}

	// 加载OI Top数据（不影响主流程）
	oiPositions, err := pool.GetOITopPositions()
	if err == nil {
		for _, pos := range oiPositions {
			// 标准化符号匹配
			symbol := pos.Symbol
			ctx.OITopDataMap[symbol] = &OITopData{
				Rank:              pos.Rank,
				OIDeltaPercent:    pos.OIDeltaPercent,
				OIDeltaValue:      pos.OIDeltaValue,
				PriceDeltaPercent: pos.PriceDeltaPercent,
				NetLong:           pos.NetLong,
				NetShort:          pos.NetShort,
			}
		}
	}

	return nil
}

// calculateMaxCandidates 根据账户状态计算需要分析的候选币种数量
func calculateMaxCandidates(ctx *Context) int {
	// 直接返回候选池的全部币种数量
	// 因为候选池已经在 auto_trader.go 中筛选过了
	// 固定分析前20个评分最高的币种（来自AI500）
	return len(ctx.CandidateCoins)
}



// buildUserPrompt 构建 User Prompt（动态数据）
func buildUserPrompt(ctx *Context) (string, error) {
	// 获取数据库连接
	var db *database.DB
	if ctx.DecisionLogger != nil {
		db = ctx.DecisionLogger.GetDB()
	}
	
	if db == nil {
		return "", fmt.Errorf("数据库连接不可用，无法构建用户提示词")
	}
	
	// 从数据库获取用户提示词模板
	templates, err := db.GetUserPromptTemplates()
	if err != nil {
		return "", fmt.Errorf("获取用户提示词模板失败: %w", err)
	}
	
	var sb strings.Builder
	
	// 准备模板数据
	templateData := buildTemplateData(ctx)
	
	// 按照display_order顺序处理模板
	for _, tmpl := range templates {
		content := renderTemplate(tmpl.Content, templateData, ctx)
		if content != "" {
			sb.WriteString(content)
			sb.WriteString("\n\n")
		}
	}
	
	return sb.String(), nil
}

// buildTemplateData 构建模板数据
func buildTemplateData(ctx *Context) map[string]interface{} {
	data := make(map[string]interface{})
	
	// 基础数据
	data["Time"] = ctx.CurrentTime
	data["CycleNumber"] = ctx.CallCount
	data["RuntimeMinutes"] = ctx.RuntimeMinutes
	data["CandidateCount"] = len(ctx.MarketDataMap)
	data["PositionCount"] = ctx.Account.PositionCount
	
	// BTC数据
	if btcData, hasBTC := ctx.MarketDataMap["BTCUSDT"]; hasBTC {
		data["BTCPrice"] = fmt.Sprintf("%.2f", btcData.CurrentPrice)
		data["BTC1hChange"] = fmt.Sprintf("%+.2f", btcData.PriceChange1h)
		data["BTC4hChange"] = fmt.Sprintf("%+.2f", btcData.PriceChange4h)
		data["BTCMACD"] = fmt.Sprintf("%.4f", btcData.CurrentMACD)
		data["BTCRSI"] = fmt.Sprintf("%.2f", btcData.CurrentRSI7)
	}
	
	// 账户数据
	data["NetValue"] = fmt.Sprintf("%.2f", ctx.Account.TotalEquity)
	data["Balance"] = fmt.Sprintf("%.2f", ctx.Account.AvailableBalance)
	data["BalancePercent"] = fmt.Sprintf("%.1f", (ctx.Account.AvailableBalance/ctx.Account.TotalEquity)*100)
	data["PnLPercent"] = fmt.Sprintf("%+.2f", ctx.Account.TotalPnLPct)
	data["MarginPercent"] = fmt.Sprintf("%.1f", ctx.Account.MarginUsedPct)
	
	// 夏普比率
	if ctx.Performance != nil {
		type PerformanceData struct {
			SharpeRatio float64 `json:"sharpe_ratio"`
		}
		var perfData PerformanceData
		if jsonData, err := json.Marshal(ctx.Performance); err == nil {
			if err := json.Unmarshal(jsonData, &perfData); err == nil {
				data["SharpeRatio"] = fmt.Sprintf("%.2f", perfData.SharpeRatio)
			}
		}
	}
	
	return data
}

// renderTemplate 渲染模板内容
func renderTemplate(template string, data map[string]interface{}, ctx *Context) string {
	content := template
	
	// 简单的字符串替换
	for key, value := range data {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		content = strings.ReplaceAll(content, placeholder, fmt.Sprintf("%v", value))
	}
	
	// 处理特殊的动态内容
	content = renderSpecialContent(content, ctx)
	
	return content
}

// renderSpecialContent 处理特殊的动态内容
func renderSpecialContent(content string, ctx *Context) string {
	// 如果是持仓标题，需要检查是否有持仓
	if strings.Contains(content, "## 当前持仓") && len(ctx.Positions) > 0 {
		// 添加持仓详情
		var positionDetails strings.Builder
		positionDetails.WriteString(content)
		positionDetails.WriteString("\n")
		
		for i, pos := range ctx.Positions {
			// 计算持仓时长
			holdingDuration := ""
			if pos.UpdateTime > 0 {
				durationMs := time.Now().UnixMilli() - pos.UpdateTime
				durationMin := durationMs / (1000 * 60)
				if durationMin < 60 {
					holdingDuration = fmt.Sprintf(" | 持仓时长%d分钟", durationMin)
				} else {
					durationHour := durationMin / 60
					durationMinRemainder := durationMin % 60
					holdingDuration = fmt.Sprintf(" | 持仓时长%d小时%d分钟", durationHour, durationMinRemainder)
				}
			}

			positionDetails.WriteString(fmt.Sprintf("%d. %s %s | 入场价%.4f 当前价%.4f | 盈亏%+.2f%% | 杠杆%dx | 保证金%.0f | 强平价%.4f%s\n\n",
				i+1, pos.Symbol, strings.ToUpper(pos.Side),
				pos.EntryPrice, pos.MarkPrice, pos.UnrealizedPnLPct,
				pos.Leverage, pos.MarginUsed, pos.LiquidationPrice, holdingDuration))

			// 添加市场数据（精简格式）
			if marketData, ok := ctx.MarketDataMap[pos.Symbol]; ok {
				positionDetails.WriteString(market.FormatCompact(marketData))
				positionDetails.WriteString("\n")
			}
		}
		return positionDetails.String()
	}
	
	// 如果是候选币种标题，添加候选币种详情
	if strings.Contains(content, "## 候选币种") {
		var candidateDetails strings.Builder
		candidateDetails.WriteString(content)
		candidateDetails.WriteString("\n\n")
		
		displayedCount := 0
		for _, coin := range ctx.CandidateCoins {
			marketData, hasData := ctx.MarketDataMap[coin.Symbol]
			if !hasData {
				continue
			}
			displayedCount++

			sourceTags := ""
			if len(coin.Sources) > 1 {
				sourceTags = " (AI500+OI_Top双重信号)"
			} else if len(coin.Sources) == 1 && coin.Sources[0] == "oi_top" {
				sourceTags = " (OI_Top持仓增长)"
			}

			candidateDetails.WriteString(fmt.Sprintf("### %d. %s%s\n", displayedCount, coin.Symbol, sourceTags))
			candidateDetails.WriteString(market.FormatCompact(marketData))
			candidateDetails.WriteString("\n")
		}
		return candidateDetails.String()
	}
	
	// 如果是AI学习总结，添加实际内容
	if strings.Contains(content, "## 📚 AI历史交易学习总结") && ctx.AILearningSummary != "" {
		return content + "\n\n" + ctx.AILearningSummary
	}
	
	return content
}



// parseFullDecisionResponse 解析AI的完整决策响应
func parseFullDecisionResponse(aiResponse string, accountEquity float64, btcEthLeverage, altcoinLeverage int) (*FullDecision, error) {
	// 提取思维链
	cotTrace := extractCoTTrace(aiResponse)

	// 提取决策JSON
	decisions, err := extractDecisions(aiResponse)
	if err != nil {
		return nil, fmt.Errorf("提取决策失败: %w", err)
	}

	// 直接返回，不在这里验证（验证在GetFullDecision中用真实ctx进行）
	return &FullDecision{
		CoTTrace:  cotTrace,
		Decisions: decisions,
		Timestamp: time.Now(),
	}, nil
}

// extractCoTTrace 提取思维链分析
func extractCoTTrace(response string) string {
	// 查找JSON数组的开始位置
	jsonStart := strings.Index(response, "[")

	if jsonStart > 0 {
		// 思维链是JSON数组之前的内容
		return strings.TrimSpace(response[:jsonStart])
	}

	// 如果找不到JSON，整个响应都是思维链
	return strings.TrimSpace(response)
}

// extractDecisions 提取JSON决策列表
func extractDecisions(response string) ([]Decision, error) {
	// 直接查找JSON数组 - 找第一个完整的JSON数组
	arrayStart := strings.Index(response, "[")
	if arrayStart == -1 {
		return nil, fmt.Errorf("无法找到JSON数组起始")
	}

	// 从 [ 开始，匹配括号找到对应的 ]
	arrayEnd := findMatchingBracket(response, arrayStart)
	if arrayEnd == -1 {
		return nil, fmt.Errorf("无法找到JSON数组结束")
	}

	jsonContent := strings.TrimSpace(response[arrayStart : arrayEnd+1])

	// 🔧 修复常见的JSON格式错误：缺少引号的字段值
	// 匹配: "reasoning": 内容"}  或  "reasoning": 内容}  (没有引号)
	// 修复为: "reasoning": "内容"}
	// 使用简单的字符串扫描而不是正则表达式
	jsonContent = fixMissingQuotes(jsonContent)

	// 解析JSON
	var decisions []Decision
	if err := json.Unmarshal([]byte(jsonContent), &decisions); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w\nJSON内容: %s", err, jsonContent)
	}

	return decisions, nil
}

// fixMissingQuotes 修复JSON中缺失的引号
func fixMissingQuotes(jsonStr string) string {
	// 修复action字段
	jsonStr = strings.ReplaceAll(jsonStr, `"action": open_long`, `"action": "open_long"`)
	jsonStr = strings.ReplaceAll(jsonStr, `"action": open_short`, `"action": "open_short"`)
	jsonStr = strings.ReplaceAll(jsonStr, `"action": close_long`, `"action": "close_long"`)
	jsonStr = strings.ReplaceAll(jsonStr, `"action": close_short`, `"action": "close_short"`)
	jsonStr = strings.ReplaceAll(jsonStr, `"action": hold`, `"action": "hold"`)
	jsonStr = strings.ReplaceAll(jsonStr, `"action": wait`, `"action": "wait"`)

	// 修复symbol字段（常见币种）
	symbols := []string{"BTCUSDT", "ETHUSDT", "SOLUSDT", "ADAUSDT", "DOTUSDT", "LINKUSDT", "AVAXUSDT", "MATICUSDT", "ATOMUSDT", "NEARUSDT", "FILUSDT", "LTCUSDT", "UNIUSDT", "AAVEUSDT", "SUSHIUSDT", "COMPUSDT", "MKRUSDT", "YFIUSDT", "SNXUSDT", "CRVUSDT", "1INCHUSDT", "ALPHAUSDT", "RENUSDT", "KSMUSDT", "WAVESUSDT", "ICXUSDT", "ONTUSDT", "ZILUSDT", "BATUSDT", "ZRXUSDT", "ENJUSDT", "STORJUSDT", "KNCUSDT", "LRCUSDT", "BANDUSDT", "SANDUSDT", "MANAUSDT", "CHZUSDT", "HOTUSDT", "VETUSDT", "WINUSDT", "DUSKUSDT", "DEFIUSDT", "YFIIUSDT", "AUDIOUSDT", "CTKUSDT", "AKROUSDT", "AXSUSDT", "HARDUSDT", "DNTUSDT", "STRKUSDT", "UNFIUSDT", "ROSEUSDT", "AVAUSDT", "XEMUSDT", "SKLUSDT", "GRTUSDT", "1000SHIBUSDT", "CELOUSDT", "RIFUSDT", "CKBUSDT", "FIROUSDT", "LITUSDT", "SFPUSDT", "DODOUSDT", "CAKEUSDT", "ACMUSDT", "BADGERUSDT", "FISUSDT", "OMUSDT", "PONDUSDT", "DEGOUSDT", "ALICEUSDT", "LINAUSDT", "PERPUSDT", "RAMPUSDT", "SUPERUSDT", "CFXUSDT", "EPSUSDT", "AUTOUSDT", "TKOUSDT", "PUNDIXUSDT", "TLMUSDT", "1000BTTUSDT", "BTCSTUSDT", "TRUUSDT", "DEXEUSDT", "CKBUSDT", "TWTUSDT", "FIROUSDT", "LITUSDT", "SFPUSDT", "DODOUSDT", "CAKEUSDT", "ACMUSDT", "BADGERUSDT", "FISUSDT", "OMUSDT", "PONDUSDT", "DEGOUSDT", "ALICEUSDT", "LINAUSDT", "PERPUSDT", "RAMPUSDT", "SUPERUSDT", "CFXUSDT", "EPSUSDT", "AUTOUSDT", "TKOUSDT", "PUNDIXUSDT", "TLMUSDT", "BTCSTUSDT", "TRUUSDT", "DEXEUSDT", "CKBUSDT", "TWTUSDT", "FTTUSDT", "HNTUSDT", "OCEANUSDT", "BELUSDT", "COTIUSDT", "STMXUSDT", "DENTUSDT", "ONEUSDT", "CHRUSDT", "SANDUSDT", "MANAUSDT", "ANKRUSDT", "BTSUSDT", "LITUSDT", "UNFIUSDT", "REEFUSDT", "RVNUSDT", "SFPUSDT", "XEMUSDT", "COTIUSDT", "CHRUSDT", "STMXUSDT", "DENTUSDT", "ONEUSDT", "BTCSTUSDT", "TRUUSDT", "DEXEUSDT", "CKBUSDT", "TWTUSDT", "FTTUSDT", "HNTUSDT", "OCEANUSDT", "BELUSDT", "COTIUSDT", "STMXUSDT", "DENTUSDT", "ONEUSDT", "CHRUSDT", "SANDUSDT", "MANAUSDT", "ANKRUSDT", "BTSUSDT", "LITUSDT", "UNFIUSDT", "REEFUSDT", "RVNUSDT", "SFPUSDT", "XEMUSDT", "BTCDOMUSDT", "DEFIUSDT", "ADAUSDT", "TRXUSDT", "AVAXUSDT", "UNIUSDT", "SUSHIUSDT", "COMPUSDT", "MKRUSDT", "YFIUSDT", "SNXUSDT", "CRVUSDT", "1INCHUSDT", "ALPHAUSDT", "RENUSDT", "KSMUSDT", "WAVESUSDT", "ICXUSDT", "ONTUSDT", "ZILUSDT", "BATUSDT", "ZRXUSDT", "ENJUSDT", "STORJUSDT", "KNCUSDT", "LRCUSDT", "BANDUSDT", "SANDUSDT", "MANAUSDT", "CHZUSDT", "HOTUSDT", "VETUSDT", "WINUSDT", "DUSKUSDT", "DEFIUSDT", "YFIIUSDT", "AUDIOUSDT", "CTKUSDT", "AKROUSDT", "AXSUSDT", "HARDUSDT", "DNTUSDT", "STRKUSDT", "UNFIUSDT", "ROSEUSDT", "AVAUSDT", "XEMUSDT", "SKLUSDT", "GRTUSDT", "1000SHIBUSDT", "CELOUSDT", "RIFUSDT", "CKBUSDT", "FIROUSDT", "LITUSDT", "SFPUSDT", "DODOUSDT", "CAKEUSDT", "ACMUSDT", "BADGERUSDT", "FISUSDT", "OMUSDT", "PONDUSDT", "DEGOUSDT", "ALICEUSDT", "LINAUSDT", "PERPUSDT", "RAMPUSDT", "SUPERUSDT", "CFXUSDT", "EPSUSDT", "AUTOUSDT", "TKOUSDT", "PUNDIXUSDT", "TLMUSDT", "1000BTTUSDT", "BTCSTUSDT", "TRUUSDT", "DEXEUSDT", "CKBUSDT", "TWTUSDT", "FTTUSDT", "HNTUSDT", "OCEANUSDT", "BELUSDT", "COTIUSDT", "STMXUSDT", "DENTUSDT", "ONEUSDT", "CHRUSDT", "SANDUSDT", "MANAUSDT", "ANKRUSDT", "BTSUSDT", "LITUSDT", "UNFIUSDT", "REEFUSDT", "RVNUSDT", "SFPUSDT", "XEMUSDT", "COTIUSDT", "CHRUSDT", "STMXUSDT", "DENTUSDT", "ONEUSDT", "BTCSTUSDT", "TRUUSDT", "DEXEUSDT", "CKBUSDT", "TWTUSDT", "FTTUSDT", "HNTUSDT", "OCEANUSDT", "BELUSDT", "COTIUSDT", "STMXUSDT", "DENTUSDT", "ONEUSDT", "CHRUSDT", "SANDUSDT", "MANAUSDT", "ANKRUSDT", "BTSUSDT", "LITUSDT", "UNFIUSDT", "REEFUSDT", "RVNUSDT", "SFPUSDT", "XEMUSDT", "BTCDOMUSDT", "DEFIUSDT", "TAOUSDT", "ZECUSDT", "XMRUSDT", "DASHUSDT", "ETCUSDT", "BCHUSDT", "BSVUSDT", "XRPUSDT", "EOSUSDT", "XLMUSDT", "TRXUSDT", "IOTAUSDT", "NEOUSDT", "QTUMUSDT", "ALGOUSDT", "ZECUSDT", "XMRUSDT", "DASHUSDT", "ETCUSDT", "BCHUSDT", "BSVUSDT", "XRPUSDT", "EOSUSDT", "XLMUSDT", "TRXUSDT", "IOTAUSDT", "NEOUSDT", "QTUMUSDT", "ALGOUSDT"}
	for _, symbol := range symbols {
		jsonStr = strings.ReplaceAll(jsonStr, `"symbol": `+symbol, `"symbol": "`+symbol+`"`)
	}

	return jsonStr
}

// findMatchingBracket 查找匹配的右括号
func findMatchingBracket(s string, start int) int {
	if start >= len(s) || s[start] != '[' {
		return -1
	}

	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return i
			}
		}
	}

	return -1
}

// validateDecisions 验证所有决策的有效性
func validateDecisions(decisions []Decision, ctx *Context) error {
	for i, decision := range decisions {
		if err := validateDecision(&decision, ctx); err != nil {
			return fmt.Errorf("决策 %d 验证失败: %w", i+1, err)
		}
	}
	return nil
}

// validateDecision 验证单个决策的有效性
func validateDecision(decision *Decision, ctx *Context) error {
	// 调试：打印传入的模式
	log.Printf("[DEBUG] validateDecision: AIAutonomyMode=%v", ctx.AIAutonomyMode)
	
	// 🤖 AI自主模式：只做基本验证，不限制AI决策
	if ctx.AIAutonomyMode {
		log.Printf("🚀 [AI自主模式] 使用宽松验证，AI完全自主决策")
		return validateDecisionAutonomy(decision, ctx)
	}
	
	// 🔧 限制模式：计算智能风险管理参数
	log.Printf("🛡️ [限制模式] 使用严格风控验证")
	smartRisk := CalculateSmartRiskParams(ctx)
	
	// 验证action是否有效
	validActions := []string{"open_long", "open_short", "close_long", "close_short", "hold", "wait"}
	isValidAction := false
	for _, validAction := range validActions {
		if decision.Action == validAction {
			isValidAction = true
			break
		}
	}
	if !isValidAction {
		return fmt.Errorf("无效的action: %s", decision.Action)
	}

	// 对于开仓操作，验证参数
	if decision.Action == "open_long" || decision.Action == "open_short" {
		// 验证杠杆
		if decision.Leverage < 1 || decision.Leverage > 20 {
			return fmt.Errorf("杠杆必须在1-20之间，当前: %d", decision.Leverage)
		}

		// 验证仓位大小
		if decision.PositionSizeUSD <= 0 {
			return fmt.Errorf("仓位大小必须大于0: %.2f", decision.PositionSizeUSD)
		}

		// 🔧 优化：动态仓位大小验证（大幅提高基础限制）
		baseMaxPositionValue := 20.0 * ctx.Account.TotalEquity // 提高基础仓位限制到20倍
		if decision.Symbol == "BTCUSDT" || decision.Symbol == "ETHUSDT" {
			baseMaxPositionValue = 30.0 * ctx.Account.TotalEquity // BTC/ETH提高到30倍
		}
		
		// 使用智能仓位计算
		adjustedMaxPositionValue := CalculateSmartPositionSize(baseMaxPositionValue, smartRisk, decision.Symbol, decision.Confidence)
		
		positionValue := decision.PositionSizeUSD * float64(decision.Leverage)
		
		// 添加调试日志
		log.Printf("🛡️ [限制模式-仓位验证] 币种:%s 基础限制:%.2f 调整后:%.2f AI仓位价值:%.2f 信心度:%d 账户净值:%.2f 亏损率:%.1f%% 近期表现:%.1f",
			decision.Symbol, baseMaxPositionValue, adjustedMaxPositionValue, positionValue, 
			decision.Confidence, ctx.Account.TotalEquity, smartRisk.TotalPnLPct, smartRisk.RecentPerformance)
		
		if positionValue > adjustedMaxPositionValue {
			return fmt.Errorf("仓位价值过大: %.2f USDT (最大允许: %.2f USDT)", positionValue, adjustedMaxPositionValue)
		}

		// 🔧 新增：单笔交易最大风险限制
		maxSingleRisk := 0.05 * ctx.Account.TotalEquity // 5%
		if decision.Symbol == "BTCUSDT" || decision.Symbol == "ETHUSDT" {
			maxSingleRisk = 0.08 * ctx.Account.TotalEquity // 8%
		}
		
		// 验证止损
		if decision.StopLoss <= 0 {
			return fmt.Errorf("必须设置止损价格")
		}

		// 验证止盈
		if decision.TakeProfit <= 0 {
			return fmt.Errorf("必须设置止盈价格")
		}

		// 验证止损止盈的合理性
		if decision.Action == "open_long" {
			if decision.StopLoss >= decision.TakeProfit {
				return fmt.Errorf("做多时止损价必须小于止盈价")
			}
		} else {
			if decision.StopLoss <= decision.TakeProfit {
				return fmt.Errorf("做空时止损价必须大于止盈价")
			}
		}

		// 🔧 优化：动态风险回报比验证
		// 计算入场价（假设当前市价）
		var entryPrice float64
		if decision.Action == "open_long" {
			// 做多：入场价在止损和止盈之间
			entryPrice = decision.StopLoss + (decision.TakeProfit-decision.StopLoss)*0.2 // 假设在20%位置入场
		} else {
			// 做空：入场价在止损和止盈之间
			entryPrice = decision.StopLoss - (decision.StopLoss-decision.TakeProfit)*0.2 // 假设在20%位置入场
		}

		var riskPercent, rewardPercent, riskRewardRatio float64
		if decision.Action == "open_long" {
			riskPercent = (entryPrice - decision.StopLoss) / entryPrice * 100
			rewardPercent = (decision.TakeProfit - entryPrice) / entryPrice * 100
			if riskPercent > 0 {
				riskRewardRatio = rewardPercent / riskPercent
			}
		} else {
			riskPercent = (decision.StopLoss - entryPrice) / entryPrice * 100
			rewardPercent = (entryPrice - decision.TakeProfit) / entryPrice * 100
			if riskPercent > 0 {
				riskRewardRatio = rewardPercent / riskPercent
			}
		}

		// 🔧 优化：根据币种和信心度调整最小风险回报比
		minRiskReward := 3.0 // 默认3:1
		if decision.Symbol == "BTCUSDT" || decision.Symbol == "ETHUSDT" {
			minRiskReward = 1.8 // BTC/ETH降低到1.8:1
		}
		
		// 根据信心度调整
		if decision.Confidence >= 80 {
			minRiskReward *= 0.8 // 高信心度时降低要求
		} else if decision.Confidence < 60 {
			minRiskReward *= 1.2 // 低信心度时提高要求
		}
		
		// 根据最近表现调整
		if smartRisk.RecentPerformance > 70 {
			minRiskReward *= 0.9 // 表现好时稍微降低要求
		} else if smartRisk.RecentPerformance < 30 {
			minRiskReward *= 1.3 // 表现差时提高要求
		}

		if riskRewardRatio < minRiskReward {
			return fmt.Errorf("风险回报比过低: %.2f (最小要求: %.2f)", riskRewardRatio, minRiskReward)
		}

		// 🔧 新增：单笔最大风险限制验证
		estimatedRisk := decision.PositionSizeUSD * (riskPercent / 100) / float64(decision.Leverage)
		if estimatedRisk > maxSingleRisk {
			return fmt.Errorf("单笔风险过高(%.2f USDT)，最大允许%.2f USDT（%.1f%%账户净值）", 
				estimatedRisk, maxSingleRisk, (maxSingleRisk/ctx.Account.TotalEquity)*100)
		}
	}

	return nil
}

// validateDecisionAutonomy AI自主模式下的验证（只做基本安全检查）
func validateDecisionAutonomy(decision *Decision, ctx *Context) error {
	// 验证action是否有效
	validActions := map[string]bool{
		"open_long": true, "open_short": true,
		"close_long": true, "close_short": true,
		"hold": true, "wait": true,
	}
	if !validActions[decision.Action] {
		return fmt.Errorf("无效的action: %s", decision.Action)
	}

	// 对于开仓操作，只做基本数值验证
	if decision.Action == "open_long" || decision.Action == "open_short" {
		// 验证数值合法性（非负、非NaN）
		if decision.Leverage < 1 {
			return fmt.Errorf("杠杆必须大于0，当前: %d", decision.Leverage)
		}
		if decision.PositionSizeUSD < 0 {
			return fmt.Errorf("仓位大小不能为负数: %.2f", decision.PositionSizeUSD)
		}
		if decision.StopLoss < 0 {
			return fmt.Errorf("止损价格不能为负数: %.2f", decision.StopLoss)
		}
		if decision.TakeProfit < 0 {
			return fmt.Errorf("止盈价格不能为负数: %.2f", decision.TakeProfit)
		}
		
		// 验证止损止盈的方向正确性（防止反向设置）
		if decision.Action == "open_long" {
			if decision.StopLoss > 0 && decision.TakeProfit > 0 && decision.StopLoss >= decision.TakeProfit {
				return fmt.Errorf("做多时止损价应小于止盈价")
			}
		} else {
			if decision.StopLoss > 0 && decision.TakeProfit > 0 && decision.StopLoss <= decision.TakeProfit {
				return fmt.Errorf("做空时止损价应大于止盈价")
			}
		}
		
		log.Printf("🚀 [AI自主模式] ✅ 决策验证通过: %s %s 仓位:%.2f USDT 杠杆:%dx 信心度:%d%% (无限制)",
			decision.Action, decision.Symbol, decision.PositionSizeUSD, decision.Leverage, decision.Confidence)
	}

	return nil
}

// CalculateRiskMetrics 计算风险管理指标
func CalculateRiskMetrics(ctx *Context) RiskMetrics {
	metrics := RiskMetrics{}
	
	// 基础风险计算
	if ctx.DecisionLogger != nil {
		db := ctx.DecisionLogger.GetDB()
		if db != nil {
			// 获取最近的决策记录用于计算风险指标
			records, err := db.Decision().GetLatest(100) // 最近100个周期
			if err == nil && len(records) > 0 {
				metrics.SharpeRatio = calculateSharpeRatioFromRecords(records)
				metrics.MaxDrawdown, metrics.MaxDrawdownUSD = calculateMaxDrawdown(records)
				metrics.VaR95, metrics.VaR99 = calculateVaR(records)
			}
		}
	}
	
	// 计算当前持仓风险
	metrics.TotalRiskExposure = calculateTotalRiskExposure(ctx.Positions)
	metrics.LeverageRisk = calculateLeverageRisk(ctx.Positions, ctx.Account.TotalEquity)
	metrics.ConcentrationRisk = calculateConcentrationRisk(ctx.Positions)
	metrics.LiquidationRisk = calculateLiquidationRisk(ctx.Positions, ctx.Account.TotalEquity)
	metrics.VolatilityRisk = calculateVolatilityRisk(ctx.Positions, ctx.MarketDataMap)
	
	return metrics
}

// calculateSharpeRatioFromRecords 从决策记录计算夏普比率
func calculateSharpeRatioFromRecords(records []*models.DecisionRecord) float64 {
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

// calculateMaxDrawdown 计算最大回撤
func calculateMaxDrawdown(records []*models.DecisionRecord) (float64, float64) {
	if len(records) < 2 {
		return 0.0, 0.0
	}

	var equities []float64
	for _, record := range records {
		if record.TotalBalance > 0 {
			equities = append(equities, record.TotalBalance)
		}
	}

	if len(equities) < 2 {
		return 0.0, 0.0
	}

	maxDrawdownPct := 0.0
	maxDrawdownUSD := 0.0
	peak := equities[0]

	for _, equity := range equities {
		if equity > peak {
			peak = equity
		}
		
		drawdownUSD := peak - equity
		drawdownPct := (drawdownUSD / peak) * 100
		
		if drawdownPct > maxDrawdownPct {
			maxDrawdownPct = drawdownPct
			maxDrawdownUSD = drawdownUSD
		}
	}

	return maxDrawdownPct, maxDrawdownUSD
}

// calculateVaR 计算风险价值（VaR）
func calculateVaR(records []*models.DecisionRecord) (float64, float64) {
	if len(records) < 10 {
		return 0.0, 0.0
	}

	var returns []float64
	for i := 1; i < len(records); i++ {
		if records[i-1].TotalBalance > 0 {
			periodReturn := (records[i].TotalBalance - records[i-1].TotalBalance) / records[i-1].TotalBalance
			returns = append(returns, periodReturn)
		}
	}

	if len(returns) < 10 {
		return 0.0, 0.0
	}

	// 排序收益率
	sort.Float64s(returns)
	
	// 计算95%和99%置信度的VaR
	index95 := int(float64(len(returns)) * 0.05) // 5%分位数
	index99 := int(float64(len(returns)) * 0.01) // 1%分位数
	
	var95 := 0.0
	var99 := 0.0
	
	if index95 < len(returns) {
		var95 = -returns[index95] // VaR为负收益率的绝对值
	}
	if index99 < len(returns) {
		var99 = -returns[index99]
	}
	
	// 转换为USD金额（假设当前账户净值）
	currentEquity := records[len(records)-1].TotalBalance
	var95USD := var95 * currentEquity
	var99USD := var99 * currentEquity
	
	return var95USD, var99USD
}

// calculateTotalRiskExposure 计算总风险敞口
func calculateTotalRiskExposure(positions []PositionInfo) float64 {
	totalExposure := 0.0
	for _, pos := range positions {
		// 风险敞口 = 持仓价值 = 数量 × 当前价格
		exposure := math.Abs(pos.Quantity) * pos.MarkPrice
		totalExposure += exposure
	}
	return totalExposure
}

// calculateLeverageRisk 计算杠杆风险评分（0-100）
func calculateLeverageRisk(positions []PositionInfo, totalEquity float64) float64 {
	if totalEquity <= 0 {
		return 100.0 // 最高风险
	}
	
	totalMarginUsed := 0.0
	weightedLeverage := 0.0
	totalPositionValue := 0.0
	
	for _, pos := range positions {
		positionValue := math.Abs(pos.Quantity) * pos.MarkPrice
		totalPositionValue += positionValue
		totalMarginUsed += pos.MarginUsed
		weightedLeverage += float64(pos.Leverage) * positionValue
	}
	
	if totalPositionValue > 0 {
		weightedLeverage /= totalPositionValue
	}
	
	// 基于保证金使用率和平均杠杆计算风险评分
	marginUsageRisk := (totalMarginUsed / totalEquity) * 100
	leverageRisk := (weightedLeverage / 20.0) * 50 // 假设20倍杠杆为中等风险
	
	riskScore := marginUsageRisk + leverageRisk
	if riskScore > 100 {
		riskScore = 100
	}
	
	return riskScore
}

// calculateConcentrationRisk 计算集中度风险评分（0-100）
func calculateConcentrationRisk(positions []PositionInfo) float64 {
	if len(positions) == 0 {
		return 0.0
	}
	
	// 计算各持仓的价值占比
	totalValue := 0.0
	positionValues := make([]float64, len(positions))
	
	for i, pos := range positions {
		value := math.Abs(pos.Quantity) * pos.MarkPrice
		positionValues[i] = value
		totalValue += value
	}
	
	if totalValue == 0 {
		return 0.0
	}
	
	// 计算赫芬达尔指数（HHI）
	hhi := 0.0
	for _, value := range positionValues {
		share := value / totalValue
		hhi += share * share
	}
	
	// 将HHI转换为风险评分（0-100）
	// HHI范围：1/n（完全分散）到1（完全集中）
	// 风险评分：集中度越高，风险越大
	riskScore := hhi * 100
	
	return riskScore
}

// calculateLiquidationRisk 计算强平风险评分（0-100）
func calculateLiquidationRisk(positions []PositionInfo, totalEquity float64) float64 {
	if len(positions) == 0 || totalEquity <= 0 {
		return 0.0
	}
	
	minDistanceToLiquidation := math.Inf(1)
	
	for _, pos := range positions {
		if pos.LiquidationPrice <= 0 || pos.MarkPrice <= 0 {
			continue
		}
		
		// 计算到强平价的距离（百分比）
		var distancePct float64
		if pos.Side == "long" {
			distancePct = (pos.MarkPrice - pos.LiquidationPrice) / pos.MarkPrice * 100
		} else {
			distancePct = (pos.LiquidationPrice - pos.MarkPrice) / pos.MarkPrice * 100
		}
		
		if distancePct < minDistanceToLiquidation {
			minDistanceToLiquidation = distancePct
		}
	}
	
	if math.IsInf(minDistanceToLiquidation, 1) {
		return 0.0
	}
	
	// 将距离转换为风险评分
	// 距离越近，风险越高
	var riskScore float64
	if minDistanceToLiquidation <= 5 {
		riskScore = 100 // 极高风险
	} else if minDistanceToLiquidation <= 10 {
		riskScore = 80
	} else if minDistanceToLiquidation <= 20 {
		riskScore = 60
	} else if minDistanceToLiquidation <= 30 {
		riskScore = 40
	} else if minDistanceToLiquidation <= 50 {
		riskScore = 20
	} else {
		riskScore = 0 // 低风险
	}
	
	return riskScore
}

// calculateVolatilityRisk 计算波动率风险评分（0-100）
func calculateVolatilityRisk(positions []PositionInfo, marketDataMap map[string]*market.Data) float64 {
	if len(positions) == 0 {
		return 0.0
	}
	
	totalValue := 0.0
	weightedVolatility := 0.0
	
	for _, pos := range positions {
		positionValue := math.Abs(pos.Quantity) * pos.MarkPrice
		totalValue += positionValue
		
		// 获取市场数据计算波动率
		if marketData, exists := marketDataMap[pos.Symbol]; exists {
			// 使用价格变化作为波动率代理
			volatility := math.Abs(marketData.PriceChange1h) + math.Abs(marketData.PriceChange4h)
			weightedVolatility += volatility * positionValue
		}
	}
	
	if totalValue == 0 {
		return 0.0
	}
	
	avgVolatility := weightedVolatility / totalValue
	
	// 将波动率转换为风险评分
	// 假设10%的4小时波动率为高风险
	riskScore := (avgVolatility / 10.0) * 100
	if riskScore > 100 {
		riskScore = 100
	}
	
	return riskScore
}

// CalculateAccountRiskMetrics 计算账户风险相关字段
func CalculateAccountRiskMetrics(account *AccountInfo, totalEquity float64, positions []PositionInfo) {
	// 计算风险容量（基于2%风险规则）
	account.RiskCapacityUSD = totalEquity * 0.02
	
	// 单笔最大风险（账户净值的1%）
	account.MaxRiskPerTrade = totalEquity * 0.01
	
	// 日风险预算（账户净值的5%）
	account.DailyRiskBudget = totalEquity * 0.05
	
	// 计算已使用的风险预算（基于当前持仓的潜在损失）
	usedRisk := 0.0
	for _, pos := range positions {
		// 估算潜在损失（到止损位的距离）
		if pos.UnrealizedPnL < 0 {
			usedRisk += math.Abs(pos.UnrealizedPnL)
		}
	}
	account.UsedRiskBudget = usedRisk
}

// getRiskLevel 根据风险评分返回风险等级描述
func getRiskLevel(score float64) string {
	if score >= 80 {
		return "🔴极高风险"
	} else if score >= 60 {
		return "🟠高风险"
	} else if score >= 40 {
		return "🟡中等风险"
	} else if score >= 20 {
		return "🟢低风险"
	} else {
		return "✅安全"
	}
}


// 🔧 新增：智能风险管理结构
type SmartRiskManager struct {
	AccountEquity     float64
	TotalPnLPct       float64
	MarginUsedPct     float64
	RecentPerformance float64 // 最近表现评分 (0-100)
}

// 🔧 新增：计算智能风险管理参数
func CalculateSmartRiskParams(ctx *Context) *SmartRiskManager {
	srm := &SmartRiskManager{
		AccountEquity: ctx.Account.TotalEquity,
		TotalPnLPct:   ctx.Account.TotalPnLPct,
		MarginUsedPct: ctx.Account.MarginUsedPct,
	}
	
	// 计算最近表现评分
	if ctx.DecisionLogger != nil {
		db := ctx.DecisionLogger.GetDB()
		if db != nil {
			records, err := db.Decision().GetLatest(20) // 最近20个周期
			if err == nil && len(records) > 0 {
				srm.RecentPerformance = calculateRecentPerformanceScore(records)
			}
		}
	}
	
	return srm
}

// 🔧 新增：计算最近表现评分
func calculateRecentPerformanceScore(records []*models.DecisionRecord) float64 {
	if len(records) == 0 {
		return 50.0 // 默认中等评分
	}
	
	var totalReturn float64
	var winCount, lossCount int
	
	for _, record := range records {
		if record.TotalUnrealizedProfit != 0 {
			// 计算收益率百分比
			returnPct := record.TotalUnrealizedProfit / record.TotalBalance * 100
			totalReturn += returnPct
			if returnPct > 0 {
				winCount++
			} else {
				lossCount++
			}
		}
	}
	
	// 综合评分：收益率 + 胜率
	avgReturn := totalReturn / float64(len(records))
	winRate := float64(winCount) / float64(winCount+lossCount) * 100
	
	// 评分公式：基础50分 + 收益率贡献 + 胜率贡献
	score := 50.0 + avgReturn*2 + (winRate-50)*0.5
	
	// 限制在0-100范围内
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}
	
	return score
}

// 🔧 新增：智能仓位大小计算
func CalculateSmartPositionSize(baseSize float64, srm *SmartRiskManager, symbol string, confidence int) float64 {
	adjustedSize := baseSize
	
	// 1. 根据账户表现调整 - 更温和的系数，避免过度限制
	if srm.TotalPnLPct < -50 { // 亏损超过50%才大幅减少
		adjustedSize *= 0.8 // 减少20%仓位
	} else if srm.TotalPnLPct < -30 { // 亏损超过30%
		adjustedSize *= 0.9 // 减少10%仓位
	} else if srm.TotalPnLPct > 20 { // 盈利超过20%
		adjustedSize *= 1.2 // 增加20%仓位
	}
	
	// 2. 根据保证金使用率调整 - 只在极高使用率时才大幅减少
	if srm.MarginUsedPct > 85 {
		adjustedSize *= 0.6 // 极高保证金使用率时减少
	} else if srm.MarginUsedPct > 70 {
		adjustedSize *= 0.8 // 高保证金使用率时适度减少
	}
	
	// 3. 根据最近表现调整 - 大幅减少惩罚
	if srm.RecentPerformance < 10 {
		adjustedSize *= 0.85 // 表现极差时轻微减少仓位
	} else if srm.RecentPerformance > 80 {
		adjustedSize *= 1.15 // 表现很好时增加仓位
	}
	// 移除20-80之间的惩罚，给AI更多空间
	
	// 4. 根据信心度调整 - 大幅提高最低信心度
	confidenceMultiplier := float64(confidence) / 100.0
	if confidenceMultiplier < 0.85 {
		confidenceMultiplier = 0.85 // 最低85%，减少惩罚
	}
	adjustedSize *= confidenceMultiplier
	
	// 5. 币种特殊调整 - 移除山寨币惩罚
	// 不再对山寨币额外惩罚，让AI自主决策
	
	return adjustedSize
}

// 🔧 新增：动态止损计算
func CalculateDynamicStopLoss(entryPrice float64, side string, atr float64, volatility float64, confidence int) float64 {
	// 基础止损距离（使用ATR）
	baseStopDistance := atr * 2.0
	
	// 根据波动率调整
	if volatility > 0.05 { // 高波动率
		baseStopDistance *= 1.5
	} else if volatility < 0.02 { // 低波动率
		baseStopDistance *= 0.8
	}
	
	// 根据信心度调整
	confidenceAdjustment := 1.0 + (float64(confidence)-70)/100.0 // 信心度70为基准
	if confidenceAdjustment < 0.7 {
		confidenceAdjustment = 0.7
	} else if confidenceAdjustment > 1.3 {
		confidenceAdjustment = 1.3
	}
	baseStopDistance *= confidenceAdjustment
	
	// 计算止损价格
	var stopLoss float64
	if side == "long" {
		stopLoss = entryPrice - baseStopDistance
	} else {
		stopLoss = entryPrice + baseStopDistance
	}
	
	return stopLoss
}

// 🔧 新增：动态止盈计算
func CalculateDynamicTakeProfit(entryPrice float64, stopLoss float64, side string, riskRewardRatio float64) float64 {
	var riskDistance float64
	if side == "long" {
		riskDistance = entryPrice - stopLoss
	} else {
		riskDistance = stopLoss - entryPrice
	}
	
	rewardDistance := riskDistance * riskRewardRatio
	
	var takeProfit float64
	if side == "long" {
		takeProfit = entryPrice + rewardDistance
	} else {
		takeProfit = entryPrice - rewardDistance
	}
	
	return takeProfit
}

// SmartMarketAnalyzer 智能市场分析器
type SmartMarketAnalyzer struct {
	ctx *Context
}

// NewSmartMarketAnalyzer 创建智能市场分析器
func NewSmartMarketAnalyzer(ctx *Context) *SmartMarketAnalyzer {
	return &SmartMarketAnalyzer{ctx: ctx}
}

// AnalyzeMarketCondition 分析市场状况
func (sma *SmartMarketAnalyzer) AnalyzeMarketCondition() MarketCondition {
	btcData, hasBTC := sma.ctx.MarketDataMap["BTCUSDT"]
	if !hasBTC {
		return MarketCondition{
			Trend:      "unknown",
			Volatility: "medium",
			Sentiment:  "neutral",
			Risk:       "medium",
		}
	}

	// 分析趋势
	trend := sma.analyzeTrend(btcData)
	
	// 分析波动率
	volatility := sma.analyzeVolatility(btcData)
	
	// 分析市场情绪
	sentiment := sma.analyzeSentiment(btcData)
	
	// 评估风险等级
	risk := sma.assessRisk(btcData)

	return MarketCondition{
		Trend:      trend,
		Volatility: volatility,
		Sentiment:  sentiment,
		Risk:       risk,
	}
}

// analyzeTrend 分析趋势
func (sma *SmartMarketAnalyzer) analyzeTrend(data *market.Data) string {
	// 基于EMA和价格变化分析趋势
	if data.PriceChange4h > 2.0 && data.PriceChange1h > 0.5 {
		return "strong_bullish"
	} else if data.PriceChange4h > 0.5 && data.PriceChange1h > 0 {
		return "bullish"
	} else if data.PriceChange4h < -2.0 && data.PriceChange1h < -0.5 {
		return "strong_bearish"
	} else if data.PriceChange4h < -0.5 && data.PriceChange1h < 0 {
		return "bearish"
	} else {
		return "sideways"
	}
}

// analyzeVolatility 分析波动率
func (sma *SmartMarketAnalyzer) analyzeVolatility(data *market.Data) string {
	// 基于价格变化幅度分析波动率
	volatility := math.Abs(data.PriceChange1h) + math.Abs(data.PriceChange4h)
	
	if volatility > 5.0 {
		return "high"
	} else if volatility > 2.0 {
		return "medium"
	} else {
		return "low"
	}
}

// analyzeSentiment 分析市场情绪
func (sma *SmartMarketAnalyzer) analyzeSentiment(data *market.Data) string {
	// 基于RSI和MACD分析情绪
	if data.CurrentRSI7 > 70 && data.CurrentMACD > 0 {
		return "greedy"
	} else if data.CurrentRSI7 < 30 && data.CurrentMACD < 0 {
		return "fearful"
	} else if data.CurrentRSI7 > 60 {
		return "optimistic"
	} else if data.CurrentRSI7 < 40 {
		return "pessimistic"
	} else {
		return "neutral"
	}
}

// assessRisk 评估风险等级
func (sma *SmartMarketAnalyzer) assessRisk(data *market.Data) string {
	riskScore := 0
	
	// 波动率风险
	if math.Abs(data.PriceChange1h) > 3.0 {
		riskScore += 2
	} else if math.Abs(data.PriceChange1h) > 1.5 {
		riskScore += 1
	}
	
	// RSI极端值风险
	if data.CurrentRSI7 > 80 || data.CurrentRSI7 < 20 {
		riskScore += 2
	} else if data.CurrentRSI7 > 70 || data.CurrentRSI7 < 30 {
		riskScore += 1
	}
	
	// 账户保证金风险
	if sma.ctx.Account.MarginUsedPct > 70 {
		riskScore += 3
	} else if sma.ctx.Account.MarginUsedPct > 50 {
		riskScore += 2
	} else if sma.ctx.Account.MarginUsedPct > 30 {
		riskScore += 1
	}
	
	if riskScore >= 5 {
		return "very_high"
	} else if riskScore >= 3 {
		return "high"
	} else if riskScore >= 2 {
		return "medium"
	} else {
		return "low"
	}
}

// MarketCondition 市场状况
type MarketCondition struct {
	Trend      string `json:"trend"`      // strong_bullish, bullish, sideways, bearish, strong_bearish
	Volatility string `json:"volatility"` // low, medium, high
	Sentiment  string `json:"sentiment"`  // greedy, optimistic, neutral, pessimistic, fearful
	Risk       string `json:"risk"`       // low, medium, high, very_high
}

// DecisionQualityAnalyzer 决策质量分析器
type DecisionQualityAnalyzer struct {
	ctx             *Context
	marketCondition MarketCondition
}

// NewDecisionQualityAnalyzer 创建决策质量分析器
func NewDecisionQualityAnalyzer(ctx *Context, marketCondition MarketCondition) *DecisionQualityAnalyzer {
	return &DecisionQualityAnalyzer{
		ctx:             ctx,
		marketCondition: marketCondition,
	}
}

// EvaluateDecisionQuality 评估决策质量
func (dqa *DecisionQualityAnalyzer) EvaluateDecisionQuality(decision *Decision) DecisionQuality {
	issues := []string{}
	
	// 各维度权重配置
	weights := map[string]float64{
		"technical": 0.30, // 技术信号 30%
		"risk":      0.35, // 风险管理 35%
		"market":    0.20, // 市场环境 20%
		"timing":    0.15, // 时机选择 15%
	}
	
	// 检查技术信号质量
	techScore, techIssues := dqa.evaluateTechnicalSignals(decision)
	issues = append(issues, techIssues...)
	
	// 检查风险管理质量
	riskScore, riskIssues := dqa.evaluateRiskManagement(decision)
	issues = append(issues, riskIssues...)
	
	// 检查市场环境适应性
	envScore, envIssues := dqa.evaluateMarketEnvironment(decision)
	issues = append(issues, envIssues...)
	
	// 检查时机选择
	timingScore, timingIssues := dqa.evaluateTiming(decision)
	issues = append(issues, timingIssues...)
	
	// 加权平均计算总分（每个子项都是0-1之间的分数）
	score := techScore*weights["technical"]*100 +
		riskScore*weights["risk"]*100 +
		envScore*weights["market"]*100 +
		timingScore*weights["timing"]*100
	
	// 确定质量等级
	var grade string
	if score >= 80 {
		grade = "excellent"
	} else if score >= 60 {
		grade = "good"
	} else if score >= 40 {
		grade = "fair"
	} else {
		grade = "poor"
	}
	
	return DecisionQuality{
		Score:  score,
		Grade:  grade,
		Issues: issues,
	}
}

// evaluateTechnicalSignals 评估技术信号质量
func (dqa *DecisionQualityAnalyzer) evaluateTechnicalSignals(decision *Decision) (float64, []string) {
	score := 1.0
	issues := []string{}
	
	data, exists := dqa.ctx.MarketDataMap[decision.Symbol]
	if !exists {
		return 0.5, []string{"缺少市场数据"}
	}
	
	// 检查RSI信号
	if decision.Action == "open_long" && data.CurrentRSI7 > 70 {
		score *= 0.7
		issues = append(issues, "RSI超买状态下做多风险较高")
	}
	if decision.Action == "open_short" && data.CurrentRSI7 < 30 {
		score *= 0.7
		issues = append(issues, "RSI超卖状态下做空风险较高")
	}
	
	// 检查MACD信号
	if decision.Action == "open_long" && data.CurrentMACD < 0 {
		score *= 0.8
		issues = append(issues, "MACD负值时做多需谨慎")
	}
	if decision.Action == "open_short" && data.CurrentMACD > 0 {
		score *= 0.8
		issues = append(issues, "MACD正值时做空需谨慎")
	}
	
	// 布林通道信号检查
	if data.EnhancedIndicators != nil && data.EnhancedIndicators.BollingerBands != nil {
		bb := data.EnhancedIndicators.BollingerBands
		
		// 检查布林带位置
		if decision.Action == "open_long" {
			// 做多时价格在上轨附近风险较高（可能回调）
			if bb.Position > 0.9 {
				score *= 0.6
				issues = append(issues, "价格触及布林上轨，做多风险高，可能回调")
			} else if bb.Position > 0.7 {
				score *= 0.8
				issues = append(issues, "价格接近布林上轨，短期超买")
			}
			// 价格在下轨附近是好的做多机会
			if bb.Position < 0.3 {
				score *= 1.1 // 加分
			}
		}
		
		if decision.Action == "open_short" {
			// 做空时价格在下轨附近风险较高（可能反弹）
			if bb.Position < 0.1 {
				score *= 0.6
				issues = append(issues, "价格触及布林下轨，做空风险高，可能反弹")
			} else if bb.Position < 0.3 {
				score *= 0.8
				issues = append(issues, "价格接近布林下轨，短期超卖")
			}
			// 价格在上轨附近是好的做空机会
			if bb.Position > 0.7 {
				score *= 1.1 // 加分
			}
		}
		
		// 检查布林带宽度（波动率）
		if bb.Width < 2.0 {
			// Bollinger Squeeze - 窄带预示即将突破
			if decision.Action == "open_long" || decision.Action == "open_short" {
				issues = append(issues, fmt.Sprintf("布林带收窄(%.2f%%)，市场可能酝酿突破", bb.Width))
			}
		} else if bb.Width > 10.0 {
			// 宽带表示高波动
			if decision.Leverage > 5 {
				score *= 0.8
				issues = append(issues, fmt.Sprintf("布林带宽幅较大(%.2f%%)，高杠杆风险较高", bb.Width))
			}
		}
	}
	
	return score, issues
}

// evaluateRiskManagement 评估风险管理质量
func (dqa *DecisionQualityAnalyzer) evaluateRiskManagement(decision *Decision) (float64, []string) {
	score := 1.0
	issues := []string{}
	
	if decision.Action == "open_long" || decision.Action == "open_short" {
		data := dqa.ctx.MarketDataMap[decision.Symbol]
		
		// 检查止损设置
		if decision.StopLoss == 0 {
			score *= 0.3
			issues = append(issues, "未设置止损，风险极高")
		}
		
		// 检查止盈设置
		if decision.TakeProfit == 0 {
			score *= 0.7
			issues = append(issues, "未设置止盈目标")
		}
		
		// 检查风险回报比
		if decision.StopLoss > 0 && decision.TakeProfit > 0 {
			var riskRewardRatio float64
			if decision.Action == "open_long" {
				risk := math.Abs(decision.StopLoss - data.CurrentPrice)
				reward := math.Abs(decision.TakeProfit - data.CurrentPrice)
				if risk > 0 {
					riskRewardRatio = reward / risk
				}
			} else {
				risk := math.Abs(data.CurrentPrice - decision.StopLoss)
				reward := math.Abs(data.CurrentPrice - decision.TakeProfit)
				if risk > 0 {
					riskRewardRatio = reward / risk
				}
			}
			
			if riskRewardRatio < 1.5 {
				score *= 0.5
				issues = append(issues, fmt.Sprintf("风险回报比%.2f过低", riskRewardRatio))
			} else if riskRewardRatio < 2.0 {
				score *= 0.8
				issues = append(issues, fmt.Sprintf("风险回报比%.2f偏低", riskRewardRatio))
			}
		}
		
		// 布林通道辅助止损验证
		if data.EnhancedIndicators != nil && data.EnhancedIndicators.BollingerBands != nil {
			bb := data.EnhancedIndicators.BollingerBands
			
			// 检查止损位置是否合理（应该在布林带外）
			if decision.Action == "open_long" && decision.StopLoss > 0 {
				// 做多止损应该在下轨以下
				if decision.StopLoss > bb.Lower {
					score *= 0.9
					issues = append(issues, fmt.Sprintf("做多止损%.2f在布林下轨%.2f之上，空间不足", decision.StopLoss, bb.Lower))
				}
				// 理想：止损在下轨下方1-2个ATR
				stopDistance := (data.CurrentPrice - decision.StopLoss) / data.CurrentPrice * 100
				bbWidth := bb.Width / 2 // 中轨到边轨的距离
				if stopDistance < bbWidth * 0.5 {
					score *= 0.9
					issues = append(issues, fmt.Sprintf("止损距离%.2f%%过小，易被噪音触发(建议>%.2f%%)", stopDistance, bbWidth*0.5))
				}
			}
			
			if decision.Action == "open_short" && decision.StopLoss > 0 {
				// 做空止损应该在上轨以上
				if decision.StopLoss < bb.Upper {
					score *= 0.9
					issues = append(issues, fmt.Sprintf("做空止损%.2f在布林上轨%.2f之下，空间不足", decision.StopLoss, bb.Upper))
				}
				stopDistance := (decision.StopLoss - data.CurrentPrice) / data.CurrentPrice * 100
				bbWidth := bb.Width / 2
				if stopDistance < bbWidth * 0.5 {
					score *= 0.9
					issues = append(issues, fmt.Sprintf("止损距离%.2f%%过小，易被噪音触发(建议>%.2f%%)", stopDistance, bbWidth*0.5))
				}
			}
		}
		
		// 根据布林带宽度调整仓位大小建议
		if data.EnhancedIndicators != nil && data.EnhancedIndicators.BollingerBands != nil {
			bb := data.EnhancedIndicators.BollingerBands
			baseMaxSize := dqa.ctx.Account.TotalEquity * 3.0
			
			// 高波动时降低仓位上限
			if bb.Width > 10.0 {
				maxPositionSize := baseMaxSize * 0.7 // 降低30%
				if decision.PositionSizeUSD > maxPositionSize {
					score *= 0.7
					issues = append(issues, fmt.Sprintf("高波动环境(BB宽度%.2f%%)，建议降低仓位", bb.Width))
				}
			} else if bb.Width < 2.0 {
				// 低波动（Squeeze）时可以适当加大仓位
				maxPositionSize := baseMaxSize * 1.2 // 提高20%
				if decision.PositionSizeUSD > maxPositionSize {
					score *= 0.8
					issues = append(issues, "即使低波动，仓位仍需控制")
				}
			} else {
				// 正常波动
				if decision.PositionSizeUSD > baseMaxSize {
					score *= 0.6
					issues = append(issues, "仓位过大，超出风险承受能力")
				}
			}
		} else {
			// 没有布林带数据时的默认检查
			maxPositionSize := dqa.ctx.Account.TotalEquity * 3.0
			if decision.PositionSizeUSD > maxPositionSize {
				score *= 0.6
				issues = append(issues, "仓位过大，超出风险承受能力")
			}
		}
	}
	
	return score, issues
}

// evaluateMarketEnvironment 评估市场环境适应性
func (dqa *DecisionQualityAnalyzer) evaluateMarketEnvironment(decision *Decision) (float64, []string) {
	score := 1.0
	issues := []string{}
	
	data := dqa.ctx.MarketDataMap[decision.Symbol]
	
	// 高风险环境下的决策评估
	if dqa.marketCondition.Risk == "very_high" || dqa.marketCondition.Risk == "high" {
		if decision.Action == "open_long" || decision.Action == "open_short" {
			score *= 0.6
			issues = append(issues, "高风险环境下开新仓需要更强的信号确认")
		}
	}
	
	// 高波动环境下的决策评估（优先使用布林带宽度）
	if data.EnhancedIndicators != nil && data.EnhancedIndicators.BollingerBands != nil {
		bb := data.EnhancedIndicators.BollingerBands
		
		if bb.Width > 10.0 {
			// 高波动
			if decision.Leverage > 5 {
				score *= 0.6
				issues = append(issues, fmt.Sprintf("高波动环境(BB宽度%.2f%%)，高杠杆风险大", bb.Width))
			}
		} else if bb.Width < 2.0 {
			// 低波动 - Bollinger Squeeze
			if decision.Action == "open_long" || decision.Action == "open_short" {
				// Squeeze后的突破往往很强劲
				if bb.Position > 0.8 || bb.Position < 0.2 {
					score *= 1.15 // 加分：突破布林带的Squeeze
					issues = append(issues, fmt.Sprintf("布林带收窄(%.2f%%)后突破，信号较强", bb.Width))
				} else {
					score *= 0.85
					issues = append(issues, fmt.Sprintf("布林带收窄(%.2f%%)，方向未明确前等待", bb.Width))
				}
			}
		}
	} else if dqa.marketCondition.Volatility == "high" {
		// 没有布林带数据时使用市场条件
		if decision.Leverage > 5 {
			score *= 0.7
			issues = append(issues, "高波动环境下使用高杠杆风险较大")
		}
	}
	
	// 极端情绪下的决策评估
	if dqa.marketCondition.Sentiment == "greedy" && decision.Action == "open_long" {
		score *= 0.8
		issues = append(issues, "市场贪婪时做多需要谨慎")
	}
	if dqa.marketCondition.Sentiment == "fearful" && decision.Action == "open_short" {
		score *= 0.8
		issues = append(issues, "市场恐慌时做空需要谨慎")
	}
	
	// 布林带整体趋势判断
	if data.EnhancedIndicators != nil && data.EnhancedIndicators.BollingerBands != nil {
		bb := data.EnhancedIndicators.BollingerBands
		
		// 价格持续在上轨运行（强势上升趋势）
		if bb.Position > 0.85 && data.CurrentPrice > bb.Upper {
			if decision.Action == "open_short" {
				score *= 0.7
				issues = append(issues, "价格沿上轨强势上涨，逆势做空风险高")
			}
		}
		
		// 价格持续在下轨运行（强势下降趋势）
		if bb.Position < 0.15 && data.CurrentPrice < bb.Lower {
			if decision.Action == "open_long" {
				score *= 0.7
				issues = append(issues, "价格沿下轨强势下跌，逆势做多风险高")
			}
		}
	}
	
	return score, issues
}

// evaluateTiming 评估时机选择
func (dqa *DecisionQualityAnalyzer) evaluateTiming(decision *Decision) (float64, []string) {
	score := 1.0
	issues := []string{}
	
	// 检查账户状态
	if dqa.ctx.Account.MarginUsedPct > 70 && (decision.Action == "open_long" || decision.Action == "open_short") {
		score *= 0.5
		issues = append(issues, "保证金使用率过高，不宜开新仓")
	}
	
	// 检查持仓数量
	if len(dqa.ctx.Positions) >= 3 && (decision.Action == "open_long" || decision.Action == "open_short") {
		score *= 0.8
		issues = append(issues, "持仓过多，增加管理难度")
	}
	
	// 检查信心度
	if decision.Confidence < 75 && (decision.Action == "open_long" || decision.Action == "open_short") {
		score *= 0.7
		issues = append(issues, "信心度不足，建议等待更好机会")
	}
	
	return score, issues
}

// DecisionQuality 决策质量
type DecisionQuality struct {
	Score  float64  `json:"score"`  // 0-100分
	Grade  string   `json:"grade"`  // excellent, good, fair, poor
	Issues []string `json:"issues"` // 问题列表
}

// ... existing code ...
