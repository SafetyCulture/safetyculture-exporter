#!/usr/bin/env bash
# Pack a Mitti Exporter CLI binary into a .mcpb (MCP Bundle) for one-click
# install into Claude Desktop.
#
# Usage: scripts/pack-mcpb.sh <platform> <binary-path> <version>
#   platform:    darwin | win32 | linux
#   binary-path: path to the built mitti-exporter binary (signed on macOS!)
#   version:     e.g. 3.2.0 (no leading v)
#
# Output: mitti-exporter-<platform>.mcpb in the current directory.
set -euo pipefail

PLATFORM="${1:?platform required: darwin|win32|linux}"
BINARY="${2:?path to mitti-exporter binary required}"
VERSION="${3:?version required, e.g. 3.2.0}"
VERSION="${VERSION#v}"

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WORK_DIR="$(mktemp -d)"
trap 'rm -rf "$WORK_DIR"' EXIT

BINARY_NAME="mitti-exporter"
if [ "$PLATFORM" = "win32" ]; then
  BINARY_NAME="mitti-exporter.exe"
fi

mkdir -p "$WORK_DIR/server"
cp "$BINARY" "$WORK_DIR/server/$BINARY_NAME"
chmod +x "$WORK_DIR/server/$BINARY_NAME"
cp "$REPO_ROOT/ui/build/appicon.png" "$WORK_DIR/icon.png"

# Render the manifest template.
sed -e "s/\${VERSION}/$VERSION/g" \
    -e "s/\${PLATFORM}/$PLATFORM/g" \
    -e "s/\${BINARY_NAME}/$BINARY_NAME/g" \
    "$REPO_ROOT/mcpb/manifest.json" > "$WORK_DIR/manifest.json"

# Validate + pack (uses npx; Node 20+ required).
npx -y @anthropic-ai/mcpb validate "$WORK_DIR/manifest.json"
npx -y @anthropic-ai/mcpb pack "$WORK_DIR" "./mitti-exporter-${PLATFORM}.mcpb"

echo "Packed: mitti-exporter-${PLATFORM}.mcpb"
