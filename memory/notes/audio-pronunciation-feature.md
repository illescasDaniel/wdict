---
name: audio-pronunciation-feature
description: Audio pronunciation: per-item play buttons; en-us/es default voices; codec failure falls back to eSpeak
type: decision
date: 2026-11-13
status: current
supersedes: 
superseded_by: 
source: 
---
Approved in chat through `sdd/audio-pronunciation.html`: per-item play buttons; a codec failure falls back to eSpeak rather
than plugin messaging; default voices en-us and es; eSpeak-NG bundled when its licence allows. Verify amendments ("Okay let's
go with your recommendations"): translated playback-failed copy (en and es) for a double failure in the webview, hover wording
"source and attribution". `PlayRequest.Headword` carries the retry data, `PlayResult.Voice` upgrades the hover text, and the
eSpeak fallback is capped at one attempt.
