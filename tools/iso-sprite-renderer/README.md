# Isometric Sprite Renderer Library

A Python library for rendering 3D models as isometric sprites using Blender. Designed for creating game assets for isometric/tycoon-style games.

## Overview

This is a **library tool** - it provides core rendering capabilities that other tools can use. For complete applications, see:
- `electrical_sprite_database` - Generate electrical grid component sprites

## Features

- **8-direction rendering**: Automatically renders models in 8 rotations (N, NE, E, SE, S, SW, W, NW)
- **Orthographic projection**: True isometric view with no perspective distortion
- **Customizable lighting**: Control sun direction, color, and ambient lighting
- **Flexible configuration**: YAML config files with CLI override support
- **Transparent backgrounds**: Ready for game engine integration
- **Extensible**: Interface-based mesh loader system

## Requirements

- Python 3.9+
- Blender 3.0+ (accessible from command line)
- PyYAML (`pip install pyyaml`)

## Installation

This is a library - add the `tools/` directory to your Python path:

```python
import sys
sys.path.insert(0, '/path/to/tools')

from iso_sprite_renderer import config, renderer
```

Or use it via command line:

```bash
python3 tools/iso-sprite-renderer/iso_render.py --input model.obj --output sprites/
```

## Command Line Usage

### Basic Rendering

```bash
python3 iso_render.py --input model.obj --output sprites/ --size 64
```

### Custom Camera Settings

```bash
python3 iso_render.py --input model.obj \
  --ortho-scale 2.828427 \  # sqrt(8) for perfect unit square fit
  --size 64 \
  --output sprites/
```

### Rotation and Positioning

```bash
python3 iso_render.py --input model.obj \
  --rotate-x 1 \            # Rotate 90° around X axis
  --camera-focus-x 0 \      # Center camera at X=0
  --camera-focus-y 0 \      # Center camera at Y=0
  --no-normalize            # Use absolute coordinates from OBJ
```

### Debugging Options

```bash
python3 iso_render.py --input model.obj \
  --show-ground-plane \     # Show unit square at Z=0
  --show-axes \             # Show X,Y,Z coordinate axes
  --render-top-view \       # Also render from directly above
  --render-side-view        # Also render from the side
```

## Command Line Options

### Input/Output
- `--input`, `-i`: Path to 3D model file (required)
- `--output`, `-o`: Output directory (default: `./sprites`)
- `--config`, `-c`: Path to YAML configuration file

### Render Settings
- `--size`, `-s`: Sprite size in pixels (e.g., `64` or `64,64`)
- `--ortho-scale`: Orthographic camera scale (lower = more zoomed in)
- `--rotate-x`: Rotate model 90° increments around X axis (0-3)
- `--rotate-y`: Rotate model 90° increments around Y axis (0-3)
- `--rotate-z`: Rotate model 90° increments around Z axis (0-3)
- `--no-normalize`: Disable automatic scaling/centering
- `--camera-focus-x`: Camera focus point X coordinate
- `--camera-focus-y`: Camera focus point Y coordinate

### Output Options
- `--no-sprite-sheet`: Don't generate sprite sheet
- `--no-individual-pngs`: Don't save individual PNG files

### Debugging
- `--show-ground-plane`: Show unit square border at Z=0
- `--show-axes`: Show coordinate axes
- `--render-top-view`: Render top-down view
- `--render-side-view`: Render side-on view

### Lighting
- `--light-dir`: Sun light direction as `X,Y,Z`
- `--light-color`: Sun light color as `R,G,B`
- `--light-energy`: Sun light intensity
- `--ambient-color`: Ambient light color as `R,G,B`
- `--ambient-energy`: Ambient light intensity

## Python API Usage

### Basic Rendering

```python
import sys
import os

# Add to path
sys.path.insert(0, '/path/to/tools')

# Import renderer
from iso_sprite_renderer.renderer import BlenderRenderer
from iso_sprite_renderer.config import RenderConfig

# Create config
config = RenderConfig()
config.load_defaults()
config.render.size = [64, 64]
config.render.ortho_scale = 2.828427  # sqrt(8)

# Render
renderer = BlenderRenderer(config)
renderer.render('model.obj', 'output_dir/')
```

