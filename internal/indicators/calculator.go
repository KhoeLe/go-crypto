package indicators

import (
	"fmt"
	"math"

	"go-crypto/internal/models"

	"github.com/shopspring/decimal"
)

// Calculator provides methods for calculating technical indicators
type Calculator struct{}

// NewCalculator creates a new indicator calculator
func NewCalculator() *Calculator {
	return &Calculator{}
}

// CalculateRSI calculates Relative Strength Index
func (c *Calculator) CalculateRSI(klines []models.Kline, period int) (decimal.Decimal, error) {
	if len(klines) < period+1 {
		return decimal.Zero, fmt.Errorf("insufficient data for RSI calculation: need %d, got %d", period+1, len(klines))
	}

	// Calculate price changes
	var gains, losses []decimal.Decimal
	for i := 1; i < len(klines); i++ {
		change := klines[i].Close.Sub(klines[i-1].Close)
		if change.GreaterThan(decimal.Zero) {
			gains = append(gains, change)
			losses = append(losses, decimal.Zero)
		} else {
			gains = append(gains, decimal.Zero)
			losses = append(losses, change.Abs())
		}
	}

	if len(gains) < period {
		return decimal.Zero, fmt.Errorf("insufficient price changes for RSI: need %d, got %d", period, len(gains))
	}

	// Calculate initial averages
	var avgGain, avgLoss decimal.Decimal
	for i := 0; i < period; i++ {
		avgGain = avgGain.Add(gains[i])
		avgLoss = avgLoss.Add(losses[i])
	}
	avgGain = avgGain.Div(decimal.NewFromInt(int64(period)))
	avgLoss = avgLoss.Div(decimal.NewFromInt(int64(period)))

	// Apply smoothing for remaining periods
	for i := period; i < len(gains); i++ {
		avgGain = avgGain.Mul(decimal.NewFromInt(int64(period - 1))).Add(gains[i]).Div(decimal.NewFromInt(int64(period)))
		avgLoss = avgLoss.Mul(decimal.NewFromInt(int64(period - 1))).Add(losses[i]).Div(decimal.NewFromInt(int64(period)))
	}

	if avgLoss.IsZero() {
		return decimal.NewFromInt(100), nil
	}

	rs := avgGain.Div(avgLoss)
	rsi := decimal.NewFromInt(100).Sub(decimal.NewFromInt(100).Div(decimal.NewFromInt(1).Add(rs)))

	return rsi, nil
}

// CalculateSMA calculates Simple Moving Average
func (c *Calculator) CalculateSMA(klines []models.Kline, period int) (decimal.Decimal, error) {
	if len(klines) < period {
		return decimal.Zero, fmt.Errorf("insufficient data for SMA calculation: need %d, got %d", period, len(klines))
	}

	var sum decimal.Decimal
	for i := len(klines) - period; i < len(klines); i++ {
		sum = sum.Add(klines[i].Close)
	}

	return sum.Div(decimal.NewFromInt(int64(period))), nil
}

// CalculateEMA calculates Exponential Moving Average
func (c *Calculator) CalculateEMA(klines []models.Kline, period int) (decimal.Decimal, error) {
	if len(klines) < period {
		return decimal.Zero, fmt.Errorf("insufficient data for EMA calculation: need %d, got %d", period, len(klines))
	}

	// Calculate multiplier
	multiplier := decimal.NewFromFloat(2.0).Div(decimal.NewFromInt(int64(period + 1)))

	// Start with SMA for the first EMA value
	var sum decimal.Decimal
	for i := 0; i < period; i++ {
		sum = sum.Add(klines[i].Close)
	}
	ema := sum.Div(decimal.NewFromInt(int64(period)))

	// Calculate EMA for remaining periods
	for i := period; i < len(klines); i++ {
		ema = klines[i].Close.Sub(ema).Mul(multiplier).Add(ema)
	}

	return ema, nil
}

