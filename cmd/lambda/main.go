package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"go-crypto/internal/api"
	"go-crypto/internal/config"
	"go-crypto/internal/models"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
)

// LambdaHandler wraps our HTTP API for Lambda
type LambdaHandler struct {
	server *api.Server
}

// getRealPrice fetches real price data from Binance via server
func (h *LambdaHandler) getRealPrice(ctx context.Context, symbol string) (*api.PriceResponse, error) {
	// Use the server's new public method to get real data
	return h.server.GetPrice(ctx, symbol)
}

// getCurrentTimestamp returns current time in GMT+7 timezone
func getCurrentTimestamp() string {
	// Load GMT+7 timezone
	loc, err := time.LoadLocation("Asia/Bangkok") // GMT+7
	if err != nil {
		// Fallback to UTC+7 if timezone loading fails
		loc = time.FixedZone("GMT+7", 7*60*60)
	}
	// Create a GMTPlus7Time and let it handle the JSON formatting
	gmt7Time := models.NewGMTPlus7Time(time.Now().In(loc))
	// Remove the quotes that MarshalJSON adds
	jsonTime, _ := gmt7Time.MarshalJSON()
	return string(jsonTime[1 : len(jsonTime)-1]) // Remove surrounding quotes
}

// NewLambdaHandler creates a new Lambda handler
func NewLambdaHandler() (*LambdaHandler, error) {
	// Load configuration with defaults for Lambda
	cfg := &config.Config{
		Binance: config.BinanceConfig{
			BaseURL:             "https://api.binance.com",
			FuturesBaseURL:      "https://fapi.binance.com",
			WebSocketURL:        "wss://stream.binance.com:9443",
			FuturesWebSocketURL: "wss://fstream.binance.com/ws",
			FuturesSymbols:      []string{"XAUUSDT", "XAGUSDT"},
			Timeout:             30,
			RateLimit:           1200,
		},
		Symbols:   []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "XAUTUSDT", "XAUUSDT", "XAGUSDT"},
		Intervals: []string{"15m", "1h", "4h", "1d"},
		Indicators: config.IndicatorConfig{
			RSI: config.RSIConfig{
				Periods: []int{6, 12, 24},
			},
			MA: config.MAConfig{
				Periods: []int{7, 20, 25, 50, 99, 200},
				Type:    "SMA",
			},
			KDJ: config.KDJConfig{
				KPeriod: 9,
				DPeriod: 3,
				JPeriod: 3,
			},
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	// Create logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Create API server
	server := api.NewServer(cfg, logger)

	return &LambdaHandler{
		server: server,
	}, nil
}

// HandleRequest processes Lambda API Gateway requests
func (h *LambdaHandler) HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Log the incoming request for debugging
	log.Printf("Lambda request - Path: %s, Method: %s", request.Path, request.HTTPMethod)

	// Remove stage prefix from path (e.g., /prod/api/v1/health -> /api/v1/health)
	path := request.Path
	if strings.HasPrefix(path, "/prod/") {
		path = strings.TrimPrefix(path, "/prod")
	} else if strings.HasPrefix(path, "/dev/") {
		path = strings.TrimPrefix(path, "/dev")
	} else if strings.HasPrefix(path, "/test/") {
		path = strings.TrimPrefix(path, "/test")
	}

	log.Printf("Cleaned path: %s", path)

	httpRequest, err := h.toHTTPRequest(ctx, request, path)
	if err != nil {
		return h.errorResponse(400, "Invalid request")
	}
	recorder := httptest.NewRecorder()
	h.server.GetHandler().ServeHTTP(recorder, httpRequest)

	headers := h.getCORSHeaders()
	for key, values := range recorder.Result().Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	statusCode := recorder.Code
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       recorder.Body.String(),
		Headers:    headers,
	}, nil
}

// handleHealth handles health check requests
func (h *LambdaHandler) handleHealth(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	body := `{"status":"healthy","service":"go-crypto-api","version":"1.0.0","deployment":"serverless-lambda"}`

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       body,
		Headers:    h.getCORSHeaders(),
	}, nil
}

// handlePrice handles price requests
func (h *LambdaHandler) handlePrice(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Clean the path to remove stage prefix
	path := request.Path
	if strings.HasPrefix(path, "/prod/") {
		path = strings.TrimPrefix(path, "/prod")
	} else if strings.HasPrefix(path, "/dev/") {
		path = strings.TrimPrefix(path, "/dev")
	} else if strings.HasPrefix(path, "/test/") {
		path = strings.TrimPrefix(path, "/test")
	}

	// Extract symbol from cleaned path
	pathParts := strings.Split(path, "/")
	if len(pathParts) < 5 {
		return h.errorResponse(400, "Invalid path format. Expected /api/v1/price/{symbol}")
	}

	symbol := pathParts[4]

	// Call the real API server method
	price, err := h.getRealPrice(ctx, symbol)
	if err != nil {
		log.Printf("Failed to get real price: %v", err)
		// Fall back to mock data if real API fails
		body := fmt.Sprintf(`{"symbol":"%s","price":"45000.50","timestamp":"%s","source":"mock"}`, symbol, getCurrentTimestamp())
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       body,
			Headers:    h.getCORSHeaders(),
		}, nil
	}

	// Return real data
	jsonData, err := json.Marshal(price)
	if err != nil {
		return h.errorResponse(500, "Failed to serialize price data")
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(jsonData),
		Headers:    h.getCORSHeaders(),
	}, nil
}

