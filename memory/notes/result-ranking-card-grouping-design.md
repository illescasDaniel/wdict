---
name: result-ranking-card-grouping-design
description: Result ranking: case-aware tiers, source-card grouping, labeled spelling reading; Design, Spec and Architecture approved
type: decision
date: 2026-11-15
status: current
supersedes: 
superseded_by: 
source: 
---
Approved in chat 2026-11-15 on `sdd/result-ranking.html`: exact-case ranking (ordinary `car` ahead of `Car`) with
conservative lexical tie-breaking, query-aware ordering on cache replay, grouping of exact-spelling/source cards that keeps
source records and pronunciation boundaries, and an explicit absent-IPA state with a separately labeled synthesized spelling
reading (never borrowed IPA: [ipa-not-borrowed-across-capitalization](ipa-not-borrowed-across-capitalization.md)). Spec S1-S16
and the Architecture were approved; automatic QA and grading were enabled. Implemented; the grader said NEEDS WORK for resource
evidence and stale traceability, both fixed afterwards without a regrade. Committed in `a2cbfcc`, before v0.8.3. Testing stayed off screen.
