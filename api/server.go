package api

import (
	"fmt"
	"log"
	"net/http"
	"nofx/database/models"
	"nofx/manager"

	"github.com/gin-gonic/gin"
)

// Server HTTP API服务器
type Server struct {
	router        *gin.Engine
	traderManager *manager.TraderManager
	port          int
}

// NewServer 创建API服务器
func NewServer(traderManager *manager.TraderManager, port int) *Server {
	// 设置为Release模式（减少日志输出）
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// 启用CORS
	router.Use(corsMiddleware())

	s := &Server{
		router:        router,
		traderManager: traderManager,
		port:          port,
	}

	// 设置路由
	s.setupRoutes()

	return s
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 健康检查
	s.router.Any("/health", s.handleHealth)

	// API路由组
	api := s.router.Group("/api")
	{
		// 竞赛总览
		api.GET("/competition", s.handleCompetition)

		// Trader列表
		api.GET("/traders", s.handleTraderList)

		// 指定trader的数据（使用query参数 ?trader_id=xxx）
		api.GET("/status", s.handleStatus)
		api.GET("/account", s.handleAccount)
		api.GET("/positions", s.handlePositions)
		api.GET("/decisions", s.handleDecisions)
		api.GET("/decisions/latest", s.handleLatestDecisions)
		api.GET("/statistics", s.handleStatistics)
		api.GET("/equity-history", s.handleEquityHistory)
		api.GET("/performance", s.handlePerformance)

		// Prompt配置相关路由（使用gin格式）
		api.GET("/prompts", s.handleGetPrompts)
		api.POST("/prompts/update", s.handleUpdatePrompt)
		api.POST("/prompts/toggle", s.handleTogglePrompt)
		api.POST("/prompts/add", s.handleAddPrompt)
		api.DELETE("/prompts/delete", s.handleDeletePrompt)
		api.GET("/prompts/preview", s.handlePreviewPrompt)

		// 系统配置管理路由（通用配置管理）
		api.GET("/config", s.handleGetConfig)
		api.POST("/config/global/update", s.handleUpdateGlobalConfig)
		api.POST("/config/trader/update", s.handleUpdateTraderConfig)
		api.POST("/config/trader/add", s.handleAddTrader)
		api.DELETE("/config/trader/delete", s.handleDeleteTrader)

		// 系统运行时配置API（风险阈值、技术指标等可配置参数）
		api.GET("/system/configs", s.handleGetSystemConfigs)              // 获取所有配置
		api.GET("/system/configs/:type", s.handleGetConfigByType)         // 按类型获取配置
		api.PUT("/system/configs", s.handleUpdateSystemConfig)            // 更新单个配置
		api.PUT("/system/configs/batch", s.handleBatchUpdateConfigs)      // 批量更新配置
		api.POST("/system/configs/:key/reset", s.handleResetConfig)       // 重置配置
		
		// 热重载路由
		api.POST("/config/reload", s.handleReloadConfig)
		
		// 交易控制路由
		api.POST("/trading/close-position", s.handleManualClosePosition)
		api.POST("/trading/toggle-trader", s.handleToggleTrader)
		
		// AI学习总结路由
		api.POST("/ai-learning/generate", s.handleGenerateAILearning)
		api.GET("/ai-learning/summary", s.handleGetAILearningSummary)
	}
}

// handleHealth 健康检查
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   c.Request.Context().Value("time"),
	})
}

// getTraderFromQuery 从query参数获取trader
func (s *Server) getTraderFromQuery(c *gin.Context) (*manager.TraderManager, string, error) {
	traderID := c.Query("trader_id")
	if traderID == "" {
		// 如果没有指定trader_id，返回第一个trader
		ids := s.traderManager.GetTraderIDs()
		if len(ids) == 0 {
			return nil, "", fmt.Errorf("没有可用的trader")
		}
		traderID = ids[0]
	}
	return s.traderManager, traderID, nil
}

// handleCompetition 竞赛总览（对比所有trader）
func (s *Server) handleCompetition(c *gin.Context) {
	comparison, err := s.traderManager.GetComparisonData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取对比数据失败: %v", err),
		})
		return
	}
	c.JSON(http.StatusOK, comparison)
}

// handleTraderList trader列表
func (s *Server) handleTraderList(c *gin.Context) {
	traders := s.traderManager.GetAllTraders()
	result := make([]map[string]interface{}, 0, len(traders))

	for _, t := range traders {
		result = append(result, map[string]interface{}{
			"trader_id":   t.GetID(),
			"trader_name": t.GetName(),
			"ai_model":    t.GetAIModel(),
		})
	}

	c.JSON(http.StatusOK, result)
}

