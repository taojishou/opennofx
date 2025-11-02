package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"nofx/database/models"
	"time"
)

// PromptConfig 简化的Prompt配置结构（用于向外暴露）
type PromptConfig struct {
	ID           int64
	SectionName  string
	Title        string
	Content      string
	PromptType   string
	Enabled      bool
	DisplayOrder int
	UpdatedAt    time.Time
}

// ConfigRepository Prompt配置数据访问层
type ConfigRepository struct {
	db *sql.DB
}

// NewConfigRepository 创建配置仓储
func NewConfigRepository(db *sql.DB) *ConfigRepository {
	repo := &ConfigRepository{db: db}
	
	// 初始化默认配置
	if err := repo.initDefaults(); err != nil {
		log.Printf("⚠️ 初始化默认Prompt配置失败: %v", err)
	}
	
	return repo
}

// GetAll 获取所有prompt配置
func (r *ConfigRepository) GetAll() ([]*models.PromptConfig, error) {
	query := `
		SELECT id, section_name, title, content, prompt_type, enabled, display_order, updated_at
		FROM prompt_configs
		ORDER BY display_order ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*models.PromptConfig
	for rows.Next() {
		cfg := &models.PromptConfig{}
		err := rows.Scan(&cfg.ID, &cfg.SectionName, &cfg.Title, &cfg.Content,
			&cfg.PromptType, &cfg.Enabled, &cfg.DisplayOrder, &cfg.UpdatedAt)
		if err != nil {
			continue
		}
		configs = append(configs, cfg)
	}

	return configs, nil
}

// GetEnabled 获取启用的prompt配置
func (r *ConfigRepository) GetEnabled() ([]*models.PromptConfig, error) {
	query := `
		SELECT id, section_name, title, content, prompt_type, enabled, display_order, updated_at
		FROM prompt_configs
		WHERE enabled = 1
		ORDER BY display_order ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*models.PromptConfig
	for rows.Next() {
		cfg := &models.PromptConfig{}
		err := rows.Scan(&cfg.ID, &cfg.SectionName, &cfg.Title, &cfg.Content,
			&cfg.PromptType, &cfg.Enabled, &cfg.DisplayOrder, &cfg.UpdatedAt)
		if err != nil {
			continue
		}
		configs = append(configs, cfg)
	}

	return configs, nil
}

