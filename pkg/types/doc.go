// Package types defines the shared vocabulary of the engine: the core value
// types and the interfaces that every other package agrees on.
//
// It sits at the bottom of the dependency graph. Because concrete components
// (canvas, sprite, mover, scene, ...) depend on types rather than on each
// other, they can be swapped, mocked, and unit-tested in isolation. Nothing in
// this package imports another engine package, which keeps the dependency
// graph acyclic.
//
// # The component model
//
// A game is a tree of GameObjects grouped into Scenes and driven by the
// engine's game loop. Each GameObject is assembled from small, single-purpose
// parts described here:
//
//   - Sprite  - the visual appearance (texture, animation frame, size).
//   - Mover   - the spatial behaviour (position, velocity, screen wrapping).
//   - State   - identity and bookkeeping (ObjectState: ID, position, frame).
//
// A GameObject combines its Sprite and Mover each frame into a
// SpriteRenderData payload, which is the only thing the renderer needs to draw
// it. This separation means "how a thing looks", "how a thing moves", and "how
// a thing is drawn" are three independent concerns.
//
// # Value types
//
// Vector2 (a 2D point/size/velocity) and UVRect (a normalised sub-rectangle of
// a texture) are the small, copyable building blocks used throughout the
// engine's public APIs.
//
// # Optional scene interfaces
//
// Scenes declare the engine services they need by implementing narrow,
// optional interfaces (SceneAssetProvider, SceneInjectable, and friends). The
// engine feature-detects these with type assertions during scene setup, so a
// scene only pays for what it uses. See scene_extras.go.
//
// # Scope
//
// This package holds only engine-generic vocabulary. Domain-specific systems
// own their own types: the combat vocabulary (BattleEntity, Action, ActionTimer)
// lives in pkg/systems/battle, not here. The input snapshot (InputState) does
// live here, since input is a core engine concern.
package types
