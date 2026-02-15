"""
Blender render script template.
This script is executed inside Blender to render isometric sprites.
Configuration is injected as JSON.
"""
import bpy
import math
import json
import sys
import os

# Configuration injected by the renderer
CONFIG_JSON = '''__CONFIG_JSON__'''

try:
    config = json.loads(CONFIG_JSON)
except json.JSONDecodeError as e:
    print(f"ERROR: Failed to parse config JSON: {e}")
    sys.exit(1)


def clear_scene():
    """Remove all objects, lights, and cameras from the scene."""
    bpy.ops.object.select_all(action='SELECT')
    bpy.ops.object.delete()


def setup_render_settings():
    """Configure render engine and output settings."""
    scene = bpy.context.scene
    
    # Use EEVEE for faster rendering (can switch to CYCLES for better quality)
    scene.render.engine = 'BLENDER_EEVEE'
    
    # Set resolution
    size = config['render']['size']
    scene.render.resolution_x = size[0]
    scene.render.resolution_y = size[1]
    scene.render.resolution_percentage = 100
    
    # PNG output
    scene.render.image_settings.file_format = 'PNG'
    scene.render.image_settings.color_mode = 'RGBA'
    
    # Transparent background
    if config['output']['transparent_bg']:
        scene.render.film_transparent = True
    
    # EEVEE settings for better quality (compatible with Blender 5.0+)
    try:
        # Try Blender 4.x attribute names
        scene.eevee.use_gtao = True
        scene.eevee.use_bloom = False
        scene.eevee.use_ssr = False
    except AttributeError:
        # Blender 5.0+ changed attribute names, skip if not available
        pass


def setup_lighting():
    """Add lighting to the scene."""
    lighting = config['lighting']
    
    # Add sun light
    bpy.ops.object.light_add(type='SUN')
    sun = bpy.context.active_object
    sun.name = "Sun"
    
    # Set sun direction
    direction = lighting['sun_direction']
    # Convert direction vector to rotation
    # Point the sun in the specified direction
    length = math.sqrt(sum(d*d for d in direction))
    if length > 0:
        direction = [d/length for d in direction]
    
    # Calculate rotation from direction
    sun.rotation_euler = (
        math.atan2(direction[2], math.sqrt(direction[0]**2 + direction[1]**2)),
        0,
        math.atan2(direction[1], direction[0])
    )
    
    # Set sun properties
    sun.data.energy = lighting['sun_energy']
    sun.data.color = tuple(lighting['sun_color'])
    
    # Add ambient light (use world settings)
    world = bpy.context.scene.world
    world.use_nodes = True
    bg_node = world.node_tree.nodes.get('Background')
    if bg_node:
        bg_node.inputs['Color'].default_value = tuple(lighting['ambient_color']) + (1.0,)
        bg_node.inputs['Strength'].default_value = lighting['ambient_energy']


def setup_camera():
    """Setup orthographic camera for isometric view."""
    camera_config = config['camera']
    render_config = config['render']
    
    # Add camera
    bpy.ops.object.camera_add()
    camera = bpy.context.active_object
    camera.name = "IsometricCamera"
    
    # Set to orthographic
    camera.data.type = 'ORTHO'
    camera.data.ortho_scale = render_config['ortho_scale']
    
    # Position camera for isometric view
    elevation = math.radians(camera_config['elevation_angle'])
    rotation_offset = math.radians(camera_config['rotation_offset'])
    
    distance = 10  # Distance from origin
    
    # Calculate camera position
    camera.location = (
        distance * math.cos(elevation) * math.cos(rotation_offset),
        distance * math.cos(elevation) * math.sin(rotation_offset),
        distance * math.sin(elevation)
    )
    
    # Point camera at origin
    from mathutils import Vector
    direction = Vector((-camera.location[0], -camera.location[1], -camera.location[2]))
    rot_quat = direction.to_track_quat('-Z', 'Y')
    camera.rotation_euler = rot_quat.to_euler()
    
    # Set as active camera
    bpy.context.scene.camera = camera
    
    return camera


def create_ground_plane():
    """Create a reference square on the ground plane."""
    render_config = config['render']
    if not render_config.get('show_ground_plane', False):
        return None
    
    # Create a curve object for the border
    curve_data = bpy.data.curves.new(name='GroundPlane', type='CURVE')
    curve_data.dimensions = '3D'
    curve_data.bevel_depth = 0.02  # Border thickness (0.2 units as requested)
    
    # Create a polyline (spline)
    polyline = curve_data.splines.new('POLY')
    
    # Define the square corners [-1, 1] x [-1, 1] at Z=0
    corners = [
        (-1.0, -1.0, 0.0),
        ( 1.0, -1.0, 0.0),
        ( 1.0,  1.0, 0.0),
        (-1.0,  1.0, 0.0),
    ]
    
    # Polyline starts with 1 point, add 3 more to get 4 total
    polyline.points.add(len(corners) - 1)
    
    for i, corner in enumerate(corners):
        x, y, z = corner
        polyline.points[i].co = (x, y, z, 1.0)  # (x, y, z, w) for homogeneous coords
    
    # Close the loop
    polyline.use_cyclic_u = True
    
    # Create object and link to scene
    curve_obj = bpy.data.objects.new('GroundPlane', curve_data)
    bpy.context.collection.objects.link(curve_obj)
    
    # Create a simple material for visibility
    mat = bpy.data.materials.new(name="GroundPlaneMaterial")
    mat.diffuse_color = (1.0, 1.0, 1.0, 1.0)  # White color
    curve_obj.data.materials.append(mat)
    
    print(f"✓ Created ground plane reference square [-1,1] x [-1,1]")
    return curve_obj


