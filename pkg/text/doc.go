// Package text renders strings using bitmap font sprite sheets.
//
// The engine has no dependency on system fonts or a glyph rasteriser. Instead,
// a font is a pre-rendered sprite sheet plus JSON metadata describing where
// each character sits within it. Drawing text is then just drawing a sequence
// of textured rectangles through the canvas - the same mechanism used for
// every other sprite. This keeps text rendering fast, pixel-perfect, and
// consistent with the engine's art style.
//
// # Fonts
//
// The Font interface abstracts "given a character, where is its glyph". SpriteFont
// is the concrete implementation: LoadFont reads a "<name>.sheet.png" texture
// and a "<name>.sheet.json" metadata file (produced by the generator in
// scripts/) and maps each rune to a UV rectangle. Loaded font metadata is
// cached globally so repeated loads of the same font are cheap.
//
// # Rendering
//
// The TextRenderer interface draws a string at a position, laying out glyphs
// left to right and handling newlines, spacing, and scaling. BasicTextRenderer
// implements it on top of a canvas.CanvasManager. This is what powers the
// debug console and in-game menus.
package text
