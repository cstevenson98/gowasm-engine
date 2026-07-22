# Go 2D Game Engine (Library)

A 2D game engine written in Go, rendered with [Ebiten](https://ebiten.org/) and
built around an **Entity Component System (ECS)**. Game flow is organised into
**States**, each owning its own ECS **World**; behaviour lives in **Systems**
that operate on pure-data **Components**. The same code runs on the desktop and
in the browser (WebAssembly), because Ebiten abstracts the platform — there is
no WebGPU, no `syscall/js`, and no `//go:build js` split.

Example games live under `examples/` and consume the engine as a Go module.

> The module path (`github.com/cstevenson98/gowasm-engine`) predates the move to
> Ebiten and is retained for compatibility.

## Quick Start

Prerequisites:
- Go 1.24+
- A C toolchain and the Ebiten system dependencies (see below)
- Git LFS (for image assets — see below)

### Desktop system dependencies

Ebiten needs OpenGL/X11 (or Wayland) development libraries to build the desktop
binary:

```bash
# Ubuntu/Debian
sudo apt-get install build-essential libgl1-mesa-dev xorg-dev

# NixOS — a ready-made shell is provided
nix-shell examples/ebiten-demo/shell.nix
```

The `shell.nix` also sets `LD_LIBRARY_PATH` for OpenGL, which fixes
`libGL.so: cannot open shared object file` at runtime.

### Git LFS

PNG images and other binary assets are stored with Git LFS. After cloning you
**must** pull the real files, or you will only have small pointer files and asset
loading will fail:

```bash
sudo apt-get install git-lfs   # Ubuntu/Debian (or: brew install git-lfs)
git lfs install
git lfs pull
```

### Build and run (desktop)

```bash
make build-desktop   # build the Ebiten desktop binary -> build/game-desktop
make run-desktop     # build and run the example game
```

### Browser (WebAssembly)

The engine and example compile cleanly to `GOOS=js GOARCH=wasm`
(`cd examples && make serve`), but the browser target is **not yet fully wired
for assets**: textures are loaded via `ebitenutil.NewImageFromFile`, which on the
web becomes an HTTP fetch relative to the page, and the example currently has no
`index.html` bootstrap and serves assets without the `assets/` path prefix the
code expects. For a robust dual desktop/web build, embed assets with `go:embed`
and load them through an `fs.FS`. Until that lands, treat the wasm build as
compile-only and run the game on the desktop.

## Architecture

High-level per-frame flow:

```
input.Poll -> engine refreshes Input resource -> active State.Update
  -> State.Schedule runs systems (input -> movement -> animation -> ...)
  -> render.Renderer draws the World (Background -> Entities -> UI, by Order.Z)
  -> State.DrawOverlays (menus / HUD / debug console)
```

- **Engine** owns the game loop and the active State, refreshes the per-frame
  `Input` resource, applies deferred state switches, and renders the active
  World.
- **State** owns one ECS World and an ordered system Schedule; it builds
  entities and registers systems in `Enter`.
- **Systems** hold all behaviour and run synchronously on the loop.
- **Components** are pure data. **Resources** are per-World singletons.

### Modules (multi-module workspace)

- **Root** `github.com/cstevenson98/gowasm-engine` — the engine library (`pkg/`).
- **`examples/basic-game`** (`example.com/basic-game`) — the example game;
  its `game/` package (`package main`) is the **browser (wasm) entry point**.
- **`cmd/ebiten-game`** (`example.com/ebiten-game`) — the **desktop entry point**.

The entry modules use `replace` directives to point at the local engine. When
dependencies change, run `go mod tidy` in each affected module.

### Engine packages (`pkg/`)

- `engine` — game loop (`ebiten.Game`), State orchestration, Input refresh,
  deferred state switches.
- `ecs` — the ECS abstraction and the **sole backend seam** (wraps
  `github.com/mlange-42/ark`). Nothing else imports Ark. Exposes `World`,
  `Entity`, `Map1..8`, `Filter1..4`, resources, `System`/`Schedule`.
- `components` — pure-data components (`Position`, `Velocity`, `Wrap`, `Sprite`,
  `Animation`), layer tags (`LayerBackground/Entities/UI`), `Order`, and the
  `ScreenBounds` / `Input` resources.
- `state` — `State` interface, `BaseState`, injected `Deps`, and the optional
  `OverlayRenderer` interface.
- `systems` — engine systems (`Movement`, `Animation`) and `systems/battle`
  (the ATB battle subsystem).
- `render` — the `Renderer`: one filtered pass per layer, ordered by `Order.Z`.
- `prefab` — entity builder helpers (e.g. `NewBackground`).
- `canvas` — thin Ebiten drawing facade. `input` — keyboard/gamepad polling.
  `ui` — immediate-mode overlay drawing. `text`, `config`, `debug`, `logger`,
  `types` — support.

## Using as a Library

```go
// main.go (desktop entry point)
eng := engine.NewEngine()

// A game-defined provider for cross-state data (optional).
eng.RegisterGameStateProvider(myGameState)

// Register states; the engine injects deps (input, UI, screen size,
// state-change callback, game-state provider) into each state on activation.
eng.RegisterState(types.MENU, states.NewMenuState())
eng.RegisterState(types.GAMEPLAY, states.NewGameplayState())

_ = eng.Initialize("")               // canvasID unused by the Ebiten backend
_ = eng.SetGameState(types.MENU)     // activates the starting state
eng.Start()
if err := ebiten.RunGame(eng); err != nil {
    log.Fatal(err)
}
```

Writing a state:

```go
type GameplayState struct{ *state.BaseState }

func NewGameplayState() *GameplayState {
    return &GameplayState{BaseState: state.NewBaseState("Gameplay")}
}

func (s *GameplayState) Enter(deps state.Deps) error {
    if err := s.BaseState.Enter(deps); err != nil { // seeds ScreenBounds + Input
        return err
    }
    // build entities (spawners / prefabs) ...
    s.Schedule().
        Add(entities.NewPlayerInputSystem(s.World())).
        Add(systems.NewMovement(s.World())).
        Add(systems.NewAnimation(s.World()))
    return nil
}
```

Local development from your game's `go.mod`:

```go
require github.com/cstevenson98/gowasm-engine v0.0.0
replace github.com/cstevenson98/gowasm-engine => ../path/to/engine/repo
```

## Configuration

Global configuration lives in `pkg/config` as `config.Global`:
- `Screen` — virtual width/height and window/canvas size.
- `Player` — spawn, size, speed, texture, sprite grid.
- `Animation` — default frame times.
- `Rendering` — `PixelArtMode`, scaling, line spacing, etc.
- `Debug` — console toggle, font path/scale, colours, message settings.
- `Battle` — example-game combat parameters.

## Build, Test, and Docs

```bash
# Tests (plain Go, no browser, no build tags)
make test        # go test ./pkg/...
make test-all    # all modules

# Build / run
make build-desktop
make run-desktop
cd examples && make serve   # wasm build + local server

# Quality
make fmt
make lint        # golangci-lint if installed
make tidy        # root + entry modules

# Docs (package overviews from doc.go)
make docs        # browsable docs at http://localhost:6060 (override DOCS_PORT)
make docs-cli    # print overviews to the terminal
```

Start reading the generated docs at the `engine` and `ecs` packages.

## Directory Layout

```
pkg/
  engine/       # Game loop, State orchestration
  ecs/          # ECS seam (only importer of the Ark backend)
  components/   # Pure-data components + resources
  state/        # State interface, BaseState, Deps
  systems/      # Movement/Animation systems + systems/battle
  render/       # Renderer (layer passes, Order.Z)
  prefab/       # Entity builder helpers
  canvas/       # Ebiten drawing facade
  input/        # Keyboard/gamepad input
  ui/ text/     # Overlay + text rendering
  config/ debug/ logger/ types/   # Support

examples/
  Makefile      # Builds examples to wasm; serves examples/dist
  basic-game/
    assets/     # Example assets (Git LFS)
    game/       # package main — browser (wasm) entry point
    states/     # Game states (Menu, Gameplay, PlayerMenu, Battle)
    game/entities/    # Game components, spawners, systems
    game/gamestate/   # Persistent cross-state game data
    go.mod
cmd/
  ebiten-game/  # package main — desktop entry point
```

## Using from a Private Repository

This repository is private; consuming it as a module needs auth config.

- **Local dev (recommended):** use a `replace` directive (see above) — no auth
  needed.
- **CI/remote:** set `GOPRIVATE` and configure git auth:

```bash
go env -w GOPRIVATE=github.com/cstevenson98/*
# SSH:
git config --global url."git@github.com:".insteadOf "https://github.com/"
# or a PAT via ~/.netrc / credential helper, or `gh auth login`.
```

Tag releases (`git tag v0.1.0 && git push origin v0.1.0`) to pin versions.

## Troubleshooting

**Images fail to load / repeated "failed to load image".** Git LFS assets
weren't pulled (PNGs are ~130-byte pointer files). Fix:

```bash
git lfs install && git lfs pull
ls -lah examples/basic-game/assets/*.png   # should be KB, not bytes
```

**Font textures missing/corrupt.** Regenerate the sprite-sheet fonts:

```bash
sudo apt-get install python3-pil
cd scripts
python3 font_spritesheet_generator.py --font Mono --size 10 \
    --output ../examples/basic-game/assets/fonts/
```

This produces `.sheet.png` + `.sheet.json` (proper fonts are ~5KB).

**Module auth errors (`go get`/`go mod tidy`).** Verify `go env GOPRIVATE` and
test git access (`git ls-remote git@github.com:cstevenson98/gowasm-engine.git`).

## Roadmap

- [ ] Collision detection system
- [ ] Audio support
- [ ] Particle effects
- [ ] ECS-backed battle components (fold Participant timers/stats into the World)
- [ ] Asset loading system
- [ ] Mobile touch controls

## License

MIT License — see the LICENSE file.
