# Dev Container Troubleshooting Guide

This guide covers common issues when using the dev container and their solutions.

## Container Won't Build

### Problem: "Failed to build dev container"

**Symptoms**: Container build fails during initialization

**Solutions**:

1. **Check Docker is running**:
   ```bash
   docker ps
   ```

2. **Clean Docker cache**:
   ```bash
   docker system prune -a
   ```

3. **Rebuild without cache**:
   - Press `F1` in VSCode
   - Select "Dev Containers: Rebuild Container Without Cache"

4. **Check internet connection**: TinyGo needs to be downloaded from GitHub

5. **Check disk space**: `docker system df` to see Docker disk usage

### Problem: "Post-create script failed"

**Symptoms**: Container builds but post-create.sh fails

**Solutions**:

1. **Check script syntax**:
   ```bash
   bash -n .devcontainer/post-create.sh
   ```

2. **Check script permissions**:
   ```bash
   ls -la .devcontainer/post-create.sh
   # Should show -rwxr-xr-x
   ```

3. **Run manually to see detailed error**:
   ```bash
   bash .devcontainer/post-create.sh
   ```

4. **Check TinyGo download URL**: Verify version exists on GitHub releases

## Go/TinyGo Issues

### Problem: "Command not found: tinygo"

**Symptoms**: `tinygo: command not found` when running TinyGo

**Solutions**:

1. **Check if TinyGo is installed**:
   ```bash
   ls -la /usr/local/tinygo
   ```

2. **Check PATH**:
   ```bash
   echo $PATH | grep tinygo
   ```

3. **Source the profile script**:
   ```bash
   source /etc/profile.d/tinygo.sh
   ```

4. **Restart terminal**: Open a new terminal in VSCode

5. **Manually add to PATH**:
   ```bash
   export PATH=$PATH:/usr/local/tinygo/bin
   ```

### Problem: "wasm_exec.js not found"

**Symptoms**: Build fails with "wasm_exec.js not found"

**Solutions**:

1. **Locate the files**:
   ```bash
   locate-wasm-exec
   ```

2. **Verify GOROOT is set**:
   ```bash
   go env GOROOT
   ```

3. **Check if file exists**:
   ```bash
   # Go 1.24+ location
   ls -la "$(go env GOROOT)/lib/wasm/wasm_exec.js"
   # Go 1.23 and older location
   ls -la "$(go env GOROOT)/misc/wasm/wasm_exec.js"
   # TinyGo location
   ls -la "$(tinygo env TINYGOROOT)/targets/wasm_exec.js"
   ```

4. **Update Makefile if needed**: Ensure using correct path for your Go version:
   ```makefile
   # Go 1.24+
   WASM_EXEC_JS:=$(GOROOT)/lib/wasm/wasm_exec.js
   # Go 1.23 and older
   # WASM_EXEC_JS:=$(GOROOT)/misc/wasm/wasm_exec.js
   ```
   **Important**: Go changed the location in version 1.24!

### Problem: "Go module errors"

**Symptoms**: `go: module not found` or dependency errors

**Solutions**:

1. **Clean module cache**:
   ```bash
   go clean -modcache
   ```

2. **Re-download dependencies**:
   ```bash
   go mod download
   go mod tidy
   ```

3. **For examples**:
   ```bash
   cd examples/basic-game
   go mod tidy
   cd ../..
   ```

4. **Check go.mod file**: Ensure module path is correct

5. **Verify network access**: Container needs internet for `go get`

## Build Issues

### Problem: "Build fails with WASM errors"

**Symptoms**: WASM build succeeds but binary doesn't work

**Solutions**:

1. **Check GOOS/GOARCH**:
   ```bash
   echo $GOOS $GOARCH
   # Should show: js wasm
   ```

2. **Verify build command**:
   ```bash
   GOOS=js GOARCH=wasm go build -o main.wasm ./game
   ```

3. **Check for build tag issues**: Files with `//go:build js` need GOOS=js

4. **Verify wasm_exec.js matches compiler**:
   - Standard Go → use Go's wasm_exec.js
   - TinyGo → use TinyGo's wasm_exec.js

5. **Check browser console**: Look for JavaScript errors

### Problem: "syscall/js not found"

**Symptoms**: `package syscall/js is not in GOROOT`

**Solutions**:

1. **Must use GOOS=js GOARCH=wasm**:
   ```bash
   GOOS=js GOARCH=wasm go build ./...
   ```

2. **Check build tags**: Ensure `//go:build js` is present

3. **For tests**:
   ```bash
   GOOS=js GOARCH=wasm go test ./...
   ```

## Port/Server Issues

### Problem: "Port 8080 already in use"

**Symptoms**: `make serve` fails with "address already in use"

**Solutions**:

1. **Makefile auto-detects**: Should automatically find next available port

2. **Kill process on port**:
   ```bash
   lsof -ti:8080 | xargs kill -9
   ```

3. **Use different port manually**:
   ```bash
   cd examples/dist
   python3 -m http.server 8081
   ```

