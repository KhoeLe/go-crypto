package indicators

import (
	"testing"

	"go-crypto/internal/models"

	"github.com/shopspring/decimal"
)

func TestCalculateRSI(t *testing.T) {
	calc := NewCalculator()

	// Create test data
	klines := createTestKlines(20)

	rsi, err := calc.CalculateRSI(klines, 14)
	if err != nil {
		t.Fatalf("Failed to calculate RSI: %v", err)
	}

	// RSI should be between 0 and 100
	if rsi.LessThan(decimal.Zero) || rsi.GreaterThan(decimal.NewFromInt(100)) {
		t.Errorf("RSI value out of range: %s", rsi.String())
	}
}

func TestCalculateSMA(t *testing.T) {
	calc := NewCalculator()

	// Create test data with known values
	klines := []models.Kline{
		{Close: decimal.NewFromInt(10)},
		{Close: decimal.NewFromInt(20)},
		{Close: decimal.NewFromInt(30)},
	}

	sma, err := calc.CalculateSMA(klines, 3)
	if err != nil {
		t.Fatalf("Failed to calculate SMA: %v", err)
	}

	expected := decimal.NewFromInt(20) // (10 + 20 + 30) / 3
	if !sma.Equal(expected) {
		t.Errorf("Expected SMA %s, got %s", expected.String(), sma.String())
	}
}

func TestCalculateKDJ(t *testing.T) {
	calc := NewCalculator()

	// Create test data
	klines := createTestKlines(15)

	kdj, err := calc.CalculateKDJ(klines, 9, 3, 3)
	if err != nil {
		t.Fatalf("Failed to calculate KDJ: %v", err)
	}

	// KDJ values should be within reasonable ranges
	if kdj.K.LessThan(decimal.Zero) || kdj.K.GreaterThan(decimal.NewFromInt(100)) {
		t.Errorf("K value out of range: %s", kdj.K.String())
	}
}

func TestCalculateAllIndicators(t *testing.T) {
	calc := NewCalculator()

	// Create test data
	klines := createTestKlines(50)

	indicators, err := calc.CalculateAllIndicators(klines, []int{14}, []int{20}, "SMA", 9, 3, 3)
	if err != nil {
		t.Fatalf("Failed to calculate all indicators: %v", err)
	}

	// Verify all indicators are calculated
	if len(indicators.RSI) == 0 {
		t.Error("RSI should be calculated")
	}
	if len(indicators.MA) == 0 {
		t.Error("MA should be calculated")
	}
	if indicators.KDJ.K.IsZero() && indicators.KDJ.D.IsZero() {
		t.Error("KDJ values should not all be zero")
	}
}

// Helper function to create test klines
func createTestKlines(count int) []models.Kline {
	klines := make([]models.Kline, count)

	for i := 0; i < count; i++ {
		base := decimal.NewFromInt(int64(100 + i))
		klines[i] = models.Kline{
			Symbol:    models.BTCUSDT,
			Open:      base,
			High:      base.Add(decimal.NewFromInt(5)),
			Low:       base.Sub(decimal.NewFromInt(3)),
			Close:     base.Add(decimal.NewFromInt(2)),
			Volume:    decimal.NewFromInt(1000),
			Timeframe: models.Timeframe15m,
		}
	}

	return klines
}
