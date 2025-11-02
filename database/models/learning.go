package models

import "time"

// AILearningSummary AI学习总结表
type AILearningSummary struct {
	ID int64
	TraderID string
	SummaryContent string
	TradesCount int
	DateRangeStart string
	DateRangeEnd string
	WinRate float64
	AvgPnL float64
	CreatedAt time.Time
	IsActive bool
}
