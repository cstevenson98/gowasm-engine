# Ebiten Migration Plan

> **STATUS (2026-07-18): COMPLETE — WebGPU/WASM backend fully removed.**
> The engine is now pure Ebiten. All `//go:build js` files, the WebGPU canvas,
> the WASM entry point, pipeline/batching abstractions, and the `cogentcore/webgpu`
> dependency have been deleted. `EbitenEngine`/`EbitenCanvasManager`/`EbitenInput`
> were renamed to `Engine`/`Canvas`/`Input`. The sections below are retained for
> historical context on how the migration was carried out.

## Executive Summary

This document outlines the migration strategy from the WebGPU WASM engine to Ebiten for Linux desktop. The migration is designed to be incremental, preserving the existing component-based architecture while replacing the rendering backend.

**Timeline Estimate**: 3-5 phases with iterative testing
**Risk Level**: Medium - Core architecture stays intact, rendering layer is abstracted
**Target Platform**: Linux Desktop (future: Windows, macOS, Web via WASM)

---

## Table of Contents

1. [Architecture Mapping](#architecture-mapping)
2. [Component-by-Component Analysis](#component-by-component-analysis)
3. [Migration Phases](#migration-phases)
4. [Technical Decisions](#technical-decisions)
5. [Testing Strategy](#testing-strategy)
6. [Build System Changes](#build-system-changes)
7. [Risk Mitigation](#risk-mitigation)

---

## Architecture Mapping

### Current WebGPU Architecture → Ebiten Architecture

| Current Component | WebGPU Role | Ebiten Equivalent | Migration Action |
|------------------|-------------|-------------------|------------------|
| `Engine` | Game loop orchestrator | `ebiten.Game` interface | **Refactor**: Implement `ebiten.Game` |
| `CanvasManager` (WebGPU) | WebGPU rendering, textures, pipelines | `*ebiten.Image` operations | **Replace**: New `EbitenCanvasManager` |
| `GameObject` | Game entities with Update/Render | Same pattern | **Keep**: No changes needed |
| `Sprite` | Sprite rendering, animation | `ebiten.DrawImageOptions` | **Adapt**: Use Ebiten draw calls |
| `Mover` | Movement physics | Same | **Keep**: No changes needed |
| `Scene` | Scene management | Same | **Keep**: Minor canvas API changes |
| `Input` (Unified) | Keyboard + Gamepad via JS | `ebiten.IsKeyPressed`, `ebiten.StandardGamepadButton` | **Replace**: Use Ebiten input APIs |
| `Config` | Settings management | Same | **Keep**: Update rendering settings |
| `Debug Console` | Debug overlay | Same pattern | **Adapt**: Use Ebiten text rendering |
| `Text/Font` | Sprite font rendering | `ebiten.Image` + custom rendering | **Keep**: Works with Ebiten |

### Key Architectural Changes

1. **Rendering Abstraction Preserved**
   - `CanvasManager` interface stays intact
   - WebGPU implementation → Ebiten implementation
   - No changes to GameObject/Scene layer

2. **Build Tags Simplified**
   - Remove `//go:build js` from most files
   - Only platform-specific code needs tags
   - Single binary for Linux desktop

3. **Entry Point Refactored**
   - `cmd/game/main.go` implements `ebiten.Game`
   - No more `syscall/js` or DOM events
   - Game loop managed by Ebiten

---

## Component-by-Component Analysis

### 1. Engine (`pkg/engine/engine.go`)

**Current State**: Manages game loop via `requestAnimationFrame`, scene transitions, input polling

**Changes Required**:
- Remove `//go:build js` tag
- Remove `syscall/js` imports
- Implement `ebiten.Game` interface:
  - `Update() error` - existing game loop logic
  - `Draw(screen *ebiten.Image)` - call scene rendering
  - `Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int)` - return virtual resolution

**Migration Strategy**:
```go
type Engine struct {
    // Keep existing fields
    canvasManager canvas.CanvasManager
    currentScene  scene.Scene
    // ... etc
    
    // NEW: Ebiten-specific
    ebitenScreen *ebiten.Image // Passed to Draw()
}

// Implement ebiten.Game
func (e *Engine) Update() error {
    // Existing update logic
    deltaTime := 1.0 / 60.0 // Ebiten runs at fixed 60 TPS
    e.currentScene.Update(deltaTime)
    return nil
}

func (e *Engine) Draw(screen *ebiten.Image) {
    e.ebitenScreen = screen
    e.currentScene.Render(e.canvasManager)
}

func (e *Engine) Layout(outsideWidth, outsideHeight int) (int, int) {
    return int(e.screenWidth), int(e.screenHeight)
}
```

**Risk**: Low - Clean separation of concerns makes this straightforward

---

### 2. CanvasManager (`pkg/canvas/`)

**Current State**: WebGPU-specific rendering with pipelines, buffers, textures

**Changes Required**:
- Create `pkg/canvas/canvas_ebiten.go` (NO build tag)
- Implement `CanvasManager` interface using Ebiten primitives
- Remove WebGPU-specific code (pipelines, buffers, bind groups)

**New Implementation**:
```go
package canvas

import "github.com/hajimehoshi/ebiten/v2"

type EbitenCanvasManager struct {
    screen  *ebiten.Image                    // Render target
    textures map[string]*ebiten.Image        // Loaded textures
    batch   []*drawCall                      // Batch rendering (optional)
}

type drawCall struct {
    texture  *ebiten.Image
    options  *ebiten.DrawImageOptions
}

func (e *EbitenCanvasManager) Initialize(canvasID string) error {
    // No-op for Ebiten (screen is passed via Draw)
    e.textures = make(map[string]*ebiten.Image)
    return nil
}

func (e *EbitenCanvasManager) LoadTexture(path string) error {
    img, _, err := ebitenutil.NewImageFromFile(path)
    if err != nil {
        return err
    }
    e.textures[path] = img
    return nil
}

func (e *EbitenCanvasManager) DrawTexture(texture types.Texture, position types.Vector2, size types.Vector2, uv types.UVRect) error {
    img := e.textures[texture.Path]
    opts := &ebiten.DrawImageOptions{}
    
    // Apply UV coordinates (subimage)
    subImg := img.SubImage(image.Rect(
        int(uv.U * float64(img.Bounds().Dx())),
        int(uv.V * float64(img.Bounds().Dy())),
        int((uv.U + uv.Width) * float64(img.Bounds().Dx())),
        int((uv.V + uv.Height) * float64(img.Bounds().Dy())),
    )).(*ebiten.Image)
    
    // Apply position
    opts.GeoM.Translate(position.X, position.Y)
    
    // Pixel-perfect filtering
    opts.Filter = ebiten.FilterNearest
    
    e.screen.DrawImage(subImg, opts)
    return nil
}

// ... implement other CanvasManager methods
```

**Key Points**:
- Ebiten manages its own rendering backend (OpenGL/Metal/DirectX)
- No manual pipeline or buffer management
- Textures are `*ebiten.Image` objects
- Drawing is immediate (or batched via `drawCall` slice)

**Risk**: Low - Interface abstraction makes this a clean swap

---

### 3. Input (`pkg/input/`)

**Current State**: `unified_input.go` with `syscall/js` for keyboard and gamepad polling

**Changes Required**:
- Create `pkg/input/ebiten_input.go` (NO build tag)
- Replace JS event listeners with Ebiten input APIs
- Keep `types.InputState` struct unchanged

**New Implementation**:
```go
package input

import (
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/cstevenson98/gowasm-engine/pkg/types"
)

type EbitenInput struct {
    state types.InputState
}

func NewEbitenInput() *EbitenInput {
    return &EbitenInput{
        state: types.InputState{},
    }
}

func (e *EbitenInput) Initialize() error {
    return nil // No setup needed
}

func (e *EbitenInput) PollInput() types.InputState {
    // Keyboard
    e.state.Up = ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW)
    e.state.Down = ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS)
    e.state.Left = ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA)
    e.state.Right = ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD)
    e.state.Enter = inpututil.IsKeyJustPressed(ebiten.KeyEnter)
    e.state.Escape = inpututil.IsKeyJustPressed(ebiten.KeyEscape)
    
    // Gamepad (first connected controller)
    gamepadIDs := ebiten.AppendGamepadIDs(nil)
    if len(gamepadIDs) > 0 {
        gid := gamepadIDs[0]
        e.state.Up = e.state.Up || ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonLeftTop)
        e.state.Down = e.state.Down || ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonLeftBottom)
        e.state.Left = e.state.Left || ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonLeftLeft)
        e.state.Right = e.state.Right || ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonLeftRight)
        e.state.Enter = e.state.Enter || inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonRightBottom)
    }
    
    return e.state
}
```

**Risk**: Low - Simple 1:1 API mapping

---

### 4. GameObject / Sprite / Mover (`pkg/gameobject/`, `pkg/sprite/`, `pkg/mover/`)

**Current State**: Component-based entities with Update/Render methods

**Changes Required**: **NONE** ✅

These components are already abstracted from rendering details. They call `CanvasManager` methods which we're replacing at the implementation level.

**Example (no changes needed)**:
```go
// pkg/gameobject/player.go
func (p *Player) Update(deltaTime float64) {
    p.mover.Update(deltaTime) // Physics - unchanged
    p.sprite.Update(deltaTime) // Animation - unchanged
}

func (p *Player) Render(canvas canvas.CanvasManager) error {
    return p.sprite.Render(canvas) // Uses CanvasManager interface
}
```

**Risk**: None - Already properly abstracted

---

### 5. Scene (`pkg/scene/`)

**Current State**: Scene management with layers, lifecycle, asset loading

**Changes Required**: Minimal
- Remove `//go:build js` tag from `base_scene.go`
- Update asset loading to use filesystem paths (no fetch API)
- Keep scene lifecycle intact

**Risk**: Low - Minor path changes only

---

### 6. Config (`pkg/config/settings.go`)

**Current State**: Global settings for screen, rendering, player, etc.

**Changes Required**:
- Update `RenderingSettings` to clarify Ebiten behavior
- Remove WebGPU-specific settings (pipelines)
- Add Ebiten-specific options (TPS, fullscreen, etc.)

**Updated Settings**:
```go
type RenderingSettings struct {
    PixelArtMode        bool    // Enable pixel-perfect rendering (FilterNearest)
    PixelScale          int     // Window scale factor (3 = 3x)
    VSyncEnabled        bool    // Enable VSync (default in Ebiten)
    TPS                 int     // Ticks per second (default 60)
    UILineSpacing       float64
    TextLineSpacing     float64
}

type ScreenSettings struct {
    Width         float64 // Virtual resolution (800)
    Height        float64 // Virtual resolution (600)
    WindowWidth   int     // Actual window size (2400)
    WindowHeight  int     // Actual window size (1800)
    Fullscreen    bool    // Fullscreen mode
    Resizable     bool    // Allow window resizing
}
```

**Risk**: Low - Configuration changes

---

### 7. Debug Console (`pkg/debug/console.go`)

**Current State**: Overlay debug console with sprite font rendering

**Changes Required**: Minor
- Console logic stays the same
- Text rendering uses `CanvasManager` (which now uses Ebiten)
- No direct changes needed

**Risk**: None - Relies on CanvasManager abstraction

---

### 8. Text/Font (`pkg/text/`)

**Current State**: Sprite font rendering with `.sheet.json` metadata

**Changes Required**: Minimal
- Font loading uses Ebiten's `ebitenutil.NewImageFromFile`
- Rendering calls `CanvasManager.DrawTexture` (which uses Ebiten)
- Keep existing sprite font system

**Risk**: Low - Abstracted through CanvasManager

---

## Migration Phases

### Phase 1: Foundation Setup ✅ (Partially Complete)

**Goal**: Set up Ebiten project structure and verify build

**Tasks**:
- [x] Create `examples/ebiten-demo/` with minimal example
- [x] Add `github.com/hajimehoshi/ebiten/v2` dependency
- [x] Verify build on Linux desktop
- [x] Create `shell.nix` for NixOS development
- [ ] Document Ebiten build process in README

**Deliverable**: Working Ebiten "hello world" with llama.png

---

### Phase 2: Canvas Manager Migration

**Goal**: Replace WebGPU CanvasManager with Ebiten implementation

**Tasks**:
1. Create `pkg/canvas/canvas_ebiten.go`
2. Implement `CanvasManager` interface methods:
   - `Initialize()`
   - `LoadTexture(path string)`
   - `DrawTexture()`, `DrawTextureRotated()`, `DrawTextureScaled()`
   - `DrawColoredRect()` (for debug console)
   - `ClearCanvas()`
   - Batch rendering (optional optimization)
3. Create `pkg/canvas/mock_canvas_ebiten_test.go` for testing
4. Test texture loading and rendering independently

**Testing**:
- Unit tests for texture loading
- Render a single sprite
- Render multiple sprites with different transforms
- Verify pixel-perfect scaling (FilterNearest)

**Deliverable**: `EbitenCanvasManager` fully functional

---

### Phase 3: Input System Migration

**Goal**: Replace JS input with Ebiten input APIs

**Tasks**:
1. Create `pkg/input/ebiten_input.go`
2. Implement `InputCapturer` interface:
   - `Initialize()`
   - `PollInput() types.InputState`
3. Map keyboard keys (arrows, WASD, Enter, Escape)
4. Map gamepad buttons (D-pad, face buttons)
5. Test input in isolation (simple test program)

**Testing**:
- Verify all keyboard inputs work
- Verify gamepad detection and input
- Test input state merging (keyboard + gamepad)

**Deliverable**: `EbitenInput` fully functional

---

### Phase 4: Engine Refactor

**Goal**: Refactor Engine to implement `ebiten.Game` interface

**Tasks**:
1. Create `pkg/engine/engine_ebiten.go` (or refactor existing)
2. Remove `//go:build js` tag
3. Remove `syscall/js` imports
4. Implement `ebiten.Game`:
   - `Update() error` - move game loop logic here
   - `Draw(screen *ebiten.Image)` - render current scene
   - `Layout()` - return virtual resolution
5. Update scene rendering to pass Ebiten screen
6. Remove `requestAnimationFrame` logic
7. Update delta time calculation (Ebiten runs at fixed TPS)

**Testing**:
- Engine initializes without errors
- Game loop runs at 60 TPS
- Scenes can be loaded and switched
- Delta time is correct

**Deliverable**: Engine working with Ebiten game loop

---

### Phase 5: Entry Point Migration

**Goal**: Create new entry point for Linux desktop

**Tasks**:
1. Create `cmd/ebiten-game/main.go`
2. Initialize Engine with Ebiten canvas and input
3. Set up Ebiten window:
   ```go
   ebiten.SetWindowSize(config.Global.Screen.WindowWidth, config.Global.Screen.WindowHeight)
   ebiten.SetWindowTitle("Game Title")
   ebiten.SetScreenFilterEnabled(false) // Pixel-perfect
   ```
4. Register scenes (Menu, Gameplay, Battle)
5. Call `ebiten.RunGame(engine)`
6. Remove WASM-specific entry point logic

**Testing**:
- Game launches in window
- Menu scene displays correctly
- Input works (keyboard + gamepad)
- Can navigate to gameplay scene

**Deliverable**: Playable game on Linux desktop

---

### Phase 6: Asset Loading & Scenes

**Goal**: Ensure all scenes work with Ebiten

**Tasks**:
1. Update asset loading paths (no fetch API)
2. Test Menu scene rendering
3. Test Gameplay scene with player movement
4. Test Battle scene with animations
5. Verify font rendering in debug console
6. Test scene transitions

**Testing**:
- All textures load correctly
- Sprite animations play
- Text renders correctly
- Debug console works
- Scene transitions are smooth

**Deliverable**: All game scenes functional

---

### Phase 7: Config & Polish

**Goal**: Finalize configuration and add Ebiten-specific features

**Tasks**:
1. Update `pkg/config/settings.go` with Ebiten settings
2. Add fullscreen support
3. Add window resizing logic (maintain aspect ratio)
4. Implement pause/unpause (Escape key)
5. Add FPS/TPS display (optional debug info)
6. Optimize batch rendering (if needed)

**Testing**:
- Test different screen resolutions
- Test fullscreen mode
- Test pixel-perfect scaling at different window sizes
- Verify performance (60 FPS stable)

**Deliverable**: Polished, configurable game

---

### Phase 8: Build System & Documentation

**Goal**: Update build system and documentation

**Tasks**:
1. Update `Makefile`:
   ```makefile
   .PHONY: build-desktop
   build-desktop:
       go build -o build/game ./cmd/ebiten-game
   
   .PHONY: run-desktop
   run-desktop: build-desktop
       ./build/game
   ```
2. Update `README.md` with Ebiten instructions
3. Document NixOS setup (`shell.nix`)
4. Add example configs for different resolutions
5. Create troubleshooting guide

**Deliverable**: Complete documentation and build scripts

---

## Technical Decisions

### 1. Pixel-Perfect Scaling

**Current Approach (WebGPU)**:
- Manual NDC coordinate calculations
- `PixelScale` applied in rendering pipeline
- Canvas size = Virtual resolution × PixelScale

**Ebiten Approach**:
- Set virtual resolution via `Layout()` (800x600)
- Set window size via `SetWindowSize()` (2400x1800 = 3x)
- Ebiten auto-scales with `FilterNearest`
- Use `SetScreenFilterEnabled(false)` for pixel-perfect rendering

**Decision**: Use Ebiten's built-in scaling (simpler, equivalent quality)

---

### 2. Batch Rendering

**Current Approach**: Manual batching via `BeginBatch()` / `EndBatch()`

**Ebiten Options**:
1. **Immediate Mode** (simplest): Call `DrawImage()` for each sprite
2. **Manual Batching**: Collect draw calls, sort by texture, batch render
3. **Ebiten's Auto-Batching**: Ebiten batches same-texture draws automatically

**Decision**: Start with immediate mode, profile, optimize if needed

---

### 3. Text Rendering

**Current Approach**: Custom sprite font system with `.sheet.json` metadata

**Ebiten Options**:
1. Keep custom sprite font system (works with any renderer)
2. Use `golang.org/x/image/font` with TTF fonts
3. Use `github.com/hajimehoshi/ebiten/v2/text` package

**Decision**: Keep custom sprite font system (already works, consistent with game style)

---

### 4. Build Tags

**Current Approach**: `//go:build js` for WASM code

**New Approach**:
- Remove build tags from most files (desktop-only for now)
- Future: Add tags if we support WASM later (Ebiten supports WASM)
- Use `canvas_webgpu.go` (js tag) vs `canvas_ebiten.go` (no tag) for dual support

**Decision**: Remove tags initially, add back if we re-enable WASM build

---

### 5. Asset Loading

**Current Approach**: Fetch API via `syscall/js`

**Ebiten Approach**:
- Use `os.ReadFile()` or `ebitenutil.NewImageFromFile()`
- Relative paths from binary location
- Consider embedding assets with `//go:embed` for distribution

**Decision**: Use filesystem loading, consider `go:embed` later

---

## Testing Strategy

### Unit Tests

**Unchanged Components** (already have tests):
- `pkg/gameobject/*_test.go`
- `pkg/mover/*_test.go`
- `pkg/sprite/*_test.go`

**New Tests Needed**:
- `pkg/canvas/canvas_ebiten_test.go` - Test EbitenCanvasManager
- `pkg/input/ebiten_input_test.go` - Test EbitenInput
- `pkg/engine/engine_ebiten_test.go` - Test Engine.Update/Draw/Layout

### Integration Tests

1. **Rendering Test**: Load texture, render sprite, verify output
2. **Input Test**: Simulate keyboard/gamepad, verify state
3. **Scene Test**: Load scene, update, render, verify game objects
4. **Full Game Test**: Run through menu → gameplay → battle

### Manual Testing Checklist

- [ ] Game launches without errors
- [ ] Menu scene displays correctly
- [ ] Can select menu options with keyboard
- [ ] Can select menu options with gamepad
- [ ] Gameplay scene loads
- [ ] Player moves with arrow keys
- [ ] Player moves with gamepad D-pad
- [ ] Sprite animations play correctly
- [ ] Screen wrapping works
- [ ] Battle scene loads
- [ ] Battle UI renders correctly
- [ ] Debug console toggles (F1/backtick)
- [ ] Debug console text is readable
- [ ] Pixel-perfect scaling maintains crisp pixels
- [ ] Window resizing maintains aspect ratio
- [ ] Fullscreen mode works
- [ ] Game runs at stable 60 FPS
- [ ] No memory leaks (run for extended period)

---

## Build System Changes

### Makefile Updates

```makefile
# Existing WASM build (keep for future)
.PHONY: build-wasm
build-wasm:
	GOOS=js GOARCH=wasm go build -o build/main.wasm ./cmd/game

# NEW: Desktop build
.PHONY: build-desktop
build-desktop:
	go build -o build/game-desktop ./cmd/ebiten-game

# NEW: Run desktop game
.PHONY: run-desktop
run-desktop: build-desktop
	./build/game-desktop

# NEW: Build with race detector (debugging)
.PHONY: build-desktop-race
build-desktop-race:
	go build -race -o build/game-desktop-race ./cmd/ebiten-game

# Run tests (unchanged)
.PHONY: test
test:
	go test ./...

# NEW: Run with verbose output
.PHONY: run-debug
run-debug: build-desktop
	EBITEN_DEBUG=1 ./build/game-desktop

# NEW: Profile performance
.PHONY: profile
profile: build-desktop
	go build -o build/game-desktop-profile -cpuprofile=cpu.prof ./cmd/ebiten-game
	./build/game-desktop-profile
	go tool pprof cpu.prof
```

### go.mod Changes

```go
module github.com/cstevenson98/gowasm-engine

go 1.21

require (
    // NEW: Ebiten dependency
    github.com/hajimehoshi/ebiten/v2 v2.6.3
    
    // Keep existing
    github.com/google/uuid v1.6.0
    
    // REMOVE: WebGPU (keep if we want dual backends)
    // github.com/cogentcore/webgpu v0.23.0
)
```

### NixOS Development (shell.nix)

Already created in Phase 1 ✅

```nix
{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  name = "gowasm-engine-ebiten";
  
  buildInputs = with pkgs; [
    go
    gcc
    pkg-config
    
    # X11 libraries
    xorg.libX11
    xorg.libXcursor
    xorg.libXrandr
    xorg.libXinerama
    xorg.libXi
    xorg.libXxf86vm
    
    # OpenGL
    libGL
    libglvnd
  ];
  
  LD_LIBRARY_PATH = with pkgs; lib.makeLibraryPath [
    libGL
    libglvnd
    xorg.libX11
    xorg.libXcursor
    xorg.libXrandr
    xorg.libXinerama
    xorg.libXi
    xorg.libXxf86vm
  ] + ":/run/opengl-driver/lib";
  
  shellHook = ''
    echo "Ebiten development environment loaded"
    echo "Run: make build-desktop && make run-desktop"
  '';
}
```

---

## Risk Mitigation

### High-Risk Areas

1. **Rendering Pipeline Replacement**
   - **Risk**: Visual glitches, incorrect scaling, missing features
   - **Mitigation**: Implement CanvasManager interface faithfully, test each method in isolation
   - **Fallback**: Keep WebGPU implementation in separate file with build tags

2. **Input System**
   - **Risk**: Missing input events, gamepad detection issues
   - **Mitigation**: Test on real hardware with multiple controllers
   - **Fallback**: Keyboard-only initially, add gamepad later

3. **Performance**
   - **Risk**: Frame rate drops, stuttering
   - **Mitigation**: Profile early, implement batch rendering if needed
   - **Fallback**: Reduce sprite count, simplify effects

### Medium-Risk Areas

1. **Asset Loading**
   - **Risk**: Path issues, missing textures
   - **Mitigation**: Use relative paths, add error handling, consider `go:embed`
   - **Fallback**: Hardcode asset paths temporarily

2. **Font Rendering**
   - **Risk**: Debug console text garbled or misaligned
   - **Mitigation**: Test sprite font system independently with Ebiten
   - **Fallback**: Use Ebiten's built-in text package temporarily

### Low-Risk Areas

1. **Game Logic** (GameObject, Mover, Scene)
   - Already abstracted, no changes needed
   - Existing tests cover functionality

2. **Configuration**
   - Simple data structures, easy to update

---

## Post-Migration: Future Enhancements

### Phase 9+ (Optional)

1. **Multi-Platform Support**
   - Windows build
   - macOS build
   - Mobile (Android/iOS via Ebiten)

2. **WASM Re-enablement**
   - Ebiten supports WASM compilation
   - Dual backend: `canvas_ebiten.go` (desktop + WASM) vs `canvas_webgpu.go` (WASM only)

3. **Audio System**
   - Add Ebiten audio (`github.com/hajimehoshi/ebiten/v2/audio`)
   - Music and sound effects

4. **Advanced Features**
   - Particle systems
   - Post-processing effects (Ebiten shaders - Kage)
   - Save/load system
   - Networking (multiplayer)

5. **Packaging**
   - Build scripts for distribution (.deb, .rpm, AppImage)
   - NixOS package (`default.nix`)
   - Flatpak/Snap packaging

---

## Success Criteria

### Phase Completion Criteria

Each phase is considered complete when:
- All tasks are implemented
- Tests pass
- Manual testing checklist items verified
- No regressions in existing functionality
- Code reviewed and documented

### Final Migration Success

The migration is successful when:
- [x] Game builds on Linux desktop without WASM dependencies
- [ ] All scenes (Menu, Gameplay, Battle) render correctly
- [ ] Input works with keyboard and gamepad
- [ ] Pixel-perfect scaling is maintained
- [ ] Debug console functions properly
- [ ] Performance is stable (60 FPS)
- [ ] All existing tests pass
- [ ] New Ebiten-specific tests added
- [ ] Documentation updated
- [ ] Build system simplified

---

## Appendix: File Structure After Migration

```
gowasm-engine/
├── cmd/
│   ├── game/                  # WASM entry point (keep for future)
│   │   └── main.go            # (//go:build js)
│   └── ebiten-game/           # NEW: Desktop entry point
│       └── main.go            # Ebiten initialization
├── pkg/
│   ├── canvas/
│   │   ├── interface.go       # CanvasManager interface (unchanged)
│   │   ├── canvas_webgpu.go   # WebGPU impl (keep, add //go:build js)
│   │   └── canvas_ebiten.go   # NEW: Ebiten impl (no build tag)
│   ├── input/
│   │   ├── interface.go       # InputCapturer interface (unchanged)
│   │   ├── unified_input.go   # JS input (keep, add //go:build js)
│   │   └── ebiten_input.go    # NEW: Ebiten input (no build tag)
│   ├── engine/
│   │   └── engine.go          # Refactored (remove js tag, implement ebiten.Game)
│   ├── gameobject/            # UNCHANGED
│   ├── sprite/                # UNCHANGED
│   ├── mover/                 # UNCHANGED
│   ├── scene/                 # Minor changes (remove js tags)
│   ├── config/                # Updated settings
│   ├── debug/                 # UNCHANGED
│   ├── text/                  # UNCHANGED
│   ├── types/                 # UNCHANGED
│   └── logger/                # UNCHANGED
├── examples/
│   ├── basic-game/            # WASM example (keep)
│   └── ebiten-demo/           # NEW: Minimal Ebiten test ✅
├── shell.nix                  # NEW: NixOS development environment ✅
├── Makefile                   # Updated with desktop targets
├── go.mod                     # Add Ebiten dependency
├── README.md                  # Updated with desktop instructions
├── EBITEN_MIGRATION_PLAN.md   # This document
└── CURSOR_HISTORY.md          # Updated with migration log
```

---

## Conclusion

This migration plan provides a structured, low-risk path from WebGPU to Ebiten. The existing architecture's clean separation of concerns makes this migration feasible without a complete rewrite.

**Key Advantages**:
- Component-based architecture is preserved
- Game logic remains unchanged
- Interface abstractions enable clean swap of rendering backend
- Incremental approach allows testing at each phase
- Can keep WASM build for future if desired

**Next Steps**:
1. Review and approve this plan
2. Begin Phase 2 (Canvas Manager Migration)
3. Test each phase thoroughly before proceeding
4. Document lessons learned in CURSOR_HISTORY.md

**Estimated Effort**: 3-5 weeks of focused development (assuming 1-2 phases per week)

---

*Document Version*: 1.0
*Last Updated*: 2026-07-16
*Author*: Cursor AI Assistant
*Status*: Ready for Review
