{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  name = "ebiten-engine-dev";
  
  buildInputs = with pkgs; [
    # Go toolchain
    go
    gcc
    pkg-config
    
    # GLFW (required for Ebiten)
    glfw
    
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
    glfw
    xorg.libX11
    xorg.libXcursor
    xorg.libXrandr
    xorg.libXinerama
    xorg.libXi
    xorg.libXxf86vm
  ] + ":/run/opengl-driver/lib";
  
  shellHook = ''
    # Set CGO flags for Ebiten/GLFW
    export CGO_CFLAGS="-I${pkgs.glfw}/include"
    export CGO_LDFLAGS="-L${pkgs.glfw}/lib"
    export PKG_CONFIG_PATH="${pkgs.glfw}/lib/pkgconfig:$PKG_CONFIG_PATH"
    
    echo "=================================================="
    echo "  Go WASM Engine Development Environment"
    echo "=================================================="
    echo ""
    echo "Available commands:"
    echo "  make build-desktop  - Build Ebiten desktop binary"
    echo "  make run-desktop    - Build and run desktop game"
    echo "  make test           - Run unit tests"
    echo "  make clean          - Clean build artifacts"
    echo ""
    echo "Environment:"
    echo "  Go: $(go version | cut -d' ' -f3-4)"
    echo "  CGO_CFLAGS: $CGO_CFLAGS"
    echo "  CGO_LDFLAGS: $CGO_LDFLAGS"
    echo ""
    echo "Tip: Type 'zsh' to use your shell (CGO flags will persist)"
    echo "=================================================="
    echo ""
  '';
}
