---
name: lookup-alternatives-design-choices
description: Alternatives appear only after a miss, need the offline dictionary, and use bounded rules instead of a morphology dataset
type: decision
date: 2026-11-06
status: current
supersedes: 
superseded_by: 
source: 
---
Approved 2026-11-06 (spec: `sdd/lookup-alternatives/SPEC.md`): show alternatives only after a language misses and only when
that language's offline dictionary is installed; include phrase headwords. No morphology dataset: UniMorph and others were
investigated and no suitable distribution was found for the first release (`reports/English and Spanish morphology.md`), so
bounded regular-form suffix rules for English and Spanish plus a bounded Unicode optimal-string-alignment distance (OSA <= 2,
a transposition counts as one edit) with deterministic tie-breaking rank the candidates. Only headwords verified in the
installed index are returned. Nothing derived is persisted. `timed_out` and `unavailable` have distinct UI messages. Shipped
in v0.6.0.

**Not done:** it was implemented with bounded rules, without a morphology dataset.
**How to apply:** ranking ties are alphabetical because there is no word-frequency data (`dictionari` ranks `dictionaric`
ahead of `dictionary`).
