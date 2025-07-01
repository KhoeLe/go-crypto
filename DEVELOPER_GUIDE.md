# Go Crypto Trading Analysis - Developer Guide

## Overview

This is a professional-grade Go application for cryptocurrency market analysis using the Binance API. It provides real-time and historical data analysis with technical indicators including RSI, Moving Averages, and KDJ stochastic oscillator.

## Features

### ✅ Implemented Features

- **Multi-timeframe Analysis**: 15m, 4h, 1d (configurable)
- **Technical Indicators**:
  - RSI (Relative Strength Index)
  - SMA/EMA (Simple/Exponential Moving Average)
  - KDJ (Stochastic Oscillator)
  - Volatility calculation
- **Real-time Data**: WebSocket streaming from Binance
- **Multiple Trading Pairs**: BTCUSDT, ETHUSDT, BNBUSDT, and more
- **Professional Architecture**: Clean code structure with proper error handling
- **Configurable**: YAML-based configuration system
- **CLI Tools**: Command-line tools for analysis and streaming
- **Comprehensive Logging**: Structured logging with logrus
- **Docker Support**: Containerized deployment
- **Testing**: Unit tests for core functionality

## Project Structure

```
go-crypto/
├── cmd/                        # Application entry points
│   ├── main.go                 # Main application
│   ├── analyzer/main.go        # CLI analysis tool
│   └── stream/main.go          # Real-time streaming tool
├── internal/                   # Private application code
│   ├── binance/               # Binance API client
│   │   └── client.go
│   ├── config/                # Configuration management
│   │   ├── config.go
│   │   └── loader.go
│   ├── indicators/            # Technical indicators
│   │   ├── calculator.go
│   │   └── calculator_test.go
│   └── models/                # Data models
│       └── types.go
├── pkg/                       # Public library code
│   └── utils/                 # Utility functions
│       └── helpers.go
├── configs/                   # Configuration files
│   └── config.yaml
├── scripts/                   # Helper scripts
│   ├── quick_analysis.sh
│   └── stream.sh
├── Dockerfile                 # Docker configuration
├── Makefile                   # Build automation
├── go.mod                     # Go module definition
└── README.md
```

## Quick Start

### Prerequisites

- Go 1.21 or later
- Internet connection for Binance API access

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd go-crypto

# Install dependencies
go mod tidy

# Build the application
make build
```

### Basic Usage

#### 1. Quick Analysis
```bash
# Analyze BTCUSDT on 15m timeframe
./scripts/quick_analysis.sh

# Analyze specific symbol and timeframe
./scripts/quick_analysis.sh ETHUSDT 4h

# Custom analysis with more data points
./scripts/quick_analysis.sh BNBUSDT 1d 100
```

#### 2. Real-time Streaming
```bash
# Stream BTCUSDT on 15m timeframe
./scripts/stream.sh

# Stream specific symbol and timeframe
./scripts/stream.sh ETHUSDT 4h
```

#### 3. Enhanced Multi-Timeframe Analysis
```bash
# Default analysis (15m, 4h, 1d) with intelligent summary
curl http://localhost:8080/api/v1/multi-analysis/BTCUSDT

# Custom timeframes
curl "http://localhost:8080/api/v1/multi-analysis/BTCUSDT?timeframes=5m,15m,1h,4h"

# Get just the summary
curl http://localhost:8080/api/v1/multi-analysis/BTCUSDT | jq '.data.summary'
```

#### 4. API Server
```bash
# Start the REST API server
make start-api

# Test enhanced multi-timeframe demo
./scripts/enhanced_api_demo.sh
```

## Configuration

The application uses `configs/config.yaml` for configuration. You can customize:

### Trading Symbols
```yaml
symbols:
  - "BTCUSDT"
  - "ETHUSDT"
  - "BNBUSDT"
  - "ADAUSDT"
  - "SOLUSDT"
```

### Timeframes
```yaml
intervals:
  - "15m"
  - "4h"
  - "1d"
```

### Technical Indicators
```yaml
indicators:
  rsi:
    period: 14
  ma:
    period: 20
    type: "SMA"  # SMA, EMA, WMA
  kdj:
    k_period: 9
    d_period: 3
    j_period: 3
