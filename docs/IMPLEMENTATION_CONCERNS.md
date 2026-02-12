# Implementation Concerns & Solutions

**Date**: 2026-02-12  
**Purpose**: Address architectural concerns about proposed ease-of-use improvements  
**Companion to**: `EASE_OF_USE_IMPROVEMENTS.md`

---

## Table of Contents

1. [Concern 1: Optional Interface Implementation in BaseScene](#concern-1-optional-interface-implementation-in-basescene)
   - [How Go Method Overriding Works](#how-go-method-overriding-works)
   - [Strategy 1: Full Default Implementation](#strategy-1-full-default-implementation-recommended)
   - [Strategy 2: Compositional Scene Components](#strategy-2-compositional-scene-components)
   - [Strategy 3: Multiple Base Types](#strategy-3-multiple-base-types)
2. [Concern 2: Dependency Injection Pattern](#concern-2-dependency-injection-pattern)
   - [Current Manual Injection Problem](#current-manual-injection-problem)
   - [Proposed Solution Architecture](#proposed-solution-architecture)
   - [Implementation Examples](#implementation-examples)

---

## Concern 1: Optional Interface Implementation in BaseScene

### The Question

> "How does overriding work in Go? If we don't want BaseScene to implement all interfaces, would we have predefined composite scene types?"

### How Go Method Overriding Works

In Go, embedding (anonymous fields) provides **method delegation**, not true inheritance. However, you can "shadow" or "override" methods:

```go
// Base type
type BaseScene struct {
    name string
}

func (b *BaseScene) GetName() string {
    return b.name
}

func (b *BaseScene) DoSomething() {
    fmt.Println("BaseScene implementation")
}

// Derived type
type MyScene struct {
    *BaseScene  // Embedded - delegates methods
}

// Override DoSomething - shadows the base implementation
func (m *MyScene) DoSomething() {
    fmt.Println("MyScene custom implementation")
}

// Usage
scene := &MyScene{BaseScene: &BaseScene{name: "test"}}
scene.GetName()      // Uses BaseScene.GetName (delegated)
scene.DoSomething()  // Uses MyScene.DoSomething (overridden)
```

**Key insight**: Methods on the outer type take precedence. If `MyScene` doesn't define `DoSomething()`, the call delegates to `BaseScene.DoSomething()`.

---

### Strategy 1: Full Default Implementation (Recommended)

**Approach**: BaseScene implements ALL optional interfaces with sensible defaults. User scenes override only what they need.

#### Architecture

```go
// pkg/scene/base_scene.go
package scene

import (
    "github.com/cstevenson98/gowasm-engine/pkg/types"
    "github.com/cstevenson98/gowasm-engine/pkg/canvas"
    "sync"
)

// BaseScene provides default implementations for all common scene interfaces.
// Embed this in your scene to get free implementations, override only what you need.
type BaseScene struct {
    // Core fields
    name          string
    screenWidth   float64
    screenHeight  float64
    
    // Layer management
    layers        map[SceneLayer][]types.GameObject
    layerMutex    sync.RWMutex
    
    // Optional interface fields (all initialized by default)
    inputCapturer       types.InputCapturer
    stateChangeCallback func(types.GameStateType, map[string]interface{}) error
    canvasManager       canvas.CanvasManager
    gameStateManager    interface{}
    
    // Saved state (for SceneStateful)
    savedState    map[string]interface{}
    
    // Text rendering (for SceneAssetProvider)
    requiredAssets types.SceneAssets
}

// Constructor
func NewBaseScene(name string, width, height float64) *BaseScene {
    return &BaseScene{
        name:         name,
        screenWidth:  width,
        screenHeight: height,
        layers:       make(map[SceneLayer][]types.GameObject),
        savedState:   make(map[string]interface{}),
        requiredAssets: types.SceneAssets{
            Textures: []string{},
            Fonts:    []string{},
        },
    }
}

// ===== Core Scene Interface =====

func (b *BaseScene) GetName() string {
    return b.name
}

func (b *BaseScene) Initialize() error {
    // Default: initialize empty layers
    b.layerMutex.Lock()
    defer b.layerMutex.Unlock()
    
    if b.layers == nil {
        b.layers = make(map[SceneLayer][]types.GameObject)
    }
    
    b.layers[BACKGROUND] = []types.GameObject{}
    b.layers[ENTITIES] = []types.GameObject{}
    b.layers[UI] = []types.GameObject{}
    
    return nil
}

func (b *BaseScene) Update(deltaTime float64) error {
    // Default: update all game objects in all layers
    b.layerMutex.RLock()
    defer b.layerMutex.RUnlock()
    
    for _, layer := range []SceneLayer{BACKGROUND, ENTITIES, UI} {
        for _, obj := range b.layers[layer] {
            if obj.IsVisible() {
                obj.Update(deltaTime)
            }
        }
    }
    return nil
}

func (b *BaseScene) GetRenderables() []types.GameObject {
    // Default: return all objects in layer order
    b.layerMutex.RLock()
    defer b.layerMutex.RUnlock()
    
    renderables := []types.GameObject{}
    renderables = append(renderables, b.layers[BACKGROUND]...)
    renderables = append(renderables, b.layers[ENTITIES]...)
    renderables = append(renderables, b.layers[UI]...)
    return renderables
}

func (b *BaseScene) Cleanup() error {
    // Default: clear all layers
    b.layerMutex.Lock()
    defer b.layerMutex.Unlock()
    
    for layer := range b.layers {
        b.layers[layer] = nil
    }
    return nil
}

// ===== SceneInputProvider (optional interface) =====

func (b *BaseScene) SetInputCapturer(inputCapturer types.InputCapturer) {
    b.inputCapturer = inputCapturer
}

func (b *BaseScene) GetInputState() types.InputState {
    if b.inputCapturer != nil {
        return b.inputCapturer.GetInputState()
    }
    return types.InputState{} // Default: no input
}

// ===== SceneStateChangeRequester (optional interface) =====

func (b *BaseScene) SetStateChangeCallback(callback func(types.GameStateType, map[string]interface{}) error) {
    b.stateChangeCallback = callback
}

func (b *BaseScene) RequestStateChange(newState types.GameStateType, data map[string]interface{}) error {
    if b.stateChangeCallback != nil {
        return b.stateChangeCallback(newState, data)
    }
    return nil // Default: no-op if not set
}

// ===== SceneGameStateUser (optional interface) =====

func (b *BaseScene) SetGameState(gameState interface{}) {
    b.gameStateManager = gameState
}

func (b *BaseScene) GetGameState() interface{} {
    return b.gameStateManager
}

// ===== SceneTextureProvider (optional interface) =====

func (b *BaseScene) SetCanvasManager(cm canvas.CanvasManager) {
    b.canvasManager = cm
}

func (b *BaseScene) GetCanvasManager() canvas.CanvasManager {
    return b.canvasManager
}

// ===== SceneAssetProvider (optional interface) =====

func (b *BaseScene) GetRequiredAssets() types.SceneAssets {
    return b.requiredAssets
}

func (b *BaseScene) SetRequiredAssets(assets types.SceneAssets) {
    b.requiredAssets = assets
}

// ===== SceneStateful (optional interface) =====

func (b *BaseScene) SaveState() (map[string]interface{}, error) {
    // Default: return saved state map (scenes can populate this)
    return b.savedState, nil
}

func (b *BaseScene) RestoreState(state map[string]interface{}) error {
    // Default: restore saved state
    b.savedState = state
    return nil
}

// ===== SceneOverlayRenderer (optional interface) =====

func (b *BaseScene) RenderOverlay() error {
    // Default: no overlay rendering
    return nil
}

// ===== Layer Management Helpers =====

func (b *BaseScene) AddBackground(obj types.GameObject) {
    b.layerMutex.Lock()
    defer b.layerMutex.Unlock()
    b.layers[BACKGROUND] = append(b.layers[BACKGROUND], obj)
}

func (b *BaseScene) AddEntity(obj types.GameObject) {
    b.layerMutex.Lock()
    defer b.layerMutex.Unlock()
    b.layers[ENTITIES] = append(b.layers[ENTITIES], obj)
}

func (b *BaseScene) AddUI(obj types.GameObject) {
    b.layerMutex.Lock()
    defer b.layerMutex.Unlock()
    b.layers[UI] = append(b.layers[UI], obj)
}

func (b *BaseScene) RemoveGameObject(id string) {
    b.layerMutex.Lock()
    defer b.layerMutex.Unlock()
    
    for layer := range b.layers {
        filtered := []types.GameObject{}
        for _, obj := range b.layers[layer] {
            if obj.GetID() != id {
                filtered = append(filtered, obj)
            }
        }
        b.layers[layer] = filtered
    }
}
```

#### Example 1: Minimal User Scene (Uses Defaults)

```go
// examples/basic-game/scenes/simple_scene.go
package scenes

import (
    "github.com/cstevenson98/gowasm-engine/pkg/scene"
    "github.com/cstevenson98/gowasm-engine/pkg/gameobject"
    "github.com/cstevenson98/gowasm-engine/pkg/types"
)

// SimpleScene doesn't need ANY boilerplate!
type SimpleScene struct {
    *scene.BaseScene  // Embed - gets ALL interface implementations
    
    // Only game-specific fields
    player *gameobject.Player
}

func NewSimpleScene(width, height float64) *SimpleScene {
    return &SimpleScene{
        BaseScene: scene.NewBaseScene("Simple", width, height),
    }
}

// Only override what you need - Initialize to add game objects
func (s *SimpleScene) Initialize() error {
    // Call base initialization (sets up layers)
    if err := s.BaseScene.Initialize(); err != nil {
        return err
    }
    
    // Add game-specific objects
    s.player = gameobject.NewPlayer(
        types.Vector2{X: 100, Y: 100},
        types.Vector2{X: 32, Y: 32},
        200.0,
    )
    
    s.AddEntity(s.player)  // Helper from BaseScene
    
    background := gameobject.NewBackground(
        types.Vector2{X: 0, Y: 0},
        types.Vector2{X: s.screenWidth, Y: s.screenHeight},
        "background.png",
    )
    s.AddBackground(background)
    
    return nil
}

// Don't need to override Update - BaseScene.Update handles it
// Don't need to override GetRenderables - BaseScene.GetRenderables handles it
// Don't need to implement SetInputCapturer, SetCanvasManager, etc - BaseScene has them!

// RESULT: ~30 lines vs 200+ lines before!
```

#### Example 2: Scene with Custom Update Logic

```go
// examples/basic-game/scenes/gameplay_scene.go
package scenes

import (
    "github.com/cstevenson98/gowasm-engine/pkg/scene"
    "github.com/cstevenson98/gowasm-engine/pkg/types"
)

type GameplayScene struct {
    *scene.BaseScene
    
    player *gameobject.Player
    score  int
}

func NewGameplayScene(width, height float64) *GameplayScene {
    base := scene.NewBaseScene("Gameplay", width, height)
    
    // Configure required assets (uses BaseScene's field)
    base.SetRequiredAssets(types.SceneAssets{
        Textures: []string{"player.png", "background.png"},
        Fonts:    []string{"font.png"},
    })
    
    return &GameplayScene{
        BaseScene: base,
    }
}

func (s *GameplayScene) Initialize() error {
    // Call base
    if err := s.BaseScene.Initialize(); err != nil {
        return err
    }
    
    // Game-specific setup
    s.player = gameobject.NewPlayer(
        types.Vector2{X: 100, Y: 100},
        types.Vector2{X: 32, Y: 32},
        200.0,
    )
    s.AddEntity(s.player)
    
    return nil
}

// Override Update to add custom game logic
func (s *GameplayScene) Update(deltaTime float64) error {
    // Call base update first (updates all game objects)
    if err := s.BaseScene.Update(deltaTime); err != nil {
        return err
    }
    
    // Custom game logic
    input := s.GetInputState()  // Uses BaseScene's method
    if input.Buttons["Escape"] {
        // Use BaseScene's RequestStateChange
        s.RequestStateChange(types.GameStateMenu, nil)
    }
    
    // Check win condition
    if s.score >= 100 {
        s.RequestStateChange(types.GameStateWin, map[string]interface{}{
            "score": s.score,
        })
    }
    
    return nil
}

// Override SaveState to save custom data
func (s *GameplayScene) SaveState() (map[string]interface{}, error) {
    state := map[string]interface{}{
        "score":          s.score,
        "playerPosition": s.player.GetState().Position,
    }
    return state, nil
}

// Override RestoreState to restore custom data
func (s *GameplayScene) RestoreState(state map[string]interface{}) error {
    if score, ok := state["score"].(int); ok {
        s.score = score
    }
    if pos, ok := state["playerPosition"].(types.Vector2); ok {
        playerState := s.player.GetState()
        playerState.Position = pos
        s.player.SetState(*playerState)
    }
    return nil
}

// RESULT: ~70 lines vs 300+ lines before, with MORE functionality!
```

#### Example 3: Scene with Custom Overlay Rendering

```go
// examples/basic-game/scenes/battle_scene.go
package scenes

import (
    "github.com/cstevenson98/gowasm-engine/pkg/scene"
    "github.com/cstevenson98/gowasm-engine/pkg/text"
)

type BattleScene struct {
    *scene.BaseScene
    
    playerHP int
    enemyHP  int
    
    // Custom text renderer for battle UI
    uiTextRenderer *text.TextRenderer
    uiFont         text.Font
}

func NewBattleScene(width, height float64) *BattleScene {
    return &BattleScene{
        BaseScene: scene.NewBaseScene("Battle", width, height),
    }
}

func (s *BattleScene) Initialize() error {
    if err := s.BaseScene.Initialize(); err != nil {
        return err
    }
    
    // Initialize custom text rendering
    s.uiFont = text.NewSpriteFont()
    s.uiFont.(*text.SpriteFont).LoadFont("font.png")
    s.uiTextRenderer = text.NewTextRenderer(s.GetCanvasManager())
    
    s.playerHP = 100
    s.enemyHP = 50
    
    return nil
}

// Override RenderOverlay for custom UI
func (s *BattleScene) RenderOverlay() error {
    // Render HP bars on top of everything
    s.uiTextRenderer.RenderText(
        s.uiFont,
        fmt.Sprintf("Player HP: %d", s.playerHP),
        types.Vector2{X: 10, Y: 10},
        types.Color{1, 1, 1, 1},
    )
    
    s.uiTextRenderer.RenderText(
        s.uiFont,
        fmt.Sprintf("Enemy HP: %d", s.enemyHP),
        types.Vector2{X: s.screenWidth - 150, Y: 10},
        types.Color{1, 0, 0, 1},
    )
    
    return nil
}

// RESULT: Only ~60 lines for battle scene with custom overlay!
```

---

### Strategy 2: Compositional Scene Components

**Approach**: Instead of one large BaseScene, provide **optional components** that scenes can compose.

#### Architecture

```go
// pkg/scene/components/layer_manager.go
package components

type LayerManager struct {
    layers     map[SceneLayer][]types.GameObject
    layerMutex sync.RWMutex
}

func NewLayerManager() *LayerManager {
    return &LayerManager{
        layers: make(map[SceneLayer][]types.GameObject),
    }
}

func (lm *LayerManager) Initialize() {
    lm.layerMutex.Lock()
    defer lm.layerMutex.Unlock()
    
    lm.layers[BACKGROUND] = []types.GameObject{}
    lm.layers[ENTITIES] = []types.GameObject{}
    lm.layers[UI] = []types.GameObject{}
}

func (lm *LayerManager) AddEntity(obj types.GameObject) {
    lm.layerMutex.Lock()
    defer lm.layerMutex.Unlock()
    lm.layers[ENTITIES] = append(lm.layers[ENTITIES], obj)
}

func (lm *LayerManager) GetRenderables() []types.GameObject {
    lm.layerMutex.RLock()
    defer lm.layerMutex.RUnlock()
    
    renderables := []types.GameObject{}
    renderables = append(renderables, lm.layers[BACKGROUND]...)
    renderables = append(renderables, lm.layers[ENTITIES]...)
    renderables = append(renderables, lm.layers[UI]...)
    return renderables
}

func (lm *LayerManager) Update(deltaTime float64) error {
    lm.layerMutex.RLock()
    defer lm.layerMutex.RUnlock()
    
    for _, layer := range []SceneLayer{BACKGROUND, ENTITIES, UI} {
        for _, obj := range lm.layers[layer] {
            if obj.IsVisible() {
                obj.Update(deltaTime)
            }
        }
    }
    return nil
}

// pkg/scene/components/input_manager.go
package components

type InputManager struct {
    inputCapturer types.InputCapturer
}

func NewInputManager() *InputManager {
    return &InputManager{}
}

func (im *InputManager) SetInputCapturer(ic types.InputCapturer) {
    im.inputCapturer = ic
}

func (im *InputManager) GetInputState() types.InputState {
    if im.inputCapturer != nil {
        return im.inputCapturer.GetInputState()
    }
    return types.InputState{}
}

// pkg/scene/components/state_manager.go
package components

type StateManager struct {
    stateChangeCallback func(types.GameStateType, map[string]interface{}) error
}

func NewStateManager() *StateManager {
    return &StateManager{}
}

func (sm *StateManager) SetStateChangeCallback(callback func(types.GameStateType, map[string]interface{}) error) {
    sm.stateChangeCallback = callback
}

func (sm *StateManager) RequestStateChange(newState types.GameStateType, data map[string]interface{}) error {
    if sm.stateChangeCallback != nil {
        return sm.stateChangeCallback(newState, data)
    }
    return nil
}
```

#### Example: User Scene with Compositional Components

```go
// examples/basic-game/scenes/compositional_scene.go
package scenes

import (
    "github.com/cstevenson98/gowasm-engine/pkg/scene/components"
)

type CompositionalScene struct {
    name         string
    screenWidth  float64
    screenHeight float64
    
    // Compose only what you need
    *components.LayerManager
    *components.InputManager
    *components.StateManager
    // NOT using SaveStateManager - don't need it!
}

func NewCompositionalScene(name string, width, height float64) *CompositionalScene {
    return &CompositionalScene{
        name:         name,
        screenWidth:  width,
        screenHeight: height,
        
        // Initialize components you want
        LayerManager: components.NewLayerManager(),
        InputManager: components.NewInputManager(),
        StateManager: components.NewStateManager(),
    }
}

func (s *CompositionalScene) GetName() string {
    return s.name
}

func (s *CompositionalScene) Initialize() error {
    // Initialize components
    s.LayerManager.Initialize()
    
    // Add game objects
    player := gameobject.NewPlayer(...)
    s.AddEntity(player)  // From LayerManager
    
    return nil
}

func (s *CompositionalScene) Update(deltaTime float64) error {
    // Use LayerManager's Update
    if err := s.LayerManager.Update(deltaTime); err != nil {
        return err
    }
    
    // Custom logic
    input := s.GetInputState()  // From InputManager
    if input.Buttons["Escape"] {
        s.RequestStateChange(types.GameStateMenu, nil)  // From StateManager
    }
    
    return nil
}

func (s *CompositionalScene) GetRenderables() []types.GameObject {
    return s.LayerManager.GetRenderables()
}

func (s *CompositionalScene) Cleanup() error {
    // Cleanup components if needed
    return nil
}

// Scene automatically implements:
// - SceneInputProvider (from InputManager)
// - SceneStateChangeRequester (from StateManager)
// But NOT SceneStateful (we didn't embed SaveStateManager)
```

**Pros of Compositional Approach:**
- Pick only what you need
- More explicit about dependencies
- Smaller memory footprint

**Cons:**
- More boilerplate (need to forward method calls)
- Need to explicitly embed each component
- More complex for beginners

---

### Strategy 3: Multiple Base Types

**Approach**: Provide predefined base scene types for common use cases.

```go
// pkg/scene/base_scenes.go

// MinimalScene - absolute minimum (no optional interfaces)
type MinimalScene struct {
    name         string
    screenWidth  float64
    screenHeight float64
    layers       map[SceneLayer][]types.GameObject
}

// StandardScene - most common needs (input, state change, canvas)
type StandardScene struct {
    *MinimalScene
    inputCapturer       types.InputCapturer
    stateChangeCallback func(types.GameStateType, map[string]interface{}) error
    canvasManager       canvas.CanvasManager
}

// StatefulScene - includes state save/restore
type StatefulScene struct {
    *StandardScene
    savedState map[string]interface{}
}

// FullScene - all interfaces (same as BaseScene in Strategy 1)
type FullScene struct {
    *StatefulScene
    gameStateManager interface{}
    requiredAssets   types.SceneAssets
}
```

**Usage:**

```go
// Use StandardScene for most games
type MyScene struct {
    *scene.StandardScene
    player *gameobject.Player
}

// Use StatefulScene for games with save/load
type SaveableScene struct {
    *scene.StatefulScene
    progress int
}

// Use MinimalScene for simple demos
type DemoScene struct {
    *scene.MinimalScene
}
```

**Pros:**
- Clear intent (name tells you what you get)
- Can choose appropriate level

**Cons:**
- More types to maintain
- Can be confusing which to use
- Still have some boilerplate

---

### Recommendation

**Use Strategy 1 (Full Default Implementation)** because:

1. ✅ **Simplest for users** - embed one thing, get everything
2. ✅ **Override only what you need** - most scenes need 3-4 methods
3. ✅ **No decision fatigue** - don't need to choose components
4. ✅ **Discoverable** - IDE autocomplete shows all available methods
5. ✅ **Backward compatible** - can still write scenes from scratch
6. ✅ **Memory impact minimal** - a few extra pointers per scene is negligible

**When to use Strategy 2 (Compositional):**
- Advanced users who want fine-grained control
- Very memory-constrained environments
- Could be offered as "pkg/scene/advanced" for power users

---

## Concern 2: Dependency Injection Pattern

### The Question

> "What would `engine.GetDependencies()` do? I don't understand the manual dependency injection concern."

### Current Manual Injection Problem

Right now, the engine manually checks and injects dependencies for each interface:

```go
// pkg/engine/engine.go:336-367 (CURRENT CODE)
func (e *Engine) RegisterScene(sceneName string, scene types.Scene) error {
    e.sceneMutex.Lock()
    defer e.sceneMutex.Unlock()
    
    e.scenes[sceneName] = scene
    
    // ❌ PROBLEM: Manual type assertion for EACH interface (42 lines!)
    
    // Check if scene wants input
    if inputProvider, ok := scene.(types.SceneInputProvider); ok {
        inputProvider.SetInputCapturer(e.inputCapturer)
    }
    
    // Check if scene wants state changes
    if stateRequester, ok := scene.(types.SceneStateChangeRequester); ok {
        stateRequester.SetStateChangeCallback(e.SetGameState)
    }
    
    // Check if scene wants canvas
    if textureProvider, ok := scene.(types.SceneTextureProvider); ok {
        textureProvider.SetCanvasManager(e.canvasManager)
    }
    
    // Check if scene wants game state
    if gameStateUser, ok := scene.(types.SceneGameStateUser); ok {
        gameStateUser.SetGameState(e.gameStateProvider)
    }
    
    // Check if scene needs assets loaded
    if assetProvider, ok := scene.(types.SceneAssetProvider); ok {
        assets := assetProvider.GetRequiredAssets()
        // Load assets...
    }
    
    // ... repeat for all 7 interfaces
    
    return nil
}
```

**Problems with this approach:**

1. **Maintenance burden**: Adding a new interface requires updating this function
2. **Boilerplate**: 42 lines of repetitive type assertions
3. **Error-prone**: Easy to forget to add injection for a new interface
4. **Not extensible**: Can't easily add custom injection logic

---

### Proposed Solution Architecture

Create a **dependency container** that scenes can query for what they need:

#### Option A: Dependency Container Struct

```go
// pkg/engine/dependencies.go (NEW FILE)
package engine

import (
    "github.com/cstevenson98/gowasm-engine/pkg/types"
    "github.com/cstevenson98/gowasm-engine/pkg/canvas"
)

// EngineDependencies holds all injectable dependencies from the engine
type EngineDependencies struct {
    InputCapturer       types.InputCapturer
    CanvasManager       canvas.CanvasManager
    StateChangeCallback func(types.GameStateType, map[string]interface{}) error
    GameStateProvider   interface{}
    AssetLoader         types.AssetLoader
    ScreenWidth         float64
    ScreenHeight        float64
}

// GetDependencies creates a dependency container with all available services
func (e *Engine) GetDependencies() *EngineDependencies {
    return &EngineDependencies{
        InputCapturer:       e.inputCapturer,
        CanvasManager:       e.canvasManager,
        StateChangeCallback: e.SetGameState,
        GameStateProvider:   e.gameStateProvider,
        AssetLoader:         e.assetLoader,
        ScreenWidth:         e.config.Screen.Width,
        ScreenHeight:        e.config.Screen.Height,
    }
}
```

#### BaseScene with Dependency Injection

```go
// pkg/scene/base_scene.go
package scene

// BaseScene can receive dependencies in one call
func (b *BaseScene) InjectDependencies(deps *engine.EngineDependencies) {
    b.inputCapturer = deps.InputCapturer
    b.canvasManager = deps.CanvasManager
    b.stateChangeCallback = deps.StateChangeCallback
    b.gameStateManager = deps.GameStateProvider
    b.screenWidth = deps.ScreenWidth
    b.screenHeight = deps.ScreenHeight
    
    // Any additional setup
    b.Initialize()
}
```

#### Simplified Engine Registration

```go
// pkg/engine/engine.go (IMPROVED)
func (e *Engine) RegisterScene(sceneName string, scene types.Scene) error {
    e.sceneMutex.Lock()
    defer e.sceneMutex.Unlock()
    
    e.scenes[sceneName] = scene
    
    // ✅ SOLUTION: Single injection call instead of 42 lines!
    if injectable, ok := scene.(types.SceneInjectable); ok {
        injectable.InjectDependencies(e.GetDependencies())
    } else {
        // Fallback: old manual injection for backward compatibility
        e.manuallyInjectDependencies(scene)
    }
    
    // Load assets if scene provides them
    if assetProvider, ok := scene.(types.SceneAssetProvider); ok {
        assets := assetProvider.GetRequiredAssets()
        return e.assetLoader.LoadAssets(assets)
    }
    
    return nil
}

// Backward compatibility method (can be removed in v2.0)
func (e *Engine) manuallyInjectDependencies(scene types.Scene) {
    if inputProvider, ok := scene.(types.SceneInputProvider); ok {
        inputProvider.SetInputCapturer(e.inputCapturer)
    }
    // ... other manual injections
}
```

#### New Interface

```go
// pkg/types/scene_extras.go
package types

import "github.com/cstevenson98/gowasm-engine/pkg/engine"

// SceneInjectable is implemented by scenes that use dependency injection
type SceneInjectable interface {
    InjectDependencies(deps *engine.EngineDependencies)
}
```

---

### Implementation Examples

#### Example 1: BaseScene Auto-Implements Injection

```go
// pkg/scene/base_scene.go
package scene

import "github.com/cstevenson98/gowasm-engine/pkg/engine"

// BaseScene automatically implements SceneInjectable
type BaseScene struct {
    // ... fields
}

func (b *BaseScene) InjectDependencies(deps *engine.EngineDependencies) {
    // Automatically populate all fields
    b.inputCapturer = deps.InputCapturer
    b.canvasManager = deps.CanvasManager
    b.stateChangeCallback = deps.StateChangeCallback
    b.gameStateManager = deps.GameStateProvider
    b.screenWidth = deps.ScreenWidth
    b.screenHeight = deps.ScreenHeight
}

// User scene gets injection for free:
type MyScene struct {
    *scene.BaseScene  // Gets InjectDependencies() automatically!
    player *gameobject.Player
}

// Engine usage:
engine.RegisterScene("my-scene", NewMyScene())
// ^ Automatically injects all dependencies via BaseScene.InjectDependencies()
```

#### Example 2: Custom Scene with Selective Injection

```go
// User scene that only wants some dependencies
type MinimalCustomScene struct {
    name          string
    inputCapturer types.InputCapturer
    canvasManager canvas.CanvasManager
    // Not using state changes or game state
}

// Implement SceneInjectable to get only what you need
func (m *MinimalCustomScene) InjectDependencies(deps *engine.EngineDependencies) {
    m.inputCapturer = deps.InputCapturer
    m.canvasManager = deps.CanvasManager
    // Ignore other dependencies
}

// Still need to implement Scene interface manually
func (m *MinimalCustomScene) GetName() string { return m.name }
func (m *MinimalCustomScene) Initialize() error { return nil }
func (m *MinimalCustomScene) Update(deltaTime float64) error { return nil }
func (m *MinimalCustomScene) GetRenderables() []types.GameObject { return nil }
func (m *MinimalCustomScene) Cleanup() error { return nil }
```

#### Example 3: Scene with Custom Post-Injection Setup

```go
type ComplexScene struct {
    *scene.BaseScene
    customService *MyCustomService
}

// Override InjectDependencies to do custom setup
func (c *ComplexScene) InjectDependencies(deps *engine.EngineDependencies) {
    // Call base injection first
    c.BaseScene.InjectDependencies(deps)
    
    // Then do custom setup that requires injected dependencies
    c.customService = NewMyCustomService(deps.CanvasManager, deps.InputCapturer)
    
    // Can access injected fields now
    logger.Info("Scene screen size: %fx%f", c.screenWidth, c.screenHeight)
}
```

---

### Option B: Interface-Based Injection (Alternative)

Instead of a struct, use an interface for the dependency container:

```go
// pkg/types/dependencies.go
package types

type DependencyProvider interface {
    GetInputCapturer() InputCapturer
    GetCanvasManager() canvas.CanvasManager
    GetStateChangeCallback() func(GameStateType, map[string]interface{}) error
    GetGameStateProvider() interface{}
    GetScreenSize() (width, height float64)
}

// BaseScene uses the interface
func (b *BaseScene) InjectDependencies(provider DependencyProvider) {
    b.inputCapturer = provider.GetInputCapturer()
    b.canvasManager = provider.GetCanvasManager()
    b.stateChangeCallback = provider.GetStateChangeCallback()
    b.gameStateManager = provider.GetGameStateProvider()
    b.screenWidth, b.screenHeight = provider.GetScreenSize()
}

// Engine implements the interface
func (e *Engine) GetInputCapturer() types.InputCapturer {
    return e.inputCapturer
}
// ... other interface methods
```

**Pros of interface approach:**
- More flexible (can mock for testing)
- Don't expose internal engine structure

**Cons:**
- More verbose
- More boilerplate in engine

---

### Recommendation

**Use Option A (Struct-based EngineDependencies)** because:

1. ✅ **Simpler** - just a data struct
2. ✅ **Clearer** - can see all available dependencies
3. ✅ **Less boilerplate** - no getter methods
4. ✅ **Efficient** - direct field access
5. ✅ **Still mockable** - can create test EngineDependencies in tests

---

## Summary

### Concern 1: Optional Interfaces
**Answer**: Use full BaseScene with default implementations. Scenes override only what they need via Go's method shadowing.

**Result**: 
- Scenes go from 200+ lines to 30-70 lines
- No interface boilerplate needed
- Override only Initialize(), Update(), and 1-2 custom methods

### Concern 2: Dependency Injection
**Answer**: Create `EngineDependencies` struct with all injectable services. BaseScene implements `InjectDependencies()` method.

**Result**:
- Engine registration: 42 lines → 5 lines
- Scenes get all dependencies in one call
- Easy to add new dependencies in the future
- Backward compatible with old manual injection

### Code Reduction Summary

| Component | Before | After | Reduction |
|-----------|--------|-------|-----------|
| Scene boilerplate | 200+ lines | 30-70 lines | **70% reduction** |
| Engine injection | 42 lines | 5 lines | **88% reduction** |
| Interface implementations | 77 lines per scene | 0 lines (inherited) | **100% reduction** |

---

**Next Steps**:
1. Implement BaseScene with full default implementations
2. Add EngineDependencies struct
3. Update engine RegisterScene to use dependency injection
4. Provide migration guide for existing scenes
5. Create examples showing both approaches (embedded BaseScene vs custom)


