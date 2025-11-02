package repositories

import (
	"database/sql"
	"nofx/database/models"
)

// TraderConfigRepository 交易员配置数据访问层
type TraderConfigRepository struct {
	db *sql.DB
}

// NewTraderConfigRepository 创建交易员配置仓储
func NewTraderConfigRepository(db *sql.DB) *TraderConfigRepository {
	return &TraderConfigRepository{db: db}
}

// Create 创建交易员配置
func (r *TraderConfigRepository) Create(config *models.TraderConfig) (int64, error) {
	query := `
		INSERT INTO trader_configs (
			user_id, trader_id, name, enabled, ai_model, exchange,
			binance_api_key, binance_secret_key,
			hyperliquid_private_key, hyperliquid_wallet_addr, hyperliquid_testnet,
			aster_user, aster_signer, aster_private_key,
			deepseek_key, qwen_key, custom_api_url, custom_api_key, custom_model_name,
			initial_balance, scan_interval_minutes, max_positions,
			btc_eth_leverage, altcoin_leverage,
			max_daily_loss, max_drawdown, stop_trading_minutes,
			enable_ai_learning, ai_learn_interval, ai_autonomy_mode, compact_mode
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query,
		config.UserID, config.TraderID, config.Name, config.Enabled, config.AIModel, config.Exchange,
		config.BinanceAPIKey, config.BinanceSecretKey,
		config.HyperliquidPrivateKey, config.HyperliquidWalletAddr, config.HyperliquidTestnet,
		config.AsterUser, config.AsterSigner, config.AsterPrivateKey,
		config.DeepSeekKey, config.QwenKey, config.CustomAPIURL, config.CustomAPIKey, config.CustomModelName,
		config.InitialBalance, config.ScanIntervalMinutes, config.MaxPositions,
		config.BTCETHLeverage, config.AltcoinLeverage,
		config.MaxDailyLoss, config.MaxDrawdown, config.StopTradingMinutes,
		config.EnableAILearning, config.AILearnInterval, config.AIAutonomyMode, config.CompactMode,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetByID 根据ID获取配置
func (r *TraderConfigRepository) GetByID(id int64) (*models.TraderConfig, error) {
	query := `
		SELECT id, user_id, trader_id, name, enabled, ai_model, exchange,
			binance_api_key, binance_secret_key,
			hyperliquid_private_key, hyperliquid_wallet_addr, hyperliquid_testnet,
			aster_user, aster_signer, aster_private_key,
			deepseek_key, qwen_key, custom_api_url, custom_api_key, custom_model_name,
			initial_balance, scan_interval_minutes, max_positions,
			btc_eth_leverage, altcoin_leverage,
			max_daily_loss, max_drawdown, stop_trading_minutes,
			enable_ai_learning, ai_learn_interval, ai_autonomy_mode, compact_mode,
			created_at, updated_at
		FROM trader_configs WHERE id = ?
	`
	config := &models.TraderConfig{}
	err := r.db.QueryRow(query, id).Scan(
		&config.ID, &config.UserID, &config.TraderID, &config.Name, &config.Enabled, &config.AIModel, &config.Exchange,
		&config.BinanceAPIKey, &config.BinanceSecretKey,
		&config.HyperliquidPrivateKey, &config.HyperliquidWalletAddr, &config.HyperliquidTestnet,
		&config.AsterUser, &config.AsterSigner, &config.AsterPrivateKey,
		&config.DeepSeekKey, &config.QwenKey, &config.CustomAPIURL, &config.CustomAPIKey, &config.CustomModelName,
		&config.InitialBalance, &config.ScanIntervalMinutes, &config.MaxPositions,
		&config.BTCETHLeverage, &config.AltcoinLeverage,
		&config.MaxDailyLoss, &config.MaxDrawdown, &config.StopTradingMinutes,
		&config.EnableAILearning, &config.AILearnInterval, &config.AIAutonomyMode, &config.CompactMode,
		&config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// GetByTraderID 根据TraderID获取配置
func (r *TraderConfigRepository) GetByTraderID(traderID string) (*models.TraderConfig, error) {
	query := `
		SELECT id, user_id, trader_id, name, enabled, ai_model, exchange,
			binance_api_key, binance_secret_key,
			hyperliquid_private_key, hyperliquid_wallet_addr, hyperliquid_testnet,
			aster_user, aster_signer, aster_private_key,
			deepseek_key, qwen_key, custom_api_url, custom_api_key, custom_model_name,
			initial_balance, scan_interval_minutes, max_positions,
			btc_eth_leverage, altcoin_leverage,
			max_daily_loss, max_drawdown, stop_trading_minutes,
			enable_ai_learning, ai_learn_interval, ai_autonomy_mode, compact_mode,
			created_at, updated_at
		FROM trader_configs WHERE trader_id = ?
	`
	config := &models.TraderConfig{}
	err := r.db.QueryRow(query, traderID).Scan(
		&config.ID, &config.UserID, &config.TraderID, &config.Name, &config.Enabled, &config.AIModel, &config.Exchange,
		&config.BinanceAPIKey, &config.BinanceSecretKey,
		&config.HyperliquidPrivateKey, &config.HyperliquidWalletAddr, &config.HyperliquidTestnet,
		&config.AsterUser, &config.AsterSigner, &config.AsterPrivateKey,
		&config.DeepSeekKey, &config.QwenKey, &config.CustomAPIURL, &config.CustomAPIKey, &config.CustomModelName,
		&config.InitialBalance, &config.ScanIntervalMinutes, &config.MaxPositions,
		&config.BTCETHLeverage, &config.AltcoinLeverage,
		&config.MaxDailyLoss, &config.MaxDrawdown, &config.StopTradingMinutes,
		&config.EnableAILearning, &config.AILearnInterval, &config.AIAutonomyMode, &config.CompactMode,
		&config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// GetByUserID 获取用户的所有交易员配置
func (r *TraderConfigRepository) GetByUserID(userID int64) ([]*models.TraderConfig, error) {
	query := `
		SELECT id, user_id, trader_id, name, enabled, ai_model, exchange,
			binance_api_key, binance_secret_key,
			hyperliquid_private_key, hyperliquid_wallet_addr, hyperliquid_testnet,
			aster_user, aster_signer, aster_private_key,
			deepseek_key, qwen_key, custom_api_url, custom_api_key, custom_model_name,
			initial_balance, scan_interval_minutes, max_positions,
			btc_eth_leverage, altcoin_leverage,
			max_daily_loss, max_drawdown, stop_trading_minutes,
			enable_ai_learning, ai_learn_interval, ai_autonomy_mode, compact_mode,
			created_at, updated_at
		FROM trader_configs WHERE user_id = ?
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*models.TraderConfig
	for rows.Next() {
		config := &models.TraderConfig{}
		err := rows.Scan(
			&config.ID, &config.UserID, &config.TraderID, &config.Name, &config.Enabled, &config.AIModel, &config.Exchange,
			&config.BinanceAPIKey, &config.BinanceSecretKey,
			&config.HyperliquidPrivateKey, &config.HyperliquidWalletAddr, &config.HyperliquidTestnet,
			&config.AsterUser, &config.AsterSigner, &config.AsterPrivateKey,
			&config.DeepSeekKey, &config.QwenKey, &config.CustomAPIURL, &config.CustomAPIKey, &config.CustomModelName,
			&config.InitialBalance, &config.ScanIntervalMinutes, &config.MaxPositions,
			&config.BTCETHLeverage, &config.AltcoinLeverage,
			&config.MaxDailyLoss, &config.MaxDrawdown, &config.StopTradingMinutes,
			&config.EnableAILearning, &config.AILearnInterval, &config.AIAutonomyMode, &config.CompactMode,
			&config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			continue
		}
		configs = append(configs, config)
	}
	return configs, nil
}

