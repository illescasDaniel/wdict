---
name: opencode-needs-its-own-mcp-config
description: OpenCode ignores .mcp.json; project MCP servers must be listed in opencode.json
type: gotcha
date: 2026-11-13
status: current
supersedes: 
superseded_by: 
source: 
---
An OpenCode session started without the project servers because `.mcp.json` is read only by Claude and Codex. `opencode.json`
uses `mcp.servers.<name>` entries with `type: "local"`, a `command` array and `environment`. The tool catalog changes only
after the OpenCode server restarts, and `opencode mcp list` can still say "No MCP servers configured" right after writing the
file, so verify via the tool catalog.

**How to apply:** `opencode.json` mirrors `.mcp.json` (webnav, codenav-go); keep them in sync. webnav `edit` rejects `.svelte`
files ("write tools only edit .cjs/.css/...") so use the standard Edit tool for Svelte components.
