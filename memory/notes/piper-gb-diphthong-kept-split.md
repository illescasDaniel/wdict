---
name: piper-gb-diphthong-kept-split
description: British əʊ is kept as separate ə and ʊ for en_GB-alba; only the US voice uses clustered oʊ
type: decision
date: 2026-11-15
status: current
supersedes: piper-schwa-upsilon-fold-to-ou
superseded_by: 
source: 
---
Verified 2026-11-15 by the implementer and an independent reviewer: en_GB-alba phonemizes "though" with separate `ə`, `ʊ`
phones, while en_US-amy uses clustered `oʊ`. The unconditional `əʊ` -> `oʊ` fold was wrong for the GB model and was removed;
both displayed variants are preserved, onset-schwa folding excludes `əʊ`, and the synthesis cache version was bumped. Regression
cases cover UK/US though, slow, oak, old and appearance; the user confirmed it by listening. This replaces
[piper-schwa-upsilon-fold-to-ou](piper-schwa-upsilon-fold-to-ou.md).

**How to apply:** do not infer Piper tokenization from the standalone eSpeak mnemonics; check the installed voice.
Open: one voice per language, so US IPA can still be spoken by alba.
