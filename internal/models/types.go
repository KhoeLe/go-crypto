package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Symbol represents a trading pair
type Symbol string

const (
	BTCUSDT   Symbol = "BTCUSDT"
	ETHUSDT   Symbol = "ETHUSDT"
	BNBUSDT   Symbol = "BNBUSDT"
	ETHFIUSDT Symbol = "ETHFIUSDT"
)

// Timeframe represents different chart timeframes
type Timeframe string

const (
	Timeframe15m Timeframe = "15m"
	Timeframe4h  Timeframe = "4h"
	Timeframe1d  Timeframe = "1d"
)

// Kline represents candlestick data
type Kline struct {
	Symbol              Symbol          `json:"symbol"`
	OpenTime            time.Time       `json:"openTime"`
	CloseTime           time.Time       `json:"closeTime"`
	Open                decimal.Decimal `json:"open"`
	High                decimal.Decimal `json:"high"`
	Low                 decimal.Decimal `json:"low"`
	Close               decimal.Decimal `json:"close"`
	Volume              decimal.Decimal `json:"volume"`
	QuoteAssetVolume    decimal.Decimal `json:"quoteAssetVolume"`
	NumberOfTrades      int64           `json:"numberOfTrades"`
	TakerBuyBaseVolume  decimal.Decimal `json:"takerBuyBaseVolume"`
	TakerBuyQuoteVolume decimal.Decimal `json:"takerBuyQuoteVolume"`
	Timeframe           Timeframe       `json:"timeframe"`
}

// TickerPrice represents current price data
type TickerPrice struct {
	Symbol             Symbol          `json:"symbol"`
	Price              decimal.Decimal `json:"price"`
	PriceChangePercent decimal.Decimal `json:"priceChangePercent"`
	Volume             decimal.Decimal `json:"volume"`
	QuoteVolume        decimal.Decimal `json:"quoteVolume"`
	Timestamp          time.Time       `json:"timestamp"`
}

// TechnicalIndicators represents calculated technical indicators
type TechnicalIndicators struct {
	Symbol    Symbol                  `json:"symbol"`
	Timeframe Timeframe               `json:"timeframe"`
	Timestamp time.Time               `json:"timestamp"`
	RSI       map[int]decimal.Decimal `json:"rsi"` // RSI values by period
	MA        map[int]decimal.Decimal `json:"ma"`  // MA values by period
	KDJ       KDJIndicator            `json:"kdj"`
	MACD      MACDIndicator           `json:"macd"`
}

// KDJIndicator represents K, D, J values
type KDJIndicator struct {
	K decimal.Decimal `json:"k"`
	D decimal.Decimal `json:"d"`
	J decimal.Decimal `json:"j"`
}

// MACDIndicator represents MACD (Moving Average Convergence Divergence) values
type MACDIndicator struct {
	MACD      decimal.Decimal `json:"macd"`      // MACD line
	Signal    decimal.Decimal `json:"signal"`    // Signal line
	Histogram decimal.Decimal `json:"histogram"` // MACD - Signal
}

// MarketData represents comprehensive market data
type MarketData struct {
	Symbol     Symbol              `json:"symbol"`
	Timeframe  Timeframe           `json:"timeframe"`
	Price      TickerPrice         `json:"price"`
	Klines     []Kline             `json:"klines"`
	Indicators TechnicalIndicators `json:"indicators"`
	UpdatedAt  time.Time           `json:"updatedAt"`
}

// WebSocketMessage represents WebSocket message structure
type WebSocketMessage struct {
	Stream string      `json:"stream"`
	Data   interface{} `json:"data"`
}

// BinanceKlineResponse represents Binance API kline response
type BinanceKlineResponse []interface{}

// BinanceTickerResponse represents Binance API ticker response
type BinanceTickerResponse struct {
	Symbol             string `json:"symbol"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	WeightedAvgPrice   string `json:"weightedAvgPrice"`
	PrevClosePrice     string `json:"prevClosePrice"`
	LastPrice          string `json:"lastPrice"`
	LastQty            string `json:"lastQty"`
	BidPrice           string `json:"bidPrice"`
	BidQty             string `json:"bidQty"`
	AskPrice           string `json:"askPrice"`
	AskQty             string `json:"askQty"`
	OpenPrice          string `json:"openPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	OpenTime           int64  `json:"openTime"`
	CloseTime          int64  `json:"closeTime"`
	Count              int64  `json:"count"`
}

// MoneyFlowIndicator represents money flow analysis
type MoneyFlowIndicator struct {
	MoneyFlowIndex    decimal.Decimal `json:"money_flow_index"`    // MFI value
	PositiveMoneyFlow decimal.Decimal `json:"positive_money_flow"` // Positive money flow
	NegativeMoneyFlow decimal.Decimal `json:"negative_money_flow"` // Negative money flow
	MoneyFlowChange   decimal.Decimal `json:"money_flow_change"`   // % change between periods
	Timestamp         time.Time       `json:"timestamp"`
}

