package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"nofx/database/models"
	"github.com/gin-gonic/gin"
)

// handleGenerateAILearning 让AI分析历史交易并生成学习总结
func (s *Server) handleGenerateAILearning(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	logger := trader.GetDecisionLogger()
	db := logger.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库未初始化"})
		return
	}

	// 获取历史交易数据进行分析
	tradeOutcomes, err := db.GetTradeOutcomes(50) // 分析最近50笔交易
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取交易数据失败: %v", err)})
		return
	}

	if len(tradeOutcomes) < 5 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "交易数据不足，至少需要5笔完成的交易才能进行AI分析",
		})
		return
	}

	// 获取最近的决策记录用于分析决策质量
	decisionRecords, err := db.GetLatestRecords(30) // 最近30条决策记录
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取决策记录失败: %v", err)})
		return
	}

	// 构建AI分析的系统提示词
	systemPrompt := `你是专业的交易分析AI，负责分析历史交易数据并生成学习总结。

请基于提供的交易数据，分析以下方面：
1. **成功模式识别**：什么样的市场条件、技术指标组合、持仓时长等导致了盈利交易
2. **失败模式识别**：什么情况下容易亏损，常见的错误决策模式
3. **风险管理评估**：止损执行情况、仓位管理、风险回报比分析
4. **时间维度分析**：不同时段的交易表现，持仓时长对收益的影响
5. **币种偏好分析**：哪些币种表现更好，为什么
6. **改进建议**：基于数据分析提出具体的策略优化建议

输出格式要求：
- 使用markdown格式
- 结构清晰，包含数据支撑
- 重点突出关键发现和可执行的改进建议
- 控制在800-1200字以内`

	// 构建用户提示词（包含交易数据）
	userPrompt := buildLearningAnalysisPrompt(tradeOutcomes, decisionRecords)

	// 调用AI进行分析
	aiResponse, err := trader.CallAI(systemPrompt, userPrompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("AI分析失败: %v", err)})
		return
	}

	// 计算统计数据
	stats := calculateTradeStatistics(tradeOutcomes)

	// 保存AI学习总结到数据库
	summary := &models.AILearningSummary{
		TraderID:       traderID,
		SummaryContent: aiResponse,
		TradesCount:    len(tradeOutcomes),
		DateRangeStart: stats.DateRangeStart,
		DateRangeEnd:   stats.DateRangeEnd,
		WinRate:        stats.WinRate,
		AvgPnL:         stats.AvgPnL,
		IsActive:       true,
	}

	err = db.SaveAILearningSummary(summary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("保存学习总结失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message":         "AI学习总结生成成功",
		"summary_content": aiResponse,
		"trades_analyzed": len(tradeOutcomes),
		"win_rate":        stats.WinRate,
		"avg_pnl":         stats.AvgPnL,
		"date_range":      fmt.Sprintf("%s ~ %s", stats.DateRangeStart, stats.DateRangeEnd),
	})
}

// handleGetAILearningSummary 获取当前激活的AI学习总结
func (s *Server) handleGetAILearningSummary(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	logger := trader.GetDecisionLogger()
	db := logger.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库未初始化"})
		return
	}

	summary, err := db.GetActiveAILearningSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取总结失败: %v", err)})
		return
	}

	if summary == nil {
		c.JSON(http.StatusOK, gin.H{"has_summary": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"has_summary":     true,
		"summary_content": summary.SummaryContent,
		"trades_count":    summary.TradesCount,
		"win_rate":        summary.WinRate,
		"avg_pnl":         summary.AvgPnL,
		"date_range":      fmt.Sprintf("%s ~ %s", summary.DateRangeStart, summary.DateRangeEnd),
		"created_at":      summary.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}

// TradeStatistics 交易统计数据
type TradeStatistics struct {
	DateRangeStart string
	DateRangeEnd   string
	WinRate        float64
	AvgPnL         float64
}

// buildLearningAnalysisPrompt 构建AI学习分析的用户提示词
func buildLearningAnalysisPrompt(tradeOutcomes []*models.TradeOutcome, decisionRecords []*models.DecisionRecord) string {
	var prompt strings.Builder
	
	prompt.WriteString("## 历史交易数据分析\n\n")
	prompt.WriteString(fmt.Sprintf("### 交易结果数据 (共%d笔交易)\n", len(tradeOutcomes)))
	
	for i, trade := range tradeOutcomes {
		if i >= 20 { // 限制显示前20笔交易，避免提示词过长
			prompt.WriteString(fmt.Sprintf("... 还有%d笔交易数据\n", len(tradeOutcomes)-20))
			break
		}
		
		// 计算持仓时长
		duration := time.Duration(trade.DurationMinutes) * time.Minute
		
		prompt.WriteString(fmt.Sprintf("**交易%d**: %s %s | 盈亏: %.2f USDT (%.2f%%) | 持仓时长: %s | 开仓: %.4f | 平仓: %.4f\n",
			i+1, trade.Symbol, trade.Side, trade.PnL, trade.PnLPct, 
			formatDuration(duration), trade.OpenPrice, trade.ClosePrice))
		
		if trade.EntryReason != "" {
			prompt.WriteString(fmt.Sprintf("  开仓理由: %s\n", trade.EntryReason))
		}
		if trade.ExitReason != "" {
			prompt.WriteString(fmt.Sprintf("  平仓理由: %s\n", trade.ExitReason))
		}
		prompt.WriteString("\n")
	}
	
	prompt.WriteString(fmt.Sprintf("\n### 最近决策记录 (共%d条)\n", len(decisionRecords)))
	for i, record := range decisionRecords {
		if i >= 10 { // 限制显示前10条决策记录
			prompt.WriteString(fmt.Sprintf("... 还有%d条决策记录\n", len(decisionRecords)-10))
			break
		}
		
		prompt.WriteString(fmt.Sprintf("**决策%d**: 周期%d | 时间: %s | 成功: %t\n",
			i+1, record.CycleNumber, record.Timestamp.Format("01-02 15:04"), record.Success))
		
		if record.ErrorMessage != "" {
			prompt.WriteString(fmt.Sprintf("  错误: %s\n", record.ErrorMessage))
		}
		prompt.WriteString("\n")
	}
	
	prompt.WriteString("\n请基于以上数据进行深入分析，识别成功和失败的模式，并提出具体的改进建议。")
	
	return prompt.String()
}

// calculateTradeStatistics 计算交易统计数据
func calculateTradeStatistics(tradeOutcomes []*models.TradeOutcome) TradeStatistics {
	if len(tradeOutcomes) == 0 {
		return TradeStatistics{}
	}
	
	var totalPnL float64
	var winCount int
	var earliestTime, latestTime time.Time
	
	for i, trade := range tradeOutcomes {
		totalPnL += trade.PnL
		if trade.PnL > 0 {
			winCount++
		}
		
		if i == 0 {
			earliestTime = trade.OpenTime
			latestTime = trade.CloseTime
		} else {
			if trade.OpenTime.Before(earliestTime) {
				earliestTime = trade.OpenTime
			}
			if trade.CloseTime.After(latestTime) {
				latestTime = trade.CloseTime
			}
		}
	}
	
	winRate := float64(winCount) / float64(len(tradeOutcomes)) * 100
	avgPnL := totalPnL / float64(len(tradeOutcomes))
	
	return TradeStatistics{
		DateRangeStart: earliestTime.Format("2006-01-02"),
		DateRangeEnd:   latestTime.Format("2006-01-02"),
		WinRate:        winRate,
		AvgPnL:         avgPnL,
	}
}

// formatDuration 格式化持仓时长
func formatDuration(duration time.Duration) string {
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
