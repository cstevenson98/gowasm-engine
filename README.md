# Go WASM WebGPU 2D Game Engine (Library)

A component-based 2D game engine written in Go, compiled to WebAssembly, and rendered with WebGPU via the `cogentcore/webgpu` wrapper. It is designed as a reusable library with clear interfaces for scenes, sprites, movers, input, and rendering. Example games live under `examples/` and consume the engine as a module.

## Quick Start

Prerequisites:
- Go 1.24+
- A WebGPU-capable browser (recent Chromium-based browser with WebGPU enabled)

Build and serve the examples:

```bash
make -C examples list
make -C examples build
make -C examples serve
# open the printed example URL(s)
```

Use as a library in your own project (local dev with replace):

```go
// go.mod (in your game)
require github.com/cstevenson98/gowasm-engine v0.0.0
replace github.com/cstevenson98/gowasm-engine => ../path/to/engine/repo
```

```go
// main.go (WASM entrypoint with //go:build js)
eng := engine.NewEngine()
myScene := NewMyScene()  // Input will be injected automatically
eng.RegisterScene(types.GAMEPLAY, myScene)
_ = eng.Initialize("canvas-id")
_ = eng.SetGameState(types.GAMEPLAY)  // Input injected here if scene implements SceneInputProvider
eng.Start()
```

## Using from a Private GitHub Repository

Since this repository is private, using it as a Go module requires authentication configuration. Here are the options:

### Option 1: Local Development with `replace` (Recommended for Development)

For local development, use a `replace` directive in your project's `go.mod`:

```go
module your-game

go 1.24

require github.com/cstevenson98/gowasm-engine v0.0.0

replace github.com/cstevenson98/gowasm-engine => ../path/to/gowasm-engine
```

This allows you to:
- Work with local changes without committing
- Test modifications immediately
- Avoid authentication setup during development

### Option 2: Authenticated Access (For CI/CD or Remote Use)

For using the module from the actual GitHub repository (CI/CD, deployment, or when working from different machines):

#### Step 1: Configure Go to Skip Public Proxy

Set `GOPRIVATE` to tell Go not to use the public module proxy for this module:

```bash
# For this module only
go env -w GOPRIVATE=github.com/cstevenson98/gowasm-engine

# Or for all modules under your GitHub organization
go env -w GOPRIVATE=github.com/cstevenson98/*
```

#### Step 2: Configure Git Authentication

Choose one of the following authentication methods:

**A. SSH Keys (Recommended)**

1. Ensure you have an SSH key set up with GitHub:
   ```bash
   ssh -T git@github.com  # Test connection
   ```

2. Configure git to use SSH for GitHub:
   ```bash
   git config --global url."git@github.com:".insteadOf "https://github.com/"
   ```

3. Your `go.mod` will use the module normally:
   ```go
   require github.com/cstevenson98/gowasm-engine v0.1.0
   ```

**B. Personal Access Token (PAT)**

1. Create a GitHub Personal Access Token with `repo` scope:
   - Go to GitHub Settings → Developer settings → Personal access tokens
   - Generate a token with `repo` permissions

2. Configure git credentials:
   ```bash
   # Option 1: Store in netrc file (~/.netrc)
   machine github.com
   login your-username
   password your-token
   
   # Option 2: Use git credential helper
   git config --global credential.helper store
   # Then on first clone/pull, enter username and token as password
   ```

3. Use HTTPS URLs in your `go.mod`:
   ```go
   require github.com/cstevenson98/gowasm-engine v0.1.0
   ```

**C. GitHub CLI Authentication**

If you use GitHub CLI (`gh`):

```bash
gh auth login
# This sets up authentication that Go will use
```

#### Step 3: Version Your Module

To use specific versions, tag your repository:

```bash
# In the engine repository
git tag v0.1.0
git push origin v0.1.0
```

Then in your game project:

```go
require github.com/cstevenson98/gowasm-engine v0.1.0
```

### Option 3: GONOPROXY and GONOSUMDB (Advanced)

For complete control over module fetching:

```bash
# Don't use proxy for private modules
go env -w GONOPROXY=github.com/cstevenson98/*

# Don't verify checksums from public sumdb for private modules
go env -w GONOSUMDB=github.com/cstevenson98/*
```

### Quick Setup Script

