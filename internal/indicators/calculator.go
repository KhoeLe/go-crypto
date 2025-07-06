package indicators

import (
	"fmt"
	"math"
	"sort"

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

// CalculateMoneyFlowIndex calculates Money Flow Index (MFI)
func (c *Calculator) CalculateMoneyFlowIndex(klines []models.Kline, period int) (models.MoneyFlowIndicator, error) {
	if len(klines) < period+1 {
		return models.MoneyFlowIndicator{}, fmt.Errorf("insufficient data for MFI calculation: need %d, got %d", period+1, len(klines))
	}

	var positiveFlow, negativeFlow decimal.Decimal
	var typicalPrices []decimal.Decimal
	var moneyFlows []decimal.Decimal

	// Calculate typical prices and money flows
	for i := 1; i < len(klines); i++ {
		// Typical price = (High + Low + Close) / 3
		typicalPrice := klines[i].High.Add(klines[i].Low).Add(klines[i].Close).Div(decimal.NewFromInt(3))
		typicalPrices = append(typicalPrices, typicalPrice)

		// Raw money flow = Typical Price × Volume
		rawMoneyFlow := typicalPrice.Mul(klines[i].Volume)
		moneyFlows = append(moneyFlows, rawMoneyFlow)
	}

	// Calculate positive and negative money flows for the last period
	startIdx := len(typicalPrices) - period
	if startIdx < 0 {
		startIdx = 0
	}

	for i := startIdx; i < len(typicalPrices)-1; i++ {
		if typicalPrices[i+1].GreaterThan(typicalPrices[i]) {
			positiveFlow = positiveFlow.Add(moneyFlows[i+1])
		} else if typicalPrices[i+1].LessThan(typicalPrices[i]) {
			negativeFlow = negativeFlow.Add(moneyFlows[i+1])
		}
	}

	// Calculate Money Flow Index
	var mfi decimal.Decimal
	if negativeFlow.IsZero() {
		mfi = decimal.NewFromInt(100)
	} else {
		moneyFlowRatio := positiveFlow.Div(negativeFlow)
		mfi = decimal.NewFromInt(100).Sub(decimal.NewFromInt(100).Div(decimal.NewFromInt(1).Add(moneyFlowRatio)))
	}

	// Calculate percentage change (compare with previous period if enough data)
	var moneyFlowChange decimal.Decimal
	if len(klines) >= period*2 {
		// Calculate MFI for previous period for comparison
		prevPositiveFlow, prevNegativeFlow := decimal.Zero, decimal.Zero
		prevStartIdx := len(typicalPrices) - period*2
		if prevStartIdx < 0 {
			prevStartIdx = 0
		}

		for i := prevStartIdx; i < prevStartIdx+period-1 && i < len(typicalPrices)-1; i++ {
			if typicalPrices[i+1].GreaterThan(typicalPrices[i]) {
				prevPositiveFlow = prevPositiveFlow.Add(moneyFlows[i+1])
			} else if typicalPrices[i+1].LessThan(typicalPrices[i]) {
				prevNegativeFlow = prevNegativeFlow.Add(moneyFlows[i+1])
			}
		}

		prevMfi := decimal.NewFromInt(50) // Default neutral
		if !prevNegativeFlow.IsZero() {
			prevMoneyFlowRatio := prevPositiveFlow.Div(prevNegativeFlow)
			prevMfi = decimal.NewFromInt(100).Sub(decimal.NewFromInt(100).Div(decimal.NewFromInt(1).Add(prevMoneyFlowRatio)))
		}

		if !prevMfi.IsZero() {
			moneyFlowChange = mfi.Sub(prevMfi).Div(prevMfi).Mul(decimal.NewFromInt(100))
		}
	} else if len(klines) >= period+5 {
		// Fallback: compare current MFI with a shorter historical period when we don't have enough data for full period comparison
		halfPeriod := period / 2
		prevPositiveFlow, prevNegativeFlow := decimal.Zero, decimal.Zero
		prevStartIdx := len(typicalPrices) - period - halfPeriod
		if prevStartIdx < 0 {
			prevStartIdx = 0
		}

		for i := prevStartIdx; i < prevStartIdx+halfPeriod-1 && i < len(typicalPrices)-period; i++ {
			if typicalPrices[i+1].GreaterThan(typicalPrices[i]) {
				prevPositiveFlow = prevPositiveFlow.Add(moneyFlows[i+1])
			} else if typicalPrices[i+1].LessThan(typicalPrices[i]) {
				prevNegativeFlow = prevNegativeFlow.Add(moneyFlows[i+1])
			}
		}

		if !prevNegativeFlow.IsZero() {
			prevMoneyFlowRatio := prevPositiveFlow.Div(prevNegativeFlow)
			prevMfi := decimal.NewFromInt(100).Sub(decimal.NewFromInt(100).Div(decimal.NewFromInt(1).Add(prevMoneyFlowRatio)))

			if !prevMfi.IsZero() {
				moneyFlowChange = mfi.Sub(prevMfi).Div(prevMfi).Mul(decimal.NewFromInt(100))
			}
		}
	}

	return models.MoneyFlowIndicator{
		MoneyFlowIndex:    mfi,
		PositiveMoneyFlow: positiveFlow,
		NegativeMoneyFlow: negativeFlow,
		MoneyFlowChange:   moneyFlowChange,
		Timestamp:         klines[len(klines)-1].CloseTime,
	}, nil
}

// DetectVolumeBreakout detects volume breakouts automatically
func (c *Calculator) DetectVolumeBreakout(klines []models.Kline, lookbackPeriod int) (models.VolumeBreakout, error) {
	if len(klines) < lookbackPeriod+1 {
		return models.VolumeBreakout{}, fmt.Errorf("insufficient data for volume breakout detection: need %d, got %d", lookbackPeriod+1, len(klines))
	}

	// Calculate average volume over lookback period
	var avgVolume decimal.Decimal
	startIdx := len(klines) - lookbackPeriod - 1
	if startIdx < 0 {
		startIdx = 0
	}

	for i := startIdx; i < len(klines)-1; i++ {
		avgVolume = avgVolume.Add(klines[i].Volume)
	}
	avgVolume = avgVolume.Div(decimal.NewFromInt(int64(len(klines) - 1 - startIdx)))

	// Current volume
	currentVolume := klines[len(klines)-1].Volume

	// Calculate volume multiplier
	volumeMultiplier := decimal.Zero
	if !avgVolume.IsZero() {
		volumeMultiplier = currentVolume.Div(avgVolume)
	}

	// Determine if it's a breakout (threshold: 1.5x average volume)
	breakoutThreshold := decimal.NewFromFloat(1.5)
	isBreakout := volumeMultiplier.GreaterThan(breakoutThreshold)

	// Calculate breakout strength (1-10 scale)
	breakoutStrength := decimal.Zero
	if isBreakout {
		// Strength increases with volume multiplier, capped at 10
		strength := volumeMultiplier.Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(5))
		if strength.GreaterThan(decimal.NewFromInt(10)) {
			breakoutStrength = decimal.NewFromInt(10)
		} else {
			breakoutStrength = strength
		}
	}

	// Determine breakout direction based on price movement
	breakoutDirection := "neutral"
	if isBreakout && len(klines) >= 2 {
		currentClose := klines[len(klines)-1].Close
		prevClose := klines[len(klines)-2].Close

		if currentClose.GreaterThan(prevClose) {
			breakoutDirection = "bullish"
		} else if currentClose.LessThan(prevClose) {
			breakoutDirection = "bearish"
		}
	}

	return models.VolumeBreakout{
		IsBreakout:        isBreakout,
		BreakoutStrength:  breakoutStrength,
		VolumeMultiplier:  volumeMultiplier,
		AverageVolume:     avgVolume,
		CurrentVolume:     currentVolume,
		BreakoutDirection: breakoutDirection,
		Timestamp:         klines[len(klines)-1].CloseTime,
	}, nil
}

