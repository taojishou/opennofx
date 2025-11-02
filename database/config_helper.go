package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
)

// ConfigHelper 配置辅助工具，用于从数据库读取配置
type ConfigHelper struct {
	db *sql.DB
}

// NewConfigHelper 创建配置辅助工具
func NewConfigHelper(db *sql.DB) *ConfigHelper {
	return &ConfigHelper{db: db}
}

// GetString 获取字符串配置
func (h *ConfigHelper) GetString(key, defaultValue string) string {
	var value string
	err := h.db.QueryRow("SELECT value FROM system_configs WHERE key = ?", key).Scan(&value)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetInt 获取整数配置
func (h *ConfigHelper) GetInt(key string, defaultValue int) int {
	value := h.GetString(key, "")
	if value == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intVal
}

// GetFloat 获取浮点数配置
func (h *ConfigHelper) GetFloat(key string, defaultValue float64) float64 {
	value := h.GetString(key, "")
	if value == "" {
		return defaultValue
	}
	floatVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return floatVal
}

// GetBool 获取布尔值配置
func (h *ConfigHelper) GetBool(key string, defaultValue bool) bool {
	value := h.GetString(key, "")
	if value == "" {
		return defaultValue
	}
	boolVal, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolVal
}

// GetJSON 获取JSON配置并反序列化
func (h *ConfigHelper) GetJSON(key string, target interface{}, defaultValue interface{}) error {
	value := h.GetString(key, "")
	if value == "" {
		// 使用默认值
		if defaultValue != nil {
			b, _ := json.Marshal(defaultValue)
			return json.Unmarshal(b, target)
		}
		return fmt.Errorf("配置不存在且无默认值")
	}
	return json.Unmarshal([]byte(value), target)
}

// SetString 设置字符串配置
func (h *ConfigHelper) SetString(key, value, description, configType string) error {
	_, err := h.db.Exec(`
		INSERT INTO system_configs (key, value, description, config_type, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = CURRENT_TIMESTAMP
	`, key, value, description, configType)
	return err
}

// SetInt 设置整数配置
func (h *ConfigHelper) SetInt(key string, value int, description, configType string) error {
	return h.SetString(key, strconv.Itoa(value), description, configType)
}

// SetFloat 设置浮点数配置
func (h *ConfigHelper) SetFloat(key string, value float64, description, configType string) error {
	return h.SetString(key, strconv.FormatFloat(value, 'f', -1, 64), description, configType)
}

// SetBool 设置布尔值配置
func (h *ConfigHelper) SetBool(key string, value bool, description, configType string) error {
	return h.SetString(key, strconv.FormatBool(value), description, configType)
}

// SetJSON 设置JSON配置
func (h *ConfigHelper) SetJSON(key string, value interface{}, description, configType string) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return h.SetString(key, string(b), description, configType)
}

// GetAllByType 获取指定类型的所有配置
func (h *ConfigHelper) GetAllByType(configType string) (map[string]string, error) {
	rows, err := h.db.Query("SELECT key, value FROM system_configs WHERE config_type = ?", configType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	configs := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		configs[key] = value
	}
	return configs, nil
}
