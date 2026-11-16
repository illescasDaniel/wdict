---
name: update-system-install-after-release
description: After a release is published, update the user's installed wdict with go install and confirm the version
type: preference
date: 2026-10-28
status: current
supersedes: 
superseded_by: 
source: 
---
Standing instruction from the user (first given for v0.2.0, repeated since): when a release is published, also update their
system installation: `go install -tags gtk3 github.com/illescasDaniel/wdict@vX.Y.Z`, then `wdict --version` must print the new
version (binary in `/home/daniel/go/bin`).

**How to apply:** do it only after the Release workflow succeeded and the tag exists; it needs network access.
