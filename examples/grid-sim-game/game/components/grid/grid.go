// Package grid holds the ECS components and resources that describe the
// placement grid: cell addresses, placed-object kinds, occupancy tracking,
// and the placement UI state. It is the data layer for the grid game; all
// behaviour lives in the systems packages.
package grid

import "github.com/cstevenson98/gowasm-engine/pkg/ecs"

// Tool identifies which kind of object the player currently has selected to
// place (or ToolNone, meaning clicks select a cell for the inspector).
type Tool int

const (
	ToolNone Tool = iota
	ToolGenerator
	ToolHouse
	ToolLine
	ToolDelete
)

// Label returns the toolbar button label for the tool.
func (t Tool) Label() string {
	switch t {
	case ToolGenerator:
		return "GEN (G)"
	case ToolHouse:
		return "HOUSE (H)"
	case ToolLine:
		return "LINE (L)"
	case ToolDelete:
		return "DEL (X)"
	default:
		return "NONE"
	}
}

// GhostLetter is the single-character overlay drawn inside a placement ghost.
func (t Tool) GhostLetter() string {
	switch t {
	case ToolGenerator:
		return "G"
	case ToolHouse:
		return "H"
	case ToolLine:
		return "L"
	case ToolDelete:
		return "X"
	default:
		return "?"
	}
}

// KindLabel is a short inspector name for a placed GridObject.Kind.
func (t Tool) KindLabel() string {
	switch t {
	case ToolGenerator:
		return "generator"
	case ToolHouse:
		return "house"
	case ToolLine:
		return "line"
	default:
		return "blank"
	}
}

// GridCoord is a cell address on the grid (column, row), independent of the
// grid's pixel size.
type GridCoord struct{ Col, Row int }

// GridObject marks a placed, single-cell entity - a generator, a house, or
// one tile of a line's path - and records which cell it occupies. Systems
// (and the renderer's own Position) key off Cell for occupancy/placement
// logic; Kind is informational (matches which sprite was spawned).
type GridObject struct {
	Kind Tool
	Cell GridCoord
}

// PlacementState is the per-World "global store of user actions" for the
// placement UI: which tool is currently selected, and - for lines, which
// need two clicks - the pending start cell. It is mutated only by
// PlacementSystem and read by GridState.DrawOverlays / ImGui to render the
// toolbar, hover/selection chrome, and placement ghost.
type PlacementState struct {
	Tool        Tool
	LinePending bool
	LineStart   GridCoord

	HoverCell    GridCoord // cell under the cursor (playfield)
	HoverValid   bool      // true when HoverCell is a valid grid cell
	HasSelection bool
	SelectedCell GridCoord
}

// HouseLoad is the power-demand component for a house entity. P and Q are
// in kilowatts / kilovars (kW, kVAR) and are always positive (consumed power).
// They are sampled at spawn and periodically re-sampled by LoadTickSystem.
type HouseLoad struct {
	PKw float64 // active power demand  [kW]
	QKw float64 // reactive power demand [kVAR]
}

// GeneratorProps holds the nameplate parameters of a generator entity.
// MaxOutputKW is the maximum active power the unit can deliver (kW).
type GeneratorProps struct {
	MaxOutputKW float64
}

// LineSegmentProps holds the electrical parameters of one line-segment tile.
// Resistance is in ohms (physical, not per-unit) for the cell length
// (see CellLengthM).
type LineSegmentProps struct {
	ResistanceOhm float64
}

// CellLengthM is the physical length represented by one grid cell / line tile.
const CellLengthM = 10.0 // metres

// CableOhmPerKm is a typical LV distribution feeder impedance
// (≈185 mm² aluminium, ~0.164 Ω/km). At 10 m/cell this supports ~100-bus
// radial feeders at LV without collapsing to unrealistically low voltages
// under a handful of house-scale loads.
const CableOhmPerKm = 0.164

// DefaultLineResistanceOhm is R for one newly-placed line segment:
// CableOhmPerKm × (CellLengthM / 1000).
const DefaultLineResistanceOhm = CableOhmPerKm * CellLengthM / 1000.0 // 0.00164 Ω

// GridOccupancy is the per-World map of occupied cells, enforcing that
// generators, houses and line tiles never overlap each other.
type GridOccupancy struct {
	Cells map[GridCoord]ecs.Entity
}

// NewGridOccupancy creates an empty occupancy map.
func NewGridOccupancy() *GridOccupancy {
	return &GridOccupancy{Cells: make(map[GridCoord]ecs.Entity)}
}

// Occupied reports whether cell already holds a placed entity.
func (g *GridOccupancy) Occupied(cell GridCoord) bool {
	_, ok := g.Cells[cell]
	return ok
}

// Occupy records that e now occupies cell. Callers must check Occupied first.
func (g *GridOccupancy) Occupy(cell GridCoord, e ecs.Entity) {
	g.Cells[cell] = e
}
