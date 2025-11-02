import { Card, Select, Input, Button } from '../ui';
import { SystemConfig } from '../../types/config';
import { theme } from '../../styles/theme';

interface KlineDataConfigProps {
  config: SystemConfig;
  onChange: (updates: Partial<SystemConfig>) => void;
}

export function KlineDataConfig({ config, onChange }: KlineDataConfigProps) {
  const hasKlines = config.market_data && config.market_data.klines && config.market_data.klines.length > 0;

  const initializeKlines = () => {
    onChange({
      market_data: {
        klines: [
          { interval: '3m', limit: 5, show_table: true },
          { interval: '15m', limit: 10, show_table: false },
          { interval: '4h', limit: 60, show_table: false }
        ]
      }
    });
  };

  const addKline = () => {
    const newKline = { interval: '15m', limit: 10, show_table: false };
    onChange({
      market_data: {
        klines: [...(config.market_data?.klines || []), newKline]
      }
    });
  };

  const updateKline = (index: number, updates: any) => {
    const newKlines = [...(config.market_data?.klines || [])];
    newKlines[index] = { ...newKlines[index], ...updates };
    onChange({ market_data: { klines: newKlines } });
  };

  const removeKline = (index: number) => {
    const newKlines = config.market_data!.klines.filter((_, i) => i !== index);
    onChange({ market_data: { klines: newKlines } });
  };

  return (
    <Card title="📊 K线数据配置">
      {!hasKlines ? (
        <div className="p-4 mb-4 rounded-lg" style={{ background: theme.colors.background.tertiary, border: `1px solid ${theme.colors.border.secondary}` }}>
          <p className="text-sm mb-3" style={{ color: theme.colors.text.secondary }}>
            未配置K线数据，将使用默认设置（3分钟20根 + 4小时60根）
          </p>
          <Button variant="success" onClick={initializeKlines}>
            初始化推荐配置
          </Button>
        </div>
      ) : (
        <div className="space-y-3">
          {config.market_data!.klines.map((kline, index) => (
            <div key={index} className="p-4 rounded-lg" style={{ background: theme.colors.background.tertiary, border: `1px solid ${theme.colors.border.secondary}` }}>
              <div className="flex items-center justify-between mb-4">
                <h4 className="font-semibold" style={{ color: theme.colors.text.primary }}>
                  K线 #{index + 1}
                </h4>
                {config.market_data!.klines.length > 1 && (
                  <Button
                    variant="danger"
                    size="sm"
                    onClick={() => removeKline(index)}
                  >
                    删除
                  </Button>
                )}
              </div>

              <div className="grid grid-cols-3 gap-4">
                <Select
                  label="时间周期"
                  value={kline.interval}
                  onChange={(e) => updateKline(index, { interval: e.target.value })}
                  fullWidth
                >
                  <option value="1m">1分钟</option>
                  <option value="3m">3分钟</option>
                  <option value="5m">5分钟</option>
                  <option value="15m">15分钟</option>
                  <option value="30m">30分钟</option>
                  <option value="1h">1小时</option>
                  <option value="2h">2小时</option>
                  <option value="4h">4小时</option>
                  <option value="6h">6小时</option>
                  <option value="12h">12小时</option>
                  <option value="1d">1天</option>
                </Select>

                <Input
                  type="number"
                  label="K线数量"
                  min="5"
                  max="200"
                  value={kline.limit || ''}
                  onChange={(e) => {
                    const val = e.target.value;
                    if (val === '') {
                      updateKline(index, { limit: null });
                    } else {
                      const num = parseInt(val);
                      if (!isNaN(num)) {
                        updateKline(index, { limit: num });
                      }
                    }
                  }}
                  onBlur={(e) => {
                    const val = e.target.value;
                    if (val === '' || parseInt(val) < 5) {
                      updateKline(index, { limit: 20 });
                    } else if (parseInt(val) > 200) {
                      updateKline(index, { limit: 200 });
                    }
                  }}
                  fullWidth
                />

                <div>
                  <label className="block text-sm mb-2" style={{ color: theme.colors.text.secondary }}>
                    显示表格
                  </label>
                  <label className="flex items-center cursor-pointer pt-2">
                    <input
                      type="checkbox"
                      checked={kline.show_table}
                      onChange={(e) => updateKline(index, { show_table: e.target.checked })}
                      className="w-5 h-5 mr-2"
                      style={{ accentColor: theme.colors.success.main }}
                    />
                    <span style={{ color: theme.colors.text.primary }}>显示K线表格</span>
                  </label>
                </div>
              </div>
            </div>
          ))}

          {config.market_data!.klines.length < 5 && (
            <Button variant="success" fullWidth onClick={addKline}>
              + 添加K线配置
            </Button>
          )}
        </div>
      )}

      <div className="mt-4 p-3 rounded-lg" style={{ background: theme.colors.warning.light, border: `1px solid ${theme.colors.warning.border}` }}>
        <p className="text-sm leading-relaxed" style={{ color: theme.colors.warning.main, margin: 0 }}>
          💡 <strong>建议</strong>: K线数据过多会增加prompt大小，可能导致AI过度交易。<br/>
          推荐：3分钟5根（参考）+ 15分钟10根（决策）+ 4小时60根（趋势）
        </p>
      </div>
    </Card>
  );
}