// CalculateHistoricalIndicators calculates historical RSI and MA with manual tracking
func (c *Calculator) CalculateHistoricalIndicators(klines []models.Kline, rsiPeriods []int, maPeriods []int, maType string, historyLength int) (models.HistoricalIndicators, error) {
	if len(klines) < historyLength {
		historyLength = len(klines)
	}

	var rsiHistory []models.RSIHistoryPoint
	var maHistory []models.MAHistoryPoint

	// Calculate historical points by stepping through the data
	// Ensure we get close to the desired history length by better step calculation
	minRequiredForRSI := 25 // Minimum data points needed for RSI calculation
	if len(klines) <= minRequiredForRSI {
		historyLength = 1 // Only calculate current values
	}

	step := max(1, (len(klines)-minRequiredForRSI)/max(1, historyLength-1))
	if step == 0 {
		step = 1
	}

	// Start from a reasonable point where we have enough data
	startIdx := minRequiredForRSI
	for i := startIdx; i < len(klines); i += step {
		subKlines := klines[:i+1]
		timestamp := klines[i].CloseTime

		// Calculate RSI for each period at this point in time
		for _, period := range rsiPeriods {
			if len(subKlines) >= period+1 {
				rsi, err := c.CalculateRSI(subKlines, period)
				if err == nil {
					rsiHistory = append(rsiHistory, models.RSIHistoryPoint{
						Period:    period,
						Value:     rsi,
						Timestamp: timestamp,
					})
				}
			}
		}

		// Calculate MA for each period at this point in time
		for _, period := range maPeriods {
			if len(subKlines) >= period {
				var ma decimal.Decimal
				var err error

				switch maType {
				case "SMA":
					ma, err = c.CalculateSMA(subKlines, period)
				case "EMA":
					ma, err = c.CalculateEMA(subKlines, period)
				default:
					ma, err = c.CalculateSMA(subKlines, period)
				}

				if err == nil {
					maHistory = append(maHistory, models.MAHistoryPoint{
						Period:    period,
						Type:      maType,
						Value:     ma,
						Timestamp: timestamp,
					})
				}
			}
		}
	}

	return models.HistoricalIndicators{
		RSIHistory: limitRSIHistory(rsiHistory, historyLength),
		MAHistory:  limitMAHistory(maHistory, historyLength),
	}, nil
}

