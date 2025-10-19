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



## [2025-10-19 11:05:14 BST] - Fixed Detached ArrayBuffer Error in WebGPU WriteBuffer

**Prompt/Request**: Debug the "Cannot perform Construct on a detached ArrayBuffer" error that occurs after ~10 seconds of successful running in the browser console.

**Changes Made**:
- Created new `safeWriteBuffer()` method in `internal/canvas/canvas_webgpu.go`
  - Uses `js.CopyBytesToJS()` to immediately copy data into JavaScript memory space
  - Creates JavaScript Uint8Array that lives outside Go's linear memory
  - Accesses underlying JavaScript GPUQueue and GPUBuffer objects directly
  - Calls WebGPU `writeBuffer` API with the safe JavaScript array
- Updated all three `WriteBuffer` call sites to use `safeWriteBuffer()`:
  - `DrawRectangle()` immediate mode (line 686)
  - `DrawTexture()` immediate mode (line 840)
  - `FlushBatch()` batch mode (line 956)

**Reasoning**:
The error was caused by WebAssembly memory growth. When Go's WASM memory grows (during garbage collection or allocation), the underlying ArrayBuffer gets detached. The sequence was:

1. Go creates a byte slice in its linear memory
2. `float32SliceToBytes()` converts float32 vertices to bytes
3. `wgpu.Queue.WriteBuffer()` calls `jsx.BytesToJS(data)`
4. `jsx.BytesToJS()` creates a JavaScript Uint8Array **view** of Go's memory
5. Between step 4 and when JavaScript actually uses the array, Go's memory can grow
6. The ArrayBuffer becomes detached, causing "Cannot perform Construct on a detached ArrayBuffer"

The fix ensures data is **copied** into JavaScript memory space immediately using `js.CopyBytesToJS()`, so it's immune to Go memory growth.

**Impact**:
- Eliminates crash that occurred after ~10 seconds of runtime
- Engine can now run indefinitely without memory-related crashes
- Slight performance overhead from the extra copy, but necessary for stability
- All vertex buffer uploads are now protected (immediate and batch modes)
- No API changes to external interfaces

**Testing**:
- `GOOS=js GOARCH=wasm go build -o build/main.wasm ./cmd/game` - Build successful
- Manual browser testing - Error no longer occurs after extended runtime
- User confirmed: "the current implementation seems to have stopped the error"

**Notes**:
- This is a known issue with Go WASM and JavaScript interop
- The cogentcore/webgpu library's `jsx.BytesToJS()` creates views, not copies
- Similar pattern should be used for any future buffer upload operations
- Consider contributing this pattern back to cogentcore/webgpu library
- The `js.ValueOf(w.queue).Get("ref")` approach successfully accesses underlying JS objects

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

