---
name: release-via-dispatched-workflow
description: Releases come from a manually dispatched workflow that creates the tag; never push v* tags by hand
type: decision
date: 2026-10-28
status: current
supersedes: 
superseded_by: 
source: 
---
Routine CI does not build AppImages. A separate `workflow_dispatch` Release workflow (input: version) checks that `VERSION`
and `frontend/package.json` match, refuses an existing tag, builds the AppImage, and creates the tag and GitHub release with
the matching `CHANGELOG.md` section (`scripts/release_notes.py` needs exactly one `## [X.Y.Z] - date`).

**Why:** on 2026-11-06 a hand-pushed `v0.6.0` was deleted before any release or proxy fetch and recreated by the workflow
on the same commit.
**How to apply:** bump `VERSION`, `frontend/package.json` and lock, README version and CHANGELOG; rebuild `frontend/build`
with `npm run build --prefix frontend` (or `go tool task build`), never sed-edit it, because its hash depends on the
version and CI rejects stale assets; wait for green CI; dispatch the workflow. The version string comes from
`-ldflags -X main.version=$(cat VERSION)` in Task and the AppImage script, falls back to the module version for
`go install`, else `dev`. Release timing is the user's call. After publishing, follow
[update-system-install-after-release](update-system-install-after-release.md).
