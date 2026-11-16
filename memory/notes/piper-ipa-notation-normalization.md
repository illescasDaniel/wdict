---
name: piper-ipa-notation-normalization
description: Wiktionary IPA uses tie bars, dots, syllabicity marks and precomposed nasal vowels that Piper voices handle badly
type: gotcha
date: 2026-11-14
status: current
supersedes: 
superseded_by: 
source: 
---
Verified facts (2026-11-14): Wiktionary writes the affricates with a tie bar (`d͡ʒ`, `t͡ʃ`, U+0361), syllable dots, syllabicity
marks (`ɫ̩`) and precomposed nasal vowels (`ẽ` U+1EBD is category Ll, so a Unicode-Mn strip alone misses it). In Piper voices the
`.` token (id 10) is a real ~0.15 s segment (isolated schwas then read as "a <word>"), and the tie and `ẽ` are silently dropped
(Enrique -> "ŋrike"). `stripNonPhonemic` now strips dots and Mn marks and maps ẽĩõũã to plain vowels before both Piper and the
mnemonic converter.

**Correction:** stripping marks alone did not fix correct pronunciation; see
[piper-model-input-must-match-phonemizer](piper-model-input-must-match-phonemizer.md). dictionaryapi.dev audio URLs return
Cloudflare HTML to plain clients, so recordings fail and the voice fallback runs.
