package models

import "time"

// TradeOutcome 交易结果表（用于统计分析）
type TradeOutcome struct {
	ID int64
	TraderID string
	Symbol string
	Side string
	Quantity float64
	Leverage int
	OpenPrice float64
	ClosePrice float64
	PositionValue float64
	MarginUsed float64
	PnL float64
	PnLPct float64
	DurationMinutes int64
	OpenTime time.Time
	CloseTime time.Time
	WasStopLoss bool
	EntryReason string
	ExitReason string
	IsPremature bool
	FailureType string
	CreatedAt time.Time
}
