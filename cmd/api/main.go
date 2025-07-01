package main

import (
	"flag"

	"go-crypto/internal/api"
	"go-crypto/internal/config"

	"github.com/sirupsen/logrus"
)

func main() {
	// Parse command line flags
	portFlag := flag.String("port", "8080", "Server port")
	configFlag := flag.String("config", "", "Config file path")
	flag.Parse()

	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Starting Go Crypto API Server")

	// Load configuration
	var cfg *config.Config
	var err error

	if *configFlag != "" {
		cfg, err = config.LoadConfig(*configFlag)
	} else {
		cfg, err = config.LoadConfig("")
	}

	if err != nil {
		logger.WithError(err).Warn("Failed to load config, using defaults")
		cfg = config.DefaultConfig()
	}

	// Validate configuration
	if err := config.ValidateConfig(cfg); err != nil {
		logger.WithError(err).Fatal("Invalid configuration")
	}

	// Create and start server
	server := api.NewServer(cfg, logger)

	logger.WithField("port", *portFlag).Info("Server starting")
	logger.Info("API Documentation:")
	logger.Info("GET /api/v1/health - Health check")
	logger.Info("GET /api/v1/price/{symbol} - Get current price")
	logger.Info("GET /api/v1/ticker/{symbol} - Get 24hr ticker stats")
	logger.Info("GET /api/v1/klines/{symbol}?interval=15m&limit=100 - Get klines data")
	logger.Info("GET /api/v1/indicators/{symbol}?interval=15m - Get technical indicators")
	logger.Info("GET /api/v1/analysis/{symbol}?interval=15m - Get complete analysis")
	logger.Info("GET /api/v1/multi-analysis/{symbol} - Get multi-timeframe analysis")
	logger.Info("GET /api/v1/signals/{symbol} - Get trading signals")
	logger.Info("GET /api/v1/symbols - Get supported symbols and intervals")
	logger.Info("GET /api/v1/config - Get current configuration")

	if err := server.Start(*portFlag); err != nil {
		logger.WithError(err).Fatal("Failed to start server")
	}
}
