---
name: picomatch-advisory-blocks-0-8-4
description: 0.8.4 release held because npm audit fails on picomatch with no patched version
type: incident
date: 2026-11-18
status: current
supersedes: 
superseded_by: 
source: 
---
CI `npm audit` fails (high, GHSA-fc5q-39xc-vjpf, picomatch <= 4.0.7 ReDoS, dev tooling only) and no patched version existed on
2026-11-18 (4.0.7 was latest). The user chose to wait for a patch rather than add an audit exception or raise the threshold.
Version 0.8.4 (whitespace trimming fix, translated playback messages) is committed and pushed (3936cfb) but NOT released.

**How to apply:** re-run `npm audit` in `frontend/`; when a patched picomatch exists, update dependencies, rebuild
`frontend/build`, push, wait for green CI, then dispatch the Release workflow with version 0.8.4.
