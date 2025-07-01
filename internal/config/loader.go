package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	// If no config path provided, try default locations
	if configPath == "" {
		configPath = findConfigFile()
	}

	// If still no config file found, return default config
	if configPath == "" {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge with defaults
	defaultCfg := DefaultConfig()
	mergeCfg := mergeConfigs(defaultCfg, &cfg)

	return mergeCfg, nil
}

// SaveConfig saves configuration to file
func SaveConfig(cfg *Config, configPath string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// findConfigFile tries to find config file in common locations
func findConfigFile() string {
	possiblePaths := []string{
		"configs/config.yaml",
		"configs/config.yml",
		"config.yaml",
		"config.yml",
		"./configs/config.yaml",
		"../configs/config.yaml",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// mergeConfigs merges user config with default config
func mergeConfigs(defaultCfg, userCfg *Config) *Config {
	// Start with default config
	merged := *defaultCfg

	// Override with user config values
	if userCfg.Binance.BaseURL != "" {
		merged.Binance.BaseURL = userCfg.Binance.BaseURL
	}
	if userCfg.Binance.WebSocketURL != "" {
		merged.Binance.WebSocketURL = userCfg.Binance.WebSocketURL
	}
	if userCfg.Binance.APIKey != "" {
		merged.Binance.APIKey = userCfg.Binance.APIKey
	}
	if userCfg.Binance.SecretKey != "" {
		merged.Binance.SecretKey = userCfg.Binance.SecretKey
	}
	if userCfg.Binance.Timeout > 0 {
		merged.Binance.Timeout = userCfg.Binance.Timeout
	}
	if userCfg.Binance.RateLimit > 0 {
		merged.Binance.RateLimit = userCfg.Binance.RateLimit
	}

	if len(userCfg.Symbols) > 0 {
		merged.Symbols = userCfg.Symbols
	}
	if len(userCfg.Intervals) > 0 {
		merged.Intervals = userCfg.Intervals
	}

	// Merge indicator configs
	if len(userCfg.Indicators.RSI.Periods) > 0 {
		merged.Indicators.RSI.Periods = userCfg.Indicators.RSI.Periods
	}
	if len(userCfg.Indicators.MA.Periods) > 0 {
		merged.Indicators.MA.Periods = userCfg.Indicators.MA.Periods
	}
	if userCfg.Indicators.MA.Type != "" {
		merged.Indicators.MA.Type = userCfg.Indicators.MA.Type
	}
	if userCfg.Indicators.KDJ.KPeriod > 0 {
		merged.Indicators.KDJ.KPeriod = userCfg.Indicators.KDJ.KPeriod
	}
	if userCfg.Indicators.KDJ.DPeriod > 0 {
		merged.Indicators.KDJ.DPeriod = userCfg.Indicators.KDJ.DPeriod
	}
	if userCfg.Indicators.KDJ.JPeriod > 0 {
		merged.Indicators.KDJ.JPeriod = userCfg.Indicators.KDJ.JPeriod
	}

	// Merge logging config
	if userCfg.Logging.Level != "" {
		merged.Logging.Level = userCfg.Logging.Level
	}
	if userCfg.Logging.Format != "" {
		merged.Logging.Format = userCfg.Logging.Format
	}
	if userCfg.Logging.Output != "" {
		merged.Logging.Output = userCfg.Logging.Output
	}

	return &merged
}

// ValidateConfig validates configuration values
func ValidateConfig(cfg *Config) error {
	if cfg.Binance.BaseURL == "" {
		return fmt.Errorf("binance base URL is required")
	}
	if cfg.Binance.WebSocketURL == "" {
		return fmt.Errorf("binance WebSocket URL is required")
	}
	if cfg.Binance.Timeout <= 0 {
		return fmt.Errorf("binance timeout must be positive")
	}
	if cfg.Binance.RateLimit <= 0 {
		return fmt.Errorf("binance rate limit must be positive")
	}

	if len(cfg.Symbols) == 0 {
		return fmt.Errorf("at least one symbol is required")
	}
	if len(cfg.Intervals) == 0 {
		return fmt.Errorf("at least one interval is required")
	}

	if len(cfg.Indicators.RSI.Periods) == 0 {
		return fmt.Errorf("at least one RSI period is required")
	}
	for _, period := range cfg.Indicators.RSI.Periods {
		if period <= 0 {
			return fmt.Errorf("RSI period must be positive: %d", period)
		}
	}

	if len(cfg.Indicators.MA.Periods) == 0 {
		return fmt.Errorf("at least one MA period is required")
	}
	for _, period := range cfg.Indicators.MA.Periods {
		if period <= 0 {
			return fmt.Errorf("MA period must be positive: %d", period)
		}
	}
	if cfg.Indicators.KDJ.KPeriod <= 0 {
		return fmt.Errorf("KDJ K period must be positive")
	}
	if cfg.Indicators.KDJ.DPeriod <= 0 {
		return fmt.Errorf("KDJ D period must be positive")
	}
	if cfg.Indicators.KDJ.JPeriod <= 0 {
		return fmt.Errorf("KDJ J period must be positive")
	}

	validMATypes := map[string]bool{"SMA": true, "EMA": true, "WMA": true}
	if !validMATypes[cfg.Indicators.MA.Type] {
		return fmt.Errorf("invalid MA type: %s", cfg.Indicators.MA.Type)
	}

	return nil
}