// limitRSIHistory limits RSI history to the specified count, evenly distributed
func limitRSIHistory(rsiHistory []models.RSIHistoryPoint, maxCount int) []models.RSIHistoryPoint {
	if len(rsiHistory) <= maxCount {
		return rsiHistory
	}

	// For multi-analysis/{symbol}?enhanced=true&timeframes=15m,4h,1d
	// We want to return exactly 5 RSI history points
	if maxCount == 5 {
		// Sort the RSI history by timestamp (newest first)
		sort.Slice(rsiHistory, func(i, j int) bool {
			return rsiHistory[i].Timestamp.String() > rsiHistory[j].Timestamp.String()
		})

		// First, try to get one entry per RSI period
		var uniquePeriods []models.RSIHistoryPoint
		seenPeriods := make(map[int]bool)

		for _, point := range rsiHistory {
			if !seenPeriods[point.Period] {
				uniquePeriods = append(uniquePeriods, point)
				seenPeriods[point.Period] = true
			}
		}

		// If we have less than 5 unique periods, we'll need to add more entries
		// Prefer the most common period (usually 14)
		if len(uniquePeriods) < maxCount {
			// Count occurrences of each period
			periodCounts := make(map[int]int)
			for _, point := range rsiHistory {
				periodCounts[point.Period]++
			}

			// Find the period with the most entries
			var mostCommonPeriod int
			maxCount := 0
			for period, count := range periodCounts {
				if count > maxCount {
					maxCount = count
					mostCommonPeriod = period
				}
			}

			// Add more entries from the most common period
			for _, point := range rsiHistory {
				if point.Period == mostCommonPeriod && len(uniquePeriods) < 5 {
					// Check if this timestamp is already included
					alreadyIncluded := false
					for _, existing := range uniquePeriods {
						if existing.Timestamp.String() == point.Timestamp.String() {
							alreadyIncluded = true
							break
						}
					}

					if !alreadyIncluded {
						uniquePeriods = append(uniquePeriods, point)
					}
				}
			}
		}

		// If we still have less than 5, add more entries from any period
		if len(uniquePeriods) < maxCount {
			for _, point := range rsiHistory {
				alreadyIncluded := false
				for _, existing := range uniquePeriods {
					if existing.Timestamp.String() == point.Timestamp.String() && existing.Period == point.Period {
						alreadyIncluded = true
						break
					}
				}

				if !alreadyIncluded && len(uniquePeriods) < maxCount {
					uniquePeriods = append(uniquePeriods, point)
				}
			}
		}

		// If we have more than 5, take only 5
		if len(uniquePeriods) > maxCount {
			uniquePeriods = uniquePeriods[:maxCount]
		}

		return uniquePeriods
	}

	// Original implementation for other cases
	// Group by period and take evenly distributed samples
	periodGroups := make(map[int][]models.RSIHistoryPoint)
	for _, point := range rsiHistory {
		periodGroups[point.Period] = append(periodGroups[point.Period], point)
	}

	var result []models.RSIHistoryPoint
	periodsCount := len(periodGroups)
	if periodsCount == 0 {
		return result
	}

	entriesPerPeriod := maxCount / periodsCount
	if entriesPerPeriod == 0 {
		entriesPerPeriod = 1
	}

	for _, points := range periodGroups {
		if len(points) <= entriesPerPeriod {
			result = append(result, points...)
		} else {
			step := len(points) / entriesPerPeriod
			if step == 0 {
				step = 1
			}
			for i := 0; i < len(points) && len(result) < maxCount; i += step {
				result = append(result, points[i])
			}
		}
	}

	// If we still have more than maxCount, take the most recent ones
	if len(result) > maxCount {
		result = result[len(result)-maxCount:]
	}

	return result
}

