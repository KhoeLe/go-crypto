#!/bin/bash

# Quick Deploy Script for Go Crypto API Lambda Function
# This script quickly updates the existing Lambda function with new code

set -e

# Configuration
FUNCTION_NAME="${FUNCTION_NAME:-go-crypto-api-sg}"
REGION="${AWS_REGION:-ap-southeast-1}"
BUILD_DIR="build"
LAMBDA_ZIP="lambda.zip"

echo "🚀 Quick Deploy: Go Crypto API Lambda Function"
echo "=============================================="
echo "API Gateway stage path: /prod/api/v1"

# Check if AWS CLI is configured
if ! aws sts get-caller-identity > /dev/null 2>&1; then
    echo "❌ AWS CLI not configured. Please run 'aws configure' first."
    exit 1
fi

# Check if function exists
echo "📋 Checking if Lambda function exists..."
if ! aws lambda get-function --function-name "$FUNCTION_NAME" --region "$REGION" > /dev/null 2>&1; then
    echo "❌ Lambda function '$FUNCTION_NAME' not found in region '$REGION'"
    echo "   Please run './scripts/aws-deploy.sh' for initial deployment"
    exit 1
fi

# Build and package
echo "🔨 Building Lambda function..."
make build-lambda

echo "📦 Packaging Lambda function..."
make package-lambda

# Deploy to existing function
echo "🚀 Updating Lambda function code..."
aws lambda update-function-code \
    --function-name "$FUNCTION_NAME" \
    --zip-file "fileb://$BUILD_DIR/$LAMBDA_ZIP" \
    --region "$REGION"

if [ $? -eq 0 ]; then
    echo "✅ Lambda function updated successfully!"
    
    # Wait a moment for deployment to complete
    echo "⏳ Waiting for deployment to complete..."
    sleep 5
    
    # Test the deployment
    echo "🧪 Testing deployment..."
    
    # Test health endpoint
    echo "Testing health endpoint..."
    aws lambda invoke \
        --function-name "$FUNCTION_NAME" \
        --payload '{"httpMethod":"GET","path":"/prod/api/v1/health"}' \
        --region "$REGION" \
        response.json > /dev/null
    
    status_code=$(jq -r '.statusCode // empty' response.json 2>/dev/null || true)
    if [ $? -eq 0 ] && [ "$status_code" = "200" ]; then
        echo "✅ Health check passed"
        cat response.json | jq .
        rm -f response.json
    else
        echo "❌ Health check failed"
        cat response.json | jq . 2>/dev/null || cat response.json
        exit 1
    fi

    echo "Testing XAU futures analysis endpoint..."
    aws lambda invoke \
        --function-name "$FUNCTION_NAME" \
        --payload '{"httpMethod":"GET","path":"/prod/api/v1/analysis/XAUUSDT","queryStringParameters":{"interval":"1h"}}' \
        --region "$REGION" \
        response.json > /dev/null

    status_code=$(jq -r '.statusCode // empty' response.json 2>/dev/null || true)
    if [ "$status_code" = "200" ]; then
        echo "✅ XAU analysis check passed"
        cat response.json | jq '.statusCode, (.body | fromjson | .data.symbol), (.body | fromjson | .data.trend)' 2>/dev/null || cat response.json
        rm -f response.json
    else
        echo "❌ XAU analysis check failed"
        cat response.json | jq . 2>/dev/null || cat response.json
        exit 1
    fi
    
    echo ""
    echo "🎉 Quick deployment completed successfully!"
    echo "Function: $FUNCTION_NAME"
    echo "Region: $REGION"
    echo ""
    echo "📚 To test all endpoints, run:"
    echo "   ./scripts/test-lambda.sh all"
    echo ""
    echo "📝 To view logs, run:"
    echo "   make lambda-logs"
    echo ""
    echo "📊 To check function usage statistics (last 24 hours):"
    echo "   aws cloudwatch get-metric-statistics \\"
    echo "     --namespace AWS/Lambda \\"
    echo "     --metric-name Invocations \\"
    echo "     --dimensions Name=FunctionName,Value=$FUNCTION_NAME \\"
    echo "     --start-time \$(date -u -v-1d +%Y-%m-%dT%H:%M:%S) \\"
    echo "     --end-time \$(date -u +%Y-%m-%dT%H:%M:%S) \\"
    echo "     --period 86400 \\"
    echo "     --statistics Sum \\"
    echo "     --region $REGION"
    
else
    echo "❌ Failed to update Lambda function"
    exit 1
fi
