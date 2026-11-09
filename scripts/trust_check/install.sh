#!/bin/sh
# Build tc (trust_check) and deploy a REAL (non-symlink) binary to ~/.local/bin.
# Usage:
#   ./install.sh                          # system go
#   GO=~/sdk/go1.27.3/bin/go ./install.sh # pinned Go when system Go lags
#   TC_INSTALL_DIR=/somewhere ./install.sh
set -e
cd "$(dirname "$0")"

GO_BIN="${GO:-go}"
"$GO_BIN" version >/dev/null 2>&1 || {
  echo "go not found in PATH. Try: GO=~/sdk/go1.27.3/bin/go ./install.sh" >&2
  exit 2
}

# stripped static binary; bin/tc and ~/.local/bin/tc stay hash-identical
"$GO_BIN" build -ldflags "-s -w" -o bin/tc .

TARGET="${TC_INSTALL_DIR:-$HOME/.local/bin}"
mkdir -p "$TARGET"
cp -f bin/tc "$TARGET/tc"
echo "installed: $TARGET/tc (static, stripped, $(du -h "$TARGET/tc" | cut -f1))"
