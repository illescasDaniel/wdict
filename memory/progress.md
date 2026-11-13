# Progress

## Open

- [ ] Commit the offline duplicate-results fix (`internal/adapters/offline.go`, `offline_format_test.go`) and the new `opencode.json`; the installed v0.7.0 binary still shows duplicate results until the next release.

- [x] Fixed duplicate offline results (2026-11-13): `scanDictionary` dedupes repeated/self-folding `.syn` aliases; regression test added; quality gate green; real `dogfood` lookup returns 3 entries (was 6). Recorded in `memory/decisions.md` and as durable note `src:9f5e5a6626331a4ce3e4`.
- [x] Added `opencode.json` (agemem, webnav, codenav-go) so the OpenCode session loads the project MCPs, mirroring `.mcp.json` (2026-11-13).

- [x] Implemented lookup alternatives and passed the quality gate on 2026-11-06 (55 frontend tests, Go tests with gtk3 tag, svelte-check clean).
- [x] Rebuilt `frontend/build` with `go tool task build` on 2026-11-06; the shipped bundle now includes the `timed_out` text (`frontend/build/_app/immutable/nodes/2.DDDowKUZ.js`).
- [ ] Get an independent grade of the lookup-alternatives branch (not yet run; ask before spawning a code-grader); fix findings. The branch is largely untracked, so the grader must review untracked files too.

- [ ] Verify fixed footer and language switching with real long results in the native WebKit window.

- [ ] Verify automatic appearance interactively in the native window.


- [ ] Resolve npm audit GHSA-gcqx-w48m-w2h4 in deepmerge; no upstream fix available as of 2026-10-28. Low findings remain visible; user approved failing the audit at moderate and above.

- [ ] Capture a usable app screenshot and verify native rendering after the SvelteKit upgrade.
- [ ] Generate and test macOS and Windows binaries in a future release.

- [ ] Test the AppImage interactively and on older target distributions; review bundled dependency licenses before distribution.

- [ ] Verify end-to-end interactive startup and cold-disk performance.
- [ ] Test the app on macOS and Windows before claiming support.
- [ ] Design inflected-form lookup and spelling suggestions.
- [x] Selected and implemented the Spanish offline dataset; see README for attribution.

- [ ] Review and trust the project hooks, then test the MCP and hook checklist.

## Done

- [x] 2026-11-11: updated to agentic-loadout 0.6.0 (agemem 0.3.0, `memory-wrapup` skill, refreshed agents/workflow, memory rules in AGENTS.md) and replaced the `sdd-feature` skill with the global `sdd-html-feature` skill (AGENTS.md and `.agents/development-workflow.md` updated; old skill removed).

- [x] User approved the lookup-alternatives memory-limit amendment (`src:bb20bbd7cfc98225d3b0`).

- [x] User approved the final lookup-alternatives amendment (`src:6dce82b0c9d386831cec`).

- [x] User approved the partial lookup-alternatives amendment (`src:8a02e642d2cb6cd1bb7f`); memory screening is pending.

- [x] User approved the lookup-alternatives core architecture (`src:fd9eb7586fb1e9d78b2b`); memory screening returned pending.

- [x] User approved the lookup-alternatives spec (`src:665fce98e49af19f680d`).

- [x] Confirmed the five lookup-alternatives spec decisions and drafted the spec (`src:de40e77c3089de7f6007`).

- [x] User approved the lookup-alternatives design and wireframe (`src:c34610a226566b0738a4`).

- [x] User confirmed v0.5.0 was published (`src:92edc1045f02c77efd04`).

- [x] Reran the public agentic-loadout 0.5.0 installer for Claude and Codex, retaining the existing provider and Entire setup and restoring the tracked SDD skill symlink.
- [x] Exercised codenav-go navigation, type-checked edit preview/application and diagnostics while adding unsupported retry-language coverage; targeted race test and full quality gate passed.
- [x] Implemented fixed footer and English/Spanish UI using i18next; rebuilt frontend, passed quality gate (46 frontend tests and Go tests), Chrome layout smoke and independent SDD regrade.
- [x] Approved i18next-based footer-language architecture (`src:24f6db6819cfaaa41e94`).
- [x] Approved the footer and interface-language spec (`src:cb4e45d81b39aa662285`).
- [x] Agreed the footer and interface-language spec decisions, including localized dates (`src:b100cb1de6d73aa010df`).
- [x] Approved the fixed-footer and interface-language wireframe (`src:43ce9da020fb43286fd5`).
- [x] Pushed commit f583548, verified CI run 40920837975, published v0.4.0 with the release workflow, and updated the Go-installed executable to v0.4.0 (`src:c306659d288af475a134`).
- [x] Replaced the subprocess cache restart test with fresh service and memory-cache instances against the same disk cache; separately tagged and documented real Spanish-data integration checks. Default race tests and both real-data checks passed locally using the verified archive and captured response. Independent test-quality regrade passed; this does not test an actual process restart.
- [x] Resolved Spanish support grader failures: preserve legacy cached entries with an attribution notice and reject duplicate Spanish language sections. Independent regrade passed with a partial observation on integration test setup; frontend tests, Svelte checks, build and Go race tests passed.

