package monitoring

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"nofx/database"
	"nofx/database/models"
	"nofx/logger"
)

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	mu                sync.RWMutex
	traderID          string
	db                *database.DB
	logger            *logger.DecisionLogger
	metrics           *PerformanceMetrics
	runtimeConfig     *database.RuntimeConfig  // 运行时配置
	alerts            []Alert
	alertHandlers     []AlertHandler
	monitoringEnabled bool
	stopChan          chan struct{}
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	// 基础指标
	TotalTrades       int     `json:"total_trades"`
	WinRate           float64 `json:"win_rate"`
	ProfitFactor      float64 `json:"profit_factor"`
	SharpeRatio       float64 `json:"sharpe_ratio"`
	MaxDrawdown       float64 `json:"max_drawdown"`
	CurrentDrawdown   float64 `json:"current_drawdown"`
	
	// 风险指标
	VaR95             float64 `json:"var_95"`
	VaR99             float64 `json:"var_99"`
	RiskScore         int     `json:"risk_score"`         // 0-100
	MarginUsageRate   float64 `json:"margin_usage_rate"`
	LiquidationRisk   float64 `json:"liquidation_risk"`   // 距离强平的百分比
	
	// 实时指标
	CurrentBalance    float64 `json:"current_balance"`
	AvailableBalance  float64 `json:"available_balance"`
	UnrealizedPnL     float64 `json:"unrealized_pnl"`
	TotalPnL          float64 `json:"total_pnl"`
	
	// 交易频率指标
	TradesPerHour     float64 `json:"trades_per_hour"`
	AvgHoldingTime    float64 `json:"avg_holding_time"`   // 分钟
	OverTradingScore  int     `json:"overtrading_score"`  // 0-100，越高越过度
	
	// 系统性能指标
	APILatency        float64 `json:"api_latency"`        // 毫秒
	DecisionLatency   float64 `json:"decision_latency"`   // 毫秒
	ErrorRate         float64 `json:"error_rate"`         // 百分比
	SystemUptime      float64 `json:"system_uptime"`      // 小时
	
	// 时间戳
	LastUpdated       time.Time `json:"last_updated"`
}

// Alert 预警信息
type Alert struct {
	ID          string    `json:"id"`
	Type        AlertType `json:"type"`
	Level       AlertLevel `json:"level"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	Resolved    bool      `json:"resolved"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// AlertType 预警类型
type AlertType string

const (
	AlertTypeRisk        AlertType = "risk"
	AlertTypePerformance AlertType = "performance"
	AlertTypeSystem      AlertType = "system"
	AlertTypeTrade       AlertType = "trade"
)

// AlertLevel 预警级别
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
)

// AlertHandler 预警处理器
type AlertHandler interface {
	HandleAlert(alert Alert) error
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor(traderID string, db *database.DB, logger *logger.DecisionLogger) *PerformanceMonitor {
	return &PerformanceMonitor{
		traderID:          traderID,
		db:                db,
		logger:            logger,
		metrics:           &PerformanceMetrics{},
		alerts:            make([]Alert, 0),
		alertHandlers:     make([]AlertHandler, 0),
		monitoringEnabled: false,
		stopChan:          make(chan struct{}),
	}
}

// Start 启动性能监控
func (pm *PerformanceMonitor) Start() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if pm.monitoringEnabled {
		return
	}
	
	pm.monitoringEnabled = true
	log.Printf("🔍 [%s] 性能监控器启动", pm.traderID)
	
	// 启动监控协程
	go pm.monitoringLoop()
}

// Stop 停止性能监控
func (pm *PerformanceMonitor) Stop() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if !pm.monitoringEnabled {
		return
	}
	
	pm.monitoringEnabled = false
	close(pm.stopChan)
	log.Printf("🔍 [%s] 性能监控器停止", pm.traderID)
}

