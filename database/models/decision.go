package models

import "time"

// DecisionRecord 决策记录主表
type DecisionRecord struct {
	ID        int64
	TraderID  string
	CycleNumber int
	Timestamp time.Time
	SystemPrompt string
	InputPrompt string
	CoTTrace string
	DecisionJSON string
	Success bool
	ErrorMessage string
	// 账户状态快照
	TotalBalance float64
	AvailableBalance float64
	TotalUnrealizedProfit float64
	PositionCount int
	MarginUsedPct float64
	CreatedAt time.Time
}

// DecisionAction 决策动作表
type DecisionAction struct {
	ID int64
	RecordID int64
	Action string
	Symbol string
	Quantity float64
	Leverage int
	Price float64
	OrderID int64
	Timestamp time.Time
	Success bool
	Error string
	WasStopLoss bool
}

// PositionSnapshot 持仓快照表（关联决策记录）
type PositionSnapshot struct {
	ID int64
	RecordID int64
	Symbol string
	Side string
	PositionAmt float64
	EntryPrice float64
	MarkPrice float64
	UnrealizedProfit float64
	Leverage float64
	LiquidationPrice float64
}

// CandidateCoin 候选币种表（关联决策记录）
type CandidateCoin struct {
	ID int64
	RecordID int64
	Symbol string
}
