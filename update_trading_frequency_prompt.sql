-- 更新交易频率section，添加持仓时长的历史数据分析
-- 在每个trader的数据库中执行此SQL

UPDATE prompt_configs 
SET 
    title = '⏱️ 交易频率与持仓时长',
    content = '**量化标准**:
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
- 开仓后，除非有强烈信号，否则耐心持有让趋势发展
- 持仓信息中会显示持仓时长，请参考历史数据做决策
- 小盈小亏不是平仓的理由，要让利润奔跑
- 如果持仓<45分钟且只有小幅波动，选择hold而不是过早平仓',
    updated_at = CURRENT_TIMESTAMP
WHERE section_name = 'trading_frequency';

-- 查看更新结果
SELECT section_name, title, substr(content, 1, 150) || '...' as content_preview 
FROM prompt_configs 
WHERE section_name = 'trading_frequency';