// monitoringLoop 监控循环
func (pm *PerformanceMonitor) monitoringLoop() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒更新一次
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			pm.updateMetrics()
			pm.checkAlerts()
		case <-pm.stopChan:
			return
		}
	}
}

// updateMetrics 更新性能指标
func (pm *PerformanceMonitor) updateMetrics() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	// 从配置获取查询限制
	queryLimits := pm.runtimeConfig.GetQueryLimits()
	
	// 获取交易表现分析
	performance, err := pm.logger.AnalyzePerformance(queryLimits.PerformanceLimit)
	if err != nil {
		log.Printf("⚠️ [%s] 获取交易表现失败: %v", pm.traderID, err)
		return
	}
	
	// 获取最新决策记录
	records, err := pm.db.Decision().GetLatest(queryLimits.MonitoringLimit)
	if err != nil {
		log.Printf("⚠️ [%s] 获取决策记录失败: %v", pm.traderID, err)
		return
	}
	
	// 更新基础指标
	pm.metrics.TotalTrades = performance.TotalTrades
	pm.metrics.WinRate = performance.WinRate
	pm.metrics.ProfitFactor = performance.ProfitFactor
	pm.metrics.SharpeRatio = performance.SharpeRatio
	
	// 计算风险指标
	pm.calculateRiskMetrics(records)
	
	// 计算交易频率指标
	pm.calculateTradingFrequencyMetrics(records)
	
	// 更新时间戳
	pm.metrics.LastUpdated = time.Now()
	
	log.Printf("📊 [%s] 性能指标已更新 - 胜率: %.1f%%, 夏普: %.2f, 风险评分: %d", 
		pm.traderID, pm.metrics.WinRate, pm.metrics.SharpeRatio, pm.metrics.RiskScore)
}

// calculateRiskMetrics 计算风险指标
func (pm *PerformanceMonitor) calculateRiskMetrics(records []*models.DecisionRecord) {
	if len(records) == 0 {
		return
	}
	
	// 计算最大回撤
	var maxBalance, minBalance float64
	var balances []float64
	
	for i, record := range records {
		balance := record.TotalBalance
		balances = append(balances, balance)
		
		if i == 0 {
			maxBalance = balance
			minBalance = balance
		} else {
			if balance > maxBalance {
				maxBalance = balance
			}
			if balance < minBalance {
				minBalance = balance
			}
		}
	}
	
	// 计算最大回撤
	pm.metrics.MaxDrawdown = pm.calculateMaxDrawdown(balances)
	
	// 计算当前回撤
	if len(balances) > 0 {
		currentBalance := balances[len(balances)-1]
		pm.metrics.CurrentBalance = currentBalance
		pm.metrics.CurrentDrawdown = (maxBalance - currentBalance) / maxBalance * 100
	}
	
	// 计算VaR
	pm.calculateVaR(balances)
	
	// 计算风险评分
	pm.calculateRiskScore(records)
}

// calculateMaxDrawdown 计算最大回撤
func (pm *PerformanceMonitor) calculateMaxDrawdown(balances []float64) float64 {
	if len(balances) < 2 {
		return 0
	}
	
	var maxDrawdown float64
	peak := balances[0]
	
	for _, balance := range balances {
		if balance > peak {
			peak = balance
		}
		
		drawdown := (peak - balance) / peak * 100
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}
	
	return maxDrawdown
}

// calculateVaR 计算风险价值
func (pm *PerformanceMonitor) calculateVaR(balances []float64) {
	if len(balances) < 10 {
		return
	}
	
	// 计算收益率序列
	returns := make([]float64, len(balances)-1)
	for i := 1; i < len(balances); i++ {
		returns[i-1] = (balances[i] - balances[i-1]) / balances[i-1]
	}
	
	// 计算VaR (简化版本，使用正态分布假设)
	mean := pm.calculateMean(returns)
	std := pm.calculateStd(returns, mean)
	
	// VaR95 = mean - 1.645 * std
	// VaR99 = mean - 2.326 * std
	pm.metrics.VaR95 = math.Abs(mean - 1.645*std) * pm.metrics.CurrentBalance
	pm.metrics.VaR99 = math.Abs(mean - 2.326*std) * pm.metrics.CurrentBalance
}

