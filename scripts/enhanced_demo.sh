#!/bin/bash

# Enhanced API Demo Script for Go Crypto Trading Analysis
# This script demonstrates the new enhanced features

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Default values
API_URL=${1:-"http://localhost:8080/api/v1"}
SYMBOL=${2:-"BTCUSDT"}

echo -e "${BLUE}🚀 Enhanced Go Crypto Trading Analysis Demo${NC}"
echo -e "${BLUE}===========================================${NC}"
echo -e "${YELLOW}API URL: ${API_URL}${NC}"
echo -e "${YELLOW}Symbol: ${SYMBOL}${NC}"
echo ""

# Health check
echo -e "${CYAN}1. Health Check${NC}"
HEALTH_RESPONSE=$(curl -s "$API_URL/health")
if [[ $HEALTH_RESPONSE == *"\"status\":\"ok\""* ]]; then
    echo -e "${GREEN}✅ API is healthy${NC}"
else
    echo -e "${RED}❌ API health check failed${NC}"
    exit 1
fi

# Enhanced Analysis (15m timeframe)
echo -e "\n${CYAN}2. Enhanced Analysis (15m timeframe)${NC}"
echo -e "${YELLOW}Features: Money Flow, Volume Breakout, Historical Indicators, Klines Data${NC}"
ENHANCED_RESPONSE=$(curl -s "$API_URL/enhanced-analysis/$SYMBOL?interval=15m")

if [[ $ENHANCED_RESPONSE == *"\"success\":true"* ]]; then
    echo -e "${GREEN}✅ Enhanced analysis completed${NC}"
    
    # Extract money flow information
    MONEY_FLOW_INDEX=$(echo "$ENHANCED_RESPONSE" | grep -o '"money_flow_index":"[^"]*"' | cut -d'"' -f4 | head -1)
    MONEY_FLOW_CHANGE=$(echo "$ENHANCED_RESPONSE" | grep -o '"money_flow_change":"[^"]*"' | cut -d'"' -f4 | head -1)
    
    if [ ! -z "$MONEY_FLOW_INDEX" ]; then
        echo -e "   💰 Money Flow Index: ${MONEY_FLOW_INDEX%.*}"
        echo -e "   📈 Money Flow Change: ${MONEY_FLOW_CHANGE%.*}%"
    fi
    
    # Extract volume breakout information
    VOLUME_BREAKOUT=$(echo "$ENHANCED_RESPONSE" | grep -o '"is_breakout":[^,]*' | cut -d':' -f2)
    BREAKOUT_STRENGTH=$(echo "$ENHANCED_RESPONSE" | grep -o '"breakout_strength":"[^"]*"' | cut -d'"' -f4 | head -1)
    BREAKOUT_DIRECTION=$(echo "$ENHANCED_RESPONSE" | grep -o '"breakout_direction":"[^"]*"' | cut -d'"' -f4 | head -1)
    
    if [[ $VOLUME_BREAKOUT == "true" ]]; then
        echo -e "   🔥 Volume Breakout: ${RED}DETECTED${NC}"
        echo -e "   💪 Strength: ${BREAKOUT_STRENGTH%.*}/10"
        echo -e "   📊 Direction: ${BREAKOUT_DIRECTION}"
    else
        echo -e "   ⚪ Volume Breakout: ${GREEN}None${NC}"
    fi
    
    # Extract klines count
    KLINES_COUNT=$(echo "$ENHANCED_RESPONSE" | grep -o '"klines":\[[^]]*\]' | grep -o '"symbol"' | wc -l | tr -d ' ')
    if [ ! -z "$KLINES_COUNT" ] && [ "$KLINES_COUNT" -gt 0 ]; then
        echo -e "   📊 Klines included: ${KLINES_COUNT} periods"
    fi
    
else
    echo -e "${RED}❌ Enhanced analysis failed${NC}"
fi

# Multi-timeframe Analysis with Enhanced Features
echo -e "\n${CYAN}3. Multi-timeframe Enhanced Analysis${NC}"
MULTI_ENHANCED_RESPONSE=$(curl -s "$API_URL/multi-analysis/$SYMBOL?enhanced=true&timeframes=15m")

