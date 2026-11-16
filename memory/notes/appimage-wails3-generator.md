---
name: appimage-wails3-generator
description: AppImage is built with the pinned Wails v3.0.0 generator plus scripts/build-appimage.sh; linuxdeploy is not pinned
type: decision
date: 2026-10-27
status: current
supersedes: 
superseded_by: 
source: 
---
The Linux AppImage uses the pinned Wails v3.0.0 AppImage generator for GTK 3 because it bundles and relocates the WebKit
helpers. Its linuxdeploy "continuous" download is not pinned. Facts verified in the module source (2026-11-15): the AppDir
is `$builddir/bundle/wdict-<arch>.AppDir` and survives the run, `AppRun` prepends `$APPDIR/usr/bin` to `PATH`, and
linuxdeploy keeps a custom `AppRun`.

**How to apply:** build on an old supported distribution for wider compatibility. Cross-distribution testing and a
licence review of bundled GTK/WebKit are still open: see [NOW](../NOW.md).
