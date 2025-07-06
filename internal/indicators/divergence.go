package indicators

import (
	"fmt"
	"go-crypto/internal/models"

	"github.com/shopspring/decimal"
)

// Define constants for divergence types
const (
	PositiveDivergence = "bullish"
	NegativeDivergence = "bearish"
	PartialDivergence  = "partial" // New constant for partial divergence
)

// DetectMoneyFlowDivergence detects divergence between RSI and Money Flow Index
// A divergence occurs when price action and indicator movement don't agree
func (c *Calculator) DetectMoneyFlowDivergence(
	klines []models.Kline,
	rsiHistory []models.RSIHistoryPoint,
	moneyFlowHistory []models.MoneyFlowIndicator,
) []models.DivergenceSignal {
	// Debug log
	fmt.Printf("[DEBUG] DetectMoneyFlowDivergence - klines length: %d, rsiHistory length: %d, moneyFlowHistory length: %d\n",
		len(klines), len(rsiHistory), len(moneyFlowHistory))

	// Make the requirement less strict: need at least 3 data points instead of 5
	if len(rsiHistory) < 3 || len(moneyFlowHistory) < 3 {
		fmt.Printf("[DEBUG] DetectMoneyFlowDivergence - Not enough data: rsi=%d, mfi=%d\n",
			len(rsiHistory), len(moneyFlowHistory))
		return []models.DivergenceSignal{} // Not enough data
	}

	var signals []models.DivergenceSignal

	// Check for bullish divergence: Price makes lower lows, RSI and MFI make higher lows
	bullishSignal, bullishFound := detectBullishDivergence(klines, rsiHistory, moneyFlowHistory)
	if bullishFound {
		fmt.Printf("[DEBUG] Bullish divergence detected: %+v\n", bullishSignal)
		signals = append(signals, bullishSignal)
	} else {
		fmt.Println("[DEBUG] No bullish divergence found")

		// Try to find partial bullish divergences
		partialSignals := detectPartialDivergences(klines, rsiHistory, moneyFlowHistory, true)
		if len(partialSignals) > 0 {
			fmt.Printf("[DEBUG] Found %d partial bullish divergences\n", len(partialSignals))
			signals = append(signals, partialSignals...)
		}
	}

	// Check for bearish divergence: Price makes higher highs, RSI and MFI make lower highs
	bearishSignal, bearishFound := detectBearishDivergence(klines, rsiHistory, moneyFlowHistory)
	if bearishFound {
		fmt.Printf("[DEBUG] Bearish divergence detected: %+v\n", bearishSignal)
		signals = append(signals, bearishSignal)
	} else {
		fmt.Println("[DEBUG] No bearish divergence found")

		// Try to find partial bearish divergences
		partialSignals := detectPartialDivergences(klines, rsiHistory, moneyFlowHistory, false)
		if len(partialSignals) > 0 {
			fmt.Printf("[DEBUG] Found %d partial bearish divergences\n", len(partialSignals))
			signals = append(signals, partialSignals...)
		}
	}

	return signals
}

