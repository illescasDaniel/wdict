---
name: espeak-bundled-in-appimage
description: The AppImage injects the unmodified espeak-ng CLI and its data dir; main.go sets ESPEAK_DATA_PATH
type: decision
date: 2026-11-15
status: current
supersedes: 
superseded_by: 
source: 
---
Implemented in `scripts/build-appimage.sh` (and verified locally: `dist/wdict-x86_64.AppImage` 133 MB, bundled
`usr/bin/espeak-ng` 1.52.0 synthesizes with `ESPEAK_DATA_PATH`): the script copies the CLI, the full data directory, the
vendored `packaging/linux/espeak-ng-COPYING` and a generated source-offer notice into the AppDir, then re-runs linuxdeploy with
`--executable`. The build hard-requires espeak-ng on `PATH`, and `release.yml` installs it. `bundledEspeak()` in `main.go`
sets `EspeakVoices.Binary` and the data path. Release CI once hung on an eSpeak output-limit test because a shell child
retained the subprocess pipes; the runner now bounds pipe waiting and closes stdout on cancellation.
