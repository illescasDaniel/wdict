# Active context

## Current pronunciation repair (2026-11-14)

The user confirmed the pronunciation repair works on 2026-11-14 and requested a release. Keep verification off screen and silent. Preserve the shown IPA rather than substituting headword speech. Diagnosis provenance: `src:0ae5aa2f26e677b96ca6`, screened `note:50`. Next step: finish version preparation, commit and push, wait for CI, then dispatch the project Release workflow.

Correction to the historical status below: G1 was approved in chat (“g1”) and resolved to ≤250 MB installed; it is not awaiting a decision. G4 English-only fallback copy and packaging/native-window checks remain separate open items. Historical build/commit status below is superseded by `progress.md` and Git.

## Natural voices / Piper (2026-11-13, Verify done — user decisions pending)

Built through `sdd/piper-voices.html` (design/spec/architecture approved; "Approved"/"Spec approved"/"Architecture approved"). Go + frontend implemented by delegated coders, tests first. QA exercised the real chain in a sandbox (real PyPI venv+pip install 5.1 s, real checksummed HF voice download, real synthesis 0.34 s peak 32767, cache replay byte-equal; all matches). Independent grader: **PASS** (criteria 2 and 6 PARTIAL). Fixed: Piper failure logging (Q2, pre-grade), traceability fill (G2), voice-specific failure copy `voiceInstallationFailed` en/es (G3, post-grade, not re-graded), Build-tab record for VoicePack.Blocked (G6). Open user decisions: **G1** installed runtime = 210 MB (spike's ≤120 MB stale — recommended: accept measured number), **G4** English-only F4/Blocked copy for v1 (recommended: accept + record). Not committed. **Next step: user decides G1/G4, real-window check of both features, then commit.**

## Audio pronunciation (2026-11-13, revision 3 — IPA-faithful eSpeak voice)

Revision 3 after the user chose the converter ("that matters to me"): the voice path now speaks the shown IPA via a pure `IPAToMnemonics` converter (94/94 table entries oracle-verified against `espeak-ng -q -x`), falling back to the plain-text headword when symbols are unmappable (F4 restored). Gate green, audibility integration tests pass in both modes. The user's real-window check of the converter (with the ɒ→0 and stress-anchor fixes) is still owed. The piper-voices feature's QA/grader pass covered the whole current chain. Not committed.

Revision 2 (superseded): the user reported "i hear no sound"; the approved SSML `<phoneme>` mechanism rendered digitally silent WAVs (ffmpeg volumedetect: max −91 dB vs −5.2 dB for plain text). With chat approval ("Speak the headword (Recommended)"), C5 was amended to plain-text headword speech; the real-binary integration test gained the audibility peak check (≥ 1000) that would have caught the original bug.

Revision 1: gates approved; implemented by delegated coders, tests first; QA matched all reachable scenarios in a sandbox (webview rendered empty there, so real-window items were unverified by QA); independent grader **PASS** (criteria 2 and 5 PARTIAL); G1 fixed after grading (plain-text headword retry).

## Offline duplicate results fix (2026-11-13)

Released as v0.7.1 (commit `c8977f6`, bump `cade98e`, release run 43461376019) and reinstalled via `go install -tags gtk3 github.com/illescasDaniel/wdict@v0.7.1` (`/home/daniel/go/bin/wdict` prints 0.7.1). `opencode.json` (project MCPs for OpenCode, mirroring `.mcp.json`) was committed with the fix; `memory/friction/2026-11-13-opencode-mcp-not-loaded.md` records the setup gap.

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
