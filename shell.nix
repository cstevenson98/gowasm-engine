{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  name = "gowasm-engine-dev";
  
  buildInputs = with pkgs; [
    # Go toolchain
    go
    gcc
    pkg-config
    
    # X11 libraries for Ebiten window management
    xorg.libX11
    xorg.libXcursor
    xorg.libXrandr
    xorg.libXinerama
    xorg.libXi
    xorg.libXxf86vm
    
    # OpenGL libraries for Ebiten rendering
    libGL
    libglvnd
    
    # Development tools
    git
    gnumake
    zsh
  ];
  
  # Set library paths for runtime linking
  LD_LIBRARY_PATH = with pkgs; lib.makeLibraryPath [
    libGL
    libglvnd
    xorg.libX11
    xorg.libXcursor
    xorg.libXrandr
    xorg.libXinerama
    xorg.libXi
    xorg.libXxf86vm
  ] + ":/run/opengl-driver/lib";
  
  shellHook = ''
    echo "=================================================="
    echo "  Go WASM Engine Development Environment"
    echo "=================================================="
    echo ""
    echo "Available commands:"
    echo "  make build-desktop  - Build Ebiten desktop binary"
    echo "  make run-desktop    - Build and run desktop game"
    echo "  make build-wasm     - Build WebGPU WASM binary"
    echo "  make test           - Run unit tests"
    echo "  make clean          - Clean build artifacts"
    echo ""
    echo "Go version: $(go version)"
    echo "=================================================="
    echo ""
    
    # Source global zsh configuration if available
    if [ -f ~/.zshrc ]; then
      export ZDOTDIR=$HOME
      exec zsh
    fi
  '';
}
