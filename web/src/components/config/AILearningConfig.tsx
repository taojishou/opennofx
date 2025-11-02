import { Card, Switch, Input } from '../ui';
import { SystemConfig } from '../../types/config';
import { theme } from '../../styles/theme';

interface AILearningConfigProps {
  config: SystemConfig;
  onChange: (updates: Partial<SystemConfig>) => void;
}

export function AILearningConfig({ config, onChange }: AILearningConfigProps) {
  return (
    <Card title="🧠 AI自动学习 & 自主模式">
      <div className="space-y-4">
        <div className="p-4 rounded-lg" style={{ background: theme.colors.background.tertiary }}>
          <Switch
            checked={config.ai_autonomy_mode || false}
            onChange={(e) => onChange({ ai_autonomy_mode: e.target.checked })}
            label="🤖 AI完全自主模式"
            description="开启后AI完全自主决策仓位、杠杆、止损止盈，不受系统限制（风险更高但更灵活）"
          />
          {config.ai_autonomy_mode && (
            <div className="mt-3 p-3 rounded-lg" style={{ background: theme.colors.error.light + '20', border: `1px solid ${theme.colors.error.light}` }}>
              <div className="text-sm font-medium" style={{ color: theme.colors.error.main }}>
                ⚠️ 风险提示
              </div>
              <div className="text-xs mt-1" style={{ color: theme.colors.text.secondary }}>
                自主模式下AI可以使用任意杠杆倍数和仓位大小，请确保你理解其中的风险。建议先用小资金测试。
              </div>
            </div>
          )}
        </div>
        
        <div className="p-4 rounded-lg" style={{ background: theme.colors.background.tertiary }}>
          <Switch
            checked={config.enable_ai_learning || false}
            onChange={(e) => onChange({ enable_ai_learning: e.target.checked })}
            label="启用AI自动学习"
            description="AI会每隔N个周期自动分析历史交易，总结成功/失败模式，避免重复错误"
          />
        </div>
        
        {config.enable_ai_learning && (
          <div className="p-4 rounded-lg" style={{ background: theme.colors.background.tertiary }}>
            <div className="flex items-center gap-3">
              <label className="text-sm font-medium" style={{ color: theme.colors.text.primary }}>
                学习间隔:
              </label>
              <Input
                type="number"
                min="5"
                max="50"
                value={config.ai_learn_interval === undefined || config.ai_learn_interval === 0 ? 10 : config.ai_learn_interval}
                onChange={(e) => {
                  const val = parseInt(e.target.value);
                  onChange({ ai_learn_interval: isNaN(val) || val < 5 ? 10 : val });
                }}
                style={{ width: '6rem', textAlign: 'center' }}
              />
              <span className="text-sm" style={{ color: theme.colors.text.secondary }}>个周期</span>
              <span
                className="text-xs px-2 py-1 rounded"
                style={{ background: theme.colors.success.light, color: theme.colors.success.main }}
              >
                推荐: 10
              </span>
            </div>
            <div className="mt-3 text-xs" style={{ color: theme.colors.text.secondary }}>
              💡 提示：间隔太短可能增加成本，间隔太长学习效果不明显
            </div>
          </div>
        )}
      </div>
    </Card>
  );
}
