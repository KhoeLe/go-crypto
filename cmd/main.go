package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go-crypto/internal/binance"
	"go-crypto/internal/config"
	"go-crypto/internal/indicators"
	"go-crypto/internal/models"
	"go-crypto/pkg/utils"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Starting Go Crypto Trading Analysis Application")

	// Load configuration
	cfg := config.DefaultConfig()

	// Initialize Binance client
	binanceClient := binance.NewClient(&cfg.Binance, logger)

	// Initialize indicator calculator
	calc := indicators.NewCalculator()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal")
		cancel()
	}()

	// Convert string symbols to models.Symbol
	symbols := make([]models.Symbol, len(cfg.Symbols))
	for i, symbolStr := range cfg.Symbols {
		if !utils.ValidateSymbol(symbolStr) {
			logger.WithField("symbol", symbolStr).Warn("Invalid symbol, skipping")
			continue
		}
		symbols[i] = models.Symbol(symbolStr)
	}

	// Convert string timeframes to models.Timeframe
	timeframes := make([]models.Timeframe, len(cfg.Intervals))
	for i, intervalStr := range cfg.Intervals {
		if !utils.ValidateTimeframe(intervalStr) {
			logger.WithField("timeframe", intervalStr).Warn("Invalid timeframe, skipping")
			continue
		}
		timeframes[i] = models.Timeframe(intervalStr)
	}

	// Start data collection and analysis
	var wg sync.WaitGroup

	// Fetch and analyze historical data for each symbol and timeframe
	for _, symbol := range symbols {
		for _, timeframe := range timeframes {
			wg.Add(1)
			go func(sym models.Symbol, tf models.Timeframe) {
				defer wg.Done()
				analyzeSymbolTimeframe(ctx, binanceClient, calc, cfg, sym, tf, logger)
			}(symbol, timeframe)
		}
	}

	// Start real-time data streaming
	wg.Add(1)
	go func() {
		defer wg.Done()
		streamRealTimeData(ctx, binanceClient, symbols, timeframes, logger)
	}()

	// Start periodic analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		runPeriodicAnalysis(ctx, binanceClient, calc, cfg, symbols, timeframes, logger)
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	// Cleanup
	if err := binanceClient.CloseWebSocket(); err != nil {
		logger.WithError(err).Error("Failed to close WebSocket connection")
	}

	logger.Info("Application shutdown complete")
}

func analyzeSymbolTimeframe(ctx context.Context, client *binance.Client, calc *indicators.Calculator, cfg *config.Config, symbol models.Symbol, timeframe models.Timeframe, logger *logrus.Logger) {
	logger.WithFields(logrus.Fields{
		"symbol":    symbol,
		"timeframe": timeframe,
	}).Info("Starting analysis")

	// Fetch historical data
	klines, err := client.GetKlines(ctx, symbol, timeframe, 100)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"symbol":    symbol,
			"timeframe": timeframe,
		}).Error("Failed to fetch klines")
		return
	}

	if len(klines) == 0 {
		logger.WithFields(logrus.Fields{
			"symbol":    symbol,
			"timeframe": timeframe,
		}).Warn("No klines data received")
		return
	}

	// Calculate technical indicators
	indicators, err := calc.CalculateAllIndicators(
		klines,
		cfg.Indicators.RSI.Periods,
		cfg.Indicators.MA.Periods,
		cfg.Indicators.MA.Type,
		cfg.Indicators.KDJ.KPeriod,
		cfg.Indicators.KDJ.DPeriod,
		cfg.Indicators.KDJ.JPeriod,
	)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"symbol":    symbol,
			"timeframe": timeframe,
		}).Error("Failed to calculate indicators")
		return
	}

	// Get current price
	ticker, err := client.GetTicker24hr(ctx, symbol)
	if err != nil {
		logger.WithError(err).WithField("symbol", symbol).Error("Failed to fetch ticker")
		return
	}

	// Calculate volatility
	volatility, err := calc.CalculateVolatility(klines, 20)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"symbol":    symbol,
			"timeframe": timeframe,
		}).Warn("Failed to calculate volatility")
	}

	// Log comprehensive analysis
	logFields := logrus.Fields{
		"symbol":           symbol,
		"timeframe":        timeframe,
		"current_price":    utils.FormatDecimal(ticker.Price, 8),
		"price_change_24h": utils.FormatPercentage(ticker.PriceChangePercent),
		"volume_24h":       utils.FormatDecimal(ticker.Volume, 2),
		"quote_volume_24h": utils.FormatDecimal(ticker.QuoteVolume, 2),
		"kdj_k":            utils.FormatDecimal(indicators.KDJ.K, 2),
		"kdj_d":            utils.FormatDecimal(indicators.KDJ.D, 2),
		"kdj_j":            utils.FormatDecimal(indicators.KDJ.J, 2),
		"volatility":       utils.FormatDecimal(volatility, 8),
		"last_kline_time":  utils.FormatTimestamp(klines[len(klines)-1].CloseTime),
		"klines_count":     len(klines),
	}

	// Add RSI values for each period
	for period, rsi := range indicators.RSI {
		logFields[fmt.Sprintf("rsi_%d", period)] = utils.FormatDecimal(rsi, 2)
	}

	// Add MA values for each period
	for period, ma := range indicators.MA {
		logFields[fmt.Sprintf("ma_%d", period)] = utils.FormatDecimal(ma, 8)
	}

	logger.WithFields(logFields).Info("Technical analysis completed")

	// Generate trading signals
	generateTradingSignals(symbol, timeframe, ticker, indicators, logger)
}

