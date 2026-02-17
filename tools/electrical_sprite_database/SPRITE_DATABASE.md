# Electrical Component Sprite Database Generator

A comprehensive system for generating layered sprite assets for isometric electrical grid simulation games.

## Overview

This system generates a complete sprite database with:
- **Base component sprites** (no wires) for each component type
- **Individual wire connection sprites** for compositing in-game
- Support for 3-phase electrical connections (A, B, C)
- All connection configurations within a configurable radius
- Metadata for game engine integration

## Features

### Layered Sprite Architecture

Instead of generating every possible combination of wires (which would be exponential), the system generates:
1. Base component sprites (pylon, substation)
2. Individual wire sprites for each connection

In-game, you composite the base sprite with the relevant wire layers using alpha blending.

### Camera Configuration

All sprites use identical camera settings for perfect alignment:
- **Resolution**: 64x64 pixels
- **Ortho scale**: √8 ≈ 2.828427 (fits 2.0x2.0 grid square perfectly)
- **Elevation**: 35.264° (true isometric)
- **Rotation**: 45°

### Connection Types

For each component type (pylon, substation), generates:
- **Pylon-to-Pylon** connections
- **Pylon-to-Substation** connections
- **Substation-to-Substation** connections

For radius 3, this generates ~28 offsets × 3 phases × 4 connection types = **~340 sprites**.

## Installation

### Prerequisites

```bash
# Blender (must be in PATH or aliased)
blender --version

# Python 3.8+
python3 --version

# PyYAML for configuration
pip install pyyaml

# Optional: Pillow for sprite validation
pip install Pillow
```

## Usage

### Quick Start

Generate the complete sprite database with default settings:

```bash
cd tools/iso-sprite-renderer
python3 generate_sprite_database.py
```

This will:
1. Read configuration from `sprite_database_config.yaml`
2. Generate base sprites for pylon and substation
3. Generate all connection sprites within radius 3
4. Save metadata to `output/sprite_database/metadata.json`
5. Take approximately 11 minutes for ~340 sprites

### Custom Configuration

Edit `sprite_database_config.yaml`:

```yaml
# Change output location
output_dir: "output/my_sprites"

# Adjust sprite size
sprite:
  size: 128  # Larger sprites (128x128)
  
# Reduce connection radius for faster generation
connection_radius: 2  # Only generate up to 2 grid squares away

# Skip existing sprites
rendering:
  overwrite_existing: false
```

Then run:

```bash
python3 generate_sprite_database.py --config sprite_database_config.yaml
```

### Validate Output

Check that all sprites were generated correctly:

```bash
python3 validate_sprite_database.py output/sprite_database
```

This validates:
- All expected sprite files exist
- Sprites have correct dimensions (64x64)
- Sprites have alpha channel for transparency
- No corrupted image files

## Output Structure

```
output/sprite_database/
├── pylon/
│   ├── base.png                                    # Pylon without wires
│   └── connections/
│       ├── to_pylon_x0_y1_A_to_A.png              # Phase A wire to pylon at (0,1)
│       ├── to_pylon_x0_y1_B_to_B.png              # Phase B wire to pylon at (0,1)
│       ├── to_pylon_x0_y1_C_to_C.png              # Phase C wire to pylon at (0,1)
│       ├── to_pylon_x1_y0_A_to_A.png              # Phase A wire to pylon at (1,0)
│       └── ... (all combinations)
├── substation/
│   ├── base.png                                    # Substation without wires
│   └── connections/
│       ├── to_pylon_x0_y1_A_to_A.png
│       └── ...
└── metadata.json                                   # Database metadata
```

### Filename Convention

Connection sprites use descriptive naming:
- `to_{target}_x{offset_x}_y{offset_y}_{phase}_to_{phase}.png`

Examples:
- `to_pylon_x2_y1_A_to_A.png` - Wire from this component to pylon at offset (2,1), phase A
- `to_substation_x0_y3_B_to_B.png` - Wire to substation at (0,3), phase B

## Game Integration

### Loading Sprites

```python
import json
from PIL import Image

# Load metadata
with open('sprite_database/metadata.json', 'r') as f:
    db = json.load(f)

# Load base sprite
pylon_base = Image.open(f"sprite_database/{db['components']['pylon']['base']}")

# Load connection sprite
conn_sprite = Image.open("sprite_database/pylon/connections/to_pylon_x1_y0_A_to_A.png")
```

### Compositing Example

To render a pylon with multiple connections:

