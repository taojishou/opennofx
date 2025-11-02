package database

import (
	"os"
	"path/filepath"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	BaseDir     string // 数据库基础目录
	TraderDir   string // 交易员数据目录
	BackupDir   string // 备份目录
	LogsDir     string // 日志目录
}

// DefaultConfig 返回默认的数据库配置
func DefaultConfig() *DatabaseConfig {
	return &DatabaseConfig{
		BaseDir:   "data",
		TraderDir: "traders",
		BackupDir: "backups",
		LogsDir:   "logs",
	}
}

// GetTraderDBPath 获取指定交易员的数据库路径
func (c *DatabaseConfig) GetTraderDBPath(traderID string) string {
	return filepath.Join(c.BaseDir, c.TraderDir, traderID, "decisions.db")
}

// GetTraderDir 获取指定交易员的数据目录
func (c *DatabaseConfig) GetTraderDir(traderID string) string {
	return filepath.Join(c.BaseDir, c.TraderDir, traderID)
}

// GetBackupPath 获取备份文件路径
func (c *DatabaseConfig) GetBackupPath(traderID, timestamp string) string {
	return filepath.Join(c.BaseDir, c.BackupDir, traderID, timestamp+".db")
}

// GetLogsDir 获取日志目录
func (c *DatabaseConfig) GetLogsDir() string {
	return filepath.Join(c.BaseDir, c.LogsDir)
}

// EnsureDirectories 确保所有必要的目录存在
func (c *DatabaseConfig) EnsureDirectories(traderID string) error {
	dirs := []string{
		c.GetTraderDir(traderID),
		filepath.Join(c.BaseDir, c.BackupDir, traderID),
		c.GetLogsDir(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// GetAllTraders 获取所有交易员ID列表
func (c *DatabaseConfig) GetAllTraders() ([]string, error) {
	tradersDir := filepath.Join(c.BaseDir, c.TraderDir)
	
	entries, err := os.ReadDir(tradersDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var traders []string
	for _, entry := range entries {
		if entry.IsDir() {
			traders = append(traders, entry.Name())
		}
	}

	return traders, nil
}

// CleanupOldBackups 清理旧的备份文件（保留最近N个）
func (c *DatabaseConfig) CleanupOldBackups(traderID string, keepCount int) error {
	backupDir := filepath.Join(c.BaseDir, c.BackupDir, traderID)
	
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 目录不存在，无需清理
		}
		return err
	}

	if len(entries) <= keepCount {
		return nil // 备份数量未超过限制
	}

	// 按修改时间排序，删除最旧的文件
	for i := 0; i < len(entries)-keepCount; i++ {
		oldFile := filepath.Join(backupDir, entries[i].Name())
		if err := os.Remove(oldFile); err != nil {
			return err
		}
	}

	return nil
}