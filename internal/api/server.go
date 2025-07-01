package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-crypto/internal/binance"
	"go-crypto/internal/config"
	"go-crypto/internal/indicators"
	"go-crypto/internal/models"
	"go-crypto/pkg/utils"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// Server represents the API server
type Server struct {
	router        *mux.Router
	binanceClient *binance.Client
	calculator    *indicators.Calculator
	config        *config.Config
	logger        *logrus.Logger
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, logger *logrus.Logger) *Server {
	s := &Server{
		router:        mux.NewRouter(),
		binanceClient: binance.NewClient(&cfg.Binance, logger),
		calculator:    indicators.NewCalculator(),
		config:        cfg,
		logger:        logger,
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Market data endpoints
	api.HandleFunc("/price/{symbol}", s.handleGetPrice).Methods("GET")
	api.HandleFunc("/ticker/{symbol}", s.handleGetTicker).Methods("GET")
	api.HandleFunc("/klines/{symbol}", s.handleGetKlines).Methods("GET")

	// Technical analysis endpoints
	api.HandleFunc("/indicators/{symbol}", s.handleGetIndicators).Methods("GET")
	api.HandleFunc("/analysis/{symbol}", s.handleGetAnalysis).Methods("GET")
	api.HandleFunc("/signals/{symbol}", s.handleGetSignals).Methods("GET")

	// Multi-timeframe analysis
	api.HandleFunc("/multi-analysis/{symbol}", s.handleGetMultiAnalysis).Methods("GET")

	// Real-time endpoints
	api.HandleFunc("/stream/{symbol}", s.handleStreamInfo).Methods("GET")

	// Health and info endpoints
	api.HandleFunc("/health", s.handleHealth).Methods("GET")
	api.HandleFunc("/config", s.handleGetConfig).Methods("GET")
	api.HandleFunc("/symbols", s.handleGetSymbols).Methods("GET")

	// CORS middleware
	s.router.Use(s.corsMiddleware)
	s.router.Use(s.loggingMiddleware)

	// Serve static files (for future web interface)
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/static/")))
}

// Response structures
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

type PriceResponse struct {
	Symbol    string          `json:"symbol"`
	Price     decimal.Decimal `json:"price"`
	Timestamp time.Time       `json:"timestamp"`
}

type IndicatorsResponse struct {
	Symbol    string                     `json:"symbol"`
	Timeframe string                     `json:"timeframe"`
	RSI       map[string]decimal.Decimal `json:"rsi"`
	MA        map[string]decimal.Decimal `json:"ma"`
	KDJ       models.KDJIndicator        `json:"kdj"`
	Timestamp time.Time                  `json:"timestamp"`
}

type AnalysisResponse struct {
	Symbol     string                     `json:"symbol"`
	Timeframe  string                     `json:"timeframe"`
	Price      *models.TickerPrice        `json:"price"`
	RSI        map[string]decimal.Decimal `json:"rsi"`
	MA         map[string]decimal.Decimal `json:"ma"`
	KDJ        models.KDJIndicator        `json:"kdj"`
	Volatility decimal.Decimal            `json:"volatility"`
	Signals    []string                   `json:"signals"`
	Timestamp  time.Time                  `json:"timestamp"`
}

type MultiAnalysisResponse struct {
	Symbol     string                      `json:"symbol"`
	Timeframes map[string]AnalysisResponse `json:"timeframes"`
	Timestamp  time.Time                   `json:"timestamp"`
}

// MultiAnalysisSummary struct is commented out since summary generation is disabled
/*
type MultiAnalysisSummary struct {
	OverallTrend     string   `json:"overall_trend"`     // bullish, bearish, neutral
	ConfidenceScore  float64  `json:"confidence_score"`  // 0-100
	ConsensusSignals []string `json:"consensus_signals"` // signals appearing in multiple timeframes
	RiskLevel        string   `json:"risk_level"`        // low, medium, high
	Recommendation   string   `json:"recommendation"`    // buy, sell, hold
}
*/

// Handler functions
func (s *Server) handleGetPrice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := models.Symbol(vars["symbol"])

	if !utils.ValidateSymbol(string(symbol)) {
		s.sendError(w, http.StatusBadRequest, "Invalid symbol")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	ticker, err := s.binanceClient.GetTicker24hr(ctx, symbol)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch ticker")
		s.sendError(w, http.StatusInternalServerError, "Failed to fetch price data")
		return
	}

	response := PriceResponse{
		Symbol:    string(symbol),
		Price:     ticker.Price,
		Timestamp: ticker.Timestamp,
	}

	s.sendSuccess(w, response)
}

