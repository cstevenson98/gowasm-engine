// Package scene organises the game objects that make up one screen of a game
// (a title screen, the overworld, a battle) and manages their lifecycle.
//
// The engine keeps a registry of scenes keyed by game state and activates one
// at a time. See [github.com/cstevenson98/gowasm-engine/pkg/engine] for how
// scenes fit into the game loop.
//
// # Layers
//
// A scene draws its objects back-to-front in three fixed layers, so you never
// manage z-order by hand:
//
//   - BACKGROUND: drawn first (skyboxes, tilemaps, backdrops).
//   - ENTITIES:   drawn second (players, enemies, items).
//   - UI:         drawn last (HUD, menus, overlays).
//
// Add objects with AddBackground, AddEntity, and AddUI. GetRenderables returns
// them already flattened in the correct draw order.
//
// # BaseScene
//
// Most scenes embed [BaseScene], which implements the full [Scene] interface
// (Initialize, Update, GetRenderables, Cleanup, GetName) plus the engine's
// dependency-injection contract. A concrete scene typically only overrides
// Initialize to build its objects, and Update if it needs custom per-frame
// logic:
//
//	type OverworldScene struct {
//		*scene.BaseScene
//	}
//
//	func NewOverworldScene() *OverworldScene {
//		return &OverworldScene{BaseScene: scene.NewBaseScene("Overworld", 640, 480)}
//	}
//
//	func (s *OverworldScene) Initialize() error {
//		if err := s.BaseScene.Initialize(); err != nil {
//			return err
//		}
//		s.AddEntity(gameobject.NewLlama(types.Vector2{X: 100, Y: 100}, types.Vector2{X: 64, Y: 64}, 50))
//		return nil
//	}
//
// Always call the embedded BaseScene.Initialize first.
//
// # Dependency injection
//
// By embedding BaseScene a scene automatically receives engine services (input,
// canvas, a state-change callback, the game's state provider, and the virtual
// screen size) when it becomes active. Read them through the accessors BaseScene
// provides rather than constructing services yourself.
//
// To switch scenes from within a scene (for example a menu picking "Play"),
// call the injected state-change callback. The engine defers the switch to the
// next frame, so it is safe to keep running the rest of the current Update.
//
// # Assets
//
// A scene can declare the textures and fonts it needs so the engine preloads
// them at activation time instead of blocking mid-frame. Declare them during
// construction/Initialize; the engine reads them before calling Initialize on
// the incoming scene.
//
// # Persisting data across switches
//
// A scene is torn down (Cleanup) when it deactivates and rebuilt (Initialize)
// when it reactivates, so scene fields do not survive a round-trip. Anything
// that must persist (a player's position, story flags) should be written to a
// shared game-state object owned by the game in Cleanup and read back in
// Initialize - the engine itself keeps no per-scene state.
package scene
