#!/bin/bash

# Quick analysis script for Go Crypto Trading Analysis

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
LIMIT=${3:-50}

echo -e "${BLUE}Go Crypto Trading Analysis${NC}"
echo -e "${BLUE}=========================${NC}"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

# Build the analyzer if needed
if [ ! -f "./analyzer" ]; then
    echo -e "${YELLOW}Building analyzer...${NC}"
    go build -o analyzer cmd/analyzer/main.go
fi

echo -e "${GREEN}Running analysis for ${SYMBOL} on ${INTERVAL} timeframe...${NC}"
echo ""

# Run the analyzer
./analyzer -symbol="$SYMBOL" -interval="$INTERVAL" -limit="$LIMIT"

echo ""
echo -e "${BLUE}Analysis complete!${NC}"

# Cleanup
if [ "$4" = "clean" ]; then
    rm -f ./analyzer
    echo -e "${YELLOW}Cleaned up build artifacts${NC}"
fi
