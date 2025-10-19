# Game Engine Architecture Documentation

This document provides comprehensive architecture diagrams for the Go WASM WebGPU Game Engine.

## Viewing the Diagrams

The diagrams are written in PlantUML format. You can view them using:

### Option 1: PlantUML Online Server
1. Go to http://www.plantuml.com/plantuml/uml/
2. Copy the contents of the `.puml` files
3. Paste and view

### Option 2: VS Code Extension
1. Install the "PlantUML" extension in VS Code
2. Open the `.puml` files
3. Press `Alt+D` to preview

### Option 3: Command Line
```bash
# Install PlantUML
sudo apt-get install plantuml

# Generate PNG images
plantuml docs/architecture-class-diagram.puml
plantuml docs/rendering-sequence-diagram.puml
```

## Architecture Overview

### Class Structure Diagram
**File**: `architecture-class-diagram.puml`

This diagram shows the complete class/interface structure of the game engine, including:

#### Core Packages:
1. **Engine** - Game loop orchestration and state management
2. **Scene** - Scene management with layered rendering (BACKGROUND, ENTITIES, UI)
3. **GameObject** - Game entities (Player, Background, Llama)
4. **Components** - Sprite and Mover interfaces with implementations
5. **Canvas** - WebGPU rendering pipeline and batch system
6. **Input** - Unified keyboard and gamepad input
7. **Text** - Font sprite sheets and text rendering
8. **Debug** - Thread-safe debug console with message queue

#### Key Design Patterns:
- **Component-Based Architecture**: GameObjects compose Sprite and Mover components
- **Interface-Driven Design**: All major systems define interfaces for mockability
- **Layered Rendering**: Scenes organize objects into BACKGROUND, ENTITIES, and UI layers
- **Batch Rendering**: Canvas groups draws by texture to minimize GPU state changes
- **Observer Pattern**: Debug console receives messages globally from any GameObject

### Rendering Flow Sequence Diagram
**File**: `rendering-sequence-diagram.puml`

This diagram shows the complete flow of a single frame, from initialization through rendering:

#### Initialization Phase:
1. Browser triggers engine initialization
2. Engine creates WebGPU canvas (device, queue, surface, pipelines)
3. Engine creates scene (GameplayScene)
4. Scene creates GameObjects (Background, Player)
5. GameObjects create Sprite and Mover components
6. Scene initializes debug console with font loading
7. Font metadata loaded asynchronously via fetch API

#### Game Loop (Every Frame):
1. **Update Phase**:
   - Calculate deltaTime
   - Scene updates all GameObjects
   - Player handles input and updates velocity
   - Movers update positions based on velocity
   - Sprites update animations
   - Debug console ages messages
   - Textures loaded asynchronously (non-blocking)

2. **Render Phase**:
   - Scene returns renderables in layer order
   - Canvas enters batch mode
   - For each GameObject:
     - Get sprite render data (texture, position, size, UV)
     - Generate quad vertices (6 vertices × 4 floats)
     - Append to batch for that texture
   - Debug console renders text characters (also batched)
   - End batch (finalize all batches by texture)

3. **GPU Submission**:
   - Get next frame texture from surface
   - Create command encoder
   - For each texture batch:
     - Upload vertices to GPU buffer
     - Set render pipeline
     - Bind texture
     - Draw vertices
   - Submit commands to GPU
   - Present frame to screen

## Architecture Principles

### 1. Build Tags (`//go:build js`)
- WASM-specific code uses `//go:build js` tag
- Mock implementations have no build tag (for native testing)
- Allows fast unit testing without browser

### 2. WebGPU Wrapper
- Uses `cogentcore/webgpu` library to minimize direct JS calls
- Direct JS only for DOM manipulation and browser APIs
- All GPU operations through type-safe Go wrapper

### 3. Component-Based GameObjects
```
GameObject Interface
├─ Sprite Component (rendering)
├─ Mover Component (physics)
└─ ObjectState (data)
```

### 4. Layered Scene System
```
Scene
├─ BACKGROUND Layer → Rendered first
├─ ENTITIES Layer   → Rendered second (player, NPCs)
└─ UI Layer         → Rendered last (debug console)
```

### 5. Batch Rendering Optimization
- All draws grouped by texture
- Minimizes GPU pipeline switches
- Single buffer upload per texture
- Dramatically improves performance

### 6. Async Texture Loading
- Textures load via JavaScript Image API
- Non-blocking game loop
- LoadTexture() called every frame (idempotent)
- Rendering skips unloaded textures gracefully

