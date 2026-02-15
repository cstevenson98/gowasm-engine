"""
Sprite sheet generation and output utilities.
"""
import os
import json
from typing import Optional
from pathlib import Path

try:
    from PIL import Image
except ImportError:
    Image = None


class SpriteSheetGenerator:
    """Generate sprite sheets from individual sprite images."""
    
    def __init__(self):
        """Initialize generator."""
        if Image is None:
            raise ImportError(
                "Pillow is required for sprite sheet generation. "
                "Install it with: pip install Pillow"
            )
    
    def generate_sprite_sheet(
        self,
        sprite_files: list[str],
        output_path: str,
        layout: str = "horizontal",
        include_metadata: bool = True
    ) -> tuple[str, Optional[str]]:
        """
        Generate a sprite sheet from individual sprite files.
        
        Args:
            sprite_files: List of paths to individual sprite PNG files
            output_path: Path for output sprite sheet
            layout: Layout style - "horizontal", "vertical", or "grid"
            include_metadata: Whether to generate metadata JSON file
            
        Returns:
            Tuple of (sprite_sheet_path, metadata_path)
        """
        if not sprite_files:
            raise ValueError("No sprite files provided")
        
        # Load all sprites
        sprites = []
        for filepath in sprite_files:
            if not os.path.exists(filepath):
                print(f"WARNING: Sprite file not found: {filepath}")
                continue
            try:
                img = Image.open(filepath)
                sprites.append((img, filepath))
            except Exception as e:
                print(f"ERROR loading {filepath}: {e}")
        
        if not sprites:
            raise RuntimeError("No sprites could be loaded")
        
        # Get sprite dimensions (assume all same size)
        sprite_width, sprite_height = sprites[0][0].size
        num_sprites = len(sprites)
        
        # Calculate sheet dimensions based on layout
        if layout == "horizontal":
            sheet_width = sprite_width * num_sprites
            sheet_height = sprite_height
            grid_cols = num_sprites
            grid_rows = 1
        elif layout == "vertical":
            sheet_width = sprite_width
            sheet_height = sprite_height * num_sprites
            grid_cols = 1
            grid_rows = num_sprites
        elif layout == "grid":
            # Try to make a square-ish grid
            grid_cols = int(num_sprites ** 0.5) + (1 if num_sprites ** 0.5 % 1 else 0)
            grid_rows = (num_sprites + grid_cols - 1) // grid_cols
            sheet_width = sprite_width * grid_cols
            sheet_height = sprite_height * grid_rows
        else:
            raise ValueError(f"Unknown layout: {layout}")
        
        # Create sprite sheet
        sprite_sheet = Image.new('RGBA', (sheet_width, sheet_height), (0, 0, 0, 0))
        
        # Paste sprites
        metadata = {
            "sprite_width": sprite_width,
            "sprite_height": sprite_height,
            "num_sprites": num_sprites,
            "layout": layout,
            "grid_cols": grid_cols,
            "grid_rows": grid_rows,
            "sprites": []
        }
        
        for i, (sprite, filepath) in enumerate(sprites):
            if layout == "horizontal":
                x = i * sprite_width
                y = 0
            elif layout == "vertical":
                x = 0
                y = i * sprite_height
            else:  # grid
                col = i % grid_cols
                row = i // grid_cols
                x = col * sprite_width
                y = row * sprite_height
            
            sprite_sheet.paste(sprite, (x, y))
            
            # Extract angle from filename if present
            filename = Path(filepath).name
            angle = None
            if "_angle_" in filename:
                try:
                    angle_str = filename.split("_angle_")[1].split(".")[0]
                    angle = int(angle_str)
                except (IndexError, ValueError):
                    pass
            
            metadata["sprites"].append({
                "index": i,
                "x": x,
                "y": y,
                "width": sprite_width,
                "height": sprite_height,
                "angle": angle,
                "source_file": Path(filepath).name
            })
        
        # Save sprite sheet
        sprite_sheet.save(output_path, 'PNG')
        print(f"✓ Sprite sheet saved: {output_path}")
        
        # Save metadata
        metadata_path = None
        if include_metadata:
            metadata_path = output_path.replace('.png', '.json')
            with open(metadata_path, 'w') as f:
                json.dump(metadata, f, indent=2)
            print(f"✓ Metadata saved: {metadata_path}")
        
        return output_path, metadata_path
    
    def generate_for_model(
        self,
        model_name: str,
        sprite_dir: str,
        directions: int = 8,
        layout: str = "horizontal"
    ) -> tuple[Optional[str], Optional[str]]:
        """
        Generate sprite sheet for a specific model.
        
        Args:
            model_name: Name of the model
            sprite_dir: Directory containing individual sprite files
            directions: Number of rotation directions
            layout: Layout style
            
        Returns:
            Tuple of (sprite_sheet_path, metadata_path) or (None, None) if failed
        """
        # Find all sprite files for this model
        angle_step = 360 / directions
        sprite_files = []
        
        for i in range(directions):
            angle = int(i * angle_step)
            filename = f"{model_name}_angle_{angle}.png"
            filepath = os.path.join(sprite_dir, filename)
            
            if os.path.exists(filepath):
                sprite_files.append(filepath)
        
        if not sprite_files:
            print(f"WARNING: No sprite files found for {model_name}")
            return None, None
        
        # Generate sprite sheet
        output_path = os.path.join(sprite_dir, f"{model_name}_spritesheet.png")
        
        try:
            return self.generate_sprite_sheet(
                sprite_files,
                output_path,
                layout=layout,
                include_metadata=True
            )
        except Exception as e:
            print(f"ERROR generating sprite sheet for {model_name}: {e}")
            return None, None


def generate_sprite_sheets(
    models: dict[str, list[str]],
    sprite_dir: str,
    directions: int = 8,
    layout: str = "horizontal"
) -> dict[str, tuple[Optional[str], Optional[str]]]:
    """
    Generate sprite sheets for multiple models.
    
    Args:
        models: Dictionary mapping model names to list of sprite files
        sprite_dir: Directory containing sprites
        directions: Number of directions
        layout: Layout style
        
    Returns:
        Dictionary mapping model names to (sprite_sheet_path, metadata_path) tuples
    """
    generator = SpriteSheetGenerator()
    results = {}
    
    for model_name in models.keys():
        print(f"\nGenerating sprite sheet for {model_name}...")
        result = generator.generate_for_model(
            model_name,
            sprite_dir,
            directions,
            layout
        )
        results[model_name] = result
    
    return results

