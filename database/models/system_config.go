package models

import "time"

// SystemConfig 系统配置表
type SystemConfig struct {
	ID          int64
	Key         string // 配置键名
	Value       string // 配置值（JSON格式）
	Description string // 配置说明
	ConfigType  string // 配置类型：market, trading, risk, api
	UpdatedAt   time.Time
}

// TraderConfig 交易员配置表
type TraderConfig struct {
	ID          int64
	UserID      int64  // 所属用户ID（0表示系统默认）
	TraderID    string // Trader唯一标识
	Name        string
	Enabled     bool
	AIModel     string // qwen, deepseek, custom
	Exchange    string // binance, hyperliquid, aster
	
	// API配置（加密存储）
	BinanceAPIKey       string
	BinanceSecretKey    string
	HyperliquidPrivateKey string
	HyperliquidWalletAddr string
	HyperliquidTestnet    bool
	AsterUser           string
	AsterSigner         string
	AsterPrivateKey     string
	
	// AI配置
	DeepSeekKey     string
	QwenKey         string
	CustomAPIURL    string
	CustomAPIKey    string
	CustomModelName string
	
	// 交易配置
	InitialBalance      float64
	ScanIntervalMinutes int // 扫描间隔（分钟）
	MaxPositions        int
	BTCETHLeverage      int
	AltcoinLeverage     int
	
	// 风控配置
	MaxDailyLoss        float64
	MaxDrawdown         float64
	StopTradingMinutes  int
	
	// AI学习配置
	EnableAILearning bool
	AILearnInterval  int
	
	// AI自主模式配置
	AIAutonomyMode bool // true=完全自主, false=限制模式(默认)
	
	// 数据优化配置
	CompactMode bool // true=紧凑模式（减少数据量），false=完整模式
	
	CreatedAt time.Time
	UpdatedAt time.Time
}

// MarketDataConfig 市场数据配置
type MarketDataConfig struct {
	ID              int64
	CoinPoolAPIURL  string
	OITopAPIURL     string
	UseDefaultCoins bool
	DefaultCoins    string // JSON数组
	KlineSettings   string // JSON数组
	UpdatedAt       time.Time
}
