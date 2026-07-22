package components

// Layer tag components select which render pass draws an entity. They are
// zero-size marker components; the render system runs one filtered pass per
// layer, in the order Background -> Entities -> UI. This replaces the fixed
// three-slice SceneLayer buckets.
type (
	LayerBackground struct{}
	LayerEntities   struct{}
	LayerUI         struct{}
)

// Order is the intra-layer draw-order tiebreak. Within a single layer pass,
// entities are drawn from lowest Z to highest. It exists because ECS iteration
// follows archetype order, not insertion order, so a stable key is needed to
// keep overlapping sprites from flickering. Every renderable carries an Order
// (default Z = 0).
type Order struct{ Z int }
