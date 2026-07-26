// Package prefab provides builder functions that spawn ready-made entities into
// an ecs.World. They are the ECS-era replacement for the gameobject types
// (Background, Llama): instead of constructing an object, you spawn an entity
// with the right component set and get its handle back.
package prefab

import (
	"github.com/cstevenson98/milo/pkg/components"
	"github.com/cstevenson98/milo/pkg/ecs"
	"github.com/cstevenson98/milo/pkg/types"
)

// NewBackground spawns a static, full-image background entity on the BACKGROUND
// layer and returns its handle. Ports gameobject.Background.
func NewBackground(w *ecs.World, position, size types.Vector2, texturePath string) ecs.Entity {
	m := ecs.NewMap4[components.Position, components.Sprite, components.LayerBackground, components.Order](w)
	return m.NewEntity(
		&components.Position{X: position.X, Y: position.Y},
		&components.Sprite{
			TexturePath: texturePath,
			Size:        size,
			Columns:     1,
			Rows:        1,
			Visible:     true,
		},
		&components.LayerBackground{},
		&components.Order{Z: 0},
	)
}

// NewLlama spawns an animated, screen-wrapping llama entity on the ENTITIES
// layer moving right at the given speed, and returns its handle. Ports
// gameobject.Llama (2x3 sheet = 6 frames). defaultFrameTime is the engine's
// fallback seconds-per-frame (config.Settings.Animation.DefaultFrameTime); the
// caller supplies it explicitly so this package has no config dependency.
func NewLlama(w *ecs.World, position, size types.Vector2, speed, defaultFrameTime float64) ecs.Entity {
	m := ecs.NewMap7[
		components.Position,
		components.Velocity,
		components.Wrap,
		components.Sprite,
		components.Animation,
		components.LayerEntities,
		components.Order,
	](w)

	frameTime := defaultFrameTime + (speed/100.0)*0.1

	return m.NewEntity(
		&components.Position{X: position.X, Y: position.Y},
		&components.Velocity{DX: speed, DY: 0},
		&components.Wrap{SpriteW: size.X, SpriteH: size.Y},
		&components.Sprite{
			TexturePath: "llama.png",
			Size:        size,
			Columns:     2,
			Rows:        3,
			Visible:     true,
		},
		&components.Animation{FrameTime: frameTime},
		&components.LayerEntities{},
		&components.Order{Z: 0},
	)
}
