package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// TraderConfig 单个trader的配置
type TraderConfig struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"` // 是否启用该trader
	AIModel string `json:"ai_model"` // "qwen" or "deepseek"

	// 交易平台选择（二选一）
	Exchange string `json:"exchange"` // "binance" or "hyperliquid"

	// 币安配置
	BinanceAPIKey    string `json:"binance_api_key,omitempty"`
	BinanceSecretKey string `json:"binance_secret_key,omitempty"`

	// Hyperliquid配置
	HyperliquidPrivateKey string `json:"hyperliquid_private_key,omitempty"`
	HyperliquidWalletAddr string `json:"hyperliquid_wallet_addr,omitempty"`
	HyperliquidTestnet    bool   `json:"hyperliquid_testnet,omitempty"`

	// Aster配置
	AsterUser       string `json:"aster_user,omitempty"`        // Aster主钱包地址
	AsterSigner     string `json:"aster_signer,omitempty"`      // Aster API钱包地址
	AsterPrivateKey string `json:"aster_private_key,omitempty"` // Aster API钱包私钥

	// AI配置
	QwenKey     string `json:"qwen_key,omitempty"`
	DeepSeekKey string `json:"deepseek_key,omitempty"`

	// 自定义AI API配置（支持任何OpenAI格式的API）
	CustomAPIURL    string `json:"custom_api_url,omitempty"`
	CustomAPIKey    string `json:"custom_api_key,omitempty"`
	CustomModelName string `json:"custom_model_name,omitempty"`

	InitialBalance      float64 `json:"initial_balance"`
	ScanIntervalMinutes int     `json:"scan_interval_minutes"`
	
	// AI自主模式（true=完全自主决策，false=限制模式）
	AIAutonomyMode bool `json:"ai_autonomy_mode"`
	
	// 数据优化配置（true=紧凑模式，false=完整模式）
	CompactMode bool `json:"compact_mode"`
}

// LeverageConfig 杠杆配置
type LeverageConfig struct {
	BTCETHLeverage  int `json:"btc_eth_leverage"` // BTC和ETH的杠杆倍数（主账户建议5-50，子账户≤5）
	AltcoinLeverage int `json:"altcoin_leverage"` // 山寨币的杠杆倍数（主账户建议5-20，子账户≤5）
}

// KlineConfig K线数据配置
type KlineConfig struct {
	Interval  string `json:"interval"`   // K线时间周期: "3m", "5m", "15m", "1h", "4h", "1d"
	Limit     int    `json:"limit"`      // 显示多少根K线
	ShowTable bool   `json:"show_table"` // 是否显示K线表格（如果false只显示技术指标序列）
}

// MarketDataConfig 市场数据配置
type MarketDataConfig struct {
	Klines []KlineConfig `json:"klines"` // 支持多个时间框架的K线
}

// Config 总配置
type Config struct {
	Traders            []TraderConfig   `json:"traders"`
	UseDefaultCoins    bool             `json:"use_default_coins"` // 是否使用默认主流币种列表
	DefaultCoins       []string         `json:"default_coins"`     // 默认主流币种池
	CoinPoolAPIURL     string           `json:"coin_pool_api_url"`
	OITopAPIURL        string           `json:"oi_top_api_url"`
	APIServerPort      int              `json:"api_server_port"`
	MaxPositions       int              `json:"max_positions"`        // 最大持仓数限制（默认3）
	MaxDailyLoss       float64          `json:"max_daily_loss"`
	MaxDrawdown        float64          `json:"max_drawdown"`
	StopTradingMinutes int              `json:"stop_trading_minutes"`
	Leverage           LeverageConfig   `json:"leverage"`           // 杠杆配置
	EnableAILearning   bool             `json:"enable_ai_learning"` // 是否启用AI自动学习
	AILearnInterval    int              `json:"ai_learn_interval"`  // AI学习间隔（周期数）
	AIAutonomyMode     bool             `json:"ai_autonomy_mode"`   // AI自主模式（全局开关）
	CompactMode        bool             `json:"compact_mode"`       // 数据优化模式（紧凑/完整）
	MarketData         MarketDataConfig `json:"market_data"`        // 市场数据配置
}