// calculateMean 计算均值
func (pm *PerformanceMonitor) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateStd 计算标准差
func (pm *PerformanceMonitor) calculateStd(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sumSquares := 0.0
	for _, v := range values {
		sumSquares += (v - mean) * (v - mean)
	}
	return math.Sqrt(sumSquares / float64(len(values)))
}

// calculateRiskScore 计算风险评分 (0-100)
func (pm *PerformanceMonitor) calculateRiskScore(records []*models.DecisionRecord) {
	// 获取风险阈值和评分配置
	thresholds := pm.runtimeConfig.GetRiskThresholds()
	scores := pm.runtimeConfig.GetRiskScores()
	
	if len(records) == 0 {
		pm.metrics.RiskScore = 50
		return
	}
	
	score := 0
	
	// 最新记录的保证金使用率
	latestRecord := records[len(records)-1]
	marginUsage := latestRecord.MarginUsedPct
	pm.metrics.MarginUsageRate = marginUsage
	
	if marginUsage > thresholds.MarginHighThreshold {
		score += scores.MarginHighScore
	} else if marginUsage > thresholds.MarginMediumThreshold {
		score += scores.MarginMediumScore
	}
	
	// 最大回撤
	if pm.metrics.MaxDrawdown > thresholds.DrawdownCriticalThreshold {
		score += scores.DrawdownCriticalScore
	} else if pm.metrics.MaxDrawdown > thresholds.DrawdownHighThreshold {
		score += scores.DrawdownHighScore
	} else if pm.metrics.MaxDrawdown > thresholds.DrawdownMediumThreshold {
		score += scores.DrawdownMediumScore
	}
	
	// 夏普比率
	if pm.metrics.SharpeRatio < thresholds.SharpeRatioLowThreshold {
		score += scores.SharpeRatioLowScore
	} else if pm.metrics.SharpeRatio < thresholds.SharpeRatioPoorThreshold {
		score += scores.SharpeRatioPoorScore
	}
	
	// 胜率
	if pm.metrics.WinRate < thresholds.WinRateLowThreshold {
		score += 10
	}
	
	pm.metrics.RiskScore = score
}

// calculateTradingFrequencyMetrics 计算交易频率指标
func (pm *PerformanceMonitor) calculateTradingFrequencyMetrics(records []*models.DecisionRecord) {
	if len(records) < 2 {
		return
	}
	
	// 计算每小时交易次数
	timeSpan := records[len(records)-1].Timestamp.Sub(records[0].Timestamp).Hours()
	if timeSpan > 0 {
		pm.metrics.TradesPerHour = float64(pm.metrics.TotalTrades) / timeSpan
	}
	
	// 计算过度交易评分
	if pm.metrics.TradesPerHour > 2 {
		pm.metrics.OverTradingScore = 100
	} else if pm.metrics.TradesPerHour > 1 {
		pm.metrics.OverTradingScore = 70
	} else if pm.metrics.TradesPerHour > 0.5 {
		pm.metrics.OverTradingScore = 40
	} else {
		pm.metrics.OverTradingScore = 10
	}
}

// checkAlerts 检查预警条件
func (pm *PerformanceMonitor) checkAlerts() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	// 获取风险阈值配置
	thresholds := pm.runtimeConfig.GetRiskThresholds()
	
	// 检查风险预警
	pm.checkRiskAlerts(thresholds)
	
	// 检查性能预警
	pm.checkPerformanceAlerts(thresholds)
	
	// 检查系统预警
	pm.checkSystemAlerts(thresholds)
}

