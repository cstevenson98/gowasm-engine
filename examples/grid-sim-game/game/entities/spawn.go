package entities

import (
	"example.com/grid-sim-game/game/gameconfig"
	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// cellPosition returns a cell's top-left world position in pixels.
func cellPosition(cell GridCoord) types.Vector2 {
	ts := gameconfig.Global.TileSize
	return types.Vector2{X: float64(cell.Col) * ts, Y: float64(cell.Row) * ts}
}

// spawnTile builds a single-cell entity carrying Position, Sprite, a layer
// tag, Order and GridObject. Shared by SpawnBlank/SpawnGenerator/SpawnHouse/
// SpawnLineSegment; only texture/kind/layer/z differ between them.
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

// SpawnBlank spawns one empty grid-cell tile at cell, on the BACKGROUND layer
// so placed entities always draw on top of it. Called once per cell when the
// grid is built.
func SpawnBlank(w *ecs.World, cell GridCoord) ecs.Entity {
	return spawnTile(w, cell, ToolNone, gameconfig.Global.BlankTexture, 0, true)
}

// SpawnGenerator spawns a generator tile at cell, on the ENTITIES layer.
func SpawnGenerator(w *ecs.World, cell GridCoord) ecs.Entity {
	return spawnTile(w, cell, ToolGenerator, gameconfig.Global.GeneratorTexture, 0, false)
}

// SpawnHouse spawns a house tile at cell, on the ENTITIES layer.
func SpawnHouse(w *ecs.World, cell GridCoord) ecs.Entity {
	return spawnTile(w, cell, ToolHouse, gameconfig.Global.HouseTexture, 0, false)
}

// SpawnLineSegment spawns one tile of a line's path at cell, on the ENTITIES
// layer. A line is represented as one entity per cell along its path (see
// ManhattanPath); there is no single entity for "the line" as a whole.
func SpawnLineSegment(w *ecs.World, cell GridCoord) ecs.Entity {
	return spawnTile(w, cell, ToolLine, gameconfig.Global.LineTexture, 0, false)
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
			// The (to.Col, from.Row) corner was already added by the loop
			// above; skip it here to avoid a duplicate cell.
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
