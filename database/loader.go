package database

import (
	"encoding/json"
	"fmt"
	"nofx/config"
	"nofx/database/repositories"
	"os"
)

// LoadConfigFromDB 从数据库加载配置
func LoadConfigFromDB() (*config.Config, error) {
	// 确保数据目录存在
	if err := ensureDataDirectory(); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 连接系统数据库
	sysConn, err := NewSystemConnection()
	if err != nil {
		return nil, fmt.Errorf("连接系统数据库失败: %w", err)
	}
	defer sysConn.Close()

	// 创建仓储
	traderRepo := repositories.NewTraderConfigRepository(sysConn.DB())
	sysConfigRepo := repositories.NewSystemConfigRepository(sysConn.DB())

	// 加载系统配置
	cfg := &config.Config{}

	// 加载API端口
	if apiPort, err := sysConfigRepo.Get("api_server_port"); err == nil {
		var port int
		if err := json.Unmarshal([]byte(apiPort.Value), &port); err == nil {
			cfg.APIServerPort = port
		}
	}
	if cfg.APIServerPort == 0 {
		cfg.APIServerPort = 8080 // 默认值
	}

	// 加载市场数据配置
	if coinPoolURL, err := sysConfigRepo.Get("coin_pool_api_url"); err == nil {
		cfg.CoinPoolAPIURL = coinPoolURL.Value
	}
	if oiTopURL, err := sysConfigRepo.Get("oi_top_api_url"); err == nil {
		cfg.OITopAPIURL = oiTopURL.Value
	}

	// 加载默认币种配置
	if useDefault, err := sysConfigRepo.Get("use_default_coins"); err == nil {
		var use bool
		json.Unmarshal([]byte(useDefault.Value), &use)
		cfg.UseDefaultCoins = use
	}
	
	if defaultCoins, err := sysConfigRepo.Get("default_coins"); err == nil {
		var coins []string
		if err := json.Unmarshal([]byte(defaultCoins.Value), &coins); err == nil {
			cfg.DefaultCoins = coins
		}
	}
	if len(cfg.DefaultCoins) == 0 {
		cfg.DefaultCoins = []string{
			"BTCUSDT", "ETHUSDT", "SOLUSDT", "BNBUSDT",
			"XRPUSDT", "DOGEUSDT", "ADAUSDT", "HYPEUSDT",
		}
	}

	// 加载K线配置
	if klineSettings, err := sysConfigRepo.Get("kline_settings"); err == nil {
		var klines []config.KlineConfig
		if err := json.Unmarshal([]byte(klineSettings.Value), &klines); err == nil {
			cfg.MarketData.Klines = klines
		}
	}
	if len(cfg.MarketData.Klines) == 0 {
		cfg.MarketData.Klines = []config.KlineConfig{
			{Interval: "3m", Limit: 20, ShowTable: true},
			{Interval: "4h", Limit: 60, ShowTable: false},
		}
	}

	// 从第一个启用的trader加载全局配置（保持向后兼容）
	enabledTraders, err := traderRepo.GetAllEnabled()
	if err != nil {
		return nil, fmt.Errorf("获取交易员配置失败: %w", err)
	}

	if len(enabledTraders) == 0 {
		return nil, fmt.Errorf("没有启用的trader，请先在数据库中配置trader")
	}

	// 使用第一个trader的配置作为全局配置
	firstTrader := enabledTraders[0]
	cfg.MaxPositions = firstTrader.MaxPositions
	cfg.MaxDailyLoss = firstTrader.MaxDailyLoss
	cfg.MaxDrawdown = firstTrader.MaxDrawdown
	cfg.StopTradingMinutes = firstTrader.StopTradingMinutes
	cfg.EnableAILearning = firstTrader.EnableAILearning
	cfg.AILearnInterval = firstTrader.AILearnInterval
	cfg.AIAutonomyMode = firstTrader.AIAutonomyMode
	cfg.CompactMode = firstTrader.CompactMode
	cfg.Leverage = config.LeverageConfig{
		BTCETHLeverage:  firstTrader.BTCETHLeverage,
		AltcoinLeverage: firstTrader.AltcoinLeverage,
	}

	// 转换trader配置
	cfg.Traders = make([]config.TraderConfig, len(enabledTraders))
	for i, dbTrader := range enabledTraders {
		cfg.Traders[i] = config.TraderConfig{
			ID:                    dbTrader.TraderID,
			Name:                  dbTrader.Name,
			Enabled:               dbTrader.Enabled,
			AIModel:               dbTrader.AIModel,
			Exchange:              dbTrader.Exchange,
			BinanceAPIKey:         dbTrader.BinanceAPIKey,
			BinanceSecretKey:      dbTrader.BinanceSecretKey,
			HyperliquidPrivateKey: dbTrader.HyperliquidPrivateKey,
			HyperliquidWalletAddr: dbTrader.HyperliquidWalletAddr,
			HyperliquidTestnet:    dbTrader.HyperliquidTestnet,
			AsterUser:             dbTrader.AsterUser,
			AsterSigner:           dbTrader.AsterSigner,
			AsterPrivateKey:       dbTrader.AsterPrivateKey,
			QwenKey:               dbTrader.QwenKey,
			DeepSeekKey:           dbTrader.DeepSeekKey,
			CustomAPIURL:          dbTrader.CustomAPIURL,
			CustomAPIKey:          dbTrader.CustomAPIKey,
			CustomModelName:       dbTrader.CustomModelName,
			InitialBalance:        dbTrader.InitialBalance,
			ScanIntervalMinutes:   dbTrader.ScanIntervalMinutes,
			AIAutonomyMode:        dbTrader.AIAutonomyMode,
			CompactMode:           dbTrader.CompactMode,
		}
	}

	return cfg, nil
}

// ensureDataDirectory 确保数据目录存在
func ensureDataDirectory() error {
	return os.MkdirAll("data", 0755)
}
