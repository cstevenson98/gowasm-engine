#!/usr/bin/env bash
#
# serve-docs.sh - Serve the engine's Go API documentation locally in a browser.
#
# This starts a documentation web server so you can browse every package's
# docs - including the package-level concept overviews added in each package's
# doc.go - as a linked, searchable site.
#
# It picks the first available backend, preferring anything already installed so
# repeated runs are instant:
#   1. an installed `pkgsite` binary  (the pkg.go.dev engine),
#   2. an installed `godoc` binary,
#   3. `go run golang.org/x/tools/cmd/godoc@latest`  (builds against your local
#      Go toolchain - no toolchain upgrade required).
#
# Usage:
#   ./scripts/serve-docs.sh [port]
#
# Environment:
#   DOCS_PORT   Port to listen on (default: 6060). Overridden by the CLI arg.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODULE_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${MODULE_ROOT}"

PORT="${1:-${DOCS_PORT:-6060}}"
MODULE_PATH="$(go list -m)"

serve_pkgsite() {
    echo "Using pkgsite."
    echo "  -> http://localhost:${PORT}/${MODULE_PATH}"
    echo "Press Ctrl+C to stop."
    echo
    exec "$1" -http=":${PORT}" .
}

serve_godoc() {
    echo "Using godoc."
    echo "  -> http://localhost:${PORT}/pkg/${MODULE_PATH}/"
    echo "Press Ctrl+C to stop."
    echo
    exec "$@" -http=":${PORT}"
}

echo "Serving Go documentation for ${MODULE_PATH}"
echo

if command -v pkgsite >/dev/null 2>&1; then
    serve_pkgsite "$(command -v pkgsite)"
elif command -v godoc >/dev/null 2>&1; then
    serve_godoc "$(command -v godoc)"
else
    echo "No docs binary found on PATH; running godoc via 'go run'"
    echo "(first run downloads it; uses your local Go toolchain)..."
    echo
    serve_godoc go run golang.org/x/tools/cmd/godoc@latest
fi
