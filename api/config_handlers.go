package api

import (
	"encoding/json"
	"fmt"
	"log"
	"nofx/config"
	"nofx/database"
	"nofx/database/models"
	"nofx/database/repositories"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	configMutex sync.RWMutex
)

// isMaskedKey 检查密钥是否是脱敏后的值
// 脱敏格式: "xxxx****xxxx" 或 "****"
func isMaskedKey(key string) bool {
	return key == "****" || len(key) > 4 && key[len(key)/2-2:len(key)/2+2] == "****"
}

// handleGetConfig 获取完整配置（脱敏）- 从数据库加载
func (s *Server) handleGetConfig(c *gin.Context) {
	configMutex.RLock()
	defer configMutex.RUnlock()

	// 从数据库加载配置
	cfg, err := database.LoadConfigFromDB()
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

// handleUpdateGlobalConfig 更新全局配置 - 更新到数据库
func (s *Server) handleUpdateGlobalConfig(c *gin.Context) {
	configMutex.Lock()
	defer configMutex.Unlock()

	var req struct {
		UseDefaultCoins    *bool                    `json:"use_default_coins"`
		DefaultCoins       []string                 `json:"default_coins"`
		CoinPoolAPIURL     *string                  `json:"coin_pool_api_url"`
		OITopAPIURL        *string                  `json:"oi_top_api_url"`
		MaxPositions       *int                     `json:"max_positions"`
		MaxDailyLoss       *float64                 `json:"max_daily_loss"`
		MaxDrawdown        *float64                 `json:"max_drawdown"`
		StopTradingMinutes *int                     `json:"stop_trading_minutes"`
		BTCETHLeverage     *int                     `json:"btc_eth_leverage"`
		AltcoinLeverage    *int                     `json:"altcoin_leverage"`
		EnableAILearning   *bool                    `json:"enable_ai_learning"`
		AILearnInterval    *int                     `json:"ai_learn_interval"`
		AIAutonomyMode     *bool                    `json:"ai_autonomy_mode"`
		CompactMode        *bool                    `json:"compact_mode"`
		MarketData         *config.MarketDataConfig `json:"market_data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "请求参数错误"})
		return
	}

	// 连接系统数据库
	sysConn, err := database.NewSystemConnection()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("连接数据库失败: %v", err)})
		return
	}
	defer sysConn.Close()

	repo := repositories.NewSystemConfigRepository(sysConn.DB())

	// 更新系统配置到数据库
	if req.UseDefaultCoins != nil {
		val := fmt.Sprintf("%v", *req.UseDefaultCoins)
		repo.Set("use_default_coins", val, "是否使用默认币种列表", "market")
	}
	if req.DefaultCoins != nil {
		jsonData, _ := json.Marshal(req.DefaultCoins)
		repo.Set("default_coins", string(jsonData), "默认币种列表", "market")
	}
	if req.CoinPoolAPIURL != nil {
		repo.Set("coin_pool_api_url", *req.CoinPoolAPIURL, "币种池API地址", "market")
	}
	if req.OITopAPIURL != nil {
		repo.Set("oi_top_api_url", *req.OITopAPIURL, "持仓量TopAPI地址", "market")
	}
	if req.MarketData != nil {
		jsonData, _ := json.Marshal(req.MarketData.Klines)
		repo.Set("kline_settings", string(jsonData), "K线配置", "market")
	}

	// 更新第一个trader的配置（全局配置）
	traderRepo := repositories.NewTraderConfigRepository(sysConn.DB())
	traders, err := traderRepo.GetAllEnabled()
	if err == nil && len(traders) > 0 {
		trader := traders[0]
		if req.MaxPositions != nil {
			trader.MaxPositions = *req.MaxPositions
		}
		if req.MaxDailyLoss != nil {
			trader.MaxDailyLoss = *req.MaxDailyLoss
		}
		if req.MaxDrawdown != nil {
			trader.MaxDrawdown = *req.MaxDrawdown
		}
		if req.StopTradingMinutes != nil {
			trader.StopTradingMinutes = *req.StopTradingMinutes
		}
		if req.BTCETHLeverage != nil {
			trader.BTCETHLeverage = *req.BTCETHLeverage
		}
		if req.AltcoinLeverage != nil {
			trader.AltcoinLeverage = *req.AltcoinLeverage
		}
		if req.EnableAILearning != nil {
			trader.EnableAILearning = *req.EnableAILearning
		}
		if req.AILearnInterval != nil {
			trader.AILearnInterval = *req.AILearnInterval
		}
		if req.AIAutonomyMode != nil {
			trader.AIAutonomyMode = *req.AIAutonomyMode
		}
		if req.CompactMode != nil {
			trader.CompactMode = *req.CompactMode
		}
		traderRepo.Update(trader)
	}

	log.Println("✓ 全局配置已更新")

	c.JSON(200, gin.H{
		"success": true,
		"message": "配置更新成功",
	})
}

// handleUpdateTraderConfig 更新单个Trader配置 - 更新到数据库
func (s *Server) handleUpdateTraderConfig(c *gin.Context) {
	configMutex.Lock()
	defer configMutex.Unlock()

	var req config.TraderConfig

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "请求参数错误"})
		return
	}

	// 连接系统数据库
	sysConn, err := database.NewSystemConnection()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("连接数据库失败: %v", err)})
		return
	}
	defer sysConn.Close()

	traderRepo := repositories.NewTraderConfigRepository(sysConn.DB())

	// 查找trader
	dbTrader, err := traderRepo.GetByTraderID(req.ID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Trader不存在"})
		return
	}

	// 保留原密钥（如果新请求中的密钥是脱敏的则不更新）
	// 脱敏格式: "xxxx****xxxx"，所以检查是否包含****
	if req.BinanceAPIKey != "" && !isMaskedKey(req.BinanceAPIKey) {
		dbTrader.BinanceAPIKey = req.BinanceAPIKey
	}
	if req.BinanceSecretKey != "" && !isMaskedKey(req.BinanceSecretKey) {
		dbTrader.BinanceSecretKey = req.BinanceSecretKey
	}
	if req.HyperliquidPrivateKey != "" && !isMaskedKey(req.HyperliquidPrivateKey) {
		dbTrader.HyperliquidPrivateKey = req.HyperliquidPrivateKey
	}
	if req.AsterPrivateKey != "" && !isMaskedKey(req.AsterPrivateKey) {
		dbTrader.AsterPrivateKey = req.AsterPrivateKey
	}
	if req.QwenKey != "" && !isMaskedKey(req.QwenKey) {
		dbTrader.QwenKey = req.QwenKey
	}
	if req.DeepSeekKey != "" && !isMaskedKey(req.DeepSeekKey) {
		dbTrader.DeepSeekKey = req.DeepSeekKey
	}
	if req.CustomAPIKey != "" && !isMaskedKey(req.CustomAPIKey) {
		dbTrader.CustomAPIKey = req.CustomAPIKey
	}

	// 打印接收到的数据用于调试
	log.Printf("[DEBUG] 接收到的Trader数据: ID=%s, AIAutonomyMode=%v, CompactMode=%v", 
		req.ID, req.AIAutonomyMode, req.CompactMode)
	
	// 更新其他字段
	dbTrader.Name = req.Name
	dbTrader.Enabled = req.Enabled
	dbTrader.AIModel = req.AIModel
	dbTrader.Exchange = req.Exchange
	dbTrader.HyperliquidWalletAddr = req.HyperliquidWalletAddr
	dbTrader.HyperliquidTestnet = req.HyperliquidTestnet
	dbTrader.AsterUser = req.AsterUser
	dbTrader.AsterSigner = req.AsterSigner
	dbTrader.CustomAPIURL = req.CustomAPIURL
	dbTrader.CustomModelName = req.CustomModelName
	dbTrader.InitialBalance = req.InitialBalance
	dbTrader.ScanIntervalMinutes = req.ScanIntervalMinutes
	dbTrader.AIAutonomyMode = req.AIAutonomyMode
	dbTrader.CompactMode = req.CompactMode

	// 更新到数据库
	if err := traderRepo.Update(dbTrader); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("更新失败: %v", err)})
		return
	}

	log.Printf("✓ Trader配置已更新: %s（需要重启服务生效）", req.ID)

	c.JSON(200, gin.H{
		"success": true,
		"message": "Trader配置更新成功，请重启服务使配置生效",
	})
}

// handleAddTrader 添加新Trader - 保存到数据库
func (s *Server) handleAddTrader(c *gin.Context) {
	configMutex.Lock()
	defer configMutex.Unlock()

	var req config.TraderConfig

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "请求参数错误"})
		return
	}

	// 连接系统数据库
	sysConn, err := database.NewSystemConnection()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("连接数据库失败: %v", err)})
		return
	}
	defer sysConn.Close()

	traderRepo := repositories.NewTraderConfigRepository(sysConn.DB())

	// 检查ID是否已存在
	_, err = traderRepo.GetByTraderID(req.ID)
	if err == nil {
		c.JSON(400, gin.H{"error": "Trader ID已存在"})
		return
	}

	// 转换为数据库模型
	dbTrader := &models.TraderConfig{
		UserID:                0, // 系统默认
		TraderID:              req.ID,
		Name:                  req.Name,
		Enabled:               req.Enabled,
		AIModel:               req.AIModel,
		Exchange:              req.Exchange,
		BinanceAPIKey:         req.BinanceAPIKey,
		BinanceSecretKey:      req.BinanceSecretKey,
		HyperliquidPrivateKey: req.HyperliquidPrivateKey,
		HyperliquidWalletAddr: req.HyperliquidWalletAddr,
		HyperliquidTestnet:    req.HyperliquidTestnet,
		AsterUser:             req.AsterUser,
		AsterSigner:           req.AsterSigner,
		AsterPrivateKey:       req.AsterPrivateKey,
		DeepSeekKey:           req.DeepSeekKey,
		QwenKey:               req.QwenKey,
		CustomAPIURL:          req.CustomAPIURL,
		CustomAPIKey:          req.CustomAPIKey,
		CustomModelName:       req.CustomModelName,
		InitialBalance:        req.InitialBalance,
		ScanIntervalMinutes:   req.ScanIntervalMinutes,
		MaxPositions:          3,
		BTCETHLeverage:        5,
		AltcoinLeverage:       5,
		MaxDailyLoss:          0,
		MaxDrawdown:           0,
		StopTradingMinutes:    0,
		EnableAILearning:      false,
		AILearnInterval:       10,
		AIAutonomyMode:        false,
		CompactMode:           true, // 默认启用紧凑模式
	}

	// 保存到数据库
	if _, err := traderRepo.Create(dbTrader); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("保存失败: %v", err)})
		return
	}

	log.Printf("✓ 新Trader已添加: %s（需要重启服务生效）", req.ID)

	c.JSON(200, gin.H{
		"success": true,
		"message": "Trader添加成功，请重启服务使配置生效",
	})
}

// handleDeleteTrader 删除Trader - 从数据库删除
func (s *Server) handleDeleteTrader(c *gin.Context) {
	configMutex.Lock()
	defer configMutex.Unlock()

	traderID := c.Query("trader_id")
	if traderID == "" {
		c.JSON(400, gin.H{"error": "trader_id参数缺失"})
		return
	}

	// 连接系统数据库
	sysConn, err := database.NewSystemConnection()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("连接数据库失败: %v", err)})
		return
	}
	defer sysConn.Close()

	traderRepo := repositories.NewTraderConfigRepository(sysConn.DB())

	// 查找trader
	dbTrader, err := traderRepo.GetByTraderID(traderID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Trader不存在"})
		return
	}

	// 删除
	if err := traderRepo.Delete(dbTrader.ID); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("删除失败: %v", err)})
		return
	}

	log.Printf("✓ Trader已删除: %s（需要重启服务生效）", traderID)

	c.JSON(200, gin.H{
		"success": true,
		"message": "Trader删除成功，请重启服务使配置生效",
	})
}
