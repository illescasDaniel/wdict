# Agemem recall degraded — 2026-10-29

A `project_memory_recall` for `startup blank window splash screenshot initial appearance Wails` returned `decision_status: degraded`, diagnostic `stage: jev_fallback`, `cause: budget_exhausted`. Reproduction: issue the same recall query with the configured cascade; the failure depends on provider budget. Impact: decision-assisted retrieval was incomplete, although local bank, notes and BM25 source results were returned and could be checked against project source. No provider or fallback configuration was changed.

The query `implement native startup snapshot placeholder Wails webview startup` also returned the same degraded diagnostic later that day.

The query `remove startup placeholder keep dictionary error fix and note experiments unsuccessful` also degraded. Capturing the user's request succeeded, but `project_memory_remember` returned pending and unscreened with the same budget-exhausted diagnostic. `project_memory_maintain` flushed zero notes and reported degraded consolidation. Impact: the durable note is not screened; the bank records the captured source directly. Retry maintenance when the configured provider is available.

The query `commit push dictionary metadata retry fix no release launch experiment memory` also returned the same degraded diagnostic; project files were inspected directly before committing.
