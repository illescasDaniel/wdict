---
name: audio-native-host-player
description: Pronunciations play through a host audio player (pw-play, paplay, aplay, ffplay), not the webview audio element; Wiktionary MP3 download removed
type: decision
date: 2026-11-15
status: current
supersedes: 
superseded_by: 
source: 
---
Since v0.8.3 pronunciation playback goes through a native host player, tried in order `pw-play`, `paplay`, `aplay`, `ffplay`,
with a webview data-URI fallback when none is found (`internal/adapters/audio_player.go`). The download of Wiktionary MP3
files was removed in the same release.

**Why:** the AppImage lacks the GStreamer `appsink` and `autoaudiosink` elements, so the webview `<audio>` element froze
WebKit there. The user confirmed by listening that the AppImage works with the host player.

**How to apply:** do not move playback back into the webview to simplify the code; test audio changes in the AppImage, not
only in a development build.