// checkRiskAlerts 检查风险预警
func (pm *PerformanceMonitor) checkRiskAlerts(thresholds database.RiskThresholds) {
	// 高风险评分预警
	if pm.metrics.RiskScore >= 80 {
		pm.addAlert(Alert{
			ID:        fmt.Sprintf("risk_score_%d", time.Now().Unix()),
			Type:      AlertTypeRisk,
			Level:     AlertLevelCritical,
			Title:     "极高风险警告",
			Message:   fmt.Sprintf("风险评分达到 %d/100，建议立即减仓或停止交易", pm.metrics.RiskScore),
			Timestamp: time.Now(),
		})
	} else if pm.metrics.RiskScore >= 60 {
		pm.addAlert(Alert{
			ID:        fmt.Sprintf("risk_score_%d", time.Now().Unix()),
			Type:      AlertTypeRisk,
			Level:     AlertLevelWarning,
			Title:     "高风险警告",
			Message:   fmt.Sprintf("风险评分达到 %d/100，建议谨慎交易", pm.metrics.RiskScore),
			Timestamp: time.Now(),
		})
	}
	
	// 保证金使用率预警
	if pm.metrics.MarginUsageRate >= 80 {
		pm.addAlert(Alert{
			ID:        fmt.Sprintf("margin_usage_%d", time.Now().Unix()),
			Type:      AlertTypeRisk,
			Level:     AlertLevelCritical,
			Title:     "保证金使用率过高",
			Message:   fmt.Sprintf("保证金使用率 %.1f%%，接近强平风险", pm.metrics.MarginUsageRate),
			Timestamp: time.Now(),
		})
	}
	
	// 最大回撤预警
	if pm.metrics.MaxDrawdown >= 30 {
		pm.addAlert(Alert{
			ID:        fmt.Sprintf("max_drawdown_%d", time.Now().Unix()),
			Type:      AlertTypeRisk,
			Level:     AlertLevelCritical,
			Title:     "最大回撤过大",
			Message:   fmt.Sprintf("最大回撤达到 %.1f%%，建议暂停交易", pm.metrics.MaxDrawdown),
			Timestamp: time.Now(),
		})
	}
}

// checkPerformanceAlerts 检查性能预警
func (pm *PerformanceMonitor) checkPerformanceAlerts(thresholds database.RiskThresholds) {
	// 夏普比率预警
	if pm.metrics.SharpeRatio < thresholds.SharpeRatioLowThreshold {
		pm.addAlert(Alert{
			ID:        fmt.Sprintf("sharpe_ratio_%d", time.Now().Unix()),
			Type:      AlertTypePerformance,
			Level:     AlertLevelWarning,
			Title:     "夏普比率过低",
			Message:   fmt.Sprintf("夏普比率 %.2f，策略表现不佳", pm.metrics.SharpeRatio),
			Timestamp: time.Now(),
		})
	}
	
	// 胜率预警
	if pm.metrics.WinRate < thresholds.WinRateLowThreshold && pm.metrics.TotalTrades >= thresholds.MinTradesForStats {
		pm.addAlert(Alert{
			ID:        fmt.Sprintf("win_rate_%d", time.Now().Unix()),
			Type:      AlertTypePerformance,
			Level:     AlertLevelWarning,
			Title:     "胜率过低",
			Message:   fmt.Sprintf("胜率仅 %.1f%%，需要优化策略", pm.metrics.WinRate),
			Timestamp: time.Now(),
		})
	}
	
	// 过度交易预警
	if pm.metrics.OverTradingScore >= 70 {
		pm.addAlert(Alert{
			ID:        fmt.Sprintf("overtrading_%d", time.Now().Unix()),
			Type:      AlertTypeTrade,
			Level:     AlertLevelWarning,
			Title:     "过度交易警告",
			Message:   fmt.Sprintf("每小时交易 %.1f 次，可能存在过度交易", pm.metrics.TradesPerHour),
			Timestamp: time.Now(),
		})
	}
}

