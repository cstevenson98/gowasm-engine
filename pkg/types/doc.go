// Package types defines the shared vocabulary of the engine: the core value
// types and the small interfaces that other packages agree on.
//
// It sits at the bottom of the dependency graph. Other packages depend on types
// rather than on each other, which keeps the dependency graph acyclic and lets
// components be swapped, mocked, and unit-tested in isolation. Nothing in this
// package imports another engine package.
//
// # Scope
//
// This package holds only engine-generic vocabulary that does not belong to a
// single subsystem:
//
//   - Vector2 - a 2D point/size/velocity, and UVRect - a normalised
//     sub-rectangle of a texture. These are the small, copyable building blocks
//     used throughout the engine's public APIs.
//   - InputState - the per-frame input snapshot, and InputCapturer - the
//     interface the engine polls each frame.
//   - GameState - the enum used to key registered states.
//   - UIManager - the immediate-mode overlay drawing facade.
//
// The ECS building blocks live elsewhere: pure-data components and resources in
// [github.com/cstevenson98/milo/pkg/components], the World/entity/query
// abstraction in [github.com/cstevenson98/milo/pkg/ecs], and the State
// model in [github.com/cstevenson98/milo/pkg/state]. Domain-specific
// vocabulary (e.g. combat: BattleEntity, Action, ActionTimer) lives with its
// system, in pkg/systems/battle.
package types
