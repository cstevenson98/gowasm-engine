# Isometric Sprite Renderer

A command-line tool for rendering 3D models as isometric sprites using Blender. Perfect for creating game assets for isometric/tycoon-style games.

## Features

- **8-direction rendering**: Automatically renders models in 8 rotations (N, NE, E, SE, S, SW, W, NW)
- **Orthographic projection**: True isometric view with no perspective distortion
- **Batch processing**: Render entire directories of models at once
- **Sprite sheet generation**: Combines individual sprites into sprite sheets with metadata
- **Customizable lighting**: Control sun direction, color, and ambient lighting
- **Flexible configuration**: YAML config files with CLI override support
- **Transparent backgrounds**: Ready for game engine integration
- **Extensible**: Interface-based mesh loader system (OBJ support, more formats coming)

## Requirements

- Python 3.9+
- Blender 3.0+ (accessible from command line)
- PyYAML (`pip install pyyaml`)
- Pillow (`pip install pillow`)

## Installation

The tool is already set up in this repository. Just ensure dependencies are installed:

```bash
pip install pyyaml pillow
```

Make sure Blender is accessible from command line:

```bash
blender --version
```

## Quick Start

### Render a single model

```bash
./iso_render --input examples/teapot.obj
```

This will create 8 PNG files in `./sprites/`:
- `teapot_angle_0.png`
- `teapot_angle_45.png`
- `teapot_angle_90.png`
- ... and so on

Plus a sprite sheet: `teapot_spritesheet.png` with metadata JSON.

### Custom size and output directory

```bash
./iso_render --input examples/teapot.obj --size 64 --output ./my-sprites
```

### Use a config file

```bash
./iso_render --input examples/teapot.obj --config examples/render.yaml
```

### Batch render multiple models

```bash
./iso_render --batch models/ --output ./sprites
```

### Customize lighting

```bash
./iso_render --input examples/teapot.obj \
  --light-dir 1,-1,1 \
  --light-color 1.0,0.9,0.8 \
  --ambient-color 0.2,0.2,0.3
```

## Command Line Options

### Input Options

- `--input`, `-i`: Path to single 3D model file
- `--batch`, `-b`: Directory containing multiple 3D models
- `--config`, `-c`: Path to YAML configuration file

### Output Options

- `--output`, `-o`: Output directory (default: `./sprites`)
- `--no-individual-pngs`: Don't save individual PNG files
- `--no-sprite-sheet`: Don't generate sprite sheet
- `--sprite-layout`: Sprite sheet layout: `horizontal`, `vertical`, or `grid` (default: horizontal)

### Render Settings

- `--size`, `-s`: Sprite size in pixels (e.g., `50` or `50,50` or `50x50`)
- `--ortho-scale`: Orthographic camera scale (default: 1.0, lower = more zoomed in)
- `--rotate-x`: Rotate model 90° increments around X axis (pitch): `0`=0°, `1`=90°, `2`=180°, `3`=270°
- `--rotate-y`: Rotate model 90° increments around Y axis (yaw): `0`=0°, `1`=90°, `2`=180°, `3`=270°
- `--rotate-z`: Rotate model 90° increments around Z axis (roll): `0`=0°, `1`=90°, `2`=180°, `3`=270°
- `--no-transparency`: Disable transparent background

**Note:** Many 3D models are exported in different orientations. Use rotation parameters to fix models that appear on their side or facing the wrong way.

### Lighting Options

- `--light-dir`: Sun light direction as `X,Y,Z` (default: `1,-1,1`)
- `--light-color`: Sun light color as `R,G,B` in 0-1 range (default: `1,1,1`)
- `--light-energy`: Sun light intensity (default: 3.0)
- `--ambient-color`: Ambient light color as `R,G,B` (default: `0.3,0.3,0.3`)
- `--ambient-energy`: Ambient light intensity (default: 0.5)

### Other Options

- `--blender-path`: Path to Blender executable (default: `blender`)
- `--verbose`, `-v`: Verbose output
- `--help`, `-h`: Show help message

## Configuration File Format

See `examples/render.yaml` for a complete example:

```yaml
render:
  size: [50, 50]
  directions: 8
  ortho_scale: 6.0

lighting:
  sun_direction: [1, -1, 1]
  sun_color: [1.0, 1.0, 1.0]
  sun_energy: 3.0
  ambient_color: [0.3, 0.3, 0.3]
  ambient_energy: 0.5

camera:
  elevation_angle: 35.264  # True isometric
  rotation_offset: 45

output:
  directory: "./sprites"
  individual_pngs: true
  sprite_sheet: true
  transparent_bg: true
```

## How Models Are Scaled

The tool automatically normalizes all models to fit the camera view:

