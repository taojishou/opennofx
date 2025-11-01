package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleGenerateAILearning 让AI分析历史交易并生成学习总结
func (s *Server) handleGenerateAILearning(c *gin.Context) {
	// TODO: 实现AI学习总结生成逻辑
	// 需要通过trader的MCP客户端调用AI
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"message": "AI分析功能开发中",
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
