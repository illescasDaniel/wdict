# Agemem recall degraded — 2026-10-29

## 2026-11-04 — Internationalization research

Reproduction: call `project_memory_recall` for Wails/Svelte internationalization; two calls returned `decision_status: degraded`, with `jev_fallback` / `budget_exhausted` on the second. Impact: retrieval sufficiency was not certified; checked current package configuration and official documentation independently. No provider configuration was changed.

The same diagnostic recurred for the dedicated-library architecture recall and for remembering the user's library choice (`src:a99625ad49fc3a442cfc`). The note remains pending and unscreened; the original choice is recorded in the Markdown decision bank.

Following the user's authorization, set `.agemem/config.json` decision backend to `qsiv` and ran maintenance. It reported 17 notes flushed; recall returned the library choice as `note:31`, but consolidation remained degraded. A subsequent remember of architecture approval (`src:24f6db6819cfaaa41e94`) was pending/degraded with no diagnostic, and recall remained degraded. The original approval is retained in the decision bank. The precise QSIV failure is unknown.

Reproduction: call `project_memory_recall` for the reported bilingual `diccionario` lookup. It returned `decision_status: degraded` and diagnostic `jev_fallback` / `budget_exhausted`, while still returning bank, note and source evidence. Impact: combined retrieval sufficiency was not certified; checked the current source and reproduced the lookup directly instead. No provider configuration was changed.

The same diagnostic recurred during the follow-up provider investigation; current source and HTTP responses were checked independently.

Remembering the captured 502 preference (`src:fbb1669e08460c8e2421`) returned pending/degraded with `jev_fallback` / `budget_exhausted`; maintenance flushed zero notes. Impact: the note remains unscreened, with the original source and decision preserved in the Markdown bank.
