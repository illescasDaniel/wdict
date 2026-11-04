# Agemem screening retry — 2026-10-29

Calling `project_memory_remember` for the Spanish architecture approval (`src:dc4bf98b737711e4b089`) returned a pending, unscreened note with `DeciderUnavailable` at the `qsiv` stage. Impact: the approval was initially hidden from durable recall. Calling `project_memory_maintain` subsequently reported one note flushed; consolidation still reported degraded. The original captured evidence and the Markdown decision record remain available.

Remembering the v0.4.0 release and install (`src:c306659d288af475a134`) also returned pending and unscreened, with `jev_fallback` / `budget_exhausted`. Impact: durable recall cannot yet rely on that note; the captured event and Markdown bank retain the verified outcome.
