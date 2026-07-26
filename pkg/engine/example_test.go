package engine_test

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/cstevenson98/milo/pkg/config"
	"github.com/cstevenson98/milo/pkg/engine"
	"github.com/cstevenson98/milo/pkg/prefab"
	"github.com/cstevenson98/milo/pkg/state"
	"github.com/cstevenson98/milo/pkg/types"
)

// MenuState is a custom state. Embedding state.BaseState provides the whole
// State interface (its own ecs.World, a system schedule, dependency accessors),
// so a state only has to override what it cares about.
type MenuState struct {
	*state.BaseState
}

// NewMenuState constructs the state.
func NewMenuState() *MenuState {
	return &MenuState{BaseState: state.NewBaseState("Menu")}
}

// Enter builds the state's entities. Always call the embedded BaseState.Enter
// first to store dependencies and seed common resources.
func (s *MenuState) Enter(deps state.Deps) error {
	if err := s.BaseState.Enter(deps); err != nil {
		return err
	}

	// A background is a full-screen static sprite entity.
	prefab.NewBackground(
		s.World(),
		types.Vector2{X: 0, Y: 0},
		types.Vector2{X: s.ScreenWidth(), Y: s.ScreenHeight()},
		"assets/art/background.png",
	)
	return nil
}

// Example shows the minimal end-to-end setup: build the engine, register one or
// more states against game states, pick a starting state, then hand control to
// Ebiten's run loop.
func Example() {
	eng := engine.NewEngine(config.Default())

	// Register states for the values your game can be in.
	eng.RegisterState(types.MENU, NewMenuState())

	// canvasID is unused by the Ebiten backend; pass "".
	if err := eng.Initialize(""); err != nil {
		log.Fatal(err)
	}

	// Activate the starting state, then start the engine.
	if err := eng.SetGameState(types.MENU); err != nil {
		log.Fatal(err)
	}
	eng.Start()

	// ebiten.RunGame blocks until the window closes.
	if err := ebiten.RunGame(eng); err != nil {
		log.Fatal(err)
	}
}
