package entities

import (
	"example.com/basic-game/game/gameconfig"
	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// SpawnPlayer creates the movable, animated player entity on the ENTITIES layer
// and returns its handle. It carries PlayerControl (for input), Stats (for
// save/load), screen wrapping, and CameraTarget so a CameraFollow system keeps
// it centered.
func SpawnPlayer(w *ecs.World, pos, size types.Vector2, speed float64, stats Stats) ecs.Entity {
	m := ecs.NewMap8[
		components.Position,
		components.Velocity,
		components.Wrap,
		components.Sprite,
		components.Animation,
		PlayerControl,
		components.LayerEntities,
		components.Order,
	](w)

	e := m.NewEntity(
		&components.Position{X: pos.X, Y: pos.Y},
		&components.Velocity{},
		&components.Wrap{SpriteW: size.X, SpriteH: size.Y},
		&components.Sprite{
			TexturePath: gameconfig.Global.Player.TexturePath,
			Size:        size,
			Columns:     gameconfig.Global.Player.SpriteColumns,
			Rows:        gameconfig.Global.Player.SpriteRows,
			Visible:     true,
		},
		&components.Animation{FrameTime: gameconfig.Global.PlayerFrameTime},
		&PlayerControl{Speed: speed},
		&components.LayerEntities{},
		&components.Order{Z: 0},
	)

	ecs.NewMap1[Stats](w).Add(e, &stats)
	ecs.NewMap1[components.CameraTarget](w).Add(e, &components.CameraTarget{})
	return e
}

// SpawnCharacter creates a static, animated character sprite entity on the
// ENTITIES layer (used for battle participants, which don't move) and returns
// its handle. frameTime is the seconds-per-frame for its animation; callers
// typically pass the engine's default (see state.BaseState.DefaultFrameTime).
func SpawnCharacter(w *ecs.World, pos, size types.Vector2, texturePath string, columns, rows int, frameTime float64) ecs.Entity {
	m := ecs.NewMap5[
		components.Position,
		components.Sprite,
		components.Animation,
		components.LayerEntities,
		components.Order,
	](w)

	return m.NewEntity(
		&components.Position{X: pos.X, Y: pos.Y},
		&components.Sprite{
			TexturePath: texturePath,
			Size:        size,
			Columns:     columns,
			Rows:        rows,
			Visible:     true,
		},
		&components.Animation{FrameTime: frameTime},
		&components.LayerEntities{},
		&components.Order{Z: 0},
	)
}
