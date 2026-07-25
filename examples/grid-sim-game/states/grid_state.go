package states

import (
	"example.com/grid-sim-game/game/components/grid"
	"example.com/grid-sim-game/game/components/network"
	"example.com/grid-sim-game/game/gameconfig"
	"example.com/grid-sim-game/game/systems/camera"
	"example.com/grid-sim-game/game/systems/loadflow"
	"example.com/grid-sim-game/game/systems/loadtick"
	"example.com/grid-sim-game/game/systems/placement"
	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/imgui"
	"github.com/cstevenson98/gowasm-engine/pkg/state"
)

// GridState is the (only) state of the grid-sim-game example: a scrollable
// grid the player populates with generators, houses and lines via a
// clickable toolbar.
type GridState struct {
	*state.BaseState
}

// NewGridState creates the grid state.
func NewGridState() *GridState {
	return &GridState{BaseState: state.NewBaseState("Grid")}
}

// Enter builds the grid (one blank tile per cell) and registers the
// placement and camera-scroll systems.
func (s *GridState) Enter(deps state.Deps) error {
	if err := s.BaseState.Enter(deps); err != nil {
		return err
	}

	ecs.SetResource(s.World(), &grid.PlacementState{Tool: grid.ToolNone})
	ecs.SetResource(s.World(), grid.NewGridOccupancy())
	ecs.SetResource(s.World(), network.NewElectricalNetwork())

	cfg := gameconfig.Global
	// Start scrolled down just enough that row 0 clears the toolbar.
	if cam := ecs.GetResource[components.Camera](s.World()); cam != nil {
		cam.Y = -cfg.ToolbarHeight
	}

	for row := 0; row < cfg.GridRows; row++ {
		for col := 0; col < cfg.GridCols; col++ {
			grid.SpawnBlank(s.World(), grid.GridCoord{Col: col, Row: row})
		}
	}

	// Placement / load-tick mutate topology or house P/Q (mark Dirty);
	// LoadflowSystem re-solves only when Dirty; camera scrolls last.
	s.Schedule().
		Add(placement.NewPlacementSystem(s.World())).
		Add(loadtick.NewLoadTickSystem(s.World())).
		Add(loadflow.NewLoadflowSystem(s.World())).
		Add(camera.NewCameraScrollSystem(cfg.CameraSpeed))

	return nil
}

// DrawOverlays draws the toolbar, hover/selection borders, placement ghost,
// and the debug console.
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
