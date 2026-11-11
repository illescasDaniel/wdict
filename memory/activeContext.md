# Active context

## Phonetic pronunciation (2026-11-11)

Released as v0.7.0 (IPA pronunciations) and installed locally. Open: grader item G8 (the approved 5-run median performance report was not re-measured; the user has not decided). Not verified: real-window check, flags without a flag font. Next task the user named: an AI voice for pronunciation (start with the `sdd-html-feature` design gate).

## Current focus

v0.6.0 is released (2026-11-06; lookup alternatives and interface-language result ordering). v0.6.1 (2026-11-09) adds `wdict --version`/`-v` (ldflags `main.version`, falling back to the `go install` module version, then `dev`) and a muted version label at the end of the footer via `App.Version()`. The user noticed v0.6.0 had no way to show its version. Release process: bump version files, rebuild `frontend/build` with `npm run build --prefix frontend`, wait for green CI, then dispatch the Release workflow (it creates the tag; never push `v*` by hand), then `go install -tags gtk3 github.com/illescasDaniel/wdict@vX.Y.Z`.

Earlier work (v0.5.0 footer and interface language, v0.4.0 Spanish support, silent HTTP 502 handling per `src:fbb1669e08460c8e2421`, Agentic-loadout 0.5.0 setup) is recorded in `progress.md` and `decisions.md`.

## Lookup alternatives status (2026-11-06)

Implementation is complete in the working tree. The quality gate is green: svelte-check reports 0 errors and 0 warnings, 55 frontend tests pass, and Go tests pass with the gtk3 tag. Fixes applied after the earlier grade: a `decisions.md` implementation entry, a test that source-unavailable starts no alternatives, and separate `timed_out` and `unavailable` UI messages in EN and ES.

Independent grade (inline `code-grader`, 2026-11-06) returned PASS: no FAILs, F1–F6 mapped to code and tests. Open PARTIAL items: (1) suggestion search is CPU-bound, not a network request: on the installed English dictionary, `dictionari` and `dictionary` both hit the 2 s deadline, `wrods` took 1.46 s and `look foward to` 1.62 s. A raw index pass takes 118 ms; 90% of the time is in `alternativeDistance` running a full DP on every length-matched headword (2026-11-06 probe and CPU profile). Fixed on 2026-11-06: `alternativeDistance` now uses a banded OSA with an early exit and input-sized rows (approved OSA ≤ 2 unchanged). Real-dictionary timings fell to 0.40–0.69 s (was 1.46–2.00 s). Oracle test `internal/core/alternatives_distance_test.go` matches a reference OSA on all strings up to length 4. Open: ranking puts `dictionaric` ahead of `dictionary` for `dictionari` (alphabetical tie-break; no word-frequency data). Gate passed; `./wdict` rebuilt. Not yet verified in the native window. Committed as `dc54795` and merged to `main`; shipped in v0.6.0; (2) the "Suggestions timed out." text departs from the spec's "unavailable" wording (the user asked for distinct display in chat; record that or reword); (3) untested release checks: the "install dictionary" message, a per-language deadline with the other language still delivered, and a no-files-written check; (4) cleanup: unused `BaseFormSource` and `VerifyHeadword` fakes, and a real 50 ms sleep in `alternatives.test.js`.

## Blockers

Release lesson: after bumping the version, rebuild `frontend/build` (`npm run build --prefix frontend`) instead of editing it by hand, because the build hash depends on the version and CI rejects stale assets. The Release workflow is dispatched manually and creates the tag; never push `v*` by hand.

Native appearance rendering remains to be verified by the user. The low-severity npm advisory, cross-distribution testing and bundled license review remain open. On 2026-11-06 `project_memory_maintain` flushed 2 pending notes; whether the remaining lookup-alternatives approval note is screened is unverified.

## Next steps

Verify the lookup-alternatives UI, footer, scrolling and language switching in the installed v0.6.0 native window with real long results. Resolve the open PARTIAL items above (ranking tie-break for `dictionari`, untested release checks, test cleanup). Native appearance verification, the low-severity npm advisory, cross-distribution testing and bundled license review remain open.
