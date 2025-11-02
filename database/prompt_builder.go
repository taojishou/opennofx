package database

import (
	"fmt"
	"log"
	"nofx/database/repositories"
	"strings"
)

// BuildSystemPrompt 从Repository构建system prompt
// 注意：maxPositionValueBTC和maxPositionValueAlt应该是动态风控调整后的实际可用限制
// aiAutonomyMode: true=自主模式（移除限制性规则），false=限制模式（包含所有规则）
func BuildSystemPrompt(repo *repositories.ConfigRepository, accountEquity float64, btcEthLeverage, altcoinLeverage int, maxPositionValueBTC, maxPositionValueAlt float64, aiAutonomyMode bool) string {
	configs, err := repo.GetByType("system")
	if err != nil {
		return "错误：无法加载system prompt配置"
	}

	// 使用传入的实际可用仓位限制（已考虑动态风控调整）

	var result strings.Builder
	
	// 自主模式提示
	if aiAutonomyMode {
		result.WriteString("你是专业的加密货币交易AI，在币安合约市场进行**完全自主交易**。\n\n")
		result.WriteString("🚀 **AI自主模式已启用**：你拥有完全的决策自由，可以根据市场情况自主决定所有参数。\n\n")
	} else {
		result.WriteString("你是专业的加密货币交易AI，在币安合约市场进行自主交易。\n\n")
	}

	// 自主模式下需要跳过的限制性规则
	restrictiveSections := map[string]bool{
		"hard_constraints":    true, // 硬约束（风险回报比、止损距离等限制）
		"opening_standards":   true, // 开仓标准（严格限制）
	}

	for _, cfg := range configs {
		// 自主模式下跳过限制性规则
		if aiAutonomyMode && restrictiveSections[cfg.SectionName] {
			log.Printf("🚀 [AI自主模式] 跳过限制性规则: %s", cfg.Title)
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

	// 添加输出格式要求（关键！）
	result.WriteString("---\n\n")
	result.WriteString("# 📤 输出格式\n\n")
	result.WriteString("**第一步: 思维链（纯文本）**\n")
	result.WriteString("简洁分析你的思考过程\n\n")
	result.WriteString("**第二步: JSON决策数组**\n\n")
	result.WriteString("```json\n[\n")
	result.WriteString(fmt.Sprintf("  {\"symbol\": \"BTCUSDT\", \"action\": \"open_short\", \"leverage\": %d, \"position_size_usd\": %.0f, \"stop_loss\": 97000, \"take_profit\": 91000, \"confidence\": 85, \"risk_usd\": 300, \"reasoning\": \"下跌趋势+MACD死叉\"},\n", btcEthLeverage, accountEquity*3))
	result.WriteString("  {\"symbol\": \"ETHUSDT\", \"action\": \"close_long\", \"reasoning\": \"止盈离场\"}\n")
	result.WriteString("]\n```\n\n")
	result.WriteString("**字段说明**:\n")
	result.WriteString("- `action`: open_long | open_short | close_long | close_short | hold | wait\n")
	result.WriteString("- `confidence`: 0-100（开仓建议≥75）\n")
	result.WriteString("- 开仓时必填: leverage, position_size_usd, stop_loss, take_profit, confidence, risk_usd, reasoning\n\n")
	
	// 添加仓位限制说明
	result.WriteString("**⚠️ 当前可用仓位限制（已动态调整）**:\n")
	result.WriteString(fmt.Sprintf("- BTC/ETH: 仓位价值(position_size_usd × leverage) ≤ %.0f USDT\n", maxPositionValueBTC))
	result.WriteString(fmt.Sprintf("- 其他币种: 仓位价值(position_size_usd × leverage) ≤ %.0f USDT\n", maxPositionValueAlt))
	result.WriteString(fmt.Sprintf("- 示例BTC（杠杆%dx）：position_size_usd不应超过 %.0f USDT\n", btcEthLeverage, maxPositionValueBTC/float64(btcEthLeverage)))
	result.WriteString(fmt.Sprintf("- 示例其他币（杠杆%dx）：position_size_usd不应超过 %.0f USDT\n", altcoinLeverage, maxPositionValueAlt/float64(altcoinLeverage)))
	result.WriteString("- ⚠️ 这是当前实际可用限制，已根据账户表现、保证金使用率等动态调整，请严格遵守！\n\n")
	
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
