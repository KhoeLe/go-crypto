package utils

import (
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

// SafeParseDecimal safely parses a string to decimal with error handling
func SafeParseDecimal(s string) (decimal.Decimal, error) {
	if s == "" {
		return decimal.Zero, fmt.Errorf("empty string provided")
	}

	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to parse decimal from '%s': %w", s, err)
	}

	return d, nil
}

// SafeParseFloat64 safely parses interface{} to float64
func SafeParseFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case string:
		return strconv.ParseFloat(val, 64)
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	default:
		return 0, fmt.Errorf("unsupported type: %T", v)
	}
}

// SafeParseInt64 safely parses interface{} to int64
func SafeParseInt64(v interface{}) (int64, error) {
	switch val := v.(type) {
	case float64:
		return int64(val), nil
	case string:
		return strconv.ParseInt(val, 10, 64)
	case int:
		return int64(val), nil
	case int64:
		return val, nil
	default:
		return 0, fmt.Errorf("unsupported type: %T", v)
	}
}

// FormatTimestamp formats timestamp for display
func FormatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatDecimal formats decimal for display with specified precision
func FormatDecimal(d decimal.Decimal, precision int32) string {
	return d.StringFixed(precision)
}

// FormatPercentage formats decimal as percentage
func FormatPercentage(d decimal.Decimal) string {
	return d.StringFixed(2) + "%"
}

// CalculatePercentageChange calculates percentage change between two values
func CalculatePercentageChange(oldValue, newValue decimal.Decimal) decimal.Decimal {
	if oldValue.IsZero() {
		return decimal.Zero
	}

	change := newValue.Sub(oldValue)
	return change.Div(oldValue).Mul(decimal.NewFromInt(100))
}

// RoundToSignificantFigures rounds decimal to specified significant figures
func RoundToSignificantFigures(d decimal.Decimal, figures int) decimal.Decimal {
	if d.IsZero() {
		return decimal.Zero
	}

	// Convert to string and back to handle significant figures
	str := d.String()
	if len(str) > figures {
		// This is a simplified approach - a more sophisticated implementation
		// would handle scientific notation and edge cases
		return d.Round(int32(figures - 1))
	}

	return d
}

// ValidateSymbol validates if symbol follows a valid format
// Let Binance API handle the actual symbol validation for better accuracy
func ValidateSymbol(symbol string) bool {
	// Basic format validation
	if len(symbol) < 6 || len(symbol) > 20 {
		return false
	}
	
	// Check if symbol contains only uppercase letters and numbers
	for _, char := range symbol {
		if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return false
		}
	}
	
	// Must end with a valid quote currency (basic check)
	validQuotes := []string{"USDT", "USDC", "BTC", "ETH", "BNB", "BUSD", "FDUSD"}
	for _, quote := range validQuotes {
		if len(symbol) > len(quote) && symbol[len(symbol)-len(quote):] == quote {
			return true
		}
	}
	
	return false
}

// ValidateTimeframe validates if timeframe is supported
func ValidateTimeframe(timeframe string) bool {
	supportedTimeframes := map[string]bool{
		"1m":  true,
		"3m":  true,
		"5m":  true,
		"15m": true,
		"30m": true,
		"1h":  true,
		"2h":  true,
		"4h":  true,
		"6h":  true,
		"8h":  true,
		"12h": true,
		"1d":  true,
		"3d":  true,
		"1w":  true,
		"1M":  true,
	}

	return supportedTimeframes[timeframe]
}

// GetTimeframeInSeconds returns timeframe duration in seconds
func GetTimeframeInSeconds(timeframe string) int64 {
	timeframeSeconds := map[string]int64{
		"1m":  60,
		"3m":  180,
		"5m":  300,
		"15m": 900,
		"30m": 1800,
		"1h":  3600,
		"2h":  7200,
		"4h":  14400,
		"6h":  21600,
		"8h":  28800,
		"12h": 43200,
		"1d":  86400,
		"3d":  259200,
		"1w":  604800,
		"1M":  2592000, // Approximate
	}

	if seconds, exists := timeframeSeconds[timeframe]; exists {
		return seconds
	}

	return 3600 // Default to 1 hour
}
