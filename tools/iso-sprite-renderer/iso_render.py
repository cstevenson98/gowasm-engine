#!/usr/bin/env python3
"""
Isometric Sprite Renderer CLI
Renders 3D models as isometric sprites using Blender.
"""
import argparse
import sys
import os
from pathlib import Path

# Add the tool directory to Python path
tool_dir = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, tool_dir)

from config import load_config, Config
from renderer import BlenderRenderer
from output import generate_sprite_sheets


def parse_args():
    """Parse command line arguments."""
    parser = argparse.ArgumentParser(
        description='Render 3D models as isometric sprites using Blender',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Render single model with defaults
  %(prog)s --input teapot.obj
  
  # Render with custom size and lighting
  %(prog)s --input teapot.obj --size 64 --light-dir 1,-1,1 --light-color 1,0.9,0.8
  
  # Batch render all models in directory
  %(prog)s --batch models/ --output ./sprites
  
  # Use config file with CLI overrides
  %(prog)s --config render.yaml --input teapot.obj --size 128
        """
    )
    
    # Input options
    input_group = parser.add_mutually_exclusive_group(required=True)
    input_group.add_argument(
        '--input', '-i',
        help='Path to single 3D model file'
    )
    input_group.add_argument(
        '--batch', '-b',
        help='Directory containing multiple 3D model files'
    )
    
    # Configuration
    parser.add_argument(
        '--config', '-c',
        help='Path to YAML configuration file'
    )
    
    # Output options
    parser.add_argument(
        '--output', '-o',
        default='./sprites',
        help='Output directory for rendered sprites (default: ./sprites)'
    )
    parser.add_argument(
        '--no-individual-pngs',
        action='store_true',
        help='Do not save individual PNG files'
    )
    parser.add_argument(
        '--no-sprite-sheet',
        action='store_true',
        help='Do not generate sprite sheet'
    )
    parser.add_argument(
        '--sprite-layout',
        choices=['horizontal', 'vertical', 'grid'],
        default='horizontal',
        help='Sprite sheet layout (default: horizontal)'
    )
    
    # Render settings
    parser.add_argument(
        '--size', '-s',
        help='Output sprite size in pixels (e.g., "50" or "50,50" or "50x50")'
    )
    parser.add_argument(
        '--ortho-scale',
        type=float,
        help='Orthographic camera scale (default: 2.0, lower = more zoomed in)'
    )
    parser.add_argument(
        '--rotate-x',
        type=int,
        choices=[0, 1, 2, 3],
        help='Rotate model around X axis: 0=0°, 1=90°, 2=180°, 3=270° (pitch)'
    )
    parser.add_argument(
        '--rotate-y',
        type=int,
        choices=[0, 1, 2, 3],
        help='Rotate model around Y axis: 0=0°, 1=90°, 2=180°, 3=270° (yaw)'
    )
    parser.add_argument(
        '--rotate-z',
        type=int,
        choices=[0, 1, 2, 3],
        help='Rotate model around Z axis: 0=0°, 1=90°, 2=180°, 3=270° (roll)'
    )
    parser.add_argument(
        '--show-ground-plane',
        action='store_true',
        help='Show reference square [-1,1] on ground plane with 0.2 unit border'
    )
    parser.add_argument(
        '--no-transparency',
        action='store_true',
        help='Disable transparent background'
    )
    
    # Lighting options
    parser.add_argument(
        '--light-dir',
        help='Sun light direction as X,Y,Z (default: 1,-1,1)'
    )
    parser.add_argument(
        '--light-color',
        help='Sun light color as R,G,B (0-1 range, default: 1,1,1)'
    )
    parser.add_argument(
        '--light-energy',
        type=float,
        help='Sun light energy/intensity (default: 3.0)'
    )
    parser.add_argument(
        '--ambient-color',
        help='Ambient light color as R,G,B (0-1 range, default: 0.3,0.3,0.3)'
    )
    parser.add_argument(
        '--ambient-energy',
        type=float,
        help='Ambient light energy/intensity (default: 0.5)'
    )
    
    # Blender options
    parser.add_argument(
        '--blender-path',
        default='blender',
        help='Path to Blender executable (default: blender from PATH)'
    )
    
    # Misc
    parser.add_argument(
        '--verbose', '-v',
        action='store_true',
        help='Verbose output'
    )
    
    return parser.parse_args()


def main():
    """Main entry point."""
    args = parse_args()
    
    # Print banner
    print("=" * 70)
    print("Isometric Sprite Renderer".center(70))
    print("Powered by Blender".center(70))
    print("=" * 70)
    print()
    
    try:
        # Load configuration
        cli_args = vars(args)
        config = load_config(args.config, cli_args)
        
        if args.verbose:
            print("Configuration:")
            print(f"  Output directory: {config.output.directory}")
            print(f"  Sprite size: {config.render.size[0]}x{config.render.size[1]}")
            print(f"  Directions: {config.render.directions}")
            print(f"  Light direction: {config.lighting.sun_direction}")
            print()
        
        # Initialize renderer
        renderer = BlenderRenderer(blender_path=args.blender_path)
        
        # Render model(s)
        results = {}
        
        if args.input:
            # Single model
            print(f"Input: {args.input}")
            print()
            
            model_path = args.input
            model_name = Path(model_path).stem
            
            output_files = renderer.render_model(model_path, config, model_name)
            results[model_name] = output_files
            
        elif args.batch:
            # Batch processing
            print(f"Batch processing directory: {args.batch}")
            print()
            
            results = renderer.render_batch(args.batch, config)
        
        # Generate sprite sheets if requested
        if config.output.sprite_sheet and results:
            print("\n" + "=" * 70)
            print("Generating sprite sheets...")
            print("=" * 70)
            
            sprite_sheets = generate_sprite_sheets(
                results,
                config.output.directory,
                config.render.directions,
                layout=args.sprite_layout
            )
            
            # Report sprite sheet generation
            num_success = sum(1 for ss, _ in sprite_sheets.values() if ss is not None)
            print(f"\n✓ Generated {num_success}/{len(sprite_sheets)} sprite sheets")
        
        # Final summary
        print("\n" + "=" * 70)
        print("RENDERING COMPLETE".center(70))
        print("=" * 70)
        
        total_sprites = sum(len(files) for files in results.values())
        print(f"\nRendered {len(results)} model(s)")
        print(f"Total sprites: {total_sprites}")
        print(f"Output directory: {os.path.abspath(config.output.directory)}")
        
        if config.output.individual_pngs:
            print(f"  ✓ Individual PNG files")
        if config.output.sprite_sheet:
            print(f"  ✓ Sprite sheets ({args.sprite_layout} layout)")
        
        print("\n✓✓✓ Success!")
        
        return 0
        
    except KeyboardInterrupt:
        print("\n\nInterrupted by user")
        return 130
    except Exception as e:
        print(f"\n❌ ERROR: {e}", file=sys.stderr)
        if args.verbose:
            import traceback
            traceback.print_exc()
        return 1


if __name__ == '__main__':
    sys.exit(main())

