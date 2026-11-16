---
name: offline-exact-headword-ranking
description: Offline results are ranked in tiers: exact raw headword, then folded match, then alias; index order inside a tier
type: decision
date: 2026-11-15
status: current
supersedes: 
superseded_by: 
source: 
---
Forensics on 2026-11-15 showed that `scanDictionary` returned alias-redirected and folded matches in raw index order with no
preference for the exact headword, so the correct card for `año` ([ˈaɲo]) sorted after `-ano`/`ano`/`anó` (`foldWord` strips
the ñ tilde) and `thought` sorted after `think`'s three alias cards. Approved in chat and implemented: tiered ranking (exact
raw headword, case-insensitive, first) with a stable sort (see `internal/adapters/offline.go`). Case-aware ranking, source-card
grouping and an explicit "no IPA" state (`Car` before or after `car`) followed in `sdd/result-ranking.html` and was committed in `a2cbfcc`
(before v0.8.3).

**How to apply:** `OfflineDictionary.Find` caches by `foldWord(query)`, so any ranking change must also handle cache replay.
