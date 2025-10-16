#!/bin/bash

# Test runner script for webgpu-triangle project

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 Running Tests for webgpu-triangle${NC}\n"

# Parse command line arguments
COVERAGE=false
VERBOSE=false
BENCH=false
HTML=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -b|--bench)
            BENCH=true
            shift
            ;;
        -h|--html)
            HTML=true
            COVERAGE=true  # HTML requires coverage
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [-c|--coverage] [-v|--verbose] [-b|--bench] [-h|--html]"
            exit 1
            ;;
    esac
done

# Build test command
TEST_CMD="go test ./internal/..."

if [ "$VERBOSE" = true ]; then
    TEST_CMD="$TEST_CMD -v"
fi

if [ "$COVERAGE" = true ]; then
    TEST_CMD="$TEST_CMD -coverprofile=coverage.out"
fi

if [ "$BENCH" = true ]; then
    TEST_CMD="$TEST_CMD -bench=. -benchmem"
fi

# Run tests
echo -e "${YELLOW}Running: $TEST_CMD${NC}\n"
eval $TEST_CMD

# Show coverage if requested
if [ "$COVERAGE" = true ]; then
    echo -e "\n${BLUE}📊 Coverage Summary:${NC}"
    go tool cover -func=coverage.out | tail -1
    
    if [ "$HTML" = true ]; then
        echo -e "\n${GREEN}✨ Opening coverage report in browser...${NC}"
        go tool cover -html=coverage.out
    fi
fi

echo -e "\n${GREEN}✅ Tests completed successfully!${NC}"

