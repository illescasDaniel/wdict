---
id: DOC-022
title: Building and installing wdict
summary: Linux and macOS prerequisites and Go installation, macOS app packaging, and AppImage builds
type: how-to
status: current
updated: 2026-11-20
verified:
covers: [Taskfile.yml, go.mod, frontend/package-lock.json, scripts/build-macos-app.sh, scripts/build-appimage.sh, .github/workflows/ci.yml]
related: []
supersedes:
superseded_by:
---

# Building and installing wdict

Use these instructions to build from source or install a released version with Go. Commands run from the repository root unless stated otherwise.

## Common setup

Install Go 1.26.4 or newer and Node.js 24.15 or newer in the 24.x series. Go modules, project-local npm dependencies and the pinned [Task](https://taskfile.dev/) Go tool provide the development tools; no global Task or Wails install is needed. The first `go tool task` invocation downloads and compiles Task. Run `go tool task --list` to see the commands. `go tool task setup` installs frontend dependencies on both platforms.

## Build and run

On Linux, install a C compiler, `pkg-config`, and GTK 3 and WebKit2GTK 4.1 system development packages. On Ubuntu, CI installs `build-essential`, `pkg-config`, `libgtk-3-dev`, `libwebkit2gtk-4.1-dev` and `libayatana-appindicator3-dev`. On macOS, install Xcode Command Line Tools. `task setup` does not install these system prerequisites. Then, on either platform, run:

```sh
go tool task setup
go tool task build
./wdict
```

The `gtk3` tag is needed on systems with GTK 3/WebKit2GTK 4.1 instead of Wails v3's default GTK 4/WebKitGTK 6. The binary embeds the static frontend; after UI changes, rebuild and commit `frontend/build` alongside the source. CI verifies the generated assets are up to date.

A downloaded AppImage may require FUSE. Alternatively run `./wdict-x86_64.AppImage --appimage-extract-and-run`.

<a id="install-with-go-linux"></a>

## Install with Go

Install Go 1.26.4 or newer and the platform system prerequisites above (Node.js is not required). On Linux, run:

```sh
go install -tags gtk3 github.com/illescasDaniel/wdict@v0.8.3
wdict
```

On macOS, the GTK build tag is unnecessary:

```sh
go install github.com/illescasDaniel/wdict@v0.8.3
wdict
```

To update later, rerun `go install` with the desired newer version (or use `@latest`); it replaces the executable in the same Go binary directory.

Go installs the executable to `GOBIN` (or `GOPATH/bin`, normally `~/go/bin`, if `GOBIN` is unset), which must be on your `PATH`. On Linux, GTK and WebKit libraries must also be installed at runtime. The frontend is embedded; Node.js is not needed for `go install`. This does not create a desktop launcher or macOS app bundle. The v0.8.3 release predates the macOS pronunciation changes on main. The `gtk3` tag selects GTK 3 / WebKit2GTK 4.1.

## Build the macOS app

Install Xcode Command Line Tools and the common Go and Node.js prerequisites, then run:

```sh
go tool task macos-app
```

This builds `dist/wdict.app` and `dist/wdict-macos-<architecture>.zip` for the host architecture. The app is unsigned and not notarized; macOS may block it until you explicitly allow it in Privacy & Security. The build refuses to overwrite existing output. CI uploads an Apple silicon zip as a workflow artifact; it is not a release. For speech synthesis, install eSpeak NG separately (for example, `brew install espeak-ng`) or try a Piper voice; its runtime installation on macOS has not been verified.

## Build a Linux AppImage

On Linux, install the build prerequisites above, plus `pkg-config`, `file`, `find`, librsvg and GObject Introspection development files. Run:

```sh
go tool task appimage
```

The script installs frontend dependencies locally, builds the app and invokes the pinned Wails v3.0.0 generator. It downloads linuxdeploy and writes `dist/wdict-<architecture>.AppImage`; the linuxdeploy download tracks its continuous release, so builds are not fully reproducible. Run the image after making it executable (`chmod +x dist/*.AppImage`). FUSE may be required; alternatively run with `--appimage-extract-and-run`.

Build on an appropriately old supported Linux distribution for wider compatibility: system library versions from the build host affect which other distributions can run the image. The bundled GTK/WebKit stack and its licenses should be reviewed before distribution. Offline dictionary data remains in the user's data directory, outside the AppImage.

The AppImage also bundles the unmodified [eSpeak-NG](https://github.com/espeak-ng/espeak-ng) CLI (the build host's package, plus its data directory, ≈ 20 MB) so the eSpeak-NG playback fallback works on systems without it. eSpeak-NG is GPLv3; it runs as a separate process (an aggregate under GPL-3.0 §5, so the application's MIT license is unaffected) and the image ships the license text and a source offer under `usr/share/licenses/wdict/`. Requires `espeak-ng` on the build machine's PATH; the build fails without it.
