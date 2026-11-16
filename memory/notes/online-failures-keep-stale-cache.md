---
name: online-failures-keep-stale-cache
description: HTTP 502 from the online endpoints is a silent error: keep stale results, do not cache a miss, show a neutral message
type: decision
date: 2026-10-29
status: current
supersedes: 
superseded_by: 
source: 
---
The user asked that HTTP 502 from the online lookup endpoints be treated as a silent error for now. Keep the stale cache
fallback, never cache a failure as a missing entry, and show a neutral message when there are no results. A real case was
`diccionario`, where a separate English online request returned a Cloudflare 502 of unknown cause.
