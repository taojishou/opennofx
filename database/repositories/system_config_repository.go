package repositories

import (
	"database/sql"
	"nofx/database/models"
)

// SystemConfigRepository 系统配置数据访问层
type SystemConfigRepository struct {
	db *sql.DB
}

// NewSystemConfigRepository 创建系统配置仓储
func NewSystemConfigRepository(db *sql.DB) *SystemConfigRepository {
	return &SystemConfigRepository{db: db}
}

// Get 获取配置项
func (r *SystemConfigRepository) Get(key string) (*models.SystemConfig, error) {
	query := `
		SELECT id, key, value, description, config_type, updated_at
		FROM system_configs WHERE key = ?
	`
	config := &models.SystemConfig{}
	err := r.db.QueryRow(query, key).Scan(
		&config.ID, &config.Key, &config.Value,
		&config.Description, &config.ConfigType, &config.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// Set 设置配置项
func (r *SystemConfigRepository) Set(key, value, description, configType string) error {
	query := `
		INSERT OR REPLACE INTO system_configs (key, value, description, config_type, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`
	_, err := r.db.Exec(query, key, value, description, configType)
	return err
}

// GetByType 获取指定类型的所有配置
func (r *SystemConfigRepository) GetByType(configType string) ([]*models.SystemConfig, error) {
	query := `
		SELECT id, key, value, description, config_type, updated_at
		FROM system_configs WHERE config_type = ?
		ORDER BY key
	`
	rows, err := r.db.Query(query, configType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*models.SystemConfig
	for rows.Next() {
		config := &models.SystemConfig{}
		err := rows.Scan(
			&config.ID, &config.Key, &config.Value,
			&config.Description, &config.ConfigType, &config.UpdatedAt,
		)
		if err != nil {
			continue
		}
		configs = append(configs, config)
	}
	return configs, nil
}

// GetAll 获取所有配置
func (r *SystemConfigRepository) GetAll() ([]*models.SystemConfig, error) {
	query := `
		SELECT id, key, value, description, config_type, updated_at
		FROM system_configs ORDER BY config_type, key
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*models.SystemConfig
	for rows.Next() {
		config := &models.SystemConfig{}
		err := rows.Scan(
			&config.ID, &config.Key, &config.Value,
			&config.Description, &config.ConfigType, &config.UpdatedAt,
		)
		if err != nil {
			continue
		}
		configs = append(configs, config)
	}
	return configs, nil
}

// Delete 删除配置项
func (r *SystemConfigRepository) Delete(key string) error {
	query := `DELETE FROM system_configs WHERE key = ?`
	_, err := r.db.Exec(query, key)
	return err
}
