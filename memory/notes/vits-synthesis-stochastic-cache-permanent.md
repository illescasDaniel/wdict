---
name: vits-synthesis-stochastic-cache-permanent
description: Piper VITS output is stochastic per run and cached for good; a garbled draw replays forever
type: gotcha
date: 2026-11-14
status: current
supersedes: 
superseded_by: 
source: 
---
Single-run durations prove nothing (noise scale, block-quantized lengths). The old cached "eight" had a 13 ms transient onset
(healthy Piper audio has 0.15-0.26 s lead-in silence). Cache keys changed with stress injection, so the cache repopulated, but a
future garbled draw on a new key would also persist. The audibility guard (peak >= 1000) filters silent draws only.

**How to apply:** measure lead-in-to-first-audible over at least 3 runs and cross-correlate a variant against a baseline.
`phonemes_to_ids` has no stress decoration, so sequence position is everything (eSpeak puts stress after the onset, before the
vowel). A latent bug was fixed: the `piper_real` replay test compared a fresh synthesis (raw against normalized string, a
different cache key) rather than a replay.
