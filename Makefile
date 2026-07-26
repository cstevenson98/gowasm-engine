# Root Makefile — engine tests + examples (build / run / wasm serve)

.PHONY: help test test-all test-verbose test-coverage tidy fmt lint \
        examples list-examples build-examples run-demo serve \
        clean clean-all docs docs-cli

.DEFAULT_GOAL := help

CYAN := \033[0;36m
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m

# ImGui (cimgui-go) is CGo on desktop. WASM uses //go:build js stubs.
CGO_ENABLED ?= 1

# Example to run with `make run` / `make run-demo` (default: demo)
EXAMPLE ?= demo

##@ General

help: ## Display this help message
	@echo "$(CYAN)gowasm-engine — Available Commands$(NC)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make $(GREEN)<target>$(NC)\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(CYAN)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Testing

test: ## Run engine library tests (./pkg/...)
	@echo "$(CYAN)Running engine tests...$(NC)"
	@CGO_ENABLED=$(CGO_ENABLED) go test ./pkg/...

test-all: ## Run all tests in this module
	@echo "$(CYAN)Running all module tests...$(NC)"
	@CGO_ENABLED=$(CGO_ENABLED) go test ./...

test-verbose: ## Run engine tests with verbose output
	@echo "$(CYAN)Running tests (verbose)...$(NC)"
	@CGO_ENABLED=$(CGO_ENABLED) go test -v ./pkg/...

test-coverage: ## Run engine tests with coverage report
	@echo "$(CYAN)Running tests with coverage...$(NC)"
	@CGO_ENABLED=$(CGO_ENABLED) go test -coverprofile=coverage.out ./pkg/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(NC)"

##@ Examples

list-examples: ## List example modules under examples/
	@$(MAKE) -C examples list

build-examples: ## Build all examples to WASM (examples/build + examples/dist)
	@echo "$(CYAN)Building examples (wasm)...$(NC)"
	@$(MAKE) -C examples build

examples: build-examples ## Alias for build-examples

run: run-demo ## Run EXAMPLE (default: demo) on desktop

run-demo: ## Run examples/$(EXAMPLE) on desktop (`go run ./game`)
	@echo "$(CYAN)Running examples/$(EXAMPLE)...$(NC)"
	@cd examples/$(EXAMPLE) && CGO_ENABLED=$(CGO_ENABLED) go run ./game

serve: ## Build examples to WASM and serve examples/dist
	@echo "$(CYAN)Building + serving examples...$(NC)"
	@$(MAKE) -C examples serve

##@ Code Quality

fmt: ## Format all Go code in this module
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

tidy: ## Tidy root + each examples/*/go.mod
	@echo "$(CYAN)Tidying dependencies...$(NC)"
	@go mod tidy
	@for d in examples/*/; do \
		if [ -f "$$d/go.mod" ]; then \
			echo "  tidy $$d"; \
			(cd "$$d" && go mod tidy); \
		fi; \
	done
	@echo "$(GREEN)✓ Dependencies tidied$(NC)"

##@ Documentation

DOCS_PORT ?= 6060

docs: ## Serve browsable API docs (override DOCS_PORT)
	@echo "$(CYAN)Starting documentation server (port $(DOCS_PORT))...$(NC)"
	@./scripts/serve-docs.sh $(DOCS_PORT)

docs-cli: ## Print package overviews to the terminal (via go doc)
	@for pkg in $$(go list ./pkg/...); do \
		printf "$(CYAN)========== %s ==========$(NC)\n" "$$pkg"; \
		go doc $$pkg; \
		echo ""; \
	done

##@ Cleaning

clean: ## Clean root build/ and examples build+dist
	@echo "$(CYAN)Cleaning build artifacts...$(NC)"
	@rm -rf build/
	@$(MAKE) -C examples clean
	@echo "$(GREEN)✓ Clean complete$(NC)"

clean-all: clean ## Also remove coverage reports
	@echo "$(CYAN)Cleaning coverage artifacts...$(NC)"
	@rm -f coverage.out coverage.html
	@find . -name "*.test" -type f -delete
	@echo "$(GREEN)✓ All artifacts cleaned$(NC)"