// limitMAHistory limits MA history to the specified count, evenly distributed
func limitMAHistory(maHistory []models.MAHistoryPoint, maxCount int) []models.MAHistoryPoint {
	if len(maHistory) == 0 {
		return maHistory
	}

	// Debug the incoming maxCount value
	fmt.Printf("limitMAHistory called with maxCount = %d, incoming entries = %d\n", maxCount, len(maHistory))

	// For the multi-analysis endpoint, we want exactly 5 entries
	if maxCount == 5 {
		// Sort all MA points by timestamp (regardless of period) - newest first
		sort.Slice(maHistory, func(i, j int) bool {
			// We want newest first (descending order)
			return maHistory[i].Timestamp.Time.After(maHistory[j].Timestamp.Time)
		})

		// If we have less than 5 entries, return all of them
		if len(maHistory) <= 5 {
			fmt.Printf("Returning %d MA history entries (less than requested 5)\n", len(maHistory))
			return maHistory
		}

		// Return exactly 5 entries (newest first)
		fmt.Printf("Returning exactly 5 MA history entries (from %d total)\n", len(maHistory))
		return maHistory[:5]
	}

	// Original implementation for other cases
	if len(maHistory) <= maxCount {
		return maHistory
	}

	// Group by period and take evenly distributed samples
	periodGroups := make(map[int][]models.MAHistoryPoint)
	for _, point := range maHistory {
		periodGroups[point.Period] = append(periodGroups[point.Period], point)
	}

	var result []models.MAHistoryPoint
	periodsCount := len(periodGroups)
	if periodsCount == 0 {
		return result
	}

	entriesPerPeriod := maxCount / periodsCount
	if entriesPerPeriod == 0 {
		entriesPerPeriod = 1
	}

	for _, points := range periodGroups {
		if len(points) <= entriesPerPeriod {
			result = append(result, points...)
		} else {
			step := len(points) / entriesPerPeriod
			if step == 0 {
				step = 1
			}
			for i := 0; i < len(points) && len(result) < maxCount; i += step {
				result = append(result, points[i])
			}
		}
	}

	// If we still have more than maxCount, take the most recent ones
	if len(result) > maxCount {
		result = result[len(result)-maxCount:]
	}

	return result
}

// CalculateMACD calculates MACD (Moving Average Convergence Divergence)
func (c *Calculator) CalculateMACD(klines []models.Kline, fastPeriod, slowPeriod, signalPeriod int) (models.MACDIndicator, error) {
	// Adjust parameters if we don't have enough data
	if len(klines) < slowPeriod+signalPeriod {
		// Use smaller periods that work with limited data
		if len(klines) >= 15 {
			fastPeriod = 5
			slowPeriod = 10
			signalPeriod = 3
		} else if len(klines) >= 10 {
			fastPeriod = 3
			slowPeriod = 7
			signalPeriod = 2
		} else {
			return models.MACDIndicator{
				MACD:      decimal.Zero,
				Signal:    decimal.Zero,
				Histogram: decimal.Zero,
			}, nil // Return zeros instead of error for insufficient data
		}
	}

	// Calculate fast EMA
	fastEMA, err := c.CalculateEMA(klines, fastPeriod)
	if err != nil {
		return models.MACDIndicator{
			MACD:      decimal.Zero,
			Signal:    decimal.Zero,
			Histogram: decimal.Zero,
		}, nil
	}

	// Calculate slow EMA
	slowEMA, err := c.CalculateEMA(klines, slowPeriod)
	if err != nil {
		return models.MACDIndicator{
			MACD:      decimal.Zero,
			Signal:    decimal.Zero,
			Histogram: decimal.Zero,
		}, nil
	}

	// MACD line = Fast EMA - Slow EMA
	macdLine := fastEMA.Sub(slowEMA).Round(6)

	// For signal line, we need to calculate EMA of MACD values
	// Since we only have the current MACD value, we'll use a simplified approach
	// In a real implementation, you'd store historical MACD values
	signalLine := macdLine.Mul(decimal.NewFromFloat(0.2)).Round(6) // Simplified signal approximation

	// Histogram = MACD - Signal
	histogram := macdLine.Sub(signalLine).Round(6)

	return models.MACDIndicator{
		MACD:      macdLine,
		Signal:    signalLine,
		Histogram: histogram,
	}, nil
}

