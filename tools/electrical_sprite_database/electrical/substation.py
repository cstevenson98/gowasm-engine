"""
Substation electrical component.

A simple substation with a cubic body and three terminal connection points.
"""

import math
from typing import Any

try:
    from .base_component import ElectricalComponent
except ImportError:
    from base_component import ElectricalComponent


class Substation(ElectricalComponent):
    """
    Electrical substation component.
    
    Features:
    - Cubic body (1.8 x 1.8 x 0.9 units)
    - Three cylindrical terminals on top
    - Three connection points at terminal centers
    """
    
    def __init__(
        self,
        name: str = "Substation",
        body_size: float = 0.9,
        terminal_radius: float = 0.05,
        terminal_height: float = 0.1,
        terminal_spacing: float = 0.06
    ):
        """
        Initialize substation.
        
        Args:
            name: Component name
            body_size: Body footprint size (fits in grid square)
            terminal_radius: Radius of terminal cylinders
            terminal_height: Height of terminal cylinders
            terminal_spacing: Space between terminals
        """
        super().__init__(name, footprint_size=body_size)
        
        self.body_size = body_size
        self.terminal_radius = terminal_radius
        self.terminal_height = terminal_height
        self.terminal_spacing = terminal_spacing
        
        # Define connection points at terminal centers
        # Body height is 0.9, terminals sit on top
        body_height = 0.9
        connection_z = body_height + terminal_height / 2
        
        # Calculate terminal X positions (centered, spaced evenly)
        terminal_offset = terminal_radius * 2 + terminal_spacing
        
        # Add three connection points (A, B, C phases)
        self.add_connection_point(-terminal_offset, 0, connection_z, "Terminal A")
        self.add_connection_point(0, 0, connection_z, "Terminal B")
        self.add_connection_point(terminal_offset, 0, connection_z, "Terminal C")
    
    def generate_geometry(self, bpy: Any) -> Any:
        """
        Generate substation 3D geometry.
        
        Args:
            bpy: Blender Python API module
        
        Returns:
            Blender object containing the substation
        """
        # Main cube body
        # Footprint: -0.9 to 0.9 in X and Y (size 1.8 x 1.8)
        # Height: 0.9
        body_height = 0.9
        bpy.ops.mesh.primitive_cube_add(
            size=1.0,
            location=(0, 0, body_height / 2)  # Center at half height
        )
        cube = bpy.context.active_object
        cube.name = f"{self.name}_Body"
        cube.scale = (self.body_size, self.body_size, body_height / 2)
        bpy.ops.object.transform_apply(scale=True)
        
        # Three cylinders on top for terminals
        terminal_base_z = body_height
        terminal_offset = self.terminal_radius * 2 + self.terminal_spacing
        
        positions_x = [
            -terminal_offset,  # Left
            0,                 # Center
            terminal_offset    # Right
        ]
        
        cylinders = []
        for i, x_pos in enumerate(positions_x):
            bpy.ops.mesh.primitive_cylinder_add(
                radius=self.terminal_radius,
                depth=self.terminal_height,
                location=(x_pos, 0, terminal_base_z + self.terminal_height / 2)
            )
            cyl = bpy.context.active_object
            cyl.name = f"{self.name}_Terminal_{i+1}"
            cylinders.append(cyl)
        
        # Join all objects into one
        bpy.ops.object.select_all(action='DESELECT')
        cube.select_set(True)
        for cyl in cylinders:
            cyl.select_set(True)
        bpy.context.view_layer.objects.active = cube
        bpy.ops.object.join()
        
        final_obj = bpy.context.active_object
        final_obj.name = self.name
        
        print(f"  ✓ Generated substation geometry ({len(final_obj.data.vertices)} vertices)")
        
        return final_obj

