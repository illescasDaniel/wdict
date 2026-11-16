---
name: alternatives-heap-target-not-hard-cap
description: The 64 MiB heap figure for lookup alternatives is a measured shared-heap target, not a per-feature cap
type: decision
date: 2026-11-06
status: current
supersedes: 
superseded_by: 
source: 
---
The user approved treating 64 MiB as a measured shared-heap target for the whole app rather than a strict cap for the
alternatives feature.

**How to apply:** do not fail or redesign the feature because it alone nears 64 MiB; measure the shared heap instead.
