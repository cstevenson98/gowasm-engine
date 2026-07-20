// Package input captures user input and exposes it as a simple, per-frame
// snapshot that gameplay code can read without touching device APIs.
//
// The engine polls input once at the start of every Update. The result is a
// types.InputState value describing which actions are active this frame, so
// scenes and game objects react to intent ("move left", "confirm") rather than
// to raw hardware.
//
// # Unified keyboard and gamepad
//
// Input is the concrete implementation, built on Ebiten. Each frame it polls
// the keyboard and the first connected gamepad and merges them into one state:
// WASD and arrow keys map to the same movement flags as the D-pad and left
// stick (with a dead zone), and gamepad face/menu buttons map onto the same
// action flags as their keyboard equivalents. Gameplay code therefore never
// needs to care which device the player is using.
//
// # Edge detection
//
// InputState carries both the current value of each key and its value on the
// previous frame (the *LastFrame fields). Comparing the two lets callers
// distinguish a held key from a fresh press - essential for menu navigation
// and one-shot actions, where a button should fire once rather than every
// frame it is held.
package input
