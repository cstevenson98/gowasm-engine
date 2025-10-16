# WebGPU Triangle - Go WASM Game Engine

A 2D game engine built with Go and WebGPU, compiled to WebAssembly for browser execution.

## Features

- 🎮 **Player-controlled gameplay** with keyboard and gamepad support
- 🎨 **WebGPU rendering** with sprite animation and batching
- 🕹️ **Input system** supporting WASD keyboard and game controllers
- 🏃 **Component-based architecture** with GameObjects, Sprites, and Movers
- 🧪 **Comprehensive testing** including browser-based WASM tests
- 📦 **Efficient batching** for rendering multiple sprites

## Quick Start

### Prerequisites

- Go 1.21 or later
- Chrome or Chromium browser (for WASM tests)
- Make (optional, for convenience commands)

### Build and Run

```bash
# Build the WASM binary
make build

# Serve and open in browser
make serve

# Or do both at once
make quick
```

Then open your browser to `http://localhost:8080`

### Controls

- **WASD** - Move the llama player
- **Game Controller** - Left stick or D-pad for movement
- **1** - Switch to sprite rendering mode
- **2** - Switch to triangle mode

## Development

### Project Structure

```
webgpu-triangle/
├── cmd/game/           # Main application entry point
├── internal/
│   ├── canvas/         # WebGPU canvas management
│   ├── engine/         # Game engine core
│   ├── gameobject/     # GameObject implementations (Player, Llama)
│   ├── input/          # Input capture (keyboard, gamepad)
│   ├── mover/          # Movement and physics
│   ├── sprite/         # Sprite rendering and animation
│   └── types/          # Shared types and interfaces
├── assets/             # Game assets (textures, etc.)
└── dist/               # Built output for deployment
```

### Architecture

The engine follows a component-based architecture:

- **GameObject** - Game entities with Update() and GetSprite()
- **Sprite** - Handles texture, animation, and UV calculations
- **Mover** - Manages position, velocity, and screen wrapping
- **InputCapturer** - Captures keyboard and gamepad input
- **Engine** - Orchestrates game loop, rendering, and state

## Testing

### Standard Tests

Run all unit tests:
```bash
make test
```

Or with coverage:
```bash
./test.sh -c
```

Generate HTML coverage report:
```bash
./test.sh -h
```

### WASM Browser Tests

