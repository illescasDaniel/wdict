---
name: offline-duplicate-results-aliases
description: Repeated .syn aliases made scanDictionary return each entry several times (dogfood); fixed in v0.7.1
type: incident
date: 2026-11-13
status: current
supersedes: 
superseded_by: 
source: 
---
Searching `dogfood` showed identical duplicate cards. Root cause (verified 2026-11-13 against the installed xxyzz 20261018
en-en dictionary): the `.syn` file lists aliases repeatedly (`dogfood` twice) and one alias folds to the target itself
(`dog-food`), while `scanDictionary` (`internal/adapters/offline.go`) appended the index word once per matching alias, so
`Find("dogfood")` returned 6 entries in 3 identical pairs. Fix: each index entry is appended at most once
(`matches(word) || aliases[i] > 0`); alias targets still resolve to their original headword. Regression test:
`TestGivenRepeatedSynonymAliasesWhenScannedThenEachIndexEntryIsReturnedOnce`.

**How to apply:** StarDict data is untrusted. `.idx` legitimately repeats ~45k headwords with distinct offsets (separate
articles), so deduplicate per index position, never per headword.
