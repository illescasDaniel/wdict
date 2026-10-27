# Active context

## Current focus

Contributor tasks are available through the pinned Task Go tool; setup, build and tests passed on Linux. Linux AppImage packaging was built and smoke-launched on CachyOS.

## Blockers

Cross-distribution AppImage testing and bundled dependency license review remain open; native interactive startup timing remains open; macOS and Windows verification is deferred to a later task. Real-dictionary search equivalence has been checked on representative words and an alias, not exhaustively. Agemem remember remains pending/degraded due to missing credentials.

## Next steps

Test the AppImage interactively and on older target distributions, and review bundled dependency licenses before distribution. Verify the new window size in native use. Measure end-to-end interactive performance and cold-disk behavior; test macOS and Windows in a later task before claiming support. Use Go modules and project-local npm only. Review project hooks and test the MCP/hook checklist when returning to Agemem evaluation.
