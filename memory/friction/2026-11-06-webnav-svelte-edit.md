# webnav edit rejects .svelte files

- **Date:** 2026-11-06
- **Reproduction:** `mcp__webnav__edit` on `frontend/src/routes/+page.svelte`.
- **Observed:** "write tools only edit .cjs/.css/.cts/.html/.js/.jsx/.mjs/.mts/.ts/.tsx files, not '.svelte'".
- **Impact:** The coder had to use the standard Edit tool for `.svelte` files, so the type-checked write path does not cover Svelte components.
