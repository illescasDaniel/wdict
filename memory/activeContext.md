# Active context

## Current focus

The user confirmed the improved offline installation and lookup work in native use. The initial window now scales the 1467 × 1029 reference against a 2560 × 1600 display using the primary display's usable area; native visual verification is next.

## Blockers

Native interactive startup timing remains open; macOS and Windows verification is deferred to a later task. Real-dictionary search equivalence has been checked on representative words and an alias, not exhaustively. Agemem remember remains pending/degraded due to missing credentials.

## Next steps

Verify the new window size in native use. Measure end-to-end interactive performance and cold-disk behavior; test macOS and Windows in a later task before claiming support. Use Go modules and project-local npm only. Review project hooks and test the MCP/hook checklist when returning to Agemem evaluation.
