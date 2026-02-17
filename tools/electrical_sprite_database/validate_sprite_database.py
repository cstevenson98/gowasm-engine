#!/usr/bin/env python3
"""
Validate sprite database integrity.

Checks that all sprites exist, have correct dimensions, and proper transparency.
"""

import os
import sys
import json
import argparse
from pathlib import Path
from typing import List, Tuple, Dict

try:
    from PIL import Image
    HAS_PIL = True
except ImportError:
    HAS_PIL = False
    print("WARNING: PIL/Pillow not installed. Image validation will be limited.")
    print("Install with: pip install Pillow")


class SpriteValidator:
    """Validate sprite database."""
    
    def __init__(self, database_dir: str):
        """
        Initialize validator.
        
        Args:
            database_dir: Path to sprite database directory
        """
        self.database_dir = Path(database_dir)
        self.metadata_path = self.database_dir / 'metadata.json'
        self.errors = []
        self.warnings = []
        self.stats = {
            'total_sprites': 0,
            'valid_sprites': 0,
            'missing_sprites': 0,
            'invalid_size': 0,
            'no_transparency': 0,
        }
    
    def validate(self) -> bool:
        """
        Run full validation.
        
        Returns:
            True if all checks pass, False otherwise
        """
        print("="*70)
        print("SPRITE DATABASE VALIDATOR")
        print("="*70)
        print(f"Database: {self.database_dir}")
        print()
        
        # Check database directory exists
        if not self.database_dir.exists():
            self.errors.append(f"Database directory not found: {self.database_dir}")
            self._print_results()
            return False
        
        # Check metadata exists
        if not self.metadata_path.exists():
            self.errors.append(f"Metadata file not found: {self.metadata_path}")
            self._print_results()
            return False
        
        # Load metadata
        with open(self.metadata_path, 'r') as f:
            metadata = json.load(f)
        
        expected_size = metadata.get('sprite_size', 64)
        print(f"Expected sprite size: {expected_size}x{expected_size}")
        print()
        
        # Validate each component
        for comp_type, comp_data in metadata.get('components', {}).items():
            print(f"Validating {comp_type}...")
            
            # Check base sprite
            base_path = self.database_dir / comp_data['base']
            self._validate_sprite(base_path, expected_size, f"{comp_type} base")
            
            # Check connection sprites
            for conn in comp_data.get('connections', []):
                conn_path = self.database_dir / conn['file']
                label = f"{comp_type} → {conn['target']} {conn['offset']} phase {conn['phase']}"
                self._validate_sprite(conn_path, expected_size, label)
        
        # Print results
        self._print_results()
        
        return len(self.errors) == 0
    
    def _validate_sprite(self, sprite_path: Path, expected_size: int, label: str):
        """
        Validate a single sprite file.
        
        Args:
            sprite_path: Path to sprite file
            expected_size: Expected width/height
            label: Sprite description for error messages
        """
        self.stats['total_sprites'] += 1
        
        # Check file exists
        if not sprite_path.exists():
            self.errors.append(f"Missing sprite: {label} ({sprite_path})")
            self.stats['missing_sprites'] += 1
            return
        
        # Check file size
        file_size = sprite_path.stat().st_size
        if file_size == 0:
            self.errors.append(f"Empty sprite file: {label} ({sprite_path})")
            return
        
        # Check image properties (if PIL available)
        if HAS_PIL:
            try:
                with Image.open(sprite_path) as img:
                    # Check dimensions
                    width, height = img.size
                    if width != expected_size or height != expected_size:
                        self.errors.append(
                            f"Invalid size: {label} - expected {expected_size}x{expected_size}, "
                            f"got {width}x{height}"
                        )
                        self.stats['invalid_size'] += 1
                        return
                    
                    # Check for alpha channel
                    if img.mode not in ('RGBA', 'LA', 'PA'):
                        self.warnings.append(
                            f"No alpha channel: {label} - mode is {img.mode}"
                        )
                        self.stats['no_transparency'] += 1
                    
                    # Check if has any transparency
                    if 'A' in img.mode:
                        alpha = img.getchannel('A')
                        alpha_data = alpha.getdata()
                        if all(a == 255 for a in alpha_data):
                            self.warnings.append(
                                f"No transparent pixels: {label} (fully opaque)"
                            )
                
                self.stats['valid_sprites'] += 1
                
            except Exception as e:
                self.errors.append(f"Failed to read image: {label} - {e}")
        else:
            # Without PIL, just check file exists and has size
            self.stats['valid_sprites'] += 1
    
    def _print_results(self):
        """Print validation results."""
        print()
        print("="*70)
        print("VALIDATION RESULTS")
        print("="*70)
        
        # Statistics
        print("\nStatistics:")
        print(f"  Total sprites checked: {self.stats['total_sprites']}")
        print(f"  Valid sprites: {self.stats['valid_sprites']}")
        print(f"  Missing sprites: {self.stats['missing_sprites']}")
        print(f"  Invalid size: {self.stats['invalid_size']}")
        print(f"  No transparency: {self.stats['no_transparency']}")
        
        # Errors
        if self.errors:
            print(f"\n❌ ERRORS ({len(self.errors)}):")
            for error in self.errors:
                print(f"  - {error}")
        
        # Warnings
        if self.warnings:
            print(f"\n⚠️  WARNINGS ({len(self.warnings)}):")
            for warning in self.warnings:
                print(f"  - {warning}")
        
        # Summary
        print()
        if self.errors:
            print("❌ VALIDATION FAILED")
        else:
            print("✓ VALIDATION PASSED")
        
        print("="*70)


def main():
    parser = argparse.ArgumentParser(
        description='Validate sprite database integrity'
    )
    parser.add_argument(
        'database_dir',
        nargs='?',
        default='output/sprite_database',
        help='Path to sprite database directory (default: output/sprite_database)'
    )
    parser.add_argument(
        '--list-missing',
        action='store_true',
        help='List only missing sprites'
    )
    
    args = parser.parse_args()
    
    validator = SpriteValidator(args.database_dir)
    success = validator.validate()
    
    return 0 if success else 1


if __name__ == "__main__":
    sys.exit(main())


