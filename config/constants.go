package config

// API 服务器配置常量
const (
	DefaultAPIPort = 8080 // 默认API服务器端口
	MinAPIPort     = 1024 // 最小有效端口
	MaxAPIPort     = 65535 // 最大有效端口
)

// 交易配置常量
const (
	DefaultMaxPositions      = 3 // 默认最大持仓数
	DefaultScanIntervalMin   = 3 // 默认扫描间隔（分钟）
	DefaultBTCETHLeverage    = 5 // BTC/ETH默认杠杆倍数
	DefaultAltcoinLeverage   = 5 // 山寨币默认杠杆倍数
	
	// 子账户杠杆限制（币安子账户限制）
	MaxSubAccountLeverage = 5 // 子账户最大杠杆（币安限制）
)

// AI 学习配置常量
const (
	DefaultAILearnInterval = 10  // 默认AI学习间隔（周期数）
	MinAILearnInterval     = 5   // 最小AI学习间隔
	DefaultEnableAILearning = false // 默认禁用AI学习
)

// 风控配置常量
const (
	DefaultMaxDailyLoss       = 0.0 // 默认最大日损失（0表示不限制）
	DefaultMaxDrawdown        = 0.0 // 默认最大回撤（0表示不限制）
	DefaultStopTradingMinutes = 0   // 默认停止交易时间（0表示不限制）
)

// 备份配置常量
const (
	DefaultBackupRetention = 5 // 默认保留备份数量
)

// 币种池配置常量
const (
	DefaultUseDefaultCoins = true // 默认使用默认币种列表
)

// 配置文件路径常量
const (
	DefaultConfigFile = "config.json"
	ConfigFileMode    = 0644 // 配置文件权限
)

// 验证规则常量
const (
	MinInitialBalance = 0.0  // 最小初始余额（必须大于0）
	MinScanInterval   = 1    // 最小扫描间隔（分钟）
	MaxScanInterval   = 1440 // 最大扫描间隔（24小时）
)

// 默认币种列表（如果配置文件中未指定）
var DefaultCoinList = []string{
	"BTCUSDT",
	"ETHUSDT",
	"SOLUSDT",
	"BNBUSDT",
	"XRPUSDT",
	"DOGEUSDT",
	"ADAUSDT",
	"HYPEUSDT",
}

// 支持的AI模型列表
const (
	AIModelQwen     = "qwen"
	AIModelDeepSeek = "deepseek"
	AIModelCustom   = "custom"
)

// 支持的交易平台列表
const (
	ExchangeBinance      = "binance"
	ExchangeHyperliquid  = "hyperliquid"
	ExchangeAster        = "aster"
)

// 默认K线配置
var DefaultKlineSettings = []KlineConfig{
	{
		Interval:  "3m",
		Limit:     20,
		ShowTable: true,
	},
	{
		Interval:  "4h",
		Limit:     60,
		ShowTable: false,
	},
}
