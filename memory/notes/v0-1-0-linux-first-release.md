---
name: v0-1-0-linux-first-release
description: v0.1.0 was released Linux-only; macOS and Windows binaries stay deferred
type: decision
date: 2026-10-28
status: current
supersedes: first-release-deferred-until-all-platforms
superseded_by: 
source: 
---
Reversal: on 2026-10-28 the user explicitly asked for the v0.1.0 release after CI and `go install` work, instead of waiting
for macOS and Windows. It shipped with a Linux x86_64 AppImage, and a clean-cache `go install` of the tagged module
succeeded. This replaces [first-release-deferred-until-all-platforms](first-release-deferred-until-all-platforms.md).

**How to apply:** macOS and Windows binaries are still future work; nothing was tested there.
