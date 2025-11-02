package database

import (
	"encoding/json"
	"fmt"
	"log"
	"nofx/config"
	"nofx/database/models"
)

// MigrateFromConfigFile 从config.json迁移配置到数据库
func MigrateFromConfigFile(configFile string, manager *Manager) error {
	log.Printf("📦 开始从 %s 迁移配置到数据库...", configFile)

	// 加载config.json
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("加载配置文件失败: %w", err)
	}

	// 1. 迁移系统配置
	if err := migrateSystemConfigs(cfg, manager); err != nil {
		return fmt.Errorf("迁移系统配置失败: %w", err)
	}

	// 2. 迁移Trader配置
	if err := migrateTraderConfigs(cfg, manager); err != nil {
		return fmt.Errorf("迁移Trader配置失败: %w", err)
	}

	log.Println("✓ 配置迁移完成！")
	return nil
}

// migrateSystemConfigs 迁移系统配置
func migrateSystemConfigs(cfg *config.Config, manager *Manager) error {
	log.Println("  → 迁移系统配置...")

	// API服务器端口
	if cfg.APIServerPort > 0 {
		err := manager.SystemConfigRepo.Set(
			"api_server_port",
			fmt.Sprintf("%d", cfg.APIServerPort),
			"API服务器端口",
			"api",
		)
		if err != nil {
			return err
		}
	}

	// 币种池配置
	if cfg.CoinPoolAPIURL != "" {
		err := manager.SystemConfigRepo.Set(
			"coin_pool_api_url",
			cfg.CoinPoolAPIURL,
			"币种池API地址",
			"market",
		)
		if err != nil {
			return err
		}
	}

	if cfg.OITopAPIURL != "" {
		err := manager.SystemConfigRepo.Set(
			"oi_top_api_url",
			cfg.OITopAPIURL,
			"持仓量TopAPI地址",
			"market",
		)
		if err != nil {
			return err
		}
	}

	// 默认币种列表
	useDefaultCoins := "false"
	if cfg.UseDefaultCoins {
		useDefaultCoins = "true"
	}
	err := manager.SystemConfigRepo.Set(
		"use_default_coins",
		useDefaultCoins,
		"是否使用默认币种列表",
		"market",
	)
	if err != nil {
		return err
	}

	if len(cfg.DefaultCoins) > 0 {
		defaultCoinsJSON, _ := json.Marshal(cfg.DefaultCoins)
		err := manager.SystemConfigRepo.Set(
			"default_coins",
			string(defaultCoinsJSON),
			"默认币种列表",
			"market",
		)
		if err != nil {
			return err
		}
	}

	// K线配置
	if len(cfg.MarketData.Klines) > 0 {
		klineSettingsJSON, _ := json.Marshal(cfg.MarketData.Klines)
		err := manager.SystemConfigRepo.Set(
			"kline_settings",
			string(klineSettingsJSON),
			"K线配置",
			"market",
		)
		if err != nil {
			return err
		}
	}

	log.Println("  ✓ 系统配置迁移完成")
	return nil
}

// migrateTraderConfigs 迁移Trader配置
func migrateTraderConfigs(cfg *config.Config, manager *Manager) error {
	log.Printf("  → 迁移 %d 个Trader配置...", len(cfg.Traders))

	for i, traderCfg := range cfg.Traders {
		// 检查是否已存在
		existing, err := manager.TraderConfigRepo.GetByTraderID(traderCfg.ID)
		if err == nil && existing != nil {
			log.Printf("  ⏭️  Trader[%s] 已存在，跳过", traderCfg.ID)
			continue
		}

		// 创建新的Trader配置
		dbTraderCfg := &models.TraderConfig{
			UserID:              0, // 默认用户
			TraderID:            traderCfg.ID,
			Name:                traderCfg.Name,
			Enabled:             traderCfg.Enabled,
			AIModel:             traderCfg.AIModel,
			Exchange:            traderCfg.Exchange,
			BinanceAPIKey:       traderCfg.BinanceAPIKey,
			BinanceSecretKey:    traderCfg.BinanceSecretKey,
			HyperliquidPrivateKey: traderCfg.HyperliquidPrivateKey,
			HyperliquidWalletAddr: traderCfg.HyperliquidWalletAddr,
			HyperliquidTestnet:    traderCfg.HyperliquidTestnet,
			AsterUser:           traderCfg.AsterUser,
			AsterSigner:         traderCfg.AsterSigner,
			AsterPrivateKey:     traderCfg.AsterPrivateKey,
			DeepSeekKey:         traderCfg.DeepSeekKey,
			QwenKey:             traderCfg.QwenKey,
			CustomAPIURL:        traderCfg.CustomAPIURL,
			CustomAPIKey:        traderCfg.CustomAPIKey,
			CustomModelName:     traderCfg.CustomModelName,
			InitialBalance:      traderCfg.InitialBalance,
			ScanIntervalMinutes: traderCfg.ScanIntervalMinutes,
			MaxPositions:        cfg.MaxPositions,
			BTCETHLeverage:      cfg.Leverage.BTCETHLeverage,
			AltcoinLeverage:     cfg.Leverage.AltcoinLeverage,
			MaxDailyLoss:        cfg.MaxDailyLoss,
			MaxDrawdown:         cfg.MaxDrawdown,
			StopTradingMinutes:  cfg.StopTradingMinutes,
			EnableAILearning:    cfg.EnableAILearning,
			AILearnInterval:     cfg.AILearnInterval,
			AIAutonomyMode:      cfg.AIAutonomyMode,
		}

		_, err = manager.TraderConfigRepo.Create(dbTraderCfg)
		if err != nil {
			return fmt.Errorf("创建Trader[%d] %s 配置失败: %w", i, traderCfg.Name, err)
		}

		log.Printf("  ✓ Trader[%s] %s 配置已迁移", traderCfg.ID, traderCfg.Name)
	}

	log.Println("  ✓ Trader配置迁移完成")
	return nil
}