4. **Check forwarded ports in VSCode**: Ports tab in bottom panel

### Problem: "Can't access http://localhost:8080"

**Symptoms**: Browser shows "can't connect" error

**Solutions**:

1. **Check if server is running**:
   ```bash
   ps aux | grep http.server
   ```

2. **Check port forwarding**: VSCode should show forwarded ports in Ports tab

3. **Try 127.0.0.1 instead of localhost**:
   ```
   http://127.0.0.1:8080
   ```

4. **Check VSCode port forwarding settings**: Ensure auto-forwarding is enabled

5. **Manually forward port**: In Ports tab, click "Add Port" → enter 8080

## Performance Issues

### Problem: "Container is slow"

**Symptoms**: Builds and tests take a long time

**Solutions**:

1. **Check Docker resources**: Increase CPU/memory in Docker settings

2. **Use volume for cache**: Already configured in devcontainer.json

3. **Limit concurrent builds**:
   ```bash
   go build -p 1 ./...  # Single process
   ```

4. **Clean build cache**:
   ```bash
   go clean -cache
   ```

5. **Restart Docker**: Sometimes Docker daemon gets slow

### Problem: "Go module downloads are slow"

**Symptoms**: `go mod download` takes forever

**Solutions**:

1. **Use Go module proxy**: Should be automatic, but verify:
   ```bash
   go env GOPROXY
   # Should show: https://proxy.golang.org,direct
   ```

2. **Check network**: Container needs internet access

3. **Use private proxy**: If behind corporate firewall:
   ```bash
   export GOPROXY=https://your-proxy.com
   ```

## VSCode Issues

### Problem: "Go extension not working"

**Symptoms**: No IntelliSense, red squiggles everywhere

**Solutions**:

1. **Check extension is installed**: Should auto-install, but verify in Extensions panel

2. **Reload window**: `F1` → "Developer: Reload Window"

3. **Check gopls is running**:
   ```bash
   ps aux | grep gopls
   ```

4. **Reinstall Go tools**: `F1` → "Go: Install/Update Tools"

5. **Check output panel**: Go → Gopls (server) for errors

### Problem: "IntelliSense shows wrong imports"

**Symptoms**: Auto-complete suggests non-WASM packages

**Solutions**:

1. **VSCode doesn't know about build tags**: This is expected

2. **Use build tags properly**: `//go:build js`

3. **Test with actual build**:
   ```bash
   GOOS=js GOARCH=wasm go build ./...
   ```

## Permission Issues

### Problem: "Permission denied" errors

**Symptoms**: Can't write files, run commands, etc.

**Solutions**:

1. **Check user**: Container runs as `vscode` user
   ```bash
   whoami  # Should show: vscode
   ```

2. **Use sudo for privileged operations**:
   ```bash
   sudo apt-get install something
   ```

3. **Check file permissions**:
   ```bash
   ls -la /path/to/file
   ```

4. **Fix ownership** (if needed):
   ```bash
   sudo chown vscode:vscode /path/to/file
   ```

## WebGPU Runtime Issues

### Problem: "WebGPU not available in browser"

**Symptoms**: Browser console shows WebGPU errors

**Solutions**:

1. **Use Chrome/Edge**: Best WebGPU support

2. **Enable WebGPU**: Go to `chrome://flags` → Enable "Unsafe WebGPU"

3. **Update browser**: WebGPU is new, update to latest version

4. **Check GPU**: Some virtual machines don't support WebGPU

5. **Use Firefox Nightly**: Has experimental WebGPU support

## Getting Help

### Diagnostic Commands

Run these to gather information for debugging:

```bash
# Environment info
wasm-env-info

# WASM files
locate-wasm-exec

# Docker info
docker version
docker info

# Go environment
go env

# TinyGo environment
tinygo env

# Build info
cd examples && make info
```

### Logs to Check

1. **Docker logs**: `docker logs <container-id>`
2. **VSCode Output panel**: View → Output → select "Dev Containers"
3. **Browser console**: F12 → Console tab
4. **Go build output**: Check terminal output from `make build`

### Reset Everything

If all else fails, nuke it from orbit:

```bash
# Exit container
exit

# In host terminal:
docker system prune -a -f
docker volume prune -f

# In VSCode:
# F1 → "Dev Containers: Rebuild Container Without Cache"
```

### Report Issues

If you've tried everything and still have problems:

1. Run diagnostic commands above
2. Check container logs
3. Open an issue with:
   - Error messages
   - Steps to reproduce
   - Output of `wasm-env-info`
   - Docker version
   - Host OS

## Useful Resources

- [VSCode Dev Containers Docs](https://code.visualstudio.com/docs/devcontainers/containers)
- [Go WASM Documentation](https://github.com/golang/go/wiki/WebAssembly)
- [TinyGo WASM Guide](https://tinygo.org/docs/guides/webassembly/)
- [Docker Documentation](https://docs.docker.com/)
- [WebGPU Browser Support](https://caniuse.com/webgpu)
