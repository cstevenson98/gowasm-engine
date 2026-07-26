// Package ui is a small immediate-mode drawing facade for screen-space UI
// (menus, HUD, labels, bars). A UI is backed by a single canvas and a default
// font, both supplied when it is constructed, and implements types.UIManager.
//
// The engine owns one UI instance and injects it into scenes (as a
// types.UIManager via BaseScene), so scenes draw with s.UI().Text(...) from
// their RenderOverlays without loading a font, constructing a text renderer, or
// holding a reference to the canvas.
//
// Immediate-mode note: these helpers draw straight to the current screen, so
// they must be called during the render phase (a scene's RenderOverlays), not
// from Update. All methods tolerate a nil receiver as a no-op, so scenes that
// render before the UI is available are safe.
package ui

import (
	"github.com/cstevenson98/gowasm-engine/pkg/canvas"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/text"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// Ensure *UI satisfies the engine-facing UI interface.
var _ types.UIManager = (*UI)(nil)

// Config holds the text-layout knobs the UI facade needs. The engine
// constructs this from its own config.Settings and passes it to New, so this
// package has no dependency on the engine's config package.
type Config struct {
	CharacterSpacingReduction float64 // Pixels to reduce character spacing by (used by Measure).
	UILineSpacing             float64 // Line spacing multiplier for UI elements (menus, logs, status).
	TextLineSpacing           float64 // Line spacing multiplier passed through to the embedded text renderer.
}

// UI is an immediate-mode text/primitive drawing facade backed by a canvas and
// a single default font. It implements types.UIManager.
type UI struct {
	canvas   canvas.CanvasManager
	renderer text.TextRenderer
	font     text.Font
	screenW  float64
	screenH  float64
	cfg      Config
}

// New creates a UI that draws on the given canvas using the font sprite sheet at
// fontPath, centering against the given virtual screen dimensions. It loads the
// font and preloads its texture.
func New(cm canvas.CanvasManager, fontPath string, screenW, screenH float64, cfg Config) (*UI, error) {
	font := text.NewSpriteFont()
	if err := font.LoadFont(fontPath); err != nil {
		return nil, err
	}

	// Preload the glyph texture so the first frame of text renders immediately.
	if err := cm.LoadTexture(font.GetTexturePath()); err != nil {
		logger.Logger.Warnf("UI: failed to preload font texture: %s", err.Error())
	}

	logger.Logger.Debugf("UI created with font %s", fontPath)
	return &UI{
		canvas: cm,
		renderer: text.NewTextRenderer(cm, text.Config{
			CharacterSpacingReduction: cfg.CharacterSpacingReduction,
			LineSpacing:               cfg.TextLineSpacing,
		}),
		font:    font,
		screenW: screenW,
		screenH: screenH,
		cfg:     cfg,
	}, nil
}

// Ready reports whether the UI is usable (non-nil with a loaded font).
func (u *UI) Ready() bool {
	return u != nil && u.font != nil && u.font.IsLoaded()
}

// Text draws a left-aligned white string with its top-left corner at (x, y).
func (u *UI) Text(x, y float64, s string) {
	u.TextColored(x, y, types.White, s)
}

// TextColored draws a left-aligned string in the given color with its top-left
// corner at (x, y).
func (u *UI) TextColored(x, y float64, c types.Color, s string) {
	if u == nil {
		return
	}
	if err := u.renderer.RenderText(s, types.Vector2{X: x, Y: y}, u.font, c); err != nil {
		logger.Logger.Tracef("UI: failed to draw text %q: %s", s, err.Error())
	}
}

// TextCentered draws a string horizontally centered on the screen at height y.
func (u *UI) TextCentered(y float64, c types.Color, s string) {
	u.TextCenteredScaled(y, 1, c, s)
}

// TextCenteredScaled draws a horizontally centered string at height y with
// glyph scale (≥1). Scale ≤0 is treated as 1.
func (u *UI) TextCenteredScaled(y, scale float64, c types.Color, s string) {
	if u == nil {
		return
	}
	if scale <= 0 {
		scale = 1
	}
	x := (u.screenW - u.MeasureScaled(s, scale)) / 2
	if err := u.renderer.RenderTextScaled(s, types.Vector2{X: x, Y: y}, u.font, scale, c); err != nil {
		logger.Logger.Tracef("UI: failed to draw scaled text %q: %s", s, err.Error())
	}
}

// Rect draws a filled rectangle of size (w, h) with its top-left at (x, y).
func (u *UI) Rect(x, y, w, h float64, c types.Color) {
	if u == nil {
		return
	}
	if err := u.canvas.DrawColoredRect(types.Vector2{X: x, Y: y}, types.Vector2{X: w, Y: h}, c); err != nil {
		logger.Logger.Tracef("UI: failed to draw rect: %s", err.Error())
	}
}

// Measure returns the rendered width (in virtual pixels) of a single-line
// string in the default font, matching how the text renderer advances glyphs.
func (u *UI) Measure(s string) float64 {
	return u.MeasureScaled(s, 1)
}

// MeasureScaled returns Measure(s) multiplied by scale.
func (u *UI) MeasureScaled(s string, scale float64) float64 {
	if u == nil {
		return 0
	}
	if scale <= 0 {
		scale = 1
	}
	cellWidth, _ := u.font.GetCellSize()
	advance := float64(cellWidth) - u.cfg.CharacterSpacingReduction
	return float64(len(s)) * advance * scale
}

// LineHeight returns the vertical distance between successive UI text lines.
func (u *UI) LineHeight() float64 {
	return u.LineHeightScaled(1)
}

// LineHeightScaled returns LineHeight multiplied by scale.
func (u *UI) LineHeightScaled(scale float64) float64 {
	if u == nil {
		return 0
	}
	if scale <= 0 {
		scale = 1
	}
	_, cellHeight := u.font.GetCellSize()
	return float64(cellHeight) * u.cfg.UILineSpacing * scale
}

// ScreenSize returns the virtual screen dimensions the UI centers against.
func (u *UI) ScreenSize() (width, height float64) {
	if u == nil {
		return 0, 0
	}
	return u.screenW, u.screenH
}
