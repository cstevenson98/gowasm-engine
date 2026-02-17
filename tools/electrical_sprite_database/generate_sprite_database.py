#!/usr/bin/env python3
"""
Generate complete sprite database for electrical components.

Creates layered sprites for all component configurations:
- Base components (no wires)
- Individual wire connections (for compositing)
"""

import os
import sys
import yaml
import json
import subprocess
import tempfile
import argparse
from typing import List, Dict, Any, Tuple
from pathlib import Path

# Add parent tools directory to path for iso-sprite-renderer library
TOOLS_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), '..'))
sys.path.insert(0, TOOLS_DIR)

# Add electrical module to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'electrical'))

from electrical.grid_utils import enumerate_grid_offsets, get_phase_index, format_offset_name
from electrical.pylon import Pylon
from electrical.substation import Substation
from electrical.scene import ComponentScene


class SpriteDatabase:
    """Manager for generating sprite database."""
    
    def __init__(self, config_path: str):
        """
        Initialize sprite database generator.
        
        Args:
            config_path: Path to YAML configuration file
        """
        with open(config_path, 'r') as f:
            self.config = yaml.safe_load(f)
        
        self.output_dir = self.config['output_dir']
        self.metadata = {
            'version': '1.0',
            'sprite_size': self.config['sprite']['size'],
            'camera': {
                'ortho_scale': self.config['sprite']['ortho_scale'],
                'elevation': self.config['camera']['elevation'],
                'rotation': self.config['camera']['rotation']
            },
            'components': {}
        }
        
        # Statistics
        self.stats = {
            'base_sprites': 0,
            'connection_sprites': 0,
            'skipped': 0,
            'failed': 0
        }
    
    def generate_all(self):
        """Generate complete sprite database."""
        print("="*70)
        print("SPRITE DATABASE GENERATOR")
        print("="*70)
        print(f"Output: {self.output_dir}")
        print(f"Sprite size: {self.config['sprite']['size']}x{self.config['sprite']['size']}")
        print(f"Connection radius: {self.config['connection_radius']}")
        print()
        
        # Create output directory structure
        self._create_directory_structure()
        
        # Generate base components
        if self.config['rendering']['generate_base']:
            print("\n" + "="*70)
            print("GENERATING BASE COMPONENTS")
            print("="*70)
            for comp_type in self.config['components']:
                self._generate_base_component(comp_type)
        
        # Generate connection sprites
        if self.config['rendering']['generate_connections']:
            print("\n" + "="*70)
            print("GENERATING CONNECTION SPRITES")
            print("="*70)
            self._generate_all_connections()
        
        # Save metadata
        self._save_metadata()
        
        # Print summary
        print("\n" + "="*70)
        print("GENERATION COMPLETE")
        print("="*70)
        print(f"Base sprites: {self.stats['base_sprites']}")
        print(f"Connection sprites: {self.stats['connection_sprites']}")
        print(f"Skipped (existing): {self.stats['skipped']}")
        print(f"Failed: {self.stats['failed']}")
        print(f"\nOutput: {self.output_dir}")
        print("="*70)
    
    def _create_directory_structure(self):
        """Create output directory structure."""
        os.makedirs(self.output_dir, exist_ok=True)
        
        for comp_type in self.config['components']:
            comp_dir = os.path.join(self.output_dir, comp_type)
            os.makedirs(comp_dir, exist_ok=True)
            os.makedirs(os.path.join(comp_dir, 'connections'), exist_ok=True)
    
    def _generate_base_component(self, comp_type: str):
        """
        Generate base component sprite (no wires).
        
        Args:
            comp_type: Component type ('pylon' or 'substation')
        """
        output_path = os.path.join(self.output_dir, comp_type, 'base.png')
        
        # Skip if exists and not overwriting
        if os.path.exists(output_path) and not self.config['rendering']['overwrite_existing']:
            print(f"✓ {comp_type}/base.png (exists, skipping)")
            self.stats['skipped'] += 1
            return
        
        print(f"\nGenerating {comp_type}/base.png...")
        
        # Create temporary scene script
        with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
            scene_script = f.name
            f.write(self._create_base_scene_script(comp_type))
        
        try:
            # Generate scene OBJ
            temp_obj = tempfile.mktemp(suffix='.obj')
            result = self._run_blender_scene_generation(scene_script, temp_obj)
            
            if result.returncode != 0:
                print(f"  ❌ Failed to generate scene: {result.stderr}")
                self.stats['failed'] += 1
                return
            
            # Render sprite
            temp_output_dir = tempfile.mkdtemp()
            result = self._render_sprite(temp_obj, temp_output_dir)
            
            if result.returncode != 0:
                print(f"  ❌ Failed to render sprite: {result.stderr}")
                self.stats['failed'] += 1
                return
            
            # Move output to final location
            # The output filename is based on the input OBJ filename
            obj_basename = os.path.splitext(os.path.basename(temp_obj))[0]
            rendered_sprite = os.path.join(temp_output_dir, f'{obj_basename}_angle_0.png')
            if os.path.exists(rendered_sprite):
                os.rename(rendered_sprite, output_path)
                print(f"  ✓ Generated: {output_path}")
                self.stats['base_sprites'] += 1
                
                # Add to metadata
                if comp_type not in self.metadata['components']:
                    self.metadata['components'][comp_type] = {
                        'base': f"{comp_type}/base.png",
                        'connection_points': self.config['phases'],
                        'connections': []
                    }
            else:
                print(f"  ❌ Sprite file not found: {rendered_sprite}")
                self.stats['failed'] += 1
            
            # Cleanup
            os.unlink(temp_obj)
            
        finally:
            os.unlink(scene_script)
    
    def _generate_all_connections(self):
        """Generate all connection sprites."""
        offsets = enumerate_grid_offsets(self.config['connection_radius'])
        
        print(f"\nFound {len(offsets)} grid offsets within radius {self.config['connection_radius']}")
        
        total_connections = len(self.config['connection_types']) * len(offsets) * len(self.config['phases'])
        print(f"Total connection sprites to generate: {total_connections}\n")
        
        for source_type, target_type in self.config['connection_types']:
            for offset_x, offset_y in offsets:
                for phase in self.config['phases']:
                    self._generate_connection_sprite(
                        source_type, target_type,
                        offset_x, offset_y,
                        phase
                    )
    
    def _generate_connection_sprite(
        self,
        source_type: str,
        target_type: str,
        offset_x: int,
        offset_y: int,
        phase: str
    ):
        """
        Generate single wire connection sprite.
        
        Args:
            source_type: Source component type
            target_type: Target component type
            offset_x: X offset to target
            offset_y: Y offset to target
            phase: Phase letter ('A', 'B', or 'C')
        """
        # Build filename
        offset_name = format_offset_name(offset_x, offset_y)
        filename = f"to_{target_type}_{offset_name}_{phase}_to_{phase}.png"
        output_path = os.path.join(
            self.output_dir,
            source_type,
            'connections',
            filename
        )
        
        # Skip if exists and not overwriting
        if os.path.exists(output_path) and not self.config['rendering']['overwrite_existing']:
            self.stats['skipped'] += 1
            return
        
        print(f"Generating {source_type} → {target_type} {offset_name} phase {phase}...", end=' ')
        
        # Create temporary scene script
        with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
            scene_script = f.name
            f.write(self._create_connection_scene_script(
                source_type, target_type,
                offset_x, offset_y,
                phase
            ))
        
        try:
            # Generate scene OBJ (wire only)
            temp_obj = tempfile.mktemp(suffix='.obj')
            result = self._run_blender_scene_generation(scene_script, temp_obj, wire_only=True)
            
            if result.returncode != 0:
                print(f"❌ Scene generation failed")
                self.stats['failed'] += 1
                return
            
            # Render sprite
            temp_output_dir = tempfile.mkdtemp()
            result = self._render_sprite(temp_obj, temp_output_dir)
            
            if result.returncode != 0:
                print(f"❌ Render failed")
                self.stats['failed'] += 1
                return
            
            # Move output to final location
            obj_basename = os.path.splitext(os.path.basename(temp_obj))[0]
            rendered_sprite = os.path.join(temp_output_dir, f'{obj_basename}_angle_0.png')
            if os.path.exists(rendered_sprite):
                os.rename(rendered_sprite, output_path)
                print(f"✓")
                self.stats['connection_sprites'] += 1
                
                # Add to metadata
                if source_type in self.metadata['components']:
                    self.metadata['components'][source_type]['connections'].append({
                        'target': target_type,
                        'offset': [offset_x, offset_y],
                        'phase': phase,
                        'file': f"{source_type}/connections/{filename}"
                    })
            else:
                print(f"❌ Output not found")
                self.stats['failed'] += 1
            
            # Cleanup
            if os.path.exists(temp_obj):
                os.unlink(temp_obj)
            
        finally:
            os.unlink(scene_script)
    
    def _create_base_scene_script(self, comp_type: str) -> str:
        """
        Create Python script for base component scene.
        
        Args:
            comp_type: Component type
        
        Returns:
            Python script content
        """
        return f"""
import sys
sys.path.insert(0, 'electrical')

from {comp_type} import {comp_type.capitalize()}
from scene import ComponentScene

def create_scene():
    scene = ComponentScene("Base_{comp_type.capitalize()}")
    component = {comp_type.capitalize()}(name="{comp_type}_base")
    scene.add_component(component, 0, 0, "{comp_type}_base")
    return scene
"""
    
    def _create_connection_scene_script(
        self,
        source_type: str,
        target_type: str,
        offset_x: int,
        offset_y: int,
        phase: str
    ) -> str:
        """
        Create Python script for connection scene.
        
        Args:
            source_type: Source component type
            target_type: Target component type
            offset_x: X offset
            offset_y: Y offset
            phase: Phase letter
        
        Returns:
            Python script content
        """
        phase_idx = get_phase_index(phase)
        
        return f"""
import sys
sys.path.insert(0, 'electrical')

from {source_type} import {source_type.capitalize()}
from {target_type} import {target_type.capitalize()}
from scene import ComponentScene

def create_scene():
    scene = ComponentScene("Connection_{source_type}_to_{target_type}")
    
    source = {source_type.capitalize()}(name="source")
    target = {target_type.capitalize()}(name="target")
    
    scene.add_component(source, 0, 0, "source")
    scene.add_component(target, {offset_x}, {offset_y}, "target")
    
    scene.connect("source", {phase_idx}, "target", {phase_idx}, sag={self.config['wire']['sag']})
    
    return scene
"""
    
    def _run_blender_scene_generation(
        self,
        scene_script: str,
        output_obj: str,
        wire_only: bool = False
    ) -> subprocess.CompletedProcess:
        """
        Run Blender to generate scene geometry.
        
        Args:
            scene_script: Path to scene script
            output_obj: Output OBJ path
            wire_only: If True, generate only wires
        
        Returns:
            subprocess.CompletedProcess result
        """
        cmd = [
            'blender',
            '--background',
            '--python', 'electrical/generate_electrical.py',
            '--',
            '--scene', scene_script,
            '--output', output_obj
        ]
        
        if wire_only:
            cmd.append('--wire-only')
        
        return subprocess.run(cmd, capture_output=True, text=True, cwd=os.path.dirname(__file__))
    
    def _render_sprite(self, obj_path: str, output_dir: str) -> subprocess.CompletedProcess:
        """
        Render sprite from OBJ file using iso-sprite-renderer library.
        
        Args:
            obj_path: Path to OBJ file
            output_dir: Output directory
        
        Returns:
            subprocess.CompletedProcess result
        """
        iso_render_path = os.path.join(TOOLS_DIR, 'iso-sprite-renderer', 'iso_render.py')
        
        cmd = [
            sys.executable,
            iso_render_path,
            '--input', obj_path,
            '--output', output_dir,
            '--size', str(self.config['sprite']['size']),
            '--ortho-scale', str(self.config['sprite']['ortho_scale']),
            '--no-normalize',
            '--camera-focus-x', '0',
            '--camera-focus-y', '0',
            '--no-sprite-sheet'
        ]
        
        return subprocess.run(cmd, capture_output=True, text=True)
    
    def _save_metadata(self):
        """Save metadata.json file."""
        metadata_path = os.path.join(self.output_dir, 'metadata.json')
        
        with open(metadata_path, 'w') as f:
            json.dump(self.metadata, f, indent=2)
        
        print(f"\n✓ Saved metadata: {metadata_path}")


def main():
    parser = argparse.ArgumentParser(
        description='Generate electrical component sprite database'
    )
    parser.add_argument(
        '--config',
        default='sprite_database_config.yaml',
        help='Path to configuration file (default: sprite_database_config.yaml)'
    )
    
    args = parser.parse_args()
    
    if not os.path.exists(args.config):
        print(f"ERROR: Config file not found: {args.config}")
        return 1
    
    db = SpriteDatabase(args.config)
    db.generate_all()
    
    return 0


if __name__ == "__main__":
    sys.exit(main())

