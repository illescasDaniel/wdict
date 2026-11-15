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
go build -tags gtk3 -ldflags "-X main.version=$(cat "$root/VERSION")" -o "$builddir/wdict" .
go run -tags gtk3 github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0 generate appimage \
  -binary "$builddir/wdict" -icon "$root/packaging/linux/wdict.png" \
  -desktopfile "$root/packaging/linux/wdict.desktop" -builddir "$builddir/bundle" \
  -outputdir "$root/dist"

# Bundle the unmodified eSpeak-NG CLI into the AppDir (decision 2026-11-13,
# sdd/audio-pronunciation.html G4): the app calls the standalone binary as a
# separate process — GPLv3 §5 aggregation, no libespeak-ng linkage — so the
# eSpeak playback fallback also works on distros without the system package.
# The license text and a source offer ship alongside, and the app points the
# bundled CLI at the bundled data directory via ESPEAK_DATA_PATH (main.go).
case $(uname -m) in
  x86_64) arch=x86_64 ;;
  aarch64) arch=aarch64 ;;
  *) echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac
espeak_bin=$(command -v espeak-ng) || {
  echo 'espeak-ng not found on PATH; install it to bundle the eSpeak fallback.' >&2
  exit 1
}
espeak_data=$(espeak-ng --version | sed -n 's/.*Data at: //p')
if [[ -z $espeak_data || ! -d $espeak_data ]]; then
  echo 'cannot locate the eSpeak-NG data directory' >&2
  exit 1
fi
espeak_version=$(espeak-ng --version | sed -n 's/eSpeak NG text-to-speech: //;s/  Data at:.*//p')
appdir="$builddir/bundle/wdict-$arch.AppDir"
install -m 755 "$espeak_bin" "$appdir/usr/bin/espeak-ng"
mkdir -p "$appdir/usr/share/espeak-ng-data" "$appdir/usr/share/licenses/wdict"
cp -a "$espeak_data/." "$appdir/usr/share/espeak-ng-data/"
cp "$root/packaging/linux/espeak-ng-COPYING" "$appdir/usr/share/licenses/wdict/espeak-ng-COPYING"
cat >"$appdir/usr/share/licenses/wdict/espeak-ng-BUNDLING.txt" <<EOF
Bundled eSpeak-NG
=================

This AppImage bundles the unmodified eSpeak-NG command-line program
(espeak-ng $espeak_version, https://github.com/espeak-ng/espeak-ng)
with its data directory. eSpeak-NG is licensed under the GNU General
Public License Version 3 (espeak-ng-COPYING in this directory); the
program runs as a separate process from the wdict application, so this
bundle is an aggregate and the application's MIT license is unaffected.
Corresponding source: https://github.com/espeak-ng/espeak-ng (tag/branch
for the version above).
EOF
linuxdeploy="$builddir/bundle/linuxdeploy-$arch.AppImage"
# the generator disables stripping when system libraries carry .relr.dyn
# sections, which its bundled strip binary cannot handle; mirror that check
# so the second pass cannot corrupt already-deployed libraries
if readelf -S /usr/lib/libgtk-3.so.0 /usr/lib64/libgtk-3.so.0 \
     /usr/lib/x86_64-linux-gnu/libgtk-3.so.0 2>/dev/null | grep -q relr; then
  export NO_STRIP=1
fi
(
  cd "$builddir/bundle"
  "$linuxdeploy" --appimage-extract-and-run --appdir "$appdir" \
    --executable "$appdir/usr/bin/espeak-ng"
  # linuxdeploy leaves a few host libraries un-deployed (e.g. libasound,
  # libxcb); copy the bundled CLI's remaining non-base dependencies so the
  # eSpeak fallback never depends on the host's audio or X stack
  for lib in $(ldd "$appdir/usr/bin/espeak-ng" | awk '$3 ~ /^\// {print $3}'); do
    case $lib in
      "$appdir"/*|*/ld-linux*|*/libc.so*|*/libm.so*|*/libmvec*|\
      */libstdc++*|*/libgcc_s*) continue ;;
    esac
    cp -L "$lib" "$appdir/usr/lib/"
  done
  OUTPUT="wdict-$arch.AppImage" \
    "$linuxdeploy" --appimage-extract-and-run --appdir "$appdir" --output appimage
)
mv "$builddir/bundle/wdict-$arch.AppImage" "$root/dist/wdict-$arch.AppImage"
