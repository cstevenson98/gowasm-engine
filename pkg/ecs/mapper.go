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

// Map5 accesses 5 components.
type Map5[A, B, C, D, E any] struct{ inner *arkecs.Map5[A, B, C, D, E] }

func NewMap5[A, B, C, D, E any](w *World) *Map5[A, B, C, D, E] {
	return &Map5[A, B, C, D, E]{inner: arkecs.NewMap5[A, B, C, D, E](w.inner)}
}

func (m *Map5[A, B, C, D, E]) NewEntity(a *A, b *B, c *C, d *D, e *E) Entity {
	return Entity(m.inner.NewEntity(a, b, c, d, e))
}
func (m *Map5[A, B, C, D, E]) Get(e Entity) (*A, *B, *C, *D, *E) {
	return m.inner.Get(arkecs.Entity(e))
}
func (m *Map5[A, B, C, D, E]) Has(e Entity) bool { return m.inner.HasAll(arkecs.Entity(e)) }
func (m *Map5[A, B, C, D, E]) Remove(e Entity)   { m.inner.Remove(arkecs.Entity(e)) }

// Map6 accesses 6 components.
type Map6[A, B, C, D, E, F any] struct {
	inner *arkecs.Map6[A, B, C, D, E, F]
}

func NewMap6[A, B, C, D, E, F any](w *World) *Map6[A, B, C, D, E, F] {
	return &Map6[A, B, C, D, E, F]{inner: arkecs.NewMap6[A, B, C, D, E, F](w.inner)}
}

func (m *Map6[A, B, C, D, E, F]) NewEntity(a *A, b *B, c *C, d *D, e *E, f *F) Entity {
	return Entity(m.inner.NewEntity(a, b, c, d, e, f))
}
func (m *Map6[A, B, C, D, E, F]) Get(e Entity) (*A, *B, *C, *D, *E, *F) {
	return m.inner.Get(arkecs.Entity(e))
}
func (m *Map6[A, B, C, D, E, F]) Has(e Entity) bool { return m.inner.HasAll(arkecs.Entity(e)) }
func (m *Map6[A, B, C, D, E, F]) Remove(e Entity)   { m.inner.Remove(arkecs.Entity(e)) }

// Map7 accesses 7 components.
type Map7[A, B, C, D, E, F, G any] struct {
	inner *arkecs.Map7[A, B, C, D, E, F, G]
}

func NewMap7[A, B, C, D, E, F, G any](w *World) *Map7[A, B, C, D, E, F, G] {
	return &Map7[A, B, C, D, E, F, G]{inner: arkecs.NewMap7[A, B, C, D, E, F, G](w.inner)}
}

func (m *Map7[A, B, C, D, E, F, G]) NewEntity(a *A, b *B, c *C, d *D, e *E, f *F, g *G) Entity {
	return Entity(m.inner.NewEntity(a, b, c, d, e, f, g))
}
func (m *Map7[A, B, C, D, E, F, G]) Get(e Entity) (*A, *B, *C, *D, *E, *F, *G) {
	return m.inner.Get(arkecs.Entity(e))
}
func (m *Map7[A, B, C, D, E, F, G]) Has(e Entity) bool { return m.inner.HasAll(arkecs.Entity(e)) }
func (m *Map7[A, B, C, D, E, F, G]) Remove(e Entity)   { m.inner.Remove(arkecs.Entity(e)) }

// Map8 accesses 8 components.
type Map8[A, B, C, D, E, F, G, H any] struct {
	inner *arkecs.Map8[A, B, C, D, E, F, G, H]
}

func NewMap8[A, B, C, D, E, F, G, H any](w *World) *Map8[A, B, C, D, E, F, G, H] {
	return &Map8[A, B, C, D, E, F, G, H]{inner: arkecs.NewMap8[A, B, C, D, E, F, G, H](w.inner)}
}

func (m *Map8[A, B, C, D, E, F, G, H]) NewEntity(a *A, b *B, c *C, d *D, e *E, f *F, g *G, h *H) Entity {
	return Entity(m.inner.NewEntity(a, b, c, d, e, f, g, h))
}
func (m *Map8[A, B, C, D, E, F, G, H]) Get(e Entity) (*A, *B, *C, *D, *E, *F, *G, *H) {
	return m.inner.Get(arkecs.Entity(e))
}
func (m *Map8[A, B, C, D, E, F, G, H]) Has(e Entity) bool { return m.inner.HasAll(arkecs.Entity(e)) }
func (m *Map8[A, B, C, D, E, F, G, H]) Remove(e Entity)   { m.inner.Remove(arkecs.Entity(e)) }
