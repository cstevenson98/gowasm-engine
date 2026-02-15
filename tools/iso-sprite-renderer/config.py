"""
Configuration management for isometric sprite renderer.
Supports YAML config files and CLI argument merging.
"""
import os
import yaml
from typing import Any, Optional
from dataclasses import dataclass, asdict, field


@dataclass
class RenderConfig:
    """Render settings."""
    size: list[int] = field(default_factory=lambda: [50, 50])
    directions: int = 8
    ortho_scale: float = 2.0  # Lower = more zoomed in, higher = more zoomed out
    # Model rotation adjustments (90° increments: 0, 1, 2, 3 = 0°, 90°, 180°, 270°)
    rotate_x: int = 0  # Rotation around X axis (pitch)
    rotate_y: int = 0  # Rotation around Y axis (yaw)
    rotate_z: int = 0  # Rotation around Z axis (roll)
    show_ground_plane: bool = False  # Show reference square on ground plane


@dataclass
class LightingConfig:
    """Lighting settings."""
    sun_direction: list[float] = field(default_factory=lambda: [1.0, -1.0, 1.0])
    sun_color: list[float] = field(default_factory=lambda: [1.0, 1.0, 1.0])
    sun_energy: float = 3.0
    ambient_color: list[float] = field(default_factory=lambda: [0.3, 0.3, 0.3])
    ambient_energy: float = 0.5


@dataclass
class CameraConfig:
    """Camera settings."""
    elevation_angle: float = 35.264  # True isometric angle
    rotation_offset: float = 45.0    # Starting rotation angle


@dataclass
class OutputConfig:
    """Output settings."""
    directory: str = "./sprites"
    individual_pngs: bool = True
    sprite_sheet: bool = True
    transparent_bg: bool = True


@dataclass
class Config:
    """Complete configuration."""
    render: RenderConfig = field(default_factory=RenderConfig)
    lighting: LightingConfig = field(default_factory=LightingConfig)
    camera: CameraConfig = field(default_factory=CameraConfig)
    output: OutputConfig = field(default_factory=OutputConfig)
    
    def to_dict(self) -> dict[str, Any]:
        """Convert config to dictionary."""
        return {
            'render': asdict(self.render),
            'lighting': asdict(self.lighting),
            'camera': asdict(self.camera),
            'output': asdict(self.output),
        }
    
    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> 'Config':
        """Create config from dictionary."""
        config = cls()
        
        if 'render' in data:
            for key, value in data['render'].items():
                if hasattr(config.render, key):
                    setattr(config.render, key, value)
        
        if 'lighting' in data:
            for key, value in data['lighting'].items():
                if hasattr(config.lighting, key):
                    setattr(config.lighting, key, value)
        
        if 'camera' in data:
            for key, value in data['camera'].items():
                if hasattr(config.camera, key):
                    setattr(config.camera, key, value)
        
        if 'output' in data:
            for key, value in data['output'].items():
                if hasattr(config.output, key):
                    setattr(config.output, key, value)
        
        return config
    
    @classmethod
    def from_yaml(cls, filepath: str) -> 'Config':
        """Load config from YAML file."""
        if not os.path.exists(filepath):
            raise FileNotFoundError(f"Config file not found: {filepath}")
        
        with open(filepath, 'r') as f:
            data = yaml.safe_load(f)
        
        return cls.from_dict(data or {})
    
    def merge_cli_args(self, args: dict[str, Any]) -> None:
        """
        Merge CLI arguments into config.
        CLI arguments take precedence over config file values.
        
        Args:
            args: Dictionary of CLI arguments
        """
        # Render settings
        if args.get('size'):
            size = args['size']
            if isinstance(size, str):
                # Parse "50x50" or "50,50" or "50" format
                if 'x' in size or ',' in size:
                    size = size.replace('x', ',')
                    self.render.size = [int(x.strip()) for x in size.split(',')]
                else:
                    # Single number means square (50 -> 50x50)
                    s = int(size.strip())
                    self.render.size = [s, s]
            elif isinstance(size, int):
                # Single int means square
                self.render.size = [size, size]
            elif isinstance(size, (list, tuple)):
                self.render.size = list(size)
        
        if args.get('ortho_scale') is not None:
            self.render.ortho_scale = float(args['ortho_scale'])
        
        if args.get('rotate_x') is not None:
            self.render.rotate_x = int(args['rotate_x'])
        
        if args.get('rotate_y') is not None:
            self.render.rotate_y = int(args['rotate_y'])
        
        if args.get('rotate_z') is not None:
            self.render.rotate_z = int(args['rotate_z'])
        
        if args.get('show_ground_plane'):
            self.render.show_ground_plane = True
        
        # Lighting settings
        if args.get('light_dir'):
            light_dir = args['light_dir']
            if isinstance(light_dir, str):
                self.lighting.sun_direction = [float(x.strip()) for x in light_dir.split(',')]
            elif isinstance(light_dir, (list, tuple)):
                self.lighting.sun_direction = list(light_dir)
        
        if args.get('light_color'):
            light_color = args['light_color']
            if isinstance(light_color, str):
                self.lighting.sun_color = [float(x.strip()) for x in light_color.split(',')]
            elif isinstance(light_color, (list, tuple)):
                self.lighting.sun_color = list(light_color)
        
        if args.get('light_energy') is not None:
            self.lighting.sun_energy = float(args['light_energy'])
        
        if args.get('ambient_color'):
            ambient_color = args['ambient_color']
            if isinstance(ambient_color, str):
                self.lighting.ambient_color = [float(x.strip()) for x in ambient_color.split(',')]
            elif isinstance(ambient_color, (list, tuple)):
                self.lighting.ambient_color = list(ambient_color)
        
        if args.get('ambient_energy') is not None:
            self.lighting.ambient_energy = float(args['ambient_energy'])
        
        # Output settings
        if args.get('output'):
            self.output.directory = args['output']
        
        if args.get('no_individual_pngs'):
            self.output.individual_pngs = False
        
        if args.get('no_sprite_sheet'):
            self.output.sprite_sheet = False
        
        if args.get('no_transparency'):
            self.output.transparent_bg = False


def load_config(config_file: Optional[str] = None, cli_args: Optional[dict[str, Any]] = None) -> Config:
    """
    Load configuration from file and merge with CLI arguments.
    
    Args:
        config_file: Optional path to YAML config file
        cli_args: Optional dictionary of CLI arguments
        
    Returns:
        Merged Config object
    """
    # Start with default config
    if config_file and os.path.exists(config_file):
        config = Config.from_yaml(config_file)
    else:
        config = Config()
    
    # Merge CLI arguments
    if cli_args:
        config.merge_cli_args(cli_args)
    
    return config

