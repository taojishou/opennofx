import { Card, Input, Switch } from '../ui';
import { SystemConfig } from '../../types/config';

interface CoinPoolConfigProps {
  config: SystemConfig;
  onChange: (updates: Partial<SystemConfig>) => void;
}

export function CoinPoolConfig({ config, onChange }: CoinPoolConfigProps) {
  return (
    <Card title="🪙 币种池配置">
      <div className="space-y-4">
        <Switch
          checked={config.use_default_coins}
          onChange={(e) => onChange({ use_default_coins: e.target.checked })}
          label="使用默认币种列表"
        />
        
        <Input
          label="默认币种 (逗号分隔)"
          value={config.default_coins.join(', ')}
          onChange={(e) => onChange({
            default_coins: e.target.value.split(',').map(s => s.trim())
          })}
          fullWidth
        />
        
        <Input
          label="AI500币种池API"
          value={config.coin_pool_api_url}
          onChange={(e) => onChange({ coin_pool_api_url: e.target.value })}
          placeholder="留空则不使用"
          fullWidth
        />
        
        <Input
          label="OI Top API"
          value={config.oi_top_api_url}
          onChange={(e) => onChange({ oi_top_api_url: e.target.value })}
          placeholder="留空则不使用"
          fullWidth
        />
      </div>
    </Card>
  );
}
