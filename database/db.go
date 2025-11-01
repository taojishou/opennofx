package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB 数据库连接管理器
type DB struct {
	conn     *sql.DB
	dbPath   string
	traderID string // Trader ID，用于区分不同trader的数据
}

// New 创建新的数据库连接
func New(traderID string) (*DB, error) {
	// 为每个trader创建独立的数据库文件
	dbDir := filepath.Join("decision_logs", traderID)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %w", err)
	}

	dbPath := filepath.Join(dbDir, "decisions.db")
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 设置连接池参数
	conn.SetMaxOpenConns(1) // SQLite 推荐单连接
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(0)

	db := &DB{
		conn:     conn,
		dbPath:   dbPath,
		traderID: traderID,
	}

	// 初始化表结构
	if err := db.initTables(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("初始化表结构失败: %w", err)
	}

	log.Printf("✓ SQLite数据库已初始化: %s", dbPath)
	return db, nil
}

// initTables 初始化数据库表结构
func (db *DB) initTables() error {
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

	_, err := db.conn.Exec(schema)
	if err != nil {
		return err
	}

	// 初始化默认prompt配置
	return db.initDefaultPrompts()
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// DecisionRecord 决策记录结构（数据库版本）
type DecisionRecord struct {
	ID           int64
	TraderID     string
	CycleNumber  int
	Timestamp    time.Time
	InputPrompt  string
	CoTTrace     string
	DecisionJSON string
	Success      bool
	ErrorMessage string
	// 账户状态快照
	TotalBalance          float64
	AvailableBalance      float64
	TotalUnrealizedProfit float64
	PositionCount         int
	MarginUsedPct         float64
}

// DecisionAction 决策动作结构
type DecisionAction struct {
	ID          int64
	RecordID    int64
	Action      string
	Symbol      string
	Quantity    float64
	Leverage    int
	Price       float64
	OrderID     int64
	Timestamp   time.Time
	Success     bool
	Error       string
	WasStopLoss bool
}

// PositionSnapshot 持仓快照结构
type PositionSnapshot struct {
	ID                int64
	RecordID          int64
	Symbol            string
	Side              string
	PositionAmt       float64
	EntryPrice        float64
	MarkPrice         float64
	UnrealizedProfit  float64
	Leverage          float64
	LiquidationPrice  float64
}

// AILearningSummary AI学习总结结构
type AILearningSummary struct {
	ID             int64
	TraderID       string
	SummaryContent string
	TradesCount    int
	DateRangeStart string
	DateRangeEnd   string
	WinRate        float64
	AvgPnL         float64
	CreatedAt      time.Time
	IsActive       bool
}

// TradeOutcome 交易结果结构
type TradeOutcome struct {
	ID              int64
	TraderID        string
	Symbol          string
	Side            string
	Quantity        float64
	Leverage        int
	OpenPrice       float64
	ClosePrice      float64
	PositionValue   float64
	MarginUsed      float64
	PnL             float64
	PnLPct          float64
	DurationMinutes int64
	OpenTime        time.Time
	CloseTime       time.Time
	WasStopLoss     bool
	EntryReason     string
	ExitReason      string
	IsPremature     bool
	FailureType     string
}

// InsertDecisionRecord 插入决策记录
func (db *DB) InsertDecisionRecord(record *DecisionRecord) (int64, error) {
	query := `
	INSERT INTO decision_records (
		trader_id, cycle_number, timestamp, input_prompt, cot_trace, decision_json,
		success, error_message, total_balance, available_balance, total_unrealized_profit,
		position_count, margin_used_pct
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		record.TraderID,
		record.CycleNumber,
		record.Timestamp,
		record.InputPrompt,
		record.CoTTrace,
		record.DecisionJSON,
		record.Success,
		record.ErrorMessage,
		record.TotalBalance,
		record.AvailableBalance,
		record.TotalUnrealizedProfit,
		record.PositionCount,
		record.MarginUsedPct,
	)

	if err != nil {
		return 0, fmt.Errorf("插入决策记录失败: %w", err)
	}

	return result.LastInsertId()
}

// InsertDecisionAction 插入决策动作
func (db *DB) InsertDecisionAction(action *DecisionAction) error {
	query := `
	INSERT INTO decision_actions (
		record_id, action, symbol, quantity, leverage, price, order_id,
		timestamp, success, error, was_stop_loss
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.conn.Exec(query,
		action.RecordID,
		action.Action,
		action.Symbol,
		action.Quantity,
		action.Leverage,
		action.Price,
		action.OrderID,
		action.Timestamp,
		action.Success,
		action.Error,
		action.WasStopLoss,
	)

	return err
}

// InsertPositionSnapshot 插入持仓快照
func (db *DB) InsertPositionSnapshot(position *PositionSnapshot) error {
	query := `
	INSERT INTO position_snapshots (
		record_id, symbol, side, position_amt, entry_price, mark_price,
		unrealized_profit, leverage, liquidation_price
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.conn.Exec(query,
		position.RecordID,
		position.Symbol,
		position.Side,
		position.PositionAmt,
		position.EntryPrice,
		position.MarkPrice,
		position.UnrealizedProfit,
		position.Leverage,
		position.LiquidationPrice,
	)

	return err
}

// InsertCandidateCoin 插入候选币种
func (db *DB) InsertCandidateCoin(recordID int64, symbol string) error {
	query := `INSERT INTO candidate_coins (record_id, symbol) VALUES (?, ?)`
	_, err := db.conn.Exec(query, recordID, symbol)
	return err
}

// InsertTradeOutcome 插入交易结果
func (db *DB) InsertTradeOutcome(trade *TradeOutcome) error {
	query := `
	INSERT INTO trade_outcomes (
		trader_id, symbol, side, quantity, leverage, open_price, close_price,
		position_value, margin_used, pnl, pnl_pct, duration_minutes,
		open_time, close_time, was_stop_loss, entry_reason, exit_reason,
		is_premature, failure_type
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.conn.Exec(query,
		trade.TraderID,
		trade.Symbol,
		trade.Side,
		trade.Quantity,
		trade.Leverage,
		trade.OpenPrice,
		trade.ClosePrice,
		trade.PositionValue,
		trade.MarginUsed,
		trade.PnL,
		trade.PnLPct,
		trade.DurationMinutes,
		trade.OpenTime,
		trade.CloseTime,
		trade.WasStopLoss,
		trade.EntryReason,
		trade.ExitReason,
		trade.IsPremature,
		trade.FailureType,
	)

	return err
}

// GetLatestRecords 获取最近N条决策记录
func (db *DB) GetLatestRecords(limit int) ([]*DecisionRecord, error) {
	query := `
	SELECT id, trader_id, cycle_number, timestamp, input_prompt, cot_trace, decision_json,
		success, error_message, total_balance, available_balance, total_unrealized_profit,
		position_count, margin_used_pct
	FROM decision_records
	WHERE trader_id = ?
	ORDER BY timestamp DESC
	LIMIT ?
	`

	rows, err := db.conn.Query(query, db.traderID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*DecisionRecord
	for rows.Next() {
		record := &DecisionRecord{}
		err := rows.Scan(
			&record.ID,
			&record.TraderID,
			&record.CycleNumber,
			&record.Timestamp,
			&record.InputPrompt,
			&record.CoTTrace,
			&record.DecisionJSON,
			&record.Success,
			&record.ErrorMessage,
			&record.TotalBalance,
			&record.AvailableBalance,
			&record.TotalUnrealizedProfit,
			&record.PositionCount,
			&record.MarginUsedPct,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	// 反转数组，让时间从旧到新排列
	for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
		records[i], records[j] = records[j], records[i]
	}

	return records, nil
}

// GetTradeOutcomes 获取最近N笔交易结果
func (db *DB) GetTradeOutcomes(limit int) ([]*TradeOutcome, error) {
	query := `
	SELECT id, trader_id, symbol, side, quantity, leverage, open_price, close_price,
		position_value, margin_used, pnl, pnl_pct, duration_minutes,
		open_time, close_time, was_stop_loss, entry_reason, exit_reason,
		is_premature, failure_type
	FROM trade_outcomes
	WHERE trader_id = ?
	ORDER BY close_time DESC
	LIMIT ?
	`

	rows, err := db.conn.Query(query, db.traderID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []*TradeOutcome
	for rows.Next() {
		trade := &TradeOutcome{}
		err := rows.Scan(
			&trade.ID,
			&trade.TraderID,
			&trade.Symbol,
			&trade.Side,
			&trade.Quantity,
			&trade.Leverage,
			&trade.OpenPrice,
			&trade.ClosePrice,
			&trade.PositionValue,
			&trade.MarginUsed,
			&trade.PnL,
			&trade.PnLPct,
			&trade.DurationMinutes,
			&trade.OpenTime,
			&trade.CloseTime,
			&trade.WasStopLoss,
			&trade.EntryReason,
			&trade.ExitReason,
			&trade.IsPremature,
			&trade.FailureType,
		)
		if err != nil {
			return nil, err
		}
		trades = append(trades, trade)
	}

	return trades, nil
}

// GetStatistics 获取统计数据
func (db *DB) GetStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总决策周期数
	var totalCycles int
	err := db.conn.QueryRow(`
		SELECT COUNT(*) FROM decision_records WHERE trader_id = ?
	`, db.traderID).Scan(&totalCycles)
	if err != nil {
		return nil, err
	}
	stats["total_cycles"] = totalCycles

	// 成功/失败周期数
	var successCycles, failedCycles int
	db.conn.QueryRow(`
		SELECT COUNT(*) FROM decision_records WHERE trader_id = ? AND success = 1
	`, db.traderID).Scan(&successCycles)
	db.conn.QueryRow(`
		SELECT COUNT(*) FROM decision_records WHERE trader_id = ? AND success = 0
	`, db.traderID).Scan(&failedCycles)
	stats["success_cycles"] = successCycles
	stats["failed_cycles"] = failedCycles

	// 交易统计
	var totalTrades, winningTrades, losingTrades int
	var totalPnL, avgWin, avgLoss float64

	db.conn.QueryRow(`
		SELECT COUNT(*) FROM trade_outcomes WHERE trader_id = ?
	`, db.traderID).Scan(&totalTrades)

	db.conn.QueryRow(`
		SELECT COUNT(*), COALESCE(AVG(pnl), 0) FROM trade_outcomes 
		WHERE trader_id = ? AND pnl > 0
	`, db.traderID).Scan(&winningTrades, &avgWin)

	db.conn.QueryRow(`
		SELECT COUNT(*), COALESCE(AVG(pnl), 0) FROM trade_outcomes 
		WHERE trader_id = ? AND pnl < 0
	`, db.traderID).Scan(&losingTrades, &avgLoss)

	db.conn.QueryRow(`
		SELECT COALESCE(SUM(pnl), 0) FROM trade_outcomes WHERE trader_id = ?
	`, db.traderID).Scan(&totalPnL)

	stats["total_trades"] = totalTrades
	stats["winning_trades"] = winningTrades
	stats["losing_trades"] = losingTrades
	stats["total_pnl"] = totalPnL
	stats["avg_win"] = avgWin
	stats["avg_loss"] = avgLoss

	if totalTrades > 0 {
		stats["win_rate"] = float64(winningTrades) / float64(totalTrades) * 100
	}

	return stats, nil
}

// QueryActions 查询指定记录的所有决策动作
func (db *DB) QueryActions(recordID int64) ([]*DecisionAction, error) {
	query := `
	SELECT id, record_id, action, symbol, quantity, leverage, price, order_id,
		timestamp, success, error, was_stop_loss
	FROM decision_actions
	WHERE record_id = ?
	ORDER BY timestamp ASC
	`

	rows, err := db.conn.Query(query, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []*DecisionAction
	for rows.Next() {
		action := &DecisionAction{}
		err := rows.Scan(
			&action.ID,
			&action.RecordID,
			&action.Action,
			&action.Symbol,
			&action.Quantity,
			&action.Leverage,
			&action.Price,
			&action.OrderID,
			&action.Timestamp,
			&action.Success,
			&action.Error,
			&action.WasStopLoss,
		)
		if err != nil {
			continue
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// PromptConfig Prompt配置结构
type PromptConfig struct {
	ID           int64     `json:"id"`
	SectionName  string    `json:"section_name"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Enabled      bool      `json:"enabled"`
	DisplayOrder int       `json:"display_order"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NewPromptConfig 创建新的Prompt配置
func (db *DB) NewPromptConfig(sectionName, title, content string, enabled bool, displayOrder int) *PromptConfig {
	return &PromptConfig{
		SectionName:  sectionName,
		Title:        title,
		Content:      content,
		Enabled:      enabled,
		DisplayOrder: displayOrder,
	}
}

// initDefaultPrompts 初始化默认prompt配置
func (db *DB) initDefaultPrompts() error {
	// 检查是否已经初始化
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM prompt_configs").Scan(&count)
	if err != nil {
		return err
	}
	
	if count > 0 {
		return nil // 已经初始化过了
	}

	log.Println("🔧 初始化默认Prompt配置...")

	defaults := []PromptConfig{
		{
			SectionName:  "core_mission",
			Title:        "🎯 核心目标",
			DisplayOrder: 1,
			Enabled:      true,
			Content: `**最大化夏普比率（Sharpe Ratio）**

夏普比率 = 平均收益 / 收益波动率

**这意味着**：
- ✅ 高质量交易（高胜率、大盈亏比）→ 提升夏普
- ✅ 稳定收益、控制回撤 → 提升夏普
- ✅ 耐心持仓、让利润奔跑 → 提升夏普
- ❌ 频繁交易、小盈小亏 → 增加波动，严重降低夏普
- ❌ 过度交易、手续费损耗 → 直接亏损
- ❌ 过早平仓、频繁进出 → 错失大行情

**关键认知**: 系统每3分钟扫描一次，但不意味着每次都要交易！
大多数时候应该是 wait 或 hold，只在极佳机会时才开仓。`,
		},
		{
			SectionName:  "hard_constraints",
			Title:        "⚖️ 硬约束（风险控制）",
			DisplayOrder: 2,
			Enabled:      true,
			Content: `1. **风险回报比**: 必须 ≥ 1:3（冒1%风险，赚3%+收益）
2. **最多持仓**: 由配置决定（用户提示中会显示持仓状态，注意查看上限值）
3. **单币仓位**: 
   - 山寨币: {{altMinSize}}-{{altMaxSize}} USDT ({{altcoinLeverage}}x杠杆)
   - BTC/ETH: {{btcMinSize}}-{{btcMaxSize}} USDT ({{btcEthLeverage}}x杠杆)
4. **保证金**: 总使用率 ≤ 90%`,
		},
		{
			SectionName:  "long_short_balance",
			Title:        "⚖️ 做多做空平衡",
			DisplayOrder: 3,
			Enabled:      true,
			Content: `**核心原则**: 做多和做空是完全平等的赚钱工具！

**判断标准**:
- 📈 上涨趋势 → 做多 (价格>EMA20>EMA50, MACD>0, RSI>50, 成交量放大)
- 📉 下跌趋势 → 做空 (价格<EMA20<EMA50, MACD<0, RSI<50, 成交量放大)
- ⏸️ 震荡市场 → 观望 (指标相互矛盾，方向不明确)

**重要等式**:
- 上涨5%做多的利润 = 下跌5%做空的利润
- 做多的风险 = 做空的风险
- 成功率不取决于方向，取决于趋势判断准确性

**严禁偏见**:
- ❌ 单边做多（错失下跌机会）
- ❌ 单边做空（错失上涨机会）
- ✅ 客观分析市场，跟随趋势`,
		},
		{
			SectionName:  "trading_frequency",
			Title:        "⏱️ 交易频率认知",
			DisplayOrder: 4,
			Enabled:      true,
			Content: `**量化标准**:
- 优秀交易员：每天2-4笔 = 每小时0.1-0.2笔
- 过度交易：每小时>2笔 = 严重问题
- 最佳节奏：开仓后持有至少30-60分钟

**自查**:
如果你发现自己每个周期都在交易 → 说明标准太低
如果你发现持仓<30分钟就平仓 → 说明太急躁`,
		},
		{
			SectionName:  "fee_awareness",
			Title:        "💰 手续费成本认知",
			DisplayOrder: 5,
			Enabled:      true,
			Content: `**每次交易的真实成本**:
- 开仓手续费：0.04% × 仓位价值（市价单Taker费率）
- 平仓手续费：0.04% × 仓位价值
- 总成本：0.08% ≈ 每100U仓位损失0.08 USDT
- 示例：150U仓位 = 0.12 USDT手续费

**盈亏平衡点**:
- 必须盈利>0.08%才能覆盖手续费
- 推荐目标：>0.3%（3-4倍手续费）才值得开仓
- 如果预期盈利<0.3%，不值得交易

**最小持仓时间**:
- 开仓后至少持有45分钟（除非止损）
- 过早平仓浪费手续费，数据显示持仓<30分钟胜率仅20-30%
- 耐心持有至少45-60分钟，让趋势充分发展

**交易频率惩罚**:
- 每天超过10笔交易 = 手续费>1.2 USDT
- 频繁进出可能吃掉所有利润
- 宁可少做，做精品交易`,
		},
		{
			SectionName:  "opening_standards",
			Title:        "🎯 开仓标准（严格）",
			DisplayOrder: 6,
			Enabled:      true,
			Content: `只在**强信号**时开仓，不确定就观望。

**你拥有的完整数据**：
- 📊 原始序列：3分钟价格序列(MidPrices数组) + 4小时K线序列
- 📈 技术序列：EMA20序列、MACD序列、RSI7序列、RSI14序列
- 💰 资金序列：成交量序列、持仓量(OI)序列、资金费率
- 🎯 筛选标记：AI500评分 / OI_Top排名（如果有标注）

**分析方法**（完全由你自主决定）：
- 自由运用序列数据，趋势分析、形态识别、支撑阻力等
- 多维度交叉验证（价格+量+OI+指标+序列形态）
- 用你认为最有效的方法发现高确定性机会
- 综合信心度 ≥ 75 才开仓

**避免低质量信号**：
- 单一维度（只看一个指标）
- 相互矛盾（涨但量萎缩）
- 横盘震荡
- 刚平仓不久（<15分钟）
- 预期盈利<0.3%（不值得支付手续费）`,
		},
		{
			SectionName:  "sharpe_optimization",
			Title:        "🧬 夏普比率自我进化",
			DisplayOrder: 7,
			Enabled:      true,
			Content: `每次你会收到**夏普比率**作为绩效反馈（周期级别）：

**夏普比率 < -0.5** (持续亏损):
  → 🛑 停止交易，连续观望至少6个周期（18分钟）
  → 🔍 深度反思：交易频率过高？持仓时间过短？信号强度不足？

**夏普比率 -0.5 ~ 0** (轻微亏损):
  → ⚠️ 严格控制：只做信心度>80的交易
  → 减少交易频率：每小时最多1笔新开仓
  → 耐心持仓：至少持有45分钟以上

**夏普比率 0 ~ 0.7** (正收益):
  → ✅ 维持当前策略

**夏普比率 > 0.7** (优异表现):
  → 🚀 可适度扩大仓位

**关键**: 夏普比率是唯一指标，它会自然惩罚频繁交易和过度进出。`,
		},
	}

	for _, cfg := range defaults {
		_, err := db.conn.Exec(`
			INSERT INTO prompt_configs (section_name, title, content, enabled, display_order)
			VALUES (?, ?, ?, ?, ?)
		`, cfg.SectionName, cfg.Title, cfg.Content, cfg.Enabled, cfg.DisplayOrder)
		
		if err != nil {
			return fmt.Errorf("插入默认prompt配置失败 [%s]: %w", cfg.SectionName, err)
		}
	}

	log.Println("✓ 默认Prompt配置初始化完成")
	return nil
}

// GetAllPromptConfigs 获取所有prompt配置
func (db *DB) GetAllPromptConfigs() ([]*PromptConfig, error) {
	query := `
		SELECT id, section_name, title, content, enabled, display_order, updated_at
		FROM prompt_configs
		ORDER BY display_order ASC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*PromptConfig
	for rows.Next() {
		cfg := &PromptConfig{}
		err := rows.Scan(&cfg.ID, &cfg.SectionName, &cfg.Title, &cfg.Content, 
			&cfg.Enabled, &cfg.DisplayOrder, &cfg.UpdatedAt)
		if err != nil {
			continue
		}
		configs = append(configs, cfg)
	}

	return configs, nil
}

// GetEnabledPromptConfigs 获取启用的prompt配置
func (db *DB) GetEnabledPromptConfigs() ([]*PromptConfig, error) {
	query := `
		SELECT id, section_name, title, content, enabled, display_order, updated_at
		FROM prompt_configs
		WHERE enabled = 1
		ORDER BY display_order ASC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*PromptConfig
	for rows.Next() {
		cfg := &PromptConfig{}
		err := rows.Scan(&cfg.ID, &cfg.SectionName, &cfg.Title, &cfg.Content, 
			&cfg.Enabled, &cfg.DisplayOrder, &cfg.UpdatedAt)
		if err != nil {
			continue
		}
		configs = append(configs, cfg)
	}

	return configs, nil
}

// UpdatePromptConfig 更新prompt配置
func (db *DB) UpdatePromptConfig(cfg *PromptConfig) error {
	query := `
		UPDATE prompt_configs 
		SET title = ?, content = ?, enabled = ?, display_order = ?, updated_at = CURRENT_TIMESTAMP
		WHERE section_name = ?
	`

	_, err := db.conn.Exec(query, cfg.Title, cfg.Content, cfg.Enabled, cfg.DisplayOrder, cfg.SectionName)
	return err
}

// SavePositionOpenTime 保存持仓开仓时间
func (db *DB) SavePositionOpenTime(symbol, side string, openTimeMs int64) error {
	query := `
		INSERT OR REPLACE INTO position_open_times (trader_id, symbol, side, open_time_ms)
		VALUES (?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query, db.traderID, symbol, side, openTimeMs)
	return err
}

// GetPositionOpenTime 获取持仓开仓时间
func (db *DB) GetPositionOpenTime(symbol, side string) (int64, bool) {
	query := `
		SELECT open_time_ms FROM position_open_times
		WHERE trader_id = ? AND symbol = ? AND side = ?
	`
	var openTimeMs int64
	err := db.conn.QueryRow(query, db.traderID, symbol, side).Scan(&openTimeMs)
	if err != nil {
		return 0, false
	}
	return openTimeMs, true
}

// DeletePositionOpenTime 删除持仓开仓时间
func (db *DB) DeletePositionOpenTime(symbol, side string) error {
	query := `
		DELETE FROM position_open_times
		WHERE trader_id = ? AND symbol = ? AND side = ?
	`
	_, err := db.conn.Exec(query, db.traderID, symbol, side)
	return err
}

// GetAllPositionOpenTimes 获取所有持仓开仓时间（用于系统启动时恢复）
func (db *DB) GetAllPositionOpenTimes() (map[string]int64, error) {
	query := `
		SELECT symbol, side, open_time_ms FROM position_open_times
		WHERE trader_id = ?
	`
	rows, err := db.conn.Query(query, db.traderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int64)
	for rows.Next() {
		var symbol, side string
		var openTimeMs int64
		if err := rows.Scan(&symbol, &side, &openTimeMs); err != nil {
			continue
		}
		key := symbol + "_" + side
		result[key] = openTimeMs
	}

	return result, nil
}

// SaveTraderState 保存Trader运行状态
func (db *DB) SaveTraderState(isPaused bool) error {
	query := `
		INSERT OR REPLACE INTO trader_states (trader_id, is_paused, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
	`
	_, err := db.conn.Exec(query, db.traderID, isPaused)
	return err
}

// GetTraderState 获取Trader运行状态
func (db *DB) GetTraderState() (isPaused bool, exists bool) {
	query := `
		SELECT is_paused FROM trader_states
		WHERE trader_id = ?
	`
	var paused int
	err := db.conn.QueryRow(query, db.traderID).Scan(&paused)
	if err != nil {
		return false, false
	}
	return paused == 1, true
}

// CleanOldRecords 清理N天前的旧记录
func (db *DB) CleanOldRecords(days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days)

	// 删除旧的决策记录（会级联删除关联数据）
	result, err := db.conn.Exec(`
		DELETE FROM decision_records 
		WHERE trader_id = ? AND timestamp < ?
	`, db.traderID, cutoffTime)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("🗑️ 已清理 %d 条旧记录（%d天前）", rowsAffected, days)
	}

	return nil
}

// SaveAILearningSummary 保存AI学习总结（将旧的设置为inactive）
func (db *DB) SaveAILearningSummary(summary *AILearningSummary) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 将该trader的所有旧总结设置为inactive
	_, err = tx.Exec(`UPDATE ai_learning_summaries SET is_active = 0 WHERE trader_id = ?`, db.traderID)
	if err != nil {
		return err
	}

	// 插入新总结
	_, err = tx.Exec(`
		INSERT INTO ai_learning_summaries (
			trader_id, summary_content, trades_count, date_range_start, date_range_end,
			win_rate, avg_pnl, is_active
		) VALUES (?, ?, ?, ?, ?, ?, ?, 1)
	`, db.traderID, summary.SummaryContent, summary.TradesCount, 
	   summary.DateRangeStart, summary.DateRangeEnd, summary.WinRate, summary.AvgPnL)
	
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetActiveAILearningSummary 获取当前激活的AI学习总结
func (db *DB) GetActiveAILearningSummary() (*AILearningSummary, error) {
	query := `
		SELECT id, trader_id, summary_content, trades_count, date_range_start, date_range_end,
		       win_rate, avg_pnl, created_at, is_active
		FROM ai_learning_summaries
		WHERE trader_id = ? AND is_active = 1
		ORDER BY created_at DESC
		LIMIT 1
	`
	
	var summary AILearningSummary
	var createdAtStr string
	
	err := db.conn.QueryRow(query, db.traderID).Scan(
		&summary.ID, &summary.TraderID, &summary.SummaryContent, &summary.TradesCount,
		&summary.DateRangeStart, &summary.DateRangeEnd, &summary.WinRate, &summary.AvgPnL,
		&createdAtStr, &summary.IsActive,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil // 没有总结，返回nil
	}
	if err != nil {
		return nil, err
	}
	
	summary.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	return &summary, nil
}

// GetAllAILearningSummaries 获取所有AI学习总结（用于前端展示历史）
func (db *DB) GetAllAILearningSummaries(limit int) ([]*AILearningSummary, error) {
	query := `
		SELECT id, trader_id, summary_content, trades_count, date_range_start, date_range_end,
		       win_rate, avg_pnl, created_at, is_active
		FROM ai_learning_summaries
		WHERE trader_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`
	
	rows, err := db.conn.Query(query, db.traderID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var summaries []*AILearningSummary
	for rows.Next() {
		var summary AILearningSummary
		var createdAtStr string
		
		err := rows.Scan(
			&summary.ID, &summary.TraderID, &summary.SummaryContent, &summary.TradesCount,
			&summary.DateRangeStart, &summary.DateRangeEnd, &summary.WinRate, &summary.AvgPnL,
			&createdAtStr, &summary.IsActive,
		)
		if err != nil {
			return nil, err
		}
		
		summary.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		summaries = append(summaries, &summary)
	}
	
	return summaries, nil
}
