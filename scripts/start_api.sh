#!/bin/bash

# API Server startup script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Default values
PORT=${1:-"8080"}
CONFIG=${2:-""}

echo -e "${BLUE}🚀 Go Crypto Trading Analysis API Server${NC}"
echo -e "${BLUE}=======================================${NC}"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Error: Go is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Go is installed${NC}"

# Install dependencies if needed
echo -e "${YELLOW}📦 Checking dependencies...${NC}"
go mod tidy
echo -e "${GREEN}✅ Dependencies ready${NC}"

# Build the API server
echo -e "${YELLOW}🔨 Building API server...${NC}"
go build -o api-server cmd/api/main.go
echo -e "${GREEN}✅ API server built${NC}"
echo ""

# Show configuration
if [ -n "$CONFIG" ]; then
    echo -e "${CYAN}📁 Using config file: ${CONFIG}${NC}"
else
    echo -e "${CYAN}📁 Using default configuration${NC}"
fi

echo -e "${CYAN}🌐 Server will start on port: ${PORT}${NC}"
echo ""

# Show API endpoints
echo -e "${PURPLE}📋 Available API Endpoints:${NC}"
echo -e "${GREEN}Health Check:${NC}     http://localhost:${PORT}/api/v1/health"
echo -e "${GREEN}Price Data:${NC}       http://localhost:${PORT}/api/v1/price/BTCUSDT"
echo -e "${GREEN}Indicators:${NC}       http://localhost:${PORT}/api/v1/indicators/ETHUSDT?interval=15m"
echo -e "${GREEN}Analysis:${NC}         http://localhost:${PORT}/api/v1/analysis/BTCUSDT?interval=4h"
echo -e "${GREEN}Multi-Analysis:${NC}   http://localhost:${PORT}/api/v1/multi-analysis/BTCUSDT"
echo -e "${GREEN}Signals:${NC}          http://localhost:${PORT}/api/v1/signals/ETHUSDT"
echo -e "${GREEN}Web Interface:${NC}    http://localhost:${PORT}/"
echo ""

# Function to cleanup on exit
cleanup() {
    echo ""
    echo -e "${YELLOW}🧹 Cleaning up...${NC}"
    rm -f ./api-server
    echo -e "${GREEN}✅ Cleanup complete${NC}"
    exit 0
}

# Set trap for cleanup
trap 'cleanup' INT TERM

# Start the server
echo -e "${GREEN}🚀 Starting API server...${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop${NC}"
echo ""

if [ -n "$CONFIG" ]; then
    ./api-server --port="$PORT" --config="$CONFIG"
else
    ./api-server --port="$PORT"
fi

cleanup
