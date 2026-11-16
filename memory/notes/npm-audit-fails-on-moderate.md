---
name: npm-audit-fails-on-moderate
description: CI and release fail on moderate or higher npm audit findings and only display low ones
type: decision
date: 2026-10-28
status: current
supersedes: 
superseded_by: 
source: 
---
The user approved `npm audit --audit-level=moderate` (see `.github/workflows/ci.yml` and `release.yml`): moderate and higher
fail, low findings stay visible. SvelteKit was upgraded on 2026-10-27 to clear findings. A low-severity advisory
(GHSA-gcqx-w48m-w2h4 in deepmerge) had no upstream fix on 2026-10-28; its current state was not checked.

**How to apply:** if a release run fails on audit, update dependencies (as `d260da3` did on 2026-11-14) rather than loosening
the threshold.
