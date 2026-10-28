# Active context

## Current focus

User requested fixing CI, then preparing Go installation and publishing v0.1.0 with an AppImage. Audit policy was approved for moderate and higher; verify CI before the Go-install changes.

## Blockers

npm audit reports low-severity GHSA-gcqx-w48m-w2h4 in the Svelte build tool's deepmerge dependency; there is no upstream fix. Low findings remain visible but no longer fail the approved audit gate. Desktop screenshot and native rendering after the SvelteKit upgrade remain unverified. Cross-distribution testing and bundled license review remain open.

## Next steps

Verify CI, prepare Go installation and run the v0.1.0 release workflow. Track the upstream npm advisory. Verify native rendering and capture a screenshot; test on older Linux distributions and review bundled licenses. Build and test macOS and Windows binaries in a future release; the user explicitly requested v0.1.0 now.
