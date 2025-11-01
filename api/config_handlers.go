package api

import (
	"fmt"
	"log"
	"nofx/config"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	configMutex sync.RWMutex
)

// GetConfigHandler 获取完整配置（脱敏）
func (s *Server) handleGetConfig(c *gin.Context) {
	configMutex.RLock()
	defer configMutex.RUnlock()

	cfg, err := config.LoadConfig(config.GetConfigFilePath())
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("加载配置失败: %v", err)})
		return
	}

	// 脱敏敏感数据
	maskedConfig := cfg.MaskSensitiveData()

	c.JSON(200, gin.H{
		"success": true,
		"data":    maskedConfig,
	})
}

// UpdateGlobalConfigHandler 更新全局配置
func (s *Server) handleUpdateGlobalConfig(c *gin.Context) {
	configMutex.Lock()
	defer configMutex.Unlock()

	var req struct {
		UseDefaultCoins    *bool                   `json:"use_default_coins"`
		DefaultCoins       []string                `json:"default_coins"`
		CoinPoolAPIURL     *string                 `json:"coin_pool_api_url"`
		OITopAPIURL        *string                 `json:"oi_top_api_url"`
		MaxPositions       *int                    `json:"max_positions"`
		MaxDailyLoss       *float64                `json:"max_daily_loss"`
		MaxDrawdown        *float64                `json:"max_drawdown"`
		StopTradingMinutes *int                    `json:"stop_trading_minutes"`
		BTCETHLeverage     *int                    `json:"btc_eth_leverage"`
		AltcoinLeverage    *int                    `json:"altcoin_leverage"`
		EnableAILearning   *bool                   `json:"enable_ai_learning"`
		AILearnInterval    *int                    `json:"ai_learn_interval"`
		MarketData         *config.MarketDataConfig `json:"market_data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "请求参数错误"})
		return
	}

	cfg, err := config.LoadConfig(config.GetConfigFilePath())
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("加载配置失败: %v", err)})
		return
	}

	// 更新配置
	if req.UseDefaultCoins != nil {
		cfg.UseDefaultCoins = *req.UseDefaultCoins
	}
	if req.DefaultCoins != nil {
		cfg.DefaultCoins = req.DefaultCoins
	}
	if req.CoinPoolAPIURL != nil {
		cfg.CoinPoolAPIURL = *req.CoinPoolAPIURL
	}
	if req.OITopAPIURL != nil {
		cfg.OITopAPIURL = *req.OITopAPIURL
	}
	if req.MaxPositions != nil {
		cfg.MaxPositions = *req.MaxPositions
	}
	if req.MaxDailyLoss != nil {
		cfg.MaxDailyLoss = *req.MaxDailyLoss
	}
	if req.MaxDrawdown != nil {
		cfg.MaxDrawdown = *req.MaxDrawdown
	}
	if req.StopTradingMinutes != nil {
		cfg.StopTradingMinutes = *req.StopTradingMinutes
	}
	if req.BTCETHLeverage != nil {
		cfg.Leverage.BTCETHLeverage = *req.BTCETHLeverage
	}
	if req.AltcoinLeverage != nil {
		cfg.Leverage.AltcoinLeverage = *req.AltcoinLeverage
	}
	if req.EnableAILearning != nil {
		cfg.EnableAILearning = *req.EnableAILearning
	}
	if req.AILearnInterval != nil {
		cfg.AILearnInterval = *req.AILearnInterval
	}
	if req.MarketData != nil {
		cfg.MarketData = *req.MarketData
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("配置验证失败: %v", err)})
		return
	}

	// 保存配置
	if err := config.SaveConfig(config.GetConfigFilePath(), cfg); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("保存配置失败: %v", err)})
		return
	}

	log.Println("✓ 全局配置已更新（需要重启服务生效）")

	c.JSON(200, gin.H{
		"success": true,
		"message": "配置更新成功，请重启服务使配置生效",
	})
}

// UpdateTraderConfigHandler 更新单个Trader配置
func (s *Server) handleUpdateTraderConfig(c *gin.Context) {
	configMutex.Lock()
	defer configMutex.Unlock()

	var req config.TraderConfig

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "请求参数错误"})
		return
	}

	cfg, err := config.LoadConfig(config.GetConfigFilePath())
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("加载配置失败: %v", err)})
		return
	}

	// 查找并更新trader
	found := false
	for i, trader := range cfg.Traders {
		if trader.ID == req.ID {
			// 保留原密钥（如果新请求中的密钥是脱敏的）
			if req.BinanceAPIKey != "" && req.BinanceAPIKey != "****" && len(req.BinanceAPIKey) > 8 {
				trader.BinanceAPIKey = req.BinanceAPIKey
			}
			if req.BinanceSecretKey != "" && req.BinanceSecretKey != "****" && len(req.BinanceSecretKey) > 8 {
				trader.BinanceSecretKey = req.BinanceSecretKey
			}
			if req.HyperliquidPrivateKey != "" && req.HyperliquidPrivateKey != "****" && len(req.HyperliquidPrivateKey) > 8 {
				trader.HyperliquidPrivateKey = req.HyperliquidPrivateKey
			}
			if req.AsterPrivateKey != "" && req.AsterPrivateKey != "****" && len(req.AsterPrivateKey) > 8 {
				trader.AsterPrivateKey = req.AsterPrivateKey
			}
			if req.QwenKey != "" && req.QwenKey != "****" && len(req.QwenKey) > 8 {
				trader.QwenKey = req.QwenKey
			}
			if req.DeepSeekKey != "" && req.DeepSeekKey != "****" && len(req.DeepSeekKey) > 8 {
				trader.DeepSeekKey = req.DeepSeekKey
			}
			if req.CustomAPIKey != "" && req.CustomAPIKey != "****" && len(req.CustomAPIKey) > 8 {
				trader.CustomAPIKey = req.CustomAPIKey
			}

			// 更新其他字段
			trader.Name = req.Name
			trader.Enabled = req.Enabled
			trader.AIModel = req.AIModel
			trader.Exchange = req.Exchange
			trader.HyperliquidWalletAddr = req.HyperliquidWalletAddr
			trader.HyperliquidTestnet = req.HyperliquidTestnet
			trader.AsterUser = req.AsterUser
			trader.AsterSigner = req.AsterSigner
			trader.CustomAPIURL = req.CustomAPIURL
			trader.CustomModelName = req.CustomModelName
			trader.InitialBalance = req.InitialBalance
			trader.ScanIntervalMinutes = req.ScanIntervalMinutes

			cfg.Traders[i] = trader
			found = true
			break
		}
	}

	if !found {
		c.JSON(404, gin.H{"error": "Trader不存在"})
		return
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("配置验证失败: %v", err)})
		return
	}

	// 保存配置
	if err := config.SaveConfig(config.GetConfigFilePath(), cfg); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("保存配置失败: %v", err)})
		return
	}

	log.Printf("✓ Trader配置已更新: %s（需要重启服务生效）", req.ID)

	c.JSON(200, gin.H{
		"success": true,
		"message": "Trader配置更新成功，请重启服务使配置生效",
	})
}

// AddTraderHandler 添加新Trader
func (s *Server) handleAddTrader(c *gin.Context) {
	configMutex.Lock()
	defer configMutex.Unlock()

	var req config.TraderConfig

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "请求参数错误"})
		return
	}

	cfg, err := config.LoadConfig(config.GetConfigFilePath())
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("加载配置失败: %v", err)})
		return
	}

	// 检查ID是否已存在
	for _, trader := range cfg.Traders {
		if trader.ID == req.ID {
			c.JSON(400, gin.H{"error": "Trader ID已存在"})
			return
		}
	}

	// 添加新trader
	cfg.Traders = append(cfg.Traders, req)

	// 验证配置
	if err := cfg.Validate(); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("配置验证失败: %v", err)})
		return
	}

	// 保存配置
	if err := config.SaveConfig(config.GetConfigFilePath(), cfg); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("保存配置失败: %v", err)})
		return
	}

	log.Printf("✓ 新Trader已添加: %s（需要重启服务生效）", req.ID)

	c.JSON(200, gin.H{
		"success": true,
		"message": "Trader添加成功，请重启服务使配置生效",
	})
}

// DeleteTraderHandler 删除Trader
func (s *Server) handleDeleteTrader(c *gin.Context) {
	configMutex.Lock()
	defer configMutex.Unlock()

	traderID := c.Query("trader_id")
	if traderID == "" {
		c.JSON(400, gin.H{"error": "trader_id参数缺失"})
		return
	}

	cfg, err := config.LoadConfig(config.GetConfigFilePath())
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("加载配置失败: %v", err)})
		return
	}

	// 查找并删除trader
	found := false
	newTraders := make([]config.TraderConfig, 0, len(cfg.Traders))
	for _, trader := range cfg.Traders {
		if trader.ID != traderID {
			newTraders = append(newTraders, trader)
		} else {
			found = true
		}
	}

	if !found {
		c.JSON(404, gin.H{"error": "Trader不存在"})
		return
	}

	cfg.Traders = newTraders

	// 保存配置
	if err := config.SaveConfig(config.GetConfigFilePath(), cfg); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("保存配置失败: %v", err)})
		return
	}

	log.Printf("✓ Trader已删除: %s（需要重启服务生效）", traderID)

	c.JSON(200, gin.H{
		"success": true,
		"message": "Trader删除成功，请重启服务使配置生效",
	})
}
