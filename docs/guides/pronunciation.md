---
id: DOC-024
title: Pronunciation and voices
summary: eSpeak NG and Piper setup, playback fallbacks, requirements and third-party voice licenses
type: how-to
status: current
updated: 2026-11-20
verified:
covers: [internal/adapters/audio_player.go, internal/adapters/piper.go, internal/adapters/piper_install.go, internal/adapters/voices_catalog.go, internal/adapters/espeak.go, scripts/build-appimage.sh]
related: []
supersedes:
superseded_by:
---

# Pronunciation and voices

Pronunciation playback can optionally use natural neural voices ([Piper](https://github.com/OHF-Voice/piper1-gpl)): when a voice pack is installed, play buttons read the shown IPA transcription with a natural voice instead of eSpeak-NG; without one, playback is unchanged and eSpeak-NG remains the automatic fallback.

Synthesised audio is played natively using host players found on `PATH`. On macOS, the app tries the built-in `afplay` first. The other candidates are `pw-play`, `paplay`, `aplay` and `ffplay`, in that order. PipeWire and PulseAudio normally provide `pw-play` and `paplay`, respectively. Without a native player, playback falls back to the webview's audio element. The AppImage ships without the GStreamer plugins that WebKit needs for audio playback.

In the app's download screen, first install the **Piper runtime** once (≈ 210 MB installed; it runs in its own Python environment inside the app data directory and installs the pinned Piper version using pip; its dependencies are recorded in a manifest), then one voice per language (≈ 63–77 MB each, checksummed at the pinned Hugging Face revision). This requires `python3` (≥ 3.9) on the PATH; on Debian/Ubuntu also install `python3-venv`, or the runtime install cannot complete. Piper is third-party GPLv3 code ([COPYING](https://github.com/OHF-Voice/piper1-gpl/blob/main/COPYING)), attributed to **rhasspy / OHF-Voice piper voices**; each voice's dataset and licence are linked from its model card in the app ([en_GB-alba-medium](https://huggingface.co/rhasspy/piper-voices/blob/6b9cfaacb4fb2e02a9621227e3c1a48bec9710c3/en/en_GB/alba/medium/MODEL_CARD), [en_US-amy-medium](https://huggingface.co/rhasspy/piper-voices/blob/6b9cfaacb4fb2e02a9621227e3c1a48bec9710c3/en/en_US/amy/medium/MODEL_CARD), [es_ES-sharvard-medium](https://huggingface.co/rhasspy/piper-voices/blob/6b9cfaacb4fb2e02a9621227e3c1a48bec9710c3/es/es_ES/sharvard/medium/MODEL_CARD)).
