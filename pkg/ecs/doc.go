// Package ecs is the engine's Entity Component System abstraction and the sole
// backend seam: it is the ONLY package in this module permitted to import a
// concrete ECS library (currently github.com/mlange-42/ark). All engine and
// game code depends on this package, never on Ark directly, so the backend can
// be swapped by rewriting this package alone.
//
// # Model
//
//   - Entity is an opaque handle. Store it by value; check World.Alive for
//     handles kept across frames.
//   - Components are plain data structs. They carry no behaviour.
//   - Systems (see System / Schedule) contain the behaviour and run each frame.
//   - Resources are per-World singletons (input, screen size, UI, ...).
//
// # Usage
//
// Create mappers and filters once (e.g. in a system's constructor) and reuse
// them; both are comparatively costly to build. Create a fresh iteration for
// every frame via the Each methods.
//
//	type MoveSystem struct{ f *ecs.Filter2[Position, Velocity] }
//	func NewMoveSystem(w *ecs.World) *MoveSystem {
//	    return &MoveSystem{f: ecs.NewFilter2[Position, Velocity](w)}
//	}
//	func (s *MoveSystem) Update(w *ecs.World, dt float64) {
//	    s.f.Each(func(_ ecs.Entity, p *Position, v *Velocity) {
//	        p.X += v.DX * dt
//	        p.Y += v.DY * dt
//	    })
//	}
//
// Component pointers obtained inside an Each callback are only valid for the
// duration of that callback; never store them.
package ecs
