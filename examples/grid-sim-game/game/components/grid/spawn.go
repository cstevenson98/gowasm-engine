package grid

import (
	"math/rand"

	"example.com/grid-sim-game/game/gameconfig"
	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

const (
	HouseLoadMinKW = 1.5
	HouseLoadMaxKW = 3.0
)

// RandLoadKW returns a uniform random value in [HouseLoadMinKW, HouseLoadMaxKW]
// for house demand (kW or kVAR). Used at spawn and by the load-tick system.
func RandLoadKW() float64 {
	return HouseLoadMinKW + (HouseLoadMaxKW-HouseLoadMinKW)*rand.Float64()
}

// cellPosition returns a cell's top-left world position in pixels.
func cellPosition(cell GridCoord) types.Vector2 {
	ts := gameconfig.Global.TileSize
	return types.Vector2{X: float64(cell.Col) * ts, Y: float64(cell.Row) * ts}
}

// spawnTile builds a single-cell entity carrying Position, Sprite, a layer
// tag, Order and GridObject. Shared by SpawnGenerator/SpawnHouse/SpawnLine;
// only texture/kind/layer/z differ between them.
func spawnTile(w *ecs.World, cell GridCoord, kind Tool, texturePath string, z int, background bool) ecs.Entity {
	ts := gameconfig.Global.TileSize
	pos := cellPosition(cell)
	size := types.Vector2{X: ts, Y: ts}

	sprite := &components.Sprite{
		TexturePath: texturePath,
		Size:        size,
		Columns:     1,
		Rows:        1,
		Visible:     true,
	}
	order := &components.Order{Z: z}

	var e ecs.Entity
	if background {
		m := ecs.NewMap4[components.Position, components.Sprite, components.LayerBackground, components.Order](w)
		e = m.NewEntity(&components.Position{X: pos.X, Y: pos.Y}, sprite, &components.LayerBackground{}, order)
	} else {
		m := ecs.NewMap4[components.Position, components.Sprite, components.LayerEntities, components.Order](w)
		e = m.NewEntity(&components.Position{X: pos.X, Y: pos.Y}, sprite, &components.LayerEntities{}, order)
	}

	ecs.NewMap1[GridObject](w).Add(e, &GridObject{Kind: kind, Cell: cell})
	return e
}

// SpawnGenerator spawns a generator tile at cell, on the ENTITIES layer.
// It also attaches a GeneratorProps component with a default 100 kW capacity.
func SpawnGenerator(w *ecs.World, cell GridCoord) ecs.Entity {
	e := spawnTile(w, cell, ToolGenerator, gameconfig.Global.GeneratorTexture, 0, false)
	ecs.NewMap1[GeneratorProps](w).Add(e, &GeneratorProps{MaxOutputKW: 100.0})
	return e
}

// SpawnHouse spawns a house tile at cell, on the ENTITIES layer.
// It attaches a HouseLoad component with P and Q sampled uniformly from
// [HouseLoadMinKW, HouseLoadMaxKW].
func SpawnHouse(w *ecs.World, cell GridCoord) ecs.Entity {
	e := spawnTile(w, cell, ToolHouse, gameconfig.Global.HouseTexture, 0, false)
	ecs.NewMap1[HouseLoad](w).Add(e, &HouseLoad{PKw: RandLoadKW(), QKw: RandLoadKW()})
	return e
}

// SpawnLine spawns one polyline line entity covering path. The entity's
// Position/Sprite sit on path[0]; remaining cells are drawn by grid overlays.
// LineSegmentProps R/X scale with max(1, len(path)-1) cell lengths.
// Caller must Occupy every cell in path with the returned entity.
func SpawnLine(w *ecs.World, path []GridCoord) ecs.Entity {
	if len(path) == 0 {
		return ecs.Entity{}
	}
	cells := append([]GridCoord(nil), path...)
	e := spawnTile(w, cells[0], ToolLine, gameconfig.Global.LineTexture, 0, false)
	hops := len(cells) - 1
	if hops < 1 {
		hops = 1
	}
	ecs.NewMap1[LineSegmentProps](w).Add(e, &LineSegmentProps{
		ResistanceOhm: DefaultLineResistanceOhm * float64(hops),
		ReactanceOhm:  DefaultLineReactanceOhm * float64(hops),
	})
	ecs.NewMap1[LinePath](w).Add(e, &LinePath{Cells: cells})
	return e
}

// SpawnLineSegment spawns a single-cell line (convenience for tests).
func SpawnLineSegment(w *ecs.World, cell GridCoord) ecs.Entity {
	return SpawnLine(w, []GridCoord{cell})
}

// ManhattanPath returns the cells of an L-shaped path from `from` to `to`:
// first horizontally along from.Row, then vertically along to.Col. Both
// endpoints are included exactly once (the corner cell is not duplicated).
// Used to fill in a line's tiles between the two clicks that define it.
func ManhattanPath(from, to GridCoord) []GridCoord {
	var path []GridCoord

	colStep := 1
	if to.Col < from.Col {
		colStep = -1
	}
	for c := from.Col; ; c += colStep {
		path = append(path, GridCoord{Col: c, Row: from.Row})
		if c == to.Col {
			break
		}
	}

	rowStep := 1
	if to.Row < from.Row {
		rowStep = -1
	}
	for r := from.Row; ; r += rowStep {
		if r == from.Row {
			if r == to.Row {
				break
			}
			continue
		}
		path = append(path, GridCoord{Col: to.Col, Row: r})
		if r == to.Row {
			break
		}
	}

	return path
}
