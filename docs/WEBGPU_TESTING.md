# WebGPU Browser Testing Guide

## Overview

This project includes comprehensive browser tests for the WebGPU canvas implementation. These tests run in a real browser environment with access to the WebGPU API.

## Quick Start

### 1. Check WebGPU Support

First, verify that your browser supports WebGPU:

```bash
make serve
```

Then open: http://localhost:8080/test-webgpu-support.html

This page will:
- ✅ Check if WebGPU API exists
- ✅ Try to request a WebGPU adapter
- ✅ Show detailed adapter information
- ❌ Provide troubleshooting steps if WebGPU is unavailable

### 2. Run WebGPU Tests

```bash
make test-webgpu-browser
```

This command:
1. Opens Chrome with WebGPU flags enabled
2. Runs all `TestWebGPU*` tests from `internal/canvas/canvas_webgpu_test.go`
3. Tests will gracefully skip if WebGPU is unavailable

### 3. Run All WASM Tests

```bash
make test-wasm-all
```

This runs ALL tests (including WebGPU tests) but may show errors for packages not designed for WASM.

## Test Coverage

The WebGPU tests cover:

### Initialization & Configuration
- ✅ `TestNewWebGPUCanvasManager` - Constructor and initial state
- ✅ `TestWebGPUCanvasManager_Initialize` - Canvas setup and WebGPU initialization
- ✅ `TestWebGPUCanvasManager_GetStatus` - Status tracking

### Pipeline Management
- ✅ `TestWebGPUCanvasManager_SetPipelines` - Pipeline configuration
- ✅ `TestWebGPUCanvasManager_PipelineSwitching` - Dynamic pipeline changes

### Rendering
- ✅ `TestWebGPUCanvasManager_DrawColoredRect` - Colored rectangle rendering
- ✅ `TestWebGPUCanvasManager_DrawTexturedRect` - Textured sprite rendering
- ✅ `TestWebGPUCanvasManager_Render` - Frame rendering

### Batch Rendering
- ✅ `TestWebGPUCanvasManager_BatchRendering` - Batch mode operations
- ✅ `TestWebGPUCanvasManager_FlushBatch` - Vertex buffer flushing

### Coordinate System
- ✅ `TestWebGPUCanvasManager_CanvasToNDC` - Canvas to NDC transformations

### Textures
- ✅ `TestWebGPUCanvasManager_LoadTexture` - Texture loading
- ✅ `TestWebGPUCanvasManager_StubMethods` - Stub method implementations

### Vertex Generation
- ✅ `TestWebGPUCanvasManager_GenerateQuadVertices` - Colored quad vertices
- ✅ `TestWebGPUCanvasManager_GenerateTexturedQuadVertices` - Textured quad vertices

### Cleanup
- ✅ `TestWebGPUCanvasManager_Cleanup` - Resource cleanup

### Utilities
- ✅ `TestFloat32SliceToBytes` - Data conversion

**Total: 17 WebGPU browser tests**

## Enabling WebGPU Support

### Chrome/Chromium (Recommended)

**Method 1: Chrome Flags (Easiest)**

1. Open `chrome://flags`
2. Search for "WebGPU"
3. Enable these flags:
   - `#enable-unsafe-webgpu`
   - `#enable-webgpu-developer-features`
4. Restart Chrome

**Method 2: Command Line (Best for Testing)**

```bash
# Linux
google-chrome --enable-unsafe-webgpu --enable-features=Vulkan --use-vulkan=native

# macOS
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --enable-unsafe-webgpu

# Windows
"C:\Program Files\Google\Chrome\Application\chrome.exe" --enable-unsafe-webgpu
```

### Firefox

WebGPU support in Firefox is experimental. To enable:

1. Open `about:config`
2. Set `dom.webgpu.enabled` to `true`
3. Restart Firefox

**Note:** Firefox WebGPU support may be incomplete.

### Edge

Edge (Chromium) supports WebGPU similarly to Chrome:

1. Open `edge://flags`
2. Enable `#enable-unsafe-webgpu`
3. Restart Edge

## Platform-Specific Issues

### WSL2 (Windows Subsystem for Linux)

**Problem:** WebGPU may not work due to limited GPU access in WSL2.

**Solutions:**

1. **Use WSLg with GPU Passthrough** (Windows 11+)
   ```bash
   # Check if WSLg is installed
   wsl --version
   
   # Should show WSL version 2.0.0 or higher
   # Update WSL:
   wsl --update
   ```

