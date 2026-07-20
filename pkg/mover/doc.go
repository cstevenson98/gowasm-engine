// Package mover provides the spatial half of a game object: where it is and how
// it moves, independent of what it looks like.
//
// A Mover (see types.Mover) owns a position and velocity and advances the
// position every frame based on elapsed time. Because movement is a separate
// component from the sprite, the same visual can be given completely different
// motion behaviour just by swapping its mover.
//
// # BasicMover
//
// BasicMover is the standard implementation. It applies velocity over
// deltaTime and performs screen wrapping: when an object travels off one edge
// of the screen it reappears on the opposite edge, which is handy for endlessly
// scrolling or looping entities. It needs the sprite's dimensions and the
// screen bounds (SetScreenBounds) to know exactly when to wrap.
//
// For objects that should never move (backgrounds, battle enemies), see the
// StaticMover in package gameobject, which satisfies the same interface but
// ignores velocity.
package mover
