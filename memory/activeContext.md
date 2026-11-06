# Active context

## Current focus

Designing inflected-form lookup and spelling suggestions using SDD. Phase 0 design is approved (`src:c34610a226566b0738a4`); user wants offline-only alternatives on language misses, dataset investigation and multiword corrections (`src:6f16694188c7fc53cc4a`). User likes the wireframe and requests phrase-headword typo search (`src:23d2651f2c57ed43c790`); the spec is approved (`src:665fce98e49af19f680d`) and Phase 2 core architecture is approved (`src:fd9eb7586fb1e9d78b2b`); quality gate passed using a writable temporary Go cache. Measured data and typo-ranking probes exposed spec issues; the partial amendment is approved (`src:8a02e642d2cb6cd1bb7f`). The final amendment is approved (`src:6dce82b0c9d386831cec`); implementation and review are in progress. The memory-limit amendment is approved (`src:bb20bbd7cfc98225d3b0`); run final verification and independent grade. The user confirmed v0.5.0 was published (`src:92edc1045f02c77efd04`).

Agentic-loadout 0.5.0 was rerun in this checkout for Claude and Codex with Go and TypeScript; the existing Agemem provider and Entire setup were preserved. The installer replaced the tracked SDD skill symlink, which was restored. Reload clients and review/trust hooks and MCP servers.

The codenav-go trial added a Go test for rejecting unsupported retry languages. Navigation, edit preview/application and diagnostics worked; targeted race and full quality gate passed. The test is included in the current checkout.

Footer and interface language implemented with i18next; full quality gate passed (46 frontend tests and Go tests), Chrome layout smoke passed four view/size combinations, and independent SDD regrade passed. Native WebKit layout with real long results remains unverified. Continue in this checkout.

Spanish online and offline support is implemented and local `./wdict` rebuilt. To address test-quality review, the cache restart test now recreates services and memory caches against shared disk without subprocesses; real-data checks are integration-tagged and documented. Default race tests and both local real-data tests passed; independent test-quality regrade passed, noting the lack of an actual process restart test. Legacy attribution amendment was approved (`src:517058a14965c942a353`); HTTP 502 handling follows `src:fbb1669e08460c8e2421`.

v0.4.0 is published with a Linux x86_64 AppImage after CI and the release workflow passed; the Go-installed executable reports v0.4.0 (`src:c306659d288af475a134`).

Launch experiments were removed at the user's request; the dictionary metadata error-and-retry fix remains (`src:951f2ed31a886ccf2478`).

## Lookup alternatives status (2026-11-06)

Implementation is complete in the working tree. The quality gate is green: svelte-check reports 0 errors and 0 warnings, 55 frontend tests pass, and Go tests pass with the gtk3 tag. Fixes applied after the earlier grade: a `decisions.md` implementation entry, a test that source-unavailable starts no alternatives, and separate `timed_out` and `unavailable` UI messages in EN and ES.

Independent grade (inline `code-grader`, 2026-11-06) returned PASS: no FAILs, F1–F6 mapped to code and tests. Open PARTIAL items: (1) suggestion search is CPU-bound, not a network request: on the installed English dictionary, `dictionari` and `dictionary` both hit the 2 s deadline, `wrods` took 1.46 s and `look foward to` 1.62 s. A raw index pass takes 118 ms; 90% of the time is in `alternativeDistance` running a full DP on every length-matched headword (2026-11-06 probe and CPU profile). Fixed on 2026-11-06: `alternativeDistance` now uses a banded OSA with an early exit and input-sized rows (approved OSA ≤ 2 unchanged). Real-dictionary timings fell to 0.40–0.69 s (was 1.46–2.00 s). Oracle test `internal/core/alternatives_distance_test.go` matches a reference OSA on all strings up to length 4. Open: ranking puts `dictionaric` ahead of `dictionary` for `dictionari` (alphabetical tie-break; no word-frequency data). Gate passed; `./wdict` rebuilt. Not yet verified in the native window, and not committed; (2) the "Suggestions timed out." text departs from the spec's "unavailable" wording (the user asked for distinct display in chat; record that or reword); (3) untested release checks: the "install dictionary" message, a per-language deadline with the other language still delivered, and a no-files-written check; (4) cleanup: unused `BaseFormSource` and `VerifyHeadword` fakes, and a real 50 ms sleep in `alternatives.test.js`.

## Blockers

Native appearance rendering remains to be verified by the user. The low-severity npm advisory, cross-distribution testing and bundled license review remain open. On 2026-11-06 `project_memory_maintain` flushed 2 pending notes; whether the remaining lookup-alternatives approval note is screened is unverified.

## Next steps

Get the independent grade (ask the user before spawning a code-grader), fix findings, then update progress. Verify the lookup-alternatives UI in the native window with real results. Commit only with the user's approval.

Verify footer, scrolling and language switching in the native window with real long results. The user reports the silent 502 behavior works.

Run the installed wdict and verify appearance controls, persistence and system changes in the native window.
