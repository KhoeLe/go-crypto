package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-crypto/internal/binance"
	"go-crypto/internal/config"
	"go-crypto/internal/indicators"
	"go-crypto/internal/models"
	"go-crypto/internal/ratelimit"
	"go-crypto/pkg/utils"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// nowGMT7 returns the current time in GMT+7 (Asia/Bangkok) timezone
func nowGMT7() models.GMTPlus7Time {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	return models.NewGMTPlus7Time(time.Now().In(loc))
}

// Server represents the API server
type Server struct {
	router        *mux.Router
	binanceClient *binance.Client
	calculator    *indicators.Calculator
	config        *config.Config
	logger        *logrus.Logger
	rateLimiter   *ratelimit.RateLimiter
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, logger *logrus.Logger) *Server {
	s := &Server{
		router:        mux.NewRouter(),
		binanceClient: binance.NewClient(&cfg.Binance, logger),
		calculator:    indicators.NewCalculator(),
		config:        cfg,
		logger:        logger,
		rateLimiter:   ratelimit.NewRateLimiter(&cfg.RateLimit, logger),
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

	// Enhanced analysis endpoint with new features
	api.HandleFunc("/enhanced-analysis/{symbol}", s.handleGetEnhancedAnalysis).Methods("GET")

	// Real-time endpoints
	api.HandleFunc("/stream/{symbol}", s.handleStreamInfo).Methods("GET")

	// Health and info endpoints
	api.HandleFunc("/health", s.handleHealth).Methods("GET")
	api.HandleFunc("/config", s.handleGetConfig).Methods("GET")
	api.HandleFunc("/symbols", s.handleGetSymbols).Methods("GET")
	api.HandleFunc("/rate-limit-status", s.handleGetRateLimitStatus).Methods("GET")

	// CORS middleware
	s.router.Use(s.corsMiddleware)
	s.router.Use(s.rateLimitMiddleware)
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
	Symbol    string              `json:"symbol"`
	Price     decimal.Decimal     `json:"price"`
	Timestamp models.GMTPlus7Time `json:"timestamp"`
}

type IndicatorsResponse struct {
	Symbol    string                     `json:"symbol"`
	Timeframe string                     `json:"timeframe"`
	RSI       map[string]decimal.Decimal `json:"rsi"`
	MA        map[string]decimal.Decimal `json:"ma"`
	KDJ       models.KDJIndicator        `json:"kdj"`
	Timestamp models.GMTPlus7Time        `json:"timestamp"`
}

type AnalysisResponse struct {
	Symbol          string                     `json:"symbol"`
	Timeframe       string                     `json:"timeframe"`
	Price           *models.TickerPrice        `json:"price"`
	RSI             map[string]decimal.Decimal `json:"rsi"`
	MA              map[string]decimal.Decimal `json:"ma"`
	KDJ             models.KDJIndicator        `json:"kdj"`
	MACD            models.MACDIndicator       `json:"macd"`
	Volatility      decimal.Decimal            `json:"volatility"`
	MarketSentiment string                     `json:"market_sentiment"`
	Signals         []string                   `json:"signals"`
	Timestamp       models.GMTPlus7Time        `json:"timestamp"`
}

type MultiAnalysisResponse struct {
	Symbol     string                      `json:"symbol"`
	Timeframes map[string]AnalysisResponse `json:"timeframes"`
	Timestamp  models.GMTPlus7Time         `json:"timestamp"`
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

// isBinanceInvalidSymbolError checks if error is a Binance invalid symbol error
func isBinanceInvalidSymbolError(err error) bool {
	if binanceErr, ok := err.(*binance.BinanceError); ok {
		return binanceErr.IsInvalidSymbol()
	}
	return false
}

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
		if isBinanceInvalidSymbolError(err) {
			s.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid symbol: %s. Please check the symbol name.", string(symbol)))
			return
		}
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
	limit := 50
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
	klines, err := s.binanceClient.GetKlines(ctx, symbol, models.Timeframe(interval), 25)
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
		Timestamp: nowGMT7(),
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

	// Default timeframes - focus on 15m as requested
	if len(timeframes) == 0 {
		timeframes = []models.Timeframe{
			models.Timeframe15m, // Primary focus as requested
			models.Timeframe4h,
			models.Timeframe1d,
		}
	}

	// Check if enhanced analysis is requested (query parameter)
	enhancedParam := r.URL.Query().Get("enhanced")
	useEnhanced := enhancedParam == "true" || enhancedParam == "1"

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	if useEnhanced {
		// Mixed analysis: enhanced for 15m, standard for others
		type mixedResult struct {
			timeframe string
			analysis  interface{} // Can be either EnhancedAnalysisResponse or AnalysisResponse
			err       error
		}

		mixedResultChan := make(chan mixedResult, len(timeframes))

		// Launch goroutines for mixed analysis
		for _, tf := range timeframes {
			go func(timeframe models.Timeframe) {
				if string(timeframe) == "15m" {
					// Enhanced analysis for 15m timeframe
					analysis, err := s.getEnhancedSymbolAnalysis(ctx, symbol, timeframe)
					mixedResultChan <- mixedResult{
						timeframe: string(timeframe),
						analysis:  analysis,
						err:       err,
					}
				} else {
					// Standard analysis for 4h and 1d timeframes
					analysis, err := s.getSymbolAnalysis(ctx, symbol, timeframe)
					mixedResultChan <- mixedResult{
						timeframe: string(timeframe),
						analysis:  analysis,
						err:       err,
					}
				}
			}(tf)
		}

		// Collect mixed results
		mixedAnalyses := make(map[string]interface{})
		for i := 0; i < len(timeframes); i++ {
			res := <-mixedResultChan
			if res.err != nil {
				s.logger.WithError(res.err).WithField("timeframe", res.timeframe).Warn("Failed to get analysis for timeframe")
				continue
			}

			// Sort klines in descending order and add proper timestamps for enhanced analysis
			if enhancedAnalysis, ok := res.analysis.(*models.EnhancedAnalysisResponse); ok {
				s.sortKlinesDescending(enhancedAnalysis.Klines)
				enhancedAnalysis.Timestamp = nowGMT7()

				// Debug the divergence signals
				s.logger.WithFields(logrus.Fields{
					"timeframe":               res.timeframe,
					"divergenceSignalsLength": len(enhancedAnalysis.Historical.DivergenceSignals),
					"rsiHistoryLength":        len(enhancedAnalysis.Historical.RSIHistory),
					"moneyFlowHistoryLength":  len(enhancedAnalysis.Historical.MoneyFlowHistory),
				}).Info("Enhanced analysis divergence signals debug")

				// Debug the divergence signals further if they're empty
				if len(enhancedAnalysis.Historical.DivergenceSignals) == 0 {
					// Attempt to recalculate divergence signals if we have enough historical data
					if len(enhancedAnalysis.Historical.RSIHistory) > 0 && len(enhancedAnalysis.Historical.MoneyFlowHistory) > 0 {
						fmt.Printf("[DEBUG] Recalculating divergence signals in multi-analysis for timeframe %s\n", res.timeframe)
						enhancedAnalysis.Historical.DivergenceSignals = s.calculator.DetectMoneyFlowDivergence(
							enhancedAnalysis.Klines,
							enhancedAnalysis.Historical.RSIHistory,
							enhancedAnalysis.Historical.MoneyFlowHistory,
						)
						s.logger.WithField("recalculatedSignalsLength", len(enhancedAnalysis.Historical.DivergenceSignals)).Info("Recalculated divergence signals")

						// If we found divergence signals, add them to the enhanced signals list
						if len(enhancedAnalysis.Historical.DivergenceSignals) > 0 {
							for _, divSignal := range enhancedAnalysis.Historical.DivergenceSignals {
								var signalText string
								if divSignal.Type == "bullish" {
									signalText = "BULLISH_DIVERGENCE_DETECTED"
								} else if divSignal.Type == "bearish" {
									signalText = "BEARISH_DIVERGENCE_DETECTED"
								} else if divSignal.Type == "partial" {
									if divSignal.RSITrend == "higher_lows" || divSignal.MFITrend == "higher_lows" {
										signalText = "PARTIAL_BULLISH_DIVERGENCE"
									} else if divSignal.RSITrend == "lower_highs" || divSignal.MFITrend == "lower_highs" {
										signalText = "PARTIAL_BEARISH_DIVERGENCE"
									}
								}

								if signalText != "" && !containsSignal(enhancedAnalysis.Signals, signalText) {
									enhancedAnalysis.Signals = append(enhancedAnalysis.Signals, signalText)
								}
							}
						}
					}
				}
			}

			mixedAnalyses[res.timeframe] = res.analysis
		}

		if len(mixedAnalyses) == 0 {
			s.sendError(w, http.StatusInternalServerError, "Failed to get analysis for any timeframe")
			return
		}

		response := map[string]interface{}{
			"symbol":     string(symbol),
			"timeframes": mixedAnalyses,
			"timestamp":  nowGMT7(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Standard analysis (existing behavior)
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
		Timestamp:  nowGMT7(),
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
		"timestamp": nowGMT7(),
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.sendSuccess(w, map[string]interface{}{
		"status":    "healthy",
		"timestamp": nowGMT7(),
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
		"timestamp": nowGMT7(),
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
		"timestamp":  nowGMT7(),
		"message":    "Use WebSocket clients to connect to the stream URLs above",
	}

	s.sendSuccess(w, info)
}

// handleGetEnhancedAnalysis handles enhanced analysis requests with new features
func (s *Server) handleGetEnhancedAnalysis(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := models.Symbol(vars["symbol"])

	// Parse query parameters
	interval := r.URL.Query().Get("interval")
	if interval == "" {
		interval = "15m" // Default to 15m as requested
	}

	if !utils.ValidateSymbol(string(symbol)) {
		s.sendError(w, http.StatusBadRequest, "Invalid symbol")
		return
	}

	if !utils.ValidateTimeframe(interval) {
		s.sendError(w, http.StatusBadRequest, "Invalid timeframe")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// Get enhanced analysis with all new features
	analysis, err := s.getEnhancedSymbolAnalysis(ctx, symbol, models.Timeframe(interval))
	if err != nil {
		s.logger.WithError(err).Error("Failed to get enhanced analysis")
		s.sendError(w, http.StatusInternalServerError, "Failed to perform enhanced analysis")
		return
	}

	s.sendSuccess(w, analysis)
}

// Helper functions
func (s *Server) getSymbolAnalysis(ctx context.Context, symbol models.Symbol, timeframe models.Timeframe) (*AnalysisResponse, error) {
	// Fetch market data
	klines, err := s.binanceClient.GetKlines(ctx, symbol, timeframe, 25)
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

	// Calculate MACD (5, 10, 3 parameters work better with limited data)
	macd, err := s.calculator.CalculateMACD(klines, 5, 10, 3)
	if err != nil {
		// Use zero values if MACD calculation fails
		macd = models.MACDIndicator{}
	}

	// Generate market sentiment
	marketSentiment := s.calculator.GenerateMarketSentiment(rsiValues, macd, kdj, ticker.PriceChangePercent)

	// Generate signals
	signals := s.generateSignals(ticker, rsiValues, maValues, kdj)

	return &AnalysisResponse{
		Symbol:          string(symbol),
		Timeframe:       string(timeframe),
		Price:           ticker,
		RSI:             rsiValues,
		MA:              maValues,
		KDJ:             kdj,
		MACD:            macd,
		Volatility:      volatility,
		MarketSentiment: marketSentiment,
		Signals:         signals,
		Timestamp:       nowGMT7(),
	}, nil
}

// getEnhancedSymbolAnalysis performs enhanced analysis with new features
func (s *Server) getEnhancedSymbolAnalysis(ctx context.Context, symbol models.Symbol, timeframe models.Timeframe) (*models.EnhancedAnalysisResponse, error) {
	// Fetch market data with limit of 50 for calculation, but only return 25 in response
	limit := 50
	klines, err := s.binanceClient.GetKlines(ctx, symbol, timeframe, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch klines: %w", err)
	}

	ticker, err := s.binanceClient.GetTicker24hr(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ticker: %w", err)
	}

	// Calculate traditional indicators
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

	// Calculate MACD (5, 10, 3 parameters work better with limited data)
	macd, err := s.calculator.CalculateMACD(klines, 5, 10, 3)
	if err != nil {
		// Use zero values if MACD calculation fails
		macd = models.MACDIndicator{}
	}

	// Calculate enhanced features
	// 1. Money Flow Analysis
	moneyFlow, err := s.calculator.CalculateMoneyFlowIndex(klines, 14)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to calculate money flow index")
		moneyFlow = models.MoneyFlowIndicator{
			MoneyFlowIndex: decimal.NewFromInt(50), // Neutral default
			Timestamp:      klines[len(klines)-1].CloseTime,
		}
	}

	// 2. Volume Breakout Detection
	volumeBreakout, err := s.calculator.DetectVolumeBreakout(klines, 20)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to detect volume breakout")
		volumeBreakout = models.VolumeBreakout{
			IsBreakout:        false,
			BreakoutDirection: "neutral",
			Timestamp:         klines[len(klines)-1].CloseTime,
		}
	}

	// 3. Volume Delta Analysis (Buy vs Sell pressure)
	volumeDelta, err := s.calculator.CalculateVolumeDelta(klines)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to calculate volume delta")
		volumeDelta = models.VolumeDelta{
			BuyVolume:    decimal.Zero,
			SellVolume:   decimal.Zero,
			Delta:        decimal.Zero,
			DeltaPercent: decimal.Zero,
			Pressure:     "balanced",
			Strength:     1,
			Timestamp:    klines[len(klines)-1].CloseTime,
		}
	}

	// 4. Whale Volume Spike Detection
	whaleActivity, err := s.calculator.CalculateWhaleVolumeSpike(klines, ticker.Price)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to calculate whale volume spike")
		whaleActivity = models.WhaleVolumeSpike{
			IsWhaleSpike:     false,
			SpikeVolume:      decimal.Zero,
			SpikeValueUSDT:   decimal.Zero,
			ThresholdUSDT:    decimal.NewFromInt(100000),
			VolumeMultiplier: decimal.NewFromInt(1),
			Timestamp:        klines[len(klines)-1].CloseTime,
		}
	}

	// 5. Historical Indicators with precise counts
	// RSI History: 5 entries, MA History: 5 entries, Money Flow History: 15 entries
	const (
		expectedRSIHistoryCount = 5
		expectedMAHistoryCount  = 5
		expectedMoneyFlowCount  = 15
	)

	rsiHistorical, err := s.calculator.CalculateHistoricalIndicators(klines, s.config.Indicators.RSI.Periods, []int{}, s.config.Indicators.MA.Type, expectedRSIHistoryCount)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to calculate RSI historical indicators")
		rsiHistorical = models.HistoricalIndicators{
			RSIHistory: []models.RSIHistoryPoint{},
			MAHistory:  []models.MAHistoryPoint{},
		}
	}

	// Debug the MA history count
	s.logger.WithField("expectedMAHistoryCount", expectedMAHistoryCount).Info("Calculating MA history")

	maHistorical, err := s.calculator.CalculateHistoricalIndicators(klines, []int{}, s.config.Indicators.MA.Periods, s.config.Indicators.MA.Type, expectedMAHistoryCount)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to calculate MA historical indicators")
		maHistorical = models.HistoricalIndicators{
			RSIHistory: []models.RSIHistoryPoint{},
			MAHistory:  []models.MAHistoryPoint{},
		}
	}

	// Debug the actual MA history length
	s.logger.WithField("actualMAHistoryLength", len(maHistorical.MAHistory)).Info("MA history calculated")

	// Calculate Money Flow History
	moneyFlowHistory, err := s.calculator.CalculateHistoricalMoneyFlow(klines, 14, expectedMoneyFlowCount)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to calculate historical money flow")
		moneyFlowHistory = []models.MoneyFlowIndicator{}
	}

	// Debug the money flow history
	s.logger.WithField("moneyFlowHistoryLength", len(moneyFlowHistory)).Info("Money flow history calculated")
	fmt.Printf("[DEBUG] Money Flow History length: %d, expected: %d\n", len(moneyFlowHistory), expectedMoneyFlowCount)

	// Detect divergence signals between RSI and Money Flow
	// Initialize with empty array to avoid null in JSON response
	divergenceSignals := make([]models.DivergenceSignal, 0)

	if len(rsiHistorical.RSIHistory) > 0 && len(moneyFlowHistory) > 0 {
		// The calculator now returns properly formatted models.DivergenceSignal objects
		divergenceSignals = s.calculator.DetectMoneyFlowDivergence(klines, rsiHistorical.RSIHistory, moneyFlowHistory)
	}

	// Ensure divergenceSignals is never null
	if divergenceSignals == nil {
		divergenceSignals = []models.DivergenceSignal{}
	}

	// Combine historical indicators
	historical := models.HistoricalIndicators{
		RSIHistory:        rsiHistorical.RSIHistory,
		MAHistory:         maHistorical.MAHistory,
		MoneyFlowHistory:  moneyFlowHistory,
		DivergenceSignals: divergenceSignals,
	}

	// Generate market sentiment
	marketSentiment := s.calculator.GenerateMarketSentiment(rsiValues, macd, kdj, ticker.PriceChangePercent)

	// Generate enhanced signals including pump detection
	signals := s.generateEnhancedSignals(ticker, rsiValues, maValues, kdj, moneyFlow, volumeBreakout, volumeDelta, whaleActivity)

	// Return only the last 15 klines in the response (but use all 50 for calculations)
	responseKlines := klines
	if len(klines) > 15 {
		responseKlines = klines[len(klines)-15:]
	}

	return &models.EnhancedAnalysisResponse{
		Symbol:          string(symbol),
		Timeframe:       string(timeframe),
		Price:           ticker,
		Klines:          responseKlines, // Only include last 25 klines in response
		RSI:             rsiValues,
		MA:              maValues,
		KDJ:             kdj,
		MACD:            macd,
		Volatility:      volatility,
		MarketSentiment: marketSentiment,
		MoneyFlow:       moneyFlow,
		VolumeBreakout:  volumeBreakout,
		VolumeDelta:     volumeDelta,   // New: Buy vs sell pressure analysis
		WhaleActivity:   whaleActivity, // New: Whale volume spike detection
		Historical:      historical,
		Signals:         signals,
		Timestamp:       nowGMT7(),
	}, nil
}

