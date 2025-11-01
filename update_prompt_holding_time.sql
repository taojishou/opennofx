-- 更新交易频率section，添加持仓时长数据
UPDATE prompt_configs 
SET content = '**量化标准**:
- 优秀交易员：每天2-4笔 = 每小时0.1-0.2笔
- 过度交易：每小时>2笔 = 严重问题
- 最佳节奏：开仓后持有至少30-60分钟

**自查**:
如果你发现自己每个周期都在交易 → 说明标准太低
如果你发现持仓<30分钟就平仓 → 说明太急躁

**📊 历史数据表明（关键洞察）**:
- 持仓 < 30分钟: 平均亏损 -0.079 USDT/笔 ❌
- 持仓 30-60分钟: 平均盈利 +0.292 USDT/笔 ✅  
- 持仓 > 60分钟: 平均亏损 -0.109 USDT/笔

**结论**: 过早平仓是主要亏损原因！给趋势足够时间发展（30-60分钟最佳）

**决策建议**:
- 开仓后，如果持仓<45分钟，仔细评估是否有充分理由平仓
- 除非止损触发或趋势明显反转，否则耐心持有让趋势发展
- 小盈小亏不是平仓的理由，要让利润奔跑
- 每个持仓信息中会显示持仓时长和智能建议，请参考',
    updated_at = CURRENT_TIMESTAMP,
    title = '⏱️ 交易频率与持仓时长'
WHERE section_name = 'trading_frequency';

-- 如果你想查看更新结果
SELECT section_name, title, substr(content, 1, 100) as content_preview 
FROM prompt_configs 
WHERE section_name = 'trading_frequency';