// handleAnalysis handles analysis requests
func (h *LambdaHandler) handleAnalysis(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Clean the path to remove stage prefix
	path := request.Path
	if strings.HasPrefix(path, "/prod/") {
		path = strings.TrimPrefix(path, "/prod")
	} else if strings.HasPrefix(path, "/dev/") {
		path = strings.TrimPrefix(path, "/dev")
	} else if strings.HasPrefix(path, "/test/") {
		path = strings.TrimPrefix(path, "/test")
	}

	// Extract symbol from cleaned path
	pathParts := strings.Split(path, "/")
	if len(pathParts) < 5 {
		return h.errorResponse(400, "Invalid path format. Expected /api/v1/analysis/{symbol}")
	}

	symbol := pathParts[4]
	interval := request.QueryStringParameters["interval"]
	if interval == "" {
		interval = "15m"
	}

	// Get real analysis data from server
	analysis, err := h.server.GetAnalysis(ctx, symbol, interval)
	if err != nil {
		log.Printf("Failed to get real analysis: %v", err)
		// Fall back to mock data if real API fails
		body := fmt.Sprintf(`{
			"symbol":"%s",
			"interval":"%s",
			"timestamp":"%s",
			"analysis":{
				"rsi":{"rsi_6":65.2,"rsi_12":58.7,"rsi_24":61.3},
				"moving_averages":{"ma_7":43100.25,"ma_25":42980.50,"ma_99":42750.75},
				"kdj":{"k":72.5,"d":68.9,"j":79.7}
			},
			"signals":{"trend":"bullish","strength":"moderate","recommendation":"hold"},
			"source":"mock"
		}`, symbol, interval, getCurrentTimestamp())

		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       body,
			Headers:    h.getCORSHeaders(),
		}, nil
	}

	// Return real data with consistent API response structure
	response := map[string]interface{}{
		"success": true,
		"data":    analysis,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return h.errorResponse(500, "Failed to serialize analysis data")
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(jsonData),
		Headers:    h.getCORSHeaders(),
	}, nil
}

// handleMultiAnalysis handles multi-timeframe analysis requests
func (h *LambdaHandler) handleMultiAnalysis(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Clean the path to remove stage prefix
	path := request.Path
	if strings.HasPrefix(path, "/prod/") {
		path = strings.TrimPrefix(path, "/prod")
	} else if strings.HasPrefix(path, "/dev/") {
		path = strings.TrimPrefix(path, "/dev")
	} else if strings.HasPrefix(path, "/test/") {
		path = strings.TrimPrefix(path, "/test")
	}

	// Extract symbol from cleaned path
	pathParts := strings.Split(path, "/")
	if len(pathParts) < 5 {
		return h.errorResponse(400, "Invalid path format. Expected /api/v1/multi-analysis/{symbol}")
	}

	symbol := pathParts[4]

	// Get timeframes from query parameter (optional)
	var timeframes []string
	if tf, exists := request.QueryStringParameters["timeframes"]; exists && tf != "" {
		timeframes = strings.Split(tf, ",")
	}

	// Get enhanced multi-analysis data from server
	// Enhanced features for 15m, basic for 4h and 1d
	multiAnalysis, err := h.server.GetEnhancedMultiAnalysis(ctx, symbol, timeframes)
	if err != nil {
		log.Printf("Failed to get enhanced multi-analysis: %v", err)
		return h.errorResponse(500, "Failed to get multi-analysis data")
	}

	// Return real data without extra wrapper
	jsonData, err := json.Marshal(multiAnalysis)
	if err != nil {
		return h.errorResponse(500, "Failed to serialize multi-analysis data")
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(jsonData),
		Headers:    h.getCORSHeaders(),
	}, nil
}

func (h *LambdaHandler) toHTTPRequest(ctx context.Context, request events.APIGatewayProxyRequest, path string) (*http.Request, error) {
	query := url.Values{}
	for key, value := range request.QueryStringParameters {
		query.Set(key, value)
	}
	for key, values := range request.MultiValueQueryStringParameters {
		for _, value := range values {
			query.Add(key, value)
		}
	}

	target := path
	if encodedQuery := query.Encode(); encodedQuery != "" {
		target += "?" + encodedQuery
	}

	method := request.HTTPMethod
	if method == "" {
		method = http.MethodGet
	}

	httpRequest, err := http.NewRequestWithContext(ctx, method, target, strings.NewReader(request.Body))
	if err != nil {
		return nil, err
	}
	for key, value := range request.Headers {
		httpRequest.Header.Set(key, value)
	}
	httpRequest.RemoteAddr = request.RequestContext.Identity.SourceIP
	return httpRequest, nil
}

// getCORSHeaders returns CORS headers for API responses
func (h *LambdaHandler) getCORSHeaders() map[string]string {
	return map[string]string{
		"Content-Type":                 "application/json",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Authorization",
	}
}

// errorResponse creates an error response
func (h *LambdaHandler) errorResponse(statusCode int, message string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       fmt.Sprintf(`{"error": "%s"}`, message),
		Headers:    h.getCORSHeaders(),
	}, nil
}

func main() {
	// Create Lambda handler
	handler, err := NewLambdaHandler()
	if err != nil {
		log.Fatalf("Failed to create Lambda handler: %v", err)
	}

	// Start Lambda
	lambda.Start(handler.HandleRequest)
}
