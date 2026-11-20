---
id: DOC-023
title: Testing wdict
summary: Shared test task, optional browser smoke test, and local Spanish integration fixtures
type: how-to
status: current
updated: 2026-11-20
verified:
covers: [Taskfile.yml, .github/workflows/ci.yml, frontend/scripts/footer-browser-check.mjs, internal/adapters/spanish_archive_local_test.go, internal/adapters/spanish_response_integration_test.go]
related: []
supersedes:
superseded_by:
---

# Testing wdict

Run the standard checks on Linux or macOS after installing the [build prerequisites](building.md) and running `go tool task setup`:

```sh
go tool task test
```

The default Go tests use deterministic fixtures and fake providers. The cache persistence test recreates the service and memory cache against the same on-disk cache, without launching a separate app process.

An optional real-browser smoke test checks footer positioning, scrolling and focus at desktop and narrow sizes. It requires the built frontend, Node.js, a Chrome-compatible browser and permission to bind a loopback server. From the repository root:

```sh
cd frontend
npm run build
CHROME_PATH=/path/to/chrome node scripts/footer-browser-check.mjs
```

If `google-chrome` is on `PATH`, omit `CHROME_PATH`. The check uses synthetic long content and does not replace testing the native WebKit window. It is not part of `go tool task test`.

To additionally check the real Spanish package and a saved MediaWiki response, supply local files and run the opt-in integration tests:

```sh
WDICT_TEST_SPANISH_ARCHIVE=/path/to/es-es.tar.zst \
WDICT_TEST_SPANISH_RESPONSE=/path/to/casa-response.json \
go test -tags 'gtk3 integration' -run '^TestGivenRealSpanish' ./internal/adapters
```

The package test verifies the pinned checksum during installation. These integration tests do not download files; tests whose local fixture is not supplied are skipped.

The CI workflow also builds the frontend and checks committed assets, audits frontend dependencies on Linux, and verifies the version in the packaged macOS app.
