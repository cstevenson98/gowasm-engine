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
    
    # Sparse linear algebra (SuperLU for power flow solver)
    superlu
    
    # Development tools
    git
    gnumake
    zsh
    direnv
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
    superlu
  ] + ":/run/opengl-driver/lib";
  
  shellHook = ''
    # Set CGO flags for Ebiten/GLFW/SuperLU
    export CGO_CFLAGS="-I${pkgs.glfw}/include -I${pkgs.superlu}/include"
    export CGO_LDFLAGS="-L${pkgs.glfw}/lib -L${pkgs.superlu}/lib -lsuperlu"
    export PKG_CONFIG_PATH="${pkgs.glfw}/lib/pkgconfig:$PKG_CONFIG_PATH"
    
    # Only show welcome message in interactive shells
    if [ -t 1 ]; then
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
      echo "Examples:"
      echo "  cd examples/grid-sim-game && go build ./..."
      echo "  cd examples/grid-sim-game && go test ./..."
      echo ""
      echo "Environment:"
      echo "  Go: $(go version | cut -d' ' -f3-4)"
      echo "  SuperLU: ${pkgs.superlu}/lib/libsuperlu.so (sparse solver)"
      echo "  CGO_CFLAGS: $CGO_CFLAGS"
      echo "  CGO_LDFLAGS: $CGO_LDFLAGS"
      echo ""
      echo "Tip: Type 'zsh' to use your shell (CGO flags will persist)"
      echo "=================================================="
      echo ""
    fi
  '';
}