```python
def render_component_with_wires(component_type, connections):
    """
    Render component with wire connections.
    
    Args:
        component_type: 'pylon' or 'substation'
        connections: List of (target_type, offset_x, offset_y, phase) tuples
    """
    # Start with base
    result = Image.open(f"sprite_database/{component_type}/base.png").copy()
    
    # Composite each wire
    for target, offset_x, offset_y, phase in connections:
        wire_file = f"sprite_database/{component_type}/connections/to_{target}_x{offset_x}_y{offset_y}_{phase}_to_{phase}.png"
        wire_img = Image.open(wire_file)
        
        # Alpha composite
        result = Image.alpha_composite(result, wire_img)
    
    return result

# Example: Pylon with two connections
sprite = render_component_with_wires('pylon', [
    ('pylon', 1, 0, 'A'),      # Connect to pylon at (1,0) on phase A
    ('substation', 0, 2, 'B')  # Connect to substation at (0,2) on phase B
])
sprite.save('rendered_pylon.png')
```

### WebGPU/WebGL Integration

For the Go WASM game engine:
1. Load all sprites as textures at startup
2. Use sprite atlas or individual textures
3. Render base component first
4. Overlay wire sprites using alpha blending
5. Use depth buffer to handle tile overlap

## Component Configuration

### Terminal Naming

Components have three terminals (3-phase power):
- **Terminal A** (index 0) - Left terminal
- **Terminal B** (index 1) - Center terminal  
- **Terminal C** (index 2) - Right terminal

Wires connect matching phases: A-to-A, B-to-B, C-to-C.

### Grid Coordinates

- Grid squares are 2.0 units wide in world space
- Components placed at integer grid coordinates (0,0), (1,0), (2,1), etc.
- Each component centers within its grid square

## Advanced Usage

### Generate Single Component

Test with just one base component:

```bash
cd tools/iso-sprite-renderer/electrical
blender --background --python generate_electrical.py -- \
    --component pylon --output /tmp/pylon.obj

cd ..
python3 iso_render.py --input /tmp/pylon.obj --output /tmp/pylon_sprite \
    --size 64 --ortho-scale 2.828427 --config examples/single_angle.yaml
```

### Generate Test Scene

Create a custom scene for testing:

```python
# my_test_scene.py
import sys
sys.path.insert(0, 'electrical')

from pylon import Pylon
from substation import Substation
from scene import ComponentScene

def create_scene():
    scene = ComponentScene("TestScene")
    
    # Add two pylons
    pylon1 = Pylon(name="pylon1")
    pylon2 = Pylon(name="pylon2")
    
    scene.add_component(pylon1, 0, 0, "pylon1")
    scene.add_component(pylon2, 2, 1, "pylon2")
    
    # Connect phase A
    scene.connect("pylon1", 0, "pylon2", 0, sag=0.1)
    
    return scene
```

Then render it:

```bash
blender --background --python electrical/generate_electrical.py -- \
    --scene my_test_scene.py --output /tmp/test_scene.obj --wire-only
```

## Performance

### Generation Time

- **Base sprites**: ~2 seconds each × 2 = 4 seconds
- **Connection sprites**: ~2 seconds each × 336 = 672 seconds (~11 minutes)
- **Total**: ~11-12 minutes for radius 3

### Optimization Tips

1. **Start with smaller radius**: Test with `connection_radius: 1` (6 offsets, ~1 minute)
2. **Skip existing sprites**: Set `overwrite_existing: false` to resume interrupted runs
3. **Parallelize**: Generate different component types in parallel (advanced)
4. **Reduce resolution**: Use 32x32 for rapid prototyping

## Troubleshooting

### Blender Not Found

```bash
# Check blender is accessible
which blender

# Or create symlink (Linux)
ln -s /path/to/blender ~/.local/bin/blender
```

### Missing Dependencies

```bash
pip install pyyaml pillow
```

### Empty/Corrupted Sprites

Check Blender output for errors:

```bash
python3 generate_sprite_database.py 2>&1 | tee generation.log
```

### Slow Generation

Reduce radius or sprite count:

```yaml
connection_radius: 2  # Fewer offsets
phases:
  - A  # Only generate phase A for testing
```

## Architecture

### Components

- **[electrical/base_component.py](electrical/base_component.py)** - Abstract base class
- **[electrical/pylon.py](electrical/pylon.py)** - Pylon implementation
- **[electrical/substation.py](electrical/substation.py)** - Substation implementation
- **[electrical/scene.py](electrical/scene.py)** - Scene management with selective rendering
- **[electrical/wire_generator.py](electrical/wire_generator.py)** - Catenary wire generation
- **[electrical/grid_utils.py](electrical/grid_utils.py)** - Grid offset utilities

### Generation Pipeline

1. **generate_sprite_database.py** - Main orchestrator
2. **generate_electrical.py** - Blender scene generation (OBJ export)
3. **iso_render.py** - Blender rendering (OBJ → PNG)
4. **validate_sprite_database.py** - Post-generation validation

## License

MIT License - See main repository LICENSE file.

## Contributing

When adding new component types:
1. Extend `ElectricalComponent` base class
2. Add component to `sprite_database_config.yaml`
3. Update metadata structure if needed


