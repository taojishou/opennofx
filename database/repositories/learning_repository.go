package repositories

import (
	"database/sql"
	"nofx/database/models"
	"time"
)

// LearningRepository AI学习数据访问层
type LearningRepository struct {
	db       *sql.DB
	traderID string
}

// NewLearningRepository 创建AI学习仓储
func NewLearningRepository(db *sql.DB, traderID string) *LearningRepository {
	return &LearningRepository{
		db:       db,
		traderID: traderID,
	}
}

// Save 保存AI学习总结（将旧的设置为inactive）
func (r *LearningRepository) Save(summary *models.AILearningSummary) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 将该trader的所有旧总结设置为inactive
	_, err = tx.Exec(`UPDATE ai_learning_summaries SET is_active = 0 WHERE trader_id = ?`, r.traderID)
	if err != nil {
		return err
	}

	// 插入新总结
	_, err = tx.Exec(`
		INSERT INTO ai_learning_summaries (
			trader_id, summary_content, trades_count, date_range_start, date_range_end,
			win_rate, avg_pnl, is_active
		) VALUES (?, ?, ?, ?, ?, ?, ?, 1)
	`, r.traderID, summary.SummaryContent, summary.TradesCount,
		summary.DateRangeStart, summary.DateRangeEnd, summary.WinRate, summary.AvgPnL)

	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetActive 获取当前激活的AI学习总结
func (r *LearningRepository) GetActive() (*models.AILearningSummary, error) {
	query := `
		SELECT id, trader_id, summary_content, trades_count, date_range_start, date_range_end,
		       win_rate, avg_pnl, created_at, is_active
		FROM ai_learning_summaries
		WHERE trader_id = ? AND is_active = 1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var summary models.AILearningSummary
	var createdAtStr string

	err := r.db.QueryRow(query, r.traderID).Scan(
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

// GetAll 获取所有AI学习总结（用于前端展示历史）
func (r *LearningRepository) GetAll(limit int) ([]*models.AILearningSummary, error) {
	query := `
		SELECT id, trader_id, summary_content, trades_count, date_range_start, date_range_end,
		       win_rate, avg_pnl, created_at, is_active
		FROM ai_learning_summaries
		WHERE trader_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, r.traderID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []*models.AILearningSummary
	for rows.Next() {
		var summary models.AILearningSummary
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
