# Active context

## Current focus

Agentic-loadout 0.5.0 was rerun in this checkout for Claude and Codex with Go and TypeScript; the existing Agemem provider and Entire setup were preserved. The installer replaced the tracked SDD skill symlink, which was restored. Reload clients and review/trust hooks and MCP servers.

The codenav-go trial added a Go test for rejecting unsupported retry languages. Navigation, edit preview/application and diagnostics worked; targeted race and full quality gate passed. The change is uncommitted.

Footer and interface language implemented with i18next; full quality gate passed (46 frontend tests and Go tests), Chrome layout smoke passed four view/size combinations, and independent SDD regrade passed. Native WebKit layout with real long results remains unverified. Continue in this checkout.

Spanish online and offline support is implemented and local `./wdict` rebuilt. To address test-quality review, the cache restart test now recreates services and memory caches against shared disk without subprocesses; real-data checks are integration-tagged and documented. Default race tests and both local real-data tests passed; independent test-quality regrade passed, noting the lack of an actual process restart test. Legacy attribution amendment was approved (`src:517058a14965c942a353`); HTTP 502 handling follows `src:fbb1669e08460c8e2421`.

v0.4.0 is published with a Linux x86_64 AppImage after CI and the release workflow passed; the Go-installed executable reports v0.4.0 (`src:c306659d288af475a134`).

Launch experiments were removed at the user's request; the dictionary metadata error-and-retry fix remains (`src:951f2ed31a886ccf2478`).

## Blockers

Native appearance rendering remains to be verified by the user. The low-severity npm advisory, cross-distribution testing and bundled license review remain open.

## Next steps

Review and commit the Go test and relevant memory update if requested.

Verify footer, scrolling and language switching in the native window with real long results. The user reports the silent 502 behavior works.

Run the installed wdict and verify appearance controls, persistence and system changes in the native window.
