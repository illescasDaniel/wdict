# Active context

## Current focus

Linux AppImage packaging has been added using the Wails v3.0.0 generator. A local x86_64 image was built on CachyOS and launched with extracted runtime; frontend assets were served.

## Blockers

Cross-distribution AppImage testing and bundled dependency license review remain open; native interactive startup timing remains open; macOS and Windows verification is deferred to a later task. Real-dictionary search equivalence has been checked on representative words and an alias, not exhaustively. Agemem remember remains pending/degraded due to missing credentials.

## Next steps

Test the AppImage interactively and on older target distributions, and review bundled dependency licenses before distribution. Verify the new window size in native use. Measure end-to-end interactive performance and cold-disk behavior; test macOS and Windows in a later task before claiming support. Use Go modules and project-local npm only. Review project hooks and test the MCP/hook checklist when returning to Agemem evaluation.
