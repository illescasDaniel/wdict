---
name: offline-dictionary-xxyzz-stardict
description: Offline dictionaries are xxyzz Wiktionary StarDict releases (20261018); first lookup uses a direct index scan
type: decision
date: 2026-10-27
status: current
supersedes: 
superseded_by: 
source: 
---
The English offline dictionary is the xxyzz 20261018 English-English Wiktionary StarDict release (published SHA-256 matched
the download; package metadata says CC BY-SA 4.0). Spanish later used the same publisher's 20261018 Spanish release
(README lists both). Facts measured 2026-10-27 on the Linux reference machine: go-stardict's first lookup took ~9.4 s
(8.68 s in the baseline run, missing the spec's 250 ms target); a direct StarDict index scan (`scanDictionary` in
`internal/adapters/offline.go`) cut the first warm-OS-cache lookup to 134-151 ms. Equivalence was checked on
representative queries only, not all entries.

**How to apply:** StarDict data is untrusted input: see [offline-duplicate-results-aliases](offline-duplicate-results-aliases.md).
