#!/bin/bash

# Lambda Statistics Script for Go Crypto API
# This script fetches CloudWatch metrics for the Lambda function

set -e

# Configuration
FUNCTION_NAME=${1:-"go-crypto-api-sg"}
REGION=${2:-"ap-southeast-1"}
DAYS=${3:-1}  # Default to last 24 hours

# Calculate time parameters
START_TIME=$(date -u -v-${DAYS}d +%Y-%m-%dT%H:%M:%S)
END_TIME=$(date -u +%Y-%m-%dT%H:%M:%S)
PERIOD=$((86400 * DAYS))  # Period in seconds (86400 = 1 day)

echo "📊 Lambda Function Statistics: $FUNCTION_NAME"
echo "===========================================" 
echo "Region: $REGION"
echo "Time Period: Last $DAYS day(s)"
echo "From: $START_TIME"
echo "To: $END_TIME"
echo ""

echo "🔢 Invocation Count (How many times the function has been called):"
aws cloudwatch get-metric-statistics \
    --namespace AWS/Lambda \
    --metric-name Invocations \
    --dimensions Name=FunctionName,Value=$FUNCTION_NAME \
    --start-time "$START_TIME" \
    --end-time "$END_TIME" \
    --period $PERIOD \
    --statistics Sum \
    --region $REGION | jq '.Datapoints[0].Sum // 0'

echo ""
echo "⏱️ Average Duration (in milliseconds):"
aws cloudwatch get-metric-statistics \
    --namespace AWS/Lambda \
    --metric-name Duration \
    --dimensions Name=FunctionName,Value=$FUNCTION_NAME \
    --start-time "$START_TIME" \
    --end-time "$END_TIME" \
    --period $PERIOD \
    --statistics Average \
    --region $REGION | jq '.Datapoints[0].Average // 0'

echo ""
echo "❌ Error Count:"
aws cloudwatch get-metric-statistics \
    --namespace AWS/Lambda \
    --metric-name Errors \
    --dimensions Name=FunctionName,Value=$FUNCTION_NAME \
    --start-time "$START_TIME" \
    --end-time "$END_TIME" \
    --period $PERIOD \
    --statistics Sum \
    --region $REGION | jq '.Datapoints[0].Sum // 0'

echo ""
echo "🔒 Throttles (times function was throttled):"
aws cloudwatch get-metric-statistics \
    --namespace AWS/Lambda \
    --metric-name Throttles \
    --dimensions Name=FunctionName,Value=$FUNCTION_NAME \
    --start-time "$START_TIME" \
    --end-time "$END_TIME" \
    --period $PERIOD \
    --statistics Sum \
    --region $REGION | jq '.Datapoints[0].Sum // 0'

echo ""
echo "💾 Usage (GB-seconds):"
aws cloudwatch get-metric-statistics \
    --namespace AWS/Lambda \
    --metric-name BillableDuration \
    --dimensions Name=FunctionName,Value=$FUNCTION_NAME \
    --start-time "$START_TIME" \
    --end-time "$END_TIME" \
    --period $PERIOD \
    --statistics Sum \
    --unit Milliseconds \
    --region $REGION | jq '.Datapoints[0].Sum // 0 | . / 1000 / 1024' # Convert ms to GB-seconds

echo ""
echo "Usage Information:"
echo "- You can specify function name: ./scripts/lambda-stats.sh custom-function-name"
echo "- You can specify region: ./scripts/lambda-stats.sh go-crypto-api-sg us-east-1"
echo "- You can specify days: ./scripts/lambda-stats.sh go-crypto-api-sg ap-southeast-1 7"
