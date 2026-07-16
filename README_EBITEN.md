# Ebiten Desktop Build Instructions

This document explains how to build and run the game engine as a native Linux desktop application using Ebiten.

## Quick Start (NixOS)

### Using the existing nix-shell environment

```bash
# Enter the development environment
nix-shell examples/ebiten-demo/shell.nix

# Build and run
make run-desktop
```

## Build Commands

### Build Desktop Binary

```bash
make build-desktop
```

This creates `build/game-desktop`

### Run Desktop Game

```bash
make run-desktop
```

This builds and runs the game in one step.

### Build WASM Version (Legacy)

```bash
make build-wasm
```

## Controls

### Keyboard
- **Arrow Keys / WASD**: Movement and menu navigation
- **Enter**: Confirm / Select
- **Space**: Cancel / Back
- **M**: Open player menu (in gameplay)
- **F2**: Toggle debug console
- **1/2**: Scene switching (debug)
- **ESC**: Quit game

### Gamepad
- **D-Pad / Left Stick**: Movement and menu navigation
- **A Button (Bottom)**: Confirm / Select
- **B Button (Right)**: Cancel / Back
- **X Button (Left)**: Scene 1
- **Y Button (Top)**: Scene 2
- **Start**: Toggle debug console
- **Select/Back**: Open player menu

## Game Saves

### Desktop (Linux)
Game saves are stored in `~/.gowasm-game-saves/` as JSON files.

### WASM (Browser)
Game saves use browser localStorage.

## Configuration

Edit `pkg/config/settings.go` to change:

- **Screen Resolution**: `Screen.Width` / `Screen.Height` (virtual resolution: 800x600)
- **Window Size**: Automatically calculated as `Resolution × PixelScale`
- **Pixel Scale**: `Rendering.PixelScale` (default: 3 = 2400x1800 window)
- **Pixel Art Mode**: `Rendering.PixelArtMode` (enables nearest-neighbor filtering)

## Architecture

### Dual Backend Support

The engine now supports both **WebGPU (WASM)** and **Ebiten (Desktop)** backends through build tags:

| Component | WebGPU (WASM) | Ebiten (Desktop) |
|-----------|---------------|------------------|
| Canvas Manager | `canvas_webgpu.go` | `canvas_ebiten.go` |
| Input | `unified_input.go` (JS) | `ebiten_input.go` |
| Engine | `engine.go` (JS) | `engine_ebiten.go` |
| Font Loading | `font.go` (fetch API) | `font_desktop.go` (os.ReadFile) |
| Storage | `storage.go` (localStorage) | `storage_desktop.go` (file system) |

### Build Tags

- **`//go:build js`**: WASM/WebGPU specific code
- **`//go:build !js`**: Desktop/Ebiten specific code
- **No tag**: Platform-agnostic code

## Troubleshooting

### "X11/Xlib.h: No such file or directory"

**Solution (NixOS)**:
```bash
nix-shell examples/ebiten-demo/shell.nix
make build-desktop
```

**Solution (Ubuntu/Debian)**:
```bash
sudo apt-get install build-essential libgl1-mesa-dev xorg-dev
make build-desktop
```

### "libGL.so: cannot open shared object file"

**Solution (NixOS)**:
```bash
# Run the game inside nix-shell
nix-shell examples/ebiten-demo/shell.nix
./build/game-desktop
```

The `shell.nix` sets up `LD_LIBRARY_PATH` correctly for OpenGL.

### Game window is blank

Check that assets are in the correct location. From the project root, assets should be at:
- `examples/basic-game/assets/llama.png`
- `examples/basic-game/assets/fonts/Mono_10.sheet.png`
- `examples/basic-game/assets/fonts/Mono_10.sheet.json`

## Development

### Project Structure

```
gowasm-engine/
├── cmd/
│   └── ebiten-game/           # Desktop entry point
├── pkg/
│   ├── canvas/
│   │   ├── canvas_ebiten.go   # Ebiten rendering
│   │   └── canvas_webgpu.go   # WebGPU rendering (WASM)
│   ├── engine/
│   │   ├── engine_ebiten.go   # Desktop engine
│   │   └── engine.go          # WASM engine
│   ├── input/
│   │   ├── ebiten_input.go    # Desktop input
│   │   └── unified_input.go   # WASM input
│   └── ...
├── examples/
│   ├── basic-game/            # Game implementation
│   └── ebiten-demo/           # Minimal Ebiten test
├── Makefile
└── README_EBITEN.md           # This file
```

### Adding New Scenes

1. Create scene in `examples/basic-game/scenes/`
2. Implement `scene.Scene` interface
3. Embed `pkscene.BaseScene` for common functionality
4. Register scene in `cmd/ebiten-game/main.go`

### Testing

```bash
# Run all tests
make test

# Run tests for specific package
go test ./pkg/canvas/...
```

## Performance

- **Target**: 60 FPS (Ebiten runs at 60 TPS by default)
- **Rendering**: Automatic batching of sprites with same texture
- **Pixel-Perfect**: Nearest-neighbor filtering with integer scaling

## Next Steps

See `EBITEN_MIGRATION_PLAN.md` for:
- Phase 7: Config & Polish (fullscreen, window resizing)
- Multi-platform support (Windows, macOS, mobile)
- Re-enabling WASM with Ebiten backend
- Audio system integration

## License

(Same as main project)
