"""
Pylon electrical component.

A tall pylon/tower with a horizontal crossbar and three terminals on top.
"""

import math
from typing import Any

try:
    from .base_component import ElectricalComponent
except ImportError:
    from base_component import ElectricalComponent


class Pylon(ElectricalComponent):
    """
    Electrical pylon/tower component.
    
    Features:
    - Tall thin cylinder pole (central support)
    - Horizontal crossbar at top
    - Three cylindrical terminals on crossbar
    - Three connection points at terminal centers
    """
    
    def __init__(
        self,
        name: str = "Pylon",
        pole_radius: float = 0.05,
        pole_height: float = 2.0,
        crossbar_width: float = 1.2,
        crossbar_depth: float = 0.08,
        crossbar_height: float = 0.08,
        terminal_radius: float = 0.04,
        terminal_height: float = 0.08,
        terminal_spacing: float = 0.05
    ):
        """
        Initialize pylon.
        
        Args:
            name: Component name
            pole_radius: Radius of central pole
            pole_height: Height of pole
            crossbar_width: Width of horizontal bar (X direction)
            crossbar_depth: Depth of horizontal bar (Y direction)
            crossbar_height: Height/thickness of horizontal bar
            terminal_radius: Radius of terminal cylinders
            terminal_height: Height of terminal cylinders
            terminal_spacing: Space between terminals
        """
        super().__init__(name, footprint_size=pole_radius * 2)
        
        self.pole_radius = pole_radius
        self.pole_height = pole_height
        self.crossbar_width = crossbar_width
        self.crossbar_depth = crossbar_depth
        self.crossbar_height = crossbar_height
        self.terminal_radius = terminal_radius
        self.terminal_height = terminal_height
        self.terminal_spacing = terminal_spacing
        
        # Define connection points at the BOTTOM of terminal cylinders (on top of crossbar)
        # This is where wires actually attach in real power lines
        crossbar_top_z = pole_height/2
        connection_z = crossbar_top_z  # Right at the crossbar surface
        
        # Calculate terminal X positions (centered, spaced evenly)
        # Three terminals spread across the crossbar
        terminal_offset = (crossbar_width - terminal_radius * 6) / 4  # Space them out nicely
        
        # Add three connection points (A, B, C phases)
        self.add_connection_point(-terminal_offset - terminal_radius, 0, connection_z, "Terminal A")
        self.add_connection_point(0, 0, connection_z, "Terminal B")
        self.add_connection_point(terminal_offset + terminal_radius, 0, connection_z, "Terminal C")
    
    def generate_geometry(self, bpy: Any) -> Any:
        """
        Generate pylon 3D geometry.
        
        Args:
            bpy: Blender Python API module
        
        Returns:
            Blender object containing the pylon
        """
        objects_to_join = []
        
        # 1. Central pole (thin cylinder)
        bpy.ops.mesh.primitive_cylinder_add(
            radius=self.pole_radius,
            depth=self.pole_height,
            location=(0, 0, 0)
        )
        pole = bpy.context.active_object
        pole.name = f"{self.name}_Pole"
        objects_to_join.append(pole)
        
        # 2. Horizontal crossbar at top
        crossbar_z = self.pole_height/2 + self.crossbar_height / 2
        bpy.ops.mesh.primitive_cube_add(
            size=1.0,
            location=(0, 0, crossbar_z)
        )
        crossbar = bpy.context.active_object
        crossbar.name = f"{self.name}_Crossbar"
        crossbar.scale = (
            self.crossbar_width / 2,
            self.crossbar_depth / 2,
            self.crossbar_height / 2
        )
        bpy.ops.object.transform_apply(scale=True)
        objects_to_join.append(crossbar)
        
        # 3. Three terminals on top of crossbar
        crossbar_top_z = self.pole_height/2 + self.crossbar_height
        terminal_base_z = crossbar_top_z + self.terminal_height / 2
        
        # Calculate terminal positions
        terminal_offset = (self.crossbar_width - self.terminal_radius * 6) / 4
        
        positions_x = [
            -terminal_offset - self.terminal_radius,  # Left
            0,                                         # Center
            terminal_offset + self.terminal_radius     # Right
        ]
        
        for i, x_pos in enumerate(positions_x):
            bpy.ops.mesh.primitive_cylinder_add(
                radius=self.terminal_radius,
                depth=self.terminal_height,
                location=(x_pos, 0, terminal_base_z)
            )
            terminal = bpy.context.active_object
            terminal.name = f"{self.name}_Terminal_{i+1}"
            objects_to_join.append(terminal)
        
        # Join all objects into one
        bpy.ops.object.select_all(action='DESELECT')
        for obj in objects_to_join:
            obj.select_set(True)
        bpy.context.view_layer.objects.active = objects_to_join[0]
        bpy.ops.object.join()
        
        final_obj = bpy.context.active_object
        final_obj.name = self.name
        
        print(f"  ✓ Generated pylon geometry ({len(final_obj.data.vertices)} vertices)")
        
        return final_obj

