---
name: launch-experiments-rejected
description: Startup experiments (prerender, window geometry, deferred metadata) showed no gain and were removed
type: decision
date: 2026-10-29
status: current
supersedes: 
superseded_by: 
source: 
---
Temporary launch experiments showed no demonstrated interactive-startup improvement, and the user saw no difference from the
startup placeholder, so all were removed. The independent fix for the dictionary metadata failure state (visible error and
retry without restart) was kept. Focus-call and animation-frame markers were only proxies; native visible readiness was never
measured. The user's earlier preference had been prerendering, initial window geometry, deferred metadata, focus on the Go
executable and setting aside a background process.

**How to apply:** do not retry these without real readiness measurements in the native window.
