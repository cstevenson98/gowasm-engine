# Ease of Use & Sprite Feature Improvements

**Date**: 2026-02-11  
**Purpose**: Comprehensive catalog of all code locations where ease-of-use and sprite feature improvements can be made.  
**Scope**: Focus on developer experience, not documentation (that's separate)

---

## Table of Contents

1. [Ease of Use Issues](#ease-of-use-issues)
   - [GameObject Boilerplate](#1-gameobject-boilerplate-duplication)
   - [Scene Setup Complexity](#2-scene-setup-complexity)
   - [Config System Issues](#3-config-system-coupling)
   - [Constructor Complexity](#4-constructor-complexity)
   - [Interface Implementation Overhead](#5-interface-implementation-overhead)
   - [Manual Coordinate Management](#6-manual-coordinate-management)
2. [Sprite Feature Gaps](#sprite-feature-gaps)
   - [Missing Core Features](#1-missing-core-sprite-features)
   - [Animation Limitations](#2-animation-system-limitations)
   - [Rendering Capabilities](#3-rendering-capabilities-gaps)

---

## EASE OF USE ISSUES

### 1. GameObject Boilerplate Duplication

Every GameObject implementation repeats the same boilerplate code. This violates DRY principles and increases maintenance burden.

#### **Issue 1.1: Identical State Management Code**

**Files affected:**
- `pkg/gameobject/player.go:71-98`
- `pkg/gameobject/llama.go:60-91`
- `pkg/gameobject/background.go:83-115`
- `pkg/gameobject/enemy.go:95-111`

**Repeated pattern:**
```go
// Every GameObject has this exact same code
func (x *GameObject) GetState() *types.ObjectState {
    return &x.state
}

func (x *GameObject) SetState(state types.ObjectState) {
    x.mu.Lock()
    defer x.mu.Unlock()
    x.state = types.CopyObjectState(state)
}

func (x *GameObject) GetID() string {
    x.mu.Lock()
    defer x.mu.Unlock()
    return x.state.ID
}
```

**Lines per file:**
- Player: 90-98, 210-216 (16 lines)
- Llama: 70-91 (21 lines)
- Background: 93-115 (22 lines)
- Enemy: 95-111 (16 lines)

**Total duplicated lines: 75**

**Proposed solution:**
```go
// pkg/gameobject/base.go (NEW FILE)
type BaseGameObject struct {
    sprite types.Sprite
    mover  types.Mover
    state  types.ObjectState
    mu     sync.Mutex
}

// Implements GetState(), SetState(), GetID(), GetSprite(), GetMover()
// All GameObjects can embed this
```

---

#### **Issue 1.2: Redundant Interface Implementations**

**Files affected:**
- `pkg/gameobject/player.go:80-88`
- `pkg/gameobject/llama.go:60-68`
- `pkg/gameobject/background.go:83-90`
- `pkg/gameobject/enemy.go:72-80`

**Repeated pattern:**
```go
// Every GameObject implements these identically
func (x *GameObject) GetSprite() types.Sprite {
    return x.sprite
}

func (x *GameObject) GetMover() types.Mover {
    return x.mover
}
```

**Total duplicated lines: 32**

These 7 lines are repeated in EVERY GameObject. With BaseGameObject, this drops to 0.

---

#### **Issue 1.3: Manual Component Composition**

**Files affected:**
- `pkg/gameobject/player.go:35-78` (constructor)
- `pkg/gameobject/llama.go:25-58` (constructor)
- `pkg/gameobject/background.go:55-81` (constructor)
- `pkg/gameobject/enemy.go:31-61` (constructor)

**Pattern - Player constructor:**
```go
func NewPlayer(position types.Vector2, size types.Vector2, moveSpeed float64) *Player {
    // 1. Manually create sprite (10 lines)
    playerSprite := sprite.NewSpriteSheet(
        config.Global.Player.TexturePath,
        sprite.Vector2{X: size.X, Y: size.Y},
        config.Global.Player.SpriteColumns,
        config.Global.Player.SpriteRows,
    )
    playerSprite.SetFrameTime(config.Global.Animation.PlayerFrameTime)

    // 2. Manually create mover (10 lines)
    playerMover := mover.NewBasicMover(
        position,
        types.Vector2{X: 0, Y: 0},
        size.X,
        size.Y,
    )
    playerMover.SetScreenBounds(config.Global.Screen.Width, config.Global.Screen.Height)

    // 3. Manually construct state (8 lines)
    return &Player{
        sprite:    playerSprite,
        mover:     playerMover,
        moveSpeed: moveSpeed,
        state: types.ObjectState{
            ID:       "Player",
            Position: position,
            Visible:  true,
        },
        // ... more fields
    }
}
```

**Problem:** 28+ lines of boilerplate PER GameObject type.

**Proposed solution:**
```go
// Convenience constructor
player := gameobject.NewAnimatedCharacter(gameobject.CharacterOptions{
    Texture:   "player.png",
    Position:  types.Vector2{X: 100, Y: 100},
    Size:      types.Vector2{X: 32, Y: 32},
    FrameGrid: types.Grid{Cols: 2, Rows: 3},
    Speed:     200.0,
})
// Result: 7 lines instead of 28
```

---

### 2. Scene Setup Complexity

Scenes require extensive boilerplate to implement 5+ optional interfaces and manage multiple systems.

#### **Issue 2.1: Excessive Interface Implementation Boilerplate**

**Files affected:**
- `examples/basic-game/scenes/gameplay_scene.go:19-96`
- `examples/basic-game/scenes/menu_scene.go:18-155`
- `examples/basic-game/scenes/battle_scene.go:19-171`

**Pattern in every scene:**
```go
type GameplayScene struct {
    name          string                // Required
    screenWidth   float64               // Required
    screenHeight  float64               // Required
    inputCapturer types.InputCapturer   // Optional interface 1
    stateChangeCallback func(...)       // Optional interface 2
    gameStateManager interface{}        // Optional interface 3
    layers        map[...]              // Required
    debugFont     text.Font             // Optional interface 4
    canvasManager canvas.CanvasManager  // Optional interface 5
    // ... 10+ more fields
}

// Must implement 7+ setter methods for dependency injection
func (s *GameplayScene) SetInputCapturer(inputCapturer types.InputCapturer) {
    s.inputCapturer = inputCapturer
}
func (s *GameplayScene) SetStateChangeCallback(callback func(...) error) {
    s.stateChangeCallback = callback
}
func (s *GameplayScene) SetGameState(gameState interface{}) {
    s.gameStateManager = gameState
}
func (s *GameplayScene) SetCanvasManager(cm canvas.CanvasManager) {
    s.canvasManager = cm
}
func (s *GameplayScene) GetRequiredAssets() types.SceneAssets {
    return types.SceneAssets{...}
}
// ... 3+ more methods
```

**Lines per scene:**
- GameplayScene: 77 lines of interface boilerplate (lines 19-96)
- MenuScene: 85 lines of interface boilerplate (lines 18-155)
- BattleScene: 89 lines of interface boilerplate (lines 19-171)

**Total: 251 lines of boilerplate across 3 scenes**

**Proposed solution:**
```go
// pkg/scene/base_scene.go (NEW FILE)
type BaseScene struct {
    name          string
    screenWidth   float64
    screenHeight  float64
    inputCapturer types.InputCapturer
    canvasManager canvas.CanvasManager
    layers        map[SceneLayer][]types.GameObject
    // Auto-implements all common interfaces
}

// User scene becomes:
type GameplayScene struct {
    *scene.BaseScene  // Embed base - gets all boilerplate
    player *gameobject.Player  // Only game-specific fields
}
```

---

#### **Issue 2.2: Manual Layer Management**

**Files affected:**
- `examples/basic-game/scenes/gameplay_scene.go:128-131, 139, 168`
- `examples/basic-game/scenes/menu_scene.go:160-163`
- `examples/basic-game/scenes/battle_scene.go:177-180, 188, 206`

**Pattern:**
```go
func (s *Scene) Initialize() error {
    // Every scene manually initializes layers (4 lines)
    s.layers[pkscene.BACKGROUND] = []types.GameObject{}
    s.layers[pkscene.ENTITIES] = []types.GameObject{}
    s.layers[pkscene.UI] = []types.GameObject{}

    // Then manually adds to layers
    s.AddGameObject(pkscene.BACKGROUND, background)
    s.AddGameObject(pkscene.ENTITIES, player)
}

func (s *Scene) AddGameObject(layer pkscene.SceneLayer, obj types.GameObject) {
    s.layers[layer] = append(s.layers[layer], obj)
}

func (s *Scene) GetRenderables() []types.GameObject {
    // Must manually concatenate in order (7+ lines)
    renderables := []types.GameObject{}
    renderables = append(renderables, s.layers[pkscene.BACKGROUND]...)
    renderables = append(renderables, s.layers[pkscene.ENTITIES]...)
    renderables = append(renderables, s.layers[pkscene.UI]...)
    return renderables
}
```

**Lines per scene:** 15-20 lines for layer management

**Proposed solution:**
```go
// BaseScene provides:
scene.AddBackground(background)
scene.AddEntity(player)
scene.AddUI(menu)
// GetRenderables() already implemented
```

---

#### **Issue 2.3: Font/Text Renderer Initialization Duplication**

**Files affected:**
- `examples/basic-game/scenes/gameplay_scene.go:98-122`
- `examples/basic-game/scenes/menu_scene.go:98-140`
- `examples/basic-game/scenes/battle_scene.go:85-171`

**Pattern repeated in EVERY scene:**
```go
// Debug console init (24 lines)
func (s *Scene) InitializeDebugConsole() error {
    if !config.Global.Debug.Enabled {
        return nil
    }
    s.debugFont = text.NewSpriteFont()
    err := s.debugFont.(*text.SpriteFont).LoadFont(config.Global.Debug.FontPath)
    if err != nil {
        return err
    }
    s.debugTextRenderer = text.NewTextRenderer(s.canvasManager)
    debug.Console.PostMessage("System", "Scene ready")
    return nil
}

// Menu text init (21 lines) - nearly identical!
func (s *Scene) InitializeMenuText() error {
    s.menuFont = text.NewSpriteFont()
    err := s.menuFont.(*text.SpriteFont).LoadFont(config.Global.Debug.FontPath)
    if err != nil {
        return err
    }
    s.menuTextRenderer = text.NewTextRenderer(s.canvasManager)
    return nil
}
```

**Total duplicated lines: 135 lines** (45 per scene × 3 scenes)

**Proposed solution:**
```go
// BaseScene auto-handles this
type BaseScene struct {
    // ...
    textSystems *scene.TextRenderingSystem  // Auto-initialized
}
// No manual setup needed
```

---

### 3. Config System Coupling

Global config creates tight coupling throughout the codebase.

#### **Issue 3.1: config.Global Usage Everywhere**

**Files affected (grep results):**
- `pkg/gameobject/player.go:38, 45, 56` (3 references)
- `pkg/gameobject/llama.go:35, 46` (2 references)
- `pkg/gameobject/enemy.go:55` (1 reference)
- `pkg/mover/basic_mover.go:22-23` (hardcoded defaults!)
- `examples/basic-game/scenes/gameplay_scene.go:90, 91, 94, 145, 181, 261, 269, 270, 359` (9 references)
- `examples/basic-game/scenes/menu_scene.go:145, 146, 151, 205, 346, 349, 350, 393, 394` (9 references)
- `examples/basic-game/scenes/battle_scene.go:145, 146, 147, 149, 196, 197, 248, 302, 439, 442, 443` (11 references)
- `examples/basic-game/game/main.go:51, 52, 63` (3 references)

**Total: 38 direct config.Global accesses**

**Example pattern:**
```go
// pkg/gameobject/player.go:38-45
playerSprite := sprite.NewSpriteSheet(
    config.Global.Player.TexturePath,    // Global access
    sprite.Vector2{X: size.X, Y: size.Y},
    config.Global.Player.SpriteColumns,  // Global access
    config.Global.Player.SpriteRows,     // Global access
)
playerSprite.SetFrameTime(config.Global.Animation.PlayerFrameTime) // Global access
```

**Problems:**
1. Can't test with custom configs
2. Can't have multiple games with different settings
3. Can't override per-scene or per-object
4. Engine code depends on game-specific config (Player, Battle, etc.)

**Proposed solution:**
```go
// Separate engine config from game config
type EngineConfig struct {
    Screen    ScreenSettings
    Rendering RenderingSettings
}

// Game-specific config stays in game code
type MyGameConfig struct {
    Player PlayerSettings
    Battle BattleSettings
}

// Pass config to constructors
player := gameobject.NewPlayer(position, &gameConfig.Player)
```

---

#### **Issue 3.2: Game-Specific Config in Engine Package**

**File:** `pkg/config/settings.go:20-76, 78-140`

**Problem:** Engine package contains game-specific configuration:
```go
// Lines 21-29: PlayerSettings - game-specific!
type PlayerSettings struct {
    SpawnX        float64
    SpawnY        float64
    Size          float64
    Speed         float64
    TexturePath   string
    SpriteColumns int
    SpriteRows    int
}

// Lines 60-76: BattleSettings - game-specific!
type BattleSettings struct {
    PlayerHP      int
    PlayerMaxHP   int
    EnemyHP       int
    EnemyMaxHP    int
    EnemyTexture  string
    // ... 7 more fields
}
```

**Why this is bad:**
- Library code shouldn't know about "Player" or "Battle"
- Makes it impossible to use engine for different game types
- Violates separation of concerns

**Proposed solution:**
Move to `examples/basic-game/config/game_settings.go`

---

### 4. Constructor Complexity

GameObjects have complex constructors with many parameters and manual setup.

#### **Issue 4.1: Many-Parameter Constructors**

**Files affected:**
- `pkg/gameobject/player.go:35`
- `pkg/gameobject/llama.go:25`
- `pkg/gameobject/background.go:55`
- `pkg/gameobject/enemy.go:31`

**Pattern:**
```go
// 3-4 parameters, BUT requires knowing internal structure
func NewPlayer(position types.Vector2, size types.Vector2, moveSpeed float64) *Player {
    // 43 lines of setup code
}

// Usage requires understanding texture, size, speed, frame grid...
player := gameobject.NewPlayer(
    types.Vector2{X: 100, Y: 100},
    types.Vector2{X: 32, Y: 32},
    200.0,
)
```

**Problems:**
1. Parameter meaning not clear from usage
2. Can't add optional parameters without breaking API
3. No validation of inputs
4. No defaults

**Proposed solution:**
```go
// Option pattern
type CharacterOptions struct {
    Position  types.Vector2
    Texture   string
    Size      types.Vector2      // Optional, auto-detected from texture
    Speed     float64            // Optional, default 100
    FrameGrid *types.Grid        // Optional, auto-detected or 1x1
    Animated  bool               // Optional, default true if grid > 1x1
}

player := gameobject.NewCharacter(gameobject.CharacterOptions{
    Position: types.Vector2{X: 100, Y: 100},
    Texture:  "player.png",
    Speed:    200.0,
    // Size and FrameGrid auto-detected!
})
```

---

#### **Issue 4.2: No Default Values or Convenience Constructors**

**Current state:** Every GameObject requires full manual construction.

**Examples of missing convenience:**
```go
// Current: Background requires manual size calculation
background := gameobject.NewBackground(
    types.Vector2{X: 0, Y: 0},
    types.Vector2{X: screenWidth, Y: screenHeight}, // Manual!
    "background.png",
)

// Current: Static sprites need animation disabled manually
sprite := sprite.NewSpriteSheet(texturePath, size, 1, 1)
sprite.SetFrameTime(999999.0) // Hack to prevent animation!

// Current: Every mover needs screen bounds set
mover := mover.NewBasicMover(pos, vel, w, h)
mover.SetScreenBounds(screenWidth, screenHeight) // Manual!
```

**Proposed convenience constructors:**
```go
// NEW: Backgrounds auto-fill screen
background := gameobject.NewScreenBackground("background.png")

// NEW: Static sprites
sprite := sprite.NewStaticSprite("item.png")

// NEW: Screen bounds from engine
mover := mover.NewScreenBoundMover(pos, vel, spriteSize)
```

---

### 5. Interface Implementation Overhead

Optional scene interfaces create massive boilerplate.

#### **Issue 5.1: Seven Optional Scene Interfaces**

**File:** `pkg/types/scene_extras.go:1-79`

**Current interfaces:**
1. `SceneOverlayRenderer` (lines 5-8)
2. `SceneTextureProvider` (lines 13-16)
3. `SceneInputProvider` (lines 21-25)
4. `SceneStateChangeRequester` (lines 30-34)
5. `SceneAssetProvider` (lines 44-51)
6. `SceneStateful` (lines 56-66)
7. `SceneGameStateUser` (lines 73-78)

**Problem:** Each interface requires:
- Field declaration in scene struct
- Setter method implementation (4-8 lines)
- Usage in scene logic

**Example - GameplayScene implements ALL 7:**
```go
// 77 lines just for interface implementations (lines 19-96)
type GameplayScene struct {
    inputCapturer       types.InputCapturer   // Interface 3
    stateChangeCallback func(...)             // Interface 4
    gameStateManager    interface{}           // Interface 7
    debugFont           text.Font             // Interface 5
    canvasManager       canvas.CanvasManager  // Interface 2
    savedPlayerPosition *types.Vector2        // Interface 6
    // ... more
}

// 7 setter methods
func (s *GameplayScene) SetInputCapturer(...) { ... }
func (s *GameplayScene) SetStateChangeCallback(...) { ... }
func (s *GameplayScene) SetGameState(...) { ... }
func (s *GameplayScene) SetCanvasManager(...) { ... }
func (s *GameplayScene) GetRequiredAssets() { ... }
func (s *GameplayScene) SaveState() { ... }
func (s *GameplayScene) RestoreState() { ... }
```

**Proposed solution:**
```go
// BaseScene handles all optional interfaces by default
type BaseScene struct {
    // Auto-implements all 7 interfaces
    // Scenes can override if needed
}
```

---

#### **Issue 5.2: Manual Dependency Injection**

**File:** `pkg/engine/engine.go:336-367`

**Engine manually injects dependencies:**
```go
// Lines 336-354: Manual injection for EACH interface
if inputProvider, ok := registeredScene.(types.SceneInputProvider); ok {
    inputProvider.SetInputCapturer(e.inputCapturer)
}
if stateRequester, ok := registeredScene.(types.SceneStateChangeRequester); ok {
    stateRequester.SetStateChangeCallback(e.SetGameState)
}
if gameStateUser, ok := registeredScene.(types.SceneGameStateUser); ok {
    gameStateUser.SetGameState(e.gameStateProvider)
}
// Repeat for 7 interfaces = 42 lines
```

**Proposed solution:**
```go
// Single injection point
scene.InjectDependencies(engine.GetDependencies())
```

---

### 6. Manual Coordinate Management

No helpers for common coordinate operations.

#### **Issue 6.1: Manual Screen Centering**

**Files affected:**
- `examples/basic-game/scenes/gameplay_scene.go:145-157`
- `examples/basic-game/scenes/menu_scene.go:343-366`
- `examples/basic-game/scenes/battle_scene.go:192-199`

**Pattern:**
```go
// Every scene manually calculates centers
playerX := s.screenWidth/2 - config.Global.Player.Size/2
playerY := s.screenHeight/2 - config.Global.Player.Size/2

// Menu centering is even worse (20+ lines!)
_, cellHeight := s.menuFont.GetCellSize()
lineHeight := float64(cellHeight)
if config.Global.Rendering.PixelPerfectScaling {
    lineHeight *= float64(config.Global.Rendering.PixelScale)
}
lineHeight *= config.Global.Rendering.UILineSpacing
totalHeight := float64(len(menu.options)) * lineHeight
startY := (s.screenHeight - totalHeight) / 2
centerX := s.screenWidth / 2
```

**Proposed solution:**
```go
// Coordinate helpers
pos := types.CenterInScreen(spriteSize, screenSize)
pos := types.AlignCenter(text, screenSize)
pos := types.PositionAt(0.2, 0.5, screenSize) // 20% from left, 50% from top
```

---

#### **Issue 6.2: No Anchor Point Support**

**Files affected:**
- `pkg/sprite/sprite.go:55-78` (GetSpriteRenderData)
- `pkg/canvas/canvas_webgpu.go` (rendering assumes top-left origin)

**Problem:**
All sprites render from top-left corner. To center a sprite at a position:
```go
// User must manually offset
centeredPos := types.Vector2{
    X: position.X - sprite.GetSize().X/2,
    Y: position.Y - sprite.GetSize().Y/2,
}
```

**Proposed solution:**
```go
// Sprite with configurable anchor
sprite.SetAnchor(types.AnchorCenter)  // or TopLeft, BottomCenter, etc.
// Position now means center of sprite
sprite.SetPosition(types.Vector2{X: 100, Y: 100})
```

---

## SPRITE FEATURE GAPS

### 1. Missing Core Sprite Features

Critical features that modern sprite engines have.

#### **Feature 1.1: No Sprite Rotation**

**Files affected:**
- `pkg/sprite/sprite.go` (no rotation field or methods)
- `pkg/types/sprite.go:1-30` (no rotation in interface or render data)
- `pkg/canvas/interface.go:26` (DrawTextureRotated stub exists but unused)

**Current state:**
- Interface has `DrawTextureRotated` at line 26
- BUT SpriteRenderData has no rotation field
- Sprite has no SetRotation() method

**What's needed:**
```go
// pkg/types/sprite.go
type SpriteRenderData struct {
    TexturePath string
    Position    Vector2
    Size        Vector2
    UV          UVRect
    Rotation    float64  // NEW: Radians
    Visible     bool
}

// pkg/sprite/sprite.go
func (s *SpriteSheet) SetRotation(radians float64) {
    s.rotation = radians
}
```

**Impact:** Can't rotate enemies, projectiles, or effects.

---

#### **Feature 1.2: No Sprite Flipping**

**Files affected:**
- `pkg/sprite/sprite.go` (no flip fields)
- `pkg/types/sprite.go` (no flip in render data)

**Problem:** Can't reuse sprites for left/right movement.

**Common workaround:** Duplicate sprites in texture (wasteful).

**What's needed:**
```go
// pkg/types/sprite.go
type SpriteRenderData struct {
    // ...
    FlipX bool  // NEW
    FlipY bool  // NEW
}

// pkg/sprite/sprite.go
func (s *SpriteSheet) SetFlipX(flip bool) { s.flipX = flip }
func (s *SpriteSheet) SetFlipY(flip bool) { s.flipY = flip }
```

**Use cases:**
- Character facing left/right
- Coins spinning
- Particles

---

#### **Feature 1.3: No Per-Sprite Scaling**

**Files affected:**
- `pkg/sprite/sprite.go` (size is fixed at construction)
- `pkg/config/settings.go:56` (only global PixelScale)

**Current limitations:**
- Can't scale individual sprites
- All sprites use same PixelScale
- Can't animate size changes (growing/shrinking effects)

**What's needed:**
```go
// pkg/sprite/sprite.go
func (s *SpriteSheet) SetScale(scale float64) {
    s.scale = scale
}

// Or per-axis:
func (s *SpriteSheet) SetScaleXY(scaleX, scaleY float64) {
    s.scaleX = scaleX
    s.scaleY = scaleY
}
```

---

#### **Feature 1.4: No Sprite Tinting/Color Multiplication**

**Files affected:**
- `pkg/types/sprite.go` (no color in render data)
- `pkg/sprite/sprite.go` (no color field)

**Problem:** Can't change sprite colors dynamically.

**Use cases:**
- Damage flash (red tint)
- Power-up glow (color change)
- Fade in/out (alpha)
- Team colors

**What's needed:**
```go
// pkg/types/sprite.go
type SpriteRenderData struct {
    // ...
    Color [4]float32  // NEW: RGBA multiplier
}

// pkg/sprite/sprite.go
func (s *SpriteSheet) SetColor(r, g, b, a float32) {
    s.color = [4]float32{r, g, b, a}
}
```

---

#### **Feature 1.5: No Opacity/Alpha Control**

**Current state:** Sprites are either visible or invisible (boolean).

**Files affected:**
- `pkg/sprite/sprite.go:99-107` (SetVisible/IsVisible are boolean)
- `pkg/types/sprite.go:9` (Visible is bool)

**Problem:** Can't fade sprites in/out smoothly.

**What's needed:**
```go
// pkg/sprite/sprite.go
func (s *SpriteSheet) SetOpacity(opacity float64) {
    s.opacity = opacity  // 0.0 = transparent, 1.0 = opaque
}

func (s *SpriteSheet) FadeIn(duration float64) {
    // Animate opacity 0->1
}
```

---

### 2. Animation System Limitations

The animation system is basic and inflexible.

#### **Feature 2.1: No Multiple Animation Sequences**

**Files affected:**
- `pkg/sprite/sprite.go:7-28` (single animation only)

**Current limitation:**
Sprites can only have ONE animation (all frames in grid).

**Problem pattern:**
```go
// Can't have separate idle, walk, attack animations
// All 6 frames (2x3 grid) play in sequence
playerSprite := sprite.NewSpriteSheet(
    "player.png",
    size,
    2,  // columns
    3,  // rows = 6 frames total
)
// Plays: 0,1,2,3,4,5,0,1,2... forever
// Can't say "play frames 0-1 for idle, 2-3 for walk"
```

**What's needed:**
```go
// Define animation sequences
sprite.AddAnimation("idle", []int{0, 1})
sprite.AddAnimation("walk", []int{2, 3, 4, 5})
sprite.AddAnimation("attack", []int{6, 7, 8})

// Play specific animation
sprite.PlayAnimation("walk")
sprite.PlayAnimation("attack", false) // Don't loop
```

**Impact:** Currently impossible to make a character with multiple actions.

---

#### **Feature 2.2: No Animation Events/Callbacks**

**Files affected:**
- `pkg/sprite/sprite.go:86-97` (Update has no event system)

**Problem:** Can't trigger effects when animation frames change.

**Use cases:**
- Play footstep sound on frame 2 of walk cycle
- Spawn particle effect on frame 5 of attack
- Change hitbox on frame 3 of attack

**What's needed:**
```go
// Frame callbacks
sprite.OnFrame(2, func() {
    audio.Play("footstep.wav")
})

// Animation complete callback
sprite.OnComplete(func() {
    sprite.PlayAnimation("idle")
})
```

---

#### **Feature 2.3: No Animation Control**

**Files affected:**
- `pkg/sprite/sprite.go:86-97` (Update is automatic)

**Current limitations:**
- Can't pause animation
- Can't set animation speed
- Can't reverse animation
- Can't jump to specific frame

**What's needed:**
```go
// Control methods
sprite.PauseAnimation()
sprite.ResumeAnimation()
sprite.SetAnimationSpeed(2.0) // 2x speed
sprite.ReverseAnimation()
sprite.SetFrame(5) // Jump to frame 5
```

---

#### **Feature 2.4: No Non-Looping Animations**

**Files affected:**
- `pkg/sprite/sprite.go:86-97` (loops forever)

**Current behavior:**
```go
// Line 94: Always loops
s.currentFrame = (s.currentFrame + 1) % s.totalFrames
```

**Problem:** Attack animations should play once, not loop.

**What's needed:**
```go
// Loop control
sprite.SetLooping(false)
sprite.OnComplete(func() {
    sprite.PlayAnimation("idle")
})
```

---

### 3. Rendering Capabilities Gaps

Missing advanced rendering features.

#### **Feature 3.1: No Texture Atlas Support**

**Files affected:**
- `pkg/sprite/sprite.go` (assumes sprite sheets in grid)
- No atlas loader code exists

**Current limitation:**
Sprites must be in regular n×m grid. Can't use packed texture atlases.

**Problem:**
- Wasted texture memory (grid has empty spaces)
- Can't use tools like TexturePacker
- Can't optimize packing

**What's needed:**
```go
// Load from atlas
atlas := sprite.LoadTextureAtlas("sprites.json")
playerSprite := atlas.GetSprite("player/idle")
enemySprite := atlas.GetSprite("enemy/walk")
```

**Atlas format to support:**
- TexturePacker JSON
- Aseprite JSON
- Generic atlas format

---

#### **Feature 3.2: No Nine-Patch/Nine-Slice Support**

**Use case:** UI panels, dialogs, buttons that can scale without distortion.

**Current state:** Not implemented anywhere.

**What's needed:**
```go
// Nine-patch sprite
panel := sprite.NewNinePatch("panel.png", types.NinePatchMargins{
    Left:   10,
    Right:  10,
    Top:    10,
    Bottom: 10,
})
panel.SetSize(types.Vector2{X: 200, Y: 100}) // Scales correctly
```

---

#### **Feature 3.3: No Sprite Batching Hints**

**Files affected:**
- `pkg/canvas/canvas_webgpu.go` (auto-batches by texture)

**Current behavior:**
- Engine auto-batches by texture
- No control over batch order
- No sprite Z-order within layer

**Problem:** Can't optimize draw order or control Z-layering precisely.

**What's needed:**
```go
// Sprite Z-order
sprite.SetZOrder(10)  // Higher = drawn later (in front)

// Batch hints
sprite.SetBatchGroup("terrain")  // Group related sprites
```

---

#### **Feature 3.4: No Sprite Shaders/Effects**

**Current state:** All sprites use same textured pipeline.

**Use cases:**
- Outline shader (for selection)
- Glow shader (for power-ups)
- Distortion shader (for portals)
- Grayscale shader (for disabled items)

**What's needed:**
```go
// Custom shader per sprite
sprite.SetShader("outline")
sprite.SetShaderParam("color", types.Color{1, 0, 0, 1})
```

---

## SUMMARY STATISTICS

### Ease of Use Issues

| Category | Total Lines of Boilerplate | Files Affected | Priority |
|----------|---------------------------|----------------|----------|
| GameObject boilerplate | 139 | 4 | **HIGH** |
| Scene setup complexity | 251 | 3 | **HIGH** |
| Config system coupling | 38 access points | 8 | **MEDIUM** |
| Constructor complexity | ~120 | 4 | **MEDIUM** |
| Interface overhead | 77 per scene | 3 | **HIGH** |
| Coordinate management | 20 per scene | 3 | **LOW** |
| **TOTAL** | **~645 lines** | **25 files** | - |

### Sprite Feature Gaps

| Feature | Impact | Implementation Effort | Priority |
|---------|--------|----------------------|----------|
| Rotation | Can't rotate sprites | 1 day | **HIGH** |
| Flipping | Can't flip sprites | 0.5 days | **HIGH** |
| Per-sprite scaling | Limited effects | 0.5 days | **MEDIUM** |
| Color tinting | No damage flash | 1 day | **MEDIUM** |
| Opacity control | No fade effects | 0.5 days | **MEDIUM** |
| Multiple animations | Can't have idle/walk/attack | 2 days | **HIGH** |
| Animation events | Can't sync sound/effects | 1 day | **MEDIUM** |
| Animation control | Limited control | 1 day | **LOW** |
| Texture atlases | Wasted memory | 3 days | **LOW** |
| Nine-patch | UI limitation | 2 days | **LOW** |

---

## RECOMMENDED IMPLEMENTATION ORDER

### Phase 1: High-Impact Ease of Use (1 week)

1. **BaseGameObject** - Eliminates 139 lines of boilerplate
   - Files to create: `pkg/gameobject/base.go`
   - Files to modify: All GameObject implementations
   
2. **BaseScene** - Eliminates 251 lines of boilerplate
   - Files to create: `pkg/scene/base_scene.go`
   - Files to modify: All scene implementations

3. **Convenience Constructors** - Reduces code by 50%
   - Files to create: `pkg/gameobject/builders.go`
   - New APIs: NewSimpleSprite, NewAnimatedCharacter, etc.

### Phase 2: Critical Sprite Features (1 week)

4. **Sprite Rotation & Flipping** - Most requested features
   - Files to modify: `pkg/sprite/sprite.go`, `pkg/types/sprite.go`
   - Rendering impact: `pkg/canvas/canvas_webgpu.go`

5. **Multiple Animation Sequences** - Unlocks gameplay variety
   - Files to modify: `pkg/sprite/sprite.go`
   - New types: AnimationSequence, AnimationState

### Phase 3: Config System (1 week)

6. **Separate Engine/Game Config** - Better architecture
   - Files to create: `pkg/config/engine_config.go`
   - Files to move: Game-specific config to examples
   - Files to modify: All config.Global usage (38 locations)

### Phase 4: Advanced Features (2 weeks)

7. **Animation Events** - Gameplay enhancement
8. **Color Tinting & Opacity** - Visual effects
9. **Texture Atlas Support** - Memory optimization
10. **Coordinate Helpers** - Quality of life

---

## BREAKING CHANGES

Changes that would break existing code:

1. **Config restructure** - All `config.Global.Player` references need updating
2. **BaseGameObject** - Existing GameObjects need refactoring (simple)
3. **SpriteRenderData fields** - Adding rotation/flip/color
4. **Constructor signatures** - Switching to option pattern (can maintain old API)

## NON-BREAKING ADDITIONS

Can add without breaking existing code:

1. New convenience constructors (keep old ones)
2. Sprite features (add to interface, default implementation)
3. BaseScene (scenes can use it or not)
4. Animation enhancements (backward compatible)

---

## CONCLUSION

The engine has a solid foundation with **excellent architecture**, but suffers from:

1. **~645 lines of duplicated boilerplate** across scenes and game objects
2. **10+ critical sprite features missing** (rotation, flip, multi-animation)
3. **Tight coupling** through global config (38 access points)
4. **High barrier to entry** for new users (7 interfaces to implement)

Implementing the **High-Impact Phase 1 improvements** would:
- Reduce boilerplate by 60%
- Cut "Hello World" example from 460 lines to ~100 lines
- Make engine competitive with other sprite engines

Implementing **Critical Sprite Features (Phase 2)** would:
- Enable rotation, flipping, and multiple animations
- Match feature parity with engines like Phaser, PixiJS
- Unlock gameplay variety (idle/walk/attack animations)

**Total estimated effort: 5 weeks for all phases**

**Recommended MVP: Phase 1 + Phase 2 = 2 weeks**
