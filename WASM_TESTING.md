# WASM Browser Testing Setup

## Overview

We've successfully integrated [wasmbrowsertest](https://github.com/agnivade/wasmbrowsertest) to run Go WASM tests directly in the browser. This allows us to test code that requires the `//go:build js` tag.

## Setup Complete ✅

The following has been configured:

1. **wasmbrowsertest installed** - Latest version from GitHub
2. **Binary renamed** - `go_js_wasm_exec` in `$GOPATH/bin`
3. **Chrome installed** - Required for running tests
4. **Tests passing** - All 10 player tests running successfully

## Quick Usage

### Run WASM Tests
```bash
GOOS=js GOARCH=wasm go test ./internal/gameobject -v
```

### Run WASM Tests Quietly
```bash
GOOS=js GOARCH=wasm go test ./internal/gameobject
```

### Run with Coverage (use gocoverdir)
```bash
GOOS=js GOARCH=wasm go test ./internal/gameobject -test.gocoverdir=/tmp/coverage
go tool covdata -i /tmp/coverage -o coverage.out
```

## How It Works

According to the [wasmbrowsertest documentation](https://github.com/agnivade/wasmbrowsertest):

1. Go's test runner looks for a binary named `go_js_wasm_exec` when `GOOS=js GOARCH=wasm`
2. wasmbrowsertest intercepts the test binary
3. It compiles the WASM, serves it with an HTTP server
4. Launches Chrome in headless mode (or visible if `WASM_HEADLESS=off`)
5. Runs the tests in the browser using ChromeDP protocol
6. Captures console output and returns test results

## Test Results

### All Tests Passing ✅

```
=== RUN   TestNewPlayer
--- PASS: TestNewPlayer (0.00s)
=== RUN   TestPlayerHandleInputNoMovement
--- PASS: TestPlayerHandleInputNoMovement (0.00s)
=== RUN   TestPlayerHandleInputSingleDirection
--- PASS: TestPlayerHandleInputSingleDirection (0.00s)
=== RUN   TestPlayerHandleInputDiagonalMovement
--- PASS: TestPlayerHandleInputDiagonalMovement (0.00s)
=== RUN   TestPlayerHandleInputOppositeDirections
--- PASS: TestPlayerHandleInputOppositeDirections (0.00s)
=== RUN   TestPlayerHandleInputDifferentSpeeds
--- PASS: TestPlayerHandleInputDifferentSpeeds (0.00s)
=== RUN   TestPlayerUpdate
--- PASS: TestPlayerUpdate (0.00s)
=== RUN   TestPlayerGetSetState
--- PASS: TestPlayerGetSetState (0.00s)
=== RUN   TestPlayerIntegration
--- PASS: TestPlayerIntegration (0.00s)
=== RUN   TestPlayerWithMockComponents
--- PASS: TestPlayerWithMockComponents (0.00s)
PASS
ok  	github.com/conor/webgpu-triangle/internal/gameobject	0.784s
```

## What Gets Tested

### Player GameObject (`internal/gameobject/player_test.go`)
- ✅ Player creation and initialization
- ✅ Input handling (WASD)
- ✅ Velocity calculations
- ✅ Diagonal movement normalization
- ✅ Opposite direction cancellation
- ✅ Different movement speeds
- ✅ State management (get/set)
- ✅ Integration with mover component
- ✅ Mock component interaction

### Future WASM Tests
- Llama GameObject
- Input system integration tests
- Engine with browser APIs
- Full game loop testing

## Debugging WASM Tests

### View Browser Window
```bash
WASM_HEADLESS=off GOOS=js GOARCH=wasm go test ./internal/gameobject -v
```

This will open a visible Chrome window so you can see the test execution.

### ChromeDP Errors
You may see some harmless errors like:
```
ERROR: could not unmarshal event: unknown IPAddressSpace value: Loopback
```

These are from the Chrome DevTools Protocol and don't affect test execution.

## CI/CD Integration

### GitHub Actions

Add to `.github/workflows/test.yml`:

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install Chrome
        uses: browser-actions/setup-chrome@latest
      
      - name: Install wasmbrowsertest
        run: |
          go install github.com/agnivade/wasmbrowsertest@latest
          mv $(go env GOPATH)/bin/wasmbrowsertest $(go env GOPATH)/bin/go_js_wasm_exec
      
      - name: Run Standard Tests
        run: go test ./internal/... -cover
      
      - name: Run WASM Tests
        run: GOOS=js GOARCH=wasm go test ./internal/gameobject -v
```

## Performance

### Test Execution Time
- **Standard tests**: ~0.01s per test
- **WASM tests**: ~0.78s total (includes browser startup)
- **Browser overhead**: ~300-500ms

The WASM tests are slightly slower due to:
1. WASM compilation
2. Chrome startup (headless)
3. HTTP server setup
4. ChromeDP connection

But still fast enough for rapid development (< 1 second total).

## Troubleshooting

### Chrome Not Found
```bash
# Install Chrome/Chromium
sudo apt-get install chromium-browser
# or
sudo apt-get install google-chrome-stable
```

### Binary Not Found
```bash
# Verify go_js_wasm_exec exists
which go_js_wasm_exec

# If not, reinstall
go install github.com/agnivade/wasmbrowsertest@latest
mv $(go env GOPATH)/bin/wasmbrowsertest $(go env GOPATH)/bin/go_js_wasm_exec
```

### Environment Variable Limit
If you see `total length of command line and environment variables exceeds limit`:

```bash
# Install cleanenv
go install github.com/agnivade/wasmbrowsertest/cmd/cleanenv@latest

# Use it to remove large env vars
cleanenv -remove-prefix GITHUB_ -- GOOS=js GOARCH=wasm go test ./internal/gameobject
```

## Advantages

1. **Real browser environment** - Tests run in actual Chrome
2. **Automatic** - No manual HTML/server setup needed
3. **Fast** - Headless mode is quick
4. **CI-friendly** - Easy to integrate
5. **Standard Go tests** - Same test syntax, just different target
6. **Debugging** - Can open visible browser with `WASM_HEADLESS=off`

## Limitations

1. **Chrome required** - Must have Chrome or Chromium installed
2. **Slower than unit tests** - ~100x slower than pure Go tests
3. **Headless only (default)** - Need flag to see visual output
4. **No WebGPU** - Browser APIs aren't fully accessible in tests

## Best Practices

1. **Keep WASM tests focused** - Test Go logic, not browser APIs
2. **Use mocks for WebGPU** - Mock canvas manager for rendering tests
3. **Separate concerns** - Standard tests for pure logic, WASM for js-specific code
4. **Fast feedback** - Run standard tests first, WASM tests in CI
5. **Document requirements** - Note Chrome dependency in README

## References

- [wasmbrowsertest GitHub](https://github.com/agnivade/wasmbrowsertest)
- [Go WASM Documentation](https://github.com/golang/go/wiki/WebAssembly)
- [ChromeDP Protocol](https://chromedevtools.github.io/devtools-protocol/)

## Summary

✅ **WASM testing fully operational**  
✅ **10 player tests passing in browser**  
✅ **CI/CD ready**  
✅ **Fast execution (< 1 second)**  
✅ **Easy to use** - Same as standard Go tests  

