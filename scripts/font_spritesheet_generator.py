#!/usr/bin/python3
"""
Font Sprite Sheet Generator

Generates sprite sheets from system fonts with character mapping metadata.
"""

import argparse
import json
import math
import os
import sys
from pathlib import Path
from typing import Dict, List, Tuple

try:
    from PIL import Image, ImageDraw, ImageFont
except ImportError:
    print("Error: Pillow is required. Install with: pip install Pillow")
    sys.exit(1)


# Default character set: A-Z, a-z, 0-9, and common punctuation
DEFAULT_CHARACTER_SET = (
    "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    "abcdefghijklmnopqrstuvwxyz"
    "0123456789"
    ".,!?;:'\"()[]{}+-=*/\\|@#$%^&*_~`<> "
)

CELL_SIZE = 16  # Fixed 16x16 cell size


def find_system_font(font_name: str) -> str:
    """
    Find a system font by name. Tries common font locations.
    """
    # Common font locations on Linux, macOS, and Windows
    font_paths = [
        f"/usr/share/fonts/truetype/{font_name.lower()}/{font_name}.ttf",
        f"/usr/share/fonts/truetype/{font_name.lower()}.ttf",
        f"/usr/share/fonts/TTF/{font_name}.ttf",
        f"/System/Library/Fonts/{font_name}.ttf",
        f"C:\\Windows\\Fonts\\{font_name}.ttf",
        f"/usr/share/fonts/truetype/dejavu/DejaVu{font_name}.ttf",
        f"/usr/share/fonts/truetype/liberation/Liberation{font_name}.ttf",
    ]
    
    # Try exact path first
    if os.path.exists(font_name):
        return font_name
    
    # Try common paths
    for path in font_paths:
        if os.path.exists(path):
            return path
    
    # Fallback to default font
    print(f"Warning: Font '{font_name}' not found. Trying default font.")
    try:
        return ImageFont.load_default()
    except Exception:
        raise ValueError(f"Could not find font: {font_name}")


def calculate_optimal_font_size(font_path: str, initial_size: int, target_size: int = CELL_SIZE) -> Tuple[ImageFont.FreeTypeFont, int]:
    """
    Calculate the optimal font size that fits within the target cell size.
    Returns the font object and the actual size used.
    """
    # Leave some padding (2px on each side = 4px total)
    max_fit_size = target_size - 4
    
    current_size = initial_size
    
    # Try decreasing sizes until we find one that fits
    for size in range(initial_size, 4, -1):
        try:
            if isinstance(font_path, str):
                font = ImageFont.truetype(font_path, size)
            else:
                # Default font doesn't support sizing
                return font_path, 10
            
            # Test with a few characters to check bounds
            test_chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789Wm@#"
            
            # Create a test image to measure bounds
            test_img = Image.new('RGBA', (100, 100), (0, 0, 0, 0))
            test_draw = ImageDraw.Draw(test_img)
            
            max_width = 0
            max_height = 0
            
            for char in test_chars[:10]:  # Test first 10 chars for speed
                bbox = test_draw.textbbox((0, 0), char, font=font)
                width = bbox[2] - bbox[0]
                height = bbox[3] - bbox[1]
                max_width = max(max_width, width)
                max_height = max(max_height, height)
            
            # Check if it fits
            if max_width <= max_fit_size and max_height <= max_fit_size:
                return font, size
            
        except Exception as e:
            print(f"Warning: Could not load font at size {size}: {e}")
            continue
    
    # If we get here, use size 8 as absolute minimum
    try:
        if isinstance(font_path, str):
            return ImageFont.truetype(font_path, 8), 8
        else:
            return font_path, 8
    except Exception:
        return ImageFont.load_default(), 8


