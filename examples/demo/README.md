# Counter demo

Minimal illustrative example for `milo`.

- **Screen:** 1280×720 (`PixelScale` 1)
- **Up arrow:** increments a counter
- **Display:** large centered text via `UI.TextCenteredScaled`

## Run

```bash
# from repo root (nix-shell recommended)
cd examples/demo
go run ./game
```

WASM (from `examples/`):

```bash
make demo
make serve
```
