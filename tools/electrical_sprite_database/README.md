# Electrical Sprite Database Generator

Generate layered sprite databases for electrical grid components in isometric tycoon-style games.

## Overview

This tool generates **complete sprite databases** for electrical components (pylons, substations) including:
1. **Base component sprites** - Components without wires
2. **Wire connection sprites** - Individual transparent wire layers for in-game compositing

Instead of pre-rendering every possible combination (exponential), this approach generates separate layers that can be dynamically composited in-game.

## Requirements

- Python 3.9+
- Blender 3.0+ (accessible from command line)
- PyYAML (`pip install pyyaml`)
- Pillow (`pip install pillow`)
- `iso-sprite-renderer` library (in `../iso-sprite-renderer/`)

## Installation

Ensure dependencies are installed:

```bash
pip install pyyaml pillow
```

Make sure Blender is accessible:

```bash
blender --version
```

## Quick Start

### Generate Pylon Sprites Only (Fast Test)

```bash
cd tools/electrical_sprite_database
python3 generate_sprite_database.py --config pylon_config.yaml
```

Generates:
- 1 base pylon sprite (64×64 px)
- 30 wire connection sprites (10 offsets × 3 phases)
- `metadata.json` with sprite information

Output: `output/sprite_database_pylon/`

### Generate Complete Database

```bash
python3 generate_sprite_database.py --config sprite_database_config.yaml
```

Generates:
- Base sprites for all component types
- Wire connections for all combinations
- ~340 total sprites

Output: `output/sprite_database/`

### Validate Generated Sprites

```bash
python3 validate_sprite_database.py output/sprite_database_pylon
```

Checks:
- All expected sprites exist
- Correct dimensions (64×64)
- Proper transparency/alpha channel

## Components

### Pylon
Three-phase transmission tower with:
- Central vertical pole
- Horizontal crossbar
- Three terminals (A, B, C) for wire connections

### Substation
Transformer station with:
- Cube base
- Three top terminals (A, B, C)

### Wire Generator
3D catenary curves with:
- Realistic sagging
- Phase-specific connections (A-to-A, B-to-B, C-to-C)
- Transparent rendering for compositing

## Configuration

### Main Config (`sprite_database_config.yaml`)

```yaml
output_dir: "output/sprite_database"

sprite:
  size: 64                    # Output resolution
  ortho_scale: 2.828427       # sqrt(8) for perfect grid fit

camera:
  elevation: 35.264           # Isometric angle
  rotation: 45.0

components:
  - pylon
  - substation

connection_radius: 3          # Max grid distance for connections

phases:
  - A
  - B
  - C

connection_types:             # All valid connection pairs
  - [pylon, pylon]
  - [pylon, substation]
  - [substation, pylon]
  - [substation, substation]

wire:
  sag: 0.1                    # Wire droop amount
  radius: 0.02

rendering:
  generate_base: true
  generate_connections: true
  overwrite_existing: false   # Set false to resume
```

### Test Config (`pylon_config.yaml`)

Faster config for testing - only generates pylon sprites:

```yaml
output_dir: "output/sprite_database_pylon"
sprite:
  size: 64
  ortho_scale: 2.828427
components:
  - pylon
connection_radius: 3
```

## Output Structure

```
output/sprite_database_pylon/
├── pylon/
│   ├── base.png                           # Component without wires
│   └── connections/
│       ├── to_pylon_x0_y1_A_to_A.png     # Wire to (0,1) phase A
│       ├── to_pylon_x0_y1_B_to_B.png     # Wire to (0,1) phase B
│       ├── to_pylon_x0_y1_C_to_C.png     # Wire to (0,1) phase C
│       ├── to_pylon_x1_y0_A_to_A.png     # Wire to (1,0) phase A
│       └── ... (30 total connection sprites)
└── metadata.json                          # Sprite database metadata
```

## Metadata Format

```json
{
  "version": "1.0",
  "sprite_size": 64,
  "camera": {
    "ortho_scale": 2.828427,
    "elevation": 35.264,
    "rotation": 45.0
  },
  "components": {
    "pylon": {
      "base": "pylon/base.png",
      "connection_points": ["A", "B", "C"],
      "connections": [
        {
          "target": "pylon",
          "offset": [0, 1],
          "phase": "A",
          "file": "pylon/connections/to_pylon_x0_y1_A_to_A.png"
        }
      ]
    }
  }
}
```

## Usage Examples

### Example 1: Generate Test Database

```bash
# Quick pylon-only generation for testing
python3 generate_sprite_database.py --config pylon_config.yaml

# Validate output
python3 validate_sprite_database.py output/sprite_database_pylon
```

### Example 2: Generate Complete Database

```bash
# All components, all connections (~340 sprites, ~12 minutes)
python3 generate_sprite_database.py --config sprite_database_config.yaml
```

### Example 3: Resume Interrupted Generation

If generation is interrupted:

```bash
# Config already has overwrite_existing: false
# Just rerun - it will skip existing sprites
python3 generate_sprite_database.py --config pylon_config.yaml
```

### Example 4: Create Custom Scene

