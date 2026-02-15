# Quick Start Guide

## Installation Complete! ✓

The isometric sprite renderer is ready to use.

## Quick Test

```bash
cd /home/conor/gowasm-engine/tools/iso-sprite-renderer

# Render the example teapot
./iso_render --input examples/teapot.obj --output ./output --size 50
```

## Usage Examples

### 1. Basic render (50x50 sprites)
```bash
./iso_render --input examples/teapot.obj
```

### 2. High-res sprites (128x128)
```bash
./iso_render --input model.obj --size 128
```

### 3. Custom lighting
```bash
./iso_render --input model.obj \
  --light-dir 1,-1,1 \
  --light-color 1.0,0.9,0.8 \
  --ambient-color 0.3,0.3,0.4
```

### 4. Use config file
```bash
./iso_render --input model.obj --config examples/render.yaml
```

### 5. Grid layout sprite sheet
```bash
./iso_render --input model.obj --sprite-layout grid
```

### 6. Only sprite sheet (no individual PNGs)
```bash
./iso_render --input model.obj --no-individual-pngs
```

### 7. Batch process directory
```bash
./iso_render --batch models/ --output ./sprites
```

## Output

For each model, you get:
- 8 individual PNG files (one per rotation: 0°, 45°, 90°, 135°, 180°, 225°, 270°, 315°)
- Combined sprite sheet PNG
- Metadata JSON with sprite coordinates and angles

## Verified Tests

✓ Single model render with default settings (50x50)
✓ Config file with CLI overrides (64x64)
✓ Grid layout sprite sheet
✓ Individual PNG suppression
✓ Sprite sheet generation with metadata
✓ Blender 5.0 compatibility
✓ Orthographic isometric projection
✓ Transparent backgrounds (RGBA)

## Performance

The teapot (3000 faces) takes about:
- ~5 seconds per angle (40 seconds total for 8 directions)
- Simpler models will render faster
- Complex models may take longer

## Next Steps

1. Try rendering your own OBJ files
2. Customize lighting in the config file
3. Integrate sprites into your game engine
4. See full documentation in README.md

Enjoy creating isometric sprites! 🎮

