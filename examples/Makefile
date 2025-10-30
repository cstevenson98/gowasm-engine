# Multi-example Makefile (run from examples/ directory)

GOOS=js
GOARCH=wasm
GOROOT:=$(shell go env GOROOT)
WASM_EXEC_JS:=$(GOROOT)/lib/wasm/wasm_exec.js

# Roots for aggregated outputs
BUILD_ROOT=build
DIST_ROOT=dist
PORT=8080

# Discover example directories (exclude build/ and dist/) using shell for reliability
EXAMPLE_DIRS:=$(shell ls -d */ 2>/dev/null | sed 's#/##' | grep -v -E '^(build|dist)$$')

.PHONY: all build clean serve deps info list $(EXAMPLE_DIRS)

all: build

list:
	@echo "Examples: $(EXAMPLE_DIRS)"

deps:
	@mkdir -p $(BUILD_ROOT) $(DIST_ROOT)
	@echo "Ensuring wasm_exec.js is available..."
	@if [ ! -f "$(WASM_EXEC_JS)" ]; then \
		echo "wasm_exec.js not found in GOROOT"; \
		exit 1; \
	fi

# Build all examples
build: deps $(EXAMPLE_DIRS)
	@echo "All examples built. Outputs in $(BUILD_ROOT)/ and $(DIST_ROOT)/"
	@# Generate an index.html at the dist root with links to each example
	@echo "<html><body><h2>Examples</h2><ul>" > $(DIST_ROOT)/index.html
	@for d in $(EXAMPLE_DIRS); do \
		echo "<li><a href=\"$$d/index.html\">$$d</a></li>" >> $(DIST_ROOT)/index.html; \
	 done
	@echo "</ul></body></html>" >> $(DIST_ROOT)/index.html

# Pattern rule: build each example directory
$(EXAMPLE_DIRS):
	@echo "Building example: $@"
	@mkdir -p $(BUILD_ROOT)/$@ $(DIST_ROOT)/$@
	@# Build the WASM within the example module
	@cd $@ && GOOS=$(GOOS) GOARCH=$(GOARCH) go mod tidy && \
		GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ../$(BUILD_ROOT)/$@/main.wasm ./game
	@# Copy assets into dist
	@if [ -d "$@/assets" ]; then \
		cp -r $@/assets/* $(DIST_ROOT)/$@/ 2>/dev/null || true; \
	fi
	@# Copy wasm binary and runtime
	@cp $(BUILD_ROOT)/$@/main.wasm $(DIST_ROOT)/$@/
	@cp $(WASM_EXEC_JS) $(DIST_ROOT)/$@/wasm_exec.js
	@echo "Built $@ → $(BUILD_ROOT)/$@ and provisioned $(DIST_ROOT)/$@"

serve: build
	@SERVE_PORT=$(PORT); \
	if command -v lsof >/dev/null 2>&1; then \
		while lsof -i :$$SERVE_PORT -sTCP:LISTEN -t >/dev/null 2>&1; do SERVE_PORT=$$((SERVE_PORT+1)); done; \
	else \
		while ss -ltn 2>/dev/null | awk '{print $$4}' | grep -q ":$$SERVE_PORT$$"; do SERVE_PORT=$$((SERVE_PORT+1)); done; \
	fi; \
	echo "Serving examples from $(DIST_ROOT) at http://localhost:$$SERVE_PORT"; \
	echo "Open one of:"; \
	for d in $(EXAMPLE_DIRS); do echo "  http://localhost:$$SERVE_PORT/$$d/index.html"; done; \
	cd $(DIST_ROOT) && python3 -m http.server $$SERVE_PORT

clean:
	@rm -rf $(BUILD_ROOT) $(DIST_ROOT)
	@for d in $(EXAMPLE_DIRS); do \
		(cd $$d && go clean -cache >/dev/null 2>&1 || true); \
	done
	@echo "Clean complete."

info:
	@echo "Go:        $(shell go version)"
	@echo "GOROOT:    $(GOROOT)"
	@echo "GOOS:      $(GOOS)"
	@echo "GOARCH:    $(GOARCH)"
	@echo "WASM exec: $(WASM_EXEC_JS)"