def load_mesh():
    """Load the mesh file."""
    mesh_path = config['mesh_path']
    loader_type = config['loader_type']
    
    if loader_type == 'obj':
        # Get objects before import
        objects_before = set(bpy.context.scene.objects)
        
        # Import OBJ (Blender 5.0+ uses new importer)
        try:
            # Try Blender 5.0+ new OBJ importer
            bpy.ops.wm.obj_import(filepath=mesh_path)
        except AttributeError:
            # Fall back to old importer for Blender < 5.0
            bpy.ops.import_scene.obj(filepath=mesh_path)
        
        # Find newly imported objects
        objects_after = set(bpy.context.scene.objects)
        new_objects = objects_after - objects_before
        
        if not new_objects:
            raise RuntimeError(f"No objects imported from {mesh_path}")
        
        # Join multiple objects if needed
        if len(new_objects) > 1:
            bpy.ops.object.select_all(action='DESELECT')
            for obj in new_objects:
                obj.select_set(True)
            obj = list(new_objects)[0]
            bpy.context.view_layer.objects.active = obj
            bpy.ops.object.join()
        else:
            obj = list(new_objects)[0]
        
        # Normalize object
        normalize_object(obj)
        
        return obj
    else:
        raise ValueError(f"Unsupported loader type: {loader_type}")


def normalize_object(obj):
    """Center and scale object to fit in view."""
    # Select only this object
    bpy.ops.object.select_all(action='DESELECT')
    obj.select_set(True)
    bpy.context.view_layer.objects.active = obj
    
    # Apply initial rotation adjustments (90° increments)
    render_config = config['render']
    rotate_x = render_config.get('rotate_x', 0) * 90  # Convert to degrees
    rotate_y = render_config.get('rotate_y', 0) * 90
    rotate_z = render_config.get('rotate_z', 0) * 90
    
    obj.rotation_euler = (
        math.radians(rotate_x),
        math.radians(rotate_y),
        math.radians(rotate_z)
    )
    
    # Apply rotation
    bpy.ops.object.transform_apply(location=False, rotation=True, scale=False)
    
    # Reset location for fresh positioning
    obj.location = (0, 0, 0)
    
    # Calculate bounding box and scale based on X-Y plane only
    # This ensures the model fills the isometric vie    w horizontally
    dimensions = obj.dimensions
    xy_max = max(dimensions.x, dimensions.y) if max(dimensions.x, dimensions.y) > 0 else 1.0
    scale_factor = 2.0 / xy_max  # Scale so widest X-Y extent is 2.0 units
    obj.scale = (scale_factor, scale_factor, scale_factor)
    
    # Apply transforms
    bpy.ops.object.transform_apply(location=False, rotation=False, scale=True)
    
    # Center to origin (ground plane at Z=0)
    bpy.ops.object.origin_set(type='ORIGIN_GEOMETRY', center='BOUNDS')
    
    # Center X and Y, but keep bottom at Z=0 (sitting on ground)
    bbox = [obj.matrix_world @ v.co for v in obj.data.vertices]
    min_z = min(v.z for v in bbox)
    obj.location = (0, 0, -min_z)  # Shift so bottom touches ground


def render_rotation(obj, angle_degrees, output_path):
    """Render object at specific rotation."""
    # Rotate object around Z axis
    obj.rotation_euler = (0, 0, math.radians(angle_degrees))
    
    # Set output path
    bpy.context.scene.render.filepath = output_path
    
    # Render
    bpy.ops.render.render(write_still=True)
    
    print(f"✓ Rendered angle {angle_degrees}° -> {output_path}")


def main():
    """Main rendering function."""
    print("=" * 60)
    print("Isometric Sprite Renderer - Blender Script")
    print("=" * 60)
    
    # Clear scene
    print("Clearing scene...")
    clear_scene()
    
    # Setup
    print("Setting up render settings...")
    setup_render_settings()
    
    print("Setting up lighting...")
    setup_lighting()
    
    print("Setting up camera...")
    setup_camera()
    
    print("Loading mesh...")
    obj = load_mesh()
    print(f"✓ Loaded mesh: {obj.name}")
    
    # Create ground plane reference if requested
    create_ground_plane()
    
    # Render all directions
    directions = config['render']['directions']
    output_dir = config['output']['directory']
    model_name = config['model_name']
    
    print(f"\nRendering {directions} directions...")
    angle_step = 360 / directions
    
    for i in range(directions):
        angle = i * angle_step
        output_path = os.path.join(output_dir, f"{model_name}_angle_{int(angle)}.png")
        render_rotation(obj, angle, output_path)
    
    print("\n" + "=" * 60)
    print(f"✓✓✓ Rendering complete! {directions} sprites generated.")
    print("=" * 60)


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        print(f"\n❌ ERROR in Blender script: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)

