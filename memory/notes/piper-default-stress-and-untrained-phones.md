---
name: piper-default-stress-and-untrained-phones
description: Inject default primary stress before each word's first vowel and fold or reject phones the voices were not trained on
type: decision
date: 2026-11-14
status: current
supersedes: 
superseded_by: 
source: 
---
Found with "eight": `/eɪt/` lost its initial sound and North India `/eʈ/` said nothing. eSpeak always stresses words and the
voices were trained on its output, while Wiktionary omits stress on monosyllables, so unstressed feeds are out of
distribution. `NormalizePiperIPA` injects default primary stress per word when a word has none. The voice maps list every IPA
symbol but the voices are untrained on exotic ones (unstressed `eʈ` gave noise, peaks 0.009-0.031 against ~0.2 for speech), so
the retroflex series (ʈ ɖ ʂ ʐ ɳ ɭ ɽ) folds to t d s z n l ɾ and the `piperSpeakable` gate rejects other phones (headword
fallback). Umlaut vowels fold to base; the liaison tie and affix hyphens are dropped; multi-word IPA keeps single spaces with
per-word stress (15,205 es spans have spaces).

A full-corpus sweep of all 182,639 unique IPA spans: 98.1% en and 98.0% es speakable; rejections are genuinely non-engine
phones (ʉ ɨ ø ʁ ç ɬ, tone letters, unbalanced brackets, ~190 dangling stress affixes) and correctly fall back. The `əʊ` -> `oʊ`
fold that went in with this change was later found wrong for British voices: see
[piper-gb-diphthong-kept-split](piper-gb-diphthong-kept-split.md).
