# Lookup and ranking
Updated: 2026-11-16. Related: [pronunciation](pronunciation.md), [releases-and-packaging](releases-and-packaging.md)

## Now
- Each search checks English and Spanish with no selector ([bilingual-lookup-without-selector](../notes/bilingual-lookup-without-selector.md)); offline data is the xxyzz StarDict release scanned directly ([offline-dictionary-xxyzz-stardict](../notes/offline-dictionary-xxyzz-stardict.md)); online results stay fresh 365 days with stale fallback ([cache-freshness-365-days](../notes/cache-freshness-365-days.md), [online-failures-keep-stale-cache](../notes/online-failures-keep-stale-cache.md), [legacy-cache-attribution-notice](../notes/legacy-cache-attribution-notice.md)).
- Offline results: each index entry once ([offline-duplicate-results-aliases](../notes/offline-duplicate-results-aliases.md)), ranked exact, folded, alias ([offline-exact-headword-ranking](../notes/offline-exact-headword-ranking.md)), then case-aware grouping ([result-ranking-card-grouping-design](../notes/result-ranking-card-grouping-design.md)); senses grouped in the frontend ([sense-grouping-presentation-only](../notes/sense-grouping-presentation-only.md)).
- Alternatives after a miss ([lookup-alternatives-design-choices](../notes/lookup-alternatives-design-choices.md), [alternatives-heap-target-not-hard-cap](../notes/alternatives-heap-target-not-hard-cap.md), [alternatives-distance-cpu-hotspot](../notes/alternatives-distance-cpu-hotspot.md)).
- Known gaps: Spanish labels dropped ([spanish-senses-drop-dt-labels](../notes/spanish-senses-drop-dt-labels.md)); upstream data ([upstream-dictionary-data-gaps](../notes/upstream-dictionary-data-gaps.md)).

## How it got here
- 2026-10-27 v1 scope draft ([v1-scope-offline-and-cache-optional](../notes/v1-scope-offline-and-cache-optional.md)) was reversed the same day ([v1-offline-download-and-cache-required](../notes/v1-offline-download-and-cache-required.md)).
- 2026-11-13 duplicate cards from repeated aliases; 2026-11-15 ranking and grouping.
