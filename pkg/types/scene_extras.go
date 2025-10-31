package types

// SceneOverlayRenderer is an optional interface a Scene can implement
// to render additional overlays (menus, HUD, debug) each frame.
type SceneOverlayRenderer interface {
	// RenderOverlays should render any overlays for the scene.
	RenderOverlays() error
}

// SceneTextureProvider is an optional interface a Scene can implement
// to provide extra texture paths that should be preloaded by the engine
// (e.g., fonts used by overlays/menus).
type SceneTextureProvider interface {
	// GetExtraTexturePaths returns additional texture paths to preload.
	GetExtraTexturePaths() []string
}

// SceneInputProvider is an optional interface a Scene can implement
// to receive the engine's input capturer during initialization.
// The engine automatically injects the input capturer when scenes implement this interface.
type SceneInputProvider interface {
	// SetInputCapturer sets the input capturer for the scene.
	// Called by the engine during scene initialization.
	SetInputCapturer(inputCapturer InputCapturer)
}
