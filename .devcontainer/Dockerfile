# Go WASM Engine Dev Container
# Custom Dockerfile for transparent, verifiable builds

FROM debian:bookworm-slim

# Avoid prompts from apt
ENV DEBIAN_FRONTEND=noninteractive

# Install system dependencies
RUN apt-get update && apt-get install -y \
    # Core tools
    git \
    curl \
    wget \
    ca-certificates \
    gnupg \
    lsb-release \
    # Build essentials
    build-essential \
    # Development tools
    vim \
    nano \
    # Process tools
    procps \
    lsof \
    # Network tools
    net-tools \
    # SSL/TLS
    openssl \
    # Sudo for non-root user
    sudo \
    && rm -rf /var/lib/apt/lists/*

# Install Go
ARG GO_VERSION=1.24.3
RUN wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz

# Set Go environment variables
ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

# Install TinyGo
ARG TINYGO_VERSION=0.34.0
RUN wget -q https://github.com/tinygo-org/tinygo/releases/download/v${TINYGO_VERSION}/tinygo${TINYGO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf tinygo${TINYGO_VERSION}.linux-amd64.tar.gz \
    && rm tinygo${TINYGO_VERSION}.linux-amd64.tar.gz

# Add TinyGo to PATH
ENV PATH=$PATH:/usr/local/tinygo/bin

# Install Python
RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    python3-venv \
    && rm -rf /var/lib/apt/lists/*

# Create symbolic link for python -> python3
RUN ln -s /usr/bin/python3 /usr/bin/python

# Install Node.js (v20)
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Create vscode user with sudo privileges
RUN useradd -m -s /bin/bash -u 1000 vscode \
    && echo "vscode ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/vscode \
    && chmod 0440 /etc/sudoers.d/vscode

# Create Go workspace directory
RUN mkdir -p /go/pkg /go/bin \
    && chown -R vscode:vscode /go

# Switch to vscode user
USER vscode

# Install Go tools as vscode user
RUN go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install honnef.co/go/tools/cmd/staticcheck@latest \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Create helper scripts directory
USER root
RUN mkdir -p /usr/local/bin

# Create locate-wasm-exec helper script
RUN cat > /usr/local/bin/locate-wasm-exec << 'EOF' && chmod +x /usr/local/bin/locate-wasm-exec
#!/bin/bash
# Helper script to locate wasm_exec.js files

echo "Go WASM Runtime Files:"
echo "======================"
echo ""

GOROOT=$(go env GOROOT)
echo "GOROOT: $GOROOT"
echo ""

# Standard Go (location changed in Go 1.24+)
GO_WASM_EXEC="${GOROOT}/lib/wasm/wasm_exec.js"
if [ -f "$GO_WASM_EXEC" ]; then
    echo "✅ Go wasm_exec.js:"
    echo "   $GO_WASM_EXEC"
else
    # Fallback for older Go versions
    GO_WASM_EXEC="${GOROOT}/misc/wasm/wasm_exec.js"
    if [ -f "$GO_WASM_EXEC" ]; then
        echo "✅ Go wasm_exec.js (legacy location):"
        echo "   $GO_WASM_EXEC"
    else
        echo "❌ Go wasm_exec.js not found"
    fi
fi

echo ""

# TinyGo
TINYGO_ROOT=$(tinygo env TINYGOROOT 2>/dev/null || echo "/usr/local/tinygo")
TINYGO_WASM_EXEC="${TINYGO_ROOT}/targets/wasm_exec.js"
if [ -f "$TINYGO_WASM_EXEC" ]; then
    echo "✅ TinyGo wasm_exec.js:"
    echo "   $TINYGO_WASM_EXEC"
else
    echo "❌ TinyGo wasm_exec.js not found"
fi

echo ""
echo "Usage in Makefile:"
echo "  # Go 1.24+:"
echo "  WASM_EXEC_JS:=\$(shell go env GOROOT)/lib/wasm/wasm_exec.js"
echo "  # Go 1.23 and older:"
echo "  WASM_EXEC_JS:=\$(shell go env GOROOT)/misc/wasm/wasm_exec.js"
echo "  # TinyGo:"
echo "  TINYGO_WASM_EXEC:=\$(shell tinygo env TINYGOROOT)/targets/wasm_exec.js"
EOF

# Create wasm-env-info helper script
RUN cat > /usr/local/bin/wasm-env-info << 'EOF' && chmod +x /usr/local/bin/wasm-env-info
#!/bin/bash
# Display WASM development environment information

echo "=========================================="
echo "Go WASM Development Environment"
echo "=========================================="
echo ""
echo "Go:"
go version
echo ""
echo "TinyGo:"
tinygo version
echo ""
echo "Python:"
python3 --version
echo ""
echo "Node:"
node --version
echo ""
echo "npm:"
npm --version
echo ""
echo "Environment Variables:"
echo "  GOROOT: $(go env GOROOT)"
echo "  GOPATH: $(go env GOPATH)"
echo "  TINYGOROOT: $(tinygo env TINYGOROOT 2>/dev/null || echo 'Not set')"
echo ""
echo "WASM Support Files:"
locate-wasm-exec
EOF

# Switch back to vscode user
USER vscode

# Set working directory
WORKDIR /workspaces/gowasm-engine

# Default command - devcontainer will override this
CMD ["/bin/bash"]
