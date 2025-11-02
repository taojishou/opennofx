package database

import (
	"nofx/database/models"
	"nofx/database/repositories"
)

// DB 简化的数据库接口（用于 decision_logger 等组件）
type DB struct {
	conn     *Connection
	traderID string
}

// New 创建新的数据库连接
func New(traderID string) (*DB, error) {
	conn, err := NewConnection(traderID)
	if err != nil {
		return nil, err
	}

	return &DB{
		conn:     conn,
		traderID: traderID,
	}, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Backup 创建数据库备份
func (db *DB) Backup(timestamp string) error {
	return db.conn.Backup(timestamp)
}

// BuildSystemPromptFromDB 从数据库构建system prompt
// maxPositionValueBTC和maxPositionValueAlt是动态风控调整后的实际可用限制
// aiAutonomyMode: true=自主模式（移除限制性规则），false=限制模式（包含所有规则）
func (db *DB) BuildSystemPromptFromDB(accountEquity float64, btcEthLeverage, altcoinLeverage int, maxPositionValueBTC, maxPositionValueAlt float64, aiAutonomyMode bool) string {
	repo := repositories.NewConfigRepository(db.conn.DB())
	return BuildSystemPrompt(repo, accountEquity, btcEthLeverage, altcoinLeverage, maxPositionValueBTC, maxPositionValueAlt, aiAutonomyMode)
}

// GetUserPromptTemplates 获取用户提示词模板
func (db *DB) GetUserPromptTemplates() ([]*repositories.PromptConfig, error) {
	repo := repositories.NewConfigRepository(db.conn.DB())
	return repo.GetByType("user")
}

// Decision 获取决策Repository
func (db *DB) Decision() *repositories.DecisionRepository {
	return repositories.NewDecisionRepository(db.conn.DB(), db.traderID)
}

// Trade 获取交易Repository
func (db *DB) Trade() *repositories.TradeRepository {
	return repositories.NewTradeRepository(db.conn.DB(), db.traderID)
}

// Position 获取持仓Repository
func (db *DB) Position() *repositories.PositionRepository {
	return repositories.NewPositionRepository(db.conn.DB(), db.traderID)
}

// Learning 获取学习Repository
func (db *DB) Learning() *repositories.LearningRepository {
	return repositories.NewLearningRepository(db.conn.DB(), db.traderID)
}

// Config 获取配置Repository
func (db *DB) Config() *repositories.ConfigRepository {
	return repositories.NewConfigRepository(db.conn.DB())
}

// GetLatestRecords 获取最近N条决策记录（兼容方法）
func (db *DB) GetLatestRecords(limit int) ([]*models.DecisionRecord, error) {
	return db.Decision().GetLatest(limit)
}

// GetAllPositionOpenTimes 获取所有持仓开仓时间
func (db *DB) GetAllPositionOpenTimes() (map[string]int64, error) {
	return db.Position().GetAllOpenTimes()
}

// GetTraderState 获取Trader状态
func (db *DB) GetTraderState() (bool, bool) {
	state, err := db.Position().GetTraderState()
	if err != nil || state == nil {
		return false, false
	}
	return state.IsPaused, true
}

// GetPositionOpenTime 获取持仓开仓时间
func (db *DB) GetPositionOpenTime(symbol, side string) (int64, bool) {
	return db.Position().GetOpenTime(symbol, side)
}

// DeletePositionOpenTime 删除持仓开仓时间
func (db *DB) DeletePositionOpenTime(symbol, side string) error {
	return db.Position().DeleteOpenTime(symbol, side)
}

// SavePositionOpenTime 保存持仓开仓时间
func (db *DB) SavePositionOpenTime(symbol, side string, openTimeMs int64) error {
	return db.Position().SaveOpenTime(symbol, side, openTimeMs)
}

// SaveTraderState 保存Trader状态
func (db *DB) SaveTraderState(isPaused bool) error {
	return db.Position().SaveTraderState(isPaused)
}

// GetActiveAILearningSummary 获取活跃的AI学习总结
func (db *DB) GetActiveAILearningSummary() (*models.AILearningSummary, error) {
	return db.Learning().GetActive()
}

// GetTradeOutcomes 获取最近N笔交易结果
func (db *DB) GetTradeOutcomes(limit int) ([]*models.TradeOutcome, error) {
	return db.Trade().GetLatest(limit)
}

// SaveAILearningSummary 保存AI学习总结
func (db *DB) SaveAILearningSummary(summary *models.AILearningSummary) error {
	return db.Learning().Save(summary)
}