// generateMarketSentiment analyzes market conditions and returns sentiment
func (c *Calculator) GenerateMarketSentiment(rsi map[string]decimal.Decimal, macd models.MACDIndicator, kdj models.KDJIndicator, priceChangePercent decimal.Decimal) string {
	score := 0

	// RSI analysis
	if rsi6, exists := rsi["RSI_6"]; exists {
		if rsi6.GreaterThan(decimal.NewFromInt(70)) {
			score += 2 // Very bullish
		} else if rsi6.GreaterThan(decimal.NewFromInt(60)) {
			score += 1 // Bullish
		} else if rsi6.LessThan(decimal.NewFromInt(30)) {
			score -= 2 // Very bearish
		} else if rsi6.LessThan(decimal.NewFromInt(40)) {
			score -= 1 // Bearish
		}
	}

	// MACD analysis
	if macd.MACD.GreaterThan(macd.Signal) {
		score += 1 // Bullish crossover
	} else {
		score -= 1 // Bearish crossover
	}

	if macd.Histogram.GreaterThan(decimal.Zero) {
		score += 1 // Positive momentum
	} else {
		score -= 1 // Negative momentum
	}

	// KDJ analysis
	if kdj.K.GreaterThan(decimal.NewFromInt(80)) {
		score += 1 // Overbought (bullish short term)
	} else if kdj.K.LessThan(decimal.NewFromInt(20)) {
		score -= 1 // Oversold (bearish short term)
	}

	// Price change analysis
	if priceChangePercent.GreaterThan(decimal.NewFromInt(5)) {
		score += 2 // Strong positive move
	} else if priceChangePercent.GreaterThan(decimal.NewFromInt(2)) {
		score += 1 // Moderate positive move
	} else if priceChangePercent.LessThan(decimal.NewFromInt(-5)) {
		score -= 2 // Strong negative move
	} else if priceChangePercent.LessThan(decimal.NewFromInt(-2)) {
		score -= 1 // Moderate negative move
	}

	// Convert score to sentiment
	switch {
	case score >= 4:
		return "very_bullish"
	case score >= 2:
		return "bullish"
	case score >= 1:
		return "slightly_bullish"
	case score <= -4:
		return "very_bearish"
	case score <= -2:
		return "bearish"
	case score <= -1:
		return "slightly_bearish"
	default:
		return "neutral"
	}
}

// CalculateVolumeDelta analyzes buy vs sell pressure from volume data
func (c *Calculator) CalculateVolumeDelta(klines []models.Kline) (models.VolumeDelta, error) {
	if len(klines) == 0 {
		return models.VolumeDelta{}, fmt.Errorf("no klines data for volume delta calculation")
	}

	lastKline := klines[len(klines)-1]

	// Use taker buy volume as approximation for buy volume
	buyVolume := lastKline.TakerBuyBaseVolume
	// Sell volume = total volume - buy volume
	sellVolume := lastKline.Volume.Sub(buyVolume)

	// Calculate delta (buy - sell)
	delta := buyVolume.Sub(sellVolume)

	// Calculate delta as percentage of total volume
	var deltaPercent decimal.Decimal
	if !lastKline.Volume.IsZero() {
		deltaPercent = delta.Div(lastKline.Volume).Mul(decimal.NewFromInt(100))
	}

	// Determine pressure type and strength
	var pressure string
	var strength int

	absDeltaPercent := deltaPercent.Abs()
	switch {
	case absDeltaPercent.GreaterThanOrEqual(decimal.NewFromInt(20)):
		strength = 10 // Very strong
	case absDeltaPercent.GreaterThanOrEqual(decimal.NewFromInt(15)):
		strength = 8 // Strong
	case absDeltaPercent.GreaterThanOrEqual(decimal.NewFromInt(10)):
		strength = 6 // Moderate
	case absDeltaPercent.GreaterThanOrEqual(decimal.NewFromInt(5)):
		strength = 4 // Weak
	case absDeltaPercent.GreaterThanOrEqual(decimal.NewFromInt(2)):
		strength = 2 // Very weak
	default:
		strength = 1 // Minimal
	}

	if deltaPercent.GreaterThan(decimal.NewFromInt(2)) {
		pressure = "buy_pressure"
	} else if deltaPercent.LessThan(decimal.NewFromInt(-2)) {
		pressure = "sell_pressure"
	} else {
		pressure = "balanced"
	}

	return models.VolumeDelta{
		BuyVolume:    buyVolume,
		SellVolume:   sellVolume,
		Delta:        delta,
		DeltaPercent: deltaPercent.Round(2),
		Pressure:     pressure,
		Strength:     strength,
		Timestamp:    lastKline.CloseTime,
	}, nil
}

