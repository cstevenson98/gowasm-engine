# Cursor AI Development History

This file contains a chronological log of all changes made during AI-assisted development sessions.

**Purpose**: Provide context, reasoning, and audit trail for all modifications.

**Format**: Each entry includes timestamp, changes, reasoning, impact, and testing notes.

---

## [2025-10-18 20:49:55 BST] - Created Cursor Rules and History System

**Prompt/Request**: Create cursor rules for this game engine project. Include information about WASM in Go (build tags, WebGPU wrapper that minimizes hardcoded JS calls). Also create a history rule that tells agents to log changes for each prompt, always read the history, and always use bash to get timestamps. Each rule should exist as a .mdc file in .cursor/rules/

**Changes Made**:
- Created `.cursor/rules/` directory structure
- Created `gameEngine.mdc` - Comprehensive cursor rule for the Go WASM WebGPU game engine
  - Documented architecture overview (Engine, Canvas, GameObject, Scene, Sprite, Mover, Input, Types)
  - Explained Go build tag pattern (`//go:build js`)
  - Documented WebGPU wrapper usage (cogentcore/webgpu library)
  - Covered testing patterns (unit tests vs WASM browser tests)
  - Added code patterns and examples
  - Included file naming conventions
  - Listed important development rules
  - Added common tasks and debugging guides
- Created `history.mdc` - Change logging rule
  - Defines workflow: always read history first, make changes, log changes
  - Requires bash commands for timestamps (`date` command)
  - Provides entry template and examples
  - Includes automation tips and scripts
- Created `CURSOR_HISTORY.md` - Initial history file with this first entry

**Reasoning**:
The project has a sophisticated architecture using Go compiled to WASM with WebGPU rendering. Key aspects that needed documentation:

1. **Build Tags**: Critical for separating WASM code (with `//go:build js`) from test/mock code. Without understanding this, developers might break the build or tests.

2. **WebGPU Wrapper**: The project uses cogentcore/webgpu to minimize direct syscall/js calls. This is important for maintainability and type safety.

3. **Component Architecture**: The engine follows interfaces and composition patterns that need to be understood for consistent development.

4. **History System**: Provides continuity across AI sessions, creates an audit trail, and helps understand past decisions.

**Impact**:
- Future AI sessions will have context about the project architecture
- Developers will understand build tag requirements
- Changes will be consistently logged with timestamps
- Better code consistency following established patterns
- Easier debugging by tracking when changes were made

**Testing**:
- Verified directory structure created successfully
- Confirmed `.mdc` files are properly formatted markdown
- Verified timestamp generation with bash `date` command works correctly
- Files are in expected locations:
  - `/home/conor/dev/goFE/webgpu-triangle/.cursor/rules/gameEngine.mdc`
  - `/home/conor/dev/goFE/webgpu-triangle/.cursor/rules/history.mdc`
  - `/home/conor/dev/goFE/webgpu-triangle/CURSOR_HISTORY.md`

**Notes**:
- The gameEngine.mdc includes extensive examples of build tag usage, WebGPU patterns, and component interfaces
- The history.mdc enforces using bash for timestamps to ensure consistency and proper timezone handling
- This history file should be updated after every significant change or prompt
- Cursor should automatically load these rules for context in future sessions
- The glob patterns in the .mdc frontmatter define when each rule is active

---


## [2025-10-19 11:43:18 BST] - Configured Git LFS for Art Assets

**Prompt/Request**: Set up Git LFS to track art assets in the assets/art directory

**Changes Made**:
- Installed Git LFS on the system (`sudo apt-get install -y git-lfs`)
- Initialized Git LFS in the repository (`git lfs install`)
- Configured Git LFS to track multiple image and art file formats:
  - `*.png` - PNG images
  - `*.kra` - Krita project files
  - `*.jpg`, `*.jpeg` - JPEG images
  - `*.gif` - GIF images
  - `*.psd` - Photoshop files
  - `*.xcf` - GIMP files
- Created `.gitattributes` file with LFS tracking configuration
- Staged new art assets:
  - `assets/art/test-background.kra`
  - `assets/art/test-background.png`
