#!/usr/bin/env python3
"""
CLI tool for generating electrical components and scenes.

Usage:
    # Single component
    blender --background --python generate_electrical.py -- \\
        --component substation --output output/substation.obj
    
    # Full scene from Python script
    blender --background --python generate_electrical.py -- \\
        --scene examples/power_grid_scene.py --output output/power_grid.obj
"""

import bpy
import sys
import os
import importlib.util

# Add parent directory to path for imports
script_dir = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, script_dir)

from base_component import ElectricalComponent
from scene import ComponentScene
from substation import Substation
from pylon import Pylon


def clear_scene():
    """Remove all default objects from scene."""
    bpy.ops.object.select_all(action='SELECT')
    bpy.ops.object.delete()


def parse_args():
    """Parse command line arguments after '--'."""
    try:
        separator_idx = sys.argv.index('--')
        args = sys.argv[separator_idx + 1:]
    except (ValueError, IndexError):
        args = []
    
    # Simple argument parsing
    parsed = {
        'component': None,
        'scene': None,
        'wire_only': False,
        'base_only': False,
        'output': 'output.obj',
        'grid_x': 0,
        'grid_y': 0,
    }
    
    i = 0
    while i < len(args):
        arg = args[i]
        
        if arg == '--component' and i + 1 < len(args):
            parsed['component'] = args[i + 1]
            i += 2
        elif arg == '--scene' and i + 1 < len(args):
            parsed['scene'] = args[i + 1]
            i += 2
        elif arg == '--output' and i + 1 < len(args):
            parsed['output'] = args[i + 1]
            i += 2
        elif arg == '--grid-x' and i + 1 < len(args):
            parsed['grid_x'] = int(args[i + 1])
            i += 2
        elif arg == '--grid-y' and i + 1 < len(args):
            parsed['grid_y'] = int(args[i + 1])
            i += 2
        elif arg == '--wire-only':
            parsed['wire_only'] = True
            i += 1
        elif arg == '--base-only':
            parsed['base_only'] = True
            i += 1
        else:
            i += 1
    
    return parsed


def create_component(component_type: str) -> ElectricalComponent:
    """
    Create a component by type name.
    
    Args:
        component_type: Component type ('substation', 'transformer', 'pylon')
    
    Returns:
        ElectricalComponent instance
    """
    component_type = component_type.lower()
    
    if component_type == 'substation':
        return Substation()
    elif component_type == 'pylon':
        return Pylon()
    else:
        raise ValueError(f"Unknown component type: {component_type}")


def load_scene_from_script(script_path: str) -> ComponentScene:
    """
    Load a scene from a Python script.
    
    The script should define a function `create_scene()` that returns a ComponentScene.
    
    Args:
        script_path: Path to Python script
    
    Returns:
        ComponentScene instance
    """
    if not os.path.exists(script_path):
        raise FileNotFoundError(f"Scene script not found: {script_path}")
    
    # Load the module
    spec = importlib.util.spec_from_file_location("scene_module", script_path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    
    # Call create_scene() function
    if not hasattr(module, 'create_scene'):
        raise ValueError(f"Script {script_path} must define a create_scene() function")
    
    scene = module.create_scene()
    if not isinstance(scene, ComponentScene):
        raise ValueError(f"create_scene() must return a ComponentScene instance")
    
    return scene


def main():
    print("=" * 60)
    print("Electrical Component Generator")
    print("=" * 60)
    
    args = parse_args()
    
    # Clear default scene
    clear_scene()
    
    # Determine output path
    output_path = args['output']
    output_dir = os.path.dirname(output_path)
    if output_dir and not os.path.exists(output_dir):
        os.makedirs(output_dir)
    
    print(f"Output: {output_path}\n")
    
    # Generate based on mode
    if args['scene']:
        # Scene mode: load and generate entire scene
        print(f"Loading scene from: {args['scene']}")
        scene = load_scene_from_script(args['scene'])
        
        if args['wire_only']:
            # Generate only wires
            print("Mode: Wire-only rendering")
            scene.generate_wires_only(bpy)
        elif args['base_only']:
            # Generate only components
            print("Mode: Base components only")
            scene.generate_components_only(bpy)
        else:
            # Generate full scene
            print("Mode: Full scene (components + wires)")
            scene.generate_full_scene(bpy)
        
        scene.export_scene(bpy, output_path)
        
    elif args['component']:
        # Single component mode
        print(f"Creating component: {args['component']}")
        component = create_component(args['component'])
        component.set_grid_position(args['grid_x'], args['grid_y'])
        
        # Generate geometry
        obj = component.generate_geometry(bpy)
        world_pos = component.get_world_position()
        obj.location = world_pos
        component._blender_object = obj
        
        print(f"✓ Generated at grid ({args['grid_x']}, {args['grid_y']})")
        print(f"  World position: {world_pos}")
        print(f"  Connection points: {len(component.connection_points)}")
        for i, point in enumerate(component.connection_points):
            world_conn_pos = component.get_world_connection_position(i)
            print(f"    [{i}] {point.name}: {world_conn_pos}")
        
        # Export
        component.export_obj(bpy, output_path)
    
    else:
        print("ERROR: Must specify either --component or --scene")
        print("\nExamples:")
        print("  blender --background --python generate_electrical.py -- \\")
        print("      --component substation --output output/sub.obj")
        print()
        print("  blender --background --python generate_electrical.py -- \\")
        print("      --scene examples/power_grid_scene.py --output output/grid.obj")
        sys.exit(1)
    
    print("\n" + "=" * 60)
    print("✓✓✓ Complete!")
    print("=" * 60)


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        print(f"\n❌ ERROR: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)

