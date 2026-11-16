---
name: version-reporting-ldflags
description: wdict --version comes from ldflags main.version, then the go install module version, else dev; footer shows it via App.Version()
type: decision
date: 2026-11-09
status: current
supersedes: 
superseded_by: 
source: 
---
Added in v0.6.1 because the user found v0.6.0 had no way to show its version. Task and `scripts/build-appimage.sh` pass
`-ldflags -X main.version=$(cat VERSION)`; `go install ...@vX.Y.Z` falls back to build info with the leading `v` trimmed;
anything else prints `dev`. The muted footer label calls the Go `App.Version()` binding and is hidden when unavailable.
