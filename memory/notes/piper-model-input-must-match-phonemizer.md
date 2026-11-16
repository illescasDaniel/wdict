---
name: piper-model-input-must-match-phonemizer
description: Raw [[IPA]] for Piper skips vowel-cluster merging; inspect actual model phones, not just audibility
type: gotcha
date: 2026-11-14
status: current
supersedes: 
superseded_by: 
source: 
---
Verified on Piper 1.8 with en_GB-alba-medium (2026-11-14): raw `[[IPA]]` extends Unicode codepoints and skips the phonemizer's
vowel-cluster merging. The model's own reading of "table" is `t ˈ eɪ b ə l` with one clustered `eɪ` token, so dictionary
`/ˈteɪbl̩/` needs vowel-anchored stress (dictionary syllable-initial stress moved to the vowel), a clustered `eɪ`, and the
syllabic ending kept; removing the syllabicity mark loses the final syllable. Dots and ties are notation, but syllabicity and
nasalization carry sound.

**How to apply:** preserve the dictionary's segment choices, normalize for the engine, reject unsupported model phones through
the existing fallback, and version the synthesis cache so old audio cannot mask a repair. Audibility, duration and inventory
membership are not enough: assert the actual model phone sequence. This corrects [piper-ipa-notation-normalization](piper-ipa-notation-normalization.md).
