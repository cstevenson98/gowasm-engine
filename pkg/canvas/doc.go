// Package canvas is the engine's rendering backend: it turns high-level draw
// requests into pixels on screen.
//
// The game-facing contract is the small CanvasManager interface. Callers work
// entirely in terms of texture paths and rectangles - "draw this region of
// this texture at this position and size" - and never touch the underlying
// graphics API. This keeps rendering concerns (render targets, texture
// caching, filtering) out of gameplay code and makes the whole engine testable
// against MockCanvasManager.
//
// # Responsibilities
//
//   - Load and cache textures by file path (LoadTexture, with lazy loading on
//     first draw so dynamically spawned objects still render).
//   - Draw a UV-selected region of a texture (DrawTexturedRect) - the workhorse
//     used for sprites, sprite-sheet frames, and text glyphs.
//   - Draw solid colour rectangles (DrawColoredRect) for UI and debug overlays.
//
// # Implementation
//
// Canvas is the concrete implementation built on Ebiten. Ebiten owns the
// window and the render target (an *ebiten.Image handed to the canvas each
// frame via SetScreen) and coalesces draw calls that share a source image, so
// no manual batching layer is needed. Pixel-art versus smooth filtering is
// chosen from config. MockCanvasManager mirrors the interface for tests that
// must run without a GPU or window.
package canvas
