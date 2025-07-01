package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-crypto/internal/config"
	"go-crypto/internal/models"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// Client represents Binance API client
type Client struct {
	config     *config.BinanceConfig
	httpClient *http.Client
	logger     *logrus.Logger
	wsConn     *websocket.Conn
}

// NewClient creates a new Binance API client
func NewClient(cfg *config.BinanceConfig, logger *logrus.Logger) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
		logger: logger,
	}
}

// GetKlines fetches historical kline data
func (c *Client) GetKlines(ctx context.Context, symbol models.Symbol, interval models.Timeframe, limit int) ([]models.Kline, error) {
	endpoint := "/api/v3/klines"
	params := url.Values{}
	params.Set("symbol", string(symbol))
	params.Set("interval", string(interval))
	params.Set("limit", strconv.Itoa(limit))

	url := fmt.Sprintf("%s%s?%s", c.config.BaseURL, endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var rawKlines []models.BinanceKlineResponse
	if err := json.Unmarshal(body, &rawKlines); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	klines := make([]models.Kline, len(rawKlines))
	for i, rawKline := range rawKlines {
		kline, err := c.parseKline(rawKline, symbol, interval)
		if err != nil {
			c.logger.WithError(err).Warn("Failed to parse kline")
			continue
		}
		klines[i] = kline
	}

	c.logger.WithFields(logrus.Fields{
		"symbol":   symbol,
		"interval": interval,
		"count":    len(klines),
	}).Info("Successfully fetched klines")

	return klines, nil
}

// GetTicker24hr fetches 24hr ticker price statistics
func (c *Client) GetTicker24hr(ctx context.Context, symbol models.Symbol) (*models.TickerPrice, error) {
	endpoint := "/api/v3/ticker/24hr"
	params := url.Values{}
	params.Set("symbol", string(symbol))

	url := fmt.Sprintf("%s%s?%s", c.config.BaseURL, endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var rawTicker models.BinanceTickerResponse
	if err := json.Unmarshal(body, &rawTicker); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	ticker, err := c.parseTicker(rawTicker, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ticker: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"symbol": symbol,
		"price":  ticker.Price,
	}).Info("Successfully fetched ticker")

	return ticker, nil
}

// ConnectWebSocket establishes WebSocket connection for real-time data
func (c *Client) ConnectWebSocket(ctx context.Context, symbols []models.Symbol, intervals []models.Timeframe) (<-chan models.WebSocketMessage, error) {
	// Build stream names
	var streams []string
	for _, symbol := range symbols {
		symbolLower := strings.ToLower(string(symbol))
		// Add ticker stream
		streams = append(streams, fmt.Sprintf("%s@ticker", symbolLower))

		// Add kline streams for each interval
		for _, interval := range intervals {
			streams = append(streams, fmt.Sprintf("%s@kline_%s", symbolLower, string(interval)))
		}
	}

	streamParam := strings.Join(streams, "/")
	wsURL := fmt.Sprintf("%s/stream?streams=%s", c.config.WebSocketURL, streamParam)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	c.wsConn = conn
	msgChan := make(chan models.WebSocketMessage, 100)

	go func() {
		defer close(msgChan)
		defer conn.Close()

		for {
			select {
			case <-ctx.Done():
				c.logger.Info("WebSocket context cancelled")
				return
			default:
				_, message, err := conn.ReadMessage()
				if err != nil {
					c.logger.WithError(err).Error("Failed to read WebSocket message")
					return
				}

				var wsMsg models.WebSocketMessage
				if err := json.Unmarshal(message, &wsMsg); err != nil {
					c.logger.WithError(err).Warn("Failed to unmarshal WebSocket message")
					continue
				}

				select {
				case msgChan <- wsMsg:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	c.logger.WithField("streams", len(streams)).Info("WebSocket connection established")
	return msgChan, nil
}

// CloseWebSocket closes the WebSocket connection
func (c *Client) CloseWebSocket() error {
	if c.wsConn != nil {
		return c.wsConn.Close()
	}
	return nil
}

// parseKline converts raw kline data to structured format
func (c *Client) parseKline(raw models.BinanceKlineResponse, symbol models.Symbol, interval models.Timeframe) (models.Kline, error) {
	if len(raw) < 12 {
		return models.Kline{}, fmt.Errorf("invalid kline data length: %d", len(raw))
	}

	// Parse timestamps
	openTime := time.Unix(int64(raw[0].(float64))/1000, 0)
	closeTime := time.Unix(int64(raw[6].(float64))/1000, 0)

	// Parse decimal values
	open, _ := decimal.NewFromString(raw[1].(string))
	high, _ := decimal.NewFromString(raw[2].(string))
	low, _ := decimal.NewFromString(raw[3].(string))
	close, _ := decimal.NewFromString(raw[4].(string))
	volume, _ := decimal.NewFromString(raw[5].(string))
	quoteVolume, _ := decimal.NewFromString(raw[7].(string))
	takerBuyBaseVolume, _ := decimal.NewFromString(raw[9].(string))
	takerBuyQuoteVolume, _ := decimal.NewFromString(raw[10].(string))

	return models.Kline{
		Symbol:              symbol,
		OpenTime:            openTime,
		CloseTime:           closeTime,
		Open:                open,
		High:                high,
		Low:                 low,
		Close:               close,
		Volume:              volume,
		QuoteAssetVolume:    quoteVolume,
		NumberOfTrades:      int64(raw[8].(float64)),
		TakerBuyBaseVolume:  takerBuyBaseVolume,
		TakerBuyQuoteVolume: takerBuyQuoteVolume,
		Timeframe:           interval,
	}, nil
}

// parseTicker converts raw ticker data to structured format
func (c *Client) parseTicker(raw models.BinanceTickerResponse, symbol models.Symbol) (*models.TickerPrice, error) {
	price, err := decimal.NewFromString(raw.LastPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price: %w", err)
	}

	priceChangePercent, err := decimal.NewFromString(raw.PriceChangePercent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price change percent: %w", err)
	}

	volume, err := decimal.NewFromString(raw.Volume)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume: %w", err)
	}

	quoteVolume, err := decimal.NewFromString(raw.QuoteVolume)
	if err != nil {
		return nil, fmt.Errorf("failed to parse quote volume: %w", err)
	}

	return &models.TickerPrice{
		Symbol:             symbol,
		Price:              price,
		PriceChangePercent: priceChangePercent,
		Volume:             volume,
		QuoteVolume:        quoteVolume,
		Timestamp:          time.Now(),
	}, nil
}
