package monitoring

// 风险评分阈值常量
const (
	// 保证金使用率阈值 (%)
	MarginUsageHighThreshold   = 50.0 // 高风险阈值
	MarginUsageMediumThreshold = 20.0 // 中风险阈值
	
	// 最大回撤阈值 (%)
	DrawdownCriticalThreshold = 30.0 // 危险阈值
	DrawdownHighThreshold     = 20.0 // 高风险阈值
	DrawdownMediumThreshold   = 10.0 // 中风险阈值
	
	// 夏普比率阈值
	SharpeRatioLowThreshold  = -0.5 // 低阈值（表现极差）
	SharpeRatioPoorThreshold = 0.0  // 差阈值（无风险调整后收益）
	
	// 胜率阈值 (%)
	WinRateLowThreshold = 30.0 // 低胜率阈值
	
	// 错误率阈值 (%)
	ErrorRateHighThreshold = 10.0 // 高错误率阈值
	
	// 交易频率阈值
	TradesPerHourHighThreshold = 1.0 // 高频交易阈值（每小时超过1笔）
	TradesPerHourMediumThreshold = 0.5 // 中频交易阈值
)

// 风险评分权重常量
const (
	// 保证金使用评分
	MarginHighScore   = 20 // 高使用率评分
	MarginMediumScore = 10 // 中使用率评分
	
	// 回撤评分
	DrawdownCriticalScore = 30 // 危险回撤评分
	DrawdownHighScore     = 20 // 高回撤评分
	DrawdownMediumScore   = 10 // 中回撤评分
	
	// 夏普比率评分
	SharpeRatioLowScore  = 20 // 低夏普比率评分
	SharpeRatioPoorScore = 10 // 差夏普比率评分
	
	// 波动率评分
	VolatilityHighScore   = 10 // 高波动率评分
	VolatilityMediumScore = 5  // 中波动率评分
	
	// 交易频率评分
	OverTradingHighScore   = 100 // 极度过度交易评分
	OverTradingMediumScore = 10  // 中度过度交易评分
)

// 统计分析常量
const (
	// 最小样本数要求
	MinTradesForWinRate  = 10 // 计算胜率的最小交易数
	MinRecordsForSharpe  = 2  // 计算夏普比率的最小记录数
	MinBalancesForDrawdown = 10 // 计算回撤的最小余额样本数
	
	// 默认风险评分
	DefaultRiskScore = 50 // 无数据时的默认风险评分
)

// 监控阈值常量
const (
	// 风险等级分界
	RiskCriticalThreshold = 80  // 危险风险等级（≥80分）
	RiskHighThreshold     = 60  // 高风险等级（≥60分）
	RiskMediumThreshold   = 40  // 中风险等级（≥40分）
	
	// 表现等级分界
	PerformanceExcellent = 80.0 // 优秀表现（≥80分）
	PerformanceGood      = 60.0 // 良好表现（≥60分）
	PerformanceFair      = 40.0 // 一般表现（≥40分）
)
