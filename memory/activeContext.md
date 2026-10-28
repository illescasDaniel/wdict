# Active context

## Current focus

User confirmed autofocus; README and UI changes committed locally. A separate-process regression test confirms persistent online cache behavior across restarts.

## Blockers

npm audit reports low-severity GHSA-gcqx-w48m-w2h4 in the Svelte build tool's deepmerge dependency; there is no upstream fix. Low findings remain visible but no longer fail the approved audit gate. Desktop screenshot remains unverified. Cross-distribution testing and bundled license review remain open.

## Next steps

Keep local commits unpushed as requested; continue with any user feedback. Track the upstream npm advisory. Verify native rendering and capture a screenshot; test on older Linux distributions and review bundled licenses. Build and test macOS and Windows binaries in a future release; the user explicitly requested v0.1.0 now.
