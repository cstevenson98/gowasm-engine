# Dev Container for Go WASM Engine

This dev container provides a complete development environment for the Go WASM WebGPU Game Engine project.

## What's Included

### Languages & Runtimes
- **Go 1.24** - Primary development language
- **TinyGo 0.34.0** - Alternative WASM compiler for smaller binaries
- **Python 3.11** - For running dev server (`make serve`)
- **Node.js 20** - For web tooling (optional)

### Tools & Extensions
- Go language server (gopls)
- Delve debugger (dlv)
- golangci-lint - Code linter
- staticcheck - Static analysis
- VSCode Go extension
- Makefile support

### WASM Support
- Automatic location of `wasm_exec.js` runtime files
- Helper scripts for WASM development
- Pre-configured paths and environment variables

## Quick Start

### 0. Verify Dockerfile Builds (Optional but Recommended)

Before using the dev container, you can verify the Dockerfile builds correctly:

```bash
cd .devcontainer
./test-build.sh
```

This will:
- Build the Docker image
- Run verification tests for all components
- Display detailed environment information
- Confirm everything works before you commit to using it

### 1. Open in Dev Container

In VSCode:
1. Install the "Dev Containers" extension
2. Press `F1` and select "Dev Containers: Reopen in Container"
3. Wait for the container to build and initialize (~2-3 minutes first time)

### 2. Verify Environment

```bash
# Show environment info
wasm-env-info

# Locate WASM runtime files
locate-wasm-exec

# Check Go version
go version

# Check TinyGo version
tinygo version
```

### 3. Build and Run

```bash
# Build all examples
cd examples
make build

# Serve examples on http://localhost:8080
make serve

# Clean build artifacts
make clean
```

### 4. Build Individual Examples

```bash
cd examples
make basic-game
```

## Helper Scripts

The dev container includes two helper scripts:

### `wasm-env-info`
Displays complete environment information including:
- Go and TinyGo versions
- Python and Node versions
- Environment variables (GOROOT, GOPATH, TINYGOROOT)
- Location of WASM support files

### `locate-wasm-exec`
Finds and displays the paths to `wasm_exec.js` files:
- Standard Go: `$(GOROOT)/misc/wasm/wasm_exec.js`
- TinyGo: `$(TINYGOROOT)/targets/wasm_exec.js`

## WASM Compilation

### Standard Go WASM
```bash
GOOS=js GOARCH=wasm go build -o main.wasm ./game
```

The standard Go WASM compiler produces larger binaries but has full Go stdlib support.

### TinyGo WASM
```bash
tinygo build -o main.wasm -target wasm ./game
```

TinyGo produces much smaller binaries but has limited stdlib support. Great for production builds.

## Environment Variables

The dev container sets up the following environment variables:

```bash
GOROOT=/usr/local/go
GOPATH=/go
PATH=$PATH:/usr/local/go/bin:/go/bin:/usr/local/tinygo/bin
```

## Port Forwarding

The following ports are automatically forwarded:
- `8080` - Main dev server (default)
- `8081` - Alternative port
- `8082` - Alternative port

## Volume Mounts

- **Go module cache**: Persisted across container rebuilds for faster dependency downloads

## Updating the Container

### Update Go Version
Edit `.devcontainer/devcontainer.json`:
```json
"image": "mcr.microsoft.com/devcontainers/go:1-1.25-bookworm"
```

### Update TinyGo Version
Edit `.devcontainer/post-create.sh`:
```bash
TINYGO_VERSION="0.35.0"
```

### Add More Tools
Edit the `postCreateCommand` section in `devcontainer.json` or modify `post-create.sh`.

## Troubleshooting

### wasm_exec.js Not Found

Run `locate-wasm-exec` to find the correct paths. The Makefile should use:

```makefile
GOROOT:=$(shell go env GOROOT)
# Go 1.24+
WASM_EXEC_JS:=$(GOROOT)/lib/wasm/wasm_exec.js
# Go 1.23 and older
# WASM_EXEC_JS:=$(GOROOT)/misc/wasm/wasm_exec.js
```

**Note**: Go changed the location in version 1.24. The helper script checks both locations.

### TinyGo Command Not Found

```bash
# Verify TinyGo installation
which tinygo
tinygo version

# Check PATH
echo $PATH | grep tinygo
```

If TinyGo is not in PATH, restart the terminal or run:
```bash
source /etc/profile.d/tinygo.sh
```

### Permission Issues

The container runs as the `vscode` user. If you need root access:
```bash
sudo <command>
```

### Go Module Issues

```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download
go mod tidy
```

## Development Workflow

1. **Edit Code** - Make changes to Go files
2. **Run Tests** - `go test ./pkg/...`
3. **Build** - `cd examples && make build`
4. **Serve** - `make serve` (in examples/)
5. **Test in Browser** - Open http://localhost:8080

## Additional Resources

- [Go WebAssembly](https://github.com/golang/go/wiki/WebAssembly)
- [TinyGo WASM](https://tinygo.org/docs/guides/webassembly/)
- [WebGPU Spec](https://www.w3.org/TR/webgpu/)
- [cogentcore/webgpu](https://github.com/cogentcore/webgpu)

## Project Structure

```
gowasm-engine/
├── .devcontainer/          # Dev container configuration
│   ├── devcontainer.json   # Container definition
│   ├── post-create.sh      # Setup script
│   └── README.md           # This file
├── examples/               # Example games
│   ├── basic-game/         # Basic game example
│   │   ├── assets/         # Game assets
│   │   └── game/           # Game code
│   └── Makefile            # Build all examples
├── pkg/                    # Engine library code
│   ├── canvas/             # WebGPU rendering
│   ├── engine/             # Game loop
│   ├── gameobject/         # Game entities
│   ├── input/              # Input handling
│   ├── scene/              # Scene management
│   ├── sprite/             # Sprite rendering
│   └── types/              # Shared types
└── Makefile                # Root makefile
```
