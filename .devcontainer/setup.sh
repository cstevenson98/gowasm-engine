#!/bin/bash
set -e

echo "🔧 Setting up Go WASM game engine development environment..."
echo ""

# Install Go to user's home directory (no sudo needed)
GO_VERSION="1.23.5"
GO_INSTALL_DIR="$HOME/.local/go"

if [ -d "$GO_INSTALL_DIR" ] && [ -x "$GO_INSTALL_DIR/bin/go" ]; then
    echo "✓ Go is already installed at $GO_INSTALL_DIR"
    EXISTING_VERSION=$($GO_INSTALL_DIR/bin/go version | awk '{print $3}')
    echo "  Version: $EXISTING_VERSION"
else
    echo "🐹 Installing Go ${GO_VERSION} to $GO_INSTALL_DIR..."
    mkdir -p "$HOME/.local"
    wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
    rm -rf "$GO_INSTALL_DIR"
    tar -C "$HOME/.local" -xzf go${GO_VERSION}.linux-amd64.tar.gz
    rm go${GO_VERSION}.linux-amd64.tar.gz
    echo "✓ Go ${GO_VERSION} installed"
fi

# Set up Go environment for this script
export PATH=$GO_INSTALL_DIR/bin:$PATH
export GOPATH=$HOME/go
export PATH=$GOPATH/bin:$PATH

# Add to shell profile for persistence (only if not already there)
setup_shell_rc() {
    local rc_file=$1
    if [ -f "$rc_file" ]; then
        if ! grep -q "export PATH=.*/.local/go/bin" "$rc_file"; then
            echo "" >> "$rc_file"
            echo "# Go installation (added by gowasm-engine setup)" >> "$rc_file"
            echo "export PATH=\$HOME/.local/go/bin:\$PATH" >> "$rc_file"
            echo "export GOPATH=\$HOME/go" >> "$rc_file"
            echo "export PATH=\$GOPATH/bin:\$PATH" >> "$rc_file"
            echo "✓ Added Go to $rc_file"
        else
            echo "✓ Go already configured in $rc_file"
        fi
    fi
}

setup_shell_rc "$HOME/.bashrc"
setup_shell_rc "$HOME/.zshrc"

# Verify Go installation
echo ""
echo "🔍 Verifying Go installation..."
if ! command -v go &> /dev/null; then
    echo "❌ Go command not found in PATH"
    echo "   Please restart your terminal or run: export PATH=$GO_INSTALL_DIR/bin:\$PATH"
    exit 1
fi

echo "✓ Go is working: $(go version)"
echo "  GOROOT: $(go env GOROOT)"
echo "  GOPATH: $(go env GOPATH)"

# Install Go tools
echo ""
echo "🔨 Installing Go development tools..."
go install golang.org/x/tools/gopls@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
echo "✓ Development tools installed"

# Install project dependencies
echo ""
echo "📚 Installing project dependencies..."
cd "$(dirname "$0")/.."
go mod download
go mod tidy
echo "✓ Main module dependencies installed"

# Install dependencies for examples
echo ""
echo "📦 Installing example dependencies..."
if [ -d "examples/basic-game" ]; then
    cd examples/basic-game
    go mod download
    go mod tidy
    echo "✓ Example dependencies installed"
    cd ../..
fi

echo ""
echo "✅ Setup complete!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Environment Summary:"
echo "  Go Version:  $(go version | awk '{print $3}')"
echo "  Go Location: $(which go)"
echo "  GOROOT:      $(go env GOROOT)"
echo "  GOPATH:      $(go env GOPATH)"
echo "  Python:      $(python3 --version 2>&1)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📋 Next Steps:"
echo ""
echo "1. Restart your terminal (or run: source ~/.zshrc)"
echo ""
echo "2. Build the example game:"
echo "   cd examples"
echo "   make build"
echo ""
echo "3. Run the example game:"
echo "   cd examples"
echo "   make serve"
echo "   Then open http://localhost:8080 in your browser"
echo ""
echo "4. Run tests:"
echo "   make test           # Run unit tests"
echo "   cd examples && make clean  # Clean build artifacts"
echo ""
