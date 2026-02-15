#!/usr/bin/env python3
"""
Render comparison: PERSPECTIVE vs ORTHOGRAPHIC projection
Shows the difference for isometric game rendering
"""
import bpy
import math

def clear_scene():
    """Remove all objects from the scene"""
    bpy.ops.object.select_all(action='SELECT')
    bpy.ops.object.delete()

def create_test_scene():
    """Create a scene with multiple cubes at different depths"""
    # Ground cubes in a grid
    for x in range(-2, 3):
        for y in range(-2, 3):
            bpy.ops.mesh.primitive_cube_add(location=(x * 2, y * 2, 0), size=1.8)
    
    # Stack some cubes to show height
    bpy.ops.mesh.primitive_cube_add(location=(0, 0, 1), size=1.8)
    bpy.ops.mesh.primitive_cube_add(location=(0, 0, 2), size=1.8)
    bpy.ops.mesh.primitive_cube_add(location=(2, 2, 1), size=1.8)
    
    print(f"✓ Created grid of cubes")

def setup_lighting():
    """Add lighting to the scene"""
    bpy.ops.object.light_add(type='SUN', location=(10, -10, 15))
    light = bpy.context.active_object
    light.data.energy = 3.0
    light.rotation_euler = (math.radians(45), 0, math.radians(45))
    print(f"✓ Sun light added")

def setup_render_settings():
    """Configure render output"""
    scene = bpy.context.scene
    scene.render.engine = 'BLENDER_EEVEE'
    scene.render.resolution_x = 512
    scene.render.resolution_y = 512
    scene.render.image_settings.file_format = 'PNG'
    scene.render.film_transparent = True
    print(f"✓ Render settings configured (512x512)")

def render_with_camera_type(camera_type, output_path):
    """Render scene with specified camera type"""
    # Add camera
    bpy.ops.object.camera_add()
    camera = bpy.context.active_object
    
    # Position for isometric view (45° rotation, ~35° elevation)
    distance = 15
    angle_h = math.radians(45)  # Horizontal rotation
    angle_v = math.radians(35)  # Vertical angle
    
    camera.location = (
        distance * math.cos(angle_v) * math.cos(angle_h),
        distance * math.cos(angle_v) * math.sin(angle_h),
        distance * math.sin(angle_v)
    )
    
    # Point camera at origin
    direction = -camera.location
    rot_quat = direction.to_track_quat('-Z', 'Y')
    camera.rotation_euler = rot_quat.to_euler()
    
    # Set camera type
    if camera_type == 'ORTHO':
        camera.data.type = 'ORTHO'
        camera.data.ortho_scale = 12.0
        print(f"✓ Camera set to ORTHOGRAPHIC (ortho_scale=12.0)")
    else:
        camera.data.type = 'PERSP'
        camera.data.lens = 50  # 50mm lens (standard)
        print(f"✓ Camera set to PERSPECTIVE (lens=50mm)")
    
    # Set as active camera
    bpy.context.scene.camera = camera
    
    # Render
    bpy.context.scene.render.filepath = output_path
    bpy.ops.render.render(write_still=True)
    print(f"✓✓ Rendered to: {output_path}")
    
    # Remove camera for next render
    bpy.ops.object.select_all(action='DESELECT')
    camera.select_set(True)
    bpy.ops.object.delete()

# Main execution
print("=" * 60)
print("PERSPECTIVE vs ORTHOGRAPHIC Comparison")
print("=" * 60)

clear_scene()
create_test_scene()
setup_lighting()
setup_render_settings()

print("\n--- Rendering PERSPECTIVE projection ---")
render_with_camera_type('PERSP', '/tmp/blender_perspective.png')

print("\n--- Rendering ORTHOGRAPHIC projection ---")
render_with_camera_type('ORTHO', '/tmp/blender_orthographic.png')

print("\n" + "=" * 60)
print("✓✓✓ COMPARISON COMPLETE!")
print("=" * 60)
print(f"PERSPECTIVE:  /tmp/blender_perspective.png")
print(f"ORTHOGRAPHIC: /tmp/blender_orthographic.png")
print("\nKey differences:")
print("  • PERSPECTIVE: Objects get smaller with distance (realistic)")
print("  • ORTHOGRAPHIC: Objects stay same size (perfect for isometric games!)")
print("=" * 60)

