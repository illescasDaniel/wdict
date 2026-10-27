# Active context

## Current focus

The first offline search now scans the compact StarDict files instead of building a full in-memory index. Three fresh-process tests on the Linux reference machine measured 134–151 ms for `hello` with likely warm OS cache, versus 8.681 s previously. The README and existing in-app links give dictionary attribution and license information; the user prefers a pragmatic review (`src:7277bd055edce9e9b478`). Independent regrade returned PASS.

## Blockers

Native interactive startup timing remains open; macOS and Windows verification is deferred to a later task. Real-dictionary search equivalence has been checked on representative words and an alias, not exhaustively. Agemem remember remains pending/degraded due to missing credentials.

## Next steps

Verify the updated app in native use. Measure end-to-end interactive performance and cold-disk behavior; test macOS and Windows in a later task before claiming support. Use Go modules and project-local npm only. Review project hooks and test the MCP/hook checklist when returning to Agemem evaluation.