// generateEnhancedSignals creates trading signals including new money flow, volume analysis, and pump detection
func (s *Server) generateEnhancedSignals(ticker *models.TickerPrice, rsiValues map[string]decimal.Decimal, maValues map[string]decimal.Decimal, kdj models.KDJIndicator, moneyFlow models.MoneyFlowIndicator, volumeBreakout models.VolumeBreakout, volumeDelta models.VolumeDelta, whaleActivity models.WhaleVolumeSpike) []string {
	var signals []string

	// Traditional RSI signals
	if rsi6, exists := rsiValues["RSI_6"]; exists {
		if rsi6.LessThan(decimal.NewFromInt(30)) {
			signals = append(signals, "RSI_6_OVERSOLD")
		} else if rsi6.GreaterThan(decimal.NewFromInt(70)) {
			signals = append(signals, "RSI_6_OVERBOUGHT")
		}
	}

	if rsi12, exists := rsiValues["RSI_12"]; exists {
		if rsi12.LessThan(decimal.NewFromInt(30)) {
			signals = append(signals, "RSI_12_OVERSOLD")
		} else if rsi12.GreaterThan(decimal.NewFromInt(70)) {
			signals = append(signals, "RSI_12_OVERBOUGHT")
		}
	}

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

	if ma25, exists := maValues["MA_25"]; exists {
		if ticker.Price.GreaterThan(ma25) {
			signals = append(signals, "PRICE_ABOVE_MA25")
		} else {
			signals = append(signals, "PRICE_BELOW_MA25")
		}
	}

	// KDJ signals
	if kdj.K.GreaterThan(kdj.D) && kdj.K.LessThan(decimal.NewFromInt(20)) {
		signals = append(signals, "KDJ_BULLISH_CROSSOVER")
	} else if kdj.K.LessThan(kdj.D) && kdj.K.GreaterThan(decimal.NewFromInt(80)) {
		signals = append(signals, "KDJ_BEARISH_CROSSOVER")
	}

	// Enhanced Money Flow signals
	if moneyFlow.MoneyFlowIndex.LessThan(decimal.NewFromInt(20)) {
		signals = append(signals, "MONEY_FLOW_OVERSOLD")
	} else if moneyFlow.MoneyFlowIndex.GreaterThan(decimal.NewFromInt(80)) {
		signals = append(signals, "MONEY_FLOW_OVERBOUGHT")
	}

	// Money flow change signals
	if moneyFlow.MoneyFlowChange.GreaterThan(decimal.NewFromInt(10)) {
		signals = append(signals, "MONEY_FLOW_INCREASING")
	} else if moneyFlow.MoneyFlowChange.LessThan(decimal.NewFromInt(-10)) {
		signals = append(signals, "MONEY_FLOW_DECREASING")
	}

	// Volume breakout signals
	if volumeBreakout.IsBreakout {
		signals = append(signals, "VOLUME_BREAKOUT_DETECTED")

		if volumeBreakout.BreakoutDirection == "bullish" {
			signals = append(signals, "VOLUME_BREAKOUT_BULLISH")
		} else if volumeBreakout.BreakoutDirection == "bearish" {
			signals = append(signals, "VOLUME_BREAKOUT_BEARISH")
		}

		// High strength breakout
		if volumeBreakout.BreakoutStrength.GreaterThan(decimal.NewFromInt(7)) {
			signals = append(signals, "STRONG_VOLUME_BREAKOUT")
		}
	}

	// Volume multiplier signals
	if volumeBreakout.VolumeMultiplier.GreaterThan(decimal.NewFromFloat(2.0)) {
		signals = append(signals, "HIGH_VOLUME_ACTIVITY")
	} else if volumeBreakout.VolumeMultiplier.LessThan(decimal.NewFromFloat(0.5)) {
		signals = append(signals, "LOW_VOLUME_ACTIVITY")
	}

	// Volume Delta signals (Buy vs Sell pressure)
	if volumeDelta.Pressure == "buy_pressure" {
		signals = append(signals, "BUY_PRESSURE_DETECTED")
		if volumeDelta.Strength >= 7 {
			signals = append(signals, "STRONG_BUY_PRESSURE")
		}
	} else if volumeDelta.Pressure == "sell_pressure" {
		signals = append(signals, "SELL_PRESSURE_DETECTED")
		if volumeDelta.Strength >= 7 {
			signals = append(signals, "STRONG_SELL_PRESSURE")
		}
	}

	// Whale Activity signals
	if whaleActivity.IsWhaleSpike {
		signals = append(signals, "WHALE_ACTIVITY_DETECTED")
		if whaleActivity.SpikeValueUSDT.GreaterThan(decimal.NewFromInt(500000)) {
			signals = append(signals, "LARGE_WHALE_ACTIVITY")
		}
	}

	// Pump Signal Detection (combining multiple indicators)
	if s.calculator.DetectPumpSignal(rsiValues, moneyFlow, volumeDelta, volumeBreakout) {
		signals = append(signals, "PUMP_SIGNAL_DETECTED")
	}

	return signals
}

