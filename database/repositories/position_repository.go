package repositories

import (
	"database/sql"
	"nofx/database/models"
)

// PositionRepository 持仓管理数据访问层
type PositionRepository struct {
	db       *sql.DB
	traderID string
}

// NewPositionRepository 创建持仓管理仓储
func NewPositionRepository(db *sql.DB, traderID string) *PositionRepository {
	return &PositionRepository{
		db:       db,
		traderID: traderID,
	}
}

// SaveOpenTime 保存持仓开仓时间
func (r *PositionRepository) SaveOpenTime(symbol, side string, openTimeMs int64) error {
	query := `
		INSERT OR REPLACE INTO position_open_times (trader_id, symbol, side, open_time_ms)
		VALUES (?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, r.traderID, symbol, side, openTimeMs)
	return err
}

// GetOpenTime 获取持仓开仓时间
func (r *PositionRepository) GetOpenTime(symbol, side string) (int64, bool) {
	query := `
		SELECT open_time_ms FROM position_open_times
		WHERE trader_id = ? AND symbol = ? AND side = ?
	`
	var openTimeMs int64
	err := r.db.QueryRow(query, r.traderID, symbol, side).Scan(&openTimeMs)
	if err != nil {
		return 0, false
	}
	return openTimeMs, true
}

// DeleteOpenTime 删除持仓开仓时间
func (r *PositionRepository) DeleteOpenTime(symbol, side string) error {
	query := `
		DELETE FROM position_open_times
		WHERE trader_id = ? AND symbol = ? AND side = ?
	`
	_, err := r.db.Exec(query, r.traderID, symbol, side)
	return err
}

// GetAllOpenTimes 获取所有持仓开仓时间（用于系统启动时恢复）
func (r *PositionRepository) GetAllOpenTimes() (map[string]int64, error) {
	query := `
		SELECT symbol, side, open_time_ms FROM position_open_times
		WHERE trader_id = ?
	`
	rows, err := r.db.Query(query, r.traderID)
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
func (r *PositionRepository) SaveTraderState(isPaused bool) error {
	query := `
		INSERT OR REPLACE INTO trader_states (trader_id, is_paused, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
	`
	_, err := r.db.Exec(query, r.traderID, isPaused)
	return err
}

// GetTraderState 获取Trader运行状态
func (r *PositionRepository) GetTraderState() (*models.TraderState, error) {
	query := `
		SELECT trader_id, is_paused, updated_at FROM trader_states
		WHERE trader_id = ?
	`
	state := &models.TraderState{}
	var pausedInt int
	err := r.db.QueryRow(query, r.traderID).Scan(&state.TraderID, &pausedInt, &state.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有保存的状态
		}
		return nil, err
	}
	state.IsPaused = (pausedInt == 1)
	return state, nil
}