// LoadConfig 从文件加载配置
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置默认值：如果use_default_coins未设置（为false）且没有配置coin_pool_api_url，则默认使用默认币种列表
	if !config.UseDefaultCoins && config.CoinPoolAPIURL == "" {
		config.UseDefaultCoins = true
	}

	// 设置默认币种池
	if len(config.DefaultCoins) == 0 {
		config.DefaultCoins = []string{
			"BTCUSDT",
			"ETHUSDT",
			"SOLUSDT",
			"BNBUSDT",
			"XRPUSDT",
			"DOGEUSDT",
			"ADAUSDT",
			"HYPEUSDT",
		}
	}

	// 设置默认市场数据配置
	if len(config.MarketData.Klines) == 0 {
		config.MarketData.Klines = []KlineConfig{
			{Interval: "3m", Limit: 20, ShowTable: true},  // 3分钟K线，显示20根（1小时）
			{Interval: "4h", Limit: 60, ShowTable: false}, // 4小时K线，不显示表格
		}
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// Validate 验证配置有效性
func (c *Config) Validate() error {
	if len(c.Traders) == 0 {
		return fmt.Errorf("至少需要配置一个trader")
	}

	traderIDs := make(map[string]bool)
	for i, trader := range c.Traders {
		if trader.ID == "" {
			return fmt.Errorf("trader[%d]: ID不能为空", i)
		}
		if traderIDs[trader.ID] {
			return fmt.Errorf("trader[%d]: ID '%s' 重复", i, trader.ID)
		}
		traderIDs[trader.ID] = true

		if trader.Name == "" {
			return fmt.Errorf("trader[%d]: Name不能为空", i)
		}
		if trader.AIModel != "qwen" && trader.AIModel != "deepseek" && trader.AIModel != "custom" {
			return fmt.Errorf("trader[%d]: ai_model必须是 'qwen', 'deepseek' 或 'custom'", i)
		}

		// 验证交易平台配置
		if trader.Exchange == "" {
			trader.Exchange = "binance" // 默认使用币安
		}
		if trader.Exchange != "binance" && trader.Exchange != "hyperliquid" && trader.Exchange != "aster" {
			return fmt.Errorf("trader[%d]: exchange必须是 'binance', 'hyperliquid' 或 'aster'", i)
		}

		// 根据平台验证对应的密钥
		if trader.Exchange == "binance" {
			if trader.BinanceAPIKey == "" || trader.BinanceSecretKey == "" {
				return fmt.Errorf("trader[%d]: 使用币安时必须配置binance_api_key和binance_secret_key", i)
			}
		} else if trader.Exchange == "hyperliquid" {
			if trader.HyperliquidPrivateKey == "" {
				return fmt.Errorf("trader[%d]: 使用Hyperliquid时必须配置hyperliquid_private_key", i)
			}
		} else if trader.Exchange == "aster" {
			if trader.AsterUser == "" || trader.AsterSigner == "" || trader.AsterPrivateKey == "" {
				return fmt.Errorf("trader[%d]: 使用Aster时必须配置aster_user, aster_signer和aster_private_key", i)
			}
		}

		if trader.AIModel == "qwen" && trader.QwenKey == "" {
			return fmt.Errorf("trader[%d]: 使用Qwen时必须配置qwen_key", i)
		}
		if trader.AIModel == "deepseek" && trader.DeepSeekKey == "" {
			return fmt.Errorf("trader[%d]: 使用DeepSeek时必须配置deepseek_key", i)
		}
		if trader.AIModel == "custom" {
			if trader.CustomAPIURL == "" {
				return fmt.Errorf("trader[%d]: 使用自定义API时必须配置custom_api_url", i)
			}
			if trader.CustomAPIKey == "" {
				return fmt.Errorf("trader[%d]: 使用自定义API时必须配置custom_api_key", i)
			}
			if trader.CustomModelName == "" {
				return fmt.Errorf("trader[%d]: 使用自定义API时必须配置custom_model_name", i)
			}
		}
		if trader.InitialBalance <= 0 {
			return fmt.Errorf("trader[%d]: initial_balance必须大于0", i)
		}
		if trader.ScanIntervalMinutes <= 0 {
			trader.ScanIntervalMinutes = 3 // 默认3分钟
		}
	}

	if c.APIServerPort <= 0 {
		c.APIServerPort = 8080 // 默认8080端口
	}

	// 设置最大持仓数默认值
	if c.MaxPositions <= 0 {
		c.MaxPositions = 3 // 默认3个持仓
	}

	// 设置杠杆默认值（适配币安子账户限制，最大5倍）
	if c.Leverage.BTCETHLeverage <= 0 {
		c.Leverage.BTCETHLeverage = 5 // 默认5倍（安全值，适配子账户）
	}
	if c.Leverage.BTCETHLeverage > 5 {
		fmt.Printf("⚠️  警告: BTC/ETH杠杆设置为%dx，如果使用子账户可能会失败（子账户限制≤5x）\n", c.Leverage.BTCETHLeverage)
	}
	if c.Leverage.AltcoinLeverage <= 0 {
		c.Leverage.AltcoinLeverage = 5 // 默认5倍（安全值，适配子账户）
	}
	if c.Leverage.AltcoinLeverage > 5 {
		fmt.Printf("⚠️  警告: 山寨币杠杆设置为%dx，如果使用子账户可能会失败（子账户限制≤5x）\n", c.Leverage.AltcoinLeverage)
	}

	return nil
}

// GetScanInterval 获取扫描间隔
func (tc *TraderConfig) GetScanInterval() time.Duration {
	return time.Duration(tc.ScanIntervalMinutes) * time.Minute
}

// SaveConfig 保存配置到文件
func SaveConfig(filename string, config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// GetConfigFilePath 获取配置文件路径
func GetConfigFilePath() string {
	return "config.json"
}

// MaskSensitiveData 脱敏敏感数据（用于API返回）
func (c *Config) MaskSensitiveData() *Config {
	masked := *c
	masked.Traders = make([]TraderConfig, len(c.Traders))
	
	for i, trader := range c.Traders {
		maskedTrader := trader
		
		// 保留exchange字段（前端需要用于显示配置）
		maskedTrader.Exchange = trader.Exchange
		
		// 脱敏API密钥
		if maskedTrader.BinanceAPIKey != "" {
			maskedTrader.BinanceAPIKey = maskString(maskedTrader.BinanceAPIKey)
		}
		if maskedTrader.BinanceSecretKey != "" {
			maskedTrader.BinanceSecretKey = maskString(maskedTrader.BinanceSecretKey)
		}
		if maskedTrader.HyperliquidPrivateKey != "" {
			maskedTrader.HyperliquidPrivateKey = maskString(maskedTrader.HyperliquidPrivateKey)
		}
		if maskedTrader.AsterPrivateKey != "" {
			maskedTrader.AsterPrivateKey = maskString(maskedTrader.AsterPrivateKey)
		}
		if maskedTrader.QwenKey != "" {
			maskedTrader.QwenKey = maskString(maskedTrader.QwenKey)
		}
		if maskedTrader.DeepSeekKey != "" {
			maskedTrader.DeepSeekKey = maskString(maskedTrader.DeepSeekKey)
		}
		if maskedTrader.CustomAPIKey != "" {
			maskedTrader.CustomAPIKey = maskString(maskedTrader.CustomAPIKey)
		}
		
		masked.Traders[i] = maskedTrader
	}
	
	// 保留MarketData配置（深拷贝）
	if c.MarketData.Klines != nil {
		masked.MarketData.Klines = make([]KlineConfig, len(c.MarketData.Klines))
		copy(masked.MarketData.Klines, c.MarketData.Klines)
	}
	
	return &masked
}

// maskString 脱敏字符串（只显示前4和后4个字符）
func maskString(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
