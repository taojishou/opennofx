import { useState, useEffect } from 'react';
import { Modal, Button, Input, Select, Switch } from './ui';
import { useToast } from './ui/Toast';
import { TraderConfig, TraderTemplate } from '../types/config';
import { theme } from '../styles/theme';

interface TraderFormModalProps {
  trader: Partial<TraderConfig>;
  isEdit: boolean;
  onSave: (trader: TraderConfig) => void;
  onCancel: () => void;
  onApplyTemplate?: () => void;
}

const TRADER_TEMPLATES: TraderTemplate[] = [
  {
    name: '币安 Qwen Trader',
    template: {
      exchange: 'binance',
      ai_model: 'qwen',
      initial_balance: 1000,
      scan_interval_minutes: 3,
      enabled: true,
    }
  },
  {
    name: '币安 DeepSeek Trader',
    template: {
      exchange: 'binance',
      ai_model: 'deepseek',
      initial_balance: 1000,
      scan_interval_minutes: 3,
      enabled: true,
    }
  },
  {
    name: 'Hyperliquid DeepSeek',
    template: {
      exchange: 'hyperliquid',
      ai_model: 'deepseek',
      initial_balance: 1000,
      scan_interval_minutes: 3,
      hyperliquid_testnet: false,
      enabled: true,
    }
  },
];

