package ecs

import arkecs "github.com/mlange-42/ark/ecs"

// Mappers create entities and access components outside of query loops.
// Create them once and reuse; they are comparatively costly to construct.

// Map1 accesses 1 component.
type Map1[A any] struct{ inner *arkecs.Map1[A] }

func NewMap1[A any](w *World) *Map1[A] { return &Map1[A]{inner: arkecs.NewMap1[A](w.inner)} }

func (m *Map1[A]) NewEntity(a *A) Entity { return Entity(m.inner.NewEntity(a)) }
func (m *Map1[A]) Get(e Entity) *A       { return m.inner.Get(arkecs.Entity(e)) }
func (m *Map1[A]) Has(e Entity) bool     { return m.inner.HasAll(arkecs.Entity(e)) }
func (m *Map1[A]) Add(e Entity, a *A)    { m.inner.Add(arkecs.Entity(e), a) }
func (m *Map1[A]) Remove(e Entity)       { m.inner.Remove(arkecs.Entity(e)) }

// Map2 accesses 2 components.
type Map2[A, B any] struct{ inner *arkecs.Map2[A, B] }

func NewMap2[A, B any](w *World) *Map2[A, B] {
	return &Map2[A, B]{inner: arkecs.NewMap2[A, B](w.inner)}
}

func (m *Map2[A, B]) NewEntity(a *A, b *B) Entity { return Entity(m.inner.NewEntity(a, b)) }
func (m *Map2[A, B]) Get(e Entity) (*A, *B)       { return m.inner.Get(arkecs.Entity(e)) }
func (m *Map2[A, B]) Has(e Entity) bool           { return m.inner.HasAll(arkecs.Entity(e)) }
func (m *Map2[A, B]) Add(e Entity, a *A, b *B)    { m.inner.Add(arkecs.Entity(e), a, b) }
func (m *Map2[A, B]) Remove(e Entity)             { m.inner.Remove(arkecs.Entity(e)) }

// Map3 accesses 3 components.
type Map3[A, B, C any] struct{ inner *arkecs.Map3[A, B, C] }

func NewMap3[A, B, C any](w *World) *Map3[A, B, C] {
	return &Map3[A, B, C]{inner: arkecs.NewMap3[A, B, C](w.inner)}
}

func (m *Map3[A, B, C]) NewEntity(a *A, b *B, c *C) Entity {
	return Entity(m.inner.NewEntity(a, b, c))
}
func (m *Map3[A, B, C]) Get(e Entity) (*A, *B, *C)      { return m.inner.Get(arkecs.Entity(e)) }
func (m *Map3[A, B, C]) Has(e Entity) bool              { return m.inner.HasAll(arkecs.Entity(e)) }
func (m *Map3[A, B, C]) Add(e Entity, a *A, b *B, c *C) { m.inner.Add(arkecs.Entity(e), a, b, c) }
func (m *Map3[A, B, C]) Remove(e Entity)                { m.inner.Remove(arkecs.Entity(e)) }

// Map4 accesses 4 components.
type Map4[A, B, C, D any] struct{ inner *arkecs.Map4[A, B, C, D] }

func NewMap4[A, B, C, D any](w *World) *Map4[A, B, C, D] {
	return &Map4[A, B, C, D]{inner: arkecs.NewMap4[A, B, C, D](w.inner)}
}

func (m *Map4[A, B, C, D]) NewEntity(a *A, b *B, c *C, d *D) Entity {
	return Entity(m.inner.NewEntity(a, b, c, d))
}
func (m *Map4[A, B, C, D]) Get(e Entity) (*A, *B, *C, *D) { return m.inner.Get(arkecs.Entity(e)) }
func (m *Map4[A, B, C, D]) Has(e Entity) bool             { return m.inner.HasAll(arkecs.Entity(e)) }
func (m *Map4[A, B, C, D]) Add(e Entity, a *A, b *B, c *C, d *D) {
	m.inner.Add(arkecs.Entity(e), a, b, c, d)
}
func (m *Map4[A, B, C, D]) Remove(e Entity) { m.inner.Remove(arkecs.Entity(e)) }
