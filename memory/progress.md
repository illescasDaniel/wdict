# Progress

## Open

- [ ] Test the AppImage interactively and on older target distributions; review bundled dependency licenses before distribution.

- [ ] Verify end-to-end interactive startup and cold-disk performance.
- [ ] In a later task, test the app on macOS and Windows before claiming support.
- [ ] After v1, consider inflected-form lookup and spelling suggestions.
- [ ] Evaluate Spanish and select a licensed dataset for the in-app offline dictionary download.

- [ ] Review and trust the project hooks, then test the MCP and hook checklist.

## Done

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
