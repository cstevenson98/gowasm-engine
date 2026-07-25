# Root Makefile - Ebiten Game Engine Development

.PHONY: help test test-all test-verbose test-coverage tidy fmt lint \
        build-desktop build-all run-desktop run-desktop-from-assets \
        dev clean clean-all docs docs-cli

.DEFAULT_GOAL := help

# Colors for output
CYAN := \033[0;36m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

##@ General

help: ## Display this help message
	@echo "$(CYAN)Go WASM Engine - Available Commands$(NC)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make $(GREEN)<target>$(NC)\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(CYAN)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Building

# ImGui (cimgui-go) is CGo. Desktop builds that pull pkg/imgui need a C/C++
# toolchain. WASM builds use //go:build js stubs and do not link cimgui.
# If you also import cimgui-go glfw/sdl backends, add:
#   -tags exclude_cimgui_glfw,exclude_cimgui_sdl
CGO_ENABLED ?= 1

build-desktop: ## Build Ebiten desktop binary
	@echo "$(CYAN)Building Ebiten desktop binary...$(NC)"
	@cd cmd/ebiten-game && go mod tidy && CGO_ENABLED=$(CGO_ENABLED) go build -o ../../build/game-desktop
	@echo "$(GREEN)✓ Build complete: build/game-desktop$(NC)"

build-all: build-desktop ## Build all binaries

##@ Running

run-desktop: build-desktop ## Build and run desktop game (from project root)
	@echo "$(CYAN)Running desktop game...$(NC)"
	@cd examples/basic-game && ../../build/game-desktop

run-desktop-from-assets: build-desktop ## Run desktop game from assets directory
	@echo "$(CYAN)Running desktop game from assets directory...$(NC)"
	@cd examples/basic-game/assets && ../../../build/game-desktop

dev: ## Quick rebuild and run (for rapid iteration)
	@echo "$(CYAN)Quick dev build...$(NC)"
	@cd cmd/ebiten-game && CGO_ENABLED=$(CGO_ENABLED) go build -o ../../build/game-desktop
	@cd examples/basic-game && ../../build/game-desktop

##@ Testing

test: ## Run engine library tests
	@echo "$(CYAN)Running engine tests...$(NC)"
	@CGO_ENABLED=$(CGO_ENABLED) go test ./pkg/...

test-all: ## Run all tests (including examples)
	@echo "$(CYAN)Running all tests...$(NC)"
	@CGO_ENABLED=$(CGO_ENABLED) go test ./...

test-verbose: ## Run tests with verbose output
	@echo "$(CYAN)Running tests (verbose)...$(NC)"
	@CGO_ENABLED=$(CGO_ENABLED) go test -v ./pkg/...

test-coverage: ## Run tests with coverage report
	@echo "$(CYAN)Running tests with coverage...$(NC)"
	@CGO_ENABLED=$(CGO_ENABLED) go test -coverprofile=coverage.out ./pkg/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(NC)"

##@ Code Quality

fmt: ## Format all Go code
	@echo "$(CYAN)Formatting Go code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

lint: ## Run linter (requires golangci-lint)
	@echo "$(CYAN)Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "$(YELLOW)⚠ golangci-lint not installed, skipping$(NC)"; \
	fi

tidy: ## Tidy and verify dependencies
	@echo "$(CYAN)Tidying dependencies...$(NC)"
	@go mod tidy
	@cd cmd/ebiten-game && go mod tidy
	@echo "$(GREEN)✓ Dependencies tidied$(NC)"

##@ Documentation

DOCS_PORT ?= 6060

docs: ## Serve browsable API docs at http://localhost:6060 (override DOCS_PORT)
	@echo "$(CYAN)Starting documentation server (port $(DOCS_PORT))...$(NC)"
	@./scripts/serve-docs.sh $(DOCS_PORT)

docs-cli: ## Print package overviews to the terminal (via go doc)
	@for pkg in $$(go list ./pkg/...); do \
		printf "$(CYAN)========== %s ==========$(NC)\n" "$$pkg"; \
		go doc $$pkg; \
		echo ""; \
	done

##@ Cleaning

clean: ## Clean build artifacts
	@echo "$(CYAN)Cleaning build artifacts...$(NC)"
	@rm -rf build/
	@echo "$(GREEN)✓ Build directory cleaned$(NC)"

clean-all: clean ## Clean all artifacts including coverage reports
	@echo "$(CYAN)Cleaning all artifacts...$(NC)"
	@rm -f coverage.out coverage.html
	@find . -name "*.test" -type f -delete
	@echo "$(GREEN)✓ All artifacts cleaned$(NC)"
