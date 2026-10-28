# Active context

## Current focus

Routine CI checks builds and tests without packaging; a manually dispatched release workflow builds the Linux AppImage, tags the selected main commit and creates a GitHub release.

## Blockers

npm audit reports low-severity GHSA-gcqx-w48m-w2h4 in the Svelte build tool's deepmerge dependency; there is no upstream fix. The existing audit gate remains unchanged and therefore fails. Desktop screenshot and native rendering after the SvelteKit upgrade remain unverified. Cross-distribution testing and bundled license review remain open.

## Next steps

Resolve the upstream npm advisory or agree on an audit policy before releasing. Verify native rendering and capture a screenshot; test on older Linux distributions and review bundled licenses. Build and test macOS and Windows binaries before the first release.