def generate_sprite_sheet(
    font_name: str,
    font_size: int,
    output_dir: str,
    character_set: str = DEFAULT_CHARACTER_SET,
    columns: int = 10
) -> Tuple[str, str]:
    """
    Generate a sprite sheet and metadata JSON for the given font.
    
    Returns: (image_path, json_path)
    """
    # Find and load font
    font_path = find_system_font(font_name)
    font, actual_size = calculate_optimal_font_size(font_path, font_size)
    
    # Calculate grid dimensions
    num_chars = len(character_set)
    rows = math.ceil(num_chars / columns)
    img_width = columns * CELL_SIZE
    img_height = rows * CELL_SIZE
    
    # Create image with transparency
    img = Image.new('RGBA', (img_width, img_height), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    
    # Character mapping metadata
    character_map = {}
    
    # Render each character
    for i, char in enumerate(character_set):
        col = i % columns
        row = i // columns
        
        # Cell position in pixels
        x = col * CELL_SIZE
        y = row * CELL_SIZE
        
        # Center of the cell
        center_x = x + CELL_SIZE // 2
        center_y = y + CELL_SIZE // 2
        
        # Draw character centered in cell
        try:
            draw.text(
                (center_x, center_y),
                char,
                font=font,
                fill=(255, 255, 255, 255),  # White with full opacity
                anchor='mm'  # Middle-middle anchor
            )
        except Exception as e:
            print(f"Warning: Could not render character '{char}': {e}")
        
        # Calculate UV coordinates
        u0 = x / img_width
        v0 = y / img_height
        u1 = (x + CELL_SIZE) / img_width
        v1 = (y + CELL_SIZE) / img_height
        
        # Store metadata
        character_map[char] = {
            "index": i,
            "x": x,
            "y": y,
            "u0": round(u0, 6),
            "v0": round(v0, 6),
            "u1": round(u1, 6),
            "v1": round(v1, 6)
        }
    
    # Create output directory if it doesn't exist
    Path(output_dir).mkdir(parents=True, exist_ok=True)
    
    # Generate output filenames
    base_name = f"{font_name.replace(' ', '_')}_{actual_size}"
    image_filename = f"{base_name}.sheet.png"
    json_filename = f"{base_name}.sheet.json"
    
    image_path = os.path.join(output_dir, image_filename)
    json_path = os.path.join(output_dir, json_filename)
    
    # Save image
    img.save(image_path, 'PNG')
    
    # Create metadata
    metadata = {
        "font_name": font_name,
        "font_size": actual_size,
        "cell_width": CELL_SIZE,
        "cell_height": CELL_SIZE,
        "columns": columns,
        "rows": rows,
        "image_width": img_width,
        "image_height": img_height,
        "character_count": num_chars,
        "character_map": character_map
    }
    
    # Save metadata JSON
    with open(json_path, 'w', encoding='utf-8') as f:
        json.dump(metadata, f, indent=2, ensure_ascii=False)
    
    return image_path, json_path


def main():
    parser = argparse.ArgumentParser(
        description='Generate sprite sheets from system fonts'
    )
    parser.add_argument(
        '--font',
        type=str,
        default='DejaVuSans',
        help='Font name (default: DejaVuSans)'
    )
    parser.add_argument(
        '--size',
        type=int,
        default=12,
        help='Font size in points (default: 12, will auto-adjust to fit 16x16)'
    )
    parser.add_argument(
        '--sizes',
        type=int,
        nargs='+',
        help='Multiple font sizes (generates separate sheets for each)'
    )
    parser.add_argument(
        '--output',
        type=str,
        default='./output',
        help='Output directory (default: ./output)'
    )
    parser.add_argument(
        '--columns',
        type=int,
        default=10,
        help='Number of columns in the grid (default: 10)'
    )
    parser.add_argument(
        '--characters',
        type=str,
        help='Custom character set (default: A-Z, a-z, 0-9, punctuation)'
    )
    
    args = parser.parse_args()
    
    # Determine which sizes to generate
    sizes = args.sizes if args.sizes else [args.size]
    
    # Use custom character set if provided
    character_set = args.characters if args.characters else DEFAULT_CHARACTER_SET
    
    print(f"Generating sprite sheets for font: {args.font}")
    print(f"Character set: {len(character_set)} characters")
    print(f"Cell size: {CELL_SIZE}x{CELL_SIZE} pixels")
    print(f"Output directory: {args.output}")
    print()
    
    # Generate sprite sheet for each size
    for size in sizes:
        try:
            print(f"Generating sprite sheet for size {size}...")
            image_path, json_path = generate_sprite_sheet(
                args.font,
                size,
                args.output,
                character_set,
                args.columns
            )
            print(f"  ✓ Created: {image_path}")
            print(f"  ✓ Created: {json_path}")
        except Exception as e:
            print(f"  ✗ Error generating size {size}: {e}")
            import traceback
            traceback.print_exc()
    
    print()
    print("Done!")


if __name__ == '__main__':
    main()

