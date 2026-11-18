# Now
Updated: 2026-11-17 (legacy memory audit and docs-1 setup)

## Doing
- macOS packaging: unsigned arm64 app and CI artifact prepared on feature/macOS; hosted CI, native GUI and audible playback remain unverified.
- Upgraded the agent setup to agentic-loadout 0.7.0 (memory-1 and docs-1 skills, agemem removed) and converted the project memory to `memory/` notes. Reviewed and merged to `main` on 2026-11-16.

## Done recently
- 2026-11-17: legacy memory audited (nothing left to convert); docs-1 set up in place for `sdd/design/`, `docs/reports/`, `docs/research_notes/` and `sdd/*/SPEC.md` (21 docs). 2026-11-18: folders moved under sdd/ and docs/, all 21 docs verified against code and stamped. Gap review: playback failure copy now translated in the Spanish UI (frontend), README Spanish license wording fixed; other gaps (hardcoded es CC BY-SA 4.0, no-heading page = unavailable, corrupt cache files, untrimmed control-char check, previous-result retention) kept on purpose; committed and pushed (3936cfb). Next: 0.8.4 release is held until npm audit passes (picomatch advisory, see memory/notes/picomatch-advisory-blocks-0-8-4.md); then dispatch the Release workflow.
- 2026-11-16: agentic-loadout 0.7.0 installed with `--without sdd`; old agemem files and `.agemem/` deleted after the user confirmed.
- 2026-11-15: v0.8.3 committed (`234a43a`): pronunciations play through the host audio player (`pw-play`, `paplay`, `aplay`); released as tag `v0.8.3` with a GitHub release ([note](notes/audio-native-host-player.md)).
- 2026-11-15: result ranking and card grouping committed (`a2cbfcc`) with the compact-senses grouping (`9dc7751`).
- 2026-11-15: British "though" fixed (`0c8cbfe`); AppImage now bundles eSpeak-NG (`46112b9`).

## Next
- Owner listens to pronunciation in the native window and the AppImage; agents only verified off screen and silent.
- Check the native window for the lookup-alternatives UI, footer, appearance and flags (never agent-verified).
- Decide whether to re-measure the 5-run performance report (grader item G8) for the IPA feature.
- Cross-distribution AppImage test and bundled-licence review.
- macOS and Windows binaries are untested.

## Open questions / blocked
- Does the low npm advisory GHSA-gcqx-w48m-w2h4 (deepmerge) still have no fix?
