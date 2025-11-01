package prompts

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// PromptSection 单个prompt部分
type PromptSection struct {
	Title   string `yaml:"title"`
	Content string `yaml:"content"`
	Enabled *bool  `yaml:"enabled,omitempty"` // nil表示默认启用
}

// PromptConfig prompt配置
type PromptConfig struct {
	Version         string                    `yaml:"version"`
	EnabledSections []string                  `yaml:"enabled_sections"`
	Sections        map[string]PromptSection  `yaml:"sections"`
	Reminders       []string                  `yaml:"reminders"`
}

// LoadPromptConfig 加载prompt配置文件
func LoadPromptConfig(filepath string) (*PromptConfig, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("读取prompt配置文件失败: %w", err)
	}

	var config PromptConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析prompt配置文件失败: %w", err)
	}

	return &config, nil
}

// BuildSystemPrompt 根据配置构建system prompt
func (c *PromptConfig) BuildSystemPrompt(accountEquity float64, btcEthLeverage, altcoinLeverage int) string {
	var sb strings.Builder

	// 按顺序构建各个section
	for _, sectionName := range c.EnabledSections {
		section, exists := c.Sections[sectionName]
		if !exists {
			continue
		}

		// 检查是否启用（nil或true表示启用）
		if section.Enabled != nil && !*section.Enabled {
			continue
		}

		// 添加标题
		sb.WriteString("# ")
		sb.WriteString(section.Title)
		sb.WriteString("\n\n")

		// 添加内容（替换变量）
		content := section.Content
		content = strings.ReplaceAll(content, "{{accountEquity}}", fmt.Sprintf("%.2f", accountEquity))
		content = strings.ReplaceAll(content, "{{btcEthLeverage}}", fmt.Sprintf("%d", btcEthLeverage))
		content = strings.ReplaceAll(content, "{{altcoinLeverage}}", fmt.Sprintf("%d", altcoinLeverage))
		
		// 动态替换仓位范围
		altMinSize := accountEquity * 0.8
		altMaxSize := accountEquity * 1.5
		btcMinSize := accountEquity * 5
		btcMaxSize := accountEquity * 10
		
		content = strings.ReplaceAll(content, "账户净值的0.8-1.5倍", 
			fmt.Sprintf("%.0f-%.0f USDT", altMinSize, altMaxSize))
		content = strings.ReplaceAll(content, "账户净值的5-10倍", 
			fmt.Sprintf("%.0f-%.0f USDT", btcMinSize, btcMaxSize))

		sb.WriteString(content)
		sb.WriteString("\n\n")
	}

	// 添加提醒事项
	if len(c.Reminders) > 0 {
		sb.WriteString("---\n\n")
		sb.WriteString("**记住**: \n")
		for _, reminder := range c.Reminders {
			sb.WriteString("- ")
			sb.WriteString(reminder)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// GetDefaultPromptPath 获取默认prompt配置路径
func GetDefaultPromptPath() string {
	return "prompts/system_prompt.yaml"
}

// LoadDefaultPrompt 加载默认prompt配置
func LoadDefaultPrompt() (*PromptConfig, error) {
	return LoadPromptConfig(GetDefaultPromptPath())
}
