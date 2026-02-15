"""
Blender renderer orchestration.
Manages Blender subprocess execution for sprite rendering.
"""
import os
import json
import subprocess
import shutil
from typing import Optional
from pathlib import Path

from config import Config
from loaders import get_loader_for_file


class BlenderRenderer:
    """Orchestrates Blender rendering process."""
    
    def __init__(self, blender_path: str = "blender"):
        """
        Initialize renderer.
        
        Args:
            blender_path: Path to Blender executable (default: 'blender' from PATH)
        """
        self.blender_path = blender_path
        
        # Verify Blender is available
        if not self._check_blender():
            raise RuntimeError(
                f"Blender not found at '{blender_path}'. "
                "Please install Blender or provide correct path."
            )
    
    def _check_blender(self) -> bool:
        """Check if Blender is available."""
        try:
            result = subprocess.run(
                [self.blender_path, "--version"],
                capture_output=True,
                timeout=10
            )
            return result.returncode == 0
        except (subprocess.SubprocessError, FileNotFoundError):
            return False
    
    def render_model(
        self,
        model_path: str,
        config: Config,
        model_name: Optional[str] = None
    ) -> list[str]:
        """
        Render a 3D model as isometric sprites.
        
        Args:
            model_path: Path to the 3D model file
            config: Configuration object
            model_name: Optional name for output files (default: filename without extension)
            
        Returns:
            List of paths to generated PNG files
        """
        # Validate model file exists
        if not os.path.exists(model_path):
            raise FileNotFoundError(f"Model file not found: {model_path}")
        
        # Get loader for this file type
        loader = get_loader_for_file(model_path)
        if not loader:
            ext = os.path.splitext(model_path)[1]
            raise ValueError(f"Unsupported file format: {ext}")
        
        # Determine model name
        if not model_name:
            model_name = Path(model_path).stem
        
        # Create output directory
        output_dir = os.path.abspath(config.output.directory)
        os.makedirs(output_dir, exist_ok=True)
        
        # Prepare config for Blender script
        render_config = config.to_dict()
        render_config['mesh_path'] = os.path.abspath(model_path)
        render_config['model_name'] = model_name
        render_config['loader_type'] = 'obj'  # For now, only OBJ
        
        # Generate render script
        script_path = self._generate_render_script(render_config, output_dir)
        
        try:
            # Run Blender
            print(f"Rendering {model_name}...")
            self._run_blender(script_path)
            
            # Collect output files
            output_files = self._collect_output_files(output_dir, model_name, config.render.directions)
            
            print(f"✓ Generated {len(output_files)} sprite files")
            return output_files
            
        finally:
            # Cleanup temporary script
            if os.path.exists(script_path):
                os.remove(script_path)
    
    def _generate_render_script(self, config: dict, output_dir: str) -> str:
        """Generate Blender Python script with injected config."""
        # Read template
        template_path = os.path.join(
            os.path.dirname(__file__),
            'templates',
            'render_script.py'
        )
        
        with open(template_path, 'r') as f:
            template = f.read()
        
        # Inject config as JSON
        config_json = json.dumps(config, indent=2)
        script = template.replace('__CONFIG_JSON__', config_json)
        
        # Write temporary script
        temp_script = os.path.join(output_dir, '_temp_render_script.py')
        with open(temp_script, 'w') as f:
            f.write(script)
        
        return temp_script
    
    def _run_blender(self, script_path: str) -> None:
        """Execute Blender with the render script."""
        cmd = [
            self.blender_path,
            '--background',  # Run without GUI
            '--python', script_path
        ]
        
        try:
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=300  # 5 minute timeout
            )
            
            # Print output
            if result.stdout:
                print(result.stdout)
            
            # Always print stderr for debugging
            if result.stderr:
                print("STDERR:", result.stderr)
            
            if result.returncode != 0:
                print("ERROR: Blender rendering failed!")
                raise RuntimeError(f"Blender exited with code {result.returncode}")
                
        except subprocess.TimeoutExpired:
            raise RuntimeError("Blender rendering timed out after 5 minutes")
    
    def _collect_output_files(
        self,
        output_dir: str,
        model_name: str,
        directions: int
    ) -> list[str]:
        """Collect paths to all rendered PNG files."""
        output_files = []
        angle_step = 360 / directions
        
        for i in range(directions):
            angle = int(i * angle_step)
            filename = f"{model_name}_angle_{angle}.png"
            filepath = os.path.join(output_dir, filename)
            
            if os.path.exists(filepath):
                output_files.append(filepath)
            else:
                print(f"WARNING: Expected output file not found: {filepath}")
        
        return output_files
    
    def render_batch(
        self,
        directory: str,
        config: Config,
        extensions: Optional[list[str]] = None
    ) -> dict[str, list[str]]:
        """
        Render all supported models in a directory.
        
        Args:
            directory: Directory containing model files
            config: Configuration object
            extensions: Optional list of extensions to filter (e.g., ['.obj'])
            
        Returns:
            Dictionary mapping model names to list of output files
        """
        if not os.path.isdir(directory):
            raise NotADirectoryError(f"Not a directory: {directory}")
        
        # Find all model files
        model_files = []
        for filename in os.listdir(directory):
            filepath = os.path.join(directory, filename)
            if not os.path.isfile(filepath):
                continue
            
            # Check if we support this file
            if get_loader_for_file(filepath):
                # Filter by extensions if specified
                if extensions:
                    ext = os.path.splitext(filepath)[1].lower()
                    if ext not in extensions:
                        continue
                model_files.append(filepath)
        
        if not model_files:
            print(f"No supported model files found in {directory}")
            return {}
        
        print(f"Found {len(model_files)} model files to render")
        
        # Render each model
        results = {}
        for i, model_path in enumerate(model_files, 1):
            model_name = Path(model_path).stem
            print(f"\n[{i}/{len(model_files)}] Rendering {model_name}...")
            
            try:
                output_files = self.render_model(model_path, config, model_name)
                results[model_name] = output_files
            except Exception as e:
                print(f"ERROR rendering {model_name}: {e}")
                results[model_name] = []
        
        return results

