# Agemem recall budget — 2026-10-28

Reproduction: call `project_memory_recall` for the cache persistence investigation. The configured cascade returned `decision_status: degraded`, with `jev_fallback` / `budget_exhausted`. Impact: no screened notes were returned; bank and source excerpts remained available and were checked against the code. No provider settings were changed.

The same `jev_fallback` / `budget_exhausted` degraded recall recurred when starting the v0.2.0 release task; current repository files were used to verify the task.
