---
name: sdd-html-feature-workflow
description: Features go through the sdd-html-feature skill (one page sdd/<feature>.html); the loadout must run with --without sdd
type: decision
date: 2026-11-11
status: current
supersedes: 
superseded_by: 
source: 
---
User decision: replace `sdd-feature` (`specs/`, `wireframes/`, `design/`) with the globally installed `sdd-html-feature`
skill, one page per feature in `sdd/`. The old directories stay as history. `AGENTS.md` carries a hand-written SDD section
without loadout markers.

**Why:** the loadout would reinstall `sdd-feature` and its Phase Gate block otherwise (it did so again on the 0.7.0 upgrade
until `--without sdd` was passed).
**How to apply:** always rerun `agentic-loadout` with `--without sdd`. The review gates (design, spec, architecture) need
separate in-chat approvals.
