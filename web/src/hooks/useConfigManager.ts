import { useState, useCallback } from 'react';
import { SystemConfig, TraderConfig } from '../types/config';

export function useConfigManager() {
  const [config, setConfig] = useState<SystemConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  const loadConfig = useCallback(async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/config');
      const data = await response.json();
      if (data.success) {
        setConfig(data.data);
      }
      return data.success;
    } catch (error) {
      console.error('加载配置失败:', error);
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  const updateGlobalConfig = useCallback((updates: Partial<SystemConfig>) => {
    if (config) {
      setConfig({ ...config, ...updates });
    }
  }, [config]);

  const reloadConfig = useCallback(async () => {
    try {
      const response = await fetch('/api/config/reload', {
        method: 'POST',
      });
      const data = await response.json();
      return data.success;
    } catch (error) {
      console.error('热重载失败:', error);
      return false;
    }
  }, []);

  const saveGlobalConfig = useCallback(async () => {
    if (!config) return { success: false, error: '配置为空' };

    try {
      setSaving(true);
      
      const aiLearnInterval = config.ai_learn_interval === undefined || config.ai_learn_interval === 0 
        ? 10 
        : config.ai_learn_interval;

      let marketData = config.market_data;
      if (marketData && marketData.klines) {
        marketData = {
          klines: marketData.klines.map(k => ({
            interval: k.interval,
            limit: k.limit || 20,
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
          ai_autonomy_mode: config.ai_autonomy_mode,
          market_data: marketData,
        }),
      });
      const data = await response.json();
      
      if (data.success) {
        const reloaded = await reloadConfig();
        await loadConfig();
        return { success: true, reloaded };
      } else {
        return { success: false, error: data.error || '未知错误' };
      }
    } catch (error: any) {
      console.error('保存失败:', error);
      return { success: false, error: error.message };
    } finally {
      setSaving(false);
    }
  }, [config, reloadConfig, loadConfig]);

  const saveTrader = useCallback(async (trader: TraderConfig) => {
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
        await loadConfig();
        return { success: true, reloaded };
      } else {
        return { success: false, error: data.error || '未知错误' };
      }
    } catch (error: any) {
      console.error('保存失败:', error);
      return { success: false, error: error.message };
    } finally {
      setSaving(false);
    }
  }, [reloadConfig, loadConfig]);

  const addTrader = useCallback(async (trader: TraderConfig) => {
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
        await loadConfig();
        return { success: true, reloaded };
      } else {
        return { success: false, error: data.error || '未知错误' };
      }
    } catch (error: any) {
      console.error('添加失败:', error);
      return { success: false, error: error.message };
    } finally {
      setSaving(false);
    }
  }, [reloadConfig, loadConfig]);

  const deleteTrader = useCallback(async (traderId: string) => {
    try {
      setSaving(true);
      const response = await fetch(`/api/config/trader/delete?trader_id=${traderId}`, {
        method: 'DELETE',
      });
      const data = await response.json();
      
      if (data.success) {
        const reloaded = await reloadConfig();
        await loadConfig();
        return { success: true, reloaded };
      } else {
        return { success: false, error: data.error || '未知错误' };
      }
    } catch (error: any) {
      console.error('删除失败:', error);
      return { success: false, error: error.message };
    } finally {
      setSaving(false);
    }
  }, [reloadConfig, loadConfig]);

  return {
    config,
    loading,
    saving,
    loadConfig,
    updateGlobalConfig,
    saveGlobalConfig,
    saveTrader,
    addTrader,
    deleteTrader,
  };
}
