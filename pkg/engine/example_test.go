package engine_test

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/cstevenson98/gowasm-engine/pkg/engine"
	"github.com/cstevenson98/gowasm-engine/pkg/gameobject"
	"github.com/cstevenson98/gowasm-engine/pkg/mover"
	"github.com/cstevenson98/gowasm-engine/pkg/scene"
	"github.com/cstevenson98/gowasm-engine/pkg/sprite"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// MenuScene is a custom scene. Embedding scene.BaseScene provides the whole
// Scene interface plus automatic dependency injection, so a scene only has to
// override what it cares about.
type MenuScene struct {
	*scene.BaseScene
}

// NewMenuScene constructs the scene at a virtual resolution.
func NewMenuScene() *MenuScene {
	return &MenuScene{BaseScene: scene.NewBaseScene("Menu", 640, 480)}
}

// Initialize builds the scene's game objects. Always call the embedded
// BaseScene.Initialize first to set up the layers.
func (s *MenuScene) Initialize() error {
	if err := s.BaseScene.Initialize(); err != nil {
		return err
	}

	// A background is a full-screen sprite (1x1 sheet) at a fixed position.
	spr := sprite.NewSpriteSheet("assets/art/background.png", sprite.Vector2{X: 640, Y: 480}, 1, 1)
	pos := mover.NewBasicMover(types.Vector2{X: 0, Y: 0}, types.Vector2{}, 640, 480)
	bg := gameobject.NewBaseGameObject(spr, pos, types.ObjectState{ID: "background", Visible: true})
	s.AddBackground(bg)

	return nil
}

// Example shows the minimal end-to-end setup: build the engine, register one or
// more scenes against game states, pick a starting state, then hand control to
// Ebiten's run loop.
func Example() {
	eng := engine.NewEngine()

	// Register scenes for the states your game can be in.
	eng.RegisterScene(types.MENU, NewMenuScene())

	// canvasID is unused by the Ebiten backend; pass "".
	if err := eng.Initialize(""); err != nil {
		log.Fatal(err)
	}

	// Activate the starting scene, then start the engine.
	if err := eng.SetGameState(types.MENU); err != nil {
		log.Fatal(err)
	}
	eng.Start()

	// ebiten.RunGame blocks until the window closes.
	if err := ebiten.RunGame(eng); err != nil {
		log.Fatal(err)
	}
}
