# Dockerfile Build Details

This document explains what the custom Dockerfile does and why, so you can verify it builds correctly and understand what's being installed.

## Base Image

```dockerfile
FROM debian:bookworm-slim
```

We use Debian Bookworm (Debian 12) slim as the base. This is:
- **Lightweight**: Minimal base, we only install what we need
- **Stable**: Debian stable release, well-tested
- **Transparent**: No hidden Microsoft/VSCode layers, you see exactly what's installed

## Build Arguments

The Dockerfile accepts two build arguments that you can customize:

```dockerfile
ARG GO_VERSION=1.24.3
ARG TINYGO_VERSION=0.34.0
```

To change versions, edit `.devcontainer/devcontainer.json`:

```json
"build": {
  "args": {
    "GO_VERSION": "1.25.0",
    "TINYGO_VERSION": "0.35.0"
  }
}
```

## Installation Steps

### 1. System Dependencies

```dockerfile
RUN apt-get update && apt-get install -y \
    git curl wget ca-certificates gnupg lsb-release \
    build-essential \
    vim nano \
    procps lsof \
    net-tools \
    openssl \
    sudo
```

**Why each package:**
- `git` - Version control (essential)
- `curl`, `wget` - Download files (used to fetch Go, TinyGo, Node.js)
- `ca-certificates` - SSL/TLS certificates for HTTPS
- `gnupg`, `lsb-release` - Needed for Node.js repository setup
- `build-essential` - C compiler and build tools (Go's cgo needs this)
- `vim`, `nano` - Text editors for quick edits
- `procps` - Process tools (ps, top, etc.)
- `lsof` - Check what's using ports
- `net-tools` - Network utilities
- `openssl` - SSL/TLS tools
- `sudo` - Allow vscode user to run privileged commands

### 2. Go Installation

```dockerfile
RUN wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz
```

- Downloads official Go binary from golang.org
- Extracts to `/usr/local/go`
- Cleans up tarball
- Sets `GOROOT=/usr/local/go` and adds to `PATH`

**Verify:** `go version` should show `go1.24.3`

### 3. TinyGo Installation

```dockerfile
RUN wget -q https://github.com/tinygo-org/tinygo/releases/download/v${TINYGO_VERSION}/tinygo${TINYGO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf tinygo${TINYGO_VERSION}.linux-amd64.tar.gz \
    && rm tinygo${TINYGO_VERSION}.linux-amd64.tar.gz
```

- Downloads official TinyGo release from GitHub
- Extracts to `/usr/local/tinygo`
- Cleans up tarball
- Adds to `PATH`

**Verify:** `tinygo version` should show `tinygo version 0.34.0`

### 4. Python Installation

```dockerfile
RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    python3-venv \
    && rm -rf /var/lib/apt/lists/*

RUN ln -s /usr/bin/python3 /usr/bin/python
```

- Installs Python 3 from Debian repositories (typically 3.11.x on Bookworm)
- Installs pip and venv for package management
- Creates `python` symlink to `python3` for convenience
- Used for `make serve` which runs `python3 -m http.server`

**Verify:** `python --version` should show `Python 3.11.x`

### 5. Node.js Installation

```dockerfile
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*
```

- Adds NodeSource repository for Node.js 20
- Installs Node.js 20 and npm
- Optional but useful for web tooling

**Verify:** `node --version` should show `v20.x.x`

### 6. User Setup

```dockerfile
RUN useradd -m -s /bin/bash -u 1000 vscode \
    && echo "vscode ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/vscode \
    && chmod 0440 /etc/sudoers.d/vscode
```

Creates a non-root user called `vscode`:
- UID 1000 (matches most host users, good for file permissions)
- Home directory at `/home/vscode`
- Bash shell
- Full sudo privileges without password (for convenience in dev)

**Security Note:** This is fine for development containers but never use NOPASSWD sudo in production!

### 7. Go Workspace

```dockerfile
RUN mkdir -p /go/pkg /go/bin \
    && chown -R vscode:vscode /go
```

- Creates `/go` directory for Go workspace
- Sets `GOPATH=/go`
- Makes vscode user the owner

### 8. Go Development Tools

```dockerfile
USER vscode
RUN go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install honnef.co/go/tools/cmd/staticcheck@latest \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

Installs essential Go tools as the vscode user:
- **gopls** - Go language server (powers IntelliSense)
- **dlv** - Delve debugger
- **staticcheck** - Static analysis tool
- **golangci-lint** - Meta-linter that runs multiple linters

**Verify:** All commands should be in PATH: `which gopls dlv staticcheck golangci-lint`

### 9. Helper Scripts

```dockerfile
RUN cat > /usr/local/bin/locate-wasm-exec << 'EOF' && chmod +x /usr/local/bin/locate-wasm-exec
#!/bin/bash
# ... script content ...
EOF

RUN cat > /usr/local/bin/wasm-env-info << 'EOF' && chmod +x /usr/local/bin/wasm-env-info
#!/bin/bash
# ... script content ...
EOF
```

Creates two helper scripts in `/usr/local/bin`:

**locate-wasm-exec:**
- Finds Go's `wasm_exec.js` at `$(GOROOT)/misc/wasm/wasm_exec.js`
- Finds TinyGo's `wasm_exec.js` at `$(TINYGOROOT)/targets/wasm_exec.js`
- Shows paths for use in Makefiles

**wasm-env-info:**
- Displays versions of all installed tools
- Shows environment variables (GOROOT, GOPATH, TINYGOROOT)
- Calls `locate-wasm-exec` to show WASM runtime files
- Quick way to verify everything is set up correctly

## File Locations

After building, here's where everything is:

```
/usr/local/go/              # Go installation
  ├── bin/go                # Go compiler
  ├── lib/wasm/             # Go 1.24+ location
  │   └── wasm_exec.js      # Go WASM runtime (IMPORTANT!)
  ├── misc/wasm/            # Legacy location (Go 1.23-)
  │   └── wasm_exec.html    # Example HTML wrapper
  └── ...

/usr/local/tinygo/          # TinyGo installation
  ├── bin/tinygo            # TinyGo compiler
  ├── targets/
  │   └── wasm_exec.js      # TinyGo WASM runtime (IMPORTANT!)
  └── ...

/go/                        # Go workspace (GOPATH)
  ├── bin/                  # Installed Go binaries
  │   ├── gopls
  │   ├── dlv
  │   ├── staticcheck
  │   └── golangci-lint
  └── pkg/                  # Go module cache

/usr/local/bin/
  ├── locate-wasm-exec      # Helper script
  └── wasm-env-info         # Helper script

/home/vscode/               # User home directory
```

## Environment Variables

The Dockerfile sets these environment variables:

```bash
GOROOT=/usr/local/go
GOPATH=/go
PATH=$PATH:$GOROOT/bin:$GOPATH/bin:/usr/local/tinygo/bin
```

These are inherited by all shells in the container.

## Testing the Build

Run the test script to verify everything:

```bash
cd .devcontainer
./test-build.sh
```

This runs 9 automated tests:
1. ✅ Go version
2. ✅ TinyGo version
3. ✅ Python version
4. ✅ Node.js version
5. ✅ wasm_exec.js files exist
6. ✅ Go tools are installed
7. ✅ Helper scripts are installed
8. ✅ Full environment info
9. ✅ Default user is vscode

## Build Time

Expected build times:
- **First build**: 2-3 minutes (downloads Go, TinyGo, Node.js, installs tools)
- **Cached build**: <10 seconds (Docker layer caching)
- **After changing versions**: ~1-2 minutes (only rebuilds changed layers)

## Size

Expected image size: ~1.5-2 GB

Breakdown:
- Base Debian: ~100 MB
- Go: ~400 MB
- TinyGo: ~150 MB
- Node.js: ~200 MB
- Tools and dependencies: ~200 MB
- Cached modules: varies

## Customization

### Change Go Version

Edit `devcontainer.json`:
```json
"build": {
  "args": {
    "GO_VERSION": "1.25.0"
  }
}
```

### Change TinyGo Version

Edit `devcontainer.json`:
```json
"build": {
  "args": {
    "TINYGO_VERSION": "0.35.0"
  }
}
```

### Add More System Packages

Edit `Dockerfile`, add to the first `apt-get install` section:
```dockerfile
RUN apt-get update && apt-get install -y \
    # ... existing packages ...
    your-new-package \
    another-package
```

### Add More Go Tools

Edit `Dockerfile`, add to the Go tools section:
```dockerfile
USER vscode
RUN go install golang.org/x/tools/gopls@latest \
    # ... existing tools ...
    && go install github.com/your/tool@latest
```

## Why Not Use Microsoft's Image?

Microsoft's pre-built dev container images are great, but:

1. **Transparency**: You can't easily see what's installed
2. **Control**: Can't customize the base without hacking
3. **Size**: Often includes extras you don't need
4. **Verification**: Can't easily test the build process
5. **Understanding**: Custom Dockerfile teaches you what's needed

Our custom Dockerfile gives you full control and visibility!

## Troubleshooting

### Build fails downloading Go

- **Issue**: Network error or version doesn't exist
- **Fix**: Check Go version exists at https://go.dev/dl/
- **Fix**: Check your internet connection

### Build fails downloading TinyGo

- **Issue**: Network error or version doesn't exist
- **Fix**: Check TinyGo version exists at https://github.com/tinygo-org/tinygo/releases
- **Fix**: Check your internet connection

### Build succeeds but wasm_exec.js not found

- **Issue**: Go changed the file location
- **Fix**: Check with `go env GOROOT` and verify path
- **Fix**: Update Makefile to use correct path

### Go tools not found

- **Issue**: PATH not set correctly
- **Fix**: Check `echo $PATH` includes `/go/bin`
- **Fix**: Rebuild container without cache

## Security Considerations

This dev container is designed for **development only**, not production:

✅ **Safe for development:**
- Non-root user by default
- sudo access for convenience
- All tools from official sources

❌ **Don't use in production:**
- NOPASSWD sudo (security risk)
- Development tools included
- Debug symbols present

## References

- [Dockerfile Best Practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- [Go Downloads](https://go.dev/dl/)
- [TinyGo Releases](https://github.com/tinygo-org/tinygo/releases)
- [Node.js Installation](https://github.com/nodesource/distributions)
- [Dev Containers Specification](https://containers.dev/)
