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

    # nix-shell forces SHELL=bash; restore zsh so children (Cursor terminals,
    # `cursor` launched from this shell, etc.) load ~/.zshrc + oh-my-zsh.
    export SHELL="${pkgs.zsh}/bin/zsh"

    # nix-shell points TMPDIR at /tmp/nix-shell-<pid>-* and deletes it when that
    # shell exits. Cursor (and its terminals) keep the stale path, so `go run`
    # fails with: creating work dir: stat /tmp/nix-shell-...: no such file.
    export TMPDIR="''${XDG_RUNTIME_DIR:-/tmp}"
    export TMP="$TMPDIR"
    export TEMP="$TMPDIR"
    export TEMPDIR="$TMPDIR"
    
    # Only show welcome message in interactive shells
    if [ -t 1 ]; then
      echo "=================================================="
      echo "  Go WASM Engine Development Environment"
      echo "=================================================="
      echo ""
      echo "Available commands:"
      echo "  make test           - Engine unit tests (./pkg/...)"
      echo "  make run-demo       - Run examples/demo on desktop"
      echo "  make build-examples - Build examples to WASM"
      echo "  make serve          - WASM build + local HTTP server"
      echo "  make clean          - Clean build artifacts"
      echo ""
      echo "Sibling games (replace → this repo):"
      echo "  cd ../rpg-game && go run ./game"
      echo "  cd ../energy-tycoon && go run ./game"
      echo ""
      echo "Environment:"
      echo "  Go: $(go version | cut -d' ' -f3-4)"
      echo "  SuperLU: ${pkgs.superlu}/lib/libsuperlu.so (sparse solver; energy-tycoon)"
      echo "  CGO_CFLAGS: $CGO_CFLAGS"
      echo "  CGO_LDFLAGS: $CGO_LDFLAGS"
      echo "  SHELL: $SHELL (oh-my-zsh via ~/.zshrc)"
      echo "=================================================="
      echo ""
    fi

    # Drop into zsh for interactive `nix-shell` only (not `nix-shell --run`,
    # and not direnv's non-interactive eval). Guard avoids exec loops.
    if [[ $- == *i* && -z "$IN_NIX_SHELL_ZSH" ]]; then
      export IN_NIX_SHELL_ZSH=1
      exec "$SHELL"
    fi
  '';
}
