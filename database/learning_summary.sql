-- AI学习总结表：存储AI自己生成的历史交易总结
CREATE TABLE IF NOT EXISTS ai_learning_summaries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    trader_id TEXT NOT NULL,
    summary_content TEXT NOT NULL,      -- AI生成的总结内容（markdown格式）
    trades_count INTEGER NOT NULL,      -- 分析的交易数量
    date_range_start TEXT,              -- 分析的时间范围起始
    date_range_end TEXT,                -- 分析的时间范围结束
    win_rate REAL,                      -- 胜率
    avg_pnl REAL,                       -- 平均盈亏
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT 1         -- 是否激活（只有最新的一条active=1）
);

CREATE INDEX IF NOT EXISTS idx_ai_learning_trader ON ai_learning_summaries(trader_id);
CREATE INDEX IF NOT EXISTS idx_ai_learning_active ON ai_learning_summaries(trader_id, is_active);
