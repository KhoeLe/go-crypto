package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"go-crypto/internal/binance"
	"go-crypto/internal/config"
	"go-crypto/internal/models"

	"github.com/sirupsen/logrus"
)

func main() {
	// Parse command line flags
	symbolFlag := flag.String("symbol", "BTCUSDT", "Trading symbol")
	intervalFlag := flag.String("interval", "15m", "Time interval")
	flag.Parse()

	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce log noise

	// Load configuration
	cfg := config.DefaultConfig()

	// Initialize client
	binanceClient := binance.NewClient(&cfg.Binance, logger)

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	symbol := models.Symbol(*symbolFlag)
	interval := models.Timeframe(*intervalFlag)

	fmt.Printf("Starting real-time stream for %s on %s...\n", symbol, interval)
	fmt.Println("Press Ctrl+C to stop")

	// Connect to WebSocket
	msgChan, err := binanceClient.ConnectWebSocket(ctx, []models.Symbol{symbol}, []models.Timeframe{interval})
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}

	// Handle messages
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				fmt.Println("WebSocket connection closed")
				return
			}

			// Process different message types
			if data, ok := msg.Data.(map[string]interface{}); ok {
				if eventType, exists := data["e"]; exists {
					switch eventType {
					case "24hrTicker":
						handleTickerUpdate(data)
					case "kline":
						handleKlineUpdate(data)
					}
				}
			}
		case <-ctx.Done():
			fmt.Println("Stream stopped")
			return
		}
	}
}

func handleTickerUpdate(data map[string]interface{}) {
	symbol := data["s"]
	price := data["c"]
	change := data["P"]
	volume := data["v"]

	fmt.Printf("[TICKER] %s | Price: %s | Change: %s%% | Volume: %s\n",
		symbol, price, change, volume)
}

func handleKlineUpdate(data map[string]interface{}) {
	if klineData, ok := data["k"].(map[string]interface{}); ok {
		symbol := klineData["s"]
		interval := klineData["i"]
		open := klineData["o"]
		high := klineData["h"]
		low := klineData["l"]
		close := klineData["c"]
		volume := klineData["v"]
		isClosed := klineData["x"]

		status := "UPDATING"
		if closed, ok := isClosed.(bool); ok && closed {
			status = "CLOSED"
		}

		fmt.Printf("[KLINE %s] %s %s | O:%s H:%s L:%s C:%s V:%s | %s\n",
			status, symbol, interval, open, high, low, close, volume, time.Now().Format("15:04:05"))
	}
}
