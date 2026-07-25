package states

import (
	"math"
	"sort"

	"example.com/grid-sim-game/game/components/grid"
	"example.com/grid-sim-game/game/components/network"
	"example.com/grid-sim-game/game/gameconfig"
	"example.com/grid-sim-game/game/systems/camera"
	"example.com/grid-sim-game/game/systems/loadflow"
	"example.com/grid-sim-game/game/systems/placement"
	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/imgui"
	"github.com/cstevenson98/gowasm-engine/pkg/state"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
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

	// Placement mutates the grid and marks ElectricalNetwork Dirty;
	// LoadflowSystem re-solves only when Dirty; camera scrolls last.
	s.Schedule().
		Add(placement.NewPlacementSystem(s.World())).
		Add(loadflow.NewLoadflowSystem(s.World())).
		Add(camera.NewCameraScrollSystem(cfg.CameraSpeed))

	return nil
}

// DrawOverlays draws the toolbar (with the selected tool highlighted and a
// hint while a line is pending) plus the debug console.
func (s *GridState) DrawOverlays() error {
	s.renderToolbar()
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

func (s *GridState) renderNetworkPanel(w *imgui.WindowBuilder, net *network.ElectricalNetwork) {
	buses := net.Buses()
	branches := net.Branches()
	st := net.State

	w.Text("Topology")
	w.Text("  Buses: %d", len(buses))
	w.Text("  Branches: %d", len(branches))
	w.Text("  Dirty: %v", net.Dirty)
	w.Separator()

	w.Text("Load flow")
	if st == nil {
		w.Text("  (no state)")
		return
	}
	w.Text("  Converged: %v", st.Converged)
	w.Text("  Iterations: %d", st.Iterations)
	w.Separator()

	// Counts by bus type.
	var nGen, nLoad, nJunc int
	for _, b := range buses {
		switch b.Type {
		case network.BusGenerator:
			nGen++
		case network.BusLoad:
			nLoad++
		case network.BusJunction:
			nJunc++
		}
	}
	w.Text("Bus types")
	w.Text("  Generators: %d", nGen)
	w.Text("  Loads: %d", nLoad)
	w.Text("  Junctions: %d", nJunc)
	w.Separator()

	// Totals from solved results.
	var pGen, pLoad, iMax float64
	for id, bs := range st.Buses {
		b, ok := buses[id]
		if !ok {
			continue
		}
		p := bs.Result.PInject
		if b.Type == network.BusGenerator || p > 0 {
			pGen += math.Max(p, 0)
		}
		if b.Type == network.BusLoad || p < 0 {
			pLoad += math.Max(-p, 0)
		}
	}
	for _, br := range st.Branches {
		if br.Result.CurrentMag > iMax {
			iMax = br.Result.CurrentMag
		}
	}
	w.Text("Power (solved)")
	w.Text("  Generation: %.2f kW", pGen/1000)
	w.Text("  Load: %.2f kW", pLoad/1000)
	w.Text("  Peak |I|: %.2f A", iMax)
	w.Separator()

	w.TreeNode("Buses", func(w *imgui.WindowBuilder) {
		ids := make([]int, 0, len(buses))
		for id := range buses {
			ids = append(ids, int(id))
		}
		sort.Ints(ids)
		for _, raw := range ids {
			id := network.BusID(raw)
			b := buses[id]
			bs := st.Buses[id]
			if bs == nil {
				w.Text("bus %d (%s)", b.ID, b.Type)
				continue
			}
			angDeg := bs.Result.VoltAng * 180 / math.Pi
			w.Text("bus %d (%s)  %s", b.ID, b.Type, formulationLabel(bs.Spec.Formulation))
			w.Text("  V=%.1f V ∠ %.2f°  P=%.2f kW  Q=%.2f kvar",
				bs.Result.VoltMag, angDeg, bs.Result.PInject/1000, bs.Result.QInject/1000)
		}
	})

	w.TreeNode("Branches", func(w *imgui.WindowBuilder) {
		ids := make([]int, 0, len(branches))
		for id := range branches {
			ids = append(ids, int(id))
		}
		sort.Ints(ids)
		for _, raw := range ids {
			id := network.BranchID(raw)
			br := branches[id]
			brs := st.Branches[id]
			rOhm := br.Resistance
			if brs == nil {
				w.Text("br %d: %d—%d  R=%.3f Ω", br.ID, br.From, br.To, rOhm)
				continue
			}
			w.Text("br %d: %d—%d  R=%.3f Ω  |I|=%.2f A  P=%.2f kW",
				br.ID, br.From, br.To, rOhm, brs.Result.CurrentMag, brs.Result.PFrom/1000)
		}
	})
}

func formulationLabel(f network.BusFormulation) string {
	switch f {
	case network.Slack:
		return "Slack"
	case network.PV:
		return "PV"
	case network.PQ:
		return "PQ"
	default:
		return "?"
	}
}

func (s *GridState) renderToolbar() {
	cfg := gameconfig.Global
	ui := s.UI()
	playW := cfg.PlayfieldWidth(s.ScreenWidth())

	ui.Rect(0, 0, playW, cfg.ToolbarHeight, types.Color{0.1, 0.1, 0.12, 1})

	placement := ecs.GetResource[grid.PlacementState](s.World())
	selected := grid.ToolNone
	linePending := false
	if placement != nil {
		selected = placement.Tool
		linePending = placement.LinePending
	}

	for _, b := range grid.ToolbarButtons() {
		bg := types.Color{0.3, 0.3, 0.34, 1}
		fg := types.White
		if b.Tool == selected {
			bg = types.Yellow
			fg = types.Black
		}
		ui.Rect(b.X, b.Y, b.W, b.H, bg)
		ui.TextColored(b.X+3, b.Y+2, fg, b.Label)
	}

	if linePending {
		ui.TextColored(cfg.ButtonMarginX, cfg.ToolbarHeight+4, types.Yellow, "Line: click the end cell")
	}
}
