#!/bin/bash

# Cleanup script for Go Crypto project
# This script removes build artifacts and temporary files

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧹 Go Crypto Project Cleanup${NC}"
echo -e "${BLUE}============================${NC}"
echo ""

# Clean build artifacts
echo -e "${YELLOW}Cleaning build artifacts...${NC}"
rm -rf build/
rm -f main analyzer streamer api-server go-crypto go-crypto-api
echo -e "${GREEN}✅ Build artifacts cleaned${NC}"

# Clean test artifacts
echo -e "${YELLOW}Cleaning test artifacts...${NC}"
rm -f coverage.out coverage.html
rm -f *.test
echo -e "${GREEN}✅ Test artifacts cleaned${NC}"

# Clean temporary files
echo -e "${YELLOW}Cleaning temporary files...${NC}"
find . -name "*.tmp" -delete 2>/dev/null || true
find . -name "*.log" -delete 2>/dev/null || true
find . -name ".DS_Store" -delete 2>/dev/null || true
echo -e "${GREEN}✅ Temporary files cleaned${NC}"

# Clean Go cache (optional)
if [ "$1" = "--deep" ]; then
    echo -e "${YELLOW}Deep cleaning Go cache...${NC}"
    go clean -cache
    go clean -modcache
    echo -e "${GREEN}✅ Go cache cleaned${NC}"
fi

# Tidy dependencies
echo -e "${YELLOW}Tidying Go modules...${NC}"
go mod tidy
echo -e "${GREEN}✅ Go modules tidied${NC}"

echo ""
echo -e "${GREEN}🎉 Cleanup completed!${NC}"
echo ""
echo -e "${BLUE}Usage:${NC}"
echo -e "  ./scripts/cleanup.sh         # Standard cleanup"
echo -e "  ./scripts/cleanup.sh --deep  # Deep cleanup (includes Go cache)"
