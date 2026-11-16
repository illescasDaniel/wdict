---
name: sense-grouping-presentation-only
description: Compact senses are grouped in the frontend from an optional Meaning.SenseGroup, not by merging in extraction
type: decision
date: 2026-11-15
status: current
supersedes: 
superseded_by: 
source: 
---
Approved 2026-11-15 (`sdd/compact-senses.html`): adjacent compatible senses share a heading and continuous numbering only
inside the same actual source section and the same entry/source/language/pronunciation/etymology context; content, order and
uncertain boundaries are preserved. `Meaning.SenseGroup` (optional, `internal/core/domain.go`) records source-run provenance,
isolated per offline payload; the frontend groups without mutating stored arrays; legacy caches are not force-refreshed.

**Rejected:** coalescing during extraction: newly cached combined definitions would survive a rollback of the presentation.
