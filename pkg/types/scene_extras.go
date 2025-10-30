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
