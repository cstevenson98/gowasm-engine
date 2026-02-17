"""
Wire generator for creating 3D cables between electrical components.

Generates realistic-looking wires with catenary sag between connection points.
"""

import math
from typing import Any, Tuple


class WireGenerator:
    """
    Generates 3D wire/cable geometry with realistic catenary sag.
    """
    
    def __init__(self):
        """Initialize wire generator."""
        pass
    
    def generate_wire(
        self,
        bpy: Any,
        start_pos: Tuple[float, float, float],
        end_pos: Tuple[float, float, float],
        sag: float = 0.1,
        segments: int = 16,
        radius: float = 0.02,
        name: str = "Wire"
    ) -> Any:
        """
        Generate a wire between two points with catenary sag.
        
        Args:
            bpy: Blender Python API module
            start_pos: (x, y, z) start position
            end_pos: (x, y, z) end position
            sag: How much the wire droops (0.0 = straight, higher = more droop)
            segments: Number of segments for curve smoothness
            radius: Wire thickness
            name: Name for the wire object
        
        Returns:
            Blender mesh object representing the wire
        """
        # Create curve data
        curve_data = bpy.data.curves.new(name=f"{name}_curve", type='CURVE')
        curve_data.dimensions = '3D'
        curve_data.resolution_u = 2
        
        # Create a NURBS or POLY spline
        spline = curve_data.splines.new('NURBS')
        
        # Calculate catenary points
        points = self._calculate_catenary_points(start_pos, end_pos, sag, segments)
        
        # Add points to spline (NURBS starts with 1 point, we need segments+1 total)
        spline.points.add(len(points) - 1)
        
        for i, point in enumerate(points):
            x, y, z = point
            spline.points[i].co = (x, y, z, 1.0)  # (x, y, z, w) for NURBS
        
        spline.use_endpoint_u = True
        
        # Create curve object
        curve_obj = bpy.data.objects.new(f"{name}_curve_obj", curve_data)
        bpy.context.collection.objects.link(curve_obj)
        
        # Add bevel (thickness) to the curve
        curve_data.bevel_depth = radius
        curve_data.bevel_resolution = 4  # Circular cross-section resolution
        
        # Convert curve to mesh for better compatibility
        bpy.context.view_layer.objects.active = curve_obj
        bpy.ops.object.select_all(action='DESELECT')
        curve_obj.select_set(True)
        bpy.ops.object.convert(target='MESH')
        
        mesh_obj = bpy.context.active_object
        mesh_obj.name = name
        
        print(f"✓ Generated wire '{name}' from {start_pos} to {end_pos} (sag={sag}, {len(points)} points)")
        
        return mesh_obj
    
    def _calculate_catenary_points(
        self,
        start: Tuple[float, float, float],
        end: Tuple[float, float, float],
        sag: float,
        segments: int
    ) -> list:
        """
        Calculate points along a catenary curve (hanging cable shape).
        
        Args:
            start: Start position (x, y, z)
            end: End position (x, y, z)
            sag: Sag amount (vertical droop at midpoint)
            segments: Number of segments
        
        Returns:
            List of (x, y, z) points
        """
        points = []
        
        # Calculate horizontal distance
        dx = end[0] - start[0]
        dy = end[1] - start[1]
        dz = end[2] - start[2]
        horizontal_dist = math.sqrt(dx * dx + dy * dy)
        
        # Avoid division by zero for vertical wires
        if horizontal_dist < 0.001:
            # For nearly vertical wires, just interpolate linearly
            for i in range(segments + 1):
                t = i / segments
                x = start[0] + dx * t
                y = start[1] + dy * t
                z = start[2] + dz * t
                points.append((x, y, z))
            return points
        
        # Generate points along the curve
        for i in range(segments + 1):
            t = i / segments  # Parameter from 0 to 1
            
            # Linear interpolation in X-Y plane
            x = start[0] + dx * t
            y = start[1] + dy * t
            
            # Catenary formula for Z (vertical sag)
            # Using parabolic approximation for simplicity: z = -4*sag*t*(1-t)
            # This creates a parabolic sag with maximum at t=0.5
            sag_z = -4.0 * sag * t * (1.0 - t)
            z = start[2] + dz * t + sag_z
            
            points.append((x, y, z))
        
        return points
    
    def generate_multiple_wires(
        self,
        bpy: Any,
        connections: list,
        sag: float = 0.1,
        radius: float = 0.02
    ) -> list:
        """
        Generate multiple wires from a list of connections.
        
        Args:
            bpy: Blender Python API module
            connections: List of (start_pos, end_pos, name) tuples
            sag: Sag amount for all wires
            radius: Wire radius for all wires
        
        Returns:
            List of Blender wire objects
        """
        wires = []
        for i, (start, end, name) in enumerate(connections):
            wire_name = name if name else f"Wire_{i}"
            wire = self.generate_wire(bpy, start, end, sag, radius=radius, name=wire_name)
            wires.append(wire)
        
        return wires