func (s *Server) handleGetTicker(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := models.Symbol(vars["symbol"])

	if !utils.ValidateSymbol(string(symbol)) {
		s.sendError(w, http.StatusBadRequest, "Invalid symbol")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	ticker, err := s.binanceClient.GetTicker24hr(ctx, symbol)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch ticker")
		s.sendError(w, http.StatusInternalServerError, "Failed to fetch ticker data")
		return
	}

	s.sendSuccess(w, ticker)
}

func (s *Server) handleGetKlines(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := models.Symbol(vars["symbol"])

	// Parse query parameters
	interval := r.URL.Query().Get("interval")
	if interval == "" {
		interval = "15m"
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if !utils.ValidateSymbol(string(symbol)) {
		s.sendError(w, http.StatusBadRequest, "Invalid symbol")
		return
	}

	if !utils.ValidateTimeframe(interval) {
		s.sendError(w, http.StatusBadRequest, "Invalid timeframe")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	klines, err := s.binanceClient.GetKlines(ctx, symbol, models.Timeframe(interval), limit)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch klines")
		s.sendError(w, http.StatusInternalServerError, "Failed to fetch klines data")
		return
	}

	s.sendSuccess(w, klines)
}

func (s *Server) handleGetIndicators(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := models.Symbol(vars["symbol"])

	// Parse query parameters
	interval := r.URL.Query().Get("interval")
	if interval == "" {
		interval = "15m"
	}

	if !utils.ValidateSymbol(string(symbol)) {
		s.sendError(w, http.StatusBadRequest, "Invalid symbol")
		return
	}

	if !utils.ValidateTimeframe(interval) {
		s.sendError(w, http.StatusBadRequest, "Invalid timeframe")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Fetch klines data
	klines, err := s.binanceClient.GetKlines(ctx, symbol, models.Timeframe(interval), 100)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch klines")
		s.sendError(w, http.StatusInternalServerError, "Failed to fetch market data")
		return
	}

	// Calculate all RSI periods
	rsiValues := make(map[string]decimal.Decimal)
	for _, period := range s.config.Indicators.RSI.Periods {
		rsi, err := s.calculator.CalculateRSI(klines, period)
		if err == nil {
			rsiValues[fmt.Sprintf("RSI_%d", period)] = rsi
		}
	}

	// Calculate all MA periods
	maValues := make(map[string]decimal.Decimal)
	for _, period := range s.config.Indicators.MA.Periods {
		var ma decimal.Decimal
		switch s.config.Indicators.MA.Type {
		case "EMA":
			ma, err = s.calculator.CalculateEMA(klines, period)
		default:
			ma, err = s.calculator.CalculateSMA(klines, period)
		}
		if err == nil {
			maValues[fmt.Sprintf("MA_%d", period)] = ma
		}
	}

	// Calculate KDJ
	kdj, err := s.calculator.CalculateKDJ(klines, s.config.Indicators.KDJ.KPeriod, s.config.Indicators.KDJ.DPeriod, s.config.Indicators.KDJ.JPeriod)
	if err != nil {
		s.logger.WithError(err).Error("Failed to calculate KDJ")
	}

	response := IndicatorsResponse{
		Symbol:    string(symbol),
		Timeframe: interval,
		RSI:       rsiValues,
		MA:        maValues,
		KDJ:       kdj,
		Timestamp: time.Now(),
	}

	s.sendSuccess(w, response)
}

func (s *Server) handleGetAnalysis(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := models.Symbol(vars["symbol"])

	// Parse query parameters
	interval := r.URL.Query().Get("interval")
	if interval == "" {
		interval = "15m"
	}

	if !utils.ValidateSymbol(string(symbol)) {
		s.sendError(w, http.StatusBadRequest, "Invalid symbol")
		return
	}

	if !utils.ValidateTimeframe(interval) {
		s.sendError(w, http.StatusBadRequest, "Invalid timeframe")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Get comprehensive analysis
	analysis, err := s.getSymbolAnalysis(ctx, symbol, models.Timeframe(interval))
	if err != nil {
		s.logger.WithError(err).Error("Failed to get analysis")
		s.sendError(w, http.StatusInternalServerError, "Failed to perform analysis")
		return
	}

	s.sendSuccess(w, analysis)
}

func (s *Server) handleGetMultiAnalysis(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := models.Symbol(vars["symbol"])

	if !utils.ValidateSymbol(string(symbol)) {
		s.sendError(w, http.StatusBadRequest, "Invalid symbol")
		return
	}

	// Parse optional timeframes query parameter
	timeframesParam := r.URL.Query().Get("timeframes")
	var timeframes []models.Timeframe

	if timeframesParam != "" {
		// Custom timeframes specified
		for _, tf := range strings.Split(timeframesParam, ",") {
			tf = strings.TrimSpace(tf)
			if utils.ValidateTimeframe(tf) {
				timeframes = append(timeframes, models.Timeframe(tf))
			}
		}
	}

	// Default timeframes if none specified or invalid
	if len(timeframes) == 0 {
		timeframes = []models.Timeframe{
			models.Timeframe15m,
			models.Timeframe4h,
			models.Timeframe1d,
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	// Process timeframes concurrently for better performance
	type result struct {
		timeframe string
		analysis  *AnalysisResponse
		err       error
	}

	resultChan := make(chan result, len(timeframes))

	// Launch goroutines for concurrent analysis
	for _, tf := range timeframes {
		go func(timeframe models.Timeframe) {
			analysis, err := s.getSymbolAnalysis(ctx, symbol, timeframe)
			resultChan <- result{
				timeframe: string(timeframe),
				analysis:  analysis,
				err:       err,
			}
		}(tf)
	}

	// Collect results
	analyses := make(map[string]AnalysisResponse)
	for i := 0; i < len(timeframes); i++ {
		res := <-resultChan
		if res.err != nil {
			s.logger.WithError(res.err).WithField("timeframe", res.timeframe).Warn("Failed to get analysis for timeframe")
			continue
		}
		analyses[res.timeframe] = *res.analysis
	}

	// Generate multi-timeframe summary
	// summary := s.generateMultiAnalysisSummary(analyses)

	response := MultiAnalysisResponse{
		Symbol:     string(symbol),
		Timeframes: analyses,
		Timestamp:  time.Now(),
	}

	s.sendSuccess(w, response)
}

func (s *Server) handleGetSignals(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := models.Symbol(vars["symbol"])

	if !utils.ValidateSymbol(string(symbol)) {
		s.sendError(w, http.StatusBadRequest, "Invalid symbol")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Get analysis for default timeframe
	analysis, err := s.getSymbolAnalysis(ctx, symbol, models.Timeframe15m)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get analysis")
		s.sendError(w, http.StatusInternalServerError, "Failed to generate signals")
		return
	}

	s.sendSuccess(w, map[string]interface{}{
		"symbol":    string(symbol),
		"signals":   analysis.Signals,
		"timestamp": time.Now(),
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.sendSuccess(w, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0", // Changed version to test hot reload
	})
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	s.sendSuccess(w, s.config)
}

func (s *Server) handleGetSymbols(w http.ResponseWriter, r *http.Request) {
	s.sendSuccess(w, map[string]interface{}{
		"symbols":   s.config.Symbols,
		"intervals": s.config.Intervals,
		"timestamp": time.Now(),
	})
}

// Missing stream info handler
func (s *Server) handleStreamInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]

	if !utils.ValidateSymbol(symbol) {
		s.sendError(w, http.StatusBadRequest, "Invalid symbol")
		return
	}

	// Return streaming information instead of actual WebSocket
	info := map[string]interface{}{
		"symbol":     symbol,
		"stream_url": fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@ticker", strings.ToLower(symbol)),
		"kline_url":  fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@kline_15m", strings.ToLower(symbol)),
		"supported":  true,
		"timestamp":  time.Now(),
		"message":    "Use WebSocket clients to connect to the stream URLs above",
	}

	s.sendSuccess(w, info)
}

// Helper functions
func (s *Server) getSymbolAnalysis(ctx context.Context, symbol models.Symbol, timeframe models.Timeframe) (*AnalysisResponse, error) {
	// Fetch market data
	klines, err := s.binanceClient.GetKlines(ctx, symbol, timeframe, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch klines: %w", err)
	}

	ticker, err := s.binanceClient.GetTicker24hr(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ticker: %w", err)
	}

	// Calculate indicators
	rsiValues := make(map[string]decimal.Decimal)
	for _, period := range s.config.Indicators.RSI.Periods {
		rsi, err := s.calculator.CalculateRSI(klines, period)
		if err == nil {
			rsiValues[fmt.Sprintf("RSI_%d", period)] = rsi
		}
	}

	maValues := make(map[string]decimal.Decimal)
	for _, period := range s.config.Indicators.MA.Periods {
		var ma decimal.Decimal
		switch s.config.Indicators.MA.Type {
		case "EMA":
			ma, err = s.calculator.CalculateEMA(klines, period)
		default:
			ma, err = s.calculator.CalculateSMA(klines, period)
		}
		if err == nil {
			maValues[fmt.Sprintf("MA_%d", period)] = ma
		}
	}

	kdj, err := s.calculator.CalculateKDJ(klines, s.config.Indicators.KDJ.KPeriod, s.config.Indicators.KDJ.DPeriod, s.config.Indicators.KDJ.JPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate KDJ: %w", err)
	}

	volatility, _ := s.calculator.CalculateVolatility(klines, 20)

	// Generate signals
	signals := s.generateSignals(ticker, rsiValues, maValues, kdj)

	return &AnalysisResponse{
		Symbol:     string(symbol),
		Timeframe:  string(timeframe),
		Price:      ticker,
		RSI:        rsiValues,
		MA:         maValues,
		KDJ:        kdj,
		Volatility: volatility,
		Signals:    signals,
		Timestamp:  time.Now(),
	}, nil
}

func (s *Server) generateSignals(ticker *models.TickerPrice, rsiValues map[string]decimal.Decimal, maValues map[string]decimal.Decimal, kdj models.KDJIndicator) []string {
	var signals []string

	// RSI signals (using RSI_6 for quick signals)
	if rsi6, exists := rsiValues["RSI_6"]; exists {
		if rsi6.LessThan(decimal.NewFromInt(30)) {
			signals = append(signals, "RSI_6_OVERSOLD")
		} else if rsi6.GreaterThan(decimal.NewFromInt(70)) {
			signals = append(signals, "RSI_6_OVERBOUGHT")
		}
	}

	// RSI signals (using RSI_24 for longer term)
	if rsi24, exists := rsiValues["RSI_24"]; exists {
		if rsi24.LessThan(decimal.NewFromInt(35)) {
			signals = append(signals, "RSI_24_OVERSOLD")
		} else if rsi24.GreaterThan(decimal.NewFromInt(65)) {
			signals = append(signals, "RSI_24_OVERBOUGHT")
		}
	}

	// Moving Average signals
	if ma7, exists := maValues["MA_7"]; exists {
		if ticker.Price.GreaterThan(ma7) {
			signals = append(signals, "PRICE_ABOVE_MA7")
		} else {
			signals = append(signals, "PRICE_BELOW_MA7")
		}
	}

	// KDJ signals
	if kdj.K.GreaterThan(kdj.D) && kdj.K.LessThan(decimal.NewFromInt(20)) {
		signals = append(signals, "KDJ_BULLISH_CROSSOVER")
	} else if kdj.K.LessThan(kdj.D) && kdj.K.GreaterThan(decimal.NewFromInt(80)) {
		signals = append(signals, "KDJ_BEARISH_CROSSOVER")
	}

	return signals
}

// Middleware
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.WithFields(logrus.Fields{
			"method":   r.Method,
			"path":     r.URL.Path,
			"duration": time.Since(start),
			"ip":       r.RemoteAddr,
		}).Info("API request")
	})
}

