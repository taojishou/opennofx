import { Card, Button, Badge } from '../ui';
import { TraderConfig } from '../../types/config';
import { theme } from '../../styles/theme';

interface TraderListProps {
  traders: TraderConfig[];
  onEdit: (trader: TraderConfig) => void;
  onDelete: (traderId: string) => void;
}

export function TraderList({ traders, onEdit, onDelete }: TraderListProps) {
  return (
    <div className="space-y-4">
      {traders.map((trader) => (
        <Card
          key={trader.id}
          variant={trader.enabled ? 'default' : 'elevated'}
          style={{
            border: trader.enabled ? `1px solid ${theme.colors.purple.border}` : undefined,
          }}
        >
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-4">
              <div
                className={`w-3 h-3 rounded-full`}
                style={{ background: trader.enabled ? theme.colors.success.main : theme.colors.text.secondary }}
              />
              <div>
                <div className="flex items-center gap-3 mb-2">
                  <h3 className="text-xl font-bold" style={{ color: theme.colors.text.primary }}>
                    {trader.name}
                  </h3>
                </div>
                <div className="text-sm" style={{ color: theme.colors.text.secondary }}>
                  ID: {trader.id} | {trader.ai_model.toUpperCase()} @ {trader.exchange.toUpperCase()}
                </div>
              </div>
            </div>
            <div className="flex gap-3">
              <Button variant="purple" size="sm" onClick={() => onEdit(trader)}>
                ✏️ 编辑
              </Button>
              <Button variant="danger" size="sm" onClick={() => onDelete(trader.id)}>
                🗑️ 删除
              </Button>
            </div>
          </div>
          <div className="grid grid-cols-3 gap-4 text-sm">
            <div>
              <span style={{ color: theme.colors.text.secondary }}>初始资金: </span>
              <span style={{ color: theme.colors.text.primary }}>${trader.initial_balance}</span>
            </div>
            <div>
              <span style={{ color: theme.colors.text.secondary }}>扫描间隔: </span>
              <span style={{ color: theme.colors.text.primary }}>{trader.scan_interval_minutes}分钟</span>
            </div>
            <div>
              <span style={{ color: theme.colors.text.secondary }}>状态: </span>
              <Badge variant={trader.enabled ? 'success' : 'default'}>
                {trader.enabled ? '✅ 已启用' : '⏸️ 已禁用'}
              </Badge>
            </div>
          </div>
        </Card>
      ))}
    </div>
  );
}
