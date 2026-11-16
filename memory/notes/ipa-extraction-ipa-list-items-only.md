---
name: ipa-extraction-ipa-list-items-only
description: Offline English IPA comes only from <li> items whose own text starts with 'IPA:', skipping <ol> definitions and Rhymes
type: decision
date: 2026-11-11
status: current
supersedes: 
superseded_by: 
source: 
---
Rejected approach: any `IPA`-class span, which also took rhyme and definition IPA (for example `diktat`). Real fixture:
`internal/adapters/testdata/pronunciation/offline_en_vegetable.html`. Old cached text labels are remapped to flags on
`DiskCache.Get` without a cache version bump; `MemoryCache` does not normalize. Verified by tests and Chromium QA with a stub
runtime, not in the real window.
