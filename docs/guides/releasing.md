---
id: DOC-026
title: Preparing a release
summary: Version updates, CI, manually dispatched AppImage release and recovery after failure
type: how-to
status: current
updated: 2026-11-20
verified:
covers: [.github/workflows/release.yml, scripts/release_notes.py, VERSION, frontend/package.json, frontend/package-lock.json, CHANGELOG.md]
related: []
supersedes:
superseded_by:
---

# Preparing a release

Before a release, update `VERSION`, the frontend package version and lockfile, and the matching dated section in `CHANGELOG.md`. Commit these changes to `main` and ensure CI passes. Start the **Release** workflow on `main` with the matching version (for example `0.4.0`). It validates the version and release notes, runs the checks, builds a Linux AppImage, creates the `v`-prefixed tag and publishes a GitHub release with the AppImage attached. Do not run it until the release is ready; this workflow publishes only the AppImage.

If publishing fails after the tag is pushed, inspect the existing tag and release before retrying; the workflow refuses to reuse an existing tag. The version tag also makes the Go module available to Go tooling. Release versions must not be retagged after publication.
