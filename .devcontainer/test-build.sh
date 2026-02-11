#!/bin/bash
# Test script to verify the Dockerfile builds correctly
# Run this before committing changes to ensure the dev container will work

set -e

echo "=========================================="
echo "Testing Dev Container Dockerfile Build"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Change to the .devcontainer directory
cd "$(dirname "$0")"

# Build the image
echo "📦 Building Docker image..."
echo ""
docker build \
  --build-arg GO_VERSION=1.24.3 \
  --build-arg TINYGO_VERSION=0.34.0 \
  -t gowasm-engine-devcontainer:test \
  -f Dockerfile \
  .

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✅ Docker image built successfully${NC}"
else
    echo ""
    echo -e "${RED}❌ Docker build failed${NC}"
    exit 1
fi

echo ""
echo "=========================================="
echo "Running Verification Tests"
echo "=========================================="
echo ""

# Test 1: Check Go version
echo "🧪 Test 1: Checking Go version..."
GO_VERSION=$(docker run --rm gowasm-engine-devcontainer:test go version)
if [[ $GO_VERSION == *"go1.24"* ]]; then
    echo -e "${GREEN}✅ Go version correct: $GO_VERSION${NC}"
else
    echo -e "${RED}❌ Go version incorrect: $GO_VERSION${NC}"
    exit 1
fi

# Test 2: Check TinyGo version
echo ""
echo "🧪 Test 2: Checking TinyGo version..."
TINYGO_VERSION=$(docker run --rm gowasm-engine-devcontainer:test tinygo version)
if [[ $TINYGO_VERSION == *"0.34"* ]]; then
    echo -e "${GREEN}✅ TinyGo version correct: $TINYGO_VERSION${NC}"
else
    echo -e "${RED}❌ TinyGo version incorrect: $TINYGO_VERSION${NC}"
    exit 1
fi

# Test 3: Check Python version
echo ""
echo "🧪 Test 3: Checking Python version..."
PYTHON_VERSION=$(docker run --rm gowasm-engine-devcontainer:test python3 --version)
if [[ $PYTHON_VERSION == *"Python 3"* ]]; then
    echo -e "${GREEN}✅ Python version correct: $PYTHON_VERSION${NC}"
else
    echo -e "${RED}❌ Python version incorrect: $PYTHON_VERSION${NC}"
    exit 1
fi

# Test 4: Check Node.js version
echo ""
echo "🧪 Test 4: Checking Node.js version..."
NODE_VERSION=$(docker run --rm gowasm-engine-devcontainer:test node --version)
if [[ $NODE_VERSION == *"v20"* ]]; then
    echo -e "${GREEN}✅ Node.js version correct: $NODE_VERSION${NC}"
else
    echo -e "${RED}❌ Node.js version incorrect: $NODE_VERSION${NC}"
    exit 1
fi

# Test 5: Check wasm_exec.js locations
echo ""
echo "🧪 Test 5: Checking wasm_exec.js files..."
docker run --rm gowasm-engine-devcontainer:test bash -c '
    # Go 1.24+ uses lib/wasm, older versions use misc/wasm
    GO_WASM_EXEC="$(go env GOROOT)/lib/wasm/wasm_exec.js"
    GO_WASM_EXEC_LEGACY="$(go env GOROOT)/misc/wasm/wasm_exec.js"
    TINYGO_WASM_EXEC="$(tinygo env TINYGOROOT)/targets/wasm_exec.js"
    
    if [ -f "$GO_WASM_EXEC" ]; then
        echo "✅ Go wasm_exec.js found: $GO_WASM_EXEC"
    elif [ -f "$GO_WASM_EXEC_LEGACY" ]; then
        echo "✅ Go wasm_exec.js found (legacy): $GO_WASM_EXEC_LEGACY"
    else
        echo "❌ Go wasm_exec.js NOT found"
        exit 1
    fi
    
    if [ -f "$TINYGO_WASM_EXEC" ]; then
        echo "✅ TinyGo wasm_exec.js found: $TINYGO_WASM_EXEC"
    else
        echo "❌ TinyGo wasm_exec.js NOT found"
        exit 1
    fi
'

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ All wasm_exec.js files found${NC}"
else
    echo -e "${RED}❌ Missing wasm_exec.js files${NC}"
    exit 1
fi

# Test 6: Check Go tools are installed
echo ""
echo "🧪 Test 6: Checking Go tools..."
docker run --rm gowasm-engine-devcontainer:test bash -c '
    if command -v gopls &> /dev/null; then
        echo "✅ gopls installed"
    else
        echo "❌ gopls NOT installed"
        exit 1
    fi
    
    if command -v dlv &> /dev/null; then
        echo "✅ delve installed"
    else
        echo "❌ delve NOT installed"
        exit 1
    fi
    
    if command -v staticcheck &> /dev/null; then
        echo "✅ staticcheck installed"
    else
        echo "❌ staticcheck NOT installed"
        exit 1
    fi
    
    if command -v golangci-lint &> /dev/null; then
        echo "✅ golangci-lint installed"
    else
        echo "❌ golangci-lint NOT installed"
        exit 1
    fi
'

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ All Go tools installed${NC}"
else
    echo -e "${RED}❌ Missing Go tools${NC}"
    exit 1
fi

# Test 7: Check helper scripts
echo ""
echo "🧪 Test 7: Checking helper scripts..."
docker run --rm gowasm-engine-devcontainer:test bash -c '
    if command -v locate-wasm-exec &> /dev/null; then
        echo "✅ locate-wasm-exec script installed"
    else
        echo "❌ locate-wasm-exec NOT installed"
        exit 1
    fi
    
    if command -v wasm-env-info &> /dev/null; then
        echo "✅ wasm-env-info script installed"
    else
        echo "❌ wasm-env-info NOT installed"
        exit 1
    fi
'

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ All helper scripts installed${NC}"
else
    echo -e "${RED}❌ Missing helper scripts${NC}"
    exit 1
fi

# Test 8: Run wasm-env-info to see full environment
echo ""
echo "🧪 Test 8: Running wasm-env-info..."
docker run --rm gowasm-engine-devcontainer:test wasm-env-info

# Test 9: Check user is vscode
echo ""
echo "🧪 Test 9: Checking default user..."
CONTAINER_USER=$(docker run --rm gowasm-engine-devcontainer:test whoami)
if [[ $CONTAINER_USER == "vscode" ]]; then
    echo -e "${GREEN}✅ Container user is 'vscode'${NC}"
else
    echo -e "${RED}❌ Container user is incorrect: $CONTAINER_USER${NC}"
    exit 1
fi

# Summary
echo ""
echo "=========================================="
echo -e "${GREEN}✅ All tests passed!${NC}"
echo "=========================================="
echo ""
echo "The Dockerfile builds correctly and includes:"
echo "  • Go 1.24.3"
echo "  • TinyGo 0.34.0"
echo "  • Python 3"
echo "  • Node.js 20"
echo "  • Go development tools (gopls, dlv, staticcheck, golangci-lint)"
echo "  • Helper scripts (locate-wasm-exec, wasm-env-info)"
echo "  • WASM runtime files (wasm_exec.js)"
echo ""
echo "You can now safely use this dev container in VSCode!"
echo ""
echo "To clean up the test image:"
echo "  docker rmi gowasm-engine-devcontainer:test"
echo ""