- [x] Implemented silent HTTP 502 behavior for English and Spanish online lookups while retaining stale cache fallback (`src:fbb1669e08460c8e2421`).

- [x] Investigated `diccionario`: the installed Spanish dictionary finds it, the English dictionary misses, and a separate English online request returned a Cloudflare HTTP 502 of uncertain underlying cause; distinguish this online failure from an ordinary missing entry.

- [x] Approved Spanish support spec and prepared core contracts (`src:6e99afe45172b4ac8823`).

- [x] Approved automatic bilingual lookup design on 2026-10-29 (`src:58170218c984fa66a08b`).

- [x] Tried launch optimizations without demonstrating a speedup; the user saw no visual difference from the placeholder, so experiments were removed (`src:92330b20cd5d33849f20`, `src:951f2ed31a886ccf2478`).
- [x] Fixed the existing dictionary metadata failure state so users can see the error and retry without restarting.

- [x] Published v0.3.0 with a Linux AppImage and updated the local Go-installed executable (`src:7b175f2d2ec3b891cb96`).

- [x] Added Given/When/Then comments and formatted appearance changes; frontend checks and 15 tests passed on 2026-10-28.

- [x] Approved and implemented the automatic appearance design with tests on 2026-10-28 (`src:f0c8c5907b9be0a2eada`).

- [x] Published v0.2.0 with a Linux AppImage and updated the user’s Go-installed executable; version metadata confirms v0.2.0 (`src:69a073eeb1caadca4326`).

- [x] Changed cached online results to remain fresh for 365 days, with expiry boundary and cross-process coverage (`src:bc2bcf9382dacdc66077`).

- [x] Verified online lookup results persist across separate processes; fresh disk hits avoid the provider and stale disk results are used when it fails.

- [x] User confirmed the search-field autofocus works; clarified README installation options (`src:f3e85e0506d5e89bd969`).

- [x] Released v0.1.0 with a Linux x86_64 AppImage; clean-cache go install of the tagged module succeeded (`src:5c122ea6bbaa9ba14638`).

- [x] Prepared the GitHub Go module path and reproducible embedded frontend; isolated versioned go install built and launched the app on 2026-10-28.
- [x] Applied the approved audit threshold and verified GitHub CI run 40772755480 passed.

- [x] Split lightweight CI from a manually dispatched GitHub release workflow on 2026-10-28; no tag or release created (source `src:a8efe69c530e6bb4f0af`).

- [x] Upgraded SvelteKit dependencies to resolve npm audit findings, added MIT license and 0.1.0 metadata, and configured Linux AppImage CI on 2026-10-27. Local tests and AppImage build passed; GitHub CI run 40527973403 built and uploaded the AppImage.

- [x] Added pinned Task Go tool and contributor setup, build, test and AppImage commands on 2026-10-27.

- [x] Added a Linux AppImage build script, icon and desktop entry; built and smoke-launched an x86_64 image on CachyOS on 2026-10-27.

- [x] User confirmed the improved offline installation and lookup work in native use on 2026-10-27; initial window dimensions now scale to the primary display while retaining the requested proportions.

- [x] Independent regrade passed after fixing punctuation-only and large offline result cases on 2026-10-27. Practical attribution approach requested by user (`src:7277bd055edce9e9b478`).

- [x] Improved first offline lookup with direct StarDict index scans and bounded in-memory result caching; representative real-dictionary equivalence tests passed, and three warm-cache backend runs took 134–151 ms on 2026-10-27. Added practical source and license attribution to the README; links already exist in the app.

- [x] Recorded Linux startup asset, memory, CPU and local lookup baseline on 2026-10-27 (`src:e03003324dc7c8106d3c`); first offline lookup misses the 250 ms target and interactive startup remains unmeasured.

- [x] Aligned test names and Given/When/Then sections; independent regrade returned PASS and test quality PASS on 2026-10-27.