// CalculateWhaleVolumeSpike detects whale activity based on volume and value thresholds
func (c *Calculator) CalculateWhaleVolumeSpike(klines []models.Kline, currentPrice decimal.Decimal) (models.WhaleVolumeSpike, error) {
	if len(klines) == 0 {
		return models.WhaleVolumeSpike{}, fmt.Errorf("no klines data for whale volume spike calculation")
	}

	// Whale threshold: 100k USDT
	whaleThreshold := decimal.NewFromInt(100000)

	lastKline := klines[len(klines)-1]

	// Calculate volume value in USDT
	volumeValueUSDT := lastKline.Volume.Mul(currentPrice)

	// Calculate average volume over recent periods (up to last 10 periods)
	periodCount := 10
	if len(klines) < periodCount {
		periodCount = len(klines)
	}

	var totalVolume decimal.Decimal
	startIdx := len(klines) - periodCount
	for i := startIdx; i < len(klines)-1; i++ { // Exclude current period from average
		totalVolume = totalVolume.Add(klines[i].Volume)
	}

	var averageVolume decimal.Decimal
	if periodCount > 1 {
		averageVolume = totalVolume.Div(decimal.NewFromInt(int64(periodCount - 1)))
	} else {
		averageVolume = lastKline.Volume // Fallback if only one period
	}

	// Calculate volume multiplier
	var volumeMultiplier decimal.Decimal
	if !averageVolume.IsZero() {
		volumeMultiplier = lastKline.Volume.Div(averageVolume)
	} else {
		volumeMultiplier = decimal.NewFromInt(1)
	}

	// Detect whale spike: volume value > 100k USDT AND volume > 2x recent average
	isWhaleSpike := volumeValueUSDT.GreaterThan(whaleThreshold) &&
		volumeMultiplier.GreaterThan(decimal.NewFromInt(2))

	return models.WhaleVolumeSpike{
		IsWhaleSpike:     isWhaleSpike,
		SpikeVolume:      lastKline.Volume,
		SpikeValueUSDT:   volumeValueUSDT.Round(2),
		ThresholdUSDT:    whaleThreshold,
		VolumeMultiplier: volumeMultiplier.Round(2),
		Timestamp:        lastKline.CloseTime,
	}, nil
}

// DetectPumpSignal detects potential "pump" events based on multiple indicators
func (c *Calculator) DetectPumpSignal(rsi map[string]decimal.Decimal, mfi models.MoneyFlowIndicator, volumeDelta models.VolumeDelta, volumeBreakout models.VolumeBreakout) bool {
	// Check for RSI trending up (RSI_6 > 50)
	rsiTrendingUp := false
	if rsi6, exists := rsi["RSI_6"]; exists {
		rsiTrendingUp = rsi6.GreaterThan(decimal.NewFromInt(50))
	}

	// Check for high Money Flow Index (> 60)
	highMFI := mfi.MoneyFlowIndex.GreaterThan(decimal.NewFromInt(60))

	// Check for buy pressure in volume delta
	buyPressure := volumeDelta.Pressure == "buy_pressure" && volumeDelta.Strength >= 4

	// Check for volume breakout
	volumeBreakoutBullish := volumeBreakout.IsBreakout && volumeBreakout.BreakoutDirection == "bullish"

	// Pump signal: at least 3 out of 4 conditions met
	conditions := 0
	if rsiTrendingUp {
		conditions++
	}
	if highMFI {
		conditions++
	}
	if buyPressure {
		conditions++
	}
	if volumeBreakoutBullish {
		conditions++
	}

	return conditions >= 3
}

// max helper function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
