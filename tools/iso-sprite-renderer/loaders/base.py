"""
Base interface for mesh loaders.
Provides abstract interface for loading different 3D file formats into Blender.
"""
from abc import ABC, abstractmethod
from typing import Any


class MeshLoader(ABC):
    """Abstract base class for mesh loaders."""
    
    @abstractmethod
    def supports(self, filepath: str) -> bool:
        """
        Check if this loader can handle the given file.
        
        Args:
            filepath: Path to the mesh file
            
        Returns:
            True if this loader supports the file format
        """
        pass
    
    @abstractmethod
    def load_into_blender(self, filepath: str, bpy: Any) -> Any:
        """
        Load mesh into Blender scene.
        
        Args:
            filepath: Path to the mesh file
            bpy: Blender Python API module
            
        Returns:
            The loaded Blender object
        """
        pass
    
    @abstractmethod
    def get_file_extensions(self) -> list[str]:
        """
        Get list of supported file extensions.
        
        Returns:
            List of file extensions (e.g., ['.obj', '.OBJ'])
        """
        pass
    
    def normalize_object(self, obj: Any, bpy: Any) -> None:
        """
        Normalize object size and center it at origin.
        Scales based on X-Y plane extent only (ignores height).
        This ensures models fill the isometric view properly.
        
        Args:
            obj: Blender object to normalize
            bpy: Blender Python API module
        """
        # Select only this object
        bpy.ops.object.select_all(action='DESELECT')
        obj.select_set(True)
        bpy.context.view_layer.objects.active = obj
        
        # Reset location and rotation
        obj.location = (0, 0, 0)
        obj.rotation_euler = (0, 0, 0)
        
        # Calculate bounding box and scale based on X-Y plane only
        # This ensures the model fills the isometric view horizontally
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

