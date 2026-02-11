#!/bin/bash
set -e

echo "=========================================="
echo "Setting up Go WASM Development Environment"
echo "=========================================="

# Verify installations
echo ""
echo "🔍 Verifying installations..."
wasm-env-info

# Install project dependencies
echo ""
echo "📚 Installing project dependencies..."
cd /workspaces/gowasm-engine

# Tidy root module
go mod tidy

# Tidy example modules
if [ -d "examples/basic-game" ]; then
    cd examples/basic-game
    go mod tidy
    cd ../..
fi

echo "✅ Dependencies installed"

echo ""
echo "=========================================="
echo "✅ Development environment ready!"
echo "=========================================="
echo ""
echo "Quick Start:"
echo "  - Run 'make build' in examples/ to build all examples"
echo "  - Run 'make serve' in examples/ to start dev server"
echo "  - Run 'locate-wasm-exec' to find wasm_exec.js files"
echo "  - Run 'wasm-env-info' to see environment details"
echo ""
