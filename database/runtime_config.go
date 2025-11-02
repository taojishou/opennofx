package database

import (
	"database/sql"
	"sync"
)

// RuntimeConfig 运行时配置管理器（支持热重载）
type RuntimeConfig struct {
	helper *ConfigHelper
	mu     sync.RWMutex
	
	// 缓存的配置（避免频繁查询数据库）
	cache map[string]interface{}
}

// NewRuntimeConfig 创建运行时配置管理器
func NewRuntimeConfig(db *sql.DB) *RuntimeConfig {
	return &RuntimeConfig{
		helper: NewConfigHelper(db),
		cache:  make(map[string]interface{}),
	}
}

// QueryLimits 查询限制配置
type QueryLimits struct {
	DefaultLimit       int
	PerformanceLimit   int
	MonitoringLimit    int
	RecentLimit        int
	TradesLimit        int
}

// GetQueryLimits 获取查询限制配置
func (rc *RuntimeConfig) GetQueryLimits() QueryLimits {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	
	return QueryLimits{
		DefaultLimit:     rc.helper.GetInt("query_limit_default", 100),
		PerformanceLimit: rc.helper.GetInt("query_limit_performance", 100),
		MonitoringLimit:  rc.helper.GetInt("query_limit_monitoring", 50),
		RecentLimit:      rc.helper.GetInt("query_limit_recent", 20),
		TradesLimit:      rc.helper.GetInt("query_limit_trades", 50),
	}
}

// RiskThresholds 风险阈值配置
type RiskThresholds struct {
	MarginHighThreshold       float64
	MarginMediumThreshold     float64
	DrawdownCriticalThreshold float64
	DrawdownHighThreshold     float64
	DrawdownMediumThreshold   float64
	SharpeRatioLowThreshold   float64
	SharpeRatioPoorThreshold  float64
	WinRateLowThreshold       float64
	ErrorRateHighThreshold    float64
	MinTradesForStats         int
}

// GetRiskThresholds 获取风险阈值配置
func (rc *RuntimeConfig) GetRiskThresholds() RiskThresholds {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	
	return RiskThresholds{
		MarginHighThreshold:       rc.helper.GetFloat("risk_margin_high_threshold", 50.0),
		MarginMediumThreshold:     rc.helper.GetFloat("risk_margin_medium_threshold", 20.0),
		DrawdownCriticalThreshold: rc.helper.GetFloat("risk_drawdown_critical_threshold", 30.0),
		DrawdownHighThreshold:     rc.helper.GetFloat("risk_drawdown_high_threshold", 20.0),
		DrawdownMediumThreshold:   rc.helper.GetFloat("risk_drawdown_medium_threshold", 10.0),
		SharpeRatioLowThreshold:   rc.helper.GetFloat("risk_sharpe_low_threshold", -0.5),
		SharpeRatioPoorThreshold:  rc.helper.GetFloat("risk_sharpe_poor_threshold", 0.0),
		WinRateLowThreshold:       rc.helper.GetFloat("risk_winrate_low_threshold", 30.0),
		ErrorRateHighThreshold:    rc.helper.GetFloat("risk_error_rate_high_threshold", 10.0),
		MinTradesForStats:         rc.helper.GetInt("risk_min_trades_for_stats", 10),
	}
}

// RiskScores 风险评分权重配置
type RiskScores struct {
	MarginHighScore       int
	MarginMediumScore     int
	DrawdownCriticalScore int
	DrawdownHighScore     int
	DrawdownMediumScore   int
	SharpeRatioLowScore   int
	SharpeRatioPoorScore  int
}

// GetRiskScores 获取风险评分配置
func (rc *RuntimeConfig) GetRiskScores() RiskScores {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	
	return RiskScores{
		MarginHighScore:       rc.helper.GetInt("risk_score_margin_high", 20),
		MarginMediumScore:     rc.helper.GetInt("risk_score_margin_medium", 10),
		DrawdownCriticalScore: rc.helper.GetInt("risk_score_drawdown_critical", 30),
		DrawdownHighScore:     rc.helper.GetInt("risk_score_drawdown_high", 20),
		DrawdownMediumScore:   rc.helper.GetInt("risk_score_drawdown_medium", 10),
		SharpeRatioLowScore:   rc.helper.GetInt("risk_score_sharpe_low", 20),
		SharpeRatioPoorScore:  rc.helper.GetInt("risk_score_sharpe_poor", 10),
	}
}

// IndicatorParams 技术指标参数配置
type IndicatorParams struct {
	BollingerPeriod  int
	BollingerStdDev  float64
	StochasticK      int
	StochasticD      int
	CCIPeriod        int
	RSIPeriod        int
	VWMAPeriod       int
	MACDFastPeriod   int
	MACDSlowPeriod   int
	MACDSignalPeriod int
}

// GetIndicatorParams 获取技术指标参数配置
func (rc *RuntimeConfig) GetIndicatorParams() IndicatorParams {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	
	return IndicatorParams{
		BollingerPeriod:  rc.helper.GetInt("indicator_bollinger_period", 20),
		BollingerStdDev:  rc.helper.GetFloat("indicator_bollinger_stddev", 2.0),
		StochasticK:      rc.helper.GetInt("indicator_stochastic_k", 14),
		StochasticD:      rc.helper.GetInt("indicator_stochastic_d", 3),
		CCIPeriod:        rc.helper.GetInt("indicator_cci_period", 20),
		RSIPeriod:        rc.helper.GetInt("indicator_rsi_period", 14),
		VWMAPeriod:       rc.helper.GetInt("indicator_vwma_period", 20),
		MACDFastPeriod:   rc.helper.GetInt("indicator_macd_fast", 12),
		MACDSlowPeriod:   rc.helper.GetInt("indicator_macd_slow", 26),
		MACDSignalPeriod: rc.helper.GetInt("indicator_macd_signal", 9),
	}
}

// PoolConfig 币种池配置
type PoolConfig struct {
	MaxRetries     int
	RetryDelayMS   int
	TimeoutSeconds int
	CacheTTLMin    int
}

// GetPoolConfig 获取币种池配置
func (rc *RuntimeConfig) GetPoolConfig() PoolConfig {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	
	return PoolConfig{
		MaxRetries:     rc.helper.GetInt("pool_max_retries", 3),
		RetryDelayMS:   rc.helper.GetInt("pool_retry_delay_ms", 100),
		TimeoutSeconds: rc.helper.GetInt("pool_timeout_seconds", 10),
		CacheTTLMin:    rc.helper.GetInt("pool_cache_ttl_minutes", 5),
	}
}

// ClearCache 清除配置缓存（用于热重载）
func (rc *RuntimeConfig) ClearCache() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.cache = make(map[string]interface{})
}

// 全局运行时配置实例（将在系统启动时初始化）
var globalRuntimeConfig *RuntimeConfig
var globalConfigMu sync.RWMutex

// InitGlobalConfig 初始化全局配置
func InitGlobalConfig(db *sql.DB) {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()
	globalRuntimeConfig = NewRuntimeConfig(db)
}

// GetGlobalConfig 获取全局配置
func GetGlobalConfig() *RuntimeConfig {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalRuntimeConfig
}

// ReloadGlobalConfig 重新加载全局配置
func ReloadGlobalConfig() {
	if globalRuntimeConfig != nil {
		globalRuntimeConfig.ClearCache()
	}
}
