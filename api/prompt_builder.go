package api

import (
	"fmt"
	"log"
	"nofx/database"
	"strings"
)

// BuildSystemPromptFromDB 从数据库构建system prompt
func BuildSystemPromptFromDB(db *database.DB, accountEquity float64, btcEthLeverage, altcoinLeverage int) string {
	configs, err := db.GetEnabledPromptConfigs()
	if err != nil {
		log.Printf("⚠️ 获取prompt配置失败: %v", err)
		return "错误：无法加载prompt配置"
	}

	var result strings.Builder
	result.WriteString("你是专业的加密货币交易AI，在币安合约市场进行自主交易。\n\n")

	for _, cfg := range configs {
		if !cfg.Enabled {
			continue
		}

		result.WriteString("# ")
		result.WriteString(cfg.Title)
		result.WriteString("\n\n")

		// 替换变量
		content := cfg.Content
		content = replacePromptVariables(content, accountEquity, btcEthLeverage, altcoinLeverage)

		result.WriteString(content)
		result.WriteString("\n\n")
	}

	// 添加提醒
	result.WriteString("---\n\n")
	result.WriteString("**记住**: \n")
	result.WriteString("- 🎯 目标是夏普比率，不是交易频率\n")
	result.WriteString("- ⚖️ 做多 = 做空，完全平等的工具\n")
	result.WriteString("- ✅ 宁可错过，不做低质量交易\n")
	result.WriteString("- 🛡️ 风险回报比1:3是底线\n")
	result.WriteString("- 📊 多空平衡是成功的关键\n")

	return result.String()
}

// replacePromptVariables 替换prompt中的变量
func replacePromptVariables(content string, accountEquity float64, btcEthLeverage, altcoinLeverage int) string {
	altMinSize := accountEquity * 0.8
	altMaxSize := accountEquity * 1.5
	btcMinSize := accountEquity * 5
	btcMaxSize := accountEquity * 10

	result := content
	result = strings.ReplaceAll(result, "{{accountEquity}}", fmt.Sprintf("%.2f", accountEquity))
	result = strings.ReplaceAll(result, "{{btcEthLeverage}}", fmt.Sprintf("%d", btcEthLeverage))
	result = strings.ReplaceAll(result, "{{altcoinLeverage}}", fmt.Sprintf("%d", altcoinLeverage))
	result = strings.ReplaceAll(result, "{{altMinSize}}", fmt.Sprintf("%.0f", altMinSize))
	result = strings.ReplaceAll(result, "{{altMaxSize}}", fmt.Sprintf("%.0f", altMaxSize))
	result = strings.ReplaceAll(result, "{{btcMinSize}}", fmt.Sprintf("%.0f", btcMinSize))
	result = strings.ReplaceAll(result, "{{btcMaxSize}}", fmt.Sprintf("%.0f", btcMaxSize))

	return result
}
