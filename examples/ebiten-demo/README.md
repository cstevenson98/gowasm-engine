# Ebiten Llama Demo

Minimal Ebiten example that renders llama.png in an 800x600 window.

## Requirements

- Go 1.24+
- Linux desktop (X11 or Wayland)

## Build and Run

```bash
# Install dependencies
go mod download

# Run
go run main.go
```

## Controls

- **Arrow Keys** - Move the llama
- **ESC** - Quit

## Features Demonstrated

- ✅ Load PNG image
- ✅ 800x600 window
- ✅ Pixel-perfect rendering (nearest-neighbor filtering)
- ✅ Keyboard input
- ✅ Game loop (60 FPS update, variable render)
- ✅ Window resizing support

## Code Size

**Total: 70 lines of Go code**

Compare to WebGPU implementation:
- No shaders required
- No manual batching
- No WebGPU pipeline setup
- No syscall/js bridge

## Next Steps

To add more features:

```go
// Pixel-perfect scaling (3x)
ebiten.SetWindowSize(2400, 1800)  // 800*3, 600*3

// Fullscreen
ebiten.SetFullscreen(true)

// Controller support
ids := ebiten.AppendGamepadIDs(nil)
if len(ids) > 0 {
    x := ebiten.GamepadAxisValue(ids[0], ebiten.GamepadAxis0)
}

// Audio
audioCtx := audio.NewContext(48000)
player, _ := audioCtx.NewPlayer(soundStream)
player.Play()
```
