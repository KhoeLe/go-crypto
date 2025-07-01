#!/bin/bash

# API Demo Script - Shows how easy it is to use the Go Crypto API

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

API_URL="http://localhost:8080/api/v1"

echo -e "${BLUE}🚀 Go Crypto API Demo${NC}"
echo -e "${BLUE}===================${NC}"

# Check if API is running
echo -e "\n${YELLOW}1. Checking API health...${NC}"
if curl -s "$API_URL/health" > /dev/null; then
    echo -e "${GREEN}✅ API is running!${NC}"
else
    echo -e "${RED}❌ API is not running. Start it with: make start-api${NC}"
    exit 1
fi

# Get current price
echo -e "\n${YELLOW}2. Getting current BTC price...${NC}"
PRICE_RESPONSE=$(curl -s "$API_URL/price/BTCUSDT")
PRICE=$(echo $PRICE_RESPONSE | grep -o '"price":"[^"]*"' | cut -d'"' -f4)
echo -e "${GREEN}💰 BTC Price: \$${PRICE}${NC}"

# Get available symbols
echo -e "\n${YELLOW}3. Getting available symbols...${NC}"
SYMBOLS_RESPONSE=$(curl -s "$API_URL/symbols")
echo -e "${CYAN}📋 Available symbols and intervals:${NC}"
echo $SYMBOLS_RESPONSE | grep -o '"symbols":\[[^]]*\]' | sed 's/"symbols"://g'

# Get technical analysis
echo -e "\n${YELLOW}4. Getting technical analysis for BTCUSDT...${NC}"
ANALYSIS_RESPONSE=$(curl -s "$API_URL/analysis/BTCUSDT?interval=15m")
echo -e "${PURPLE}📊 Technical Analysis:${NC}"

# Extract RSI values
RSI_6=$(echo $ANALYSIS_RESPONSE | grep -o '"RSI_6":"[^"]*"' | cut -d'"' -f4)
RSI_12=$(echo $ANALYSIS_RESPONSE | grep -o '"RSI_12":"[^"]*"' | cut -d'"' -f4)
RSI_24=$(echo $ANALYSIS_RESPONSE | grep -o '"RSI_24":"[^"]*"' | cut -d'"' -f4)

if [ ! -z "$RSI_6" ]; then
    echo -e "   RSI-6:  ${RSI_6%.*}"
    echo -e "   RSI-12: ${RSI_12%.*}"
    echo -e "   RSI-24: ${RSI_24%.*}"
fi

# Get trading signals
echo -e "\n${YELLOW}5. Getting trading signals...${NC}"
SIGNALS_RESPONSE=$(curl -s "$API_URL/signals/BTCUSDT?interval=15m")
SIGNALS=$(echo $SIGNALS_RESPONSE | grep -o '"signals":\[[^]]*\]')
if [[ $SIGNALS == *"PRICE_BELOW_MA"* ]] || [[ $SIGNALS == *"PRICE_ABOVE_MA"* ]] || [[ $SIGNALS == *"RSI"* ]]; then
    echo -e "${RED}🚨 Active signals detected!${NC}"
    echo $SIGNALS
else
    echo -e "${GREEN}✅ No active signals${NC}"
fi

# Multi-timeframe analysis
echo -e "\n${YELLOW}6. Getting multi-timeframe analysis...${NC}"
MULTI_RESPONSE=$(curl -s "$API_URL/multi-analysis/BTCUSDT")
echo -e "${CYAN}🕐 Multi-timeframe analysis completed${NC}"

echo -e "\n${GREEN}🎉 Demo completed successfully!${NC}"
echo -e "\n${BLUE}💡 Try these API endpoints:${NC}"
echo -e "${CYAN}   GET /api/v1/health${NC}                    - Health check"
echo -e "${CYAN}   GET /api/v1/price/BTCUSDT${NC}             - Current price"
echo -e "${CYAN}   GET /api/v1/ticker/ETHUSDT${NC}            - 24h ticker data"
echo -e "${CYAN}   GET /api/v1/analysis/BTCUSDT?interval=15m${NC} - Complete analysis"
echo -e "${CYAN}   GET /api/v1/signals/BTCUSDT${NC}           - Trading signals"
echo -e "${CYAN}   GET /api/v1/multi-analysis/BTCUSDT${NC}    - Multi-timeframe"
echo -e "${CYAN}   GET /api/v1/symbols${NC}                   - Available symbols"
echo -e "${CYAN}   GET /api/v1/config${NC}                    - API configuration"

echo -e "\n${PURPLE}📖 Open http://localhost:8080 for interactive documentation${NC}"
