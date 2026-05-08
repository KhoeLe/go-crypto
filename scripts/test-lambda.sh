#!/bin/bash

# Lambda Function Testing Script for Go Crypto API
# Tests all available endpoints of the deployed Lambda function

set -e

# Configuration
FUNCTION_NAME="${FUNCTION_NAME:-go-crypto-api-sg}"
REGION="${AWS_REGION:-ap-southeast-1}"
API_PREFIX="/prod/api/v1"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results
PASSED=0
FAILED=0

echo -e "${BLUE}🧪 Testing Lambda Function: $FUNCTION_NAME${NC}"
echo "================================================"

# Function to test endpoint
test_endpoint() {
    local test_name="$1"
    local payload="$2"
    local expected_status="$3"
    
    echo -e "\n${YELLOW}Testing: $test_name${NC}"
    echo "Payload: $payload"
    
    # Invoke Lambda function
    if aws lambda invoke \
        --function-name "$FUNCTION_NAME" \
        --payload "$payload" \
        --region "$REGION" \
        response.json > invoke_output.txt 2>&1; then
        
        # Check if response file exists and has content
        if [ -f response.json ] && [ -s response.json ]; then
            echo "Response:"
            cat response.json | jq . 2>/dev/null || cat response.json
            
            # Parse status code from response
            status_code=$(cat response.json | jq -r '.statusCode' 2>/dev/null || echo "unknown")
            
            if [ "$status_code" = "$expected_status" ] || [ "$expected_status" = "any" ]; then
                echo -e "${GREEN}✅ $test_name: PASSED${NC}"
                ((PASSED++))
            else
                echo -e "${RED}❌ $test_name: FAILED (Expected status: $expected_status, Got: $status_code)${NC}"
                ((FAILED++))
            fi
        else
            echo -e "${RED}❌ $test_name: FAILED (Empty response)${NC}"
            ((FAILED++))
        fi
    else
        echo -e "${RED}❌ $test_name: FAILED (Invoke error)${NC}"
        cat invoke_output.txt
        ((FAILED++))
    fi
    
    # Cleanup
    rm -f response.json invoke_output.txt
}

# Check if function exists
echo "Checking if Lambda function exists..."
if ! aws lambda get-function --function-name "$FUNCTION_NAME" --region "$REGION" > /dev/null 2>&1; then
    echo -e "${RED}❌ Lambda function '$FUNCTION_NAME' not found in region '$REGION'${NC}"
    echo "Please deploy the function first using: make deploy-lambda"
    exit 1
fi

echo -e "${GREEN}✅ Lambda function found${NC}"

# Test cases based on command line argument
if [ "$1" = "health" ] || [ "$1" = "all" ]; then
    test_endpoint "Health Check" "{\"httpMethod\":\"GET\",\"path\":\"$API_PREFIX/health\"}" "200"
fi

if [ "$1" = "price" ] || [ "$1" = "all" ]; then
    test_endpoint "Price - BTCUSDT" "{\"httpMethod\":\"GET\",\"path\":\"$API_PREFIX/price/BTCUSDT\"}" "200"
    test_endpoint "Price - XAUUSDT" "{\"httpMethod\":\"GET\",\"path\":\"$API_PREFIX/price/XAUUSDT\"}" "200"
    test_endpoint "Price - XAGUSDT" "{\"httpMethod\":\"GET\",\"path\":\"$API_PREFIX/price/XAGUSDT\"}" "200"
    test_endpoint "Price - Invalid Symbol" "{\"httpMethod\":\"GET\",\"path\":\"$API_PREFIX/price/INVALID\"}" "400"
fi

if [ "$1" = "indicators" ] || [ "$1" = "all" ]; then
    test_endpoint "Indicators - BTCUSDT" "{\"httpMethod\":\"GET\",\"path\":\"$API_PREFIX/indicators/BTCUSDT\"}" "200"
    test_endpoint "Indicators - XAUUSDT" "{\"httpMethod\":\"GET\",\"path\":\"$API_PREFIX/indicators/XAUUSDT\",\"queryStringParameters\":{\"interval\":\"1h\"}}" "200"
    test_endpoint "Indicators - XAGUSDT" "{\"httpMethod\":\"GET\",\"path\":\"$API_PREFIX/indicators/XAGUSDT\",\"queryStringParameters\":{\"interval\":\"1h\"}}" "200"
fi

if [ "$1" = "analysis" ] || [ "$1" = "all" ]; then
    test_endpoint "Analysis - BTCUSDT" "{\"httpMethod\":\"GET\",\"path\":\"$API_PREFIX/analysis/BTCUSDT\"}" "200"
    test_endpoint "Analysis - XAUUSDT" "{\"httpMethod\":\"GET\",\"path\":\"$API_PREFIX/analysis/XAUUSDT\",\"queryStringParameters\":{\"interval\":\"1h\"}}" "200"
    test_endpoint "Analysis - XAGUSDT" "{\"httpMethod\":\"GET\",\"path\":\"$API_PREFIX/analysis/XAGUSDT\",\"queryStringParameters\":{\"interval\":\"1h\"}}" "200"
fi

# Handle specific symbol testing
if [ "$1" = "price" ] && [ ! -z "$2" ]; then
    test_endpoint "Price - $2" "{\"httpMethod\":\"GET\",\"path\":\"$API_PREFIX/price/$2\"}" "any"
fi

if [ "$1" = "indicators" ] && [ ! -z "$2" ]; then
    test_endpoint "Indicators - $2" "{\"httpMethod\":\"GET\",\"path\":\"$API_PREFIX/indicators/$2\"}" "any"
fi

if [ "$1" = "analysis" ] && [ ! -z "$2" ]; then
    test_endpoint "Analysis - $2" "{\"httpMethod\":\"GET\",\"path\":\"$API_PREFIX/analysis/$2\"}" "any"
fi

# Show usage if no arguments
if [ -z "$1" ]; then
    echo -e "${YELLOW}Usage:${NC}"
    echo "  $0 all                    # Test all endpoints"
    echo "  $0 health                 # Test health endpoint only"
    echo "  $0 price [SYMBOL]         # Test price endpoint(s)"
    echo "  $0 indicators [SYMBOL]    # Test indicators endpoint(s)"
    echo "  $0 analysis [SYMBOL]      # Test analysis endpoint(s)"
    echo ""
    echo -e "${YELLOW}Examples:${NC}"
    echo "  $0 all"
    echo "  $0 price BTCUSDT"
    echo "  $0 indicators XAUUSDT"
    exit 0
fi

# Summary
echo ""
echo "================================================"
echo -e "${BLUE}Test Summary:${NC}"
echo -e "${GREEN}✅ Passed: $PASSED${NC}"
echo -e "${RED}❌ Failed: $FAILED${NC}"
echo "================================================"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed. Check the output above.${NC}"
    exit 1
fi