We use [wasmbrowsertest](https://github.com/agnivade/wasmbrowsertest) to run tests that require the browser environment.

**Setup (one-time):**
```bash
# Install wasmbrowsertest
go install github.com/agnivade/wasmbrowsertest@latest

# Rename to go_js_wasm_exec
mv $(go env GOPATH)/bin/wasmbrowsertest $(go env GOPATH)/bin/go_js_wasm_exec
```

**Run WASM tests:**
```bash
# Using Make
make test-wasm

# Or directly
GOOS=js GOARCH=wasm go test ./internal/gameobject -v
```

**Debug in browser (visible window):**
```bash
WASM_HEADLESS=off GOOS=js GOARCH=wasm go test ./internal/gameobject -v
```

### WebGPU Browser Testing

The `canvas_webgpu.go` implementation has comprehensive browser tests that verify WebGPU functionality in a real browser environment.

**Check WebGPU Support:**
```bash
# Open the WebGPU capability checker
make serve
# Then navigate to: http://localhost:8080/test-webgpu-support.html
```

**Run WebGPU Tests:**
```bash
# Run WebGPU tests in a visible Chrome window
make test-webgpu-browser

# This will:
# 1. Open Chrome with WebGPU flags enabled
# 2. Run canvas_webgpu tests in the browser
# 3. Tests will skip gracefully if WebGPU is unavailable
```

**WebGPU Test Coverage:**
- Canvas initialization and configuration
- Pipeline creation (triangle, sprite, textured)
- Pipeline switching and management
- Batch rendering and vertex buffering
- Texture loading and binding
- Coordinate system transformations (NDC)
- Resource cleanup and lifecycle

**Limitations in WSL2:**
WebGPU may not work in WSL2 without GPU passthrough. To enable:
1. Ensure you have WSLg installed (Windows 11 or WSL 2.0+)
2. Update GPU drivers on Windows
3. Run tests in native Windows Chrome, or
4. Use the mock canvas manager for testing without WebGPU

**Alternative: Run tests on native OS:**
The tests work best on:
- Native Linux with GPU drivers
- macOS with Metal support
- Native Windows with DirectX 12 support

### Test Coverage

| Package | Coverage | Type |
|---------|----------|------|
| input   | 100%     | Unit tests |
| mover   | 61.8%    | Unit tests |
| sprite  | 64.5%    | Unit tests |
| types   | 46.7%    | Unit tests |
| canvas  | 41.5%    | Unit tests + Mock |
| canvas  | ✅       | **WebGPU browser tests** |
| gameobject | ✅    | WASM browser tests |

**Total: 74 tests** (47 standard + 10 WASM + 17 WebGPU browser)

See [internal/README_TESTING.md](internal/README_TESTING.md) for detailed testing documentation.

## Makefile Commands

```bash
make deps               # Install dependencies
make build              # Build WASM binary
make serve              # Start development server
make quick              # Build and serve
make test               # Run standard tests
make test-wasm          # Run WASM browser tests
make test-wasm-all      # Run all WASM tests (all packages)
make test-webgpu-browser # Run WebGPU tests in visible Chrome
make clean              # Clean build artifacts
make prod               # Production build (optimized)
make info               # Show build information
```

## Components

### GameObject System

GameObjects represent entities in the game:
- **Player** - User-controlled character with input handling
- **Llama** - NPC character (example)

Each GameObject has:
- A **Sprite** for rendering
- A **Mover** for movement (optional)
- **State** for position and visibility

### Input System

Unified input system supporting:
- **Keyboard** - WASD keys
- **Gamepad** - Analog sticks and D-pad
- **Automatic detection** - Controllers hot-plugged automatically

### Rendering Pipeline

WebGPU-based rendering with:
- **Batch rendering** - Multiple sprites in one draw call
- **Texture atlas** support
- **Sprite sheet animation** - Frame-based animation with UV calculation
- **Pipeline switching** - Multiple render pipelines per game state

## Performance

- **60 FPS** target frame rate
- **Batch rendering** for efficient sprite drawing
- **Screen wrapping** with minimal overhead
- **Test execution** < 1 second for all tests

## Documentation

- [WASM Testing Guide](WASM_TESTING.md) - Detailed browser testing documentation
- [Testing Guide](internal/README_TESTING.md) - Complete testing reference
- [Architecture Overview](internal/) - Component documentation

## Browser Compatibility

Requires a browser with WebGPU support:
- Chrome 113+
- Edge 113+
- Opera 99+
- Firefox (experimental, behind flag)
- Safari (experimental, Technology Preview)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass: `make test && make test-wasm`
5. Submit a pull request

## License

MIT License - See LICENSE file for details

## Acknowledgments

- Built with [Go](https://golang.org/)
- Uses [wasmbrowsertest](https://github.com/agnivade/wasmbrowsertest) for browser testing
- Inspired by modern 2D game engines

## Troubleshooting

### WASM tests fail with "chrome not found"
Install Chrome or Chromium:
```bash
sudo apt-get install chromium-browser
```

### Build fails
Ensure you have Go 1.21+ and all dependencies:
```bash
go version
make deps
```

### Tests show "IPAddressSpace" errors
These are harmless ChromeDP warnings and can be ignored. The tests will still pass.

## Roadmap

- [ ] Collision detection system
- [ ] Audio support
- [ ] Particle effects
- [ ] Level/scene management
- [ ] Asset loading system
- [ ] Mobile touch controls
- [ ] Additional GameObject types
- [ ] Performance profiling tools

---

Built with ❤️ using Go and WebGPU

