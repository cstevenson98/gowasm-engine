module example.com/basic-game

go 1.24.3

require github.com/conor/webgpu-triangle v0.0.0

require (
	github.com/cogentcore/webgpu v0.23.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
)

replace github.com/conor/webgpu-triangle => ../..