// VolumeBreakout represents volume breakout analysis
type VolumeBreakout struct {
	IsBreakout        bool            `json:"is_breakout"`        // Whether volume breakout detected
	BreakoutStrength  decimal.Decimal `json:"breakout_strength"`  // Strength of breakout (1-10)
	VolumeMultiplier  decimal.Decimal `json:"volume_multiplier"`  // Current volume vs average
	AverageVolume     decimal.Decimal `json:"average_volume"`     // Average volume over period
	CurrentVolume     decimal.Decimal `json:"current_volume"`     // Current period volume
	BreakoutDirection string          `json:"breakout_direction"` // "bullish" or "bearish"
	Timestamp         time.Time       `json:"timestamp"`
}

// VolumeDelta represents buy vs sell pressure analysis
type VolumeDelta struct {
	BuyVolume    decimal.Decimal `json:"buy_volume"`    // Approximated buy volume (taker buy)
	SellVolume   decimal.Decimal `json:"sell_volume"`   // Approximated sell volume (total - taker buy)
	Delta        decimal.Decimal `json:"delta"`         // Buy volume - sell volume
	DeltaPercent decimal.Decimal `json:"delta_percent"` // Delta as percentage of total volume
	Pressure     string          `json:"pressure"`      // "buy_pressure", "sell_pressure", "balanced"
	Strength     int             `json:"strength"`      // Pressure strength (1-10)
	Timestamp    time.Time       `json:"timestamp"`
}

// WhaleVolumeSpike represents large volume spike detection
type WhaleVolumeSpike struct {
	IsWhaleSpike     bool            `json:"is_whale_spike"`    // Whether whale volume detected
	SpikeVolume      decimal.Decimal `json:"spike_volume"`      // Volume of the spike
	SpikeValueUSDT   decimal.Decimal `json:"spike_value_usdt"`  // Value in USDT (volume * price)
	ThresholdUSDT    decimal.Decimal `json:"threshold_usdt"`    // Threshold for whale detection (100k USDT)
	VolumeMultiplier decimal.Decimal `json:"volume_multiplier"` // Current volume vs recent average
	Timestamp        time.Time       `json:"timestamp"`
}

// HistoricalIndicators represents historical tracking of indicators
type HistoricalIndicators struct {
	RSIHistory       []RSIHistoryPoint    `json:"rsi_history"`        // Historical RSI values
	MAHistory        []MAHistoryPoint     `json:"ma_history"`         // Historical MA values
	MoneyFlowHistory []MoneyFlowIndicator `json:"money_flow_history"` // Historical money flow
	VolumeHistory    []VolumeBreakout     `json:"volume_history"`     // Historical volume analysis
}

// RSIHistoryPoint represents a single RSI calculation point
type RSIHistoryPoint struct {
	Period    int             `json:"period"`    // RSI period (6, 12, 24)
	Value     decimal.Decimal `json:"value"`     // RSI value
	Timestamp time.Time       `json:"timestamp"` // Calculation time
}

// MAHistoryPoint represents a single MA calculation point
type MAHistoryPoint struct {
	Period    int             `json:"period"`    // MA period (7, 25, 99)
	Type      string          `json:"type"`      // MA type (SMA, EMA)
	Value     decimal.Decimal `json:"value"`     // MA value
	Timestamp time.Time       `json:"timestamp"` // Calculation time
}

// EnhancedAnalysisResponse represents enhanced analysis with new features
type EnhancedAnalysisResponse struct {
	Symbol          string                     `json:"symbol"`
	Timeframe       string                     `json:"timeframe"`
	Price           *TickerPrice               `json:"price"`
	Klines          []Kline                    `json:"klines"` // Include klines data
	RSI             map[string]decimal.Decimal `json:"rsi"`    // RSI values
	MA              map[string]decimal.Decimal `json:"ma"`     // MA values
	KDJ             KDJIndicator               `json:"kdj"`
	MACD            MACDIndicator              `json:"macd"`
	Volatility      decimal.Decimal            `json:"volatility"`
	MarketSentiment string                     `json:"market_sentiment"`
	MoneyFlow       MoneyFlowIndicator         `json:"money_flow"`      // Money flow analysis
	VolumeBreakout  VolumeBreakout             `json:"volume_breakout"` // Volume breakout detection
	VolumeDelta     VolumeDelta                `json:"volume_delta"`    // Buy vs sell pressure analysis
	WhaleActivity   WhaleVolumeSpike           `json:"whale_activity"`  // Whale volume spike detection
	Historical      HistoricalIndicators       `json:"historical"`      // Historical data
	Signals         []string                   `json:"signals"`
	Timestamp       time.Time                  `json:"timestamp"`
}
