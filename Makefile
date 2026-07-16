# Root Makefile - engine library and desktop builds

.PHONY: test test-all tidy build-desktop run-desktop build-wasm clean

test:
	go test ./pkg/...

test-all:
	go test ./...

tidy:
	go mod tidy

# Build Ebiten desktop binary
build-desktop:
	@echo "Building Ebiten desktop binary..."
	cd cmd/ebiten-game && go mod tidy && go build -o ../../build/game-desktop
	@echo "Build complete: build/game-desktop"

# Run Ebiten desktop game
run-desktop: build-desktop
	@echo "Running desktop game..."
	./build/game-desktop

# Build WASM binary (legacy WebGPU version) 
build-wasm:
	@echo "Building WASM binary (WebGPU)..."
	cd examples/basic-game/game && GOOS=js GOARCH=wasm go build -o ../../../build/main.wasm
	@echo "Build complete: build/main.wasm"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf build/
	@echo "Clean complete"
