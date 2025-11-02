import { useState, useEffect } from 'react';
import { Card, Button } from './ui';
import { useToast } from './ui/Toast';
import { LeverageConfig } from './config/LeverageConfig';
import { RiskControlConfig } from './config/RiskControlConfig';
import { AILearningConfig } from './config/AILearningConfig';
import { KlineDataConfig } from './config/KlineDataConfig';
import { CoinPoolConfig } from './config/CoinPoolConfig';
import { TraderList } from './config/TraderList';
import { useConfigManager } from '../hooks/useConfigManager';
import { TraderConfig } from '../types/config';
import TraderFormModal from './TraderFormModal';
import PromptConfig from './PromptConfig';
import { theme } from '../styles/theme';

export default function ConfigManagement() {
  const {
    config,
    loading,
    saving,
    loadConfig,
    updateGlobalConfig,
    saveGlobalConfig,
    saveTrader,
    addTrader,
    deleteTrader,
  } = useConfigManager();

  const toast = useToast();
  const [activeTab, setActiveTab] = useState<'global' | 'traders' | 'prompts'>('global');
  const [editingTrader, setEditingTrader] = useState<TraderConfig | null>(null);
  const [showAddTrader, setShowAddTrader] = useState(false);
  const [traderForm, setTraderForm] = useState<Partial<TraderConfig>>({});

  useEffect(() => {
    loadConfig();
  }, [loadConfig]);

  const handleSaveGlobal = async () => {
    const result = await saveGlobalConfig();
    if (result.success) {
      toast.success(
        result.reloaded 
          ? '✅ 保存成功\n🔄 配置已热重载生效！' 
          : '✅ 保存成功'
      );
    } else {
      toast.error(`❌ 保存失败: ${result.error}`);
    }
  };

  const handleSaveTrader = async (trader: TraderConfig) => {
    const result = await saveTrader(trader);
    if (result.success) {
      toast.success(
        result.reloaded 
          ? '✅ 保存成功\n🔄 配置已热重载生效！' 
          : '✅ 保存成功'
      );
      setEditingTrader(null);
    } else {
      toast.error(`❌ 保存失败: ${result.error}`);
    }
  };

  const handleAddTrader = async (trader: TraderConfig) => {
    const result = await addTrader(trader);
    if (result.success) {
      toast.success(
        result.reloaded 
          ? '✅ 添加成功\n🔄 配置已热重载生效！' 
          : '✅ 添加成功'
      );
      setShowAddTrader(false);
      setTraderForm({});
    } else {
      toast.error(`❌ 添加失败: ${result.error}`);
    }
  };

  const handleDeleteTrader = async (traderId: string) => {
    if (!confirm('确定要删除该Trader吗？此操作不可恢复！')) return;

    const result = await deleteTrader(traderId);
    if (result.success) {
      toast.success(
        result.reloaded 
          ? '✅ 删除成功\n🔄 配置已热重载生效！' 
          : '✅ 删除成功'
      );
    } else {
      toast.error(`❌ 删除失败: ${result.error}`);
    }
  };

  if (loading) {
    return (
      <Card>
        <div style={{ color: theme.colors.text.secondary }}>⏳ 加载配置中...</div>
      </Card>
    );
  }

  if (!config) {
    return (
      <Card>
        <div style={{ color: theme.colors.error.main }}>❌ 加载配置失败</div>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* 标题栏 */}
      <Card variant="purple" icon="⚙️" title="系统配置管理" subtitle="✨ 支持热重载，修改后自动生效，无需重启服务">
        {/* Empty */}
      </Card>

      {/* 标签页 */}
      <div
        className="flex gap-3 p-2 rounded-xl overflow-x-auto"
        style={{ background: theme.colors.background.secondary }}
      >
        <Button
          variant={activeTab === 'global' ? 'primary' : 'ghost'}
          onClick={() => setActiveTab('global')}
          style={{ flex: 1, whiteSpace: 'nowrap' }}
        >
          🌐 全局配置
        </Button>
        <Button
          variant={activeTab === 'traders' ? 'primary' : 'ghost'}
          onClick={() => setActiveTab('traders')}
          style={{ flex: 1, whiteSpace: 'nowrap' }}
        >
          🤖 Trader管理 ({config.traders.length})
        </Button>
        <Button
          variant={activeTab === 'prompts' ? 'primary' : 'ghost'}
          onClick={() => setActiveTab('prompts')}
          style={{ flex: 1, whiteSpace: 'nowrap' }}
        >
          💬 Prompt配置
        </Button>
      </div>

      {/* Prompt配置面板 */}
      {activeTab === 'prompts' && (
        <PromptConfig traderId={config.traders.length > 0 ? config.traders[0].id : ''} />
      )}

      {/* 全局配置面板 */}
      {activeTab === 'global' && (
        <div className="space-y-4">
          <LeverageConfig config={config} onChange={updateGlobalConfig} />
          <RiskControlConfig config={config} onChange={updateGlobalConfig} />
          <AILearningConfig config={config} onChange={updateGlobalConfig} />
          <KlineDataConfig config={config} onChange={updateGlobalConfig} />
          <CoinPoolConfig config={config} onChange={updateGlobalConfig} />

          <Button
            variant="success"
            fullWidth
            onClick={handleSaveGlobal}
            isLoading={saving}
            className="text-lg py-4"
          >
            {saving ? '⏳ 保存中...' : '💾 保存全局配置'}
          </Button>
        </div>
      )}

      {/* Trader管理面板 */}
      {activeTab === 'traders' && (
        <div className="space-y-4">
          <Button
            variant="purple"
            fullWidth
            onClick={() => {
              setShowAddTrader(true);
              setTraderForm({
                id: '',
                name: '',
                enabled: true,
                ai_model: 'deepseek',
                exchange: 'binance',
                initial_balance: 1000,
                scan_interval_minutes: 3,
              });
            }}
          >
            ➕ 添加新Trader
          </Button>

          <TraderList
            traders={config.traders}
            onEdit={setEditingTrader}
            onDelete={handleDeleteTrader}
          />
        </div>
      )}

      {/* 提示信息 */}
      <Card variant="gradient">
        <div className="flex items-start gap-4">
          <div className="text-2xl">⚠️</div>
          <div>
            <h4 className="font-bold mb-2" style={{ color: theme.colors.brand.primary }}>
              重要提示
            </h4>
            <ul className="space-y-2 text-sm" style={{ color: theme.colors.text.secondary }}>
              <li>• 🔄 修改配置后会自动热重载，无需重启服务</li>
              <li>• 🔒 敏感信息（API密钥）已脱敏显示，不修改则保留原值</li>
              <li>• ⚡ 添加/编辑/删除Trader会立即生效</li>
              <li>• 💾 建议修改前先备份config.json文件</li>
              <li>• ⚠️ 如果热重载失败，请手动重启服务</li>
            </ul>
          </div>
        </div>
      </Card>

      {/* Trader编辑弹窗 */}
      {editingTrader && (
        <TraderFormModal
          trader={editingTrader}
          isEdit={true}
          onSave={handleSaveTrader}
          onCancel={() => setEditingTrader(null)}
        />
      )}

      {/* 添加Trader弹窗 */}
      {showAddTrader && (
        <TraderFormModal
          trader={traderForm}
          isEdit={false}
          onSave={handleAddTrader}
          onCancel={() => {
            setShowAddTrader(false);
            setTraderForm({});
          }}
        />
      )}
    </div>
  );
}
