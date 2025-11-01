import { useState, useEffect } from 'react';
import TraderFormModal from './TraderFormModal';
import PromptConfig from './PromptConfig';

interface TraderConfig {
  id: string;
  name: string;
  enabled: boolean;
  is_paused?: boolean;
  is_running?: boolean;
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

interface KlineConfig {
  interval: string;
  limit: number;
  show_table: boolean;
}

interface MarketDataConfig {
  klines: KlineConfig[];
}

interface SystemConfig {
  traders: TraderConfig[];
  leverage: {
    btc_eth_leverage: number;
    altcoin_leverage: number;
  };
  use_default_coins: boolean;
  default_coins: string[];
  coin_pool_api_url: string;
  oi_top_api_url: string;
  api_server_port: number;
  max_positions: number;
  max_daily_loss: number;
  max_drawdown: number;
  stop_trading_minutes: number;
  enable_ai_learning?: boolean;
  ai_learn_interval?: number;
  market_data?: MarketDataConfig;
}

export default function ConfigManagement() {
  const [config, setConfig] = useState<SystemConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [activeTab, setActiveTab] = useState<'global' | 'traders' | 'prompts'>('global');
  const [editingTrader, setEditingTrader] = useState<TraderConfig | null>(null);
  const [showAddTrader, setShowAddTrader] = useState(false);
  const [traderForm, setTraderForm] = useState<Partial<TraderConfig>>({});
  const [showTemplates, setShowTemplates] = useState(false);

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/config');
      const data = await response.json();
      if (data.success) {
        setConfig(data.data);
      }
    } catch (error) {
      console.error('加载配置失败:', error);
      alert('加载配置失败');
    } finally {
      setLoading(false);
    }
  };

  const updateGlobalConfig = (updates: Partial<SystemConfig>) => {
    if (config) {
      setConfig({ ...config, ...updates });
    }
  };

  const reloadConfig = async () => {
    try {
      const response = await fetch('/api/config/reload', {
        method: 'POST',
      });
      const data = await response.json();
      if (data.success) {
        return true;
      } else {
        console.error('热重载失败:', data.error);
        return false;
      }
    } catch (error) {
      console.error('热重载请求失败:', error);
      return false;
    }
  };

  const saveGlobalConfig = async () => {
    if (!config) return;

    try {
      setSaving(true);
      
      // 确保ai_learn_interval有默认值
      const aiLearnInterval = config.ai_learn_interval === undefined || config.ai_learn_interval === 0 
        ? 10 
        : config.ai_learn_interval;

      // 验证并修复market_data中的空值
      let marketData = config.market_data;
      if (marketData && marketData.klines) {
        marketData = {
          klines: marketData.klines.map(k => ({
            interval: k.interval,
            limit: k.limit || 20, // 空值设为默认20
            show_table: k.show_table
          }))
        };
      }
      
      const response = await fetch('/api/config/global/update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          use_default_coins: config.use_default_coins,
          default_coins: config.default_coins,
          coin_pool_api_url: config.coin_pool_api_url,
          oi_top_api_url: config.oi_top_api_url,
          max_positions: config.max_positions,
          max_daily_loss: config.max_daily_loss,
          max_drawdown: config.max_drawdown,
          stop_trading_minutes: config.stop_trading_minutes,
          btc_eth_leverage: config.leverage.btc_eth_leverage,
          altcoin_leverage: config.leverage.altcoin_leverage,
          enable_ai_learning: config.enable_ai_learning,
          ai_learn_interval: aiLearnInterval,
          market_data: marketData,
        }),
      });
      const data = await response.json();
      if (data.success) {
        // 尝试热重载
        const reloaded = await reloadConfig();
        if (reloaded) {
          alert('✅ ' + data.message + '\n🔄 配置已热重载生效！');
        } else {
          alert('✅ ' + data.message);
        }
        loadConfig(); // 重新加载配置
      } else {
        alert('❌ 保存失败: ' + (data.error || '未知错误'));
      }
    } catch (error) {
      console.error('保存失败:', error);
      alert('保存失败');
    } finally {
      setSaving(false);
    }
  };

  const saveTraderConfig = async (trader: TraderConfig) => {
    try {
      setSaving(true);
      const response = await fetch('/api/config/trader/update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(trader),
      });
      const data = await response.json();
      if (data.success) {
        const reloaded = await reloadConfig();
        if (reloaded) {
          alert('✅ ' + data.message + '\n🔄 配置已热重载生效！');
        } else {
          alert('✅ ' + data.message);
        }
        setEditingTrader(null);
        loadConfig();
      } else {
        alert('❌ 保存失败: ' + (data.error || '未知错误'));
      }
    } catch (error) {
      console.error('保存失败:', error);
      alert('保存失败');
    } finally {
      setSaving(false);
    }
  };

  const addTrader = async (trader: TraderConfig) => {
    try {
      setSaving(true);
      const response = await fetch('/api/config/trader/add', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(trader),
      });
      const data = await response.json();
      if (data.success) {
        const reloaded = await reloadConfig();
        if (reloaded) {
          alert('✅ ' + data.message + '\n🔄 配置已热重载生效！');
        } else {
          alert('✅ ' + data.message);
        }
        setShowAddTrader(false);
        setTraderForm({});
        loadConfig();
      } else {
        alert('❌ 添加失败: ' + (data.error || '未知错误'));
      }
    } catch (error) {
      console.error('添加失败:', error);
      alert('添加失败');
    } finally {
      setSaving(false);
    }
  };

  const deleteTrader = async (traderId: string) => {
    if (!confirm('确定要删除该Trader吗？此操作不可恢复！')) return;

    try {
      setSaving(true);
      const response = await fetch(`/api/config/trader/delete?trader_id=${traderId}`, {
        method: 'DELETE',
      });
      const data = await response.json();
      if (data.success) {
        const reloaded = await reloadConfig();
        if (reloaded) {
          alert('✅ ' + data.message + '\n🔄 配置已热重载生效！');
        } else {
          alert('✅ ' + data.message);
        }
        loadConfig();
      } else {
        alert('❌ 删除失败: ' + (data.error || '未知错误'));
      }
    } catch (error) {
      console.error('删除失败:', error);
      alert('删除失败');
    } finally {
      setSaving(false);
    }
  };

  const applyTemplate = (template: Partial<TraderConfig>) => {
    setTraderForm({ ...traderForm, ...template });
    setShowTemplates(false);
  };

  const traderTemplates = [
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

  if (loading) {
    return (
      <div className="rounded-2xl p-8" style={{ background: '#1E2329', border: '1px solid #2B3139' }}>
        <div style={{ color: '#848E9C' }}>⏳ 加载配置中...</div>
      </div>
    );
  }

  if (!config) {
    return (
      <div className="rounded-2xl p-8" style={{ background: '#1E2329', border: '1px solid #2B3139' }}>
        <div style={{ color: '#F6465D' }}>❌ 加载配置失败</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* 标题栏 */}
      <div className="relative rounded-2xl p-6 overflow-hidden" style={{
        background: 'linear-gradient(135deg, rgba(99, 102, 241, 0.15) 0%, rgba(139, 92, 246, 0.1) 50%, rgba(30, 35, 41, 0.8) 100%)',
        border: '1px solid rgba(99, 102, 241, 0.3)',
        boxShadow: '0 8px 32px rgba(99, 102, 241, 0.2)'
      }}>
        <div className="absolute top-0 right-0 w-96 h-96 rounded-full opacity-10" style={{
          background: 'radial-gradient(circle, #6366F1 0%, transparent 70%)',
          filter: 'blur(60px)'
        }} />
        <div className="relative flex items-center gap-4">
          <div className="w-16 h-16 rounded-2xl flex items-center justify-center text-3xl" style={{
            background: 'linear-gradient(135deg, #6366F1 0%, #8B5CF6 100%)',
            boxShadow: '0 8px 24px rgba(99, 102, 241, 0.5)',
            border: '2px solid rgba(255, 255, 255, 0.1)'
          }}>
            ⚙️
          </div>
          <div>
            <h2 className="text-3xl font-bold mb-1" style={{
              color: '#EAECEF',
              textShadow: '0 2px 8px rgba(99, 102, 241, 0.3)'
            }}>
              系统配置管理
            </h2>
            <p className="text-base" style={{ color: '#0ECB81' }}>
              ✨ 支持热重载，修改后自动生效，无需重启服务
            </p>
          </div>
        </div>
      </div>

      {/* 标签页 */}
      <div className="flex gap-3 p-2 rounded-xl overflow-x-auto" style={{ background: '#1E2329' }}>
        <button
          onClick={() => setActiveTab('global')}
          className="flex-1 px-6 py-3 rounded-lg font-bold transition-all whitespace-nowrap"
          style={activeTab === 'global'
            ? { background: 'linear-gradient(135deg, #6366F1 0%, #8B5CF6 100%)', color: '#FFF' }
            : { background: 'transparent', color: '#848E9C' }
          }
        >
          🌐 全局配置
        </button>
        <button
          onClick={() => setActiveTab('traders')}
          className="flex-1 px-6 py-3 rounded-lg font-bold transition-all whitespace-nowrap"
          style={activeTab === 'traders'
            ? { background: 'linear-gradient(135deg, #6366F1 0%, #8B5CF6 100%)', color: '#FFF' }
            : { background: 'transparent', color: '#848E9C' }
          }
        >
          🤖 Trader管理 ({config.traders.length})
        </button>
        <button
          onClick={() => setActiveTab('prompts')}
          className="flex-1 px-6 py-3 rounded-lg font-bold transition-all whitespace-nowrap"
          style={activeTab === 'prompts'
            ? { background: 'linear-gradient(135deg, #6366F1 0%, #8B5CF6 100%)', color: '#FFF' }
            : { background: 'transparent', color: '#848E9C' }
          }
        >
          💬 Prompt配置
        </button>
      </div>

      {/* Prompt配置面板 */}
      {activeTab === 'prompts' && (
        <div>
          <PromptConfig traderId={config.traders.length > 0 ? config.traders[0].id : ''} />
        </div>
      )}

      {/* 全局配置面板 */}
      {activeTab === 'global' && (
        <div className="space-y-4">
          {/* 杠杆配置 */}
          <div className="rounded-2xl p-6" style={{ background: '#1E2329', border: '1px solid #2B3139' }}>
            <h3 className="text-xl font-bold mb-4" style={{ color: '#EAECEF' }}>⚖️ 杠杆配置</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>
                  BTC/ETH杠杆倍数
                  <span className="ml-2 text-xs" style={{ color: '#6EE7B7' }}>建议: 3-10倍</span>
                </label>
                <input
                  type="number"
                  min="1"
                  max="50"
                  value={config.leverage.btc_eth_leverage}
                  onChange={(e) => setConfig({
                    ...config,
                    leverage: { ...config.leverage, btc_eth_leverage: parseInt(e.target.value) }
                  })}
                  className="w-full px-4 py-2 rounded-lg text-lg font-semibold"
                  style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                />
              </div>
              <div>
                <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>
                  山寨币杠杆倍数
                  <span className="ml-2 text-xs" style={{ color: '#6EE7B7' }}>建议: 2-5倍</span>
                </label>
                <input
                  type="number"
                  min="1"
                  max="20"
                  value={config.leverage.altcoin_leverage}
                  onChange={(e) => setConfig({
                    ...config,
                    leverage: { ...config.leverage, altcoin_leverage: parseInt(e.target.value) }
                  })}
                  className="w-full px-4 py-2 rounded-lg text-lg font-semibold"
                  style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                />
              </div>
            </div>
          </div>

          {/* 风控配置 */}
          <div className="rounded-2xl p-6" style={{ background: '#1E2329', border: '1px solid #2B3139' }}>
            <h3 className="text-xl font-bold mb-4 flex items-center gap-2" style={{ color: '#EAECEF' }}>
              <span>🛡️ 风险控制</span>
              <span className="text-xs px-2 py-1 rounded" style={{ background: 'rgba(16, 185, 129, 0.2)', color: '#6EE7B7' }}>
                重要
              </span>
            </h3>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>最大持仓数</label>
                <input
                  type="number"
                  value={config.max_positions}
                  onChange={(e) => setConfig({ ...config, max_positions: parseInt(e.target.value) })}
                  className="w-full px-4 py-2 rounded-lg"
                  style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                />
              </div>
              <div>
                <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>最大日亏损(%)</label>
                <input
                  type="number"
                  step="0.1"
                  value={config.max_daily_loss}
                  onChange={(e) => setConfig({ ...config, max_daily_loss: parseFloat(e.target.value) })}
                  className="w-full px-4 py-2 rounded-lg"
                  style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                />
              </div>
              <div>
                <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>最大回撤(%)</label>
                <input
                  type="number"
                  step="0.1"
                  value={config.max_drawdown}
                  onChange={(e) => setConfig({ ...config, max_drawdown: parseFloat(e.target.value) })}
                  className="w-full px-4 py-2 rounded-lg"
                  style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                />
              </div>
              <div>
                <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>暂停交易时长(分钟)</label>
                <input
                  type="number"
                  value={config.stop_trading_minutes}
                  onChange={(e) => setConfig({ ...config, stop_trading_minutes: parseInt(e.target.value) })}
                  className="w-full px-4 py-2 rounded-lg"
                  style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                />
              </div>
            </div>
          </div>

          {/* AI学习配置 */}
          <div className="rounded-2xl p-6" style={{ background: '#1E2329', border: '1px solid #2B3139' }}>
            <h3 className="text-xl font-bold mb-4 flex items-center gap-2" style={{ color: '#EAECEF' }}>
              <span>🧠 AI自动学习</span>
            </h3>
            <div className="space-y-4">
              <div className="flex items-center justify-between p-4 rounded-lg" style={{ background: '#2B3139' }}>
                <div className="flex-1">
                  <div className="font-semibold mb-1" style={{ color: '#EAECEF' }}>启用AI自动学习</div>
                  <div className="text-sm" style={{ color: '#848E9C' }}>
                    AI会每隔N个周期自动分析历史交易，总结成功/失败模式，避免重复错误
                  </div>
                </div>
                <label className="relative inline-flex items-center cursor-pointer ml-4">
                  <input
                    type="checkbox"
                    checked={config.enable_ai_learning || false}
                    onChange={(e) => updateGlobalConfig({ enable_ai_learning: e.target.checked })}
                    className="sr-only peer"
                  />
                  <div className="w-11 h-6 rounded-full peer peer-focus:ring-2 peer-focus:ring-blue-300 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all" style={{
                    background: config.enable_ai_learning ? '#0ECB81' : '#474D57'
                  }}></div>
                </label>
              </div>
              {config.enable_ai_learning && (
                <div className="p-4 rounded-lg" style={{ background: '#2B3139' }}>
                  <div className="flex items-center gap-3">
                    <label className="text-sm font-medium" style={{ color: '#EAECEF' }}>学习间隔:</label>
                    <input
                      type="number"
                      min="5"
                      max="50"
                      value={config.ai_learn_interval === undefined || config.ai_learn_interval === 0 ? 10 : config.ai_learn_interval}
                      onChange={(e) => {
                        const val = parseInt(e.target.value);
                        updateGlobalConfig({ ai_learn_interval: isNaN(val) || val < 5 ? 10 : val });
                      }}
                      className="w-24 px-3 py-2 rounded-lg text-center"
                      style={{ background: '#1E2329', color: '#EAECEF', border: '1px solid #474D57' }}
                    />
                    <span className="text-sm" style={{ color: '#848E9C' }}>个周期</span>
                    <span className="text-xs px-2 py-1 rounded" style={{ background: 'rgba(14, 203, 129, 0.1)', color: '#0ECB81' }}>
                      推荐: 10
                    </span>
                  </div>
                  <div className="mt-3 text-xs" style={{ color: '#848E9C' }}>
                    💡 提示：间隔太短可能增加成本，间隔太长学习效果不明显
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* K线数据配置 */}
          <div className="rounded-2xl p-6 mb-6" style={{ background: '#1E2329', border: '1px solid #2B3139' }}>
            <h3 className="text-xl font-bold mb-4 flex items-center gap-2" style={{ color: '#EAECEF' }}>
              <span>📊 K线数据配置</span>
            </h3>

            {(!config.market_data || !config.market_data.klines || config.market_data.klines.length === 0) ? (
              <div className="p-4 mb-4 rounded-lg" style={{ background: '#2B3139', border: '1px solid #474D57' }}>
                <p className="text-sm mb-3" style={{ color: '#848E9C' }}>
                  未配置K线数据，将使用默认设置（3分钟20根 + 4小时60根）
                </p>
                <button
                  onClick={() => {
                    updateGlobalConfig({
                      market_data: {
                        klines: [
                          { interval: '3m', limit: 5, show_table: true },
                          { interval: '15m', limit: 10, show_table: false },
                          { interval: '4h', limit: 60, show_table: false }
                        ]
                      }
                    });
                  }}
                  className="px-4 py-2 rounded-lg font-medium"
                  style={{ background: '#0ECB81', color: '#FFFFFF', border: 'none', cursor: 'pointer' }}
                >
                  初始化推荐配置
                </button>
              </div>
            ) : (
              <div className="space-y-3">
                {config.market_data.klines.map((kline, index) => (
                  <div key={index} className="p-4 rounded-lg" style={{ background: '#2B3139', border: '1px solid #474D57' }}>
                    <div className="flex items-center justify-between mb-4">
                      <h4 className="font-semibold" style={{ color: '#EAECEF' }}>
                        K线 #{index + 1}
                      </h4>
                      {config.market_data!.klines.length > 1 && (
                        <button
                          onClick={() => {
                            const newKlines = config.market_data!.klines.filter((_, i) => i !== index);
                            updateGlobalConfig({
                              market_data: { klines: newKlines }
                            });
                          }}
                          className="px-3 py-1 rounded text-sm"
                          style={{ background: '#F6465D', color: '#FFFFFF', border: 'none', cursor: 'pointer' }}
                        >
                          删除
                        </button>
                      )}
                    </div>

                    <div className="grid grid-cols-3 gap-4">
                      <div>
                        <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>时间周期</label>
                        <select
                          value={kline.interval}
                          onChange={(e) => {
                            const newKlines = [...config.market_data!.klines];
                            newKlines[index].interval = e.target.value;
                            updateGlobalConfig({ market_data: { klines: newKlines } });
                          }}
                          className="w-full px-3 py-2 rounded-lg"
                          style={{ background: '#1E2329', color: '#EAECEF', border: '1px solid #474D57' }}
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
                        </select>
                      </div>

                      <div>
                        <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>K线数量</label>
                        <input
                          type="number"
                          min="5"
                          max="200"
                          value={kline.limit || ''}
                          onChange={(e) => {
                            const val = e.target.value;
                            const newKlines = [...config.market_data!.klines];
                            // 允许空值或有效数字
                            if (val === '') {
                              newKlines[index].limit = null as any; // 临时允许空值
                            } else {
                              const num = parseInt(val);
                              if (!isNaN(num)) {
                                newKlines[index].limit = num;
                              }
                            }
                            updateGlobalConfig({ market_data: { klines: newKlines } });
                          }}
                          onBlur={(e) => {
                            // 失焦时确保有效值
                            const val = e.target.value;
                            if (val === '' || parseInt(val) < 5) {
                              const newKlines = [...config.market_data!.klines];
                              newKlines[index].limit = 20;
                              updateGlobalConfig({ market_data: { klines: newKlines } });
                            } else if (parseInt(val) > 200) {
                              const newKlines = [...config.market_data!.klines];
                              newKlines[index].limit = 200;
                              updateGlobalConfig({ market_data: { klines: newKlines } });
                            }
                          }}
                          className="w-full px-3 py-2 rounded-lg"
                          style={{ background: '#1E2329', color: '#EAECEF', border: '1px solid #474D57' }}
                        />
                      </div>

                      <div>
                        <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>显示表格</label>
                        <label className="flex items-center cursor-pointer pt-2">
                          <input
                            type="checkbox"
                            checked={kline.show_table}
                            onChange={(e) => {
                              const newKlines = [...config.market_data!.klines];
                              newKlines[index].show_table = e.target.checked;
                              updateGlobalConfig({ market_data: { klines: newKlines } });
                            }}
                            className="w-5 h-5 mr-2"
                            style={{ accentColor: '#0ECB81' }}
                          />
                          <span style={{ color: '#EAECEF' }}>显示K线表格</span>
                        </label>
                      </div>
                    </div>
                  </div>
                ))}

                {config.market_data.klines.length < 5 && (
                  <button
                    onClick={() => {
                      const newKline = { interval: '15m', limit: 10, show_table: false };
                      updateGlobalConfig({
                        market_data: {
                          klines: [...(config.market_data?.klines || []), newKline]
                        }
                      });
                    }}
                    className="w-full px-4 py-2 rounded-lg font-medium"
                    style={{ background: '#0ECB81', color: '#FFFFFF', border: 'none', cursor: 'pointer' }}
                  >
                    + 添加K线配置
                  </button>
                )}
              </div>
            )}

            <div className="mt-4 p-3 rounded-lg" style={{ background: 'rgba(240, 185, 11, 0.1)', border: '1px solid rgba(240, 185, 11, 0.3)' }}>
              <p className="text-sm leading-relaxed" style={{ color: '#F0B90B', margin: 0 }}>
                💡 <strong>建议</strong>: K线数据过多会增加prompt大小，可能导致AI过度交易。<br/>
                推荐：3分钟5根（参考）+ 15分钟10根（决策）+ 4小时60根（趋势）
              </p>
            </div>
          </div>

          {/* 币种池配置 */}
          <div className="rounded-2xl p-6" style={{ background: '#1E2329', border: '1px solid #2B3139' }}>
            <h3 className="text-xl font-bold mb-4 flex items-center gap-2" style={{ color: '#EAECEF' }}>
              <span>🪙 币种池配置</span>
            </h3>
            <div className="space-y-4">
              <div className="flex items-center gap-3">
                <input
                  type="checkbox"
                  checked={config.use_default_coins}
                  onChange={(e) => setConfig({ ...config, use_default_coins: e.target.checked })}
                  className="w-5 h-5"
                  style={{ accentColor: '#6366F1' }}
                />
                <label style={{ color: '#EAECEF' }}>使用默认币种列表</label>
              </div>
              <div>
                <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>
                  默认币种 (逗号分隔)
                </label>
                <input
                  type="text"
                  value={config.default_coins.join(', ')}
                  onChange={(e) => setConfig({
                    ...config,
                    default_coins: e.target.value.split(',').map(s => s.trim())
                  })}
                  className="w-full px-4 py-2 rounded-lg"
                  style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                />
              </div>
              <div>
                <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>AI500币种池API</label>
                <input
                  type="text"
                  value={config.coin_pool_api_url}
                  onChange={(e) => setConfig({ ...config, coin_pool_api_url: e.target.value })}
                  placeholder="留空则不使用"
                  className="w-full px-4 py-2 rounded-lg"
                  style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                />
              </div>
              <div>
                <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>OI Top API</label>
                <input
                  type="text"
                  value={config.oi_top_api_url}
                  onChange={(e) => setConfig({ ...config, oi_top_api_url: e.target.value })}
                  placeholder="留空则不使用"
                  className="w-full px-4 py-2 rounded-lg"
                  style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
                />
              </div>
            </div>
          </div>

          {/* 保存按钮 */}
          <button
            onClick={saveGlobalConfig}
            disabled={saving}
            className="w-full px-6 py-4 rounded-xl font-bold text-lg transition-all hover:scale-105 disabled:opacity-50"
            style={{
              background: 'linear-gradient(135deg, #10B981 0%, #0ECB81 100%)',
              color: '#FFFFFF',
              boxShadow: '0 4px 16px rgba(16, 185, 129, 0.3)'
            }}
          >
            {saving ? '⏳ 保存中...' : '💾 保存全局配置'}
          </button>
        </div>
      )}

      {/* Trader管理面板 */}
      {activeTab === 'traders' && (
        <div className="space-y-4">
          {/* 添加新Trader按钮 */}
          <button
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
            className="w-full px-6 py-4 rounded-xl font-bold transition-all hover:scale-105"
            style={{
              background: 'linear-gradient(135deg, #6366F1 0%, #8B5CF6 100%)',
              color: '#FFF',
              boxShadow: '0 4px 16px rgba(99, 102, 241, 0.3)'
            }}
          >
            ➕ 添加新Trader
          </button>

          {config.traders.map((trader) => (
            <div
              key={trader.id}
              className="rounded-2xl p-6"
              style={{
                background: '#1E2329',
                border: trader.enabled ? '1px solid rgba(99, 102, 241, 0.3)' : '1px solid #2B3139'
              }}
            >
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-4">
                  <div className={`w-3 h-3 rounded-full ${trader.enabled ? 'bg-green-500' : 'bg-gray-500'}`} />
                  <div>
                    <div className="flex items-center gap-3 mb-2">
                      <h3 className="text-xl font-bold" style={{ color: '#EAECEF' }}>{trader.name}</h3>
                    </div>
                    <div className="text-sm" style={{ color: '#848E9C' }}>
                      ID: {trader.id} | {trader.ai_model.toUpperCase()} @ {trader.exchange.toUpperCase()}
                    </div>
                  </div>
                </div>
                <div className="flex gap-3">
                  <button
                    onClick={() => setEditingTrader(trader)}
                    className="px-4 py-2 rounded-lg font-semibold transition-all hover:scale-105"
                    style={{
                      background: 'rgba(99, 102, 241, 0.2)',
                      color: '#A78BFA',
                      border: '1px solid rgba(99, 102, 241, 0.3)'
                    }}
                  >
                    ✏️ 编辑
                  </button>
                  <button
                    onClick={() => deleteTrader(trader.id)}
                    className="px-4 py-2 rounded-lg font-semibold transition-all hover:scale-105"
                    style={{
                      background: 'rgba(248, 113, 113, 0.2)',
                      color: '#FCA5A5',
                      border: '1px solid rgba(248, 113, 113, 0.3)'
                    }}
                  >
                    🗑️ 删除
                  </button>
                </div>
              </div>
              <div className="grid grid-cols-3 gap-4 text-sm">
                <div>
                  <span style={{ color: '#848E9C' }}>初始资金: </span>
                  <span style={{ color: '#EAECEF' }}>${trader.initial_balance}</span>
                </div>
                <div>
                  <span style={{ color: '#848E9C' }}>扫描间隔: </span>
                  <span style={{ color: '#EAECEF' }}>{trader.scan_interval_minutes}分钟</span>
                </div>
                <div>
                  <span style={{ color: '#848E9C' }}>状态: </span>
                  <span style={{ color: trader.enabled ? '#0ECB81' : '#848E9C' }}>
                    {trader.enabled ? '✅ 已启用' : '⏸️ 已禁用'}
                  </span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* 提示信息 */}
      <div className="rounded-2xl p-6" style={{
        background: 'linear-gradient(135deg, rgba(240, 185, 11, 0.1) 0%, rgba(252, 213, 53, 0.05) 100%)',
        border: '1px solid rgba(240, 185, 11, 0.2)'
      }}>
        <div className="flex items-start gap-4">
          <div className="text-2xl">⚠️</div>
          <div>
            <h4 className="font-bold mb-2" style={{ color: '#FCD34D' }}>重要提示</h4>
            <ul className="space-y-2 text-sm" style={{ color: '#CBD5E1' }}>
              <li>• 🔄 修改配置后会自动热重载，无需重启服务</li>
              <li>• 🔒 敏感信息（API密钥）已脱敏显示，不修改则保留原值</li>
              <li>• ⚡ 添加/编辑/删除Trader会立即生效</li>
              <li>• 💾 建议修改前先备份config.json文件</li>
              <li>• ⚠️ 如果热重载失败，请手动重启服务</li>
            </ul>
          </div>
        </div>
      </div>

      {/* Trader编辑弹窗 */}
      {editingTrader && (
        <TraderFormModal
          trader={editingTrader}
          isEdit={true}
          onSave={saveTraderConfig}
          onCancel={() => setEditingTrader(null)}
        />
      )}

      {/* 添加Trader弹窗 */}
      {showAddTrader && (
        <TraderFormModal
          trader={traderForm}
          isEdit={false}
          onSave={addTrader}
          onCancel={() => {
            setShowAddTrader(false);
            setTraderForm({});
          }}
          onApplyTemplate={() => setShowTemplates(true)}
        />
      )}

      {/* 模板选择弹窗 */}
      {showTemplates && (
        <div
          className="fixed inset-0 z-[60] flex items-center justify-center p-4"
          style={{ background: 'rgba(0, 0, 0, 0.8)' }}
          onClick={() => setShowTemplates(false)}
        >
          <div
            className="rounded-2xl p-6 max-w-2xl w-full"
            style={{
              background: '#1E2329',
              border: '1px solid rgba(240, 185, 11, 0.3)',
              boxShadow: '0 20px 60px rgba(0, 0, 0, 0.5)'
            }}
            onClick={(e) => e.stopPropagation()}
          >
            <h3 className="text-xl font-bold mb-4" style={{ color: '#EAECEF' }}>
              📋 选择配置模板
            </h3>
            <div className="space-y-3">
              {traderTemplates.map((template, i) => (
                <button
                  key={i}
                  onClick={() => applyTemplate(template.template)}
                  className="w-full p-4 rounded-xl text-left transition-all hover:scale-105"
                  style={{
                    background: '#2B3139',
                    border: '1px solid rgba(99, 102, 241, 0.3)'
                  }}
                >
                  <div className="font-bold mb-1" style={{ color: '#EAECEF' }}>
                    {template.name}
                  </div>
                  <div className="text-sm" style={{ color: '#848E9C' }}>
                    {template.template.exchange?.toUpperCase()} + {template.template.ai_model?.toUpperCase()}
                  </div>
                </button>
              ))}
            </div>
            <button
              onClick={() => setShowTemplates(false)}
              className="w-full mt-4 px-6 py-3 rounded-xl font-bold"
              style={{
                background: 'rgba(248, 113, 113, 0.2)',
                color: '#FCA5A5',
                border: '1px solid rgba(248, 113, 113, 0.3)'
              }}
            >
              取消
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