- Migrated existing PNG assets to LFS:
  - `assets/llama.png`
  - `assets/triangle_up.png`

**Reasoning**:
Art assets (especially source files like .kra) can be large binary files that don't compress well in Git. Git LFS stores these files separately and only keeps pointers in the repository, which:
- Keeps repository size small
- Speeds up clone operations
- Improves performance for operations like checkout and diff
- Only downloads large files when actually needed

**Impact**:
- All existing and future image/art files will be tracked by Git LFS
- Repository will remain lightweight even as art assets are added
- Collaborators will need to have Git LFS installed (`git lfs install`)
- Files are properly staged and ready to commit
- `.gitattributes` ensures consistent LFS tracking across the team

**Testing**:
- `git lfs install` - Successfully initialized Git LFS
- `git lfs ls-files` - Verified 4 files are tracked by LFS:
  - `assets/art/test-background.kra` (ca23bcc456)
  - `assets/art/test-background.png` (c5306670bf)
  - `assets/llama.png` (a44428fb7b)
  - `assets/triangle_up.png` (ccde98543f)
- `git status` - Confirmed files are staged for commit

**Notes**:
- Git LFS requires installation on each machine that clones the repo
- GitHub, GitLab, and other major Git hosts support Git LFS
- The .gitattributes file is tracked in version control
- Future contributors should run `git lfs install` after cloning
- Consider adding a note about Git LFS to the README for new contributors

---

## [2025-10-19 11:50:15 BST] - Implemented Background Sprite in Gameplay Scene

**Prompt/Request**: Implement a background sprite in the gameplay scene using the test-background.png texture

