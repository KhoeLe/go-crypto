#!/bin/bash

# Real-time streaming script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
SYMBOL=${1:-"BTCUSDT"}
INTERVAL=${2:-"15m"}

echo -e "${BLUE}Go Crypto Real-Time Stream${NC}"
echo -e "${BLUE}=========================${NC}"
echo ""

# Build the streamer if needed
if [ ! -f "./streamer" ]; then
    echo -e "${YELLOW}Building streamer...${NC}"
    go build -o streamer cmd/stream/main.go
fi

echo -e "${GREEN}Starting real-time stream for ${SYMBOL} on ${INTERVAL} timeframe...${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop${NC}"
echo ""

# Handle cleanup on exit
cleanup() {
    echo ""
    echo -e "${YELLOW}Stopping stream...${NC}"
    if [ "$1" = "clean" ]; then
        rm -f ./streamer
        echo -e "${YELLOW}Cleaned up build artifacts${NC}"
    fi
    exit 0
}

# Set trap for cleanup
trap 'cleanup' INT TERM

# Run the streamer
./streamer -symbol="$SYMBOL" -interval="$INTERVAL"

cleanup
