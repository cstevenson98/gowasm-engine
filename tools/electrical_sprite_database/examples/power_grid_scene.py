"""
Example power grid scene with multiple substations connected by wires.

This script demonstrates how to create a scene with electrical components
and wire connections between them.
"""

import sys
import os

# Add electrical module to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'electrical'))

from scene import ComponentScene
from substation import Substation


def create_scene():
    """
    Create an example power grid scene.
    
    Returns:
        ComponentScene with components and connections
    """
    scene = ComponentScene(name="PowerGrid")
    
    # Create three substations
    sub1 = Substation(name="Substation_A")
    sub2 = Substation(name="Substation_B")
    sub3 = Substation(name="Substation_C")
    
    # Place on grid in a line
    # Grid squares are 2.0 units apart (from -1 to +1)
    scene.add_component(sub1, grid_x=0, grid_y=0, component_id="sub1")
    scene.add_component(sub2, grid_x=2, grid_y=0, component_id="sub2")
    scene.add_component(sub3, grid_x=4, grid_y=0, component_id="sub3")
    
    # Connect them with wires
    # sub1 right terminal -> sub2 left terminal
    scene.connect(
        comp1_id='sub1',
        point1_idx=2,  # Right terminal (index 2)
        comp2_id='sub2',
        point2_idx=0,  # Left terminal (index 0)
        sag=0.15,
        wire_name="Wire_1_to_2"
    )
    
    # sub2 right terminal -> sub3 left terminal
    scene.connect(
        comp1_id='sub2',
        point1_idx=2,
        comp2_id='sub3',
        point2_idx=0,
        sag=0.15,
        wire_name="Wire_2_to_3"
    )
    
    # Optional: Connect center terminals between sub1 and sub2 for redundancy
    scene.connect(
        comp1_id='sub1',
        point1_idx=1,  # Center terminal
        comp2_id='sub2',
        point2_idx=1,  # Center terminal
        sag=0.2,
        wire_name="Wire_1_to_2_center"
    )
    
    return scene


if __name__ == "__main__":
    print("This is a scene definition file.")
    print("Run with: blender --background --python generate_electrical.py -- \\")
    print("              --scene examples/power_grid_scene.py --output output/grid.obj")



