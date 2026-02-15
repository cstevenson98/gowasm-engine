"""
Loader registry and utilities.
"""
from loaders.base import MeshLoader
from loaders.obj_loader import OBJLoader

# Registry of all available loaders
LOADERS: list[MeshLoader] = [
    OBJLoader(),
]


def get_loader_for_file(filepath: str) -> MeshLoader | None:
    """
    Find appropriate loader for the given file.
    
    Args:
        filepath: Path to the mesh file
        
    Returns:
        Appropriate MeshLoader instance, or None if no loader supports the file
    """
    for loader in LOADERS:
        if loader.supports(filepath):
            return loader
    return None


def get_all_supported_extensions() -> list[str]:
    """
    Get all supported file extensions across all loaders.
    
    Returns:
        List of supported file extensions
    """
    extensions = []
    for loader in LOADERS:
        extensions.extend(loader.get_file_extensions())
    return list(set(extensions))  # Remove duplicates


__all__ = ['MeshLoader', 'OBJLoader', 'get_loader_for_file', 'get_all_supported_extensions']