2. **Run Tests in Native Windows**
   - Install Go on Windows
   - Run tests from Windows command prompt/PowerShell
   - Chrome will have full GPU access

3. **Use Mock Canvas Manager**
   - For unit testing without WebGPU
   - See `canvas_test.go` for examples

### Native Linux

**Requirements:**
- Up-to-date GPU drivers (NVIDIA, AMD, or Intel)
- Vulkan support
- Chrome/Chromium with WebGPU enabled

**Verify Vulkan:**
```bash
# Install vulkan-tools
sudo apt-get install vulkan-tools

# Check Vulkan support
vulkaninfo | grep "deviceName"
```

### macOS

**Requirements:**
- macOS 11 (Big Sur) or later
- Metal-compatible GPU
- Chrome 113+ or Safari 17+

**Note:** macOS uses Metal as the backend for WebGPU.

### Windows

**Requirements:**
- Windows 10 version 1809 or later
- DirectX 12 compatible GPU
- Up-to-date GPU drivers
- Chrome 113+ or Edge 113+

## Troubleshooting

### "no WebGPU adapter available"

**Cause:** Browser can't access GPU or WebGPU is disabled.

**Solutions:**
1. Enable WebGPU flags in browser (see above)
2. Update GPU drivers
3. Check hardware acceleration: `chrome://gpu`
4. Try running Chrome with `--enable-unsafe-webgpu` flag

### "WebGPU not supported" in test-webgpu-support.html

**Cause:** Browser doesn't have WebGPU API.

**Solutions:**
1. Update Chrome to version 113 or later
2. Enable WebGPU flags in `chrome://flags`
3. Try a different browser (Edge, Chrome Canary)

### Tests open browser but immediately fail

**Cause:** Browser opens but WebGPU adapter request fails.

**Solutions:**
1. Check `chrome://gpu` for errors
2. Ensure hardware acceleration is enabled
3. Update GPU drivers
4. Try running with `--use-vulkan=native` flag (Linux)

### Browser window closes too fast

**Cause:** Tests complete before you can see results.

**Solutions:**
```bash
# Run with verbose output
GOOS=js GOARCH=wasm go test -v ./internal/canvas -run TestWebGPU

# Or use the built-in script
./scripts/test-webgpu-browser.sh
```

### WSL2: "Failed to open X display"

**Cause:** No X11 server running.

**Solutions:**
1. Install WSLg (comes with Windows 11)
2. Or use native Windows for testing
3. Or run tests in headless mode (won't support WebGPU fully)

## Running Tests Manually

For more control, run tests directly:

```bash
# Set environment for WASM
export GOOS=js
export GOARCH=wasm

# Run specific test
go test -v ./internal/canvas -run TestWebGPUCanvasManager_Initialize

# Run all WebGPU tests
go test -v ./internal/canvas -run TestWebGPU

# Run with custom browser flags
export CHROME_FLAGS="--enable-unsafe-webgpu --enable-features=Vulkan"
go test -v ./internal/canvas
```

## CI/CD Integration

For automated testing in CI/CD:

```yaml
# GitHub Actions example
- name: Install Chrome
  run: |
    wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | sudo apt-key add -
    echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" | sudo tee /etc/apt/sources.list.d/google-chrome.list
    sudo apt-get update
    sudo apt-get install google-chrome-stable

- name: Run WebGPU Tests
  run: make test-webgpu-browser
  env:
    HEADLESS: "true"  # Note: WebGPU may not work in headless mode
```

**Important:** Most CI environments don't have GPU access, so WebGPU tests may need to:
- Run in headless mode (limited functionality)
- Use software rendering (slow)
- Skip WebGPU-specific tests

Consider using mock canvas manager for CI testing.

## Additional Resources

- [WebGPU Specification](https://www.w3.org/TR/webgpu/)
- [WebGPU Best Practices](https://toji.github.io/webgpu-best-practices/)
- [Chrome WebGPU Status](https://chromestatus.com/feature/6213121689518080)
- [cogentcore/webgpu Go Bindings](https://github.com/cogentcore/webgpu)

## Need Help?

1. Check the test-webgpu-support.html page for specific error messages
2. Review browser console for detailed errors
3. Check `chrome://gpu` for GPU/WebGPU status
4. Ensure you're using Chrome 113+ or equivalent
5. Try running tests with verbose output: `-v` flag

---

**Last Updated:** 2025-10-16