// detectBullishDivergence checks for bullish divergence pattern and returns a properly formatted signal
func detectBullishDivergence(klines []models.Kline, rsiHistory []models.RSIHistoryPoint, moneyFlowHistory []models.MoneyFlowIndicator) (models.DivergenceSignal, bool) {
	fmt.Println("[DEBUG] detectBullishDivergence - checking for bullish divergence")

	// Make conditions less strict: minimum 6 klines and 3 history points
	if len(klines) < 6 || len(rsiHistory) < 3 || len(moneyFlowHistory) < 3 {
		fmt.Printf("[DEBUG] detectBullishDivergence - Not enough data: klines=%d, rsi=%d, mfi=%d (need 6,3,3)\n",
			len(klines), len(rsiHistory), len(moneyFlowHistory))
		return models.DivergenceSignal{}, false
	}

	// For bullish divergence:
	// 1. Price makes a lower low
	// 2. RSI and MFI make higher lows

	// Find local price lows and check if we have a lower low
	var priceTrend string
	if len(klines) >= 6 {
		// Dynamically adapt indices based on available data
		firstLowIdx := max(0, len(klines)-min(7, len(klines)-1))  // Older low
		secondLowIdx := max(0, len(klines)-min(2, len(klines)-1)) // Recent low

		firstLow := klines[firstLowIdx].Low
		secondLow := klines[secondLowIdx].Low

		fmt.Printf("[DEBUG] Price comparison - firstLow: %s, secondLow: %s\n",
			firstLow.String(), secondLow.String())

		if secondLow.LessThan(firstLow) {
			priceTrend = "lower_lows"
			fmt.Println("[DEBUG] Price trend detected: lower_lows")
		} else {
			fmt.Println("[DEBUG] Price is not making lower lows")
		}
	}

	// Check if RSI is making higher lows
	var rsiTrend string
	if len(rsiHistory) >= 3 {
		// Dynamically adapt indices based on available data
		firstLowIdx := min(len(rsiHistory)-1, max(2, len(rsiHistory)-2)) // Older low (remember rsiHistory is newest first)
		secondLowIdx := min(len(rsiHistory)-1, 1)                        // Recent low

		firstRSI := rsiHistory[firstLowIdx].Value
		secondRSI := rsiHistory[secondLowIdx].Value

		fmt.Printf("[DEBUG] RSI comparison - firstRSI: %s, secondRSI: %s\n",
			firstRSI.String(), secondRSI.String())

		if secondRSI.GreaterThan(firstRSI) {
			rsiTrend = "higher_lows"
			fmt.Println("[DEBUG] RSI trend detected: higher_lows")
		} else {
			fmt.Println("[DEBUG] RSI is not making higher lows")
		}
	}

	// Check if MFI is making higher lows
	var mfiTrend string
	if len(moneyFlowHistory) >= 3 {
		// Dynamically adapt indices based on available data
		firstLowIdx := min(len(moneyFlowHistory)-1, max(2, len(moneyFlowHistory)-2)) // Older low (assuming moneyFlowHistory is newest first)
		secondLowIdx := min(len(moneyFlowHistory)-1, 1)                              // Recent low

		firstMFI := moneyFlowHistory[firstLowIdx].MoneyFlowIndex
		secondMFI := moneyFlowHistory[secondLowIdx].MoneyFlowIndex

		fmt.Printf("[DEBUG] MFI comparison - firstMFI: %s, secondMFI: %s\n",
			firstMFI.String(), secondMFI.String())

		if secondMFI.GreaterThan(firstMFI) {
			mfiTrend = "higher_lows"
			fmt.Println("[DEBUG] MFI trend detected: higher_lows")
		} else {
			fmt.Println("[DEBUG] MFI is not making higher lows")
		}
	}

	// Bullish divergence exists if:
	// 1. Price is making lower lows
	// 2. RSI and MFI are making higher lows
	if priceTrend == "lower_lows" && rsiTrend == "higher_lows" && mfiTrend == "higher_lows" {
		// Construct time range with dynamic indices
		firstIndex := max(0, len(klines)-min(7, len(klines)-1))
		secondIndex := max(0, len(klines)-min(2, len(klines)-1))

		timeRange := []models.GMTPlus7Time{
			klines[firstIndex].CloseTime,  // Older low
			klines[secondIndex].CloseTime, // Recent low
		}

		return models.DivergenceSignal{
			Type:        PositiveDivergence,
			PriceTrend:  priceTrend,
			RSITrend:    rsiTrend,
			MFITrend:    mfiTrend,
			Confirmed:   true,
			TimeRange:   timeRange,
			Description: "Price makes lower lows while RSI and MFI make higher lows → bullish reversal possible",
		}, true
	}

	return models.DivergenceSignal{}, false
}

