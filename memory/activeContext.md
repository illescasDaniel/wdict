# Active context

## Current focus

v0.1.0 is published with a Linux x86_64 AppImage; a clean-cache Go install of the tagged module succeeded. CI and the release workflow passed.

## Blockers

npm audit reports low-severity GHSA-gcqx-w48m-w2h4 in the Svelte build tool's deepmerge dependency; there is no upstream fix. Low findings remain visible but no longer fail the approved audit gate. Desktop screenshot and native rendering after the SvelteKit upgrade remain unverified. Cross-distribution testing and bundled license review remain open.

## Next steps

Track the upstream npm advisory and update when a fix is available. Verify native rendering and capture a screenshot; test on older Linux distributions and review bundled licenses. Build and test macOS and Windows binaries in a future release; the user explicitly requested v0.1.0 now.
