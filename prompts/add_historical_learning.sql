-- 添加历史学习配置（用户可在前端开关）
INSERT INTO prompt_configs (section_name, title, content, enabled, display_order) VALUES
('historical_learning', '📚 历史交易学习', 
'## 📚 你的历史交易学习

这个section的内容将动态生成，包括：
- 最近5笔失败案例（避免重复错误）
- 最近3笔成功案例（值得复制的策略）
- 统计洞察（胜率、持仓时长、最佳/最差币种）

启用此功能后，AI每次决策都能看到自己的历史表现，从过去的成功和失败中学习。

配置参数：
- 失败案例数量：5笔
- 成功案例数量：3笔
- 分析范围：最近20笔交易
',
1, -- 默认启用
8  -- 显示顺序（在其他sections之后）
);