// detectBearishDivergence checks for bearish divergence pattern and returns a properly formatted signal
func detectBearishDivergence(klines []models.Kline, rsiHistory []models.RSIHistoryPoint, moneyFlowHistory []models.MoneyFlowIndicator) (models.DivergenceSignal, bool) {
	fmt.Println("[DEBUG] detectBearishDivergence - checking for bearish divergence")

	// Make conditions less strict: minimum 6 klines and 3 history points
	if len(klines) < 6 || len(rsiHistory) < 3 || len(moneyFlowHistory) < 3 {
		fmt.Printf("[DEBUG] detectBearishDivergence - Not enough data: klines=%d, rsi=%d, mfi=%d (need 6,3,3)\n",
			len(klines), len(rsiHistory), len(moneyFlowHistory))
		return models.DivergenceSignal{}, false
	}

	// For bearish divergence:
	// 1. Price makes a higher high
	// 2. RSI and MFI make lower highs

	// Find local price highs and check if we have a higher high
	var priceTrend string
	if len(klines) >= 6 {
		// Dynamically adapt indices based on available data
		firstHighIdx := max(0, len(klines)-min(7, len(klines)-1))  // Older high
		secondHighIdx := max(0, len(klines)-min(2, len(klines)-1)) // Recent high

		firstHigh := klines[firstHighIdx].High
		secondHigh := klines[secondHighIdx].High

		fmt.Printf("[DEBUG] Price comparison - firstHigh: %s, secondHigh: %s\n",
			firstHigh.String(), secondHigh.String())

		if secondHigh.GreaterThan(firstHigh) {
			priceTrend = "higher_highs"
			fmt.Println("[DEBUG] Price trend detected: higher_highs")
		} else {
			fmt.Println("[DEBUG] Price is not making higher highs")
		}
	}

	// Check if RSI is making lower highs
	var rsiTrend string
	if len(rsiHistory) >= 3 {
		// Dynamically adapt indices based on available data
		firstHighIdx := min(len(rsiHistory)-1, max(2, len(rsiHistory)-2)) // Older high (remember rsiHistory is newest first)
		secondHighIdx := min(len(rsiHistory)-1, 1)                        // Recent high

		firstRSI := rsiHistory[firstHighIdx].Value
		secondRSI := rsiHistory[secondHighIdx].Value

		fmt.Printf("[DEBUG] RSI comparison - firstRSI: %s, secondRSI: %s\n",
			firstRSI.String(), secondRSI.String())

		if secondRSI.LessThan(firstRSI) {
			rsiTrend = "lower_highs"
			fmt.Println("[DEBUG] RSI trend detected: lower_highs")
		} else {
			fmt.Println("[DEBUG] RSI is not making lower highs")
		}
	}

	// Check if MFI is making lower highs
	var mfiTrend string
	if len(moneyFlowHistory) >= 3 {
		// Dynamically adapt indices based on available data
		firstHighIdx := min(len(moneyFlowHistory)-1, max(2, len(moneyFlowHistory)-2)) // Older high (assuming moneyFlowHistory is newest first)
		secondHighIdx := min(len(moneyFlowHistory)-1, 1)                              // Recent high

		firstMFI := moneyFlowHistory[firstHighIdx].MoneyFlowIndex
		secondMFI := moneyFlowHistory[secondHighIdx].MoneyFlowIndex

		fmt.Printf("[DEBUG] MFI comparison - firstMFI: %s, secondMFI: %s\n",
			firstMFI.String(), secondMFI.String())

		if secondMFI.LessThan(firstMFI) {
			mfiTrend = "lower_highs"
			fmt.Println("[DEBUG] MFI trend detected: lower_highs")
		} else {
			fmt.Println("[DEBUG] MFI is not making lower highs")
		}
	}

	// Bearish divergence exists if:
	// 1. Price is making higher highs
	// 2. RSI and MFI are making lower highs
	if priceTrend == "higher_highs" && rsiTrend == "lower_highs" && mfiTrend == "lower_highs" {
		// Construct time range with dynamic indices
		firstIndex := max(0, len(klines)-min(7, len(klines)-1))
		secondIndex := max(0, len(klines)-min(2, len(klines)-1))

		timeRange := []models.GMTPlus7Time{
			klines[firstIndex].CloseTime,  // Older high
			klines[secondIndex].CloseTime, // Recent high
		}

		return models.DivergenceSignal{
			Type:        NegativeDivergence,
			PriceTrend:  priceTrend,
			RSITrend:    rsiTrend,
			MFITrend:    mfiTrend,
			Confirmed:   true,
			TimeRange:   timeRange,
			Description: "Price makes higher highs while RSI and MFI make lower highs → bearish reversal warning",
		}, true
	}

	return models.DivergenceSignal{}, false
}

