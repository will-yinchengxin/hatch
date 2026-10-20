#!/usr/bin/env bash
# release.sh — build, sign, and package Hatch for macOS (Apple Silicon)
# Usage: ./scripts/release.sh [version]
# Example: ./scripts/release.sh 0.2.0
set -euo pipefail

# Always run from project root regardless of where the script is called from
cd "$(dirname "$0")/.."

VERSION="${1:-0.1.0}"
WAILS="${GOPATH:-$HOME/go}/bin/wails"
DMG_OUT="build/bin/Hatch-${VERSION}-arm64.dmg"

echo "▶ Building Hatch ${VERSION} (darwin/arm64)…"
# Backend (Go module + wails.json) lives in src/; output goes to ../build/bin
(cd src && "$WAILS" build -platform darwin/arm64 -clean)

echo "▶ Signing with ad-hoc identity…"
codesign --force --deep --sign - build/bin/Hatch.app
codesign --verify --deep --strict build/bin/Hatch.app
echo "  codesign OK"

# Remove previous DMG if it exists
[ -f "$DMG_OUT" ] && rm "$DMG_OUT"

echo "▶ Creating DMG: ${DMG_OUT}…"
/opt/homebrew/bin/create-dmg \
  --volname "Hatch" \
  --volicon "build/appicon.png" \
  --window-pos 200 120 \
  --window-size 600 400 \
  --icon-size 100 \
  --icon "Hatch.app" 160 185 \
  --hide-extension "Hatch.app" \
  --app-drop-link 440 185 \
  "$DMG_OUT" \
  "build/bin/Hatch.app"

echo ""
echo "✓ Done: $(ls -lh "$DMG_OUT" | awk '{print $5, $NF}')"
echo ""
echo "To open on first launch (bypass Gatekeeper):"
echo "  xattr -cr build/bin/Hatch.app && open build/bin/Hatch.app"
echo "  # or double-click $DMG_OUT and drag to /Applications"