```python
# examples/custom_grid.py
from electrical.pylon import Pylon
from electrical.substation import Substation
from electrical.scene import ComponentScene

def create_scene():
    scene = ComponentScene("my_grid")
    
    # Add components at grid positions
    pylon1 = Pylon(name="P1")
    pylon2 = Pylon(name="P2")
    sub = Substation(name="S1")
    
    scene.add_component(pylon1, x=0, y=0, id="P1")
    scene.add_component(pylon2, x=2, y=0, id="P2")
    scene.add_component(sub, x=4, y=0, id="S1")
    
    # Connect phase A: P1 -> P2 -> S1
    scene.connect("P1", 0, "P2", 0, sag=0.1)  # Terminal A
    scene.connect("P2", 0, "S1", 0, sag=0.1)  # Terminal A
    
    return scene
```

Render it:

```bash
python3 electrical/generate_electrical.py \
  --scene examples/custom_grid.py \
  --output output/my_grid.obj

python3 ../iso-sprite-renderer/iso_render.py \
  --input output/my_grid.obj \
  --output output/my_grid_sprites \
  --size 64 \
  --ortho-scale 2.828427
```

## Game Integration

### Load Metadata

```python
import json
from PIL import Image

# Load sprite database
with open('output/sprite_database_pylon/metadata.json') as f:
    db = json.load(f)

# Access component info
pylon_base = db['components']['pylon']['base']
connections = db['components']['pylon']['connections']
```

### Composite Sprites

```python
from PIL import Image

# Example: Pylon at (0,0) connected to pylon at (1,0) on phase A
base_path = 'output/sprite_database_pylon/pylon/base.png'
wire_path = 'output/sprite_database_pylon/pylon/connections/to_pylon_x1_y0_A_to_A.png'

base = Image.open(base_path)
wire = Image.open(wire_path)

# Alpha composite
result = Image.alpha_composite(base, wire)
result.save('rendered_pylon.png')
```

### In WebGPU Game Engine

```go
// Load sprites as textures
baseTexture := LoadTexture("pylon/base.png")
wireTexture := LoadTexture("pylon/connections/to_pylon_x1_y0_A_to_A.png")

// Render with alpha blending
DrawSprite(baseTexture, position, size)      // Base layer
DrawSprite(wireTexture, position, size)      // Wire layer (alpha blended)
```

## Performance

- **Per sprite**: ~2 seconds (scene generation + Blender render)
- **31 sprites** (pylon test): ~62 seconds total
- **340 sprites** (full database): ~11-12 minutes

## Architecture

```
electrical_sprite_database/
├── electrical/                    # Component system
│   ├── base_component.py         # Abstract component class
│   ├── pylon.py                  # Pylon implementation
│   ├── substation.py             # Substation implementation
│   ├── wire_generator.py         # Catenary wire generation
│   ├── scene.py                  # Scene management
│   ├── grid_utils.py             # Grid utilities
│   └── generate_electrical.py   # Scene -> OBJ generator
├── examples/                      # Example scenes
│   ├── power_grid_scene.py
│   ├── transmission_line_scene.py
│   ├── two_pylons_scene.py
│   └── two_pylons_short.py
├── generate_sprite_database.py   # Main sprite generator
├── validate_sprite_database.py   # Validation tool
├── sprite_database_config.yaml   # Full database config
├── pylon_config.yaml             # Pylon-only test config
├── render_config.yaml            # Renderer config
├── SPRITE_DATABASE.md            # Detailed documentation
└── README.md                     # This file
```

## Extending the System

### Add New Component Type

1. Create `electrical/transformer.py`:

```python
from electrical.base_component import ElectricalComponent

class Transformer(ElectricalComponent):
    def __init__(self, name="transformer"):
        super().__init__(name)
        
        # Add connection terminals
        self.add_connection_point("A", x=-0.3, y=0, z=1.0)
        self.add_connection_point("B", x=0, y=0, z=1.0)
        self.add_connection_point("C", x=0.3, y=0, z=1.0)
    
    def generate_geometry(self):
        # Return Blender-compatible geometry
        return [
            # ... define 3D geometry ...
        ]
```

2. Update `sprite_database_config.yaml`:

```yaml
components:
  - pylon
  - substation
  - transformer  # Add new component

connection_types:
  # ... existing types ...
  - [pylon, transformer]
  - [transformer, substation]
```

3. Regenerate database:

```bash
python3 generate_sprite_database.py --config sprite_database_config.yaml
```

## Troubleshooting

### Blender Not Found

```bash
# Check PATH
blender --version

# For Flatpak installations
ln -s /var/lib/flatpak/exports/bin/org.blender.Blender ~/.local/bin/blender
```

### Generation Timeout

```bash
# Set overwrite_existing: false in config
# Then resume - skips completed sprites
python3 generate_sprite_database.py --config pylon_config.yaml
```

### Connection Points Misaligned

Check component geometry. Use debugging to visualize:

```bash
python3 electrical/generate_electrical.py \
  --component pylon \
  --output /tmp/pylon.obj

python3 ../iso-sprite-renderer/iso_render.py \
  --input /tmp/pylon.obj \
  --show-axes \
  --render-top-view \
  --render-side-view
```

### Wires Appear Cut Off

Increase camera `ortho_scale` or adjust component heights. Default `sqrt(8)` fits 2×2 grid square perfectly.

## Documentation

- **README.md** (this file) - Overview and usage
- **SPRITE_DATABASE.md** - Complete technical guide
- **electrical/README.md** - Component system details

## License

Part of the gowasm-engine project.