if [[ $MULTI_ENHANCED_RESPONSE == *"\"enhanced\":true"* ]]; then
    echo -e "${GREEN}✅ Multi-timeframe enhanced analysis completed${NC}"
    echo -e "   🕐 Focused on 15m timeframe as requested"
    echo -e "   ⚡ Features: Money Flow, Volume Analysis, Historical tracking"
else
    echo -e "${YELLOW}⚠️  Multi-timeframe enhanced analysis not available, using standard${NC}"
fi

# Traditional Multi-timeframe Analysis for comparison
echo -e "\n${CYAN}4. Traditional Multi-timeframe Analysis${NC}"
MULTI_RESPONSE=$(curl -s "$API_URL/multi-analysis/$SYMBOL?timeframes=15m,4h,1d")

if [[ $MULTI_RESPONSE == *"\"15m\""* ]]; then
    echo -e "${GREEN}✅ Traditional multi-timeframe analysis completed${NC}"
    
    # Extract RSI values for 15m
    RSI_6_15M=$(echo "$MULTI_RESPONSE" | grep -A 20 '"15m"' | grep -o '"RSI_6":"[^"]*"' | cut -d'"' -f4 | head -1)
    RSI_12_15M=$(echo "$MULTI_RESPONSE" | grep -A 20 '"15m"' | grep -o '"RSI_12":"[^"]*"' | cut -d'"' -f4 | head -1)
    
    if [ ! -z "$RSI_6_15M" ]; then
        echo -e "   📊 15m RSI-6:  ${RSI_6_15M%.*}"
        echo -e "   📊 15m RSI-12: ${RSI_12_15M%.*}"
    fi
else
    echo -e "${RED}❌ Multi-timeframe analysis failed${NC}"
fi

# Enhanced signals
echo -e "\n${CYAN}5. Enhanced Trading Signals${NC}"
SIGNALS=$(echo "$ENHANCED_RESPONSE" | grep -o '"signals":\[[^]]*\]')

if [[ $SIGNALS == *"MONEY_FLOW"* ]] || [[ $SIGNALS == *"VOLUME_BREAKOUT"* ]]; then
    echo -e "${RED}🚨 Enhanced signals detected!${NC}"
    
    # Money flow signals
    if [[ $SIGNALS == *"MONEY_FLOW_INCREASING"* ]]; then
        echo -e "   💹 Money Flow: INCREASING"
    elif [[ $SIGNALS == *"MONEY_FLOW_DECREASING"* ]]; then
        echo -e "   💹 Money Flow: DECREASING"
    fi
    
    # Volume signals
    if [[ $SIGNALS == *"VOLUME_BREAKOUT_BULLISH"* ]]; then
        echo -e "   🔥 Volume Breakout: BULLISH"
    elif [[ $SIGNALS == *"VOLUME_BREAKOUT_BEARISH"* ]]; then
        echo -e "   🔥 Volume Breakout: BEARISH"
    fi
    
else
    echo -e "${GREEN}✅ No enhanced signals detected${NC}"
fi

echo -e "\n${GREEN}🎉 Enhanced demo completed successfully!${NC}"
echo -e "\n${BLUE}💡 New API Endpoints:${NC}"
echo -e "${CYAN}   GET /api/v1/enhanced-analysis/{symbol}?interval=15m${NC}     - Enhanced analysis with new features"
echo -e "${CYAN}   GET /api/v1/multi-analysis/{symbol}?enhanced=true${NC}       - Multi-timeframe with enhancements"
echo -e "${CYAN}   GET /api/v1/multi-analysis/{symbol}?timeframes=15m${NC}      - Focus on 15m timeframe"

echo -e "\n${BLUE}🆕 New Features Included:${NC}"
echo -e "${YELLOW}   • Money Flow Index with % change calculation${NC}"
echo -e "${YELLOW}   • Automatic volume breakout detection${NC}"
echo -e "${YELLOW}   • Historical RSI & MA tracking${NC}"
echo -e "${YELLOW}   • Klines data included (limit=50)${NC}"
echo -e "${YELLOW}   • Enhanced signal generation${NC}"
echo -e "${YELLOW}   • Focus on 15m timeframe as requested${NC}"
