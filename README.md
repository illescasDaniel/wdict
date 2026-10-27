# wdict

A Linux-first desktop dictionary built with Go, Wails v3 and SvelteKit. Search English words and phrases online, reuse saved results, or download a Wiktionary dictionary in the app for offline use. Queries are sent to the online provider only after an offline and cache miss or an expired cache entry.

## Build and run on Linux

Install Go 1.26.4 or newer, Node.js compatible with Vite 7, and the development packages for GTK 3 and WebKit2GTK 4.1. These are system build prerequisites; the project uses Go modules and project-local npm dependencies, not globally installed Wails or JavaScript tools.

The project pins [Task](https://taskfile.dev/) as a Go tool, so no global Task installation is required. From the repository root, run:

```sh
go tool task setup
go tool task build
./wdict
```

The first `go tool task` invocation downloads and compiles the pinned tool. Run `go tool task --list` to see all available commands.

The `gtk3` tag is needed on systems with GTK 3/WebKit2GTK 4.1 instead of Wails v3's default GTK 4/WebKitGTK 6. The binary embeds the static frontend, so rebuild the frontend before rebuilding Go after UI changes. macOS and Windows builds have not yet been verified.

For checks:

```sh
go tool task test
```

## Linux AppImage

On Linux, install the build prerequisites above, plus `pkg-config`, `file`, `find`, librsvg and GObject Introspection development files. Run:

```sh
go tool task appimage
```

The script installs frontend dependencies locally, builds the app and invokes the pinned Wails v3.0.0 generator. It downloads linuxdeploy and writes `dist/wdict-<architecture>.AppImage`; the linuxdeploy download tracks its continuous release, so builds are not fully reproducible. Run the image after making it executable (`chmod +x dist/*.AppImage`). FUSE may be required; alternatively run with `--appimage-extract-and-run`.

Build on an appropriately old supported Linux distribution for wider compatibility: system library versions from the build host affect which other distributions can run the image. The bundled GTK/WebKit stack and its licenses should be reviewed before distribution. Offline dictionary data remains in the user's data directory, outside the AppImage.

## Data and sources

The app offers the pinned [20261018 English Wiktionary StarDict release](https://github.com/xxyzz/wiktionary_stardict/releases/tag/20261018), compiled by [xxyzz](https://xxyzz.github.io/wiktionary_stardict/) from [Wiktionary](https://en.wiktionary.org/). Its package metadata identifies the snapshot as [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/); credit belongs to Wiktionary contributors and xxyzz. wdict extracts definitions and changes their formatting. The download is approximately 142 MB, is user initiated, and is verified against the publisher's SHA-256 checksum before installation.

Online lookups use [Free Dictionary API](https://dictionaryapi.dev/). The app displays each response's supplied source and license links; sampled responses identified [CC BY-SA 3.0](https://creativecommons.org/licenses/by-sa/3.0/). Offline results link to their Wiktionary source and license, and the offline dictionary screen also links to its source and license. Wiktionary may contain material from other sources with separate requirements; see [Wikimedia's reuse terms](https://foundation.wikimedia.org/wiki/Policy:Terms_of_Use).

Offline dictionary data is stored separately from the disposable response cache, in the user's data directory. Inflected-form resolution and spelling suggestions are planned for later.
