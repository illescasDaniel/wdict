---
name: dependencies-project-local
description: Dependencies come from Go modules and project-local npm with no global installs; Task runs as a pinned Go tool
type: preference
date: 2026-10-27
status: current
supersedes: 
superseded_by: 
source: 
---
The user asked for package managers for Go and the frontend and for no global installations. Go modules and project-local
npm are used, and Task is a pinned Go tool, so commands are `go tool task setup|build|test`. Task's shared module
dependencies may also move application dependencies.

**How to apply:** do not tell the user to install tools globally; the quality gate is `go tool task test`.
