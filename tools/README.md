# Tools Directory

This directory contains reusable tools and utilities for the gowasm-engine project.

## Structure

### Libraries
- **iso-sprite-renderer/** - Generic isometric sprite renderer library
  - Renders 3D models (OBJ) to isometric 2D sprites
  - Configurable camera, lighting, and output
  - Used by other tools for sprite generation

### Tools
- **electrical_sprite_database/** - Electrical grid component sprite generator
  - Generates layered sprites for pylons, substations, wires
  - Uses `iso-sprite-renderer` library for rendering
  - Outputs complete sprite databases for game integration

### Utilities
- **compare_projections.py** - Projection comparison utility

## Quick Start

### Use iso-sprite-renderer Library

```bash
# Command line
cd tools/iso-sprite-renderer
python3 iso_render.py --input model.obj --output sprites/ --size 64

# As Python library
import sys
sys.path.insert(0, 'tools')
# Then import from iso_sprite_renderer module
```

### Generate Electrical Sprites

```bash
cd tools/electrical_sprite_database
python3 generate_sprite_database.py --config pylon_config.yaml
python3 validate_sprite_database.py output/sprite_database_pylon
```

## Documentation

Each tool/library has its own README:
- `iso-sprite-renderer/README.md` - Library documentation
- `electrical_sprite_database/README.md` - Tool documentation
- `REFACTORING_SUMMARY.md` - Structure explanation

## Adding New Tools

To create a new tool that uses `iso-sprite-renderer`:

1. Create directory: `tools/my_new_tool/`
2. Add to Python path:
   ```python
   import sys
   import os
   TOOLS_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), '..'))
   sys.path.insert(0, TOOLS_DIR)
   ```
3. Call iso-sprite-renderer:
   ```python
   iso_render_path = os.path.join(TOOLS_DIR, 'iso-sprite-renderer', 'iso_render.py')
   subprocess.run([sys.executable, iso_render_path, ...])
   ```

## Requirements

- Python 3.9+
- Blender 3.0+ (for rendering tools)
- PyYAML: `pip install pyyaml`
- Pillow: `pip install pillow`

## License

Part of the gowasm-engine project.
