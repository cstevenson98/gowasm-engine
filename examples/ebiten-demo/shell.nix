{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  name = "ebiten-dev-shell";
  
  buildInputs = with pkgs; [
    go
    gcc
    pkg-config
    
    # X11 libraries for window management
    xorg.libX11
    xorg.libXcursor
    xorg.libXrandr
    xorg.libXinerama
    xorg.libXi
    xorg.libXxf86vm
    
    # OpenGL libraries  
    libGL
    libglvnd
    
    # For runtime
    makeWrapper
  ];
  
  # Set library paths for runtime
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
    echo "Ebiten development environment loaded"
    echo "LD_LIBRARY_PATH set for OpenGL support"
    echo ""
    echo "To build: go build -o ebiten-demo main.go"
    echo "To run: ./ebiten-demo"
  '';
}
