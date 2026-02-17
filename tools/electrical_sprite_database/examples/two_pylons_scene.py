"""
Simple two-pylon scene connected by a wire.
"""

import sys
import os

# Add electrical module to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'electrical'))

from scene import ComponentScene
from pylon import Pylon


def create_scene():
    """
    Create a simple scene with two pylons and one wire.
    
    Layout:
    [Pylon 1] ------------ [Pylon 2]
    (0, 0)                 (2, 0)
    
    Returns:
        ComponentScene with two pylons
    """
    scene = ComponentScene(name="TwoPylons")
    
    # Create two pylons
    pylon_1 = Pylon(name="Pylon_1")
    pylon_2 = Pylon(name="Pylon_2")
    
    # Place on grid
    scene.add_component(pylon_1, grid_x=0, grid_y=0, component_id="pylon_1")
    scene.add_component(pylon_2, grid_x=2, grid_y=0, component_id="pylon_2")
    
    # Connect them with a wire (center terminal to center terminal)
    scene.connect(
        comp1_id='pylon_1',
        point1_idx=1,  # Center terminal
        comp2_id='pylon_2',
        point2_idx=1,  # Center terminal
        sag=0.15,
        wire_name="Wire"
    )
    
    return scene


if __name__ == "__main__":
    print("Two pylons connected by a wire")



