---
name: ipa-not-borrowed-across-capitalization
description: Never fill a missing IPA for one capitalization from another (Car vs car, Polish vs polish)
type: decision
date: 2026-11-15
status: current
supersedes: 
superseded_by: 
source: 
---
Installed `Car` records have no IPA while `car` has one, and `Polish`/`polish` are pronounced differently, so borrowing a
transcription across capitalizations is unsafe. The approved design shows an explicit "source IPA absent" state and a
separately labeled synthesized spelling reading instead (`sdd/result-ranking.html`).
