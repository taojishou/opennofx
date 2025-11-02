import { Card, Input } from '../ui';
import { SystemConfig } from '../../types/config';

interface LeverageConfigProps {
  config: SystemConfig;
  onChange: (updates: Partial<SystemConfig>) => void;
}

export function LeverageConfig({ config, onChange }: LeverageConfigProps) {
  return (
    <Card title="⚖️ 杠杆配置">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Input
          type="number"
          label="BTC/ETH杠杆倍数"
          hint="建议: 3-10倍"
          min="1"
          max="50"
          value={config.leverage.btc_eth_leverage}
          onChange={(e) => onChange({
            leverage: { ...config.leverage, btc_eth_leverage: parseInt(e.target.value) }
          })}
          fullWidth
        />
        <Input
          type="number"
          label="山寨币杠杆倍数"
          hint="建议: 2-5倍"
          min="1"
          max="20"
          value={config.leverage.altcoin_leverage}
          onChange={(e) => onChange({
            leverage: { ...config.leverage, altcoin_leverage: parseInt(e.target.value) }
          })}
          fullWidth
        />
      </div>
    </Card>
  );
}
