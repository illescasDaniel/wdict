---
name: alternatives-distance-cpu-hotspot
description: Suggestion search was CPU-bound in alternativeDistance and hit its 2 s deadline; a banded OSA fixed it
type: incident
date: 2026-11-06
status: current
supersedes: 
superseded_by: 
source: 
---
Probe and CPU profile on 2026-11-06: on the installed English dictionary `dictionari` and `dictionary` hit the 2 s deadline,
`wrods` took 1.46 s and `look foward to` 1.62 s. A raw index pass costs 118 ms; ~90% of the time was a full dynamic-programming
distance on every length-matched headword. Fix (same day): `alternativeDistance` (`internal/core/alternatives_service.go`)
uses a banded OSA with early exit and input-sized rows, with the approved OSA <= 2 unchanged. Real-dictionary timings fell
to 0.40-0.69 s. `internal/core/alternatives_distance_test.go` checks it against a reference OSA for all strings up to length 4.
