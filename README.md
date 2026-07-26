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

# NixOS — a ready-made shell is provided at the repo root
nix-shell
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

### Test / run examples

```bash
make test            # engine unit tests (./pkg/...)
make run-demo        # desktop: examples/demo (counter)
make build-examples  # WASM → examples/build + examples/dist
make serve           # WASM build + local HTTP server
```

Sibling games (`rpg-game`, `energy-tycoon`) are separate repos — run them with
`go run ./game` from their own tree.

### Browser (WebAssembly)

Examples compile to `GOOS=js GOARCH=wasm` via `make serve`. The browser target
may still need an `index.html` bootstrap and asset path wiring per example;
treat wasm as compile/serve for demos and prefer desktop (`make run-demo`) for
day-to-day work.

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
- **`examples/demo`** (`example.com/demo`) — minimal counter demo (engine + state + UI).
- **Sibling games** (separate repos, local `replace` → this engine):
  [`rpg-game`](https://github.com/cstevenson98/rpg-game), `energy-tycoon`.
  Each owns its own `game/` entry point.

Consumer modules use `replace` directives to point at the local engine. When
dependencies change, run `go mod tidy` in each affected module.

### Engine packages (`pkg/`)

- `engine` — game loop (`ebiten.Game`), State orchestration, Input refresh,
  deferred state switches.
- `ecs` — the ECS abstraction and the **sole backend seam** (wraps
  `github.com/mlange-42/ark`). Nothing else imports Ark. Exposes `World`,
  `Entity`, `Map1..8`, `Filter1..4`, resources, `System`/`Schedule`.
- `components` — pure-data components (`Position`, `Velocity`, `Wrap`, `Sprite`,
  `Animation`, `CameraTarget`), layer tags (`LayerBackground/Entities/UI`),
  `Order`, and the `ScreenBounds` / `Input` / `Camera` resources.
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
cfg := config.Default()          // or build a custom config.Settings
eng := engine.NewEngine(cfg)

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

The engine's own configuration is `config.Settings` (`pkg/config`) - there is no
global instance. Build one explicitly, or start from `config.Default()`, and
pass it to `engine.NewEngine`; the engine threads it down into the canvas, UI,
text, and debug console:
- `Screen` — virtual width/height (window/canvas size is derived via
  `Settings.WindowWidth/WindowHeight`).
- `Animation` — `DefaultFrameTime`, the fallback used by generic prefab helpers.
- `Rendering` — `PixelArtMode`, scaling, line spacing, etc.
- `Debug` — console toggle, font path/scale, colours, message settings.

Game-specific configuration (player stats/appearance, enemy/battle content,
...) does not live here - each game defines its own config package. See
`rpg-game/game/gameconfig` (sibling repo) for an example, which - unlike the
engine's config - is a plain package-level global, since it configures one
specific game rather than a reusable engine.

## Build, Test, and Docs

```bash
# Tests (plain Go, no browser, no build tags)
make test        # go test ./pkg/...
make test-all    # all modules

# Examples
make run-demo
make build-examples
make serve       # wasm build + local server

# Quality
make fmt
make lint        # golangci-lint if installed
make tidy        # root + examples/*/go.mod

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
  Makefile      # WASM build + serve (also via root `make serve`)
  demo/         # Minimal counter demo (1280×720, Up arrow)
    assets/fonts/
    game/       # package main
    states/
    go.mod
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
ls -lah examples/demo/assets/fonts/*.png   # should be KB, not bytes
```

**Font textures missing/corrupt.** Regenerate the sprite-sheet fonts:

```bash
sudo apt-get install python3-pil
cd scripts
python3 font_spritesheet_generator.py --font Mono --size 10 \
    --output ../examples/demo/assets/fonts/
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