// generateSignals creates basic trading signals (traditional method)
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

	if ma25, exists := maValues["MA_25"]; exists {
		if ticker.Price.GreaterThan(ma25) {
			signals = append(signals, "PRICE_ABOVE_MA25")
		} else {
			signals = append(signals, "PRICE_BELOW_MA25")
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

// Helper to check if a signal is already in the list
func containsSignal(signals []string, signal string) bool {
	for _, s := range signals {
		if s == signal {
			return true
		}
	}
	return false
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

func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for health checks and static files
		if strings.HasPrefix(r.URL.Path, "/health") ||
			strings.HasPrefix(r.URL.Path, "/static") ||
			!strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		// Get client IP
		clientIP := ratelimit.GetClientIP(r)

		// Check rate limit
		allowed, status := s.rateLimiter.IsAllowed(clientIP)

		// Add rate limit headers to response
		for key, value := range status.ToHeaders() {
			w.Header().Set(key, value)
		}

		if !allowed {
			// Log rate limit violation
			s.logger.WithFields(logrus.Fields{
				"ip":             clientIP,
				"path":           r.URL.Path,
				"method":         r.Method,
				"tier":           status.Tier,
				"remaining_min":  status.RemainingMinute,
				"remaining_hour": status.RemainingHour,
				"blocked_until":  status.BlockedUntil,
			}).Warn("Rate limit exceeded")

			// Return rate limit error
			errorMessage := "Rate limit exceeded"
			if !status.BlockedUntil.IsZero() {
				errorMessage = fmt.Sprintf("Rate limit exceeded. Blocked until %s", status.BlockedUntil.Format(time.RFC3339))
			}

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", time.Until(status.ResetTimeMinute).Seconds()))
			w.WriteHeader(http.StatusTooManyRequests)

			json.NewEncoder(w).Encode(APIResponse{
				Success: false,
				Error:   errorMessage,
				Data: map[string]interface{}{
					"rate_limit_status": status,
					"retry_after":       time.Until(status.ResetTimeMinute).Seconds(),
				},
			})
			return
		}

		// Log successful rate limit check (debug level)
		s.logger.WithFields(logrus.Fields{
			"ip":             clientIP,
			"tier":           status.Tier,
			"remaining_min":  status.RemainingMinute,
			"remaining_hour": status.RemainingHour,
			"request_count":  status.RequestCount,
		}).Debug("Rate limit check passed")

		next.ServeHTTP(w, r)
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
	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Return the server so it can be started
	return srv.ListenAndServe()
}

// Stop gracefully stops the server and cleans up resources
func (s *Server) Stop() {
	if s.rateLimiter != nil {
		s.rateLimiter.Stop()
	}
}

// Public wrapper methods for Lambda handlers

// GetPrice fetches real price data for a symbol
func (s *Server) GetPrice(ctx context.Context, symbol string) (*PriceResponse, error) {
	if !utils.ValidateSymbol(symbol) {
		return nil, fmt.Errorf("invalid symbol: %s", symbol)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	ticker, err := s.binanceClient.GetTicker24hr(timeoutCtx, models.Symbol(symbol))
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch ticker")
		return nil, fmt.Errorf("failed to fetch price data: %w", err)
	}

	response := &PriceResponse{
		Symbol:    symbol,
		Price:     ticker.Price,
		Timestamp: ticker.Timestamp,
	}

	return response, nil
}

// GetAnalysis performs real technical analysis for a symbol and timeframe
func (s *Server) GetAnalysis(ctx context.Context, symbol, interval string) (*AnalysisResponse, error) {
	if !utils.ValidateSymbol(symbol) {
		return nil, fmt.Errorf("invalid symbol: %s", symbol)
	}

	if !utils.ValidateTimeframe(interval) {
		return nil, fmt.Errorf("invalid timeframe: %s", interval)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Get comprehensive analysis using existing method
	analysis, err := s.getSymbolAnalysis(timeoutCtx, models.Symbol(symbol), models.Timeframe(interval))
	if err != nil {
		s.logger.WithError(err).Error("Failed to get analysis")
		return nil, fmt.Errorf("failed to perform analysis: %w", err)
	}

	return analysis, nil
}

// GetMultiAnalysis performs multi-timeframe analysis for a symbol
func (s *Server) GetMultiAnalysis(ctx context.Context, symbol string, timeframes []string) (*MultiAnalysisResponse, error) {
	if !utils.ValidateSymbol(symbol) {
		return nil, fmt.Errorf("invalid symbol: %s", symbol)
	}

	// Default timeframes if none provided
	if len(timeframes) == 0 {
		timeframes = []string{"15m", "4h", "1d"}
	}

	// Validate all timeframes
	for _, tf := range timeframes {
		if !utils.ValidateTimeframe(tf) {
			return nil, fmt.Errorf("invalid timeframe: %s", tf)
		}
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Perform analysis for each timeframe concurrently
	type result struct {
		timeframe string
		analysis  *AnalysisResponse
		err       error
	}

	resultChan := make(chan result, len(timeframes))

	for _, tf := range timeframes {
		go func(timeframe string) {
			analysis, err := s.getSymbolAnalysis(timeoutCtx, models.Symbol(symbol), models.Timeframe(timeframe))
			resultChan <- result{timeframe: timeframe, analysis: analysis, err: err}
		}(tf)
	}

	// Collect results
	analyses := make(map[string]AnalysisResponse)
	for i := 0; i < len(timeframes); i++ {
		res := <-resultChan
		if res.err != nil {
			s.logger.WithError(res.err).WithField("timeframe", res.timeframe).Error("Failed to get analysis for timeframe")
			continue // Skip failed timeframes instead of failing entirely
		}
		analyses[res.timeframe] = *res.analysis
	}

	if len(analyses) == 0 {
		return nil, fmt.Errorf("failed to get analysis for any timeframe")
	}

	response := &MultiAnalysisResponse{
		Symbol:     symbol,
		Timeframes: analyses,
		Timestamp:  nowGMT7(),
	}

	return response, nil
}

// GetEnhancedMultiAnalysis performs multi-timeframe analysis with enhanced features for 15m and basic for others
func (s *Server) GetEnhancedMultiAnalysis(ctx context.Context, symbol string, timeframes []string) (map[string]interface{}, error) {
	if !utils.ValidateSymbol(symbol) {
		return nil, fmt.Errorf("invalid symbol: %s", symbol)
	}

	// Default timeframes if none provided
	if len(timeframes) == 0 {
		timeframes = []string{"15m", "4h", "1d"}
	}

	// Validate all timeframes
	for _, tf := range timeframes {
		if !utils.ValidateTimeframe(tf) {
			return nil, fmt.Errorf("invalid timeframe: %s", tf)
		}
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	// Mixed analysis: enhanced for 15m, standard for others
	type mixedResult struct {
		timeframe string
		analysis  interface{} // Can be either EnhancedAnalysisResponse or AnalysisResponse
		err       error
	}

	mixedResultChan := make(chan mixedResult, len(timeframes))

	// Launch goroutines for mixed analysis
	for _, tf := range timeframes {
		go func(timeframe string) {
			if timeframe == "15m" {
				// Enhanced analysis for 15m timeframe
				analysis, err := s.getEnhancedSymbolAnalysis(timeoutCtx, models.Symbol(symbol), models.Timeframe(timeframe))
				mixedResultChan <- mixedResult{
					timeframe: timeframe,
					analysis:  analysis,
					err:       err,
				}
			} else {
				// Standard analysis for 4h and 1d timeframes
				analysis, err := s.getSymbolAnalysis(timeoutCtx, models.Symbol(symbol), models.Timeframe(timeframe))
				mixedResultChan <- mixedResult{
					timeframe: timeframe,
					analysis:  analysis,
					err:       err,
				}
			}
		}(tf)
	}

	// Collect mixed results
	mixedAnalyses := make(map[string]interface{})
	for i := 0; i < len(timeframes); i++ {
		res := <-mixedResultChan
		if res.err != nil {
			s.logger.WithError(res.err).WithField("timeframe", res.timeframe).Warn("Failed to get analysis for timeframe")
			continue
		}

		// Sort klines in descending order and add proper timestamps for enhanced analysis
		if enhancedAnalysis, ok := res.analysis.(*models.EnhancedAnalysisResponse); ok {
			s.sortKlinesDescending(enhancedAnalysis.Klines)
			enhancedAnalysis.Timestamp = nowGMT7()
		}

		mixedAnalyses[res.timeframe] = res.analysis
	}

	if len(mixedAnalyses) == 0 {
		return nil, fmt.Errorf("failed to get analysis for any timeframe")
	}

	response := map[string]interface{}{
		"symbol":     symbol,
		"timeframes": mixedAnalyses,
		"timestamp":  nowGMT7(),
	}

	return response, nil
}

// sortKlinesDescending sorts klines by close time in descending order (newest first)
func (s *Server) sortKlinesDescending(klines []models.Kline) {
	sort.Slice(klines, func(i, j int) bool {
		return klines[i].CloseTime.Time.After(klines[j].CloseTime.Time)
	})
}

// GetRateLimitStatus returns the current rate limit status for a client IP
func (s *Server) GetRateLimitStatus(ip string) *ratelimit.RateLimitStatus {
	_, status := s.rateLimiter.IsAllowed(ip)
	return status
}

// GetHandler returns the router for server configuration
func (s *Server) GetHandler() http.Handler {
	return s.router
}

// handleGetRateLimitStatus handles rate limit status request for the current client
func (s *Server) handleGetRateLimitStatus(w http.ResponseWriter, r *http.Request) {
	// Get client IP
	clientIP := ratelimit.GetClientIP(r)

	// Get rate limit status for the client
	_, status := s.rateLimiter.IsAllowed(clientIP)

	// Add rate limit headers to response
	for key, value := range status.ToHeaders() {
		w.Header().Set(key, value)
	}

	// Get tier information
	tierInfo := s.config.RateLimit.Tiers[status.Tier]

	// Create response with tier details and client status
	response := map[string]interface{}{
		"ip":               clientIP,
		"allowed":          status.Allowed,
		"tier":             status.Tier,
		"remaining_minute": status.RemainingMinute,
		"remaining_hour":   status.RemainingHour,
		"reset_minute":     status.ResetTimeMinute,
		"reset_hour":       status.ResetTimeHour,
		"request_count":    status.RequestCount,
		"burst_tokens":     status.BurstTokens,
		"tier_limits": map[string]interface{}{
			"requests_per_minute":    tierInfo.RequestsPerMinute,
			"requests_per_hour":      tierInfo.RequestsPerHour,
			"burst_allowance":        tierInfo.BurstAllowance,
			"block_duration_minutes": tierInfo.BlockDurationMinutes,
		},
		"rate_limit_enabled": s.config.RateLimit.Enabled,
		"timestamp":          nowGMT7(),
	}

	// Add blocked status if applicable
	if !status.BlockedUntil.IsZero() {
		response["blocked_until"] = status.BlockedUntil
		response["blocked"] = true
	}

	s.sendSuccess(w, response)
}
