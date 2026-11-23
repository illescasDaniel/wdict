---
name: release-0-9-0-published
description: 0.9.0 published with French support; it also carries the held 0.8.4 fixes
type: decision
date: 2026-11-23
status: current
supersedes: picomatch-advisory-blocks-0-8-4
superseded_by: 
source: 
---
The picomatch audit hold cleared (npm audit: 0 vulnerabilities). 0.9.0 was released instead of 0.8.4 because French is a
feature; the 0.8.4 changelog entries moved into 0.9.0 and no 0.8.4 tag exists. Release steps that bit us: the committed
`frontend/build` embeds the package version, so rebuild it after a version bump or CI fails the stale-assets check.
