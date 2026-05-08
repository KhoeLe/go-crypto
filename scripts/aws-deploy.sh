#!/bin/bash

# AWS Lambda Deployment Script for Go Crypto API
# This script performs a complete deployment setup including IAM roles and Lambda function

set -e

# Configuration
FUNCTION_NAME="${FUNCTION_NAME:-go-crypto-api-sg}"
REGION="${AWS_REGION:-ap-southeast-1}"
ROLE_NAME="lambda-execution-role"
BUILD_DIR="build"
LAMBDA_ZIP="lambda.zip"

echo "🚀 AWS Lambda Deployment: Go Crypto API"
echo "========================================"
echo "Function: $FUNCTION_NAME"
echo "Region: $REGION"
echo "API Gateway stage path: /prod/api/v1"
echo ""

# Check if AWS CLI is configured
if ! aws sts get-caller-identity > /dev/null 2>&1; then
    echo "❌ AWS CLI not configured. Please run 'aws configure' first."
    exit 1
fi

# Get account ID
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
echo "📋 AWS Account ID: $ACCOUNT_ID"

# Check if IAM role exists, create if not
echo "🔐 Checking IAM execution role..."
if ! aws iam get-role --role-name "$ROLE_NAME" > /dev/null 2>&1; then
    echo "Creating IAM execution role..."
    
    # Create role with trust policy
    aws iam create-role \
        --role-name "$ROLE_NAME" \
        --assume-role-policy-document '{
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Effect": "Allow",
                    "Principal": {
                        "Service": "lambda.amazonaws.com"
                    },
                    "Action": "sts:AssumeRole"
                }
            ]
        }' \
        --region "$REGION"
    
    # Attach basic execution policy
    aws iam attach-role-policy \
        --role-name "$ROLE_NAME" \
        --policy-arn "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole" \
        --region "$REGION"
    
    echo "✅ IAM role created successfully"
    
    # Wait for role to be available
    echo "⏳ Waiting for IAM role to be available..."
    sleep 10
else
    echo "✅ IAM role already exists"
fi

# Build and package
echo "🔨 Building Lambda function..."
make build-lambda

echo "📦 Packaging Lambda function..."
make package-lambda

# Check if Lambda function exists
ROLE_ARN="arn:aws:iam::$ACCOUNT_ID:role/$ROLE_NAME"

echo "🔍 Checking if Lambda function exists..."
if aws lambda get-function --function-name "$FUNCTION_NAME" --region "$REGION" > /dev/null 2>&1; then
    echo "📝 Function exists, updating code..."
    
    # Update function code
    aws lambda update-function-code \
        --function-name "$FUNCTION_NAME" \
        --zip-file "fileb://$BUILD_DIR/$LAMBDA_ZIP" \
        --region "$REGION"
    
    # Update function configuration
    aws lambda update-function-configuration \
        --function-name "$FUNCTION_NAME" \
        --timeout 30 \
        --memory-size 256 \
        --environment Variables='{
            "BINANCE_BASE_URL":"https://api.binance.com",
            "API_TIMEOUT":"30",
            "LOG_LEVEL":"info"
        }' \
        --region "$REGION"
    
    echo "✅ Function updated successfully"
else
    echo "🆕 Creating new Lambda function..."
    
    # Create function
    aws lambda create-function \
        --function-name "$FUNCTION_NAME" \
        --runtime "provided.al2" \
        --role "$ROLE_ARN" \
        --handler "bootstrap" \
        --zip-file "fileb://$BUILD_DIR/$LAMBDA_ZIP" \
        --timeout 30 \
        --memory-size 256 \
        --environment Variables='{
            "BINANCE_BASE_URL":"https://api.binance.com",
            "API_TIMEOUT":"30",
            "LOG_LEVEL":"info"
        }' \
        --region "$REGION"
    
    echo "✅ Function created successfully"
fi

# Wait for deployment to complete
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
    cat response.json | jq . 2>/dev/null || cat response.json
    rm -f response.json
else
    echo "❌ Health check failed"
    cat response.json | jq . 2>/dev/null || cat response.json
    exit 1
fi

# Test futures price endpoint
echo "Testing XAU futures price endpoint..."
aws lambda invoke \
    --function-name "$FUNCTION_NAME" \
    --payload '{"httpMethod":"GET","path":"/prod/api/v1/price/XAUUSDT"}' \
    --region "$REGION" \
    response.json > /dev/null

status_code=$(jq -r '.statusCode // empty' response.json 2>/dev/null || true)
if [ $? -eq 0 ] && [ "$status_code" = "200" ]; then
    echo "✅ XAU price endpoint test passed"
    cat response.json | jq . 2>/dev/null || cat response.json
    rm -f response.json
else
    echo "❌ XAU price endpoint test failed"
    cat response.json | jq . 2>/dev/null || cat response.json
    exit 1
fi

echo ""
echo "🎉 Deployment completed successfully!"
echo "Function: $FUNCTION_NAME"
echo "Region: $REGION"
echo "Role: $ROLE_ARN"
echo ""
echo "📚 Next steps:"
echo "1. Test all endpoints: ./scripts/test-lambda.sh all"
echo "2. View logs: make lambda-logs"
echo "3. Set up API Gateway (optional): see LAMBDA_DEPLOYMENT.md"
echo ""
echo "💰 Expected monthly cost: ~$1-2 for typical usage"
