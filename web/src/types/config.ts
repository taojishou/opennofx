// 配置相关的类型定义

export interface TraderConfig {
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
  ai_autonomy_mode?: boolean;
  compact_mode?: boolean;
}

export interface KlineConfig {
  interval: string;
  limit: number;
  show_table: boolean;
}

export interface MarketDataConfig {
  klines: KlineConfig[];
}

export interface SystemConfig {
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
  ai_autonomy_mode?: boolean;
  market_data?: MarketDataConfig;
}

export interface PromptSection {
  id: number;
  section_name: string;
  title: string;
  content: string;
  prompt_type: 'system' | 'user'; // 添加prompt_type字段
  enabled: boolean;
  display_order: number;
  updated_at: string;
}

export interface RuntimeConfigItem {
  key: string;
  value: string;
  description: string;
  type: string;
  updated_at: string;
}

export interface ConfigGroup {
  [key: string]: RuntimeConfigItem[];
}

export interface TraderTemplate {
  name: string;
  template: Partial<TraderConfig>;
}
