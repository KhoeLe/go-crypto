#!/bin/bash

# Enhanced Multi-Timeframe Analysis Demo

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

echo -e "${BLUE}🚀 Enhanced Multi-Timeframe Analysis Demo${NC}"
echo -e "${BLUE}===========================================${NC}"

# Check if API is running
echo -e "\n${YELLOW}1. Checking API health...${NC}"
if curl -s "$API_URL/health" > /dev/null; then
    echo -e "${GREEN}✅ API is running!${NC}"
else
    echo -e "${RED}❌ API is not running. Start it with: make start-api${NC}"
    exit 1
fi

# Test default multi-timeframe analysis
echo -e "\n${YELLOW}2. Default Multi-Timeframe Analysis (15m, 4h, 1d)...${NC}"
echo -e "${CYAN}📊 Getting comprehensive analysis for BTCUSDT...${NC}"
RESPONSE=$(curl -s "$API_URL/multi-analysis/BTCUSDT")

# Extract and display summary
SUMMARY=$(echo $RESPONSE | jq -r '.data.summary')
TREND=$(echo $SUMMARY | jq -r '.overall_trend')
CONFIDENCE=$(echo $SUMMARY | jq -r '.confidence_score')
RECOMMENDATION=$(echo $SUMMARY | jq -r '.recommendation')
RISK=$(echo $SUMMARY | jq -r '.risk_level')

echo -e "${PURPLE}📈 Analysis Summary:${NC}"
echo -e "   Overall Trend: ${TREND}"
echo -e "   Confidence: ${CONFIDENCE}%"
echo -e "   Recommendation: ${RECOMMENDATION}"
echo -e "   Risk Level: ${RISK}"

# Show consensus signals
CONSENSUS=$(echo $SUMMARY | jq -r '.consensus_signals[]' 2>/dev/null || echo "No consensus signals")
if [ "$CONSENSUS" != "No consensus signals" ]; then
    echo -e "${RED}🚨 Consensus Signals:${NC}"
    echo $SUMMARY | jq -r '.consensus_signals[]' | while read signal; do
        echo -e "   • $signal"
    done
else
    echo -e "${GREEN}✅ No consensus signals${NC}"
fi

# Test custom timeframes
echo -e "\n${YELLOW}3. Custom Timeframes Analysis (5m, 15m, 1h, 4h)...${NC}"
echo -e "${CYAN}📊 Getting short to medium-term analysis...${NC}"
CUSTOM_RESPONSE=$(curl -s "$API_URL/multi-analysis/BTCUSDT?timeframes=5m,15m,1h,4h")
CUSTOM_SUMMARY=$(echo $CUSTOM_RESPONSE | jq -r '.data.summary')
CUSTOM_TREND=$(echo $CUSTOM_SUMMARY | jq -r '.overall_trend')
CUSTOM_CONFIDENCE=$(echo $CUSTOM_SUMMARY | jq -r '.confidence_score')

echo -e "${PURPLE}📈 Short-Medium Term Summary:${NC}"
echo -e "   Trend: ${CUSTOM_TREND}"
echo -e "   Confidence: ${CUSTOM_CONFIDENCE}%"

# Compare different symbols
echo -e "\n${YELLOW}4. Multi-Symbol Comparison...${NC}"
SYMBOLS=("BTCUSDT" "ETHUSDT" "BNBUSDT")

for symbol in "${SYMBOLS[@]}"; do
    echo -e "\n${CYAN}📊 ${symbol} Summary:${NC}"
    SYMBOL_RESPONSE=$(curl -s "$API_URL/multi-analysis/$symbol")
    SYMBOL_SUMMARY=$(echo $SYMBOL_RESPONSE | jq -r '.data.summary')
    
    SYMBOL_TREND=$(echo $SYMBOL_SUMMARY | jq -r '.overall_trend')
    SYMBOL_CONFIDENCE=$(echo $SYMBOL_SUMMARY | jq -r '.confidence_score')
    SYMBOL_RECOMMENDATION=$(echo $SYMBOL_SUMMARY | jq -r '.recommendation')
    
    # Color code the trend
    case $SYMBOL_TREND in
        "bullish") TREND_COLOR=$GREEN ;;
        "bearish") TREND_COLOR=$RED ;;
        *) TREND_COLOR=$YELLOW ;;
    esac
    
    echo -e "   Trend: ${TREND_COLOR}${SYMBOL_TREND}${NC} (${SYMBOL_CONFIDENCE}%)"
    echo -e "   Action: ${SYMBOL_RECOMMENDATION}"
done

# Performance test
echo -e "\n${YELLOW}5. Performance Test...${NC}"
echo -e "${CYAN}📊 Testing concurrent processing speed...${NC}"
START_TIME=$(date +%s.%N)
curl -s "$API_URL/multi-analysis/BTCUSDT?timeframes=1m,5m,15m,30m,1h,4h,1d" > /dev/null
END_TIME=$(date +%s.%N)
DURATION=$(echo "$END_TIME - $START_TIME" | bc)
echo -e "${GREEN}✅ Processed 7 timeframes in ${DURATION} seconds${NC}"

echo -e "\n${GREEN}🎉 Enhanced Multi-Timeframe Analysis Demo Completed!${NC}"
echo -e "\n${BLUE}💡 Available Features:${NC}"
echo -e "${CYAN}   • Default analysis: GET /multi-analysis/SYMBOL${NC}"
echo -e "${CYAN}   • Custom timeframes: GET /multi-analysis/SYMBOL?timeframes=15m,4h,1d${NC}"
echo -e "${CYAN}   • Intelligent summary with trend analysis${NC}"
echo -e "${CYAN}   • Consensus signals across timeframes${NC}"
echo -e "${CYAN}   • Risk assessment and recommendations${NC}"
echo -e "${CYAN}   • Concurrent processing for speed${NC}"

echo -e "\n${PURPLE}📖 Try these examples:${NC}"
echo -e "${CYAN}curl '$API_URL/multi-analysis/BTCUSDT'${NC}"
echo -e "${CYAN}curl '$API_URL/multi-analysis/ETHUSDT?timeframes=15m,1h,4h'${NC}"
echo -e "${CYAN}curl '$API_URL/multi-analysis/BNBUSDT' | jq '.data.summary'${NC}"