// detectPartialDivergences detects partial divergences where not all conditions are met
// If bullishCheck is true, it looks for bullish divergence conditions
// If bullishCheck is false, it looks for bearish divergence conditions
func detectPartialDivergences(klines []models.Kline, rsiHistory []models.RSIHistoryPoint, moneyFlowHistory []models.MoneyFlowIndicator, bullishCheck bool) []models.DivergenceSignal {
	if len(klines) < 6 || len(rsiHistory) < 3 || len(moneyFlowHistory) < 3 {
		return []models.DivergenceSignal{}
	}

	var signals []models.DivergenceSignal

	// Dynamic indices based on available data
	firstIndex := max(0, len(klines)-min(7, len(klines)-1))
	secondIndex := max(0, len(klines)-min(2, len(klines)-1))

	// Calculate price change
	firstPrice := klines[firstIndex].Close
	secondPrice := klines[secondIndex].Close
	priceChange := secondPrice.Sub(firstPrice)

	// Get RSI trend
	firstRSIIdx := min(len(rsiHistory)-1, max(2, len(rsiHistory)-2))
	secondRSIIdx := min(len(rsiHistory)-1, 1)
	firstRSI := rsiHistory[firstRSIIdx].Value
	secondRSI := rsiHistory[secondRSIIdx].Value
	rsiHigherLow := secondRSI.GreaterThan(firstRSI)
	rsiLowerHigh := secondRSI.LessThan(firstRSI)

	// Get MFI trend
	firstMFIIdx := min(len(moneyFlowHistory)-1, max(2, len(moneyFlowHistory)-2))
	secondMFIIdx := min(len(moneyFlowHistory)-1, 1)
	firstMFI := moneyFlowHistory[firstMFIIdx].MoneyFlowIndex
	secondMFI := moneyFlowHistory[secondMFIIdx].MoneyFlowIndex
	mfiHigherLow := secondMFI.GreaterThan(firstMFI)
	mfiLowerHigh := secondMFI.LessThan(firstMFI)

	timeRange := []models.GMTPlus7Time{
		klines[firstIndex].CloseTime,
		klines[secondIndex].CloseTime,
	}

	// Check for partial bullish divergence - price down but at least one indicator up
	// OR when MFI is going up, which can be bullish even if price isn't dropping
	if bullishCheck {
		if (priceChange.LessThan(decimal.Zero) && (rsiHigherLow || mfiHigherLow)) ||
			(mfiHigherLow) {
			divergence := models.DivergenceSignal{
				Type:        PartialDivergence,
				PriceTrend:  formatTrend(priceChange.LessThan(decimal.Zero), "lower", "neutral"),
				RSITrend:    formatTrend(rsiHigherLow, "higher_lows", "no_trend"),
				MFITrend:    formatTrend(mfiHigherLow, "higher_lows", "no_trend"),
				Confirmed:   false,
				TimeRange:   timeRange,
				Description: "Partial bullish divergence: " + formatTrend(mfiHigherLow, "Money flow increasing", ""),
			}
			signals = append(signals, divergence)
		}
	} else {
		// Check for partial bearish divergence - price up but at least one indicator down
		if priceChange.GreaterThan(decimal.Zero) && (rsiLowerHigh || mfiLowerHigh) {
			divergence := models.DivergenceSignal{
				Type:        PartialDivergence,
				PriceTrend:  "higher",
				RSITrend:    formatTrend(rsiLowerHigh, "lower_highs", "no_trend"),
				MFITrend:    formatTrend(mfiLowerHigh, "lower_highs", "no_trend"),
				Confirmed:   false,
				TimeRange:   timeRange,
				Description: "Partial bearish divergence: Price up while RSI or MFI trending lower",
			}
			signals = append(signals, divergence)
		}
	}

	return signals
}

// formatTrend helper function to format trend text
func formatTrend(condition bool, trueText, falseText string) string {
	if condition {
		return trueText
	}
	return falseText
}
