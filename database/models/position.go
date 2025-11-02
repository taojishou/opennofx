package models

import "time"

// PositionOpenTime 持仓开仓时间表（用于系统重启后恢复）
type PositionOpenTime struct {
	TraderID string
	Symbol string
	Side string
	OpenTimeMs int64
	CreatedAt time.Time
}

// TraderState Trader运行状态表（用于系统重启后恢复）
type TraderState struct {
	TraderID string
	IsPaused bool
	UpdatedAt time.Time
}
