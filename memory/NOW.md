# Now
Updated: 2026-11-16 (session 2026-11-16: agentic-loadout 0.7.0 upgrade and memory conversion)

## Doing
- On branch `agentic-loadout-0.7.0` (not pushed, not merged): upgraded the agent setup to agentic-loadout 0.7.0 (memory-1 and docs-1 skills, agemem removed) and converted the project memory to `memory/` notes. Waiting for the owner to review.

## Done recently
- 2026-11-16: agentic-loadout 0.7.0 installed with `--without sdd`; old agemem files kept in place for now.
- 2026-11-15: v0.8.3 committed (`234a43a`): pronunciations play through the host audio player (`pw-play`, `paplay`, `aplay`); local tags stop at `v0.8.2`, so confirm the release run and the system install.
- 2026-11-15: result ranking and card grouping committed (`a2cbfcc`) with the compact-senses grouping (`9dc7751`).
- 2026-11-15: British "though" fixed (`0c8cbfe`); AppImage now bundles eSpeak-NG (`46112b9`).
- 2026-11-14/15: Piper pronunciation repairs (appearance, Sábado, eight, vegetable) shipped in v0.8.0 and v0.8.1.

## Next
- Owner listens to pronunciation in the native window and the AppImage; agents only verified off screen and silent.
- Check the native window for the lookup-alternatives UI, footer, appearance and flags (never agent-verified).
- Decide whether to re-measure the 5-run performance report (grader item G8) for the IPA feature.
- Cross-distribution AppImage test and bundled-licence review.
- macOS and Windows binaries are untested.

## Open questions / blocked
- Can the old agemem memory files now be deleted: `memory/activeContext.md`, `progress.md`, `decisions.md`, `archive.md`, `memory/friction/`, and the `.agemem/` folder? They are kept only as the source of the converted notes.
- Should docs-1 be set up for the existing docs (`design/`, `reports/`, `sdd/`)? It is only installed.
- Does the low npm advisory GHSA-gcqx-w48m-w2h4 (deepmerge) still have no fix?