// GetAllEnabled 获取所有启用的配置
func (r *TraderConfigRepository) GetAllEnabled() ([]*models.TraderConfig, error) {
	query := `
		SELECT id, user_id, trader_id, name, enabled, ai_model, exchange,
			binance_api_key, binance_secret_key,
			hyperliquid_private_key, hyperliquid_wallet_addr, hyperliquid_testnet,
			aster_user, aster_signer, aster_private_key,
			deepseek_key, qwen_key, custom_api_url, custom_api_key, custom_model_name,
			initial_balance, scan_interval_minutes, max_positions,
			btc_eth_leverage, altcoin_leverage,
			max_daily_loss, max_drawdown, stop_trading_minutes,
			enable_ai_learning, ai_learn_interval, ai_autonomy_mode, compact_mode,
			created_at, updated_at
		FROM trader_configs WHERE enabled = 1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*models.TraderConfig
	for rows.Next() {
		config := &models.TraderConfig{}
		err := rows.Scan(
			&config.ID, &config.UserID, &config.TraderID, &config.Name, &config.Enabled, &config.AIModel, &config.Exchange,
			&config.BinanceAPIKey, &config.BinanceSecretKey,
			&config.HyperliquidPrivateKey, &config.HyperliquidWalletAddr, &config.HyperliquidTestnet,
			&config.AsterUser, &config.AsterSigner, &config.AsterPrivateKey,
			&config.DeepSeekKey, &config.QwenKey, &config.CustomAPIURL, &config.CustomAPIKey, &config.CustomModelName,
			&config.InitialBalance, &config.ScanIntervalMinutes, &config.MaxPositions,
			&config.BTCETHLeverage, &config.AltcoinLeverage,
			&config.MaxDailyLoss, &config.MaxDrawdown, &config.StopTradingMinutes,
			&config.EnableAILearning, &config.AILearnInterval, &config.AIAutonomyMode, &config.CompactMode,
			&config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			continue
		}
		configs = append(configs, config)
	}
	return configs, nil
}

// Update 更新交易员配置
func (r *TraderConfigRepository) Update(config *models.TraderConfig) error {
	query := `
		UPDATE trader_configs SET
			name = ?, enabled = ?, ai_model = ?, exchange = ?,
			binance_api_key = ?, binance_secret_key = ?,
			hyperliquid_private_key = ?, hyperliquid_wallet_addr = ?, hyperliquid_testnet = ?,
			aster_user = ?, aster_signer = ?, aster_private_key = ?,
			deepseek_key = ?, qwen_key = ?, custom_api_url = ?, custom_api_key = ?, custom_model_name = ?,
			initial_balance = ?, scan_interval_minutes = ?, max_positions = ?,
			btc_eth_leverage = ?, altcoin_leverage = ?,
			max_daily_loss = ?, max_drawdown = ?, stop_trading_minutes = ?,
			enable_ai_learning = ?, ai_learn_interval = ?, ai_autonomy_mode = ?, compact_mode = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.Exec(query,
		config.Name, config.Enabled, config.AIModel, config.Exchange,
		config.BinanceAPIKey, config.BinanceSecretKey,
		config.HyperliquidPrivateKey, config.HyperliquidWalletAddr, config.HyperliquidTestnet,
		config.AsterUser, config.AsterSigner, config.AsterPrivateKey,
		config.DeepSeekKey, config.QwenKey, config.CustomAPIURL, config.CustomAPIKey, config.CustomModelName,
		config.InitialBalance, config.ScanIntervalMinutes, config.MaxPositions,
		config.BTCETHLeverage, config.AltcoinLeverage,
		config.MaxDailyLoss, config.MaxDrawdown, config.StopTradingMinutes,
		config.EnableAILearning, config.AILearnInterval, config.AIAutonomyMode, &config.CompactMode,
		config.ID,
	)
	return err
}

// Delete 删除交易员配置
func (r *TraderConfigRepository) Delete(id int64) error {
	query := `DELETE FROM trader_configs WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}
