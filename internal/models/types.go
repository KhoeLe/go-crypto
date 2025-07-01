package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Symbol represents a trading pair
type Symbol string

const (
	BTCUSDT Symbol = "BTCUSDT"
	ETHUSDT Symbol = "ETHUSDT"
	BNBUSDT Symbol = "BNBUSDT"
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
}

// KDJIndicator represents K, D, J values
type KDJIndicator struct {
	K decimal.Decimal `json:"k"`
	D decimal.Decimal `json:"d"`
	J decimal.Decimal `json:"j"`
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