func generateTradingSignals(symbol models.Symbol, timeframe models.Timeframe, ticker *models.TickerPrice, indicators *models.TechnicalIndicators, logger *logrus.Logger) {
	var signals []string

	// RSI signals for each period
	for period, rsi := range indicators.RSI {
		if rsi.LessThan(decimal.NewFromInt(30)) {
			signals = append(signals, fmt.Sprintf("RSI_%d_OVERSOLD", period))
		} else if rsi.GreaterThan(decimal.NewFromInt(70)) {
			signals = append(signals, fmt.Sprintf("RSI_%d_OVERBOUGHT", period))
		}
	}

	// Moving Average signals for each period
	for period, ma := range indicators.MA {
		if ticker.Price.GreaterThan(ma) {
			signals = append(signals, fmt.Sprintf("PRICE_ABOVE_MA_%d", period))
		} else {
			signals = append(signals, fmt.Sprintf("PRICE_BELOW_MA_%d", period))
		}
	}

	// KDJ signals
	if indicators.KDJ.K.GreaterThan(indicators.KDJ.D) && indicators.KDJ.K.LessThan(decimal.NewFromInt(20)) {
		signals = append(signals, "KDJ_BULLISH_CROSSOVER")
	} else if indicators.KDJ.K.LessThan(indicators.KDJ.D) && indicators.KDJ.K.GreaterThan(decimal.NewFromInt(80)) {
		signals = append(signals, "KDJ_BEARISH_CROSSOVER")
	}

	if len(signals) > 0 {
		logger.WithFields(logrus.Fields{
			"symbol":    symbol,
			"timeframe": timeframe,
			"signals":   signals,
		}).Info("Trading signals generated")
	}
}

func streamRealTimeData(ctx context.Context, client *binance.Client, symbols []models.Symbol, timeframes []models.Timeframe, logger *logrus.Logger) {
	logger.Info("Starting real-time data streaming")

	msgChan, err := client.ConnectWebSocket(ctx, symbols, timeframes)
	if err != nil {
		logger.WithError(err).Error("Failed to connect to WebSocket")
		return
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("Real-time streaming stopped")
			return
		case msg, ok := <-msgChan:
			if !ok {
				logger.Warn("WebSocket channel closed")
				return
			}

			logger.WithFields(logrus.Fields{
				"stream": msg.Stream,
			}).Debug("Received WebSocket message")

			// Process different message types
			if ticker, ok := msg.Data.(map[string]interface{}); ok {
				if eventType, exists := ticker["e"]; exists && eventType == "24hrTicker" {
					logger.WithFields(logrus.Fields{
						"symbol": ticker["s"],
						"price":  ticker["c"],
						"change": ticker["P"],
					}).Info("Price update received")
				}
			}
		}
	}
}

func runPeriodicAnalysis(ctx context.Context, client *binance.Client, calc *indicators.Calculator, cfg *config.Config, symbols []models.Symbol, timeframes []models.Timeframe, logger *logrus.Logger) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	logger.Info("Starting periodic analysis")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Periodic analysis stopped")
			return
		case <-ticker.C:
			logger.Info("Running periodic analysis")

			for _, symbol := range symbols {
				for _, timeframe := range timeframes {
					go analyzeSymbolTimeframe(ctx, client, calc, cfg, symbol, timeframe, logger)
				}
			}
		}
	}
}

// Helper function for safe decimal parsing
func safeParseDec(s string) string {
	return s
}
