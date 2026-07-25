// Package entities holds the grid-sim-game's ECS components, resources,
// spawners and systems: everything specific to placing generators, houses and
// lines on a grid.
package entities

import "github.com/cstevenson98/gowasm-engine/pkg/ecs"

// Tool identifies which kind of object the player currently has selected to
// place (or ToolNone, meaning clicks on the grid do nothing).
type Tool int

const (
	ToolNone Tool = iota
	ToolGenerator
	ToolHouse
	ToolLine
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
	default:
		return "NONE"
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
// PlacementSystem and read by GridState.DrawOverlays to render the toolbar
// highlight/hint, so it is the single source of truth for "what is the
// player doing right now".
type PlacementState struct {
	Tool        Tool
	LinePending bool
	LineStart   GridCoord
}

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
