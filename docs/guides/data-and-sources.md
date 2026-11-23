---
id: DOC-025
title: Dictionary data and sources
summary: Offline and online dictionary sources, download sizes, licenses, attribution and data storage
type: reference
status: current
updated: 2026-11-21
verified:
covers: [internal/adapters/catalog.go, internal/adapters/install.go, internal/adapters/offline.go, internal/adapters/spanish.go, internal/adapters/french.go, internal/adapters/provider.go]
related: []
supersedes:
superseded_by:
---

# Dictionary data and sources

The app offers the pinned [20261018 English Wiktionary StarDict release](https://github.com/xxyzz/wiktionary_stardict/releases/tag/20261018), compiled by [xxyzz](https://xxyzz.github.io/wiktionary_stardict/) from [Wiktionary](https://en.wiktionary.org/). Its package metadata identifies the snapshot as [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/); credit belongs to Wiktionary contributors and xxyzz. wdict extracts definitions and changes their formatting. The download is approximately 142 MB, is user initiated, and is verified against the publisher's SHA-256 checksum before installation.

Spanish offline definitions use the pinned [20261018 Spanish Wiktionary StarDict release](https://github.com/xxyzz/wiktionary_stardict/releases/tag/20261018), compiled by [xxyzz](https://xxyzz.github.io/wiktionary_stardict/) from [Wikcionario](https://es.wiktionary.org/); package metadata identifies CC BY-SA 4.0. The download is about 23 MB and is verified using the publisher’s SHA-256 checksum. Credit belongs to Wiktionary contributors and xxyzz; wdict extracts definitions and changes their formatting.

French offline definitions use the pinned [20261018 French Wiktionary StarDict release](https://github.com/xxyzz/wiktionary_stardict/releases/tag/20261018), compiled by [xxyzz](https://xxyzz.github.io/wiktionary_stardict/) from [Wiktionnaire](https://fr.wiktionary.org/). Its package metadata identifies CC BY-SA 4.0. The user-initiated download is about 87 MB and is verified against the publisher’s SHA-256 checksum; credit belongs to Wiktionary contributors and xxyzz. wdict extracts definitions and changes their formatting.

English online lookups use [Free Dictionary API](https://dictionaryapi.dev/); the app displays each response's supplied source and license links, and sampled responses identified [CC BY-SA 3.0](https://creativecommons.org/licenses/by-sa/3.0/). Spanish online lookups use the [Spanish Wiktionary](https://es.wiktionary.org/) MediaWiki API, with attribution to Wiktionary contributors and links to the entry and to [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/), the site's license; wdict extracts definitions and changes their formatting. French online lookups use the [French Wiktionary](https://fr.wiktionary.org/) MediaWiki API, with attribution to Wiktionary contributors and links to the entry and to [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/); wdict extracts definitions and changes their formatting. Offline results link to their Wiktionary source and license, and the offline dictionary screen also links to its source and license. Wiktionary may contain material from other sources with separate requirements; see [Wikimedia's reuse terms](https://foundation.wikimedia.org/wiki/Policy:Terms_of_Use).

Offline dictionary data is stored separately from the disposable response cache, in the user's data directory. When a lookup misses, installed dictionaries can provide possible base forms and spelling suggestions for words and phrases. Inflection rules are limited, and suggestions may be unavailable if a search exceeds its time limit.
