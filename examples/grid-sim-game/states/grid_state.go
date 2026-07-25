package states

import (
	"fmt"
	"math"
	"sort"

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

	// Placement / load-tick mutate topology or house P/Q (mark Dirty);
	// LoadflowSystem re-solves only when Dirty; camera scrolls last.
	s.Schedule().
		Add(placement.NewPlacementSystem(s.World())).
		Add(loadtick.NewLoadTickSystem(s.World())).
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

	s.renderBusHistoryCharts(w, net)
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

const perBusPlotHeight = 220.0

// busHist holds one bus entity's solve history in kW / kvar / volts.
type busHist struct {
	id     network.BusID
	pKW    []float64
	qKVAR  []float64
	vVolts []float64
}

// collectBusHistories returns BusHistory series for buses of the given type,
// sorted by bus ID. If consumerSign is true, P/Q are negated (load demand);
// otherwise they stay in generator-convention kW (positive = injection).
func (s *GridState) collectBusHistories(net *network.ElectricalNetwork, typ network.BusType, consumerSign bool) []busHist {
	busHistMap := ecs.NewMap1[network.BusHistory](s.World())
	ids := make([]int, 0)
	for id, b := range net.Buses() {
		if b.Type != typ {
			continue
		}
		ids = append(ids, int(id))
	}
	sort.Ints(ids)

	out := make([]busHist, 0, len(ids))
	for _, raw := range ids {
		id := network.BusID(raw)
		b := net.Buses()[id]
		h := busHistMap.Get(b.Entity)
		if h == nil || h.P.Len() == 0 {
			continue
		}
		p := h.P.Values()
		q := h.Q.Values()
		v := h.V.Values()
		pKW := make([]float64, len(p))
		qKVAR := make([]float64, len(q))
		sign := 1.0
		if consumerSign {
			sign = -1.0
		}
		for i := range p {
			pKW[i] = sign * p[i] / 1000
			if i < len(q) {
				qKVAR[i] = sign * q[i] / 1000
			}
		}
		out = append(out, busHist{id: id, pKW: pKW, qKVAR: qKVAR, vVolts: v})
	}
	return out
}

func (s *GridState) renderBusHistoryCharts(w *imgui.WindowBuilder, net *network.ElectricalNetwork) {
	gens := s.collectBusHistories(net, network.BusGenerator, false)
	houses := s.collectBusHistories(net, network.BusLoad, true)

	w.Text("Generators")
	if len(gens) == 0 {
		w.Text("  (none with history yet)")
	} else {
		for _, h := range gens {
			s.renderOneBusHistory(w, "Gen", h, "kW / kvar (+gen)")
		}
	}
	w.Separator()

	w.Text("Houses")
	if len(houses) == 0 {
		w.Text("  (none with history yet)")
	} else {
		for _, h := range houses {
			s.renderOneBusHistory(w, "House", h, "kW / kvar (demand)")
		}
	}
}

// renderOneBusHistory draws one bus's P/Q (left) and |V| (right) plots.
// Plot titles include the bus id so ImPlot keys stay unique across the panel.
func (s *GridState) renderOneBusHistory(w *imgui.WindowBuilder, kind string, h busHist, pqAxis string) {
	w.Text("%s bus %d  (%d samples)", kind, h.id, len(h.pKW))
	w.Columns(2)
	w.Plot(fmt.Sprintf("%s%d P/Q", kind, h.id), perBusPlotHeight, func(p *imgui.PlotBuilder) {
		p.SetupAxes("solve #", pqAxis)
		p.Line("P", h.pKW)
		p.Line("Q", h.qKVAR)
	})
	w.NextColumn()
	w.Plot(fmt.Sprintf("%s%d |V|", kind, h.id), perBusPlotHeight, func(p *imgui.PlotBuilder) {
		p.SetupAxes("solve #", "V")
		p.Line("|V|", h.vVolts)
	})
	w.Columns(1)
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
