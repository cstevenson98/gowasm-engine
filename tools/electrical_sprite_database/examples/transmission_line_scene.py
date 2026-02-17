"""
Example transmission line scene with substations connected by pylons.

This demonstrates a more realistic power grid with:
- Two substations at ends
- Three pylons in between
- Wires connecting all components in sequence
"""

import sys
import os

# Add electrical module to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'electrical'))

from scene import ComponentScene
from substation import Substation
from pylon import Pylon


def create_scene():
    """
    Create a transmission line scene.
    
    Layout:
    [Sub A] -- [Pylon 1] -- [Pylon 2] -- [Pylon 3] -- [Sub B]
    (0, 0)     (2, 0)       (4, 0)       (6, 0)       (8, 0)
    
    Returns:
        ComponentScene with transmission line
    """
    scene = ComponentScene(name="TransmissionLine")
    
    # Create components
    sub_a = Substation(name="Substation_A")
    pylon_1 = Pylon(name="Pylon_1")
    pylon_2 = Pylon(name="Pylon_2")
    pylon_3 = Pylon(name="Pylon_3")
    sub_b = Substation(name="Substation_B")
    
    # Place on grid in a line (2 units apart)
    scene.add_component(sub_a, grid_x=0, grid_y=0, component_id="sub_a")
    scene.add_component(pylon_1, grid_x=2, grid_y=0, component_id="pylon_1")
    scene.add_component(pylon_2, grid_x=4, grid_y=0, component_id="pylon_2")
    scene.add_component(pylon_3, grid_x=6, grid_y=0, component_id="pylon_3")
    scene.add_component(sub_b, grid_x=8, grid_y=0, component_id="sub_b")
    
    # Connect them sequentially with wires
    # Use right terminal from each component to left terminal of next
    
    # Sub A (right terminal) -> Pylon 1 (left terminal)
    scene.connect(
        comp1_id='sub_a',
        point1_idx=2,  # Right terminal
        comp2_id='pylon_1',
        point2_idx=0,  # Left terminal
        sag=0.15,
        wire_name="Wire_SubA_to_Pylon1"
    )
    
    # Pylon 1 (right) -> Pylon 2 (left)
    scene.connect(
        comp1_id='pylon_1',
        point1_idx=2,
        comp2_id='pylon_2',
        point2_idx=0,
        sag=0.15,
        wire_name="Wire_Pylon1_to_Pylon2"
    )
    
    # Pylon 2 (right) -> Pylon 3 (left)
    scene.connect(
        comp1_id='pylon_2',
        point1_idx=2,
        comp2_id='pylon_3',
        point2_idx=0,
        sag=0.15,
        wire_name="Wire_Pylon2_to_Pylon3"
    )
    
    # Pylon 3 (right) -> Sub B (left)
    scene.connect(
        comp1_id='pylon_3',
        point1_idx=2,
        comp2_id='sub_b',
        point2_idx=0,
        sag=0.15,
        wire_name="Wire_Pylon3_to_SubB"
    )
    
    return scene


if __name__ == "__main__":
    print("This is a scene definition file.")
    print("Run with: blender --background --python generate_electrical.py -- \\")
    print("              --scene examples/transmission_line_scene.py --output output/transmission_line.obj")



