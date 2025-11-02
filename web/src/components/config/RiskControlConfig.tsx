import { Card, Input, Badge } from '../ui';
import { SystemConfig } from '../../types/config';

interface RiskControlConfigProps {
  config: SystemConfig;
  onChange: (updates: Partial<SystemConfig>) => void;
}

export function RiskControlConfig({ config, onChange }: RiskControlConfigProps) {
  return (
    <Card>
      <div className="flex items-center gap-2 mb-4">
        <h3 className="text-xl font-bold" style={{ color: '#EAECEF' }}>🛡️ 风险控制</h3>
        <Badge variant="success">重要</Badge>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <Input
          type="number"
          label="最大持仓数"
          value={config.max_positions}
          onChange={(e) => onChange({ max_positions: parseInt(e.target.value) })}
          fullWidth
        />
        <Input
          type="number"
          label="最大日亏损(%)"
          step="0.1"
          value={config.max_daily_loss}
          onChange={(e) => onChange({ max_daily_loss: parseFloat(e.target.value) })}
          fullWidth
        />
        <Input
          type="number"
          label="最大回撤(%)"
          step="0.1"
          value={config.max_drawdown}
          onChange={(e) => onChange({ max_drawdown: parseFloat(e.target.value) })}
          fullWidth
        />
        <Input
          type="number"
          label="暂停交易时长(分钟)"
          value={config.stop_trading_minutes}
          onChange={(e) => onChange({ stop_trading_minutes: parseInt(e.target.value) })}
          fullWidth
        />
      </div>
    </Card>
  );
}
