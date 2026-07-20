// Package engine is the heart of a small component-based 2D game engine built
// on top of [Ebiten]. It owns the game loop and coordinates the other packages
// (canvas, input, scene, gameobject, sprite, mover) into a running game.
//
// This package doc is the recommended starting point for understanding the
// engine. Each collaborating package (scene, gameobject, sprite) has its own
// overview with more detail.
//
// # Mental model
//
// A game is a set of scenes, one active at a time, selected by a game state:
//
//		GameState ──▶ Scene ──▶ GameObjects ──▶ Sprite (look) + Mover (position)
//
//	  - An [Engine] holds a registry of scenes keyed by [github.com/cstevenson98/gowasm-engine/pkg/types.GameState].
//	  - A Scene owns GameObjects organised into back-to-front layers.
//	  - A GameObject is identity + behaviour; it delegates its appearance to a
//	    Sprite and its position to a Mover.
//	  - The Engine drives everything by forwarding the per-frame Update and Draw
//	    calls to the active scene.
//
// # The game loop
//
// Engine implements Ebiten's game interface, so Ebiten calls these three
// methods for you, 60 times per second, on a single goroutine:
//
//   - Update: polls input and calls the active scene's Update(deltaTime).
//   - Draw: asks the active scene for its renderables and draws each one, then
//     renders scene overlays (menus, HUD, debug console).
//   - Layout: reports the virtual resolution; Ebiten scales it to the window.
//
// You do not call Update/Draw yourself. You build the engine, register scenes,
// pick a starting state, then hand the engine to [ebiten.RunGame].
//
// # Scene lifecycle
//
// Activating a scene runs a fixed sequence:
//
//  1. Cleanup the outgoing scene.
//  2. Preload the incoming scene's declared assets (textures and fonts) so
//     blocking I/O happens up front, not mid-frame.
//  3. Inject dependencies (see below).
//  4. Initialize the incoming scene and make it active.
//
// Use [Engine.SetGameState] to activate the starting scene before the loop
// runs. Scenes request a switch from within their own Update via the injected
// state-change callback ([Engine.RequestStateChange]), which defers the switch
// to the start of the next frame. Deferring means a scene is never cleaned up
// while it is still executing its own Update, so scene code needs no
// "am I still alive?" guards. Persisting anything across a switch (for example
// a player's position) is the game's responsibility - typically by writing it
// to a shared game-state object in Cleanup and reading it back in Initialize.
//
// # Dependency injection
//
// Scenes never construct engine services themselves. Instead the engine passes
// them in. A scene that implements SceneInjectable receives everything in one
// call via InjectDependencies(deps), where deps exposes the input capturer, the
// canvas manager, a state-change callback, the game's state provider, and the
// virtual screen size. Embedding [github.com/cstevenson98/gowasm-engine/pkg/scene.BaseScene]
// implements SceneInjectable for you, so most scenes get injection for free.
//
// The state-change callback is how a scene asks to switch scenes (for example a
// menu selecting "Play"): it calls [Engine.RequestStateChange].
//
// # Textures
//
// Textures are addressed by file path. Declared assets are preloaded at scene
// switch, and the canvas also lazy-loads any texture the first time it is
// drawn, so objects spawned at runtime render without extra bookkeeping.
//
// # Quick start
//
// A minimal program wires an engine to Ebiten:
//
//	eng := engine.NewEngine()
//	eng.RegisterScene(types.MENU, NewMenuScene())
//	if err := eng.Initialize(""); err != nil { // canvasID is unused by Ebiten
//		log.Fatal(err)
//	}
//	_ = eng.SetGameState(types.MENU)
//	eng.Start()
//	if err := ebiten.RunGame(eng); err != nil {
//		log.Fatal(err)
//	}
//
// See the runnable examples in this package for a complete, compiling setup
// including a custom scene.
//
// [Ebiten]: https://ebiten.org/
// [ebiten.RunGame]: https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2#RunGame
package engine
