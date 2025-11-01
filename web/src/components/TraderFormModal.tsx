import { useState, useEffect } from 'react';

interface TraderConfig {
  id: string;
  name: string;
  enabled: boolean;
  ai_model: string;
  exchange: string;
  binance_api_key?: string;
  binance_secret_key?: string;
  hyperliquid_private_key?: string;
  hyperliquid_wallet_addr?: string;
  hyperliquid_testnet?: boolean;
  aster_user?: string;
  aster_signer?: string;
  aster_private_key?: string;
  qwen_key?: string;
  deepseek_key?: string;
  custom_api_url?: string;
  custom_api_key?: string;
  custom_model_name?: string;
  initial_balance: number;
  scan_interval_minutes: number;
}

interface TraderFormModalProps {
  trader: Partial<TraderConfig>;
  isEdit: boolean;
  onSave: (trader: TraderConfig) => void;
  onCancel: () => void;
  onApplyTemplate?: () => void;
}

export default function TraderFormModal({
  trader,
  isEdit,
  onSave,
  onCancel,
  onApplyTemplate
}: TraderFormModalProps) {
  const [form, setForm] = useState<Partial<TraderConfig>>(trader);
  const [showSecrets, setShowSecrets] = useState(false);

  useEffect(() => {
    // 确保exchange字段有默认值
    const updatedTrader = {
      ...trader,
      exchange: trader.exchange || 'binance'
    };
    setForm(updatedTrader);
    console.log('Trader form loaded:', updatedTrader);
  }, [trader]);

  const handleSubmit = () => {
    // 验证必填字段
    if (!form.id || !form.name || !form.ai_model || !form.exchange) {
      alert('请填写所有必填字段');
      return;
    }

    // 验证交易所配置
    if (form.exchange === 'binance' && (!form.binance_api_key || !form.binance_secret_key)) {
      alert('使用币安时必须配置API Key和Secret Key');
      return;
    }
    if (form.exchange === 'hyperliquid' && !form.hyperliquid_private_key) {
      alert('使用Hyperliquid时必须配置Private Key');
      return;
    }

    // 验证AI配置
    if (form.ai_model === 'qwen' && !form.qwen_key) {
      alert('使用Qwen时必须配置API Key');
      return;
    }
    if (form.ai_model === 'deepseek' && !form.deepseek_key) {
      alert('使用DeepSeek时必须配置API Key');
      return;
    }

    // 处理私钥：自动去掉0x前缀
    const processedForm = { ...form };
    if (processedForm.hyperliquid_private_key?.startsWith('0x') || processedForm.hyperliquid_private_key?.startsWith('0X')) {
      processedForm.hyperliquid_private_key = processedForm.hyperliquid_private_key.slice(2);
    }
    if (processedForm.aster_private_key?.startsWith('0x') || processedForm.aster_private_key?.startsWith('0X')) {
      processedForm.aster_private_key = processedForm.aster_private_key.slice(2);
    }

    onSave(processedForm as TraderConfig);
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4"
      style={{ background: 'rgba(0, 0, 0, 0.8)' }}
      onClick={onCancel}
    >
      <div
        className="rounded-2xl p-6 max-w-4xl w-full max-h-[90vh] overflow-y-auto"
        style={{
          background: '#1E2329',
          border: '1px solid rgba(99, 102, 241, 0.3)',
          boxShadow: '0 20px 60px rgba(0, 0, 0, 0.5)'
        }}
        onClick={(e) => e.stopPropagation()}
      >
        {/* 标题 */}
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-2xl font-bold" style={{ color: '#EAECEF' }}>
            {isEdit ? '✏️ 编辑Trader' : '➕ 添加新Trader'}
          </h2>
          <div className="flex gap-2">
            {!isEdit && onApplyTemplate && (
              <button
                onClick={onApplyTemplate}
                className="px-4 py-2 rounded-lg font-semibold transition-all hover:scale-105"
                style={{
                  background: 'rgba(240, 185, 11, 0.2)',
                  color: '#FCD34D',
                  border: '1px solid rgba(240, 185, 11, 0.3)'
                }}
              >
                📋 使用模板
              </button>
            )}
            <button
              onClick={onCancel}
              className="text-2xl hover:scale-110 transition-transform"
              style={{ color: '#848E9C' }}
            >
              ✕
            </button>
          </div>
        </div>

        <div className="space-y-6">
          {/* 基本信息 */}
          <div className="space-y-4">
            <h3 className="text-lg font-bold" style={{ color: '#EAECEF' }}>基本信息</h3>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>
                  Trader ID * {!isEdit && <span className="text-xs">(唯一标识，不可重复)</span>}
                </label>
                <input
                  type="text"
                  value={form.id || ''}
                  onChange={(e) => setForm({ ...form, id: e.target.value })}
                  disabled={isEdit}
                  placeholder="例如: binance_qwen_01"
                  className="w-full px-4 py-2 rounded-lg"
                  style={{
                    background: isEdit ? '#2B3139' : '#0B0E11',
                    border: '1px solid #2B3139',
                    color: '#EAECEF'
                  }}
                />
              </div>
              <div>
                <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>Trader名称 *</label>
                <input
                  type="text"
                  value={form.name || ''}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  placeholder="例如: 币安Qwen交易员"
                  className="w-full px-4 py-2 rounded-lg"
                  style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                />
              </div>
              <div>
                <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>AI模型 *</label>
                <select
                  value={form.ai_model || 'deepseek'}
                  onChange={(e) => setForm({ ...form, ai_model: e.target.value })}
                  className="w-full px-4 py-2 rounded-lg"
                  style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                >
                  <option value="deepseek">DeepSeek</option>
                  <option value="qwen">Qwen (通义千问)</option>
                  <option value="custom">Custom API</option>
                </select>
              </div>
              <div>
                <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>交易平台 *</label>
                <select
                  value={form.exchange || 'binance'}
                  onChange={(e) => setForm({ ...form, exchange: e.target.value })}
                  className="w-full px-4 py-2 rounded-lg"
                  style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                >
                  <option value="binance">Binance (币安)</option>
                  <option value="hyperliquid">Hyperliquid</option>
                  <option value="aster">Aster</option>
                </select>
              </div>
              <div>
                <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>初始资金 (USDT)</label>
                <input
                  type="number"
                  value={form.initial_balance || 1000}
                  onChange={(e) => setForm({ ...form, initial_balance: parseFloat(e.target.value) })}
                  className="w-full px-4 py-2 rounded-lg"
                  style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                />
              </div>
              <div>
                <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>扫描间隔 (分钟)</label>
                <input
                  type="number"
                  value={form.scan_interval_minutes || 3}
                  onChange={(e) => setForm({ ...form, scan_interval_minutes: parseInt(e.target.value) })}
                  className="w-full px-4 py-2 rounded-lg"
                  style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                />
              </div>
            </div>
            <div className="flex items-center gap-3">
              <input
                type="checkbox"
                checked={form.enabled ?? true}
                onChange={(e) => setForm({ ...form, enabled: e.target.checked })}
                className="w-5 h-5"
                style={{ accentColor: '#6366F1' }}
              />
              <label style={{ color: '#EAECEF' }}>启用该Trader</label>
            </div>
          </div>

          {/* 交易所配置 */}
          <div className="space-y-4">
            <h3 className="text-lg font-bold flex items-center gap-2" style={{ color: '#EAECEF' }}>
              交易所配置
              <button
                onClick={() => setShowSecrets(!showSecrets)}
                className="text-sm px-2 py-1 rounded"
                style={{ background: 'rgba(99, 102, 241, 0.2)', color: '#A78BFA' }}
              >
                {showSecrets ? '🔓 隐藏密钥' : '🔒 显示密钥'}
              </button>
            </h3>

            {form.exchange === 'binance' && (
              <div className="space-y-4 p-4 rounded-lg" style={{ background: '#0B0E11', border: '1px solid #2B3139' }}>
                <div className="flex items-center justify-between mb-3">
                  <h4 className="font-semibold flex items-center gap-2" style={{ color: '#EAECEF' }}>
                    🔶 Binance API配置
                  </h4>
                  {isEdit && (
                    <span className="text-xs px-2 py-1 rounded" style={{ background: 'rgba(14, 203, 129, 0.1)', color: '#0ECB81' }}>
                      ✓ 已配置
                    </span>
                  )}
                </div>
                <div>
                  <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>
                    Binance API Key * {isEdit && <span className="text-xs">(留空保持不变)</span>}
                  </label>
                  <input
                    type={showSecrets ? 'text' : 'password'}
                    value={form.binance_api_key || ''}
                    onChange={(e) => setForm({ ...form, binance_api_key: e.target.value })}
                    placeholder={isEdit ? '••••••••••••••••（已配置）' : '输入您的API Key'}
                    className="w-full px-4 py-2 rounded-lg font-mono text-sm"
                    style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                  />
                </div>
                <div>
                  <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>
                    Binance Secret Key * {isEdit && <span className="text-xs">(留空保持不变)</span>}
                  </label>
                  <input
                    type={showSecrets ? 'text' : 'password'}
                    value={form.binance_secret_key || ''}
                    onChange={(e) => setForm({ ...form, binance_secret_key: e.target.value })}
                    placeholder={isEdit ? '••••••••••••••••（已配置）' : '输入您的Secret Key'}
                    className="w-full px-4 py-2 rounded-lg font-mono text-sm"
                    style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                  />
                </div>
              </div>
            )}

            {form.exchange === 'hyperliquid' && (
              <div className="space-y-4 p-4 rounded-lg" style={{ background: '#0B0E11', border: '1px solid #2B3139' }}>
                <div className="flex items-center justify-between mb-3">
                  <h4 className="font-semibold flex items-center gap-2" style={{ color: '#EAECEF' }}>
                    🌊 Hyperliquid配置
                  </h4>
                  {isEdit && (
                    <span className="text-xs px-2 py-1 rounded" style={{ background: 'rgba(14, 203, 129, 0.1)', color: '#0ECB81' }}>
                      ✓ 已配置
                    </span>
                  )}
                </div>
                <div>
                  <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>
                    Private Key * {isEdit && <span className="text-xs">(留空保持不变)</span>}
                  </label>
                  <input
                    type={showSecrets ? 'text' : 'password'}
                    value={form.hyperliquid_private_key || ''}
                    onChange={(e) => setForm({ ...form, hyperliquid_private_key: e.target.value })}
                    placeholder={isEdit ? '••••••••••••••••（已配置）' : '输入以太坊私钥（自动去除0x前缀）'}
                    className="w-full px-4 py-2 rounded-lg font-mono text-sm"
                    style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                  />
                  <div className="text-xs mt-1" style={{ color: '#848E9C' }}>
                    💡 系统会自动去除0x前缀，可直接粘贴完整私钥
                  </div>
                </div>
                <div>
                  <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>钱包地址</label>
                  <input
                    type="text"
                    value={form.hyperliquid_wallet_addr || ''}
                    onChange={(e) => setForm({ ...form, hyperliquid_wallet_addr: e.target.value })}
                    placeholder="0x..."
                    className="w-full px-4 py-2 rounded-lg font-mono text-sm"
                    style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                  />
                </div>
                <div className="flex items-center gap-3">
                  <input
                    type="checkbox"
                    checked={form.hyperliquid_testnet ?? false}
                    onChange={(e) => setForm({ ...form, hyperliquid_testnet: e.target.checked })}
                    className="w-5 h-5"
                    style={{ accentColor: '#6366F1' }}
                  />
                  <label style={{ color: '#EAECEF' }}>使用测试网</label>
                </div>
              </div>
            )}

            {form.exchange === 'aster' && (
              <div className="space-y-4 p-4 rounded-lg" style={{ background: '#0B0E11', border: '1px solid #2B3139' }}>
                <div className="flex items-center justify-between mb-3">
                  <h4 className="font-semibold flex items-center gap-2" style={{ color: '#EAECEF' }}>
                    ⭐ Aster配置
                  </h4>
                  {isEdit && (
                    <span className="text-xs px-2 py-1 rounded" style={{ background: 'rgba(14, 203, 129, 0.1)', color: '#0ECB81' }}>
                      ✓ 已配置
                    </span>
                  )}
                </div>
                <div>
                  <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>主钱包地址 (User) *</label>
                  <input
                    type="text"
                    value={form.aster_user || ''}
                    onChange={(e) => setForm({ ...form, aster_user: e.target.value })}
                    placeholder="0x..."
                    className="w-full px-4 py-2 rounded-lg font-mono text-sm"
                    style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                  />
                </div>
                <div>
                  <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>API钱包地址 (Signer) *</label>
                  <input
                    type="text"
                    value={form.aster_signer || ''}
                    onChange={(e) => setForm({ ...form, aster_signer: e.target.value })}
                    placeholder="0x..."
                    className="w-full px-4 py-2 rounded-lg font-mono text-sm"
                    style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                  />
                </div>
                <div>
                  <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>
                    API钱包私钥 * {isEdit && <span className="text-xs">(留空保持不变)</span>}
                  </label>
                  <input
                    type={showSecrets ? 'text' : 'password'}
                    value={form.aster_private_key || ''}
                    onChange={(e) => setForm({ ...form, aster_private_key: e.target.value })}
                    placeholder={isEdit ? '••••••••••••••••（已配置）' : '输入私钥（自动去除0x前缀）'}
                    className="w-full px-4 py-2 rounded-lg font-mono text-sm"
                    style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                  />
                  <div className="text-xs mt-1" style={{ color: '#848E9C' }}>
                    💡 系统会自动去除0x前缀，可直接粘贴完整私钥
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* AI配置 */}
          <div className="space-y-4">
            <h3 className="text-lg font-bold" style={{ color: '#EAECEF' }}>AI配置</h3>
            {form.ai_model === 'qwen' && (
              <div>
                <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>Qwen API Key *</label>
                <input
                  type={showSecrets ? 'text' : 'password'}
                  value={form.qwen_key || ''}
                  onChange={(e) => setForm({ ...form, qwen_key: e.target.value })}
                  placeholder={isEdit ? '留空则保留原值' : '输入Qwen API Key'}
                  className="w-full px-4 py-2 rounded-lg font-mono text-sm"
                  style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                />
              </div>
            )}
            {form.ai_model === 'deepseek' && (
              <div>
                <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>DeepSeek API Key *</label>
                <input
                  type={showSecrets ? 'text' : 'password'}
                  value={form.deepseek_key || ''}
                  onChange={(e) => setForm({ ...form, deepseek_key: e.target.value })}
                  placeholder={isEdit ? '留空则保留原值' : '输入DeepSeek API Key'}
                  className="w-full px-4 py-2 rounded-lg font-mono text-sm"
                  style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                />
              </div>
            )}
            {form.ai_model === 'custom' && (
              <div className="space-y-4">
                <div>
                  <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>API URL *</label>
                  <input
                    type="text"
                    value={form.custom_api_url || ''}
                    onChange={(e) => setForm({ ...form, custom_api_url: e.target.value })}
                    placeholder="https://api.openai.com/v1"
                    className="w-full px-4 py-2 rounded-lg"
                    style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                  />
                </div>
                <div>
                  <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>API Key *</label>
                  <input
                    type={showSecrets ? 'text' : 'password'}
                    value={form.custom_api_key || ''}
                    onChange={(e) => setForm({ ...form, custom_api_key: e.target.value })}
                    placeholder={isEdit ? '留空则保留原值' : 'sk-...'}
                    className="w-full px-4 py-2 rounded-lg font-mono text-sm"
                    style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                  />
                </div>
                <div>
                  <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>模型名称 *</label>
                  <input
                    type="text"
                    value={form.custom_model_name || ''}
                    onChange={(e) => setForm({ ...form, custom_model_name: e.target.value })}
                    placeholder="gpt-4o"
                    className="w-full px-4 py-2 rounded-lg"
                    style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                  />
                </div>
              </div>
            )}
          </div>

          {/* 按钮 */}
          <div className="flex gap-3 pt-4">
            <button
              onClick={handleSubmit}
              className="flex-1 px-6 py-3 rounded-xl font-bold transition-all hover:scale-105"
              style={{
                background: 'linear-gradient(135deg, #10B981 0%, #0ECB81 100%)',
                color: '#FFFFFF',
                boxShadow: '0 4px 16px rgba(16, 185, 129, 0.3)'
              }}
            >
              💾 {isEdit ? '保存修改' : '添加Trader'}
            </button>
            <button
              onClick={onCancel}
              className="px-6 py-3 rounded-xl font-bold transition-all hover:scale-105"
              style={{
                background: 'rgba(248, 113, 113, 0.2)',
                color: '#FCA5A5',
                border: '1px solid rgba(248, 113, 113, 0.3)'
              }}
            >
              ❌ 取消
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
