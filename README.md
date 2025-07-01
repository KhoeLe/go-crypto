# Go Crypto Trading Analysis

A professional Go application for cryptocurrency market analysis using Binance API with **REST API** for easy integration.

## ✨ Features

- **REST API Server** - Easy-to-use HTTP API for integration with any application
- **Real-time price and volume data** from Binance API
- **Technical indicators**: RSI (6,12,24), MA (7,25,99), KDJ stochastic oscillator
- **Multi-timeframe analysis** (15m, 4h, 1d) with more timeframes available
- **WebSocket streaming** for real-time updates
- **Professional project structure** with clean architecture
- **Comprehensive error handling** and logging
- **Multiple tools**: CLI analyzer, real-time streamer, multi-timeframe analysis
- **Interactive API documentation** and web interface

## 🚀 Quick Start

### 🌐 API Server (Recommended)

Start the REST API server for easy access from any application:

```bash
# Start the API server
make start-api

# Or run directly
go run cmd/api/main.go -port=8080
```

**API is now running at `http://localhost:8080`**

#### Quick API Examples:
```bash
# Health check
curl http://localhost:8080/api/v1/health

# Get current BTC price
curl http://localhost:8080/api/v1/price/BTCUSDT

# Get complete analysis
curl "http://localhost:8080/api/v1/analysis/BTCUSDT?interval=15m"

# Get trading signals
curl http://localhost:8080/api/v1/signals/BTCUSDT
```

### 📖 API Documentation
- **Interactive docs**: Open `http://localhost:8080` in your browser
- **Complete guide**: See [API_GUIDE.md](API_GUIDE.md)
- **Go client example**: See [examples/api_client.go](examples/api_client.go)

### 💻 CLI Tools

```bash
# Quick analysis for BTCUSDT on 15m timeframe
./scripts/quick_analysis.sh

# Analyze ETHUSDT on 4h timeframe with 100 data points
./scripts/quick_analysis.sh ETHUSDT 4h 100

# Real-time streaming
./scripts/stream.sh BTCUSDT 15m

# Multi-timeframe analysis
go run cmd/main.go

# Complete demo
./scripts/demo.sh
```

## 📋 Prerequisites

- Go 1.21 or later
- Internet connection for Binance API

## Project Structure

```
go-crypto/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── binance/               # Binance API client
│   ├── indicators/            # Technical indicators
│   ├── models/                # Data models
│   └── config/                # Configuration
├── pkg/
│   └── utils/                 # Utility functions
├── configs/
│   └── config.yaml            # Configuration file
└── README.md
```

## Getting Started

1. Install dependencies:
```bash
go mod tidy
```

2. Run the application:
```bash
go run cmd/main.go
```

3. Analyze :
```bash
go build cmd/analyzer/main.go
```
## Configuration

Edit `configs/config.yaml` to customize:
- Trading symbols
- Timeframes
- Indicator parameters
- API settings

## Technical Indicators

- **RSI (Relative Strength Index)**: Momentum oscillator
- **MA (Moving Average)**: Trend indicator
- **KDJ**: Stochastic oscillator for momentum analysis
