package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// SystemConnection 系统数据库连接管理器（存储用户、trader配置等全局数据）
type SystemConnection struct {
	db     *sql.DB
	dbPath string
}

// NewSystemConnection 创建系统数据库连接
func NewSystemConnection() (*SystemConnection, error) {
	dbPath := "data/system.db"
	
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开系统数据库失败: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	conn := &SystemConnection{
		db:     db,
		dbPath: dbPath,
	}

	// 初始化系统表结构
	if err := conn.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("初始化系统表结构失败: %w", err)
	}

	log.Printf("✓ 系统数据库已初始化: %s", dbPath)
	return conn, nil
}

// DB 获取原始的 sql.DB 对象
func (c *SystemConnection) DB() *sql.DB {
	return c.db
}

// Close 关闭数据库连接
func (c *SystemConnection) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// initSchema 初始化系统表结构
func (c *SystemConnection) initSchema() error {
	schema := `
	-- 用户表
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		is_active BOOLEAN DEFAULT 1
	);
	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

	-- 会话表
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL UNIQUE,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);

	-- 系统配置表
	CREATE TABLE IF NOT EXISTS system_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL UNIQUE,
		value TEXT NOT NULL,
		description TEXT,
		config_type TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_system_configs_key ON system_configs(key);
	CREATE INDEX IF NOT EXISTS idx_system_configs_type ON system_configs(config_type);

	-- Trader配置表
	CREATE TABLE IF NOT EXISTS trader_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL DEFAULT 0,
		trader_id TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		enabled BOOLEAN DEFAULT 1,
		ai_model TEXT NOT NULL,
		exchange TEXT NOT NULL,
		-- API配置（建议加密存储）
		binance_api_key TEXT,
		binance_secret_key TEXT,
		hyperliquid_private_key TEXT,
		hyperliquid_wallet_addr TEXT,
		hyperliquid_testnet BOOLEAN DEFAULT 0,
		aster_user TEXT,
		aster_signer TEXT,
		aster_private_key TEXT,
		-- AI配置
		deepseek_key TEXT,
		qwen_key TEXT,
		custom_api_url TEXT,
		custom_api_key TEXT,
		custom_model_name TEXT,
		-- 交易配置
		initial_balance REAL NOT NULL,
		scan_interval_minutes INTEGER NOT NULL DEFAULT 3,
		max_positions INTEGER NOT NULL DEFAULT 3,
		btc_eth_leverage INTEGER NOT NULL DEFAULT 5,
		altcoin_leverage INTEGER NOT NULL DEFAULT 5,
		-- 风控配置
		max_daily_loss REAL DEFAULT 0,
		max_drawdown REAL DEFAULT 0,
		stop_trading_minutes INTEGER DEFAULT 0,
		-- AI学习配置
		enable_ai_learning BOOLEAN DEFAULT 0,
		ai_learn_interval INTEGER DEFAULT 10,
		-- AI自主模式配置
		ai_autonomy_mode BOOLEAN DEFAULT 0,
		-- 数据优化配置
		compact_mode BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_trader_configs_trader_id ON trader_configs(trader_id);
	CREATE INDEX IF NOT EXISTS idx_trader_configs_user_id ON trader_configs(user_id);
	CREATE INDEX IF NOT EXISTS idx_trader_configs_enabled ON trader_configs(enabled);
	`

	_, err := c.db.Exec(schema)
	if err != nil {
		return err
	}

	// 初始化默认系统配置
	return c.initDefaultConfigs()
}