// CalculateKDJ calculates KDJ stochastic oscillator
func (c *Calculator) CalculateKDJ(klines []models.Kline, kPeriod, dPeriod, jPeriod int) (models.KDJIndicator, error) {
	if len(klines) < kPeriod {
		return models.KDJIndicator{}, fmt.Errorf("insufficient data for KDJ calculation: need %d, got %d", kPeriod, len(klines))
	}

	// Calculate %K values
	var kValues []decimal.Decimal
	for i := kPeriod - 1; i < len(klines); i++ {
		// Find highest high and lowest low in the period
		high := klines[i-kPeriod+1].High
		low := klines[i-kPeriod+1].Low

		for j := i - kPeriod + 2; j <= i; j++ {
			if klines[j].High.GreaterThan(high) {
				high = klines[j].High
			}
			if klines[j].Low.LessThan(low) {
				low = klines[j].Low
			}
		}

		// Calculate %K
		if high.Equal(low) {
			kValues = append(kValues, decimal.NewFromInt(50))
		} else {
			k := klines[i].Close.Sub(low).Div(high.Sub(low)).Mul(decimal.NewFromInt(100))
			kValues = append(kValues, k)
		}
	}

	if len(kValues) == 0 {
		return models.KDJIndicator{}, fmt.Errorf("no K values calculated")
	}

	// Calculate %D (moving average of %K)
	var d decimal.Decimal
	if len(kValues) >= dPeriod {
		var sum decimal.Decimal
		for i := len(kValues) - dPeriod; i < len(kValues); i++ {
			sum = sum.Add(kValues[i])
		}
		d = sum.Div(decimal.NewFromInt(int64(dPeriod)))
	} else {
		// If not enough data for D period, use average of all available K values
		var sum decimal.Decimal
		for _, k := range kValues {
			sum = sum.Add(k)
		}
		d = sum.Div(decimal.NewFromInt(int64(len(kValues))))
	}

	// Get latest %K
	k := kValues[len(kValues)-1]

	// Calculate %J
	j := k.Mul(decimal.NewFromInt(int64(jPeriod))).Sub(d.Mul(decimal.NewFromInt(int64(jPeriod - 1))))

	return models.KDJIndicator{
		K: k,
		D: d,
		J: j,
	}, nil
}

// CalculateAllIndicators calculates all technical indicators for given klines
func (c *Calculator) CalculateAllIndicators(klines []models.Kline, rsiPeriods []int, maPeriods []int, maType string, kPeriod, dPeriod, jPeriod int) (*models.TechnicalIndicators, error) {
	if len(klines) == 0 {
		return nil, fmt.Errorf("no kline data provided")
	}

	lastKline := klines[len(klines)-1]

	// Calculate RSI for multiple periods
	rsiValues := make(map[int]decimal.Decimal)
	for _, period := range rsiPeriods {
		rsi, err := c.CalculateRSI(klines, period)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate RSI for period %d: %w", period, err)
		}
		rsiValues[period] = rsi
	}

	// Calculate Moving Average for multiple periods
	maValues := make(map[int]decimal.Decimal)
	for _, period := range maPeriods {
		var ma decimal.Decimal
		var err error
		switch maType {
		case "SMA":
			ma, err = c.CalculateSMA(klines, period)
		case "EMA":
			ma, err = c.CalculateEMA(klines, period)
		default:
			ma, err = c.CalculateSMA(klines, period)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to calculate MA for period %d: %w", period, err)
		}
		maValues[period] = ma
	}

	// Calculate KDJ
	kdj, err := c.CalculateKDJ(klines, kPeriod, dPeriod, jPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate KDJ: %w", err)
	}

	return &models.TechnicalIndicators{
		Symbol:    lastKline.Symbol,
		Timeframe: lastKline.Timeframe,
		Timestamp: lastKline.CloseTime,
		RSI:       rsiValues,
		MA:        maValues,
		KDJ:       kdj,
	}, nil
}

// CalculateVolatility calculates price volatility (standard deviation)
func (c *Calculator) CalculateVolatility(klines []models.Kline, period int) (decimal.Decimal, error) {
	if len(klines) < period {
		return decimal.Zero, fmt.Errorf("insufficient data for volatility calculation: need %d, got %d", period, len(klines))
	}

	// Calculate mean
	var sum decimal.Decimal
	for i := len(klines) - period; i < len(klines); i++ {
		sum = sum.Add(klines[i].Close)
	}
	mean := sum.Div(decimal.NewFromInt(int64(period)))

	// Calculate variance
	var variance decimal.Decimal
	for i := len(klines) - period; i < len(klines); i++ {
		diff := klines[i].Close.Sub(mean)
		variance = variance.Add(diff.Mul(diff))
	}
	variance = variance.Div(decimal.NewFromInt(int64(period)))

	// Calculate standard deviation
	varianceFloat, _ := variance.Float64()
	stdDev := decimal.NewFromFloat(math.Sqrt(varianceFloat))

	return stdDev, nil
}