// handleStatus 系统状态
func (s *Server) handleStatus(c *gin.Context) {
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

	status := trader.GetStatus()
	c.JSON(http.StatusOK, status)
}

// handleAccount 账户信息
func (s *Server) handleAccount(c *gin.Context) {
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

	log.Printf("📊 收到账户信息请求 [%s]", trader.GetName())
	account, err := trader.GetAccountInfo()
	if err != nil {
		log.Printf("❌ 获取账户信息失败 [%s]: %v", trader.GetName(), err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取账户信息失败: %v", err),
		})
		return
	}

	log.Printf("✓ 返回账户信息 [%s]: 净值=%.2f, 可用=%.2f, 盈亏=%.2f (%.2f%%)",
		trader.GetName(),
		account["total_equity"],
		account["available_balance"],
		account["total_pnl"],
		account["total_pnl_pct"])
	c.JSON(http.StatusOK, account)
}

// handlePositions 持仓列表
func (s *Server) handlePositions(c *gin.Context) {
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

	positions, err := trader.GetPositions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取持仓列表失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, positions)
}

// handleDecisions 决策日志列表
func (s *Server) handleDecisions(c *gin.Context) {
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

	// 获取所有历史决策记录（无限制）
	records, err := trader.GetDecisionLogger().GetLatestRecords(10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取决策日志失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, records)
}

// handleLatestDecisions 最新决策日志（最近5条，最新的在前）
func (s *Server) handleLatestDecisions(c *gin.Context) {
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

	records, err := trader.GetDecisionLogger().GetLatestRecords(5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取决策日志失败: %v", err),
		})
		return
	}

	// 反转数组，让最新的在前面（用于列表显示）
	// GetLatestRecords返回的是从旧到新（用于图表），这里需要从新到旧
	for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
		records[i], records[j] = records[j], records[i]
	}

	c.JSON(http.StatusOK, records)
}

// handleStatistics 统计信息
func (s *Server) handleStatistics(c *gin.Context) {
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

	stats, err := trader.GetDecisionLogger().GetStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取统计信息失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// handleEquityHistory 收益率历史数据
func (s *Server) handleEquityHistory(c *gin.Context) {
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

	// 获取尽可能多的历史数据（几天的数据）
	// 每3分钟一个周期：10000条 = 约20天的数据
	records, err := trader.GetDecisionLogger().GetLatestRecords(10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取历史数据失败: %v", err),
		})
		return
	}

	// 构建收益率历史数据点
	type EquityPoint struct {
		Timestamp        string  `json:"timestamp"`
		TotalEquity      float64 `json:"total_equity"`      // 账户净值（wallet + unrealized）
		AvailableBalance float64 `json:"available_balance"` // 可用余额
		TotalPnL         float64 `json:"total_pnl"`         // 总盈亏（相对初始余额）
		TotalPnLPct      float64 `json:"total_pnl_pct"`     // 总盈亏百分比
		PositionCount    int     `json:"position_count"`    // 持仓数量
		MarginUsedPct    float64 `json:"margin_used_pct"`   // 保证金使用率
		CycleNumber      int     `json:"cycle_number"`
	}

	// 从AutoTrader获取初始余额（用于计算盈亏百分比）
	initialBalance := 0.0
	if status := trader.GetStatus(); status != nil {
		if ib, ok := status["initial_balance"].(float64); ok && ib > 0 {
			initialBalance = ib
		}
	}

	// 如果无法从status获取，且有历史记录，则从第一条记录获取
	if initialBalance == 0 && len(records) > 0 {
		// 第一条记录的equity作为初始余额
		initialBalance = records[0].AccountState.TotalBalance
	}

	// 如果还是无法获取，返回错误
	if initialBalance == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "无法获取初始余额",
		})
		return
	}

	var history []EquityPoint
	for _, record := range records {
		// TotalBalance字段实际存储的是TotalEquity
		totalEquity := record.AccountState.TotalBalance
		// TotalUnrealizedProfit字段实际存储的是TotalPnL（相对初始余额）
		totalPnL := record.AccountState.TotalUnrealizedProfit

		// 计算盈亏百分比
		totalPnLPct := 0.0
		if initialBalance > 0 {
			totalPnLPct = (totalPnL / initialBalance) * 100
		}

		history = append(history, EquityPoint{
			Timestamp:        record.Timestamp.Format("2006-01-02 15:04:05"),
			TotalEquity:      totalEquity,
			AvailableBalance: record.AccountState.AvailableBalance,
			TotalPnL:         totalPnL,
			TotalPnLPct:      totalPnLPct,
			PositionCount:    record.AccountState.PositionCount,
			MarginUsedPct:    record.AccountState.MarginUsedPct,
			CycleNumber:      record.CycleNumber,
		})
	}

	c.JSON(http.StatusOK, history)
}

