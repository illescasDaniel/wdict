#!/usr/bin/env bash
set -euo pipefail

root=$(cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"
if [[ $(uname -s) != Linux ]]; then
  echo 'AppImage builds require Linux.' >&2
  exit 1
fi
mkdir -p dist
builddir=$(mktemp -d "${TMPDIR:-/tmp}/wdict-appimage.XXXXXXXX")
trap 'rm -rf -- "$builddir"' EXIT
npm ci --prefix frontend
npm run build --prefix frontend
go build -tags gtk3 -o "$builddir/wdict" .
go run -tags gtk3 github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0 generate appimage \
  -binary "$builddir/wdict" -icon "$root/packaging/linux/wdict.png" \
  -desktopfile "$root/packaging/linux/wdict.desktop" -builddir "$builddir/bundle" \
  -outputdir "$root/dist"
