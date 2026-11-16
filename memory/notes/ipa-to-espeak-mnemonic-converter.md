---
name: ipa-to-espeak-mnemonic-converter
description: The voice path speaks the shown IPA through a pure IPAToMnemonics converter, oracle-checked against espeak-ng
type: decision
date: 2026-11-13
status: current
supersedes: 
superseded_by: 
source: 
---
The user said the exact transcription matters, so revision 3 of `sdd/audio-pronunciation.html` converts IPA to eSpeak
mnemonics (`IPAToMnemonics`, tables in `internal/core/playback.go`, 94/94 entries oracle-verified with `espeak-ng -q -x`,
espeak-ng 1.52.0). Unmappable symbols fall back to the plain headword. Facts: es ɣ -> `Q` (`[[G]]` is invalid in es); bare
`[[A]]` is invalid in en (ɑ needs the ɑː -> `A:` pair); every vowel mnemonic accepts `:` lengthening; `I@` echoes `I;@`;
isolated voiced consonants synthesize silence but are audible with a vowel, so audibility tests use full streams; mnemonic
input `[[...]]` is audible without `-m` in raw text mode.
