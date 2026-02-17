"""
Scene manager for electrical component layouts.

Manages multiple components on a grid and connections between them.
"""

import os
from typing import Any, Dict, List, Tuple, Optional

try:
    from .base_component import ElectricalComponent
    from .wire_generator import WireGenerator
except ImportError:
    from base_component import ElectricalComponent
    from wire_generator import WireGenerator


class ComponentScene:
    """
    Manages a scene with multiple electrical components and wire connections.
    
    Components are placed on an integer grid and can be connected with wires
    between their connection points.
    """
    
    def __init__(self, name: str = "ElectricalScene"):
        """
        Initialize component scene.
        
        Args:
            name: Name for this scene
        """
        self.name = name
        self.components: Dict[str, ElectricalComponent] = {}
        self.connections: List[Dict] = []  # List of connection specs
        self.wire_generator = WireGenerator()
        self._blender_objects: List[Any] = []
    
    def add_component(
        self,
        component: ElectricalComponent,
        grid_x: int,
        grid_y: int,
        component_id: Optional[str] = None
    ) -> str:
        """
        Add a component to the scene at a grid position.
        
        Args:
            component: ElectricalComponent instance
            grid_x: Grid X coordinate
            grid_y: Grid Y coordinate
            component_id: Optional custom ID (uses component.name if not provided)
        
        Returns:
            Component ID used for connections
        """
        comp_id = component_id if component_id else component.name
        
        if comp_id in self.components:
            raise ValueError(f"Component with ID '{comp_id}' already exists in scene")
        
        component.set_grid_position(grid_x, grid_y)
        self.components[comp_id] = component
        
        print(f"✓ Added {component} to scene at grid ({grid_x}, {grid_y})")
        return comp_id
    
    def connect(
        self,
        comp1_id: str,
        point1_idx: int,
        comp2_id: str,
        point2_idx: int,
        sag: float = 0.1,
        wire_name: Optional[str] = None
    ):
        """
        Create a wire connection between two component connection points.
        
        Args:
            comp1_id: First component ID
            point1_idx: Connection point index on first component
            comp2_id: Second component ID
            point2_idx: Connection point index on second component
            sag: Wire sag amount (droop)
            wire_name: Optional name for the wire
        """
        # Validate components exist
        if comp1_id not in self.components:
            raise ValueError(f"Component '{comp1_id}' not found in scene")
        if comp2_id not in self.components:
            raise ValueError(f"Component '{comp2_id}' not found in scene")
        
        comp1 = self.components[comp1_id]
        comp2 = self.components[comp2_id]
        
        # Validate connection points exist
        if point1_idx >= len(comp1.connection_points):
            raise ValueError(f"Connection point {point1_idx} not found on {comp1_id}")
        if point2_idx >= len(comp2.connection_points):
            raise ValueError(f"Connection point {point2_idx} not found on {comp2_id}")
        
        # Store connection spec
        connection = {
            'comp1_id': comp1_id,
            'point1_idx': point1_idx,
            'comp2_id': comp2_id,
            'point2_idx': point2_idx,
            'sag': sag,
            'name': wire_name or f"Wire_{comp1_id}_to_{comp2_id}"
        }
        self.connections.append(connection)
        
        print(f"✓ Queued connection: {comp1_id}[{point1_idx}] -> {comp2_id}[{point2_idx}]")
    
    def generate_full_scene(self, bpy: Any) -> List[Any]:
        """
        Generate all components and wires in Blender.
        
        Args:
            bpy: Blender Python API module
        
        Returns:
            List of all Blender objects created
        """
        self._blender_objects = []
        
        print(f"\n{'='*60}")
        print(f"Generating Scene: {self.name}")
        print(f"{'='*60}")
        print(f"Components: {len(self.components)}")
        print(f"Connections: {len(self.connections)}")
        print()
        
        # Generate all components
        for comp_id, component in self.components.items():
            print(f"Generating {comp_id}...")
            obj = component.generate_geometry(bpy)
            
            # Move to world position
            world_pos = component.get_world_position()
            obj.location = world_pos
            
            component._blender_object = obj
            self._blender_objects.append(obj)
            print(f"  ✓ Positioned at {world_pos}")
        
        print()
        
        # Generate all wires
        for conn in self.connections:
            comp1 = self.components[conn['comp1_id']]
            comp2 = self.components[conn['comp2_id']]
            
            start_pos = comp1.get_world_connection_position(conn['point1_idx'])
            end_pos = comp2.get_world_connection_position(conn['point2_idx'])
            
            print(f"Generating {conn['name']}...")
            wire_obj = self.wire_generator.generate_wire(
                bpy,
                start_pos,
                end_pos,
                sag=conn['sag'],
                name=conn['name']
            )
            self._blender_objects.append(wire_obj)
        
        print(f"\n{'='*60}")
        print(f"✓✓✓ Scene generation complete!")
        print(f"Total objects: {len(self._blender_objects)}")
        print(f"{'='*60}\n")
        
        return self._blender_objects
    
    def generate_components_only(self, bpy: Any) -> List[Any]:
        """
        Generate only component geometry, no wires.
        
        Args:
            bpy: Blender Python API module
        
        Returns:
            List of component Blender objects
        """
        self._blender_objects = []
        
        print(f"\n{'='*60}")
        print(f"Generating Components Only: {self.name}")
        print(f"{'='*60}")
        print(f"Components: {len(self.components)}")
        print()
        
        # Generate all components
        for comp_id, component in self.components.items():
            print(f"Generating {comp_id}...")
            obj = component.generate_geometry(bpy)
            
            # Move to world position
            world_pos = component.get_world_position()
            obj.location = world_pos
            
            component._blender_object = obj
            self._blender_objects.append(obj)
            print(f"  ✓ Positioned at {world_pos}")
        
        print(f"\n{'='*60}")
        print(f"✓✓✓ Component generation complete!")
        print(f"Total objects: {len(self._blender_objects)}")
        print(f"{'='*60}\n")
        
        return self._blender_objects
    
    def generate_single_wire(self, bpy: Any, conn_index: int) -> Any:
        """
        Generate only one specific wire connection.
        
        Args:
            bpy: Blender Python API module
            conn_index: Index of connection to generate
        
        Returns:
            Wire Blender object
        """
        if conn_index < 0 or conn_index >= len(self.connections):
            raise IndexError(f"Connection index {conn_index} out of range (have {len(self.connections)} connections)")
        
        conn = self.connections[conn_index]
        
        # Ensure components have been generated
        comp1 = self.components[conn['comp1_id']]
        comp2 = self.components[conn['comp2_id']]
        
        if comp1._blender_object is None or comp2._blender_object is None:
            raise ValueError("Components must be generated before wires. Call generate_components_only() first.")
        
        start_pos = comp1.get_world_connection_position(conn['point1_idx'])
        end_pos = comp2.get_world_connection_position(conn['point2_idx'])
        
        print(f"\nGenerating single wire: {conn['name']}...")
        print(f"  From: {conn['comp1_id']}[{conn['point1_idx']}] at {start_pos}")
        print(f"  To: {conn['comp2_id']}[{conn['point2_idx']}] at {end_pos}")
        
        wire_obj = self.wire_generator.generate_wire(
            bpy,
            start_pos,
            end_pos,
            sag=conn['sag'],
            name=conn['name']
        )
        
        print(f"✓ Wire generated")
        
        return wire_obj
    
    def generate_wires_only(self, bpy: Any) -> List[Any]:
        """
        Generate only wires (assumes components already exist invisibly).
        
        Args:
            bpy: Blender Python API module
        
        Returns:
            List of wire Blender objects
        """
        wire_objects = []
        
        print(f"\n{'='*60}")
        print(f"Generating Wires Only: {self.name}")
        print(f"{'='*60}")
        print(f"Connections: {len(self.connections)}")
        print()
        
        # Generate all wires
        for conn in self.connections:
            comp1 = self.components[conn['comp1_id']]
            comp2 = self.components[conn['comp2_id']]
            
            start_pos = comp1.get_world_connection_position(conn['point1_idx'])
            end_pos = comp2.get_world_connection_position(conn['point2_idx'])
            
            print(f"Generating {conn['name']}...")
            wire_obj = self.wire_generator.generate_wire(
                bpy,
                start_pos,
                end_pos,
                sag=conn['sag'],
                name=conn['name']
            )
            wire_objects.append(wire_obj)
            self._blender_objects.append(wire_obj)  # Add to scene objects for export
        
        print(f"\n{'='*60}")
        print(f"✓✓✓ Wire generation complete!")
        print(f"Total wires: {len(wire_objects)}")
        print(f"{'='*60}\n")
        
        return wire_objects
    
    def export_scene(self, bpy: Any, output_path: str):
        """
        Export entire scene as a single OBJ file.
        
        Args:
            bpy: Blender Python API module
            output_path: Output file path
        """
        if not self._blender_objects:
            raise ValueError("No objects generated. Call generate_full_scene() first.")
        
        # Select all scene objects
        bpy.ops.object.select_all(action='DESELECT')
        for obj in self._blender_objects:
            obj.select_set(True)
        
        if self._blender_objects:
            bpy.context.view_layer.objects.active = self._blender_objects[0]
        
        # Export as OBJ
        bpy.ops.wm.obj_export(
            filepath=output_path,
            export_selected_objects=True,
            forward_axis='Y',
            up_axis='Z'
        )
        
        print(f"✓ Exported scene to: {output_path}")
    
    def export_separate(self, bpy: Any, output_dir: str):
        """
        Export each component separately.
        
        Args:
            bpy: Blender Python API module
            output_dir: Output directory for component files
        """
        if not os.path.exists(output_dir):
            os.makedirs(output_dir)
        
        for comp_id, component in self.components.items():
            filename = f"{comp_id}.obj"
            output_path = os.path.join(output_dir, filename)
            component.export_obj(bpy, output_path)
    
    def clear_scene(self, bpy: Any):
        """
        Remove all objects from the Blender scene.
        
        Args:
            bpy: Blender Python API module
        """
        bpy.ops.object.select_all(action='SELECT')
        bpy.ops.object.delete()
        self._blender_objects = []
        print("✓ Cleared scene")
    
    def __repr__(self):
        return (f"ComponentScene('{self.name}', "
                f"{len(self.components)} components, "
                f"{len(self.connections)} connections)")

