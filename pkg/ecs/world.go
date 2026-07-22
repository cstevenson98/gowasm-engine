package ecs

import arkecs "github.com/mlange-42/ark/ecs"

// Entity is an opaque handle to an entity. It is a comparable value type and
// may be used as a map key or stored in components/resources by value.
type Entity arkecs.Entity

// ID returns the entity's stable identifier (excluding generation).
func (e Entity) ID() uint32 { return arkecs.Entity(e).ID() }

// World owns all entities, components and resources. Each State owns one World
// and drops it on exit, giving clean teardown between states.
type World struct {
	inner *arkecs.World
}

// NewWorld creates an empty World. An optional initial entity capacity may be
// supplied as a sizing hint.
func NewWorld(initialCapacity ...int) *World {
	return &World{inner: arkecs.NewWorld(initialCapacity...)}
}

// Remove deletes an entity and all of its components from the world.
func (w *World) Remove(e Entity) { w.inner.RemoveEntity(arkecs.Entity(e)) }

// Alive reports whether the entity handle still refers to a live entity.
func (w *World) Alive(e Entity) bool { return w.inner.Alive(arkecs.Entity(e)) }

// Reset clears all entities, components and resources, retaining allocations.
func (w *World) Reset() { w.inner.Reset() }

// ===== Resources (per-World singletons) =====

// SetResource stores (or replaces) the singleton resource of type T.
func SetResource[T any](w *World, res *T) {
	r := arkecs.NewResource[T](w.inner)
	if r.Has() {
		r.Remove()
	}
	r.Add(res)
}

// GetResource returns the singleton resource of type T, or nil if unset.
func GetResource[T any](w *World) *T {
	r := arkecs.NewResource[T](w.inner)
	if !r.Has() {
		return nil
	}
	return r.Get()
}

// HasResource reports whether a resource of type T is present.
func HasResource[T any](w *World) bool {
	r := arkecs.NewResource[T](w.inner)
	return r.Has()
}

// RemoveResource deletes the resource of type T if present.
func RemoveResource[T any](w *World) {
	r := arkecs.NewResource[T](w.inner)
	if r.Has() {
		r.Remove()
	}
}
