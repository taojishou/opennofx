package manager

import (
	"fmt"
	"log"
	"nofx/config"
	"nofx/trader"
	"sync"
	"time"
)

// TraderManager 管理多个trader实例
type TraderManager struct {
	traders map[string]*trader.AutoTrader // key: trader ID
	mu      sync.RWMutex
}

// NewTraderManager 创建trader管理器
func NewTraderManager() *TraderManager {
	return &TraderManager{
		traders: make(map[string]*trader.AutoTrader),
	}
}

// AddTrader 添加一个trader
func (tm *TraderManager) AddTrader(cfg config.TraderConfig, coinPoolURL string, maxDailyLoss, maxDrawdown float64, stopTradingMinutes int, leverage config.LeverageConfig, maxPositions int, enableAILearning bool, aiLearnInterval int) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.traders[cfg.ID]; exists {
		return fmt.Errorf("trader ID '%s' 已存在", cfg.ID)
	}

	// 构建AutoTraderConfig
	traderConfig := trader.AutoTraderConfig{
		ID:                    cfg.ID,
		Name:                  cfg.Name,
		AIModel:               cfg.AIModel,
		Exchange:              cfg.Exchange,
		BinanceAPIKey:         cfg.BinanceAPIKey,
		BinanceSecretKey:      cfg.BinanceSecretKey,
		HyperliquidPrivateKey: cfg.HyperliquidPrivateKey,
		HyperliquidWalletAddr: cfg.HyperliquidWalletAddr,
		HyperliquidTestnet:    cfg.HyperliquidTestnet,
		AsterUser:             cfg.AsterUser,
		AsterSigner:           cfg.AsterSigner,
		AsterPrivateKey:       cfg.AsterPrivateKey,
		CoinPoolAPIURL:        coinPoolURL,
		UseQwen:               cfg.AIModel == "qwen",
		DeepSeekKey:           cfg.DeepSeekKey,
		QwenKey:               cfg.QwenKey,
		CustomAPIURL:          cfg.CustomAPIURL,
		CustomAPIKey:          cfg.CustomAPIKey,
		CustomModelName:       cfg.CustomModelName,
		ScanInterval:          cfg.GetScanInterval(),
		InitialBalance:        cfg.InitialBalance,
		BTCETHLeverage:        leverage.BTCETHLeverage,  // 使用配置的杠杆倍数
		AltcoinLeverage:       leverage.AltcoinLeverage, // 使用配置的杠杆倍数
		MaxPositions:          maxPositions,             // 使用配置的最大持仓数
		EnableAILearning:      enableAILearning,         // AI学习开关
		AILearnInterval:       aiLearnInterval,          // AI学习间隔
		MaxDailyLoss:          maxDailyLoss,
		MaxDrawdown:           maxDrawdown,
		StopTradingTime:       time.Duration(stopTradingMinutes) * time.Minute,
	}

	// 创建trader实例
	at, err := trader.NewAutoTrader(traderConfig)
	if err != nil {
		return fmt.Errorf("创建trader失败: %w", err)
	}

	tm.traders[cfg.ID] = at
	log.Printf("✓ Trader '%s' (%s) 已添加", cfg.Name, cfg.AIModel)
	return nil
}

// GetTrader 获取指定ID的trader
func (tm *TraderManager) GetTrader(id string) (*trader.AutoTrader, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	t, exists := tm.traders[id]
	if !exists {
		return nil, fmt.Errorf("trader ID '%s' 不存在", id)
	}
	return t, nil
}

// GetAllTraders 获取所有trader
func (tm *TraderManager) GetAllTraders() map[string]*trader.AutoTrader {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	result := make(map[string]*trader.AutoTrader)
	for id, t := range tm.traders {
		result[id] = t
	}
	return result
}

// GetTraderIDs 获取所有trader ID列表
func (tm *TraderManager) GetTraderIDs() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	ids := make([]string, 0, len(tm.traders))
	for id := range tm.traders {
		ids = append(ids, id)
	}
	return ids
}

// StartAll 启动所有trader
func (tm *TraderManager) StartAll() {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	log.Println("🚀 启动所有Trader...")
	for id, t := range tm.traders {
		go func(traderID string, at *trader.AutoTrader) {
			log.Printf("▶️  启动 %s...", at.GetName())
			if err := at.Run(); err != nil {
				log.Printf("❌ %s 运行错误: %v", at.GetName(), err)
			}
		}(id, t)
	}
}

// StopAll 停止所有trader
func (tm *TraderManager) StopAll() {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	log.Println("⏹  停止所有Trader...")
	for _, t := range tm.traders {
		t.Stop()
	}
}

// GetComparisonData 获取对比数据
func (tm *TraderManager) GetComparisonData() (map[string]interface{}, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	comparison := make(map[string]interface{})
	traders := make([]map[string]interface{}, 0, len(tm.traders))

	for _, t := range tm.traders {
		account, err := t.GetAccountInfo()
		if err != nil {
			continue
		}

		status := t.GetStatus()
		isPaused := t.IsPaused()

		traders = append(traders, map[string]interface{}{
			"trader_id":       t.GetID(),
			"trader_name":     t.GetName(),
			"ai_model":        t.GetAIModel(),
			"exchange":        status["exchange"],
			"total_equity":    account["total_equity"],
			"total_pnl":       account["total_pnl"],
			"total_pnl_pct":   account["total_pnl_pct"],
			"position_count":  account["position_count"],
			"margin_used_pct": account["margin_used_pct"],
			"call_count":      status["call_count"],
			"is_running":      status["is_running"].(bool) && !isPaused,
			"is_paused":       isPaused,
		})
	}

	comparison["traders"] = traders
	comparison["count"] = len(traders)

	return comparison, nil
}