// handlePerformance AI历史表现分析（用于展示AI学习和反思）
func (s *Server) handlePerformance(c *gin.Context) {
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

	// 分析最近100个周期的交易表现（避免长期持仓的交易记录丢失）
	// 假设每3分钟一个周期，100个周期 = 5小时，足够覆盖大部分交易
	performance, err := trader.GetDecisionLogger().AnalyzePerformance(100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("分析历史表现失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, performance)
}

// handleGetPrompts 获取prompt配置
func (s *Server) handleGetPrompts(c *gin.Context) {
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

	db := trader.GetDecisionLogger().GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库未初始化"})
		return
	}

	configs, err := db.Config().GetAll()
	if err != nil {
		log.Printf("获取prompt配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    configs,
	})
}

// handleUpdatePrompt 更新prompt配置
func (s *Server) handleUpdatePrompt(c *gin.Context) {
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

	var req struct {
		SectionName  string `json:"section_name"`
		Title        string `json:"title"`
		Content      string `json:"content"`
		PromptType   string `json:"prompt_type"`
		Enabled      bool   `json:"enabled"`
		DisplayOrder int    `json:"display_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 设置默认的prompt_type
	if req.PromptType == "" {
		req.PromptType = "system"
	}

	// 验证prompt_type
	if req.PromptType != "system" && req.PromptType != "user" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt_type must be 'system' or 'user'"})
		return
	}

	// 设置默认的prompt_type
	if req.PromptType == "" {
		req.PromptType = "system"
	}

	// 验证prompt_type
	if req.PromptType != "system" && req.PromptType != "user" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt_type must be 'system' or 'user'"})
		return
	}

	if req.SectionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "section_name is required"})
		return
	}

	db := trader.GetDecisionLogger().GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库未初始化"})
		return
	}

	cfg := &models.PromptConfig{
		SectionName:  req.SectionName,
		Title:        req.Title,
		Content:      req.Content,
		PromptType:   req.PromptType,
		Enabled:      req.Enabled,
		DisplayOrder: req.DisplayOrder,
	}

	if err := db.Config().Update(cfg); err != nil {
		log.Printf("更新prompt配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新配置失败"})
		return
	}

	log.Printf("✓ Prompt配置已更新: %s", req.SectionName)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置更新成功",
	})
}

// handleTogglePrompt 切换prompt启用状态
func (s *Server) handleTogglePrompt(c *gin.Context) {
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

	var req struct {
		SectionName string `json:"section_name"`
		Enabled     bool   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	db := trader.GetDecisionLogger().GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库未初始化"})
		return
	}

	configs, err := db.Config().GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取配置失败"})
		return
	}

	var found bool
	for _, cfg := range configs {
		if cfg.SectionName == req.SectionName {
			cfg.Enabled = req.Enabled
			if err := db.Config().Update(cfg); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "更新配置失败"})
				return
			}
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "配置不存在"})
		return
	}

	status := "禁用"
	if req.Enabled {
		status = "启用"
	}
	log.Printf("✓ Prompt配置已%s: %s", status, req.SectionName)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("已%s配置", status),
	})
}

// handlePreviewPrompt 预览完整prompt
func (s *Server) handlePreviewPrompt(c *gin.Context) {
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

	db := trader.GetDecisionLogger().GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库未初始化"})
		return
	}

	// 获取账户信息用于变量替换
	accountInfo, _ := trader.GetAccountInfo()
	accountEquity := 100.0
	if eq, ok := accountInfo["total_equity"].(float64); ok {
		accountEquity = eq
	}

	// 获取杠杆配置（简化处理，使用默认值）
	btcLeverage := 3
	altLeverage := 10
	
	// 计算实际仓位限制（简化版，使用默认风控参数）
	// 实际系统会根据账户状态动态调整
	baseMaxBTC := accountEquity * 30.0
	baseMaxAlt := accountEquity * 20.0
	actualMaxBTC := baseMaxBTC * 0.85 // 默认应用85%信心度调整
	actualMaxAlt := baseMaxAlt * 0.85

	// 预览时默认使用限制模式（false），展示完整规则
	prompt := db.BuildSystemPromptFromDB(accountEquity, btcLeverage, altLeverage, actualMaxBTC, actualMaxAlt, false)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"prompt":         prompt,
			"account_equity": accountEquity,
			"btc_leverage":   btcLeverage,
			"alt_leverage":   altLeverage,
		},
	})
}

// Start 启动服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("🌐 API服务器启动在 http://localhost%s", addr)
	log.Printf("📊 API文档:")
	log.Printf("  • GET  /api/competition      - 竞赛总览（对比所有trader）")
	log.Printf("  • GET  /api/traders          - Trader列表")
	log.Printf("  • GET  /api/status?trader_id=xxx     - 指定trader的系统状态")
	log.Printf("  • GET  /api/account?trader_id=xxx    - 指定trader的账户信息")
	log.Printf("  • GET  /api/positions?trader_id=xxx  - 指定trader的持仓列表")
	log.Printf("  • GET  /api/decisions?trader_id=xxx  - 指定trader的决策日志")
	log.Printf("  • GET  /api/decisions/latest?trader_id=xxx - 指定trader的最新决策")
	log.Printf("  • GET  /api/statistics?trader_id=xxx - 指定trader的统计信息")
	log.Printf("  • GET  /api/equity-history?trader_id=xxx - 指定trader的收益率历史数据")
	log.Printf("  • GET  /api/performance?trader_id=xxx - 指定trader的AI学习表现分析")
	log.Printf("  • GET  /api/prompts?trader_id=xxx    - 获取Prompt配置")
	log.Printf("  • POST /api/prompts/update?trader_id=xxx - 更新Prompt配置")
	log.Printf("  • POST /api/prompts/toggle?trader_id=xxx - 切换Prompt启用状态")
	log.Printf("  • GET  /api/prompts/preview?trader_id=xxx - 预览完整Prompt")
	log.Printf("  • GET  /health               - 健康检查")
	log.Println()

	return s.router.Run(addr)
}

// handleAddPrompt 添加新的prompt配置
func (s *Server) handleAddPrompt(c *gin.Context) {
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

	var req struct {
		SectionName  string `json:"section_name"`
		Title        string `json:"title"`
		Content      string `json:"content"`
		PromptType   string `json:"prompt_type"`
		Enabled      bool   `json:"enabled"`
		DisplayOrder int    `json:"display_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.SectionName == "" || req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "section_name and title are required"})
		return
	}

	db := trader.GetDecisionLogger().GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库未初始化"})
		return
	}

	// 检查是否已存在
	existing, _ := db.Config().GetAll()
	for _, cfg := range existing {
		if cfg.SectionName == req.SectionName {
			c.JSON(http.StatusBadRequest, gin.H{"error": "该section_name已存在"})
			return
		}
	}

	// 如果没有指定display_order，使用最大值+1
	if req.DisplayOrder == 0 {
		maxOrder := 0
		for _, cfg := range existing {
			if cfg.DisplayOrder > maxOrder {
				maxOrder = cfg.DisplayOrder
			}
		}
		req.DisplayOrder = maxOrder + 1
	}

	// 插入新配置
	cfg := &models.PromptConfig{
		SectionName:  req.SectionName,
		Title:        req.Title,
		Content:      req.Content,
		PromptType:   req.PromptType,
		Enabled:      req.Enabled,
		DisplayOrder: req.DisplayOrder,
	}
	if err := db.Config().Insert(cfg); err != nil{
		log.Printf("添加prompt配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加配置失败"})
		return
	}

	log.Printf("✓ 新增Prompt配置: %s - %s", req.SectionName, req.Title)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "添加成功",
	})
}

// handleDeletePrompt 删除prompt配置
func (s *Server) handleDeletePrompt(c *gin.Context) {
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

	sectionName := c.Query("section_name")
	if sectionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "section_name is required"})
		return
	}

	db := trader.GetDecisionLogger().GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库未初始化"})
		return
	}

	// 删除配置
	rows, err := db.Config().Delete(sectionName)
	if err != nil {
		log.Printf("删除prompt配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除配置失败"})
		return
	}

	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "配置不存在"})
		return
	}

	log.Printf("✓ 删除Prompt配置: %s", sectionName)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "删除成功",
	})
}
