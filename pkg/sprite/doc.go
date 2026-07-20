// Package sprite handles the visual appearance of a game object: which texture
// it uses, how big it is, and which frame of an animation is currently showing.
//
// A sprite is intentionally decoupled from position. It answers "what does this
// look like right now?" while a Mover answers "where is it?"; a
// [github.com/cstevenson98/gowasm-engine/pkg/gameobject.BaseGameObject]
// combines the two into a render packet for the engine.
//
// # Sprite sheets
//
// [SpriteSheet] treats a single texture as a grid of equally sized frames laid
// out in columns and rows. Instead of shipping one file per animation frame,
// you pack them into one image and let the sprite pick a cell:
//
//	// A 2x3 sheet: 2 columns, 3 rows, 6 frames total, each drawn at 64x64.
//	spr := sprite.NewSpriteSheet("assets/art/llama.png", sprite.Vector2{X: 64, Y: 64}, 2, 3)
//	spr.SetFrameTime(0.15) // seconds per frame
//
// # UV coordinates
//
// The engine draws a full texture by default; a sprite narrows that to one
// frame by reporting a UV rectangle (a sub-region of the texture in 0..1
// space). GetUV computes the rectangle for the current frame from the column,
// row, and sheet dimensions, so the engine only ever sees "draw this texture
// with this UV window."
//
// # Animation
//
// Update advances an internal timer and steps to the next frame once
// SetFrameTime worth of time has elapsed, wrapping back to the first frame at
// the end. A single-frame sheet (1x1) is effectively a static image and never
// animates.
//
// # Visibility
//
// SetVisible / IsVisible toggle whether the object is drawn at all. An invisible
// sprite produces a render packet marked not-visible, and the engine skips it.
package sprite
