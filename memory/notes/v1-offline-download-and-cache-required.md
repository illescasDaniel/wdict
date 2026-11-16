---
name: v1-offline-download-and-cache-required
description: v1 requires in-memory and on-disk caching and an in-app download of a curated offline dictionary
type: decision
date: 2026-10-27
status: current
supersedes: v1-scope-offline-and-cache-optional
superseded_by: 
source: 
---
Reversal: the user asked for both an in-memory and an on-disk response cache to avoid unnecessary requests, and approved
the spec with in-app download of a curated dictionary required for v1 (installing it stays optional per user). This
replaces [v1-scope-offline-and-cache-optional](v1-scope-offline-and-cache-optional.md).

**Why:** avoid repeat requests; users should not have to install StarDict files by hand (go-stardict was the reader
candidate).
**Rejected:** manual dictionary installation by the user.
**How to apply:** inflected forms and spelling suggestions, deferred at the time, were built later as lookup alternatives
(v0.6.0): see [lookup-alternatives-design-choices](lookup-alternatives-design-choices.md).
