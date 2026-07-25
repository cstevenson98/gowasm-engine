module example.com/ebiten-game

go 1.24.3

replace github.com/cstevenson98/gowasm-engine => ../../

replace example.com/basic-game => ../../examples/basic-game

require (
	example.com/basic-game v0.0.0
	github.com/cstevenson98/gowasm-engine v0.0.0
	github.com/hajimehoshi/ebiten/v2 v2.9.9
)

require (
	github.com/AllenDang/cimgui-go v1.5.0 // indirect
	github.com/ebitengine/gomobile v0.0.0-20250923094054-ea854a63cce1 // indirect
	github.com/ebitengine/hideconsole v1.0.0 // indirect
	github.com/ebitengine/purego v0.9.0 // indirect
	github.com/jezek/xgb v1.1.1 // indirect
	github.com/mlange-42/ark v0.8.3 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
)
