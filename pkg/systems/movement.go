// Package systems provides the engine's built-in ECS systems (movement,
// animation, ...). Systems hold their filters as fields, constructed once
// against a World, and iterate them each frame in Update. Game-specific systems
// live in the game (or in sub-packages like systems/battle).
package systems

import (
	"github.com/cstevenson98/milo/pkg/components"
	"github.com/cstevenson98/milo/pkg/ecs"
)

// Movement integrates Position by Velocity and applies screen-edge wrapping to
// entities that also have a Wrap component, using the ScreenBounds resource.
// Replaces Mover.Update / BasicMover.
type Movement struct {
	move *ecs.Filter2[components.Position, components.Velocity]
	wrap *ecs.Filter3[components.Position, components.Velocity, components.Wrap]
}

// NewMovement builds the Movement system for world w.
func NewMovement(w *ecs.World) *Movement {
	return &Movement{
		move: ecs.NewFilter2[components.Position, components.Velocity](w),
		wrap: ecs.NewFilter3[components.Position, components.Velocity, components.Wrap](w),
	}
}

// Update integrates all moving entities, then wraps those with a Wrap component.
func (s *Movement) Update(w *ecs.World, dt float64) {
	s.move.Each(func(_ ecs.Entity, p *components.Position, v *components.Velocity) {
		p.X += v.DX * dt
		p.Y += v.DY * dt
	})

	bounds := ecs.GetResource[components.ScreenBounds](w)
	if bounds == nil {
		return
	}

	s.wrap.Each(func(_ ecs.Entity, p *components.Position, v *components.Velocity, wr *components.Wrap) {
		// Wrap in the direction of travel, mirroring the original BasicMover:
		// leave the far edge and reappear just off the opposite edge.
		if v.DX > 0 && p.X > bounds.W {
			p.X = -wr.SpriteW
		} else if v.DX < 0 && p.X+wr.SpriteW < 0 {
			p.X = bounds.W
		}
		if v.DY > 0 && p.Y > bounds.H {
			p.Y = -wr.SpriteH
		} else if v.DY < 0 && p.Y+wr.SpriteH < 0 {
			p.Y = bounds.H
		}
	})
}
