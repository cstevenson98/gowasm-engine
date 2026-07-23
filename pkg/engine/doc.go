// Package engine is the heart of a small 2D game engine built on top of
// [Ebiten] and an Entity Component System (ECS). It owns the game loop and
// coordinates the other packages (ecs, state, components, systems, render,
// canvas, input, ui) into a running game.
//
// This package doc is the recommended starting point for understanding the
// engine; the [github.com/cstevenson98/gowasm-engine/pkg/ecs] and
// [github.com/cstevenson98/gowasm-engine/pkg/state] packages are the next reads.
//
// # Mental model
//
// A game is a set of States, one active at a time, selected by a game state
// value:
//
//		GameState ──▶ State ──▶ ecs.World ──▶ entities (Components) + Systems
//
//	  - An [Engine] holds a registry of states keyed by
//	    [github.com/cstevenson98/gowasm-engine/pkg/types.GameState].
//	  - A State owns one ecs.World and an ordered system Schedule.
//	  - Entities are composed of pure-data Components; Systems hold the
//	    behaviour and run each frame.
//	  - The Engine drives everything by forwarding the per-frame Update and Draw
//	    calls to the active state and its World.
//
// # The game loop
//
// Engine implements Ebiten's game interface, so Ebiten calls these three
// methods for you, 60 times per second, on a single goroutine:
//
//   - Update: polls input, refreshes the World's Input resource, applies any
//     deferred state switch, then calls the active state's Update(dt) (which
//     runs its system Schedule).
//   - Draw: renders the active World with the render.Renderer (one pass per
//     layer, ordered by Order.Z), then draws state overlays (menus, HUD, debug
//     console) for states implementing state.OverlayRenderer.
//   - Layout: reports the virtual resolution; Ebiten scales it to the window.
//
// You do not call Update/Draw yourself. You build the engine, register states,
// pick a starting state, then hand the engine to [ebiten.RunGame].
//
// # State lifecycle
//
// Activating a state runs a fixed sequence:
//
//  1. Exit the outgoing state (its World is reset).
//  2. Enter the incoming state with injected dependencies (see below), which is
//     where it builds entities and registers systems.
//  3. Build a render.Renderer over the new state's World and make it active.
//
// There is no separate asset-preload step: the canvas lazy-loads (and caches)
// each texture the first time it is drawn.
//
// Use [Engine.SetGameState] to activate the starting state before the loop
// runs. States request a switch from within their own Update via the injected
// state-change callback ([Engine.RequestStateChange]), which defers the switch
// to the start of the next frame - so a state is never torn down while it is
// still executing its own Update.
//
// # Dependency injection
//
// States never construct engine services themselves. The engine passes them in
// via state.Deps at Enter: the input capturer, the UI facade, the virtual
// screen size, the state-change callback, the game-defined game-state
// provider, and the relevant slices of the engine's own config.Settings (debug
// console enablement, default animation frame time). Embedding
// [github.com/cstevenson98/gowasm-engine/pkg/state.BaseState] stores these and
// exposes them through accessors, and seeds the ScreenBounds, Input, and
// Camera resources on the World.
//
// Because each State owns its own World, data that must survive a switch (for
// example a player's position) is the game's responsibility - typically written
// to the game-state provider on Exit and read back on Enter.
//
// # Textures
//
// Textures are addressed by file path. Declared assets are preloaded at the
// state switch, and the canvas also lazy-loads any texture the first time it is
// drawn, so entities spawned at runtime render without extra bookkeeping.
//
// # Quick start
//
//	eng := engine.NewEngine(config.Default())
//	eng.RegisterState(types.MENU, NewMenuState())
//	if err := eng.Initialize(""); err != nil { // canvasID is unused by Ebiten
//		log.Fatal(err)
//	}
//	_ = eng.SetGameState(types.MENU)
//	eng.Start()
//	if err := ebiten.RunGame(eng); err != nil {
//		log.Fatal(err)
//	}
//
// [Ebiten]: https://ebiten.org/
// [ebiten.RunGame]: https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2#RunGame
package engine