For a quick setup, create a `.envrc` file (if using direnv) or a setup script:

```bash
#!/bin/bash
# setup-private-module.sh

# Configure Go for private module
go env -w GOPRIVATE=github.com/cstevenson98/gowasm-engine

# Ensure SSH is configured for GitHub
git config --global url."git@github.com:".insteadOf "https://github.com/"

echo "Private module configured! Run: go get github.com/cstevenson98/gowasm-engine@latest"
```

### Troubleshooting

**Error: `go get` fails with authentication error**
- Verify `GOPRIVATE` is set correctly: `go env GOPRIVATE`
- Test git access: `git ls-remote git@github.com:cstevenson98/gowasm-engine.git`
- For HTTPS, verify credentials: `git config --global credential.helper`

**Error: `go mod tidy` fails**
- Ensure you're authenticated with GitHub (test with `gh auth status`)
- Verify the repository exists and you have access
- Check that `GOPRIVATE` includes your module path

**CI/CD Setup**
- For GitHub Actions: Use the built-in `GITHUB_TOKEN` (automatically configured)
- For other CI: Set up SSH keys or use PAT as secrets
- Don't forget to set `GOPRIVATE` in your CI environment

### Recommended Workflow

1. **Development**: Use `replace` directive for fast iteration
2. **Version control**: Commit `go.mod` with version pin (remove `replace` for releases)
3. **CI/CD**: Use authenticated access with version tags
4. **Teams**: Share SSH key setup or use GitHub PATs with team members

## Architecture Overview

High-level flow: Input → Scene.Update → Scene.GetRenderables → Canvas batching → WebGPU

- WASM boundary: Files with `//go:build js` can access browser APIs (DOM, timing). All WebGPU calls go through the `cogentcore/webgpu` wrapper, minimizing direct `syscall/js` usage.
- Engine owns the main loop, input, and scene orchestration; scenes own game state; canvas owns GPU details and batching.

## Core Packages and Responsibilities

- `pkg/engine`
  - Game loop (requestAnimationFrame), delta time, render loop
  - Engine state: current scene and pipelines by `types.GameState`
  - Scene registration (`RegisterScene`) and state switching (`SetGameState`)
  - Owns and initializes input; injects input into scenes via `SceneInputProvider` interface
  - Loads textures required by current scene

- `pkg/canvas`
  - WebGPU abstraction (via `cogentcore/webgpu`)
  - Pipeline setup (e.g., textured pipeline)
  - Texture management and batched sprite rendering (batch per texture)
  - Helpers to draw textured quads and begin/end batches

- `pkg/scene`
  - Scene interfaces and render layers (`SceneLayer`: BACKGROUND, ENTITIES, UI)
  - Scenes implement: `Initialize()`, `Update(dt)`, `GetRenderables()`, `Cleanup()`, `GetName()`