// GetByType 获取指定类型的启用配置
func (r *ConfigRepository) GetByType(promptType string) ([]*PromptConfig, error) {
	query := `
		SELECT id, section_name, title, content, prompt_type, enabled, display_order, updated_at
		FROM prompt_configs
		WHERE enabled = 1 AND prompt_type = ?
		ORDER BY display_order ASC
	`

	rows, err := r.db.Query(query, promptType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*PromptConfig
	for rows.Next() {
		cfg := &PromptConfig{}
		err := rows.Scan(&cfg.ID, &cfg.SectionName, &cfg.Title, &cfg.Content,
			&cfg.PromptType, &cfg.Enabled, &cfg.DisplayOrder, &cfg.UpdatedAt)
		if err != nil {
			continue
		}
		configs = append(configs, cfg)
	}

	return configs, nil
}

// Update 更新prompt配置
func (r *ConfigRepository) Update(cfg *models.PromptConfig) error {
	query := `
		UPDATE prompt_configs 
		SET title = ?, content = ?, prompt_type = ?, enabled = ?, display_order = ?, updated_at = CURRENT_TIMESTAMP
		WHERE section_name = ?
	`

	_, err := r.db.Exec(query, cfg.Title, cfg.Content, cfg.PromptType, cfg.Enabled, cfg.DisplayOrder, cfg.SectionName)
	return err
}

// Insert 添加新的prompt配置
func (r *ConfigRepository) Insert(cfg *models.PromptConfig) error {
	query := `INSERT INTO prompt_configs (section_name, title, content, enabled, display_order, prompt_type) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, cfg.SectionName, cfg.Title, cfg.Content, cfg.Enabled, cfg.DisplayOrder, cfg.PromptType)
	return err
}

// Delete 删除prompt配置
func (r *ConfigRepository) Delete(sectionName string) (int64, error) {
	query := `DELETE FROM prompt_configs WHERE section_name = ?`
	result, err := r.db.Exec(query, sectionName)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// initDefaults 初始化默认prompt配置
func (r *ConfigRepository) initDefaults() error {
	// 检查是否已经初始化
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM prompt_configs").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // 已经初始化过了
	}

	log.Println("🔧 初始化默认Prompt配置...")

	defaults := []models.PromptConfig{
		{
			SectionName:  "core_mission",
			Title:        "🎯 核心目标",
			PromptType:   "system",
			DisplayOrder: 1,
			Enabled:      true,
			Content: `**最大化夏普比率（Sharpe Ratio）**

夏普比率 = 平均收益 / 收益波动率

**这意味着**：
- ✅ 高质量交易（高胜率、大盈亏比）→ 提升夏普
- ✅ 稳定收益、控制回撤 → 提升夏普
- ✅ 耐心持仓、让利润奔跑 → 提升夏普
- ❌ 频繁交易、小盈小亏 → 增加波动，严重降低夏普
- ❌ 过度交易、手续费损耗 → 直接亏损
- ❌ 过早平仓、频繁进出 → 错失大行情

**关键认知**: 系统每3分钟扫描一次，但不意味着每次都要交易！
大多数时候应该是 wait 或 hold，只在极佳机会时才开仓。`,
		},
		{
			SectionName:  "hard_constraints",
			Title:        "⚖️ 硬约束（风险控制）",
			PromptType:   "system",
			DisplayOrder: 2,
			Enabled:      true,
			Content: `1. **风险回报比**: 必须 ≥ 1:3（冒1%风险，赚3%+收益）

2. **止损设置** (关键！避免被噪音扫掉):
   - 必须参考ATR14（平均真实波幅）设置止损距离
   - **最小止损距离**: 1.5倍ATR14（给予足够的波动空间）
   - 做多止损 = 当前价 - (1.5~2.5)×ATR14
   - 做空止损 = 当前价 + (1.5~2.5)×ATR14
   - 关键支撑/阻力位可作为止损参考，但不能小于1.5倍ATR
   - 示例：ETH价格3888，ATR14=50 → 止损至少距离75点(1.5×50)，即做多止损≤3813

3. **止盈设置**:
   - 确保风险回报比≥3:1
   - 如止损距离2%，则止盈至少6%
   - 可参考前高/前低、斐波那契扩展位设置目标

4. **最多持仓**: 由配置决定（用户提示中会显示持仓状态，注意查看上限值）

5. **单币仓位**: 
   - 山寨币: {{altMinSize}}-{{altMaxSize}} USDT ({{altcoinLeverage}}x杠杆)
   - BTC/ETH: {{btcMinSize}}-{{btcMaxSize}} USDT ({{btcEthLeverage}}x杠杆)

6. **保证金**: 总使用率 ≤ 90%`,
		},
		{
			SectionName:  "long_short_balance",
			Title:        "⚖️ 做多做空平衡",
			PromptType:   "system",
			DisplayOrder: 3,
			Enabled:      true,
			Content: `**核心原则**: 做多和做空是完全平等的赚钱工具！

**判断标准**:
- 📈 上涨趋势 → 做多 (价格>EMA20>EMA50, MACD>0, RSI>50, 成交量放大)
- 📉 下跌趋势 → 做空 (价格<EMA20<EMA50, MACD<0, RSI<50, 成交量放大)
- ⏸️ 震荡市场 → 观望 (指标相互矛盾，方向不明确)

**重要等式**:
- 上涨5%做多的利润 = 下跌5%做空的利润
- 做多的风险 = 做空的风险
- 成功率不取决于方向，取决于趋势判断准确性

**严禁偏见**:
- ❌ 单边做多（错失下跌机会）
- ❌ 单边做空（错失上涨机会）
- ✅ 客观分析市场，跟随趋势`,
		},
		// 更多默认配置...（为节省篇幅省略，实际代码中应包含完整配置）
	}

	for _, cfg := range defaults {
		_, err := r.db.Exec(`
			INSERT INTO prompt_configs (section_name, title, content, enabled, display_order, prompt_type)
			VALUES (?, ?, ?, ?, ?, ?)
		`, cfg.SectionName, cfg.Title, cfg.Content, cfg.Enabled, cfg.DisplayOrder, cfg.PromptType)

		if err != nil {
			return fmt.Errorf("插入默认prompt配置失败 [%s]: %w", cfg.SectionName, err)
		}
	}

	log.Println("✓ 默认Prompt配置初始化完成")
	return nil
}
