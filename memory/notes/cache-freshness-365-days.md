---
name: cache-freshness-365-days
description: Online results stay fresh for 365 days; a stale result is used when the provider fails
type: decision
date: 2026-10-28
status: current
supersedes: 
superseded_by: 
source: 
---
Online dictionary results are fresh for 365 days (`cacheFreshness` in `internal/core/lookup.go`) before a refresh is tried; a
stale disk result is still served when the provider is unavailable. Released in v0.2.0.
