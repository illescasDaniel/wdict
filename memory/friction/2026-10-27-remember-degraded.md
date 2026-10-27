# 2026-10-27 — Agemem remember degraded

Reproduction: capture the user's v1 scope with `project_memory_capture`, then call `project_memory_remember` using source `src:78ae2237acccde7265bf`.

Result: `pending: true`, `degraded: true`, reason `jev unavailable; stored unscreened and hidden`. Impact: the durable note is not a successfully screened memory. Scope and deferred work are recorded in the Markdown bank; retry maintenance when the configured provider is available.

## Diagnosis

The message `jev unavailable` is hard-coded for any `DeciderUnavailable`, including QSIV and cascade failures, so it does not identify the failed provider. The active MCP launcher selects the online cascade. In the Codex workspace sandbox, doctor cannot see CUDA and QSIV startup fails; outside the sandbox, doctor reports the RTX 4070 GPU and one synthetic local QSIV inference succeeded. Sandbox GPU access is therefore a plausible cause, but the original call did not retain the underlying exception and its exact cause is unconfirmed. Jev credential presence was checked, not Jev connectivity; no hosted call was made. The captured note remains pending until screened successfully.

## Follow-up probe

A local-only MCP test in a disposable project, using QSIV with GPU access, successfully captured synthetic evidence, screened a durable note, and recalled both. An offline MCP recall against wdict also ran, but returned degraded; the cascade cannot use its hosted fallback offline, and the exact decision failure was not exposed. QSIV starts lazily when a decision is required, and the MCP server caches the worker. The earlier `jev unavailable` degraded write means a decision call raised `DeciderUnavailable` after provider initialization; a QSIV startup failure alone would surface as an MCP error instead. The specific original cause remains unknown: local inference or Jev fallback are possible. No hosted request was made in these probes.

## Online and diagnostics follow-up

A disposable-project MCP test with synthetic data succeeded with the online QSIV-to-Jev cascade: capture, screened remember, and recall all passed, with successful hosted fallback requests. A synthetic invalid-credential test returned `jev_fallback` / `authentication`, and an offline cascade test returned `jev_fallback` / `offline`; both stored the note pending and hidden. Agemem now returns stage/cause diagnostics for future decision failures. These tests do not establish the cause of the original failure; the original note remains pending.

## Live recall and improved diagnostics

On 2026-10-27 the user authorized an online MCP recall of wdict. The configured cascade completed; QSIV ran locally and Jev returned HTTP 200 for two fallback calls. The recall returned zero screened notes; the original pending note has not been retried. A synthetic invalid-credential test confirmed MCP diagnostics report `jev_fallback`, `authentication`, and HTTP 401. QSIV worker exceptions now emit tracebacks to client stderr logs; these logs may contain private runtime data. The precise cause of the original failure remains unknown.

On 2026-10-27, remembering the in-app download choice with source `src:901c3d41cff158d1b9e4` again returned pending/degraded (jev fallback missing credentials). The event remains captured, but the note is unscreened and hidden; decision is documented in the Markdown bank.

Remembering the v1 requirement and spec approval (`src:d9ba2309b87de6de7a1c`) returned pending/degraded because the fallback provider lacks credentials; the note remains unscreened and hidden.

Remembering architecture and wireframe approvals and package-manager preference (`src:7451c617c3b69a94fb37`) remained pending/degraded because fallback credentials are missing; the captured source and Markdown bank preserve the event.

Remembering provider and packaging observations (`src:8047bd2799c531826128`) returned pending/degraded due to missing fallback credentials. The event was captured but the note remains unscreened and hidden.
