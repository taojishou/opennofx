package market

// 技术指标参数常量
const (
	// 布林带 (Bollinger Bands)
	DefaultBollingerPeriod = 20  // 布林带周期
	DefaultBollingerStdDev = 2.0 // 标准差倍数
	
	// 随机指标 (Stochastic Oscillator)
	DefaultStochasticK = 14 // K值周期
	DefaultStochasticD = 3  // D值周期
	
	// 商品通道指数 (CCI)
	DefaultCCIPeriod = 20 // CCI周期
	
	// 成交量加权移动平均 (VWMA)
	DefaultVWMAPeriod = 20 // VWMA周期
	
	// 历史波动率
	DefaultHistoricalVolPeriod = 20 // 历史波动率计算周期
	
	// RSI (Relative Strength Index)
	DefaultRSIPeriod = 14 // RSI周期
	
	// MACD (Moving Average Convergence Divergence)
	DefaultMACDFastPeriod   = 12 // 快线周期
	DefaultMACDSlowPeriod   = 26 // 慢线周期
	DefaultMACDSignalPeriod = 9  // 信号线周期
	
	// 移动平均线 (Moving Average)
	DefaultSMAPeriod = 20 // 简单移动平均周期
	DefaultEMAPeriod = 20 // 指数移动平均周期
)

// K线数据要求常量
const (
	MinKlinesForIndicators = 50 // 计算指标的最小K线数量
	MinKlinesForAnalysis   = 20 // 基础分析的最小K线数量
)

// 技术指标阈值常量
const (
	// 随机指标阈值
	StochasticOversoldThreshold = 20.0 // 超卖阈值
	StochasticOverboughtThreshold = 80.0 // 超买阈值
	
	// RSI阈值
	RSIOversoldThreshold = 30.0 // RSI超卖阈值
	RSIOverboughtThreshold = 70.0 // RSI超买阈值
	
	// 布林带宽度阈值
	BollingerWidthNarrowThreshold = 2.0  // 窄带阈值 (%)
	BollingerWidthWideThreshold   = 10.0 // 宽带阈值 (%)
)

// 价格计算常量
const (
	TypicalPriceDivisor = 3.0 // 典型价格计算除数: (H+L+C)/3
	PercentageMultiplier = 100.0 // 百分比计算乘数
)

// 市场分析配置常量
const (
	DefaultKlineInterval = "3m"  // 默认K线时间间隔
	DefaultKlineLimit    = 20    // 默认K线数量
	DefaultShowTable     = true  // 默认是否显示K线表格
	
	// 多时间框架配置
	ShortTermInterval  = "3m"  // 短期时间框架
	MediumTermInterval = "4h"  // 中期时间框架
	LongTermInterval   = "1d"  // 长期时间框架
)

// 恐慌贪婪指数范围
const (
	FearGreedIndexMin = 0   // 最小值（极度恐慌）
	FearGreedIndexMax = 100 // 最大值（极度贪婪）
	
	FearGreedExtremeFear   = 25 // 极度恐慌阈值
	FearGreedFear          = 45 // 恐慌阈值
	FearGreedGreed         = 55 // 贪婪阈值
	FearGreedExtremeGreed  = 75 // 极度贪婪阈值
)

// CompactMode 紧凑模式开关（减少发送给AI的数据量）
var CompactMode = true
