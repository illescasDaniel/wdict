# Agemem recall degraded — 2026-10-29

Reproduction: call `project_memory_recall` for the reported bilingual `diccionario` lookup. It returned `decision_status: degraded` and diagnostic `jev_fallback` / `budget_exhausted`, while still returning bank, note and source evidence. Impact: combined retrieval sufficiency was not certified; checked the current source and reproduced the lookup directly instead. No provider configuration was changed.

The same diagnostic recurred during the follow-up provider investigation; current source and HTTP responses were checked independently.

Remembering the captured 502 preference (`src:fbb1669e08460c8e2421`) returned pending/degraded with `jev_fallback` / `budget_exhausted`; maintenance flushed zero notes. Impact: the note remains unscreened, with the original source and decision preserved in the Markdown bank.