// LoadSystemConfig 从数据库加载系统配置
func LoadSystemConfig(manager *Manager) (*config.Config, error) {
	cfg := &config.Config{}

	// 加载API服务器端口
	if portCfg, err := manager.SystemConfigRepo.Get("api_server_port"); err == nil {
		fmt.Sscanf(portCfg.Value, "%d", &cfg.APIServerPort)
	}

	// 加载币种池配置
	if poolCfg, err := manager.SystemConfigRepo.Get("coin_pool_api_url"); err == nil {
		cfg.CoinPoolAPIURL = poolCfg.Value
	}

	if oiCfg, err := manager.SystemConfigRepo.Get("oi_top_api_url"); err == nil {
		cfg.OITopAPIURL = oiCfg.Value
	}

	// 加载默认币种配置
	if useCfg, err := manager.SystemConfigRepo.Get("use_default_coins"); err == nil {
		cfg.UseDefaultCoins = (useCfg.Value == "true")
	}

	if coinsCfg, err := manager.SystemConfigRepo.Get("default_coins"); err == nil {
		json.Unmarshal([]byte(coinsCfg.Value), &cfg.DefaultCoins)
	}

	// 加载K线配置
	if klineCfg, err := manager.SystemConfigRepo.Get("kline_settings"); err == nil {
		json.Unmarshal([]byte(klineCfg.Value), &cfg.MarketData.Klines)
	}

	// 加载所有启用的Trader配置
	traderConfigs, err := manager.TraderConfigRepo.GetAllEnabled()
	if err != nil {
		return nil, fmt.Errorf("加载Trader配置失败: %w", err)
	}

	// 转换为config.TraderConfig格式
	cfg.Traders = make([]config.TraderConfig, len(traderConfigs))
	for i, tc := range traderConfigs {
		cfg.Traders[i] = config.TraderConfig{
			ID:                    tc.TraderID,
			Name:                  tc.Name,
			Enabled:               tc.Enabled,
			AIModel:               tc.AIModel,
			Exchange:              tc.Exchange,
			BinanceAPIKey:         tc.BinanceAPIKey,
			BinanceSecretKey:      tc.BinanceSecretKey,
			HyperliquidPrivateKey: tc.HyperliquidPrivateKey,
			HyperliquidWalletAddr: tc.HyperliquidWalletAddr,
			HyperliquidTestnet:    tc.HyperliquidTestnet,
			AsterUser:             tc.AsterUser,
			AsterSigner:           tc.AsterSigner,
			AsterPrivateKey:       tc.AsterPrivateKey,
			QwenKey:               tc.QwenKey,
			DeepSeekKey:           tc.DeepSeekKey,
			CustomAPIURL:          tc.CustomAPIURL,
			CustomAPIKey:          tc.CustomAPIKey,
			CustomModelName:       tc.CustomModelName,
			InitialBalance:        tc.InitialBalance,
			ScanIntervalMinutes:   tc.ScanIntervalMinutes,
		}

		// 从第一个trader配置中提取全局配置
		if i == 0 {
			cfg.MaxPositions = tc.MaxPositions
			cfg.Leverage.BTCETHLeverage = tc.BTCETHLeverage
			cfg.Leverage.AltcoinLeverage = tc.AltcoinLeverage
			cfg.MaxDailyLoss = tc.MaxDailyLoss
			cfg.MaxDrawdown = tc.MaxDrawdown
			cfg.StopTradingMinutes = tc.StopTradingMinutes
			cfg.EnableAILearning = tc.EnableAILearning
			cfg.AILearnInterval = tc.AILearnInterval
			cfg.AIAutonomyMode = tc.AIAutonomyMode
		}
	}

	return cfg, nil
}
