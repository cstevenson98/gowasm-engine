package systems

import (
	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
)

// Animation advances Sprite.Frame for entities that have both a Sprite and an
// Animation component. Sprites without an Animation are static. Replaces
// SpriteSheet.Update.
type Animation struct {
	f *ecs.Filter2[components.Sprite, components.Animation]
}

// NewAnimation builds the Animation system for world w.
func NewAnimation(w *ecs.World) *Animation {
	return &Animation{
		f: ecs.NewFilter2[components.Sprite, components.Animation](w),
	}
}

// Update accumulates elapsed time and cycles frames.
func (s *Animation) Update(w *ecs.World, dt float64) {
	s.f.Each(func(_ ecs.Entity, sp *components.Sprite, an *components.Animation) {
		total := sp.TotalFrames()
		if total <= 1 || an.FrameTime <= 0 {
			return
		}
		an.Elapsed += dt
		for an.Elapsed >= an.FrameTime {
			an.Elapsed -= an.FrameTime
			sp.Frame = (sp.Frame + 1) % total
		}
	})
}