// initDefaultConfigs 初始化默认系统配置
func (c *SystemConnection) initDefaultConfigs() error {
	// 检查是否已初始化
	var count int
	err := c.db.QueryRow("SELECT COUNT(*) FROM system_configs").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // 已初始化
	}

	log.Println("🔧 初始化默认系统配置...")

	defaults := []struct {
		Key         string
		Value       string
		Description string
		ConfigType  string
	}{
		// API配置
		{"api_server_port", "8080", "API服务器端口", "api"},
		
		// 市场数据配置
		{"coin_pool_api_url", "", "币种池API地址", "market"},
		{"oi_top_api_url", "", "持仓量TopAPI地址", "market"},
		{"use_default_coins", "true", "是否使用默认币种列表", "market"},
		{"default_coins", `["BTCUSDT","ETHUSDT","SOLUSDT","BNBUSDT","XRPUSDT","DOGEUSDT","ADAUSDT","HYPEUSDT"]`, "默认币种列表", "market"},
		{"kline_settings", `[{"interval":"3m","limit":20,"show_table":true},{"interval":"4h","limit":60,"show_table":false}]`, "K线配置", "market"},
		
		// 查询限制配置
		{"query_limit_default", "100", "默认记录查询数量", "database"},
		{"query_limit_performance", "100", "性能分析记录数量", "database"},
		{"query_limit_monitoring", "50", "监控记录数量", "database"},
		{"query_limit_recent", "20", "近期表现记录数量", "database"},
		{"query_limit_trades", "50", "交易结果查询数量", "database"},
		
		// 风险阈值配置
		{"risk_margin_high_threshold", "50.0", "保证金使用率高风险阈值(%)", "risk"},
		{"risk_margin_medium_threshold", "20.0", "保证金使用率中风险阈值(%)", "risk"},
		{"risk_drawdown_critical_threshold", "30.0", "回撤危险阈值(%)", "risk"},
		{"risk_drawdown_high_threshold", "20.0", "回撤高风险阈值(%)", "risk"},
		{"risk_drawdown_medium_threshold", "10.0", "回撤中风险阈值(%)", "risk"},
		{"risk_sharpe_low_threshold", "-0.5", "夏普比率低阈值", "risk"},
		{"risk_sharpe_poor_threshold", "0.0", "夏普比率差阈值", "risk"},
		{"risk_winrate_low_threshold", "30.0", "胜率低阈值(%)", "risk"},
		{"risk_error_rate_high_threshold", "10.0", "错误率高阈值(%)", "risk"},
		{"risk_min_trades_for_stats", "10", "统计分析最小交易数", "risk"},
		
		// 风险评分权重配置
		{"risk_score_margin_high", "20", "保证金高使用率评分", "risk"},
		{"risk_score_margin_medium", "10", "保证金中使用率评分", "risk"},
		{"risk_score_drawdown_critical", "30", "危险回撤评分", "risk"},
		{"risk_score_drawdown_high", "20", "高回撤评分", "risk"},
		{"risk_score_drawdown_medium", "10", "中回撤评分", "risk"},
		{"risk_score_sharpe_low", "20", "低夏普比率评分", "risk"},
		{"risk_score_sharpe_poor", "10", "差夏普比率评分", "risk"},
		
		// 技术指标参数配置
		{"indicator_bollinger_period", "20", "布林带周期", "indicator"},
		{"indicator_bollinger_stddev", "2.0", "布林带标准差倍数", "indicator"},
		{"indicator_stochastic_k", "14", "随机指标K值周期", "indicator"},
		{"indicator_stochastic_d", "3", "随机指标D值周期", "indicator"},
		{"indicator_cci_period", "20", "CCI周期", "indicator"},
		{"indicator_rsi_period", "14", "RSI周期", "indicator"},
		{"indicator_vwma_period", "20", "VWMA周期", "indicator"},
		{"indicator_macd_fast", "12", "MACD快线周期", "indicator"},
		{"indicator_macd_slow", "26", "MACD慢线周期", "indicator"},
		{"indicator_macd_signal", "9", "MACD信号线周期", "indicator"},
		
		// 币种池配置
		{"pool_max_retries", "3", "API请求最大重试次数", "pool"},
		{"pool_retry_delay_ms", "100", "重试延迟(毫秒)", "pool"},
		{"pool_timeout_seconds", "10", "请求超时时间(秒)", "pool"},
		{"pool_cache_ttl_minutes", "5", "缓存有效期(分钟)", "pool"},
		
		// 交易配置
		{"trading_max_positions", "3", "最大持仓数", "trading"},
		{"trading_scan_interval_minutes", "3", "扫描间隔(分钟)", "trading"},
		
		// 备份配置
		{"backup_retention_count", "5", "保留备份数量", "backup"},
	}

	for _, cfg := range defaults {
		_, err := c.db.Exec(`
			INSERT INTO system_configs (key, value, description, config_type)
			VALUES (?, ?, ?, ?)
		`, cfg.Key, cfg.Value, cfg.Description, cfg.ConfigType)

		if err != nil {
			return fmt.Errorf("插入默认配置失败 [%s]: %w", cfg.Key, err)
		}
	}

	log.Println("✓ 默认系统配置初始化完成")
	return nil
}
