---
name: upstream-dictionary-data-gaps
description: Some pronunciation gaps are upstream data (quiso, many common headwords); the 'sing -> think alias' claim was wrong
type: gotcha
date: 2026-11-15
status: current
supersedes: 
superseded_by: 
source: 
---
Verified 2026-11-15 against the installed xxyzz 20261018 StarDict data: `quiso` has no `.idx` entry, only `.syn` aliases to
`querer`, whose only stored pronunciation is the infinitive [keˈɾeɾ]; the form's [ˈkiso] is not in the data. Many common
headwords (cinco, zarpa, llave, song, new, girl, world) have no pronunciation items at all. The fix belongs upstream.

Correction of an earlier survey note: there is no `sing` -> `think` alias. Only `thought` showed `think`'s cards first (a
ranking defect, see [offline-exact-headword-ranking](offline-exact-headword-ranking.md)); `sing`'s real defect was `shit`
cards arriving through the censored aliases `s****ing` and `s—ing`, which fold to "sing".
