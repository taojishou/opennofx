package repositories

import (
	"database/sql"
	"nofx/database/models"
)

// TradeRepository 交易结果数据访问层
type TradeRepository struct {
	db       *sql.DB
	traderID string
}

// NewTradeRepository 创建交易结果仓储
func NewTradeRepository(db *sql.DB, traderID string) *TradeRepository {
	return &TradeRepository{
		db:       db,
		traderID: traderID,
	}
}

// Insert 插入交易结果
func (r *TradeRepository) Insert(trade *models.TradeOutcome) error {
	query := `
	INSERT INTO trade_outcomes (
		trader_id, symbol, side, quantity, leverage, open_price, close_price,
		position_value, margin_used, pnl, pnl_pct, duration_minutes,
		open_time, close_time, was_stop_loss, entry_reason, exit_reason,
		is_premature, failure_type
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
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

// GetLatest 获取最近N笔交易结果
func (r *TradeRepository) GetLatest(limit int) ([]*models.TradeOutcome, error) {
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

	rows, err := r.db.Query(query, r.traderID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []*models.TradeOutcome
	for rows.Next() {
		trade := &models.TradeOutcome{}
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

// GetStatistics 获取交易统计
func (r *TradeRepository) GetStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 交易统计
	var totalTrades, winningTrades, losingTrades int
	var totalPnL, avgWin, avgLoss float64

	r.db.QueryRow(`
		SELECT COUNT(*) FROM trade_outcomes WHERE trader_id = ?
	`, r.traderID).Scan(&totalTrades)

	r.db.QueryRow(`
		SELECT COUNT(*), COALESCE(AVG(pnl), 0) FROM trade_outcomes 
		WHERE trader_id = ? AND pnl > 0
	`, r.traderID).Scan(&winningTrades, &avgWin)

	r.db.QueryRow(`
		SELECT COUNT(*), COALESCE(AVG(pnl), 0) FROM trade_outcomes 
		WHERE trader_id = ? AND pnl < 0
	`, r.traderID).Scan(&losingTrades, &avgLoss)

	r.db.QueryRow(`
		SELECT COALESCE(SUM(pnl), 0) FROM trade_outcomes WHERE trader_id = ?
	`, r.traderID).Scan(&totalPnL)

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

// DeleteOld 删除N天前的旧记录
func (r *TradeRepository) DeleteOld(days int) (int64, error) {
	query := `
		DELETE FROM trade_outcomes 
		WHERE trader_id = ? AND close_time < datetime('now', '-' || ? || ' days')
	`
	result, err := r.db.Exec(query, r.traderID, days)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
