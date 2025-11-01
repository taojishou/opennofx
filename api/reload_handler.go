package api

import (
	"log"
	"nofx/config"
	"nofx/market"

	"github.com/gin-gonic/gin"
)

// handleReloadConfig 热重载配置
func (s *Server) handleReloadConfig(c *gin.Context) {
	log.Println("🔄 收到热重载请求...")

	// 1. 重新读取config.json
	newConfig, err := config.LoadConfig(config.GetConfigFilePath())
	if err != nil {
		log.Printf("❌ 加载配置失败: %v\n", err)
		c.JSON(500, gin.H{
			"success": false,
			"error":   "加载配置文件失败: " + err.Error(),
		})
		return
	}

	// 2. 更新市场数据K线配置
	if len(newConfig.MarketData.Klines) > 0 {
		klineSettings := make([]market.KlineSettings, len(newConfig.MarketData.Klines))
		for i, kline := range newConfig.MarketData.Klines {
			klineSettings[i] = market.KlineSettings{
				Interval:  kline.Interval,
				Limit:     kline.Limit,
				ShowTable: kline.ShowTable,
			}
		}
		market.SetKlineSettings(klineSettings)
		log.Printf("✓ K线配置已热重载: %d个时间框架", len(klineSettings))
	}

	// 3. 调用TraderManager的ReloadConfig方法
	err = s.traderManager.ReloadConfig(newConfig)
	if err != nil {
		log.Printf("❌ 热重载失败: %v\n", err)
		c.JSON(500, gin.H{
			"success": false,
			"error":   "热重载失败: " + err.Error(),
		})
		return
	}

	log.Println("✅ 热重载成功")

	// 3. 返回成功响应
	c.JSON(200, gin.H{
		"success": true,
		"message": "配置已热重载，无需重启服务",
		"traders": len(newConfig.Traders),
	})
}