### 7. Debug Console Architecture
- Global singleton accessible anywhere
- Thread-safe circular message buffer
- Renders as text using font sprite sheet
- Updates and renders in game loop

## Data Flow

### Input → GameObject → Rendering
```
UnifiedInput.GetInputState()
    ↓
Player.HandleInput(InputState)
    ↓
Mover.SetVelocity(velocity)
    ↓
Mover.Update(deltaTime)
    ↓
Mover.GetPosition() → Sprite.GetSpriteRenderData()
    ↓
Canvas.DrawTexturedRect(texture, position, size, UV)
    ↓
GPU renders frame
```

### Debug Message Flow
```
GameObject.Update()
    ↓
types.PostDebugMessage("source", "message")
    ↓
DebugConsole.PostMessage() [thread-safe]
    ↓
DebugConsole.Render() [in game loop]
    ↓
TextRenderer.RenderTextScaled()
    ↓
Font.GetCharacterUV(char) → Canvas.DrawTexturedRect()
```

## Performance Characteristics

### Batch Rendering Benefits:
- **Without batching**: 1 GPU draw call per sprite = ~100 calls/frame
- **With batching**: 1-5 GPU draw calls per frame (grouped by texture)
- **Result**: 20-100x reduction in GPU state changes

### Texture Loading:
- **Async loading**: Game continues running during texture load
- **Idempotent**: Safe to call LoadTexture() every frame
- **Fallback**: Unloaded textures skipped gracefully

### Debug Console:
- **Circular buffer**: Constant memory usage (max 10 messages)
- **Batch rendered**: All text in single batch
- **No allocation**: Reuses message buffer slots

## File Organization

```
internal/
├── canvas/          # WebGPU rendering
│   ├── canvas_webgpu.go      # Implementation (js build tag)
│   ├── interface.go          # CanvasManager interface
│   └── mock_canvas.go        # Mock for testing
├── debug/           # Debug console system
│   ├── console.go            # Thread-safe console (js build tag)
│   └── message.go            # Message structure
├── engine/          # Game loop and state
│   └── engine.go             # Engine implementation (js build tag)
├── gameobject/      # Game entities
│   ├── player.go             # Player implementation
│   ├── background.go         # Background implementation
│   └── llama.go              # Llama NPC
├── input/           # Input handling
│   ├── unified_input.go      # Keyboard + Gamepad
│   ├── keyboard_input.go     # Keyboard handling
│   └── gamepad_input.go      # Gamepad handling
├── mover/           # Physics components
│   ├── basic_mover.go        # Simple mover (js build tag)
│   └── mock_mover.go         # Mock for testing
├── scene/           # Scene management
│   ├── scene.go              # Scene interface
│   └── gameplay_scene.go     # Main gameplay scene
├── sprite/          # Sprite components
│   ├── sprite.go             # SpriteSheet implementation
│   └── mock_sprite.go        # Mock for testing
├── text/            # Text rendering
│   ├── font.go               # Font sprite sheet loader
│   ├── text_renderer.go      # Text rendering implementation
│   ├── interface.go          # Font and TextRenderer interfaces
│   └── mock_text.go          # Mocks for testing
└── types/           # Shared interfaces
    ├── gameobject.go         # GameObject interface
    ├── sprite.go             # Sprite interface
    ├── mover.go              # Mover interface
    └── input.go              # InputState structure
```

## Future Enhancements

### Rendering:
- [ ] Implement colored rectangle rendering in separate pass
- [ ] Add debug console background (without texture artifacts)
- [ ] Texture atlases for better batching
- [ ] Sprite rotation and scaling

### Text System:
- [ ] Text alignment (left, center, right)
- [ ] Word wrapping
- [ ] Multiple font support
- [ ] Rich text formatting

### Debug Console:
- [ ] Console commands system
- [ ] Scrolling and history
- [ ] Toggle visibility (keyboard shortcut)
- [ ] Filter messages by source

### Scene System:
- [ ] Scene transitions
- [ ] Multiple scene types (Menu, Pause, GameOver)
- [ ] Scene loading/unloading
- [ ] Asset preloading per scene

## References

- [WebGPU Specification](https://www.w3.org/TR/webgpu/)
- [Go WASM](https://github.com/golang/go/wiki/WebAssembly)
- [cogentcore/webgpu](https://github.com/cogentcore/webgpu)
- [PlantUML Documentation](https://plantuml.com/)







