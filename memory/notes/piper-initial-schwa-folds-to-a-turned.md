---
name: piper-initial-schwa-folds-to-a-turned
description: English word-initial schwa was near-silent in Piper; fold it to ɐ. Mid-word ə stays
type: incident
date: 2026-11-14
status: current
supersedes: 
superseded_by: 
source: 
---
"appearance" lost its start. The two pinned en voices are trained on eSpeak text-mode output (config `phoneme_type: ESPEAK`),
whose dictionaries write the word-initial unstressed prefix vowel as `ɐ` (about -> `ɐbˈaʊt`; `occasion` -> `əkˈeɪʒən` is an
exception). `əpˈiɹəns` against `pˈiɹəns` over three trials each: lead-in 0.29 s against 0.20 s, and the ə variant contained
the vowel-less audio shifted 82 ms (normalized cross-correlation 0.64), so the schwa added only silence. `NormalizePiperIPA`
folds a word-initial English ə to ɐ (id 50 in alba); ɐ and ə are perceptually the same unstressed central vowel. The
superscript nasal release ⁿ (U+207F, Lm not Mn) folds to n unless a nasal already follows.
