package repositories

import (
	"database/sql"
	"fmt"
	"nofx/database/models"
)

// DecisionRepository 决策记录数据访问层
type DecisionRepository struct {
	db       *sql.DB
	traderID string
}

// NewDecisionRepository 创建决策记录仓储
func NewDecisionRepository(db *sql.DB, traderID string) *DecisionRepository {
	return &DecisionRepository{
		db:       db,
		traderID: traderID,
	}
}

// Insert 插入决策记录
func (r *DecisionRepository) Insert(record *models.DecisionRecord) (int64, error) {
	query := `
	INSERT INTO decision_records (
		trader_id, cycle_number, timestamp, system_prompt, input_prompt, cot_trace, decision_json,
		success, error_message, total_balance, available_balance, total_unrealized_profit,
		position_count, margin_used_pct
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		record.TraderID,
		record.CycleNumber,
		record.Timestamp,
		record.SystemPrompt,
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

// GetLatest 获取最近N条决策记录
func (r *DecisionRepository) GetLatest(limit int) ([]*models.DecisionRecord, error) {
	query := `
	SELECT id, trader_id, cycle_number, timestamp, 
		COALESCE(system_prompt, '') as system_prompt, 
		COALESCE(input_prompt, '') as input_prompt, 
		COALESCE(cot_trace, '') as cot_trace, 
		COALESCE(decision_json, '') as decision_json,
		success, 
		COALESCE(error_message, '') as error_message, 
		total_balance, available_balance, total_unrealized_profit,
		position_count, margin_used_pct
	FROM decision_records
	WHERE trader_id = ?
	ORDER BY timestamp DESC
	LIMIT ?
	`

	rows, err := r.db.Query(query, r.traderID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.DecisionRecord
	for rows.Next() {
		record := &models.DecisionRecord{}
		err := rows.Scan(
			&record.ID,
			&record.TraderID,
			&record.CycleNumber,
			&record.Timestamp,
			&record.SystemPrompt,
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

// InsertAction 插入决策动作
func (r *DecisionRepository) InsertAction(action *models.DecisionAction) error {
	query := `
	INSERT INTO decision_actions (
		record_id, action, symbol, quantity, leverage, price, order_id,
		timestamp, success, error, was_stop_loss
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
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

// GetActions 查询指定记录的所有决策动作
func (r *DecisionRepository) GetActions(recordID int64) ([]*models.DecisionAction, error) {
	query := `
	SELECT id, record_id, action, symbol, quantity, leverage, price, order_id,
		timestamp, success, error, was_stop_loss
	FROM decision_actions
	WHERE record_id = ?
	ORDER BY timestamp ASC
	`

	rows, err := r.db.Query(query, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []*models.DecisionAction
	for rows.Next() {
		action := &models.DecisionAction{}
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

// InsertPositionSnapshot 插入持仓快照
func (r *DecisionRepository) InsertPositionSnapshot(position *models.PositionSnapshot) error {
	query := `
	INSERT INTO position_snapshots (
		record_id, symbol, side, position_amt, entry_price, mark_price,
		unrealized_profit, leverage, liquidation_price
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
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
func (r *DecisionRepository) InsertCandidateCoin(recordID int64, symbol string) error {
	query := `INSERT INTO candidate_coins (record_id, symbol) VALUES (?, ?)`
	_, err := r.db.Exec(query, recordID, symbol)
	return err
}

// GetStatistics 获取统计数据
func (r *DecisionRepository) GetStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总决策周期数
	var totalCycles int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM decision_records WHERE trader_id = ?
	`, r.traderID).Scan(&totalCycles)
	if err != nil {
		return nil, err
	}
	stats["total_cycles"] = totalCycles

	// 成功/失败周期数
	var successCycles, failedCycles int
	r.db.QueryRow(`
		SELECT COUNT(*) FROM decision_records WHERE trader_id = ? AND success = 1
	`, r.traderID).Scan(&successCycles)
	r.db.QueryRow(`
		SELECT COUNT(*) FROM decision_records WHERE trader_id = ? AND success = 0
	`, r.traderID).Scan(&failedCycles)
	stats["success_cycles"] = successCycles
	stats["failed_cycles"] = failedCycles

	return stats, nil
}
