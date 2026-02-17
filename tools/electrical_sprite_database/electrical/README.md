# Electrical Component System

A class-based system for generating electrical components (substations, transformers, pylons) with grid placement, connection points, and wire generation for isometric game assets.

## Features

- **Component-Based Architecture**: Reusable electrical components with OOP design
- **Grid Positioning**: Integer grid coordinates for easy placement in isometric games
- **Connection Points**: Define attachment points for wires on each component
- **Wire Generation**: Realistic catenary (sagging) wires between connection points
- **Scene Management**: Compose multiple components and connections into complete scenes
- **Blender Integration**: Uses Blender's Python API for 3D geometry generation
- **Isometric Rendering**: Integrates with the existing `iso_render` tool

## Quick Start

### Generate a Single Component

```bash
cd tools/iso-sprite-renderer

# Generate a substation
blender --background --python electrical/generate_electrical.py -- \
    --component substation \
    --output output/electrical/substation.obj

# Render as isometric sprites
./iso_render \
    --input output/electrical/substation.obj \
    --output output/electrical/substation_sprites \
    --size 128
```

### Generate a Scene with Connections

```bash
# Generate a power grid scene with 3 substations and wires
blender --background --python electrical/generate_electrical.py -- \
    --scene examples/power_grid_scene.py \
    --output output/electrical/power_grid.obj

# Render the full scene
./iso_render \
    --input output/electrical/power_grid.obj \
    --output output/electrical/power_grid_sprites \
    --size 256
```

## Architecture

### Grid System

- Each grid square is **2.0 units** (from -1.0 to +1.0 in local space)
- Components occupy integer grid positions: `(0, 0)`, `(1, 0)`, `(2, 0)`, etc.
- World position: `world_pos = (grid_x * 2.0, grid_y * 2.0, 0)`

### Connection Points

Components define connection points as local (x, y, z) offsets from their center:

```python
# Substation has 3 terminals on top
self.add_connection_point(-0.22, 0, 0.95, "Left Terminal")
self.add_connection_point(0, 0, 0.95, "Center Terminal")
self.add_connection_point(0.22, 0, 0.95, "Right Terminal")
```

World positions are calculated automatically based on grid placement.

### Components

#### Substation

- **Footprint**: 1.8 x 1.8 units (fits in one grid square)
- **Height**: 0.9 units
- **Features**: Cubic body with 3 cylindrical terminals on top
- **Connection Points**: 3 (one per terminal)

Example usage:
```python
from electrical.substation import Substation

sub = Substation(name="MySubstation")
sub.set_grid_position(0, 0)
```

## Creating Custom Components

### 1. Subclass ElectricalComponent

```python
from electrical.base_component import ElectricalComponent

class MyComponent(ElectricalComponent):
    def __init__(self, name="MyComponent"):
        super().__init__(name, footprint_size=0.8)
        
        # Define connection points
        self.add_connection_point(0, 0, 1.0, "Top Connection")
    
    def generate_geometry(self, bpy):
        # Create geometry using Blender primitives
        bpy.ops.mesh.primitive_cube_add(size=0.8, location=(0, 0, 0.4))
        obj = bpy.context.active_object
        obj.name = self.name
        return obj
```

### 2. Use in Scenes

```python
from electrical.scene import ComponentScene
from my_component import MyComponent

scene = ComponentScene()
comp = MyComponent()
scene.add_component(comp, grid_x=0, grid_y=0)
```

## Creating Scene Scripts

Create a Python file in `examples/` that defines a `create_scene()` function:

```python
# examples/my_scene.py
from electrical.scene import ComponentScene
from electrical.substation import Substation

def create_scene():
    scene = ComponentScene(name="MyScene")
    
    # Add components
    sub1 = Substation(name="Sub1")
    sub2 = Substation(name="Sub2")
    
    scene.add_component(sub1, grid_x=0, grid_y=0, component_id="sub1")
    scene.add_component(sub2, grid_x=2, grid_y=0, component_id="sub2")
    
    # Connect them
    scene.connect(
        comp1_id='sub1',
        point1_idx=2,  # Right terminal
        comp2_id='sub2',
        point2_idx=0,  # Left terminal
        sag=0.15
    )
    
    return scene
```

Run with:
```bash
blender --background --python electrical/generate_electrical.py -- \
    --scene examples/my_scene.py \
    --output output/my_scene.obj
```

## Wire Generation

Wires are automatically generated with realistic catenary (sagging) curves:

```python
scene.connect(
    comp1_id='sub1',
    point1_idx=0,
    comp2_id='sub2',
    point2_idx=0,
    sag=0.15,  # Amount of droop (0.0 = straight, higher = more sag)
    wire_name="PowerLine"
)
```

The wire generator:
- Calculates a parabolic catenary curve
- Creates a NURBS curve in Blender
- Adds thickness with circular cross-section
- Converts to mesh for export

## File Structure

```
electrical/
├── __init__.py                  # Module exports
├── base_component.py            # ElectricalComponent base class
├── substation.py                # Substation component
├── wire_generator.py            # Wire generation logic
├── scene.py                     # ComponentScene manager
└── generate_electrical.py       # CLI tool

examples/
└── power_grid_scene.py          # Example scene definition

output/electrical/
├── substation_component.obj     # Generated component
├── power_grid.obj               # Generated scene
├── substation_sprites/          # Rendered isometric sprites
└── power_grid_sprites/          # Rendered scene sprites
```

## CLI Reference

### generate_electrical.py

```bash
blender --background --python electrical/generate_electrical.py -- [OPTIONS]

Options:
  --component TYPE      Generate single component (e.g., 'substation')
  --output PATH         Output OBJ file path
  --grid-x X           Grid X position (default: 0)
  --grid-y Y           Grid Y position (default: 0)
  --scene PATH         Generate scene from Python script
```

### Integration with iso_render

All generated OBJ files work seamlessly with the isometric renderer:

```bash
./iso_render \
    --input output/electrical/component.obj \
    --output sprites/ \
    --size 128 \
    --show-ground-plane \
    --show-axes \
    --render-side-view
```

## Examples

### Single Component
```bash
# Generate and render a substation
blender --background --python electrical/generate_electrical.py -- \
    --component substation --output output/sub.obj

./iso_render --input output/sub.obj --output sprites/sub --size 128
```

### Power Grid Scene
```bash
# Generate scene with 3 substations and 3 wires
blender --background --python electrical/generate_electrical.py -- \
    --scene examples/power_grid_scene.py \
    --output output/grid.obj

# Render with larger size for full scene
./iso_render --input output/grid.obj --output sprites/grid --size 256
```

## Future Extensions

Planned component types:
- **Transformer**: Smaller component with 2 connection points
- **Pylon**: Tall structure (height 2.0) with 3-6 connection points
- **Generator**: Power source component
- **Switch**: Control component with input/output connections

## Technical Details

- **Grid Size**: 2.0 units per square
- **Component Footprint**: Typically 0.6 - 0.9 (fits within grid square)
- **Wire Thickness**: Default 0.02 radius
- **Wire Segments**: 16 segments for smooth curves
- **Connection Height**: Usually 0.9 - 1.0 units (top of components)

## Notes

- Components are centered at their grid position
- Connection points use absolute Z coordinates (height from ground)
- Wires use parabolic approximation of catenary curves
- All geometry uses Blender primitives for simplicity
- Export format is OBJ (compatible with most 3D tools)



