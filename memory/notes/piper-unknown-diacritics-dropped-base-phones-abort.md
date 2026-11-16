---
name: piper-unknown-diacritics-dropped-base-phones-abort
description: The Piper script drops missing Unicode-Mn diacritics but still aborts on unknown base phones
type: decision
date: 2026-11-14
status: current
supersedes: 
superseded_by: 
source: 
---
Spanish lenition notation `β̞ ð̞ ɣ̞` (U+031E; sábado, luego, nada, cada, todo, usted, debe, ave, lava, sobre, adverso) and the
retraction mark U+0320 are in no voice map, so the inventory check raised and the chain fell back to robotic eSpeak. Fix: the
synthesis script drops missing Mn diacritics (the base phone carries the sound). Unknown base phones still abort: that guard
prevents `ŋrike`-style corruption and must stay. Verified map facts: nasal tilde U+0303 is in all maps (kept, Enrique path),
dental U+032A in sharvard, aspiration ʰ and parentheses in alba, ɐ everywhere. A 105-item survey of both dictionaries then
resolved 100% through Piper with no eSpeak fallbacks.
