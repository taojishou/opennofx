package database

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Connection 数据库连接管理器
type Connection struct {
	db       *sql.DB
	dbPath   string
	traderID string
}

// NewConnection 创建新的数据库连接
func NewConnection(traderID string) (*Connection, error) {
	config := DefaultConfig()
	
	// 确保所有必要的目录存在
	if err := config.EnsureDirectories(traderID); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %w", err)
	}

	dbPath := config.GetTraderDBPath(traderID)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(1) // SQLite 推荐单连接
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	conn := &Connection{
		db:       db,
		dbPath:   dbPath,
		traderID: traderID,
	}

	// 初始化表结构
	if err := conn.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("初始化表结构失败: %w", err)
	}

	log.Printf("✓ SQLite数据库已初始化: %s", dbPath)
	return conn, nil
}

// DB 获取原始的 sql.DB 对象（用于执行查询）
func (c *Connection) DB() *sql.DB {
	return c.db
}

// TraderID 获取 Trader ID
func (c *Connection) TraderID() string {
	return c.traderID
}

// Close 关闭数据库连接
func (c *Connection) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Backup 创建数据库备份
func (c *Connection) Backup(timestamp string) error {
	config := DefaultConfig()
	backupPath := config.GetBackupPath(c.traderID, timestamp)
	
	// 确保备份目录存在
	if err := config.EnsureDirectories(c.traderID); err != nil {
		return fmt.Errorf("创建备份目录失败: %w", err)
	}
	
	// 执行备份
	backupQuery := fmt.Sprintf("VACUUM INTO '%s'", backupPath)
	_, err := c.db.Exec(backupQuery)
	if err != nil {
		return fmt.Errorf("备份数据库失败: %w", err)
	}
	
	log.Printf("✓ 数据库备份完成: %s", backupPath)
	
	// 清理旧备份（保留最近5个）
	if err := config.CleanupOldBackups(c.traderID, 5); err != nil {
		log.Printf("⚠️ 清理旧备份失败: %v", err)
	}
	
	return nil
}

// BeginTx 开始事务
func (c *Connection) BeginTx() (*sql.Tx, error) {
	return c.db.Begin()
}

