# API Usage Guide

## 🚀 Quick Start

The Go Crypto Trading Analysis API provides comprehensive cryptocurrency trading analysis with technical indicators. This guide shows you how to use the API effectively.

## Starting the API Server

### Option 1: Using Makefile (Recommended)
```bash
# Build and start the API server
make start-api

# Or build first, then run
make build-api
./build/go-crypto-api -port=8080
```

### Option 2: Using Go directly
```bash
# Run directly
go run cmd/api/main.go -port=8080

# Or using the startup script
./scripts/start_api.sh 8080
```

The API will start on `http://localhost:8080` by default.

## API Endpoints

All endpoints are prefixed with `/api/v1/`

### Market Data
- `GET /price/{symbol}` - Get current price
- `GET /ticker/{symbol}` - Get 24h ticker data  
- `GET /klines/{symbol}` - Get OHLCV klines data

### Technical Analysis
- `GET /indicators/{symbol}` - Get technical indicators (RSI, MA, KDJ)
- `GET /analysis/{symbol}` - Get complete market analysis
- `GET /signals/{symbol}` - Get trading signals
- `GET /multi-analysis/{symbol}` - Get multi-timeframe analysis

### System
- `GET /health` - Health check
- `GET /config` - Current configuration
- `GET /symbols` - Available symbols and intervals

## Example Usage

### 1. Health Check
```bash
curl http://localhost:8080/api/v1/health
```

### 2. Get Current Price
```bash
curl http://localhost:8080/api/v1/price/BTCUSDT
```

### 3. Get Complete Analysis
```bash
curl "http://localhost:8080/api/v1/analysis/BTCUSDT?interval=15m&limit=50"
```

### 4. Get Multi-Timeframe Analysis
```bash
# Default timeframes (15m, 4h, 1d)
curl http://localhost:8080/api/v1/multi-analysis/BTCUSDT

# Custom timeframes
curl "http://localhost:8080/api/v1/multi-analysis/BTCUSDT?timeframes=15m,1h,4h,1d"

# Just short-term analysis
curl "http://localhost:8080/api/v1/multi-analysis/ETHUSDT?timeframes=5m,15m,30m"
```

### 5. Get Trading Signals
```bash
curl "http://localhost:8080/api/v1/signals/ETHUSDT?interval=4h"
```

## Query Parameters

- `interval` - Time interval (15m, 4h, 1d) [default: 15m]
- `limit` - Number of klines to fetch [default: 50]
- `timeframes` - Comma-separated list for multi-analysis (15m,4h,1d) [default: 15m,4h,1d]

## Enhanced Multi-Timeframe Analysis

The `/multi-analysis` endpoint now provides:

### **📊 Individual Timeframe Data**
- Complete analysis for each requested timeframe
- RSI, MA, KDJ indicators for all periods
- Trading signals per timeframe

### **🧠 Intelligent Summary**
- `overall_trend`: bullish, bearish, or neutral
- `confidence_score`: 0-100 based on signal consensus
- `consensus_signals`: signals appearing across multiple timeframes
- `risk_level`: low, medium, high based on volatility
- `recommendation`: buy, sell, hold

### **⚡ Performance Features**
- Concurrent processing of all timeframes
- Customizable timeframe selection
- Optimized for speed and accuracy

## Response Format

All responses follow this structure:
```json
{
  "success": true,
  "data": { ... },
  "timestamp": "2025-07-01T10:00:00Z"
}
```

## Error Handling

Error responses:
```json
{
  "success": false,
  "error": "Error message",
  "timestamp": "2025-07-01T10:00:00Z"
}
```

## Programming Examples

### Go Client
See `examples/api_client.go` for a complete Go client example.

### Python Client
```python
import requests

# Get current price
response = requests.get('http://localhost:8080/api/v1/price/BTCUSDT')
data = response.json()
print(f"BTC Price: ${data['data']['price']}")

# Get analysis
response = requests.get('http://localhost:8080/api/v1/analysis/BTCUSDT?interval=15m')
analysis = response.json()['data']
print(f"RSI: {analysis['rsi']}")
print(f"Signals: {analysis['signals']}")
```

### JavaScript/Node.js Client
```javascript
const axios = require('axios');

async function getAnalysis(symbol) {
  try {
    const response = await axios.get(`http://localhost:8080/api/v1/analysis/${symbol}`);
    return response.data.data;
  } catch (error) {
    console.error('Error:', error.message);
  }
}

// Usage
getAnalysis('BTCUSDT').then(data => {
  console.log('Analysis:', data);
});
```

## Web Interface

Open `http://localhost:8080` in your browser to access the interactive API documentation and test interface.

## Configuration

The API uses the configuration file at `configs/config.yaml`. You can customize:
- RSI periods (default: 6, 12, 24)
- MA periods (default: 7, 25, 99)
- KDJ parameters
- Supported symbols and timeframes

## Rate Limiting

The API respects Binance rate limits. For production use, consider implementing additional rate limiting and caching.

## Docker Usage

```bash
# Build Docker image
docker build -t go-crypto-api .

# Run container
docker run -p 8080:8080 go-crypto-api
```

## Support

For issues and questions, refer to the main README.md or check the API documentation at `http://localhost:8080`.
