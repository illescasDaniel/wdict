# Active context

## Current focus

v0.2.0 is published with a Linux AppImage; the local Go-installed executable was updated and verified at v0.2.0. Online cache freshness is 365 days.

## Blockers

npm audit reports low-severity GHSA-gcqx-w48m-w2h4 in the Svelte build tool's deepmerge dependency; there is no upstream fix. Low findings remain visible but no longer fail the approved audit gate. Desktop screenshot remains unverified. Cross-distribution testing and bundled license review remain open.

## Next steps

Track the npm advisory and test the release on additional platforms. Track the upstream npm advisory. Verify native rendering and capture a screenshot; test on older Linux distributions and review bundled licenses. Build and test macOS and Windows binaries in a future release; the user explicitly requested v0.1.0 now.
