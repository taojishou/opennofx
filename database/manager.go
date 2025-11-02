package database

import (
	"fmt"
	"nofx/database/repositories"
)

// Manager 数据库管理器（统一管理所有数据库连接和Repository）
type Manager struct {
	// 系统数据库连接
	systemConn *SystemConnection
	
	// Trader数据库连接（按traderID索引）
	traderConns map[string]*Connection
	
	// 系统级Repository
	UserRepo          *repositories.UserRepository
	SystemConfigRepo  *repositories.SystemConfigRepository
	TraderConfigRepo  *repositories.TraderConfigRepository
}

// NewManager 创建数据库管理器
func NewManager() (*Manager, error) {
	// 初始化系统数据库
	systemConn, err := NewSystemConnection()
	if err != nil {
		return nil, fmt.Errorf("初始化系统数据库失败: %w", err)
	}

	manager := &Manager{
		systemConn:  systemConn,
		traderConns: make(map[string]*Connection),
		// 初始化系统级Repository
		UserRepo:         repositories.NewUserRepository(systemConn.DB()),
		SystemConfigRepo: repositories.NewSystemConfigRepository(systemConn.DB()),
		TraderConfigRepo: repositories.NewTraderConfigRepository(systemConn.DB()),
	}

	return manager, nil
}

// GetTraderConnection 获取或创建Trader数据库连接
func (m *Manager) GetTraderConnection(traderID string) (*Connection, error) {
	// 检查是否已存在连接
	if conn, exists := m.traderConns[traderID]; exists {
		return conn, nil
	}

	// 创建新连接
	conn, err := NewConnection(traderID)
	if err != nil {
		return nil, fmt.Errorf("创建Trader数据库连接失败: %w", err)
	}

	m.traderConns[traderID] = conn
	return conn, nil
}

// GetDecisionRepo 获取决策记录Repository
func (m *Manager) GetDecisionRepo(traderID string) (*repositories.DecisionRepository, error) {
	conn, err := m.GetTraderConnection(traderID)
	if err != nil {
		return nil, err
	}
	return repositories.NewDecisionRepository(conn.DB(), traderID), nil
}

// GetTradeRepo 获取交易结果Repository
func (m *Manager) GetTradeRepo(traderID string) (*repositories.TradeRepository, error) {
	conn, err := m.GetTraderConnection(traderID)
	if err != nil {
		return nil, err
	}
	return repositories.NewTradeRepository(conn.DB(), traderID), nil
}

// GetPositionRepo 获取持仓管理Repository
func (m *Manager) GetPositionRepo(traderID string) (*repositories.PositionRepository, error) {
	conn, err := m.GetTraderConnection(traderID)
	if err != nil {
		return nil, err
	}
	return repositories.NewPositionRepository(conn.DB(), traderID), nil
}

// GetLearningRepo 获取AI学习Repository
func (m *Manager) GetLearningRepo(traderID string) (*repositories.LearningRepository, error) {
	conn, err := m.GetTraderConnection(traderID)
	if err != nil {
		return nil, err
	}
	return repositories.NewLearningRepository(conn.DB(), traderID), nil
}

// GetConfigRepo 获取Prompt配置Repository
func (m *Manager) GetConfigRepo(traderID string) (*repositories.ConfigRepository, error) {
	conn, err := m.GetTraderConnection(traderID)
	if err != nil {
		return nil, err
	}
	return repositories.NewConfigRepository(conn.DB()), nil
}

// Close 关闭所有数据库连接
func (m *Manager) Close() error {
	var lastErr error

	// 关闭系统数据库
	if m.systemConn != nil {
		if err := m.systemConn.Close(); err != nil {
			lastErr = err
		}
	}

	// 关闭所有Trader数据库
	for traderID, conn := range m.traderConns {
		if err := conn.Close(); err != nil {
			lastErr = fmt.Errorf("关闭Trader[%s]数据库失败: %w", traderID, err)
		}
	}

	return lastErr
}

// BackupTraderDB 备份指定Trader的数据库
func (m *Manager) BackupTraderDB(traderID, timestamp string) error {
	conn, err := m.GetTraderConnection(traderID)
	if err != nil {
		return err
	}
	return conn.Backup(timestamp)
}
