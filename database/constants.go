package database

// 数据库文件名常量
const (
	SystemDBName  = "system.db"
	TradingDBName = "trading.db"
)

// 默认目录结构常量
const (
	DefaultBaseDir    = "data"
	DefaultTraderDir  = "traders"
	DefaultBackupDir  = "backups"
	DefaultLogsDir    = "logs"
	DefaultCacheDir   = "cache"
)

// 查询限制常量
const (
	// 默认记录查询数量
	DefaultRecordLimit = 100
	
	// 性能分析相关
	PerformanceAnalysisLimit = 100 // 性能分析记录数量
	MonitoringRecordLimit    = 50  // 监控使用的记录数量
	RecentPerformanceLimit   = 20  // 近期表现记录数量
	
	// 交易结果相关
	TradeOutcomesLookbackLimit = 50  // 交易结果回顾数量
	TradeAnalysisMinTrades     = 10  // 交易分析最小交易数
	
	// 回溯窗口倍数
	LookbackWindowMultiplier = 3 // 获取决策记录时的窗口放大倍数
)

// 备份管理常量
const (
	DefaultBackupRetention = 5  // 默认保留备份数量
	BackupTimestampFormat  = "20060102_150405"
)

// 数据库连接池配置
const (
	SQLiteMaxOpenConns    = 1 // SQLite 推荐单连接
	SQLiteMaxIdleConns    = 1
	SQLiteConnMaxLifetime = 0 // 不限制连接生命周期
	
	SystemDBMaxOpenConns = 10 // 系统数据库可以多连接
	SystemDBMaxIdleConns = 5
)