1. **Scaling**: Based on the **X-Y plane extent** (horizontal footprint), not the height
   - The widest horizontal dimension (X or Y) is scaled to 2.0 units (-1 to +1)
   - This ensures models fill the isometric view properly
   - Height (Z) does not affect scaling

2. **Positioning**: Models are centered at the origin with their **bottom touching the ground plane (Z=0)**
   - This ensures consistent baseline for all sprites
   - Models "sit" on the ground rather than floating

**Example**: A tall thin tower will appear larger than if scaled by overall dimensions (which would shrink it to fit the height).

## Supported File Formats

Currently supported:
- **OBJ** (`.obj`)

Coming soon:
- FBX (`.fbx`)
- GLTF/GLB (`.gltf`, `.glb`)
- Blender files (`.blend`)

## Output Files

For a model named `teapot`, the tool generates:

### Individual PNGs (if enabled)
```
sprites/
├── teapot_angle_0.png
├── teapot_angle_45.png
├── teapot_angle_90.png
├── teapot_angle_135.png
├── teapot_angle_180.png
├── teapot_angle_225.png
├── teapot_angle_270.png
└── teapot_angle_315.png
```

### Sprite Sheet (if enabled)
```
sprites/
├── teapot_spritesheet.png  # Combined sprite sheet
└── teapot_spritesheet.json # Metadata (dimensions, angles, etc.)
```

### Metadata JSON Example

```json
{
  "sprite_width": 50,
  "sprite_height": 50,
  "num_sprites": 8,
  "layout": "horizontal",
  "sprites": [
    {
      "index": 0,
      "x": 0,
      "y": 0,
      "width": 50,
      "height": 50,
      "angle": 0,
      "source_file": "teapot_angle_0.png"
    },
    ...
  ]
}
```

## Integration with Game Engine

The generated sprites can be loaded directly into the gowasm-engine:

1. Copy sprite sheet to `assets/textures/`
2. Load metadata JSON to get sprite coordinates
3. Use sprite coordinates for UV mapping in your game

## Examples

### Example 1: Basic Render

```bash
./iso_render --input examples/teapot.obj
```

### Example 2: High-Res Sprites

```bash
./iso_render --input examples/teapot.obj --size 256 --ortho-scale 4.0
```

### Example 3: Warm Sunset Lighting

```bash
./iso_render --input examples/teapot.obj \
  --light-color 1.0,0.7,0.4 \
  --ambient-color 0.4,0.3,0.2
```

### Example 4: Batch Process with Config

```bash
./iso_render --batch models/ --config examples/render.yaml --output ./game-assets
```

### Example 5: Grid Layout Sprite Sheet

```bash
./iso_render --input examples/teapot.obj --sprite-layout grid
```

## Architecture

The tool uses a modular architecture:

- **config.py**: Configuration management (YAML + CLI merging)
- **loaders/**: Mesh loader interface and implementations
  - `base.py`: Abstract loader interface
  - `obj_loader.py`: OBJ file loader
- **renderer.py**: Blender orchestration and subprocess management
- **output.py**: Sprite sheet generation with Pillow
- **templates/render_script.py**: Blender Python script template
- **iso_render.py**: Main CLI entry point

## Adding New File Format Support

To add support for a new 3D format:

1. Create a new loader in `loaders/` (e.g., `fbx_loader.py`)
2. Extend the `MeshLoader` base class
3. Implement `supports()`, `load_into_blender()`, and `get_file_extensions()`
4. Register in `loaders/__init__.py`

Example:

```python
from .base import MeshLoader

class FBXLoader(MeshLoader):
    def supports(self, filepath: str) -> bool:
        return filepath.lower().endswith('.fbx')
    
    def get_file_extensions(self) -> list[str]:
        return ['.fbx', '.FBX']
    
    def load_into_blender(self, filepath: str, bpy):
        bpy.ops.import_scene.fbx(filepath=filepath)
        # ... handle imported objects
```

## Troubleshooting

### "Blender not found"

Make sure Blender is in your PATH or use `--blender-path`:

```bash
./iso_render --input teapot.obj --blender-path /path/to/blender
```

### "Pillow is required"

Install Pillow:

```bash
pip install pillow
```

### Sprites are too small/large

Adjust `--ortho-scale` (lower = more zoomed in):

```bash
./iso_render --input teapot.obj --ortho-scale 4.0  # More zoomed in
./iso_render --input teapot.obj --ortho-scale 8.0  # More zoomed out
```

### Output is too dark/bright

Adjust lighting:

```bash
./iso_render --input teapot.obj --light-energy 5.0 --ambient-energy 1.0
```

## License

Part of the gowasm-engine project.

## Credits

- Example teapot model from [Stanford CS148](https://graphics.stanford.edu/courses/cs148-10-summer/)
- Powered by [Blender](https://www.blender.org/)

