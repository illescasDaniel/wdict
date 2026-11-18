---
name: dictionary-app-direction
description: wdict is a Linux-first desktop dictionary built with Wails v3, Go and SvelteKit, English and Spanish
type: decision
date: 2026-10-27
status: current
supersedes: 
superseded_by: 
source: 
---
The user approved the initial design (chat, 2026-10-26) and then picked the stack (2026-10-27): a Linux-first desktop
dictionary for English and Spanish with same-language definitions, aiming for macOS and Windows later, built with Wails
(v3.0.0 in `go.mod`) and Go, and SvelteKit with the static adapter.

**Why:** Tkinter was unsuitable for the user's high-DPI display. Wails v3 had dropped its built-in SvelteKit starter, so
a pinned v3 starter was adapted to SvelteKit's static adapter.
**Rejected:** Tkinter.
**How to apply:** more detail is in `sdd/design/word-lookup/tech-stack-and-decisions.md`. macOS and Windows are not tested; do
not claim support for them.
