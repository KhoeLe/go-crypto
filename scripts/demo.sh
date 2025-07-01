#!/bin/bash

# Comprehensive demo script for Go Crypto Trading Analysis

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Go Crypto Trading Analysis - Complete Demo${NC}"
echo -e "${BLUE}=============================================${NC}"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Error: Go is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Go is installed${NC}"

# Install dependencies
echo -e "${YELLOW}📦 Installing dependencies...${NC}"
go mod tidy
echo -e "${GREEN}✅ Dependencies installed${NC}"
echo ""

# Build all tools
echo -e "${YELLOW}🔨 Building tools...${NC}"
go build -o analyzer cmd/analyzer/main.go
go build -o streamer cmd/stream/main.go
echo -e "${GREEN}✅ All tools built successfully${NC}"
echo ""

# Run single analysis
echo -e "${PURPLE}📊 Running Single Symbol Analysis${NC}"
echo -e "${PURPLE}=================================${NC}"
./analyzer -symbol=BTCUSDT -interval=15m -limit=20
echo ""

# Show multi-timeframe analysis via API
echo -e "${CYAN}📈 Multi-Timeframe Analysis (via API)${NC}"
echo -e "${CYAN}=====================================${NC}"
echo -e "${YELLOW}Multi-timeframe analysis is available via:${NC}"
echo -e "${GREEN}  API: curl http://localhost:8080/api/v1/multi-analysis/BTCUSDT${NC}"
echo -e "${GREEN}  CLI: go run cmd/main.go${NC}"
echo ""

# Show available commands
echo -e "${BLUE}🛠️  Available Commands${NC}"
echo -e "${BLUE}===================${NC}"
echo -e "${GREEN}1. Quick Analysis:${NC}"
echo -e "   ./scripts/quick_analysis.sh [SYMBOL] [INTERVAL] [LIMIT]"
echo -e "   Example: ./scripts/quick_analysis.sh ETHUSDT 4h 50"
echo ""
echo -e "${GREEN}2. Real-time Streaming:${NC}"
echo -e "   ./scripts/stream.sh [SYMBOL] [INTERVAL]"
echo -e "   Example: ./scripts/stream.sh BTCUSDT 15m"
echo ""
echo -e "${GREEN}3. Multi-timeframe Analysis:${NC}"
echo -e "   go run cmd/main.go              # Full CLI application"
echo -e "   make start-api                  # Start API, then:"
echo -e "   curl http://localhost:8080/api/v1/multi-analysis/BTCUSDT"
echo ""
echo -e "${GREEN}4. Full Application:${NC}"
echo -e "   go run cmd/main.go"
echo ""
echo -e "${GREEN}5. Make Commands:${NC}"
echo -e "   make build    # Build the application"
echo -e "   make run      # Run the application"
echo -e "   make test     # Run tests"
echo -e "   make clean    # Clean build artifacts"
echo ""

# Show supported symbols and timeframes
echo -e "${BLUE}📋 Supported Symbols${NC}"
echo -e "${BLUE}===================${NC}"
echo "BTCUSDT, ETHUSDT, BNBUSDT, ADAUSDT, SOLUSDT"
echo ""
echo -e "${BLUE}📋 Supported Timeframes${NC}"
echo -e "${BLUE}=======================${NC}"
echo "1m, 3m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 8h, 12h, 1d, 3d, 1w"
echo ""

# Cleanup option
echo -e "${YELLOW}🧹 Cleaning up build artifacts...${NC}"
rm -f analyzer streamer
echo -e "${GREEN}✅ Cleanup complete${NC}"

echo ""
echo -e "${GREEN}🎉 Demo completed successfully!${NC}"
echo -e "${BLUE}For more information, see DEVELOPER_GUIDE.md${NC}"
