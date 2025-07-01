# Go Crypto Trading Analysis Makefile

.PHONY: build run test clean install deps help build-api run-api start-api

# Variables
APP_NAME=go-crypto
API_NAME=go-crypto-api
BUILD_DIR=build
MAIN_PATH=cmd/main.go
API_PATH=cmd/api/main.go

# Default target
help: ## Show this help message
	@echo "Go Crypto Trading Analysis"
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "🚀 Quick Start with API:"
	@echo "  make start-api           # Start REST API server"
	@echo "  ./scripts/api_demo.sh    # Run API demonstration"
	@echo "  Open http://localhost:8080 for web documentation"

deps: ## Install dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

build: deps ## Build the application
	@echo "Building application..."
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)

build-api: deps ## Build the API server
	@echo "Building API server..."
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(API_NAME) $(API_PATH)

run: deps ## Run the application
	@echo "Running application..."
	go run $(MAIN_PATH)

run-api: deps ## Run the API server
	@echo "Running API server..."
	go run $(API_PATH)

start-api: build-api ## Build and start the API server
	@echo "Starting API server on port 8080..."
	./$(BUILD_DIR)/$(API_NAME) -port=8080

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -f main analyzer streamer api-server go-crypto go-crypto-api
	rm -f *.test *.tmp *.log
	find . -name ".DS_Store" -delete 2>/dev/null || true

deep-clean: clean ## Deep clean including Go cache
	@echo "Deep cleaning..."
	go clean -cache
	go clean -modcache

format: ## Format code
	@echo "Formatting code..."
	go fmt ./...

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

install: build ## Install the application
	@echo "Installing application..."
	sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .

docker-run: ## Run in Docker container
	@echo "Running in Docker..."
	docker run --rm -it $(APP_NAME):latest

dev: ## Run in development mode with live reload
	@echo "Running in development mode..."
	air

setup: ## Setup development environment
	@echo "Setting up development environment..."
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

all: clean deps test build ## Run all: clean, deps, test, build