- `pkg/types`
  - Shared interfaces and types: `GameObject`, `Sprite`, `Mover`, `InputCapturer`, `Vector2`, `UVRect`, `Pipeline`, `GameState`, etc.
  - Optional scene extension interfaces:
    - `SceneInputProvider` with `SetInputCapturer(inputCapturer)` (receive engine's input capturer)
    - `SceneOverlayRenderer` with `RenderOverlays()` (HUD/menus/debug rendered inside batch)
    - `SceneTextureProvider` with `GetExtraTexturePaths() []string` (extra textures to preload)

- `pkg/sprite`
  - Sprite sheet representation, UV calculations, animation frame management
  - Produce `SpriteRenderData` (texture path, position, size, UV, visibility)

- `pkg/mover`
  - Movement integration (velocity, update per frame), screen bounds, wrapping

- `pkg/input`
  - Unified input for keyboard and gamepad
  - Thread-safe state read via `GetInputState()`

- `pkg/text`, `pkg/debug`
  - Text rendering from sprite fonts and optional debug console overlay

## Rendering Pipeline

- Pipelines: the engine configures pipelines per `types.GameState` (e.g., `TexturedPipeline`).
- Batching: draw calls are batched by texture to minimize bind group switches. When the texture changes, a new batch begins.
- Vertex generation: positions/sizes are converted to NDC; for pixel art, integer snapping and scaling are applied in the canvas layer.
- Texture loading: engine preloads textures used by renderables and any extra paths provided by the scene via `SceneTextureProvider`.

## Input System

- Engine owns the `InputCapturer` and initializes it during `Initialize()`.
- Scenes receive the engine's input capturer automatically if they implement the `SceneInputProvider` interface. The engine injects it during scene initialization when `SetGameState()` is called.
- This ensures listeners are registered once and input state is shared across scenes.

## Scenes and Extensibility

- Implement the `Scene` interface to define your game state. Typical lifecycle:
  - `Initialize()`: create objects, layers, and resources
  - `Update(dt)`: update movers, sprites, and gameplay logic; read input from `InputCapturer`
  - `GetRenderables()`: return objects in render order (layered)
  - `Cleanup()`: release references/resources
- Register scenes with `engine.RegisterScene(state, scene)` and set the state with `engine.SetGameState(state)`.
- Optional: implement `SceneInputProvider` to receive input from the engine, `SceneOverlayRenderer` for batched HUD/menus, and `SceneTextureProvider` for extra preloads (e.g., font textures).

## Configuration

The global configuration lives in `pkg/config` as `config.Global` with:
- `Screen`: virtual width/height, canvas width/height
- `Player`: spawn, size, speed, texture, sprite grid
- `Animation`: default frame times
- `Rendering`: `PixelArtMode`, `TextureFiltering`, `PixelPerfectScaling`, `PixelScale`, `UILineSpacing`, `TextLineSpacing`
- `Debug`: console toggle, font path/scale, colors, message settings
- `Battle`: example game parameters (used in examples)

Canvas creation: examples create the canvas element at runtime and pass its `id` to `engine.Initialize(canvasID)`.

## Build, Test, and Run

Library (root):

```bash
make test       # pkg/...
make test-all   # ./...
make tidy
```

Examples (multi-example orchestrator):

```bash
make -C examples list
make -C examples build
make -C examples serve   # serves examples/dist on an available port
```

Notes:
- WASM builds require `GOOS=js GOARCH=wasm` (handled by the examples Makefile).
- Use a WebGPU-capable browser; ensure WebGPU is enabled.

## Directory Layout

```
pkg/
  battle/         # Battle system primitives (used by examples)
  canvas/         # WebGPU wrapper integration, pipelines, batching
  config/         # Global configuration (config.Global)
  debug/          # Optional debug console
  engine/         # Engine loop, scene orchestration, input ownership
  gameobject/     # Example game objects (shareable components)
  input/          # Unified input system
  mover/          # Movement/physics helpers
  scene/          # Scene interfaces and layers
  sprite/         # Sprite sheet and animation
  text/           # Text rendering
  types/          # Shared types and interfaces

examples/
  Makefile        # Builds all examples to examples/build, provisions examples/dist
  basic-game/
    assets/       # Example-specific assets
    game/         # WASM entrypoint
    scenes/       # Game-specific scenes (moved out of the library)
    go.mod        # Separate module importing the engine
```

## Using as a Library

Minimal pattern:

```go
eng := engine.NewEngine()
scene := NewMyScene()  // Input injected automatically if scene implements SceneInputProvider
eng.RegisterScene(types.GAMEPLAY, scene)
_ = eng.Initialize("canvas-id")
_ = eng.SetGameState(types.GAMEPLAY)  // Input injected here
eng.Start()
```

Local development with replace in your game’s `go.mod`:

```go
require github.com/cstevenson98/gowasm-engine v0.0.0
replace github.com/cstevenson98/gowasm-engine => ../path/to/engine/repo
```

## Performance Notes

- Batch by texture to minimize pipeline/bind group switches
- Prefer texture atlases to increase batch sizes
- Minimize per-frame allocations; consider object pooling
- Use browser performance tools for profiling

## Troubleshooting / FAQ

- WebGPU not available: ensure you’re using a supported browser and WebGPU is enabled
- Port in use: `make -C examples serve` auto-picks a free port
- Assets missing in dist: ensure your example has `assets/` and the Makefile copied them
- Input not registering: ensure your scene implements `SceneInputProvider` interface to receive input from the engine
- Build tags: WASM files must include `//go:build js`

## Examples

Examples live under `examples/` and can be built and served with the examples Makefile. They are intentionally separate modules and are not explained in detail here.


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

