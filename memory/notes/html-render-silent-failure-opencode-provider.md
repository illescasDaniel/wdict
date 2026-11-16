---
name: html-render-silent-failure-opencode-provider
description: In T3 Code with opencode-provider models, html_render attachments never display; paste the inline_view.py --format md output
type: gotcha
date: 2026-11-13
status: current
supersedes: 
superseded_by: 
source: 
---
`html_render` returns success with an attachment id but shows no frame in that combination (path and inline-html variants both
tested, `html_preview` of the same document renders fine). Working replacement: run the SDD skill's `inline_view.py <page>
<tab> --format md` and paste the Markdown in chat, optionally opening the full page in the Review pane. Scope (the user): only
opencode-provider models in T3 Code; Claude Code and Codex render `html_render` correctly, so do not degrade their render
points.