### Custom Configuration

```python
from iso_sprite_renderer.config import RenderConfig

config = RenderConfig()
config.load_yaml('my_config.yaml')

# Override specific settings
config.render.size = [128, 128]
config.camera.ortho_scale = 4.0
config.lighting.sun_energy = 5.0

# Use config...
```

## Configuration File Format

```yaml
render:
  size: [64, 64]
  directions: 8          # Number of rotation angles
  ortho_scale: 2.828427  # Camera scale
  transparent_bg: true
  no_normalize: false    # Set true to use absolute coordinates
  
  # Initial model rotation (0-3 = 0°, 90°, 180°, 270°)
  rotate_x: 0
  rotate_y: 0
  rotate_z: 0
  
  # Debugging options
  show_ground_plane: false
  show_axes: false
  render_top_view: false
  render_side_view: false

lighting:
  sun_direction: [1, -1, 1]
  sun_color: [1.0, 1.0, 1.0]
  sun_energy: 3.0
  ambient_color: [0.3, 0.3, 0.3]
  ambient_energy: 0.5

camera:
  elevation_angle: 35.264  # Isometric angle
  rotation_offset: 45
  focus_x: 0               # Camera focus point
  focus_y: 0

output:
  directory: "./sprites"
  individual_pngs: true
  sprite_sheet: false
  transparent_bg: true
```

## Coordinate System

- **Ground Plane**: X-Y plane at Z=0
- **Unit Square**: X=[-1,1], Y=[-1,1]
- **Camera Scale**: `ortho_scale = sqrt(8) ≈ 2.828427` fits the unit square perfectly in frame

### Perfect Grid Fit

For grid-based games (e.g., 2×2 unit squares):
```python
config.render.ortho_scale = 2.828427  # sqrt(8)
config.render.no_normalize = True
config.camera.focus_x = 0
config.camera.focus_y = 0
```

This ensures:
- All sprites use the same world scale
- Components align perfectly when composited
- Grid coordinates map 1:1 to game world

## Architecture

```
iso-sprite-renderer/
├── config.py           # Configuration management
├── renderer.py         # Blender orchestration
├── iso_render.py       # CLI entry point
├── iso_render          # Shell wrapper
├── loaders/
│   ├── base.py        # Abstract mesh loader
│   └── obj_loader.py  # OBJ file loader
└── templates/
    └── render_script.py  # Internal Blender script
```

## Extending the Library

### Add New File Format

1. Create loader in `loaders/`:

```python
from .base import MeshLoader

class FBXLoader(MeshLoader):
    def supports(self, filepath: str) -> bool:
        return filepath.lower().endswith('.fbx')
    
    def get_file_extensions(self) -> list[str]:
        return ['.fbx']
    
    def load_into_blender(self, filepath: str, bpy):
        bpy.ops.import_scene.fbx(filepath=filepath)
```

2. Register in `loaders/__init__.py`:

```python
from .fbx_loader import FBXLoader
```

## Tools Using This Library

- **electrical_sprite_database** - Generate layered sprites for electrical grid components (pylons, substations, wires)

## Supported File Formats

- **OBJ** (`.obj`) - Fully supported

Coming soon:
- FBX (`.fbx`)
- GLTF/GLB (`.gltf`, `.glb`)

## Troubleshooting

### Blender Not Found

Ensure Blender is in your PATH:

```bash
blender --version
```

For Flatpak installations:
```bash
ln -s /var/lib/flatpak/exports/bin/org.blender.Blender ~/.local/bin/blender
```

### Sprites Too Small/Large

Adjust `ortho_scale` (lower = more zoomed in):
```bash
--ortho-scale 2.0  # More zoomed in
--ortho-scale 4.0  # More zoomed out
```

### Model Not Grounded

Use debugging options to verify positioning:
```bash
--show-ground-plane --show-axes --render-top-view --render-side-view
```

The model should sit on the X-Y plane (Z=0).

## License

Part of the gowasm-engine project.
