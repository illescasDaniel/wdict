<p align="center"><img src="packaging/linux/wdict.svg" alt="wdict book icon" width="96"></p>

# wdict

[![CI](https://github.com/illescasDaniel/wdict/actions/workflows/ci.yml/badge.svg)](https://github.com/illescasDaniel/wdict/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-green)](LICENSE)

A desktop dictionary built with Go, Wails v3 and SvelteKit. Development version: **0.8.4**. Search English and Spanish words and phrases online, reuse saved results, or download Wiktionary dictionaries in the app for offline use. Queries are sent to the online provider only after an offline and cache miss or an expired cache entry. Online results stay fresh for 365 days; when refresh fails, a stale saved result can still be shown.

## How it works

| Step | What happens |
| --- | --- |
| Search | A query is checked against installed English and Spanish offline dictionaries. |
| Reuse | A saved online result is used when available and fresh. |
| Look up | On a miss or expired result, the appropriate online provider is queried; a stale result can be shown if refresh fails. |
| Listen | Optional pronunciation uses eSpeak NG or an installed Piper voice and a host audio player. |

## Quick start

On Linux, download `wdict-x86_64.AppImage` from the [latest release](https://github.com/illescasDaniel/wdict/releases/latest), then run:

```sh
chmod +x wdict-x86_64.AppImage
./wdict-x86_64.AppImage
```

On macOS, download the unsigned Apple silicon ZIP artifact from a successful [CI run](https://github.com/illescasDaniel/wdict/actions/workflows/ci.yml), or [build the app](#build-the-macos-app).

<a id="status"></a>

## Requirements and limits

Linux and macOS builds and automated tests are supported. Native macOS GUI and audible playback still need manual verification; Windows is untested. The release provides a Linux AppImage; macOS CI artifacts are unsigned and not notarized. Offline dictionaries and optional voices are separate downloads.

## Build from source

Install Go 1.26.4 or newer and Node.js 24.15 or newer in the 24.x series, plus the platform prerequisites in the [build guide](docs/guides/building.md). From the repository root, install frontend dependencies:

```sh
go tool task setup
```

### Build and run on Linux

After installing GTK 3 and WebKit2GTK 4.1 development dependencies:

```sh
go tool task build
./wdict
```

### Build the macOS app

With Xcode Command Line Tools installed:

```sh
go tool task macos-app
```

This creates an unsigned `dist/wdict.app` and ZIP for the host architecture. See the [build guide](docs/guides/building.md) for prerequisites, output and macOS security details.

### Build a Linux AppImage

With the additional packaging prerequisites in the [build guide](docs/guides/building.md):

```sh
go tool task appimage
```

### Install with Go (Linux)

See the [Go installation instructions](docs/guides/building.md#install-with-go-linux) for the released version and runtime prerequisites.

## Tests

On either platform, after installing the build prerequisites and running setup:

```sh
go tool task test
```

This runs frontend checks and tests, followed by Go tests. See the [testing guide](docs/guides/testing.md) for optional browser and integration checks.

## Evaluation

[CI](https://github.com/illescasDaniel/wdict/actions/workflows/ci.yml) builds and tests on Linux and macOS. Historical measurements and their limits are in the [documentation index](docs/INDEX.md); they are not a current app-wide quality score.

## Documentation

- [Building and installation](docs/guides/building.md)
- [Testing](docs/guides/testing.md)
- [Pronunciation and voices](docs/guides/pronunciation.md)
- [Dictionary data and sources](docs/guides/data-and-sources.md)
- [Preparing a release](docs/guides/releasing.md)
- [Documentation index](docs/INDEX.md)

<a id="install-on-linux"></a><a id="natural-voices"></a><a id="preparing-a-release"></a><a id="data-and-sources"></a>

## Author and license

Created by Daniel Illescas Romero ([illescas.daniel@protonmail.com](mailto:illescas.daniel@protonmail.com)). The project source is licensed under the [MIT License](LICENSE). Dictionary content and bundled third-party components have their own licenses.
