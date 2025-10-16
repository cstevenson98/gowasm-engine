#!/bin/bash
# Run WebGPU tests in a real browser with WebGPU support

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Running WebGPU tests in Chrome with WebGPU enabled...${NC}"
echo ""

# Check if wasmbrowsertest is installed
if ! command -v go_js_wasm_exec &> /dev/null; then
    echo -e "${RED}ERROR: go_js_wasm_exec not found${NC}"
    echo -e "${YELLOW}Please install wasmbrowsertest:${NC}"
    echo "  go install github.com/agnivade/wasmbrowsertest@latest"
    echo "  mv \$(go env GOPATH)/bin/wasmbrowsertest \$(go env GOPATH)/bin/go_js_wasm_exec"
    exit 1
fi

# Find Chrome/Chromium
CHROME_PATH=""
for path in /usr/bin/google-chrome /usr/bin/chromium /usr/bin/chromium-browser /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome; do
    if [ -f "$path" ] || [ -f "${path}" ]; then
        CHROME_PATH="$path"
        break
    fi
done

if [ -z "$CHROME_PATH" ]; then
    echo -e "${RED}ERROR: Chrome/Chromium not found${NC}"
    echo -e "${YELLOW}Please install Chrome or Chromium${NC}"
    exit 1
fi

echo -e "${YELLOW}Browser: ${CHROME_PATH}${NC}"
echo -e "${YELLOW}Running tests in visible Chrome window with WebGPU enabled${NC}"
echo -e "${YELLOW}Note: Chrome window will open - keep it open until tests complete${NC}"
echo -e "${YELLOW}This may take 10-30 seconds...${NC}"
echo ""

# Set environment variables for wasmbrowsertest
export GOOS=js
export GOARCH=wasm
export BROWSER_PATH="$CHROME_PATH"
export HEADLESS=false

# Chrome flags to enable WebGPU (add to existing flags if any)
export CHROME_FLAGS="${CHROME_FLAGS} --enable-unsafe-webgpu --enable-features=Vulkan --use-vulkan=native"

# Run the canvas WebGPU tests specifically
echo -e "${BLUE}Testing: internal/canvas (WebGPU tests only)${NC}"
GOOS=js GOARCH=wasm go test -v ./internal/canvas -run "TestWebGPU" 2>&1 | grep -v "IPAddressSpace" || {
    echo ""
    echo -e "${YELLOW}Some tests may have failed - check output above${NC}"
    echo -e "${YELLOW}Note: Tests requiring actual WebGPU adapter will skip if WebGPU is unavailable${NC}"
}

echo ""
echo -e "${GREEN}✓ WebGPU browser tests completed${NC}"
echo ""
echo -e "${BLUE}Tip: To run all canvas tests (including mocks), use:${NC}"
echo -e "  make test-wasm-all"

