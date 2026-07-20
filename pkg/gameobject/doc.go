// Package gameobject provides the entities that live inside a scene: players,
// enemies, items, backgrounds, and anything else the game draws or updates.
//
// # Composition, not inheritance
//
// A game object is deliberately thin. It is identity and behaviour, and it
// delegates the two things that vary most to separate components:
//
//   - a Sprite decides how it looks (texture, animation frame, size), and
//   - a Mover decides where it is (position, velocity, screen wrapping).
//
// This keeps objects small and lets you mix and match: a static background is a
// sprite with a stationary mover; a walking enemy is the same sprite with a
// velocity-driven mover.
//
// # BaseGameObject
//
// [BaseGameObject] wires a Sprite and a Mover together and implements the
// engine's GameObject interface, including GetRenderData, which assembles the
// per-frame render packet (texture path, position, size, UV, visibility) that
// the engine hands to the canvas. Concrete objects embed it:
//
//	type Coin struct {
//		*gameobject.BaseGameObject
//	}
//
//	func NewCoin(pos types.Vector2) *Coin {
//		spr := sprite.NewSpriteSheet("assets/art/coin.png", sprite.Vector2{X: 16, Y: 16}, 4, 1)
//		mv := mover.NewBasicMover(pos, types.Vector2{}, 16, 16)
//		state := types.ObjectState{ID: "coin", Position: pos, Visible: true}
//		return &Coin{BaseGameObject: gameobject.NewBaseGameObject(spr, mv, state)}
//	}
//
// Because BaseGameObject already satisfies the interface, a type like Coin is
// immediately usable with scene.AddEntity and friends. Override Update only when
// the object needs custom behaviour (for example reading input or animating on a
// condition); the embedded Update advances the sprite and mover for you.
//
// # Ready-made objects
//
// [Player] and [Llama] are complete examples: Player reads the injected input
// each frame to move, while Llama drifts across the screen and wraps around,
// demonstrating the sprite + mover split with no custom input.
package gameobject