// Response helpers
func (s *Server) sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

func (s *Server) sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   message,
	})
}

// generateMultiAnalysisSummary and helper functions are commented out since summary generation is disabled
/*
// generateMultiAnalysisSummary creates a comprehensive summary from multiple timeframe analyses
func (s *Server) generateMultiAnalysisSummary(analyses map[string]AnalysisResponse) MultiAnalysisSummary {
	if len(analyses) == 0 {
		return MultiAnalysisSummary{
			OverallTrend:     "neutral",
			ConfidenceScore:  0,
			ConsensusSignals: []string{},
			RiskLevel:        "high",
			Recommendation:   "hold",
		}
	}

	// Count signal occurrences across timeframes
	signalCounts := make(map[string]int)
	trendScores := 0 // positive = bullish, negative = bearish
	rsiOversoldCount := 0
	rsiOverboughtCount := 0
	totalTimeframes := len(analyses)

	for timeframe, analysis := range analyses {
		// Count signals
		for _, signal := range analysis.Signals {
			signalCounts[signal]++
		}

		// Analyze RSI trends across timeframes
		for period, rsi := range analysis.RSI {
			if period == "RSI_12" { // Use RSI-12 as primary indicator
				if rsi.LessThan(decimal.NewFromInt(30)) {
					rsiOversoldCount++
					trendScores -= 2 // Strong bearish signal
				} else if rsi.LessThan(decimal.NewFromInt(40)) {
					trendScores -= 1 // Weak bearish signal
				} else if rsi.GreaterThan(decimal.NewFromInt(70)) {
					rsiOverboughtCount++
					trendScores += 2 // Strong bullish signal (potential reversal)
				} else if rsi.GreaterThan(decimal.NewFromInt(60)) {
					trendScores += 1 // Weak bullish signal
				}
			}
		}

		// Check price vs MA signals for trend confirmation
		for _, signal := range analysis.Signals {
			switch signal {
			case "PRICE_ABOVE_MA7", "PRICE_ABOVE_MA25":
				trendScores += 1
			case "PRICE_BELOW_MA7", "PRICE_BELOW_MA25":
				trendScores -= 1
			}
		}

		// Weight longer timeframes more heavily
		switch timeframe {
		case "1d":
			trendScores *= 3 // Daily trend is most important
		case "4h":
			trendScores *= 2 // 4h trend is moderately important
			// 15m keeps weight of 1
		}
	}

	// Determine consensus signals (appearing in multiple timeframes)
	var consensusSignals []string
	for signal, count := range signalCounts {
		if count >= 2 || (count == 1 && totalTimeframes == 1) {
			consensusSignals = append(consensusSignals, signal)
		}
	}

	// Calculate overall trend
	var overallTrend string
	var confidenceScore float64

	if trendScores > 3 {
		overallTrend = "bullish"
		confidenceScore = min(90, float64(50+trendScores*5))
	} else if trendScores < -3 {
		overallTrend = "bearish"
		confidenceScore = min(90, float64(50+abs(trendScores)*5))
	} else {
		overallTrend = "neutral"
		confidenceScore = 30 + float64(abs(trendScores)*5)
	}

	// Adjust confidence based on signal consensus
	consensusStrength := float64(len(consensusSignals)) / float64(max(1, totalTimeframes))
	confidenceScore += consensusStrength * 20

	// Cap confidence score
	confidenceScore = min(95, confidenceScore)

	// Determine risk level
	var riskLevel string
	volatilityCount := 0
	for _, analysis := range analyses {
		if analysis.Volatility.GreaterThan(decimal.NewFromFloat(0.05)) { // 5% volatility threshold
			volatilityCount++
		}
	}

	if volatilityCount >= totalTimeframes/2 {
		riskLevel = "high"
	} else if len(consensusSignals) <= 1 {
		riskLevel = "medium"
	} else {
		riskLevel = "low"
	}

	// Generate recommendation
	var recommendation string
	if overallTrend == "bullish" && confidenceScore > 70 && riskLevel != "high" {
		recommendation = "buy"
	} else if overallTrend == "bearish" && confidenceScore > 70 && riskLevel != "high" {
		recommendation = "sell"
	} else {
		recommendation = "hold"
	}

	return MultiAnalysisSummary{
		OverallTrend:     overallTrend,
		ConfidenceScore:  confidenceScore,
		ConsensusSignals: consensusSignals,
		RiskLevel:        riskLevel,
		Recommendation:   recommendation,
	}
}

// Helper functions for summary calculation
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
*/

// Start starts the API server
func (s *Server) Start(port string) error {
	s.logger.WithField("port", port).Info("Starting API server")
	return http.ListenAndServe(":"+port, s.router)
}
