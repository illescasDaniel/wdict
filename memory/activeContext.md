# Active context

## Current focus

Preparing v0.2.0: online cache freshness is 365 days; tests and release notes are ready. User requested publishing and installing the release locally.

## Blockers

npm audit reports low-severity GHSA-gcqx-w48m-w2h4 in the Svelte build tool's deepmerge dependency; there is no upstream fix. Low findings remain visible but no longer fail the approved audit gate. Desktop screenshot remains unverified. Cross-distribution testing and bundled license review remain open.

## Next steps

Commit and push the v0.2.0 changes, verify CI, run the release workflow, then update the user’s installed binary. Track the upstream npm advisory. Verify native rendering and capture a screenshot; test on older Linux distributions and review bundled licenses. Build and test macOS and Windows binaries in a future release; the user explicitly requested v0.1.0 now.
