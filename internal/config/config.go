package config

import (
	"go-crypto/internal/models"
)

// Config represents application configuration
type Config struct {
	Binance    BinanceConfig   `yaml:"binance"`
	Symbols    []string        `yaml:"symbols"`
	Intervals  []string        `yaml:"intervals"`
	Indicators IndicatorConfig `yaml:"indicators"`
	Logging    LoggingConfig   `yaml:"logging"`
	RateLimit  RateLimitConfig `yaml:"rate_limit"`
}

// BinanceConfig represents Binance API configuration
type BinanceConfig struct {
	BaseURL      string `yaml:"base_url"`
	WebSocketURL string `yaml:"websocket_url"`
	APIKey       string `yaml:"api_key,omitempty"`
	SecretKey    string `yaml:"secret_key,omitempty"`
	Timeout      int    `yaml:"timeout"`
	RateLimit    int    `yaml:"rate_limit"`
}

// IndicatorConfig represents technical indicator parameters
type IndicatorConfig struct {
	RSI RSIConfig `yaml:"rsi"`
	MA  MAConfig  `yaml:"ma"`
	KDJ KDJConfig `yaml:"kdj"`
}

// RSIConfig represents RSI indicator configuration
type RSIConfig struct {
	Periods []int `yaml:"periods"`
}

// MAConfig represents Moving Average configuration
type MAConfig struct {
	Periods []int  `yaml:"periods"`
	Type    string `yaml:"type"` // SMA, EMA, WMA
}

// KDJConfig represents KDJ indicator configuration
type KDJConfig struct {
	KPeriod int `yaml:"k_period"`
	DPeriod int `yaml:"d_period"`
	JPeriod int `yaml:"j_period"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled         bool                     `yaml:"enabled"`
	DefaultTier     string                   `yaml:"default_tier"`
	Tiers           map[string]RateLimitTier `yaml:"tiers"`
	IPWhitelist     []string                 `yaml:"ip_whitelist"`
	IPBlacklist     []string                 `yaml:"ip_blacklist"`
	CleanupInterval int                      `yaml:"cleanup_interval"` // minutes
}

// RateLimitTier represents a rate limiting tier with different limits
type RateLimitTier struct {
	RequestsPerMinute    int `yaml:"requests_per_minute"`
	RequestsPerHour      int `yaml:"requests_per_hour"`
	BurstAllowance       int `yaml:"burst_allowance"`
	BlockDurationMinutes int `yaml:"block_duration_minutes"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Binance: BinanceConfig{
			BaseURL:      "https://api.binance.com",
			WebSocketURL: "wss://stream.binance.com:9443/ws",
			Timeout:      30,
			RateLimit:    1200, // requests per minute
		},
		Symbols: []string{
			string(models.BTCUSDT),
			string(models.ETHUSDT),
			string(models.BNBUSDT),
			string(models.ETHFIUSDT),
		},
		Intervals: []string{
			string(models.Timeframe15m),
			string(models.Timeframe4h),
			string(models.Timeframe1d),
		},
		Indicators: IndicatorConfig{
			RSI: RSIConfig{
				Periods: []int{6, 12, 24},
			},
			MA: MAConfig{
				Periods: []int{7, 25, 99},
				Type:    "SMA",
			},
			KDJ: KDJConfig{
				KPeriod: 9,
				DPeriod: 3,
				JPeriod: 3,
			},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		RateLimit: RateLimitConfig{
			Enabled:         true,
			DefaultTier:     "premium",
			CleanupInterval: 60, // cleanup old entries every hour
			IPWhitelist:     []string{},
			IPBlacklist:     []string{},
			Tiers: map[string]RateLimitTier{
				"basic": {
					RequestsPerMinute:    10,
					RequestsPerHour:      200,
					BurstAllowance:       5,
					BlockDurationMinutes: 15,
				},
				"standard": {
					RequestsPerMinute:    60,
					RequestsPerHour:      1000,
					BurstAllowance:       20,
					BlockDurationMinutes: 5,
				},
				"premium": {
					RequestsPerMinute:    300,
					RequestsPerHour:      5000,
					BurstAllowance:       50,
					BlockDurationMinutes: 1,
				},
				"unlimited": {
					RequestsPerMinute:    0, // 0 means unlimited
					RequestsPerHour:      0,
					BurstAllowance:       0,
					BlockDurationMinutes: 0,
				},
			},
		},
	}
}
