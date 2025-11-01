package api

import (
	"fmt"
	"log"
	"net/http"
	"time"
	
	"nofx/logger"

	"github.com/gin-gonic/gin"
)

// ManualClosePositionRequest 手动平仓请求
type ManualClosePositionRequest struct {
	TraderID string `json:"trader_id"`
	Symbol   string `json:"symbol"`
	Side     string `json:"side"` // "long" or "short"
}

// handleManualClosePosition 处理手动平仓请求
func (s *Server) handleManualClosePosition(c *gin.Context) {
	var req ManualClosePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数: " + err.Error(),
		})
		return
	}

	log.Printf("📤 收到手动平仓请求: Trader=%s, Symbol=%s, Side=%s", req.TraderID, req.Symbol, req.Side)

	// 获取指定的trader
	trader, err := s.traderManager.GetTrader(req.TraderID)
	if err != nil {
		log.Printf("❌ 获取Trader失败: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Trader不存在: " + req.TraderID,
		})
		return
	}

	// 获取平仓前的持仓信息用于记录到AI学习
	positions, _ := trader.GetPositions()
	var positionInfo struct {
		EntryPrice     float64
		MarkPrice      float64
		Quantity       float64
		Leverage       int
		UnrealizedPnL  float64
		PnLPercentage  float64
		MarginUsed     float64
	}
	
	for _, pos := range positions {
		if symbol, ok := pos["symbol"].(string); ok && symbol == req.Symbol {
			if side, ok := pos["side"].(string); ok && side == req.Side {
				if entry, ok := pos["entry_price"].(float64); ok {
					positionInfo.EntryPrice = entry
				}
				if mark, ok := pos["mark_price"].(float64); ok {
					positionInfo.MarkPrice = mark
				}
				if qty, ok := pos["quantity"].(float64); ok {
					positionInfo.Quantity = qty
				}
				if lev, ok := pos["leverage"].(int); ok {
					positionInfo.Leverage = lev
				}
				if margin, ok := pos["margin_used"].(float64); ok {
					positionInfo.MarginUsed = margin
				}
				if pnl, ok := pos["unrealized_pnl"].(float64); ok {
					positionInfo.UnrealizedPnL = pnl
				}
				if positionInfo.EntryPrice > 0 {
					positionInfo.PnLPercentage = (positionInfo.MarkPrice - positionInfo.EntryPrice) / positionInfo.EntryPrice * 100
				}
				break
			}
		}
	}

	// 调用trader的手动平仓方法
	err = trader.ManualClosePosition(req.Symbol, req.Side)
	if err != nil {
		log.Printf("❌ 手动平仓失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "平仓失败: " + err.Error(),
		})
		return
	}

	// 记录到历史成交表
	if positionInfo.EntryPrice > 0 && positionInfo.Quantity > 0 {
		// 计算盈亏
		pnl := 0.0
		if req.Side == "long" {
			pnl = (positionInfo.MarkPrice - positionInfo.EntryPrice) * positionInfo.Quantity
		} else {
			pnl = (positionInfo.EntryPrice - positionInfo.MarkPrice) * positionInfo.Quantity
		}
		
		// 计算盈亏百分比和其他信息
		positionValue := positionInfo.Quantity * positionInfo.EntryPrice
		
		// 使用保证金计算盈亏百分比（更准确）
		marginUsed := positionInfo.MarginUsed
		if marginUsed == 0 && positionInfo.Leverage > 0 {
			marginUsed = positionValue / float64(positionInfo.Leverage)
		}
		
		pnlPct := 0.0
		if marginUsed > 0 {
			pnlPct = (pnl / marginUsed) * 100
		}
		
		// 从AutoTrader获取真实的开仓时间
		closeTime := time.Now()
		openTime := closeTime.Add(-30 * time.Minute) // 默认值：30分钟前
		durationMinutes := int64(30)                  // 默认持仓30分钟
		
		// 尝试获取真实的开仓时间
		if realOpenTime, exists := trader.GetPositionOpenTime(req.Symbol, req.Side); exists {
			openTime = realOpenTime
			durationMinutes = int64(closeTime.Sub(openTime).Minutes())
			if durationMinutes < 0 {
				durationMinutes = 0
			}
			log.Printf("📅 获取到真实开仓时间: %s, 持仓时长: %d分钟", openTime.Format("15:04:05"), durationMinutes)
		} else {
			log.Printf("⚠️  无法获取开仓时间，使用默认值: 30分钟前")
		}
		
		// 判断退出原因
		exitReason := "手动平仓"
		isPremature := durationMinutes < 45 // 小于45分钟认为是过早平仓
		
		// 失败原因分析
		failureType := ""
		if pnl < 0 {
			if isPremature {
				failureType = "手动平仓（可能过早）+ 亏损"
			} else {
				failureType = "手动平仓 + 亏损"
			}
		}
		
		// 开仓原因（根据是否获取到真实时间）
		entryReason := "AI自动开仓"
		if _, exists := trader.GetPositionOpenTime(req.Symbol, req.Side); !exists {
			entryReason = "历史持仓（系统重启前开仓）"
		}
		
		// 构建交易记录
		trade := &logger.TradeOutcome{
			Symbol:          req.Symbol,
			Side:            req.Side,
			Quantity:        positionInfo.Quantity,
			Leverage:        positionInfo.Leverage,
			OpenPrice:       positionInfo.EntryPrice,
			ClosePrice:      positionInfo.MarkPrice,
			PositionValue:   positionValue,
			MarginUsed:      marginUsed,
			PnL:             pnl,
			PnLPct:          pnlPct,
			DurationMinutes: durationMinutes,
			OpenTime:        openTime,
			CloseTime:       closeTime,
			WasStopLoss:     false, // 手动平仓不是止损
			EntryReason:     entryReason,
			ExitReason:      exitReason,
			IsPremature:     isPremature,
			FailureType:     failureType,
		}
		
		// 保存到数据库
		if err := trader.GetDecisionLogger().SaveTradeOutcome(trade); err != nil {
			log.Printf("⚠️ 保存交易记录失败: %v", err)
		} else {
			log.Printf("📝 已记录到历史成交表: PnL=%+.2f USDT (%.2f%%), 杠杆=%dx", pnl, pnlPct, positionInfo.Leverage)
		}
	}

	// 记录手动平仓到AI学习系统
	if positionInfo.EntryPrice > 0 {
		account := trader.GetStatus()
		reasoning := fmt.Sprintf("🖐️ 手动平仓操作\n持仓信息: 入场价 %.4f, 标记价 %.4f, 数量 %.4f\n未实现盈亏: %.2f USDT (%.2f%%)\n\n这是一次手动干预，AI应该分析：\n1. 为什么需要人工介入？\n2. 当前持仓是否有明显问题？\n3. 如何在未来自动识别类似情况？",
			positionInfo.EntryPrice,
			positionInfo.MarkPrice,
			positionInfo.Quantity,
			positionInfo.UnrealizedPnL,
			positionInfo.PnLPercentage)
		
		// 获取账户余额
		var totalEquity, availBalance, unrealizedPnL float64
		if balance, ok := account["total_equity"].(float64); ok {
			totalEquity = balance
		}
		if avail, ok := account["available_balance"].(float64); ok {
			availBalance = avail
		}
		if pnl, ok := account["total_unrealized_pnl"].(float64); ok {
			unrealizedPnL = pnl
		}
		
		// 构建决策记录
		record := &logger.DecisionRecord{
			CoTTrace:     reasoning,
			DecisionJSON: fmt.Sprintf(`{"action":"close","symbol":"%s","side":"%s","reason":"manual"}`, req.Symbol, req.Side),
			AccountState: logger.AccountSnapshot{
				TotalBalance:          totalEquity,
				AvailableBalance:      availBalance,
				TotalUnrealizedProfit: unrealizedPnL,
			},
			Decisions: []logger.DecisionAction{
				{
					Action:    fmt.Sprintf("close_%s", req.Side),
					Symbol:    req.Symbol,
					Quantity:  positionInfo.Quantity,
					Price:     positionInfo.MarkPrice,
					Timestamp: time.Now(),
					Success:   true,
				},
			},
			Success: true,
		}
		trader.GetDecisionLogger().LogDecision(record)
		log.Printf("📝 已记录手动平仓到AI学习系统")
	}

	log.Printf("✅ 手动平仓成功: %s %s", req.Symbol, req.Side)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "平仓成功，已记录到AI学习系统",
		"trader":  req.TraderID,
		"symbol":  req.Symbol,
		"side":    req.Side,
	})
}

// handleToggleTrader 启用/停止Trader
func (s *Server) handleToggleTrader(c *gin.Context) {
	traderID := c.Query("trader_id")
	action := c.Query("action") // "start" or "stop"

	if traderID == "" || action == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "缺少trader_id或action参数",
		})
		return
	}

	log.Printf("🔄 收到Trader控制请求: Trader=%s, Action=%s", traderID, action)

	// 获取指定的trader
	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		log.Printf("❌ 获取Trader失败: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Trader不存在: " + traderID,
		})
		return
	}

	var message string
	switch action {
	case "start":
		trader.Resume()
		message = "Trader已启动"
		log.Printf("✅ Trader已启动: %s", traderID)
	case "stop":
		trader.Pause()
		message = "Trader已暂停"
		log.Printf("⏸️  Trader已暂停: %s", traderID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的action参数，必须是start或stop",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"trader":  traderID,
		"action":  action,
	})
}
