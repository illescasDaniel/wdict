---
name: spanish-senses-drop-dt-labels
description: spanishMeanings drops dt labels and nested lists, so labels like 'Rugby' or 'Uso: coloquial' do not reach the UI
type: gotcha
date: 2026-11-15
status: current
supersedes: 
superseded_by: 
source: 
---
Found while investigating the Spanish `touch` entry (installed entry and live Wikcionario both have six senses and
[ˈt̪owt͡ʃ]). The installed raw HTML includes `Rugby` in a `dt` and a nested `Uso: coloquial`, but `spanishMeanings`
(`internal/adapters/spanish.go`) drops `dt` labels and nested `ul`/`ol`/`sup`. Which other live labels are missing from the raw
data was not checked exhaustively. Recovering them is outside the approved spec S7 (preserve current presentation strings).
