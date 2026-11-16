---
name: piper-runtime-budget-250mb
description: Piper runtime disk budget is at most 250 MB installed; the measured piper-tts 1.8.0 runtime is 210 MB
type: decision
date: 2026-11-14
status: current
supersedes: 
superseded_by: 
source: 
---
Amendment G1 (user "g1", 2026-11-14): the spike's ~120 MB estimate was too low. A real pip install of piper-tts 1.8.0 measures
210 MB (Python 3.14, onnxruntime 1.32.0, numpy 2.5.4), so the budget is 250 MB and the README says about 210 MB. G4: the
blocked-voice copy is translated (`core.VoicePack.Reason` python-missing / runtime-first with localized frontend strings);
that was implemented 2026-11-15.
