package states

import (
	"example.com/grid-sim-game/game/components/grid"
	"example.com/grid-sim-game/game/components/network"
	"example.com/grid-sim-game/game/gameconfig"
	"example.com/grid-sim-game/game/systems/camera"
	"example.com/grid-sim-game/game/systems/loadflow"
	"example.com/grid-sim-game/game/systems/loadtick"
	"example.com/grid-sim-game/game/systems/placement"
	"example.com/grid-sim-game/game/systems/pointer"
	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/imgui"
	"github.com/cstevenson98/gowasm-engine/pkg/prefab"
	"github.com/cstevenson98/gowasm-engine/pkg/state"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// GridState is the (only) state of the grid-sim-game example: a scrollable
// grid the player populates with generators, houses and lines via a
// clickable toolbar.
type GridState struct {
	*state.BaseState
	linePathFilter   *ecs.Filter1[grid.LinePath]  // thick polylines
	junctionFilter   *ecs.Filter1[grid.GridObject] // junction circles
	gridObjectFilter *ecs.Filter1[grid.GridObject] // gen ghost ports
}

// NewGridState creates the grid state.
func NewGridState() *GridState {
	return &GridState{BaseState: state.NewBaseState("Grid")}
}

// Enter seeds resources and registers systems. Empty cells are drawn
// procedurally (no blank ECS entities).
func (s *GridState) Enter(deps state.Deps) error {
	if err := s.BaseState.Enter(deps); err != nil {
		return err
	}

	ecs.SetResource(s.World(), &grid.PlacementState{Tool: grid.ToolNone})
	ecs.SetResource(s.World(), grid.NewGridOccupancy())
	ecs.SetResource(s.World(), network.NewElectricalNetwork())

	cfg := gameconfig.Global
	if cam := ecs.GetResource[components.Camera](s.World()); cam != nil {
		cam.Y = -cfg.ToolbarHeight
	}

	// One background quad under entities (overlays must not paint opaque fills
	// over the playfield — they run after the world pass).
	prefab.NewBackground(
		s.World(),
		types.Vector2{},
		types.Vector2{X: cfg.WorldWidth(), Y: cfg.WorldHeight()},
		cfg.BlankTexture,
	)

	// Pointer (hover/select) before placement; load-tick mutates P/Q;
	// LoadflowSystem re-solves only when Dirty; camera scrolls last.
	s.Schedule().
		Add(pointer.NewPointerSystem(s.World())).
		Add(placement.NewPlacementSystem(s.World())).
		Add(loadtick.NewLoadTickSystem(s.World())).
		Add(loadflow.NewLoadflowSystem(s.World())).
		Add(camera.NewCameraScrollSystem(cfg.CameraSpeed))

	return nil
}

// DrawOverlays draws the toolbar, procedural grid chrome, and debug console.
func (s *GridState) DrawOverlays() error {
	s.renderToolbar()
	s.renderGridChrome()
	return s.BaseState.DrawOverlays()
}

// RenderImGui draws the right-half network inspector panel.
func (s *GridState) RenderImGui(ctx *imgui.Context) {
	cfg := gameconfig.Global
	screenW := s.ScreenWidth()
	screenH := s.ScreenHeight()
	panelW := cfg.SidePanelWidth(screenW)
	panelX := screenW - panelW

	ctx.Panel("Network", panelX, 0, panelW, screenH, func(w *imgui.WindowBuilder) {
		net := ecs.GetResource[network.ElectricalNetwork](s.World())
		if net == nil {
			w.Text("No electrical network resource")
			return
		}
		s.renderNetworkPanel(w, net)
	})
}
