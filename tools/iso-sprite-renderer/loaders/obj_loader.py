"""
OBJ mesh loader implementation.
"""
import os
from typing import Any
from loaders.base import MeshLoader


class OBJLoader(MeshLoader):
    """Loader for Wavefront OBJ files."""
    
    def supports(self, filepath: str) -> bool:
        """Check if file is an OBJ file."""
        ext = os.path.splitext(filepath)[1].lower()
        return ext in self.get_file_extensions()
    
    def get_file_extensions(self) -> list[str]:
        """Return supported OBJ extensions."""
        return ['.obj', '.OBJ']
    
    def load_into_blender(self, filepath: str, bpy: Any) -> Any:
        """
        Load OBJ file into Blender.
        
        Args:
            filepath: Path to OBJ file
            bpy: Blender Python API module
            
        Returns:
            The loaded Blender object
        """
        # Get objects before import
        objects_before = set(bpy.context.scene.objects)
        
        # Import OBJ file
        bpy.ops.import_scene.obj(filepath=filepath)
        
        # Find newly imported objects
        objects_after = set(bpy.context.scene.objects)
        new_objects = objects_after - objects_before
        
        if not new_objects:
            raise RuntimeError(f"No objects were imported from {filepath}")
        
        # If multiple objects were imported, join them into one
        if len(new_objects) > 1:
            # Select all imported objects
            bpy.ops.object.select_all(action='DESELECT')
            for obj in new_objects:
                obj.select_set(True)
            
            # Set one as active
            obj = list(new_objects)[0]
            bpy.context.view_layer.objects.active = obj
            
            # Join all selected objects
            bpy.ops.object.join()
        else:
            obj = list(new_objects)[0]
        
        # Normalize the object (center and scale)
        self.normalize_object(obj, bpy)
        
        return obj

