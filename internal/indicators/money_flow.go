package indicators

import (
	"fmt"

	"go-crypto/internal/models"

	"github.com/shopspring/decimal"
)

// CalculateHistoricalMoneyFlow calculates money flow index for a series of historical points
func (c *Calculator) CalculateHistoricalMoneyFlow(klines []models.Kline, period int, count int) ([]models.MoneyFlowIndicator, error) {
	if len(klines) < period+1 {
		return nil, fmt.Errorf("insufficient data for historical MFI calculation: need %d, got %d", period+1, len(klines))
	}

	// Determine how many historical points we can calculate
	// We need at least 8 points but no more than 15 as per requirements
	historyPoints := len(klines) - period
	if historyPoints <= 0 {
		return nil, fmt.Errorf("insufficient data for historical points")
	}

	// If count is specified, limit to that many points
	if count > 0 && count < historyPoints {
		historyPoints = count
	}

	// Limit to 8-15 points as per requirements
	if historyPoints < 8 {
		historyPoints = 8
	} else if historyPoints > 15 {
		historyPoints = 15
	}

	result := make([]models.MoneyFlowIndicator, 0, historyPoints)

	// First, calculate all MFI values
	mfiValues := make([]models.MoneyFlowIndicator, 0, historyPoints)

	for i := 0; i < historyPoints; i++ {
		// Use a sliding window of period+1 klines for each calculation
		endIdx := len(klines) - i
		startIdx := endIdx - period - 1
		if startIdx < 0 {
			break
		}

		windowKlines := klines[startIdx:endIdx]
		mfi, err := c.calculateSingleMoneyFlow(windowKlines, period)
		if err != nil {
			continue
		}

		mfiValues = append(mfiValues, mfi)
	}

	// Then, calculate percentage changes and ensure flow types are correct
	for i := 0; i < len(mfiValues); i++ {
		// Calculate percentage change
		if i+1 < len(mfiValues) && !mfiValues[i+1].MoneyFlowIndex.IsZero() {
			current := mfiValues[i].MoneyFlowIndex
			previous := mfiValues[i+1].MoneyFlowIndex
			change := current.Sub(previous)
			percentChange := change.Div(previous).Mul(decimal.NewFromInt(100))
			mfiValues[i].MoneyFlowChange = percentChange
		}

		// Add to result
		result = append(result, mfiValues[i])
	}

	// Ensure we have a proper sequence of entries with correctly calculated flow types
	if len(result) > 1 {
		for i := 1; i < len(result); i++ {
			// Double check flow type using the previous entry's typical price for comparison
			if result[i].TypicalPrice.GreaterThan(result[i-1].TypicalPrice) {
				result[i].FlowType = "positive"
			} else if result[i].TypicalPrice.LessThan(result[i-1].TypicalPrice) {
				result[i].FlowType = "negative"
			} else {
				result[i].FlowType = "neutral"
			}
		}
	}

	return result, nil
}

// calculateSingleMoneyFlow calculates a single money flow index point
func (c *Calculator) calculateSingleMoneyFlow(klines []models.Kline, period int) (models.MoneyFlowIndicator, error) {
	if len(klines) < period+1 {
		return models.MoneyFlowIndicator{}, fmt.Errorf("insufficient data for MFI calculation")
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

	// Calculate positive and negative money flows for the period
	for i := 0; i < period && i < len(typicalPrices)-1; i++ {
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

	// Get the last kline for timestamp and typical price calculation
	lastKline := klines[len(klines)-1]
	typicalPrice := lastKline.High.Add(lastKline.Low).Add(lastKline.Close).Div(decimal.NewFromInt(3))
	rawMoneyFlow := typicalPrice.Mul(lastKline.Volume)

	// Determine flow type based on price direction
	flowType := "neutral"
	if len(klines) > 1 {
		prevKline := klines[len(klines)-2]
		prevTypicalPrice := prevKline.High.Add(prevKline.Low).Add(prevKline.Close).Div(decimal.NewFromInt(3))

		if typicalPrice.GreaterThan(prevTypicalPrice) {
			flowType = "positive"
		} else if typicalPrice.LessThan(prevTypicalPrice) {
			flowType = "negative"
		}
	}

	return models.MoneyFlowIndicator{
		MoneyFlowIndex:    mfi,
		PositiveMoneyFlow: positiveFlow,
		NegativeMoneyFlow: negativeFlow,
		TypicalPrice:      typicalPrice,
		RawMoneyFlow:      rawMoneyFlow,
		FlowType:          flowType,
		Timestamp:         lastKline.CloseTime,
	}, nil
}
