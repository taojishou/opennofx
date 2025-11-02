# 数据库架构说明

## 架构概览

项目采用 **双数据库 + Repository模式** 架构：

```
database/
├── models/              # 领域模型（纯数据结构）
├── repositories/        # 数据访问层（Repository模式）
├── migrations/          # 数据库迁移（预留）
├── connection.go        # Trader数据库连接
├── system_connection.go # 系统数据库连接
├── manager.go           # 数据库管理器（统一入口）
├── migrate.go           # 配置迁移工具
├── config.go            # 数据库配置
└── db.go                # 向后兼容适配器
```

## 数据库说明

### 1. 系统数据库（`data/system.db`）

存储全局配置和用户信息：

- **users** - 用户表
- **sessions** - 会话表
- **system_configs** - 系统配置
- **trader_configs** - Trader配置

### 2. Trader数据库（`data/traders/{trader_id}/trading.db`）

每个Trader独立数据库，存储交易数据：

- **decision_records** - 决策记录
- **decision_actions** - 决策动作
- **position_snapshots** - 持仓快照
- **candidate_coins** - 候选币种
- **trade_outcomes** - 交易结果
- **prompt_configs** - Prompt配置
- **position_open_times** - 持仓时间管理
- **trader_states** - Trader状态
- **ai_learning_summaries** - AI学习总结

## 使用方式

### 旧代码（向后兼容）

```go
// 仍然可以使用旧的API
db, err := database.New(traderID)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// 使用旧的方法
recordID, err := db.InsertDecisionRecord(record)
```

### 新代码（推荐）

```go
// 创建Manager
manager, err := database.NewManager()
if err != nil {
    log.Fatal(err)
}
defer manager.Close()

// 使用Repository
decisionRepo, _ := manager.GetDecisionRepo(traderID)
recordID, err := decisionRepo.Insert(record)

// 系统配置
config, _ := manager.SystemConfigRepo.Get("api_server_port")

// Trader配置
traderConfigs, _ := manager.TraderConfigRepo.GetAllEnabled()
```

## 配置迁移

### 从config.json迁移到数据库

```bash
# 运行迁移工具
go run cmd/migrate_config.go config.json

# 备份旧配置
cp config.json config.json.bak
```

### 从数据库加载配置

```go
manager, _ := database.NewManager()
config, err := database.LoadSystemConfig(manager)
// config 现在包含从数据库加载的配置
```

## Repository说明

### DecisionRepository

管理决策记录：

```go
repo := repositories.NewDecisionRepository(db, traderID)

// 插入决策
recordID, err := repo.Insert(record)

// 获取最近记录
records, err := repo.GetLatest(10)

// 插入动作
err = repo.InsertAction(action)

// 获取统计
stats, err := repo.GetStatistics()
```

### TradeRepository

管理交易结果：

```go
repo := repositories.NewTradeRepository(db, traderID)

// 插入交易结果
err := repo.Insert(trade)

// 获取最近交易
trades, err := repo.GetLatest(20)

// 获取统计
stats, err := repo.GetStatistics()

// 删除旧记录
deleted, err := repo.DeleteOld(30) // 删除30天前的
```

### PositionRepository

管理持仓状态：

```go
repo := repositories.NewPositionRepository(db, traderID)

// 保存开仓时间
err := repo.SaveOpenTime("BTCUSDT", "long", timeMs)

// 获取开仓时间
timeMs, exists := repo.GetOpenTime("BTCUSDT", "long")

// 保存Trader状态
err := repo.SaveTraderState(isPaused)
```

### LearningRepository

管理AI学习数据：

```go
repo := repositories.NewLearningRepository(db, traderID)

// 保存学习总结
err := repo.Save(summary)

// 获取激活的总结
summary, err := repo.GetActive()

// 获取所有历史
summaries, err := repo.GetAll(10)
```

### ConfigRepository

管理Prompt配置：

```go
repo := repositories.NewConfigRepository(db)

// 获取所有配置
configs, err := repo.GetAll()

// 获取指定类型
systemConfigs, err := repo.GetByType("system")

// 更新配置
err = repo.Update(cfg)
```

### SystemConfigRepository

管理系统配置：

```go
repo := manager.SystemConfigRepo

// 获取配置
cfg, err := repo.Get("api_server_port")

// 设置配置
err = repo.Set("api_server_port", "8080", "API端口", "api")

// 按类型获取
configs, err := repo.GetByType("market")
```

### TraderConfigRepository

管理Trader配置：

```go
repo := manager.TraderConfigRepo

// 获取所有启用的Trader
configs, err := repo.GetAllEnabled()

// 按TraderID获取
cfg, err := repo.GetByTraderID("my_trader")

// 按用户ID获取
configs, err := repo.GetByUserID(userID)

// 创建新Trader
id, err := repo.Create(traderConfig)

// 更新Trader
err = repo.Update(traderConfig)
```

## 数据库文件位置

```
data/
├── system.db                           # 系统数据库
└── traders/
    └── {trader_id}/
        ├── trading.db                   # Trader数据库
        ├── backups/
        │   └── {timestamp}.db          # 备份文件
        └── logs/
            └── trader.log              # 日志文件
```

## 优势

1. **清晰分层**：模型、Repository、连接管理分离
2. **易于测试**：Repository接口可Mock
3. **向后兼容**：旧代码无需修改
4. **多用户支持**：系统数据库支持用户表
5. **动态配置**：配置存储在数据库，可动态管理
6. **数据隔离**：每个Trader独立数据库

## 迁移注意事项

1. **备份数据**：迁移前先备份 `config.json`
2. **验证配置**：迁移后检查 `data/system.db` 中的数据
3. **测试功能**：确保现有功能正常工作
4. **保留旧文件**：`config.json.bak` 作为备份保留
