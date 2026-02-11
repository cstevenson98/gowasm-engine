# Dev Container Quick Start

## 🚀 Getting Started (30 seconds)

### Open in Container
1. Open this project in VSCode
2. Click the popup "Reopen in Container" (or press `F1` → "Dev Containers: Reopen in Container")
3. Wait ~2-3 minutes for initial setup

### Verify Setup
```bash
wasm-env-info
```

You should see:
- ✅ Go 1.24.x
- ✅ TinyGo 0.34.0
- ✅ Python 3.11.x
- ✅ Node v20.x.x

## 🎮 Build and Run

```bash
# Build all examples
cd examples
make build

# Start dev server
make serve

# Open browser to http://localhost:8080
```

## 🔧 Common Commands

| Command | Description |
|---------|-------------|
| `make build` | Build all examples |
| `make serve` | Start dev server |
| `make clean` | Clean build artifacts |
| `make info` | Show build configuration |
| `go test ./pkg/...` | Run engine tests |
| `locate-wasm-exec` | Find wasm_exec.js paths |
| `wasm-env-info` | Show environment details |

## 📁 Where is wasm_exec.js?

The dev container automatically locates WASM runtime files:

```bash
# Find both Go and TinyGo versions
locate-wasm-exec
```

**Standard Go**: `$(GOROOT)/lib/wasm/wasm_exec.js` - Use with standard Go WASM builds (Go 1.24+)
  - **Note**: Older Go versions (1.23 and below) use `$(GOROOT)/misc/wasm/wasm_exec.js`

**TinyGo**: `$(TINYGOROOT)/targets/wasm_exec.js` - Use with TinyGo WASM builds

The Makefile handles this automatically, but if you need to copy manually:

```bash
# Standard Go (1.24+)
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ./assets/

# Standard Go (1.23 and older)
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ./assets/

# TinyGo
cp "$(tinygo env TINYGOROOT)/targets/wasm_exec.js" ./assets/wasm_exec_tinygo.js
```

## 🏗️ Building WASM

### With Standard Go (full stdlib, larger binary)
```bash
GOOS=js GOARCH=wasm go build -o main.wasm ./game
```

### With TinyGo (smaller binary, limited stdlib)
```bash
tinygo build -o main.wasm -target wasm ./game
```

## 🧪 Testing

```bash
# Standard tests (fast, no browser)
go test ./pkg/...

# Run with coverage
go test -cover ./pkg/...

# Specific package
go test ./pkg/canvas/
```

## 🐛 Troubleshooting

### Port Already in Use
The dev server automatically finds an available port starting from 8080.

### wasm_exec.js Not Found
```bash
locate-wasm-exec  # Shows correct paths
```

### Module Changes Not Reflected
```bash
go clean -modcache
go mod tidy
cd examples/basic-game && go mod tidy
```

### TinyGo Not Found
```bash
# Restart terminal or source the profile
source /etc/profile.d/tinygo.sh
```

## 📚 Project Structure

```
gowasm-engine/
├── pkg/                    # Engine library
│   ├── canvas/             # WebGPU rendering
│   ├── engine/             # Game loop
│   ├── gameobject/         # Game entities
│   └── ...
├── examples/
│   ├── basic-game/         # Example game
│   │   ├── game/           # Game code
│   │   └── assets/         # Assets
│   └── Makefile            # Build all examples
└── .devcontainer/          # This configuration
```

## 💡 Tips

- **Auto-save**: Enabled by default for Go files
- **Format on save**: Automatically formats Go code
- **Organize imports**: Automatically on save
- **Ports**: 8080-8082 are auto-forwarded
- **Go modules**: Cached in a volume for speed

## 🔗 Useful Links

- [WebGPU Spec](https://www.w3.org/TR/webgpu/)
- [Go WASM](https://github.com/golang/go/wiki/WebAssembly)
- [TinyGo WASM](https://tinygo.org/docs/guides/webassembly/)
- [cogentcore/webgpu](https://github.com/cogentcore/webgpu)

## Next Steps

1. Explore `examples/basic-game/` to see how the engine works
2. Read `pkg/README_TESTING.md` for testing strategies
3. Check `docs/ARCHITECTURE.md` for engine design
4. Build something cool! 🎨
