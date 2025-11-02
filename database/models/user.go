package models

import "time"

// User 用户表
type User struct {
	ID        int64
	Username  string
	Email     string
	Password  string // 哈希后的密码
	Role      string // admin, user
	CreatedAt time.Time
	UpdatedAt time.Time
	IsActive  bool
}

// Session 用户会话表
type Session struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}
