# Font Sprite Sheet Generator

A Python script that generates PNG sprite sheets from system fonts with JSON metadata for use in game engines.

## Features

- Generates 16x16 pixel character cells in a grid layout
- Supports A-Z, a-z, 0-9, and common punctuation characters
- Auto-adjusts font size to fit within 16x16 cells
- Creates PNG with transparency
- Outputs JSON metadata with character-to-sprite mapping and UV coordinates
- Supports multiple font sizes in a single run

## Installation

Install the required dependencies using system python3:

```bash
python3 -m pip install --user -r requirements.txt
```

**Note:** Virtual environments are currently incompatible with Cursor. Use system python3 instead.

## Usage

### Basic Usage

Generate a sprite sheet with a specific font and size:

```bash
python3 font_spritesheet_generator.py --font DejaVuSans --size 12 --output ../assets/fonts/
```

### Multiple Font Sizes

Generate sprite sheets for multiple sizes at once:

```bash
python3 font_spritesheet_generator.py --font Courier --sizes 10 12 14 16 --output ../assets/fonts/
```

### Custom Character Set

Specify a custom set of characters:

```bash
python3 font_spritesheet_generator.py --font Arial --size 12 --characters "ABCD1234!@#$" --output ./output/
```

### Available Options

- `--font` - Font name (default: DejaVuSans)
- `--size` - Font size in points (default: 12, auto-adjusted to fit 16x16)
- `--sizes` - Multiple font sizes (generates separate sheets for each)
- `--output` - Output directory (default: ./output)
- `--columns` - Number of columns in grid (default: 10)
- `--characters` - Custom character set (default: A-Z, a-z, 0-9, punctuation)

## Output Format

The script generates two files per font size:

### PNG Sprite Sheet

Filename: `<font_name>_<font_size>.sheet.png`

- Transparent background
- 16x16 pixel cells arranged in a grid
- Characters centered in each cell

### JSON Metadata

Filename: `<font_name>_<font_size>.sheet.json`

Contains:
- Font information (name, size)
- Grid dimensions (columns, rows, cell size)
- Character map with:
  - Character index
  - Pixel coordinates (x, y)
  - UV coordinates (u0, v0, u1, v1) for texture sampling

Example JSON structure:

```json
{
  "font_name": "DejaVuSans",
  "font_size": 10,
  "cell_width": 16,
  "cell_height": 16,
  "columns": 10,
  "rows": 10,
  "image_width": 160,
  "image_height": 160,
  "character_count": 96,
  "character_map": {
    "A": {
      "index": 0,
      "x": 0,
      "y": 0,
      "u0": 0.0,
      "v0": 0.0,
      "u1": 0.1,
      "v1": 0.1
    },
    ...
  }
}
```

## Finding System Fonts

The script searches common font locations on Linux, macOS, and Windows:

**Linux:**
- `/usr/share/fonts/truetype/`
- `/usr/share/fonts/TTF/`

**macOS:**
- `/System/Library/Fonts/`

**Windows:**
- `C:\Windows\Fonts\`

You can also provide an absolute path to a `.ttf` file:

```bash
python3 font_spritesheet_generator.py --font /path/to/custom/font.ttf --size 12
```

To list available fonts on your system:

```bash
# Linux
fc-list | grep -i "font name"

# macOS
system_profiler SPFontsDataType

# Windows
dir C:\Windows\Fonts\*.ttf
```

## Integration with Game Engine

Use the generated sprite sheet and JSON in your game engine:

1. Load the PNG texture
2. Parse the JSON metadata
3. Use the UV coordinates to render specific characters
4. Map keyboard input to character lookups in the character_map

Example UV coordinate usage:
```go
// Get character 'A' from the sprite sheet
charData := metadata.CharacterMap["A"]
uvRect := types.UVRect{
    U0: charData.U0,
    V0: charData.V0,
    U1: charData.U1,
    V1: charData.V1,
}
canvas.DrawTexture(fontTexture, position, size, uvRect)
```

## Troubleshooting

**Font not found:**
- Ensure the font is installed on your system
- Try using the full path to the `.ttf` file
- The script will fall back to the default font if not found

**Characters appear cut off:**
- The script auto-adjusts font size to fit 16x16 cells
- If using a custom size, it will be reduced as needed
- Some fonts may require manual size adjustment

**Missing Pillow:**
```bash
python3 -m pip install --user Pillow
```

## License

Part of the gowasm-engine project.