// ReloadConfig 热重载配置
func (tm *TraderManager) ReloadConfig(newConfig *config.Config) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	log.Println("🔄 开始热重载配置...")

	// 1. 记录现有traders
	oldTraders := make(map[string]*trader.AutoTrader)
	for id, t := range tm.traders {
		oldTraders[id] = t
	}

	// 2. 构建币种池URL
	coinPoolURL := ""
	if newConfig.UseDefaultCoins {
		coinPoolURL = newConfig.CoinPoolAPIURL
	} else {
		coinPoolURL = newConfig.OITopAPIURL
	}

	// 3. 处理新配置中的每个trader
	newTraders := make(map[string]*trader.AutoTrader)
	
	for _, traderCfg := range newConfig.Traders {
		if !traderCfg.Enabled {
			log.Printf("⏸️  Trader '%s' 已禁用，跳过", traderCfg.ID)
			continue
		}

		// 如果trader已存在，保留它
		if existingTrader, exists := oldTraders[traderCfg.ID]; exists {
			log.Printf("✓ Trader '%s' 已存在，保留", traderCfg.ID)
			newTraders[traderCfg.ID] = existingTrader
			delete(oldTraders, traderCfg.ID)
		} else {
			// 创建新trader
			log.Printf("➕ 创建新Trader: %s", traderCfg.ID)
			err := tm.addTraderUnlocked(traderCfg, coinPoolURL, 
				newConfig.MaxDailyLoss, newConfig.MaxDrawdown, 
				newConfig.StopTradingMinutes, newConfig.Leverage, 
				newConfig.MaxPositions)
			if err != nil {
				log.Printf("❌ 创建Trader %s 失败: %v", traderCfg.ID, err)
				continue
			}
			newTraders[traderCfg.ID] = tm.traders[traderCfg.ID]
		}
	}

	// 4. 停止已删除的traders
	for id, t := range oldTraders {
		log.Printf("⏹  停止并删除Trader: %s", id)
		t.Stop()
	}

	// 5. 更新traders map
	tm.traders = newTraders

	log.Printf("✅ 热重载完成，当前活跃Traders: %d", len(tm.traders))
	return nil
}

// addTraderUnlocked 添加trader（不加锁版本，供ReloadConfig使用）
func (tm *TraderManager) addTraderUnlocked(cfg config.TraderConfig, coinPoolURL string, maxDailyLoss, maxDrawdown float64, stopTradingMinutes int, leverage config.LeverageConfig, maxPositions int) error {
	if _, exists := tm.traders[cfg.ID]; exists {
		return fmt.Errorf("trader ID '%s' 已存在", cfg.ID)
	}

	// 构建AutoTraderConfig
	traderConfig := trader.AutoTraderConfig{
		ID:                    cfg.ID,
		Name:                  cfg.Name,
		AIModel:               cfg.AIModel,
		Exchange:              cfg.Exchange,
		BinanceAPIKey:         cfg.BinanceAPIKey,
		BinanceSecretKey:      cfg.BinanceSecretKey,
		HyperliquidPrivateKey: cfg.HyperliquidPrivateKey,
		HyperliquidWalletAddr: cfg.HyperliquidWalletAddr,
		HyperliquidTestnet:    cfg.HyperliquidTestnet,
		AsterUser:             cfg.AsterUser,
		AsterSigner:           cfg.AsterSigner,
		AsterPrivateKey:       cfg.AsterPrivateKey,
		CoinPoolAPIURL:        coinPoolURL,
		UseQwen:               cfg.AIModel == "qwen",
		DeepSeekKey:           cfg.DeepSeekKey,
		QwenKey:               cfg.QwenKey,
		CustomAPIURL:          cfg.CustomAPIURL,
		CustomAPIKey:          cfg.CustomAPIKey,
		CustomModelName:       cfg.CustomModelName,
		ScanInterval:          cfg.GetScanInterval(),
		InitialBalance:        cfg.InitialBalance,
		BTCETHLeverage:        leverage.BTCETHLeverage,
		AltcoinLeverage:       leverage.AltcoinLeverage,
		MaxPositions:          maxPositions,
		MaxDailyLoss:          maxDailyLoss,
		MaxDrawdown:           maxDrawdown,
		StopTradingTime:       time.Duration(stopTradingMinutes) * time.Minute,
	}

	// 创建trader实例
	at, err := trader.NewAutoTrader(traderConfig)
	if err != nil {
		return fmt.Errorf("创建trader失败: %w", err)
	}

	tm.traders[cfg.ID] = at
	
	// 立即启动新trader
	go func() {
		if err := at.Run(); err != nil {
			log.Printf("❌ %s 运行错误: %v", at.GetName(), err)
		}
	}()

	return nil
}