// initSchema 初始化数据库表结构
func (c *Connection) initSchema() error {
	schema := `
	-- AI学习总结表
	CREATE TABLE IF NOT EXISTS ai_learning_summaries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		trader_id TEXT NOT NULL,
		summary_content TEXT NOT NULL,
		trades_count INTEGER NOT NULL,
		date_range_start TEXT,
		date_range_end TEXT,
		win_rate REAL,
		avg_pnl REAL,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP,
		is_active BOOLEAN DEFAULT 1
	);
	CREATE INDEX IF NOT EXISTS idx_ai_learning_trader ON ai_learning_summaries(trader_id);
	CREATE INDEX IF NOT EXISTS idx_ai_learning_active ON ai_learning_summaries(trader_id, is_active);

	-- 决策记录主表
	CREATE TABLE IF NOT EXISTS decision_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		trader_id TEXT NOT NULL,
		cycle_number INTEGER NOT NULL,
		timestamp DATETIME NOT NULL,
		system_prompt TEXT,
		input_prompt TEXT,
		cot_trace TEXT,
		decision_json TEXT,
		success BOOLEAN NOT NULL,
		error_message TEXT,
		-- 账户状态快照
		total_balance REAL NOT NULL,
		available_balance REAL NOT NULL,
		total_unrealized_profit REAL NOT NULL,
		position_count INTEGER NOT NULL,
		margin_used_pct REAL NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- 决策动作表
	CREATE TABLE IF NOT EXISTS decision_actions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		record_id INTEGER NOT NULL,
		action TEXT NOT NULL,
		symbol TEXT NOT NULL,
		quantity REAL NOT NULL,
		leverage INTEGER,
		price REAL NOT NULL,
		order_id INTEGER,
		timestamp DATETIME NOT NULL,
		success BOOLEAN NOT NULL,
		error TEXT,
		was_stop_loss BOOLEAN DEFAULT 0,
		FOREIGN KEY (record_id) REFERENCES decision_records(id) ON DELETE CASCADE
	);

	-- 持仓快照表
	CREATE TABLE IF NOT EXISTS position_snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		record_id INTEGER NOT NULL,
		symbol TEXT NOT NULL,
		side TEXT NOT NULL,
		position_amt REAL NOT NULL,
		entry_price REAL NOT NULL,
		mark_price REAL NOT NULL,
		unrealized_profit REAL NOT NULL,
		leverage REAL NOT NULL,
		liquidation_price REAL NOT NULL,
		FOREIGN KEY (record_id) REFERENCES decision_records(id) ON DELETE CASCADE
	);

	-- 候选币种表
	CREATE TABLE IF NOT EXISTS candidate_coins (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		record_id INTEGER NOT NULL,
		symbol TEXT NOT NULL,
		FOREIGN KEY (record_id) REFERENCES decision_records(id) ON DELETE CASCADE
	);

	-- 交易结果表（用于统计分析）
	CREATE TABLE IF NOT EXISTS trade_outcomes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		trader_id TEXT NOT NULL,
		symbol TEXT NOT NULL,
		side TEXT NOT NULL,
		quantity REAL NOT NULL,
		leverage INTEGER NOT NULL,
		open_price REAL NOT NULL,
		close_price REAL NOT NULL,
		position_value REAL NOT NULL,
		margin_used REAL NOT NULL,
		pnl REAL NOT NULL,
		pnl_pct REAL NOT NULL,
		duration_minutes INTEGER NOT NULL,
		open_time DATETIME NOT NULL,
		close_time DATETIME NOT NULL,
		was_stop_loss BOOLEAN DEFAULT 0,
		entry_reason TEXT,
		exit_reason TEXT,
		is_premature BOOLEAN DEFAULT 0,
		failure_type TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Prompt配置表
	CREATE TABLE IF NOT EXISTS prompt_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		section_name TEXT NOT NULL UNIQUE,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		prompt_type TEXT NOT NULL DEFAULT 'system',
		enabled BOOLEAN DEFAULT 1,
		display_order INTEGER DEFAULT 0,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- 当前持仓开仓时间表（用于系统重启后恢复）
	CREATE TABLE IF NOT EXISTS position_open_times (
		trader_id TEXT NOT NULL,
		symbol TEXT NOT NULL,
		side TEXT NOT NULL,
		open_time_ms INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (trader_id, symbol, side)
	);

	-- Trader运行状态表（用于系统重启后恢复）
	CREATE TABLE IF NOT EXISTS trader_states (
		trader_id TEXT PRIMARY KEY,
		is_paused BOOLEAN NOT NULL DEFAULT 0,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- 创建索引
	CREATE INDEX IF NOT EXISTS idx_decision_records_trader_id ON decision_records(trader_id);
	CREATE INDEX IF NOT EXISTS idx_decision_records_timestamp ON decision_records(timestamp);
	CREATE INDEX IF NOT EXISTS idx_decision_actions_record_id ON decision_actions(record_id);
	CREATE INDEX IF NOT EXISTS idx_decision_actions_symbol ON decision_actions(symbol);
	CREATE INDEX IF NOT EXISTS idx_position_snapshots_record_id ON position_snapshots(record_id);
	CREATE INDEX IF NOT EXISTS idx_trade_outcomes_trader_id ON trade_outcomes(trader_id);
	CREATE INDEX IF NOT EXISTS idx_trade_outcomes_symbol ON trade_outcomes(symbol);
	CREATE INDEX IF NOT EXISTS idx_trade_outcomes_close_time ON trade_outcomes(close_time);
	CREATE INDEX IF NOT EXISTS idx_prompt_configs_section_name ON prompt_configs(section_name);
	CREATE INDEX IF NOT EXISTS idx_prompt_configs_display_order ON prompt_configs(display_order);
	CREATE INDEX IF NOT EXISTS idx_position_open_times_trader ON position_open_times(trader_id);
	`

	_, err := c.db.Exec(schema)
	return err
}

// GetDBPath 获取数据库文件路径
func (c *Connection) GetDBPath() string {
	return c.dbPath
}

// GetBackupDir 获取备份目录路径
func (c *Connection) GetBackupDir() string {
	config := DefaultConfig()
	return filepath.Join(config.BaseDir, config.BackupDir, c.traderID)
}
