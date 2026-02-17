"""
Electrical component system for isometric game asset generation.

This module provides classes for creating electrical components (substations,
transformers, pylons) that can be placed on a grid, connected with wires,
and exported as 3D models for rendering as isometric sprites.
"""

from .base_component import ElectricalComponent
from .wire_generator import WireGenerator
from .scene import ComponentScene
from .substation import Substation
from .pylon import Pylon

__all__ = [
    'ElectricalComponent',
    'WireGenerator',
    'ComponentScene',
    'Substation',
    'Pylon',
]