- [x] User confirmed the revised live offline download and app check on Linux on 2026-10-27 (`src:52da923f20299efdd7ee`). Documented content-rights evidence and added an offline extraction and formatting notice.

- [x] Independent regrade returned PASS on 2026-10-27, with live download, performance baseline and rights review still partial; fixed query boundary control validation.

- [x] User confirmed readable high-DPI layout, visible keyboard focus and responsiveness during a lookup on their Linux display on 2026-10-27 (`src:655dd7c5a439c627122e`).
- [x] Added duplicate-download state protection and readable offline HTML definition extraction, with regression tests including real `AI` entry.

- [x] Added deterministic failed disk replacement preservation coverage and improved focus ring contrast; GTK3 Go tests, frontend checks, 9 tests and frontend build passed. Independent regrade remains NEEDS WORK because native high-DPI and visible focus are unverified. Real archive integration measured 9.20 s first lookup and 0.35 ms next lookup on 2026-10-27.

- [x] Built and smoke-launched the Linux Wails/SvelteKit app; tested the verified English Wiktionary archive and offline word/phrase lookups on 2026-10-27. A further implementation cycle fixed cache startup, provider status, installer recovery and cache eviction; the final regrade still returned NEEDS WORK due to remaining tests and native high-DPI verification.

- [x] Checked online provider terms and sample phrases; verified a minimal Wails/SvelteKit Linux asset build and retained license metadata on 2026-10-27. See `design/word-lookup/verification-2026-10-27.md`.

- [x] Wireframe and architecture approved; Go modules and project-local npm selected on 2026-10-27.
- [x] Started lookup orchestration, response-cache adapters and focused tests on 2026-10-27.

- [x] Drafted the offline dictionary download wireframe for review on 2026-10-27.

- [x] Spec approved with in-app download required for v1; prepared Go core contracts on 2026-10-27.

- [x] Specified optional in-app dictionary download and local-first lookup on 2026-10-27.

- [x] Added in-memory and on-disk caching to the v1 spec and design on 2026-10-27.

- [x] Agreed v1 scope and five design decisions, and drafted the spec on 2026-10-27.

- [x] Selected Wails, Go and SvelteKit and researched starter options on 2026-10-27.

- [x] Initial design approved with changes in chat on 2026-10-26; documented stack options and proposed decisions on 2026-10-27.

- [x] Prepared the initial dictionary GUI and system designs on 2026-10-26.

- [x] Listed repository MCP and hook configuration on 2026-10-26.

- [x] Approved and implemented result order that follows the interface language (2026-11-06): spec amendment in `specs/footer-language/SPEC.md`, two frontend tests, `$:` ordering in `+page.svelte`, rebuilt `frontend/build`. Quality gate and independent review passed; not committed.

- [x] Released v0.6.0 on 2026-11-06 (bump `1de1e80`, asset rebuild `ad7cfee`): CI green, Release workflow succeeded, AppImage attached, proxy serves it, and `go install` updated `/home/daniel/go/bin/wdict` to v0.6.0.

- [x] Added `wdict --version`/`-v` and a footer version label for v0.6.1 on 2026-11-09 (Go: `resolveVersion` with tests, ldflags in Taskfile and `build-appimage.sh`; frontend: footer label with 3 tests). Gate green (60 frontend tests, Go tests). Not yet verified in the native window.

- [x] Phonetic pronunciation (IPA) built through `sdd/phonetic-pronunciation.html` on 2026-11-11, revisions R1 to R3 approved in chat (design, spec, architecture, design R2, spec R2, architecture R2, spec R3). Quality gate green; QA matched S1 to S13 and the independent grader returned PASS on R2. The spec R3 change (G7) and the G6 test were fixed after grading and are not re-graded. Not committed. Not verified: real Wails/WebKitGTK window, flags on systems without a flag font, screen readers.

- [x] Phonetic pronunciation revisions R3 (limit after flag mapping) and R4 (ignore "dialectal"/"obsolete"; generic UK+US shows no flags) built and graded PASS on 2026-11-11; not committed. Open: G8 (5-run performance medians), real-window check.

- [x] Committed phonetic pronunciation and bumped to v0.7.0 on 2026-11-11 (VERSION, frontend package files, README, CHANGELOG, rebuilt `frontend/build`). Released v0.7.0 on 2026-11-11: CI green, Release workflow succeeded, AppImage `wdict-x86_64.AppImage` attached, and `go install -tags gtk3 github.com/illescasDaniel/wdict@v0.7.0` updated `/home/daniel/go/bin/wdict` (`wdict --version` printed 0.7.0).
- [ ] Next feature: AI voice for pronunciation (not started).