```

## Technical Indicators Explained

### RSI (Relative Strength Index)
- **Range**: 0-100
- **Overbought**: > 70
- **Oversold**: < 30
- **Usage**: Momentum oscillator for identifying potential reversal points

### Moving Average (MA)
- **Types**: SMA (Simple), EMA (Exponential), WMA (Weighted)
- **Usage**: Trend identification
- **Signal**: Price above MA = bullish, below MA = bearish

### KDJ Stochastic Oscillator
- **Components**: %K, %D, %J
- **Range**: 0-100
- **Signals**: 
  - K > D crossover in oversold area = buy signal
  - K < D crossover in overbought area = sell signal

## API Usage Examples

### Using the Binance Client
```go
// Initialize client
client := binance.NewClient(&config.BinanceConfig{
    BaseURL: "https://api.binance.com",
    Timeout: 30,
}, logger)

// Fetch klines
klines, err := client.GetKlines(ctx, models.BTCUSDT, models.Timeframe15m, 100)
if err != nil {
    log.Fatal(err)
}

// Get ticker data
ticker, err := client.GetTicker24hr(ctx, models.BTCUSDT)
if err != nil {
    log.Fatal(err)
}
```

### Calculating Indicators
```go
// Initialize calculator
calc := indicators.NewCalculator()

// Calculate RSI
rsi, err := calc.CalculateRSI(klines, 14)

// Calculate SMA
sma, err := calc.CalculateSMA(klines, 20)

// Calculate all indicators
indicators, err := calc.CalculateAllIndicators(klines, 14, 20, "SMA", 9, 3, 3)
```

## Development

### Running Tests
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Open coverage report
open coverage.html
```

### Code Quality
```bash
# Format code
make format

# Run linter
make lint

# Run go vet
make vet
```

### Building
```bash
# Build for current platform
make build

# Build Docker image
make docker-build

# Run in Docker
make docker-run
```

## Deployment

### Docker Deployment
```bash
# Build image
docker build -t go-crypto:latest .

# Run container
docker run --rm -it go-crypto:latest
```

### Production Considerations

1. **API Rate Limits**: Binance has rate limits (1200 requests/minute)
2. **WebSocket Reconnection**: Implement reconnection logic for production
3. **Error Handling**: Monitor and handle API errors gracefully
4. **Logging**: Use structured logging for production monitoring
5. **Configuration**: Use environment variables for sensitive data

## Trading Signals

The application generates several types of trading signals:

### RSI Signals
- `RSI_OVERSOLD`: RSI < 30 (potential buy)
- `RSI_OVERBOUGHT`: RSI > 70 (potential sell)

### Moving Average Signals
- `PRICE_ABOVE_MA`: Price above moving average (bullish)
- `PRICE_BELOW_MA`: Price below moving average (bearish)

### KDJ Signals
- `KDJ_BULLISH_CROSSOVER`: K crosses above D in oversold area
- `KDJ_BEARISH_CROSSOVER`: K crosses below D in overbought area

## Supported Timeframes

- `1m`, `3m`, `5m`: Short-term scalping
- `15m`, `30m`: Short-term trading
- `1h`, `2h`, `4h`: Medium-term trading
- `6h`, `8h`, `12h`: Swing trading
- `1d`, `3d`, `1w`: Long-term analysis

## Error Handling

The application implements comprehensive error handling:

- **Network Errors**: Automatic retry with exponential backoff
- **API Errors**: Proper error parsing and logging
- **Data Validation**: Input validation for all parameters
- **Graceful Shutdown**: Clean resource cleanup on exit

## Performance Considerations

- **Concurrent Processing**: Goroutines for parallel symbol analysis
- **Memory Management**: Efficient data structures
- **Rate Limiting**: Respects Binance API limits
- **Caching**: Intelligent data caching for better performance

## Troubleshooting

### Common Issues

1. **Build Errors**: Run `go mod tidy` to resolve dependencies
2. **API Errors**: Check internet connection and Binance API status
3. **WebSocket Issues**: Verify network connectivity and firewall settings
4. **Configuration Errors**: Validate YAML syntax in config files

### Debug Mode
```bash
# Enable debug logging
export LOG_LEVEL=debug
go run cmd/main.go
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is for educational and research purposes. Please respect Binance's terms of service when using their API.

## Disclaimer

This software is for educational purposes only. Do not use it for actual trading without proper risk management and understanding of financial markets. The authors are not responsible for any financial losses.