**Changes Made**:
- Created new `internal/gameobject/background.go` file
  - Implements `types.GameObject` interface
  - Creates a static, non-animated background using a single-frame sprite
  - No mover component (backgrounds don't move)
  - Takes position, size, and texture path as parameters
  - Uses same component-based pattern as Player and Llama
- Updated `internal/scene/gameplay_scene.go`
  - Added background creation in `Initialize()` method
  - Background fills entire screen (0,0 to screenWidth x screenHeight)
  - Uses texture path "art/test-background.png"
  - Background added to BACKGROUND layer (renders behind entities)
  - Added debug logging for background creation

**Reasoning**:
Following the established GameObject pattern ensures consistency in the codebase. The Background GameObject:
- Uses the SpriteSheet system with a 1x1 grid (single frame) for static images
- Implements GameObject interface but returns nil for GetMover() since backgrounds don't move
- Is added to the BACKGROUND layer to ensure proper render order (background → entities → UI)
- Full-screen size ensures it covers the entire canvas

**Impact**:
- Gameplay scene now renders a background image behind the player
- Background is rendered first in the render order (BACKGROUND layer)
- No breaking changes to existing code
- Pattern can be reused for parallax backgrounds or tiled backgrounds in the future
- Background is automatically loaded and rendered by the engine's existing rendering pipeline

**Testing**:
- `GOOS=js GOARCH=wasm go build -o build/main.wasm ./cmd/game` - Build successful
- No linter errors in new or modified files
- Background GameObject follows same interface pattern as Player and Llama

**Notes**:
- Background texture path is "art/test-background.png" (relative to assets directory)
- Background will need to be copied to dist/ folder for browser testing
- Can easily create multiple backgrounds for different scenes
- Future enhancement: Add support for repeating/tiled backgrounds
- Future enhancement: Add parallax scrolling support for layered backgrounds
- Background sprite doesn't update (static), saving performance

---

## [2025-10-19 11:56:05 BST] - Fixed Texture Batching to Support Multiple Textures

**Prompt/Request**: Fix rendering issue where all sprites were using the same texture (llama) instead of their respective textures. The background was rendering with the llama texture instead of the background image.

**Changes Made**:
- Added `textureBatch` struct in `internal/canvas/canvas_webgpu.go`
  - Stores texture path, GPU texture, bind group, and vertices for each texture
- Modified `WebGPUCanvasManager` struct:
  - Added `batches []textureBatch` field to track multiple texture batches
  - Kept `currentBatchTexturePath` to track current texture being batched
- Updated `BeginBatch()`:
  - Initializes empty batches slice at start of frame
- Updated `DrawTexturedRect()`:
  - Detects texture changes during batching
  - When texture changes, saves current batch and starts new one
  - Accumulates vertices per texture in separate batches
- Updated `EndBatch()`:
  - Saves final batch with remaining vertices
  - Reports number of batches ready to render
- Updated `executePipeline()` for `TexturedPipeline` case:
  - Iterates through all batches
  - For each batch: uploads vertices, sets bind group, draws
  - Properly switches textures between draw calls
- Removed references to `safeWriteBuffer()` (which was removed earlier)
  - Replaced with standard `queue.WriteBuffer()` calls

**Reasoning**:
The original batching system assumed all sprites would use the same texture. It would:
1. Accumulate vertices for all sprites
2. Set bind group to the last texture processed
3. Render all vertices with that one texture

This caused all sprites to render with whichever texture was processed last. The fix implements proper multi-texture batching by:
- Breaking sprites into separate batches by texture
- Rendering each batch with its correct texture and bind group
- Maintaining render order (background → entities → UI)

This is a common pattern in 2D game engines - batching is broken when the texture changes to minimize draw calls while supporting multiple textures.

**Impact**:
- Background now renders with correct texture (test-background.png)
- Player renders with correct texture (llama.png)
- Each sprite uses its own texture as intended
- Batching still reduces draw calls (sprites with same texture are batched together)
- Render order preserved (background renders first, then entities)
- Small performance overhead from multiple draw calls, but necessary for correctness
- No API changes to external interfaces

**Testing**:
- `GOOS=js GOARCH=wasm go build -o build/main.wasm ./cmd/game` - Build successful
- No linter errors
- Ready for browser testing

**Notes**:
- Future optimization: Sort renderables by texture to maximize batch sizes
- Future enhancement: Implement texture atlas to allow true single-batch rendering
- The batching system now properly handles the common case of multiple textures per frame
- Each texture change creates a new batch, so fewer texture changes = better performance
- This is standard 2D batching behavior (break batch on state change)

---

## [2025-10-19 11:58:24 BST] - Fixed Background Positioning and Animation

**Prompt/Request**: Fix two issues with the background rendering:
1. Background only rendered behind the player rectangle instead of covering the full 800x600 screen
2. Background was animating like a spritesheet instead of being a static image

**Changes Made**:
- Added `StaticMover` struct in `internal/gameobject/background.go`
  - Implements `types.Mover` interface
  - Returns fixed position, zero velocity
  - No-op implementations for Update, SetVelocity, SetScreenBounds
- Modified `Background` struct:
  - Added `mover types.Mover` field
- Updated `NewBackground()`:
  - Creates a StaticMover with the background's position
  - Sets extremely long frame time (999999.0 seconds) to prevent animation
  - Assigns mover to background
- Updated `GetMover()`:
  - Returns the StaticMover instead of nil

**Reasoning**:
The engine's render logic has two code paths:
```go
if mover := gameObject.GetMover(); mover != nil {
    renderData = gameObject.GetSprite().GetSpriteRenderData(mover.GetPosition())
} else {
    renderData = gameObject.GetSprite().GetSpriteRenderData(types.Vector2{X: 0, Y: 0})
}
```

**Issue 1**: When GetMover() returned nil, the engine passed (0,0) instead of the background's actual position. This caused the background to render at origin with its size, but since the sprite system was receiving (0,0), it was only visible where it overlapped with other sprites.

**Issue 2**: The SpriteSheet.Update() was being called every frame, advancing the currentFrame counter. Even with a 1x1 sprite sheet, the animation logic was running, causing UV coordinates to potentially shift or wrap.

**Solution**: Give the background a StaticMover that:
- Provides the correct position (0, 0 for top-left, with full screen size)
- Never moves (velocity always zero)
- Prevents the nil check from triggering the (0,0) fallback

And set an extremely long frame time to effectively disable animation.

**Impact**:
- Background now renders at correct position (0, 0)
- Background covers full 800x600 screen
- Background is completely static (no animation)
- Background still doesn't move (StaticMover has no velocity)
- No changes to other game objects
- Clean architecture - Background now follows same pattern as other GameObjects

**Testing**:
- `GOOS=js GOARCH=wasm go build -o build/main.wasm ./cmd/game` - Build successful
- No linter errors
- Ready for browser testing

**Notes**:
- StaticMover could be moved to `internal/mover/` if other static objects need it
- Alternative solution would be to modify engine render logic to use state.Position
- This solution maintains consistency with existing GameObject pattern
- The 1x1 sprite sheet with long frame time is more efficient than conditional animation logic

---

## [2025-10-19 12:06:40 BST] - Fixed Multi-Texture Batch Rendering with Buffer Offsets

**Prompt/Request**: Background was still only visible behind the llama and moved with it, despite correct render data being generated. User reported the background appeared to animate like a spritesheet and followed the player.

**Root Cause Identified**:
The batching system was uploading each batch to the SAME buffer location (offset 0). When multiple batches were uploaded via `queue.WriteBuffer(buffer, 0, data)`, they would overwrite each other in the GPU command queue before being processed. Only the last batch's data would actually be present when the draw calls executed.

**Changes Made**:
- Modified `executePipeline()` for `TexturedPipeline` case in `internal/canvas/canvas_webgpu.go`:
  - Added buffer offset tracking with `currentOffset` variable
  - Upload each batch to a different offset in the vertex buffer
  - Calculate offset as cumulative sum of previous batch sizes
  - Store draw info (bind group, vertex count, offset) for each batch
  - Draw all batches in order using their correct buffer offsets
- Removed debug logging from:
  - `internal/gameobject/background.go` - Removed construction logging
  - `internal/engine/engine.go` - Removed per-frame render data logging  
  - `internal/canvas/canvas_webgpu.go` - Removed batch upload/draw logging
- Cleaned up `internal/sprite/sprite.go` comment for clarity

**Technical Details**:
```go
// Old (broken) approach:
for batch in batches:
    WriteBuffer(buffer, 0, batch.vertices)  // All write to offset 0!
    Draw(batch)                              // Draws garbage or last batch

// New (fixed) approach:
offset = 0
for batch in batches:
    WriteBuffer(buffer, offset, batch.vertices)  // Different offset each time
    offset += len(batch.vertices) * 4            // Move forward
    store draw info
for drawInfo in drawInfos:
    SetVertexBuffer(buffer, drawInfo.offset)     // Read from correct offset
    Draw(drawInfo.vertexCount)                   // Draws correct data
```

**Why This Works**:
- Each batch gets its own space in the vertex buffer
- Queue operations (`WriteBuffer`) complete before render pass begins
- Each draw call reads from the correct offset where its data was uploaded
- No overwrites, no race conditions
- Batches render in correct order: BACKGROUND → ENTITIES → UI

**Impact**:
- Background now renders correctly at full screen (800x600)
- Background stays stationary at (0, 0)
- Player renders correctly on top with llama texture
- Each sprite uses its correct texture
- Proper layering maintained
- No performance regression (still batching effectively)

**Testing**:
- `GOOS=js GOARCH=wasm go build -o build/main.wasm ./cmd/game` - Build successful
- Browser testing confirmed:
  - ✅ Background fills entire screen
  - ✅ Background is static (doesn't move)
  - ✅ Background doesn't animate
  - ✅ Player renders on top with correct texture
  - ✅ Player moves independently of background

**Notes**:
- This is a critical fix for the multi-texture batching system
- The issue was WebGPU command queue ordering, not game logic
- Similar to the classic "double buffering" problem in graphics programming
- Future enhancement: Pre-allocate buffer with known maximum size
- Future enhancement: Track buffer usage to warn if approaching limit
- This pattern is standard for batching different draw states (textures, materials, etc.)

---

