---
name: ipa-flag-labels-first-card-only
description: Pronunciation line shows flag emoji for accents, only on the first card of a word, label limit applied after mapping
type: decision
date: 2026-11-11
status: current
supersedes: 
superseded_by: 
source: 
---
Approved 2026-11-11 (`sdd/` pronunciation page): accent labels become flags (General American/US 🇺🇸, Received
Pronunciation/UK 🇬🇧, Australia 🇦🇺, New Zealand 🇳🇿, Canada 🇨🇦, Ireland 🇮🇪, South Africa 🇿🇦); compound labels split at commas and
an unknown part keeps the whole label as written. The original goes in `Pronunciation.LabelTitle` (hover text, accessible
name). The line appears only on the first card of a headword; a later card with the same headword (case-insensitive),
language and identical list shows none. Spec R3: the 32-rune limit applies after mapping, the original may be up to 128.
Revision R4: the qualifier words "dialectal" and "obsolete" are ignored, and a label mapping to both 🇬🇧 and 🇺🇸 is the generic
form and shows no label.

**Rejected:** "US"/"UK" text; repeating the line on every card.
