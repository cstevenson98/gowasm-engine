"""
Base class for electrical components in the isometric game.

Components can be placed on a grid, have connection points for wires,
and generate their own 3D geometry.
"""

from abc import ABC, abstractmethod
from typing import Any, Tuple, List, Optional


class ConnectionPoint:
    """Represents a connection point on a component."""
    
    def __init__(self, x: float, y: float, z: float, name: str = ""):
        """
        Create a connection point with local coordinates.
        
        Args:
            x: X offset from component center
            y: Y offset from component center
            z: Z offset from ground (absolute height)
            name: Optional name for this connection point
        """
        self.x = x
        self.y = y
        self.z = z
        self.name = name
    
    def __repr__(self):
        name_str = f" '{self.name}'" if self.name else ""
        return f"ConnectionPoint({self.x:.3f}, {self.y:.3f}, {self.z:.3f}{name_str})"


class ElectricalComponent(ABC):
    """
    Abstract base class for electrical components.
    
    Components occupy grid positions and can have connection points
    for attaching wires. Each subclass implements its own geometry.
    """
    
    # Grid square size (each square is 2.0 units, from -1 to +1)
    GRID_SIZE = 2.0
    
    def __init__(self, name: str, footprint_size: float = 1.0):
        """
        Initialize electrical component.
        
        Args:
            name: Component identifier
            footprint_size: Size within grid square (0.0 to 1.0)
        """
        self.name = name
        self.footprint_size = footprint_size
        self.grid_position: Optional[Tuple[int, int]] = None
        self.connection_points: List[ConnectionPoint] = []
        self._blender_object: Optional[Any] = None
    
    def add_connection_point(self, x: float, y: float, z: float, name: str = "") -> int:
        """
        Add a connection point to this component.
        
        Args:
            x: X offset from component center (local coordinates)
            y: Y offset from component center (local coordinates)
            z: Z coordinate (absolute height from ground)
            name: Optional name for this connection point
        
        Returns:
            Index of the added connection point
        """
        point = ConnectionPoint(x, y, z, name)
        self.connection_points.append(point)
        return len(self.connection_points) - 1
    
    def get_connection_point(self, index: int) -> ConnectionPoint:
        """Get connection point by index."""
        if 0 <= index < len(self.connection_points):
            return self.connection_points[index]
        raise IndexError(f"Connection point {index} not found (component has {len(self.connection_points)} points)")
    
    def get_world_connection_position(self, index: int) -> Tuple[float, float, float]:
        """
        Get world-space position of a connection point.
        
        Args:
            index: Connection point index
        
        Returns:
            (x, y, z) tuple in world coordinates
        """
        if self.grid_position is None:
            raise ValueError(f"Component {self.name} has not been placed on grid")
        
        point = self.get_connection_point(index)
        grid_x, grid_y = self.grid_position
        
        world_x = grid_x * self.GRID_SIZE + point.x
        world_y = grid_y * self.GRID_SIZE + point.y
        world_z = point.z
        
        return (world_x, world_y, world_z)
    
    def set_grid_position(self, grid_x: int, grid_y: int):
        """
        Place component at grid position.
        
        Args:
            grid_x: Grid X coordinate (integer)
            grid_y: Grid Y coordinate (integer)
        """
        self.grid_position = (grid_x, grid_y)
    
    def get_world_position(self) -> Tuple[float, float, float]:
        """
        Get world-space position of component center.
        
        Returns:
            (x, y, z) tuple in world coordinates (z is always 0 for ground level)
        """
        if self.grid_position is None:
            return (0, 0, 0)
        
        grid_x, grid_y = self.grid_position
        return (grid_x * self.GRID_SIZE, grid_y * self.GRID_SIZE, 0)
    
    @abstractmethod
    def generate_geometry(self, bpy: Any) -> Any:
        """
        Generate 3D geometry for this component.
        
        Args:
            bpy: Blender Python API module
        
        Returns:
            Blender object
        """
        pass
    
    def export_obj(self, bpy: Any, output_path: str):
        """
        Export component as OBJ file.
        
        Args:
            bpy: Blender Python API module
            output_path: Output file path
        """
        if self._blender_object is None:
            raise ValueError(f"Component {self.name} has not generated geometry yet")
        
        # Select only this object
        bpy.ops.object.select_all(action='DESELECT')
        self._blender_object.select_set(True)
        bpy.context.view_layer.objects.active = self._blender_object
        
        # Export as OBJ
        bpy.ops.wm.obj_export(
            filepath=output_path,
            export_selected_objects=True,
            forward_axis='Y',
            up_axis='Z'
        )
        
        print(f"✓ Exported {self.name} to: {output_path}")
    
    def __repr__(self):
        pos_str = f"@{self.grid_position}" if self.grid_position else "unplaced"
        return f"{self.__class__.__name__}('{self.name}', {pos_str}, {len(self.connection_points)} connections)"



