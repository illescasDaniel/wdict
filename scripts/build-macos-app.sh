#!/usr/bin/env bash
set -euo pipefail

root=$(cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"
if [[ $(uname -s) != Darwin ]]; then
  echo 'macOS app builds require macOS.' >&2
  exit 1
fi
version=$(tr -d '\n' < VERSION)
if [[ ! $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo 'VERSION must be a numeric semantic version.' >&2
  exit 1
fi
arch=$(uname -m)
case "$arch" in arm64) goarch=arm64 ;; x86_64) goarch=amd64 ;; *) echo "Unsupported architecture: $arch" >&2; exit 1 ;; esac
mkdir -p dist
lock="$root/dist/.wdict-macos-build.lock"
if ! mkdir "$lock" 2>/dev/null; then
  echo "Another packaging build is running, or lock exists: $lock" >&2
  exit 1
fi
builddir=''
published_app=0
cleanup() {
  if [[ $published_app == 1 ]]; then rm -rf -- "$app"; fi
  if [[ -n $builddir ]]; then rm -rf -- "$builddir"; fi
  rmdir "$lock"
}
trap cleanup EXIT
app="$root/dist/wdict.app"
archive="$root/dist/wdict-macos-$arch.zip"
if [[ -e $app || -L $app || -e $archive || -L $archive ]]; then
  echo "Output already exists: move or remove $app and $archive before rebuilding." >&2
  exit 1
fi
builddir=$(mktemp -d "$root/dist/.wdict-macos.XXXXXXXX")
bundle="$builddir/wdict.app"
mkdir -p "$bundle/Contents/MacOS" "$bundle/Contents/Resources" "$builddir/wdict.iconset"
npm ci --prefix frontend
npm run build --prefix frontend
GOOS=darwin GOARCH="$goarch" CGO_ENABLED=1 go build -ldflags "-X main.version=$version" -o "$bundle/Contents/MacOS/wdict-bin" .
cat > "$bundle/Contents/MacOS/wdict" <<'LAUNCHER'
#!/bin/sh
# Finder starts apps with a minimal PATH; include standard Homebrew locations
# for optional speech and Python tools, preserving inherited entries.
export PATH="${PATH:-/usr/bin:/bin:/usr/sbin:/sbin}:/opt/homebrew/bin:/usr/local/bin"
exec "$(dirname "$0")/wdict-bin" "$@"
LAUNCHER
chmod 755 "$bundle/Contents/MacOS/wdict"
sed "s/VERSION_PLACEHOLDER/$version/g" packaging/macos/Info.plist > "$bundle/Contents/Info.plist"
for size in 16 32 128 256 512; do
  sips -z "$size" "$size" packaging/linux/wdict.png --out "$builddir/wdict.iconset/icon_${size}x${size}.png" >/dev/null
  double=$((size * 2))
  sips -z "$double" "$double" packaging/linux/wdict.png --out "$builddir/wdict.iconset/icon_${size}x${size}@2x.png" >/dev/null
done
iconutil -c icns "$builddir/wdict.iconset" -o "$bundle/Contents/Resources/wdict.icns"
plutil -lint "$bundle/Contents/Info.plist"
ditto -c -k --sequesterRsrc --keepParent "$bundle" "$builddir/wdict-macos-$arch.zip"
mv "$bundle" "$app"
published_app=1
mv "$builddir/wdict-macos-$arch.zip" "$archive"
published_app=0
printf 'Created %s and %s (unsigned)\n' "$app" "$archive"