export default function TraderFormModal({
  trader,
  isEdit,
  onSave,
  onCancel,
  onApplyTemplate,
}: TraderFormModalProps) {
  const [form, setForm] = useState<Partial<TraderConfig>>({ ...trader, exchange: trader.exchange || 'binance' });
  const [showSecrets, setShowSecrets] = useState(false);
  const [showTemplates, setShowTemplates] = useState(false);
  const toast = useToast();

  useEffect(() => {
    setForm({ ...trader, exchange: trader.exchange || 'binance' });
  }, [trader]);

  const validateForm = (): string | null => {
    if (!form.id || !form.name || !form.ai_model || !form.exchange) {
      return '请填写所有必填字段';
    }

    if (form.exchange === 'binance' && (!form.binance_api_key || !form.binance_secret_key)) {
      return '使用币安时必须配置API Key和Secret Key';
    }
    if (form.exchange === 'hyperliquid' && !form.hyperliquid_private_key) {
      return '使用Hyperliquid时必须配置Private Key';
    }
    if (form.ai_model === 'qwen' && !form.qwen_key) {
      return '使用Qwen时必须配置API Key';
    }
    if (form.ai_model === 'deepseek' && !form.deepseek_key) {
      return '使用DeepSeek时必须配置API Key';
    }

    return null;
  };

  const handleSubmit = () => {
    const error = validateForm();
    if (error) {
      toast.error(error);
      return;
    }

    const processedForm = { ...form };
    if (processedForm.hyperliquid_private_key?.startsWith('0x') || processedForm.hyperliquid_private_key?.startsWith('0X')) {
      processedForm.hyperliquid_private_key = processedForm.hyperliquid_private_key.slice(2);
    }
    if (processedForm.aster_private_key?.startsWith('0x') || processedForm.aster_private_key?.startsWith('0X')) {
      processedForm.aster_private_key = processedForm.aster_private_key.slice(2);
    }

    onSave(processedForm as TraderConfig);
  };

  const applyTemplate = (template: Partial<TraderConfig>) => {
    setForm({ ...form, ...template });
    setShowTemplates(false);
  };

  return (
    <>
      <Modal
        isOpen={true}
        onClose={onCancel}
        title={isEdit ? '✏️ 编辑Trader' : '➕ 添加新Trader'}
        maxWidth="4xl"
        footer={
          <div className="flex gap-3">
            <Button variant="success" onClick={handleSubmit} fullWidth>
              💾 {isEdit ? '保存修改' : '添加Trader'}
            </Button>
            <Button variant="danger" onClick={onCancel}>
              ❌ 取消
            </Button>
          </div>
        }
      >
        {!isEdit && onApplyTemplate && (
          <div className="mb-4">
            <Button variant="secondary" onClick={() => setShowTemplates(true)}>
              📋 使用模板
            </Button>
          </div>
        )}

        <div className="space-y-6">
          {/* 基本信息 */}
          <div>
            <h3 className="text-lg font-bold mb-4" style={{ color: theme.colors.text.primary }}>
              基本信息
            </h3>
            <div className="grid grid-cols-2 gap-4">
              <Input
                label="Trader ID *"
                hint={!isEdit ? '唯一标识，不可重复' : ''}
                value={form.id || ''}
                onChange={(e) => setForm({ ...form, id: e.target.value })}
                disabled={isEdit}
                placeholder="例如: binance_qwen_01"
                fullWidth
              />
              <Input
                label="Trader名称 *"
                value={form.name || ''}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder="例如: 币安Qwen交易员"
                fullWidth
              />
              <Select
                label="AI模型 *"
                value={form.ai_model || 'deepseek'}
                onChange={(e) => setForm({ ...form, ai_model: e.target.value })}
                fullWidth
              >
                <option value="deepseek">DeepSeek</option>
                <option value="qwen">Qwen (通义千问)</option>
                <option value="custom">Custom API</option>
              </Select>
              <Select
                label="交易平台 *"
                value={form.exchange || 'binance'}
                onChange={(e) => setForm({ ...form, exchange: e.target.value })}
                fullWidth
              >
                <option value="binance">Binance (币安)</option>
                <option value="hyperliquid">Hyperliquid</option>
                <option value="aster">Aster</option>
              </Select>
              <Input
                type="number"
                label="初始资金 (USDT)"
                value={form.initial_balance || 1000}
                onChange={(e) => setForm({ ...form, initial_balance: parseFloat(e.target.value) })}
                fullWidth
              />
              <Input
                type="number"
                label="扫描间隔 (分钟)"
                value={form.scan_interval_minutes || 3}
                onChange={(e) => setForm({ ...form, scan_interval_minutes: parseInt(e.target.value) })}
                fullWidth
              />
            </div>
            <div className="mt-4">
              <Switch
                checked={form.enabled ?? true}
                onChange={(e) => setForm({ ...form, enabled: e.target.checked })}
                label="启用该Trader"
              />
            </div>
          </div>

          {/* 交易所配置 */}
          <div>
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-bold" style={{ color: theme.colors.text.primary }}>
                交易所配置
              </h3>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowSecrets(!showSecrets)}
              >
                {showSecrets ? '🔓 隐藏密钥' : '🔒 显示密钥'}
              </Button>
            </div>

            {form.exchange === 'binance' && (
              <div className="space-y-4">
                <Input
                  type={showSecrets ? 'text' : 'password'}
                  label="Binance API Key *"
                  hint={isEdit ? '留空保持不变' : ''}
                  value={form.binance_api_key || ''}
                  onChange={(e) => setForm({ ...form, binance_api_key: e.target.value })}
                  placeholder={isEdit ? '••••••••••••••••（已配置）' : '输入您的API Key'}
                  fullWidth
                />
                <Input
                  type={showSecrets ? 'text' : 'password'}
                  label="Binance Secret Key *"
                  hint={isEdit ? '留空保持不变' : ''}
                  value={form.binance_secret_key || ''}
                  onChange={(e) => setForm({ ...form, binance_secret_key: e.target.value })}
                  placeholder={isEdit ? '••••••••••••••••（已配置）' : '输入您的Secret Key'}
                  fullWidth
                />
              </div>
            )}

            {form.exchange === 'hyperliquid' && (
              <div className="space-y-4">
                <Input
                  type={showSecrets ? 'text' : 'password'}
                  label="Private Key *"
                  hint="系统会自动去除0x前缀，可直接粘贴完整私钥"
                  value={form.hyperliquid_private_key || ''}
                  onChange={(e) => setForm({ ...form, hyperliquid_private_key: e.target.value })}
                  placeholder={isEdit ? '••••••••••••••••（已配置）' : '输入以太坊私钥'}
                  fullWidth
                />
                <Input
                  label="钱包地址"
                  value={form.hyperliquid_wallet_addr || ''}
                  onChange={(e) => setForm({ ...form, hyperliquid_wallet_addr: e.target.value })}
                  placeholder="0x..."
                  fullWidth
                />
                <Switch
                  checked={form.hyperliquid_testnet ?? false}
                  onChange={(e) => setForm({ ...form, hyperliquid_testnet: e.target.checked })}
                  label="使用测试网"
                />
              </div>
            )}

            {form.exchange === 'aster' && (
              <div className="space-y-4">
                <Input
                  label="主钱包地址 (User) *"
                  value={form.aster_user || ''}
                  onChange={(e) => setForm({ ...form, aster_user: e.target.value })}
                  placeholder="0x..."
                  fullWidth
                />
                <Input
                  label="API钱包地址 (Signer) *"
                  value={form.aster_signer || ''}
                  onChange={(e) => setForm({ ...form, aster_signer: e.target.value })}
                  placeholder="0x..."
                  fullWidth
                />
                <Input
                  type={showSecrets ? 'text' : 'password'}
                  label="API钱包私钥 *"
                  hint="系统会自动去除0x前缀"
                  value={form.aster_private_key || ''}
                  onChange={(e) => setForm({ ...form, aster_private_key: e.target.value })}
                  placeholder={isEdit ? '••••••••••••••••（已配置）' : '输入私钥'}
                  fullWidth
                />
              </div>
            )}
          </div>

          {/* AI配置 */}
          <div>
            <h3 className="text-lg font-bold mb-4" style={{ color: theme.colors.text.primary }}>
              AI配置
            </h3>
            
            {/* AI模式配置 */}
            <div className="space-y-3 mb-6 p-4 rounded-lg" style={{ background: theme.colors.background.tertiary }}>
              <div className="flex items-center justify-between">
                <div>
                  <span className="font-medium" style={{ color: theme.colors.text.primary }}>
                    🚀 AI完全自主模式
                  </span>
                  <div className="text-sm mt-1" style={{ color: theme.colors.text.secondary }}>
                    开启后AI将不受风控限制，完全自主决策（高风险）
                  </div>
                </div>
                <Switch
                  checked={form.ai_autonomy_mode ?? false}
                  onChange={(e) => setForm({ ...form, ai_autonomy_mode: e.target.checked })}
                />
              </div>
              
              <div className="flex items-center justify-between">
                <div>
                  <span className="font-medium" style={{ color: theme.colors.text.primary }}>
                    📦 数据紧凑模式
                  </span>
                  <div className="text-sm mt-1" style={{ color: theme.colors.text.secondary }}>
                    精简候选币种数据，提升AI响应速度（推荐开启）
                  </div>
                </div>
                <Switch
                  checked={form.compact_mode ?? true}
                  onChange={(e) => setForm({ ...form, compact_mode: e.target.checked })}
                />
              </div>
            </div>
            
            {form.ai_model === 'qwen' && (
              <Input
                type={showSecrets ? 'text' : 'password'}
                label="Qwen API Key *"
                value={form.qwen_key || ''}
                onChange={(e) => setForm({ ...form, qwen_key: e.target.value })}
                placeholder={isEdit ? '留空则保留原值' : '输入Qwen API Key'}
                fullWidth
              />
            )}
            {form.ai_model === 'deepseek' && (
              <Input
                type={showSecrets ? 'text' : 'password'}
                label="DeepSeek API Key *"
                value={form.deepseek_key || ''}
                onChange={(e) => setForm({ ...form, deepseek_key: e.target.value })}
                placeholder={isEdit ? '留空则保留原值' : '输入DeepSeek API Key'}
                fullWidth
              />
            )}
            {form.ai_model === 'custom' && (
              <div className="space-y-4">
                <Input
                  label="API URL *"
                  value={form.custom_api_url || ''}
                  onChange={(e) => setForm({ ...form, custom_api_url: e.target.value })}
                  placeholder="https://api.openai.com/v1"
                  fullWidth
                />
                <Input
                  type={showSecrets ? 'text' : 'password'}
                  label="API Key *"
                  value={form.custom_api_key || ''}
                  onChange={(e) => setForm({ ...form, custom_api_key: e.target.value })}
                  placeholder={isEdit ? '留空则保留原值' : 'sk-...'}
                  fullWidth
                />
                <Input
                  label="模型名称 *"
                  value={form.custom_model_name || ''}
                  onChange={(e) => setForm({ ...form, custom_model_name: e.target.value })}
                  placeholder="gpt-4o"
                  fullWidth
                />
              </div>
            )}
          </div>
        </div>
      </Modal>

      {/* 模板选择弹窗 */}
      {showTemplates && (
        <Modal
          isOpen={true}
          onClose={() => setShowTemplates(false)}
          title="📋 选择配置模板"
          maxWidth="lg"
        >
          <div className="space-y-3">
            {TRADER_TEMPLATES.map((template, i) => (
              <button
                key={i}
                onClick={() => applyTemplate(template.template)}
                className="w-full p-4 rounded-xl text-left transition-all hover:scale-105"
                style={{
                  background: theme.colors.background.tertiary,
                  border: `1px solid ${theme.colors.purple.border}`,
                }}
              >
                <div className="font-bold mb-1" style={{ color: theme.colors.text.primary }}>
                  {template.name}
                </div>
                <div className="text-sm" style={{ color: theme.colors.text.secondary }}>
                  {template.template.exchange?.toUpperCase()} + {template.template.ai_model?.toUpperCase()}
                </div>
              </button>
            ))}
          </div>
        </Modal>
      )}
    </>
  );
}