// checkSystemAlerts 检查系统预警
func (pm *PerformanceMonitor) checkSystemAlerts(thresholds database.RiskThresholds) {
	// API延迟预警
	if pm.metrics.APILatency > 5000 { // 5秒
		pm.addAlert(Alert{
			ID:        fmt.Sprintf("api_latency_%d", time.Now().Unix()),
			Type:      AlertTypeSystem,
			Level:     AlertLevelWarning,
			Title:     "API延迟过高",
			Message:   fmt.Sprintf("API延迟 %.0f ms，可能影响交易执行", pm.metrics.APILatency),
			Timestamp: time.Now(),
		})
	}
	
	// 错误率预警
	if pm.metrics.ErrorRate > thresholds.ErrorRateHighThreshold {
		pm.addAlert(Alert{
			ID:        fmt.Sprintf("error_rate_%d", time.Now().Unix()),
			Type:      AlertTypeSystem,
			Level:     AlertLevelWarning,
			Title:     "系统错误率过高",
			Message:   fmt.Sprintf("错误率 %.1f%%，系统可能存在问题", pm.metrics.ErrorRate),
			Timestamp: time.Now(),
		})
	}
}

// addAlert 添加预警
func (pm *PerformanceMonitor) addAlert(alert Alert) {
	// 检查是否已存在相同类型的未解决预警
	for _, existingAlert := range pm.alerts {
		if existingAlert.Type == alert.Type && existingAlert.Level == alert.Level && !existingAlert.Resolved {
			return // 避免重复预警
		}
	}
	
	pm.alerts = append(pm.alerts, alert)
	
	// 触发预警处理器
	for _, handler := range pm.alertHandlers {
		go func(h AlertHandler, a Alert) {
			if err := h.HandleAlert(a); err != nil {
				log.Printf("⚠️ [%s] 预警处理失败: %v", pm.traderID, err)
			}
		}(handler, alert)
	}
	
	log.Printf("🚨 [%s] %s: %s - %s", pm.traderID, alert.Level, alert.Title, alert.Message)
}

// GetMetrics 获取性能指标
func (pm *PerformanceMonitor) GetMetrics() *PerformanceMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	// 返回副本
	metrics := *pm.metrics
	return &metrics
}

// GetAlerts 获取预警列表
func (pm *PerformanceMonitor) GetAlerts(limit int) []Alert {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	alerts := make([]Alert, len(pm.alerts))
	copy(alerts, pm.alerts)
	
	// 按时间倒序排序
	for i := 0; i < len(alerts)-1; i++ {
		for j := i + 1; j < len(alerts); j++ {
			if alerts[i].Timestamp.Before(alerts[j].Timestamp) {
				alerts[i], alerts[j] = alerts[j], alerts[i]
			}
		}
	}
	
	if limit > 0 && len(alerts) > limit {
		alerts = alerts[:limit]
	}
	
	return alerts
}

// ResolveAlert 解决预警
func (pm *PerformanceMonitor) ResolveAlert(alertID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	for i := range pm.alerts {
		if pm.alerts[i].ID == alertID {
			pm.alerts[i].Resolved = true
			now := time.Now()
			pm.alerts[i].ResolvedAt = &now
			return nil
		}
	}
	
	return fmt.Errorf("预警 %s 不存在", alertID)
}

// AddAlertHandler 添加预警处理器
func (pm *PerformanceMonitor) AddAlertHandler(handler AlertHandler) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.alertHandlers = append(pm.alertHandlers, handler)
}

// GetStatus 获取监控状态
func (pm *PerformanceMonitor) GetStatus() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	return map[string]interface{}{
		"enabled":      pm.monitoringEnabled,
		"trader_id":    pm.traderID,
		"last_updated": pm.metrics.LastUpdated,
		"alerts_count": len(pm.alerts),
		"risk_score":   pm.metrics.RiskScore,
	}
}