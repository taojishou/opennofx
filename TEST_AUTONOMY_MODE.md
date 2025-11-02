# AI自主模式配置修复总结

## 问题
用户反馈后台无法配置AI自主模式，设置后不起作用。

## 根本原因
1. **API层缺少字段传递**：`handleUpdateGlobalConfig`没有处理`ai_autonomy_mode`字段
2. **前端发送缺少字段**：`saveGlobalConfig`没有在请求体中包含`ai_autonomy_mode`
3. **配置加载缺少字段**：`LoadConfigFromDB`没有加载`ai_autonomy_mode`到全局Config
4. **Config结构缺少字段**：全局`Config`结构没有定义`AIAutonomyMode`字段

## 修复内容

### 1. 后端API层 (api/config_handlers.go)
```go
// handleUpdateGlobalConfig - 添加AIAutonomyMode处理
if req.AIAutonomyMode != nil {
    trader.AIAutonomyMode = *req.AIAutonomyMode
}
```

### 2. 前端发送层 (web/src/hooks/useConfigManager.ts)
```typescript
// saveGlobalConfig - 添加字段到请求体
ai_autonomy_mode: config.ai_autonomy_mode,
```

### 3. 配置加载层 (database/loader.go)
```go
// LoadConfigFromDB - 添加全局配置加载
cfg.AIAutonomyMode = firstTrader.AIAutonomyMode

// 同时也添加到每个TraderConfig
AIAutonomyMode: dbTrader.AIAutonomyMode,
```

### 4. Config结构 (config/config.go)
```go
// Config结构添加全局字段
AIAutonomyMode bool `json:"ai_autonomy_mode"` // AI自主模式（全局开关）
```

### 5. 前端UI层 (web/src/components/config/AILearningConfig.tsx)
- 修复theme颜色：`danger` → `error`（因为theme中没有danger颜色）

## 测试验证

### 数据库验证
```bash
# 查看当前配置
sqlite3 data/system.db "SELECT trader_id, ai_autonomy_mode FROM trader_configs"
# 输出: my_trader|1

# 手动设置测试
sqlite3 data/system.db "UPDATE trader_configs SET ai_autonomy_mode = 1"
```

### API测试
```bash
# 测试配置更新
curl -X POST http://localhost:8080/api/config/global/update \
  -H "Content-Type: application/json" \
  -d '{"ai_autonomy_mode": true}'

# 测试配置读取
curl http://localhost:8080/api/config
```

## 配置流程

### Web界面配置流程
1. 用户在Web界面打开"系统配置" → "AI自动学习 & 自主模式"
2. 切换"🤖 AI完全自主模式"开关
3. 点击"💾 保存全局配置"按钮
4. 前端调用`saveGlobalConfig()` → 发送POST请求到`/api/config/global/update`
5. 后端`handleUpdateGlobalConfig`处理请求，更新数据库
6. 系统重启或热重载后生效

### 配置生效路径
```
数据库 trader_configs.ai_autonomy_mode 
  ↓
database/loader.go LoadConfigFromDB() 
  ↓
config.Config.AIAutonomyMode 
  ↓
manager/trader_manager.go AddTrader() 
  ↓
trader/auto_trader.go AutoTraderConfig.AIAutonomyMode 
  ↓
trader/auto_trader.go buildTradingContext() 
  ↓
decision.Context.AIAutonomyMode 
  ↓
decision/engine.go validateDecision()
  - 如果AIAutonomyMode=true → validateDecisionAutonomy()（宽松验证）
  - 如果AIAutonomyMode=false → 正常验证（严格限制）
```

## 功能说明

### 限制模式（ai_autonomy_mode=false，默认）
- 仓位大小：根据账户净值和SmartRiskManager动态限制
- 杠杆倍数：1-20倍
- 风险回报比：最低3:1（山寨币）或1.8:1（BTC/ETH）
- 止损止盈：必须设置，需符合ATR要求
- 智能风控：根据亏损自动缩减仓位

### 完全自主模式（ai_autonomy_mode=true）
- 仓位大小：AI完全自主决定
- 杠杆倍数：1-125倍（仅受交易所限制）
- 风险回报比：AI自主评估
- 止损止盈：AI可选择设置或不设置
- 智能风控：不限制

## 前端构建
```bash
cd web
npm run build
# 成功输出: ✓ built in 1.55s
```

## 建议测试步骤
1. 启动系统
2. 访问Web界面，进入系统配置
3. 开启AI自主模式开关
4. 保存配置
5. 重启trader
6. 观察日志中是否有`[AI自主模式]`标签
7. 检查AI决策是否不再受仓位限制

## 注意事项
⚠️ AI自主模式风险更高，建议：
- 先用小资金（<100 USDT）测试
- 密切监控日志和账户情况
- 对比两种模式的表现（夏普比率、最大回撤）
