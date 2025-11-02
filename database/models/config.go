package models

import "time"

// PromptConfig Prompt配置表
type PromptConfig struct {
	ID           int64     `json:"id"`
	SectionName  string    `json:"section_name"`  // 配置section名称
	Title        string    `json:"title"`         // 显示标题
	Content      string    `json:"content"`       // 内容
	PromptType   string    `json:"prompt_type"`   // 类型: system / user
	Enabled      bool      `json:"enabled"`       // 是否启用
	DisplayOrder int       `json:"display_order"` // 显示顺序
	UpdatedAt    time.Time `json:"updated_at"`
}
