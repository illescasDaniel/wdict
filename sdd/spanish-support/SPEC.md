# Automatic English and Spanish lookup

## Metadata

- **Feature:** Search English and Spanish for every query, online or using optional downloadable offline dictionaries.
- **Wireframe:** [Spanish support](../../wireframes/spanish-support.html), approved 2026-10-29 in chat.
- **System design:** [Spanish support](../../design/spanish-support/DESIGN.md), approved 2026-10-29 in chat (`src:58170218c984fa66a08b`).
- **Design decisions:** [proposals](../../design/spanish-support/DECISIONS.md), approved 2026-10-29 in chat (`src:e86c89718081e601095e`).
- **Spec approved:** 2026-10-29 in chat (`src:6e99afe45172b4ac8823`).
- **Legacy cached attribution amendment approved:** 2026-10-29 in chat (`src:517058a14965c942a353`).
- **Related:** [word lookup spec](../word-lookup/SPEC.md); this spec supersedes its English-only scope and single-result assumptions for bilingual lookup.

## Design decisions

### Success criteria

Every valid search checks both languages without a selector. Results are displayed independently as each finishes, grouped and identified by language, including both for shared words. A Spanish dictionary downloaded in the app answers offline after restart. Existing English installs and caches remain usable without migration. Results link to their actual source and license when available; legacy cached entries missing a valid link show an attribution-unavailable notice instead of inventing one.

### Failure handling

For each language independently use installed offline data, fresh memory/disk cache, then online lookup, with stale fallback only if online lookup fails. A failure in one language never suppresses the other's result. Show a retry for unavailable languages; if neither language yields a result, show not found only when both report not found, otherwise unavailable. HTTP 502 from either online lookup endpoint is a silent temporary failure: use a stale entry if available; otherwise do not display a per-language error, do not cache it as not found, and show “No definitions available right now” if no language has a result, at least one silently failed, and no other visible failure occurred. Other failures retain their existing visible handling. This is a user-requested exception on 2026-10-29 (`src:fbb1669e08460c8e2421`). The latest search supersedes earlier pending results. Failed or interrupted downloads preserve both existing dictionaries, are reported and can be retried; unknown IDs cannot start a download.

### Performance & resource budget

Run language lookups concurrently and deliver each independently. Each online request has a 10-second timeout and a 2 MiB response limit. Retain 365-day cache freshness, 128 memory records and 1,000 disk records. Use a 20-minute download timeout, exact catalog size and SHA-256 check, and at most 1 GiB extracted archive data. No startup or lookup latency guarantee is asserted.

### Trust boundary

Queries, provider responses, URLs within responses and archives are untrusted. Validate queries before I/O, use fixed HTTPS hosts, encode titles and do not allow the UI to choose a URL or filesystem path. Extract definitions to plain text, never render provider HTML, and validate displayed links. Only catalog IDs select packages; verify pinned size and checksum, reject unsafe paths, links, unsupported files, excessive expansion and invalid dictionaries before atomic installation. Ambiguous or unparseable Spanish HTML is unavailable rather than a wrong-language result.

### Compatibility & rollout

Retain English dictionary path and provider/cache identities; store Spanish independently. Preserve current data through installation, recovery and rollback. Older app versions continue to use English data but do not expose Spanish. No flag, database migration or automatic download is needed. Publishing a release is outside this approval.

## Triggers & routing

Enter or the lookup button starts two language lookups for the same query. Input must be valid UTF-8, nonempty after trimming, at most 256 Unicode characters and free of control characters; retain internal spaces. No lookup or download occurs on startup. Search can continue during an installation. A new search cancels or invalidates all results and retries belonging to the prior search.

## Visual & UI rules

Follow the approved wireframe: no language selector; input, search button and notice that queries may be sent online. Display English and Spanish groups in that order when available, with headword, meanings, source, license, origin and stale indicator where applicable. Show a pending indicator for each still-running lookup without delaying a completed group. Show a per-language unavailable message and retry beside any successful group. When both settle, show no definitions only if both are not found; otherwise show an unavailable state if there are no results and at least one failure. Invalid queries display a validation message. Preserve the previous successful result when reporting a total error, consistent with the existing UI, while ensuring it is clearly associated with its original query. Do not ship review-only annotations.

For a legacy cached entry missing a valid source or license link, keep its definitions available, render any valid link that is present, and show “Source or license information unavailable for this cached entry.” Apply the same rule when displaying a previous result after an error. Do not manufacture attribution or force an online refresh solely because a cached link is missing.

The dictionaries screen lists English and Spanish separately, each with name, language, source, license, size and installed state. User-initiated download shows that dictionary's progress, checking/installing, completion or actionable error and retry; metadata loading has retry. Source and license links appear in the app and attribution in README.

## Interfaces & contracts

The frontend starts `Lookup(query)` and receives independent English and Spanish outcomes, each correlated to the initiating query and identifying language, status (`found`, `not_found`, `source_unavailable`, `silent_unavailable` or `invalid_query`), entries, origin, fetched time and stale flag as applicable. The delivery mechanism must support a completed language reaching the UI before the other completes; all callbacks/events must be ignored after supersession. Retry a failed language without discarding a successful group or unnecessarily repeating its request.

For each language: offline hit ends that language's lookup without its online request; otherwise check memory then disk using provider ID, language and Unicode NFC/case-folded query. A fresh hit avoids online; otherwise request the corresponding fixed provider. Cache successful results only. A failed provider uses a stale record if present; a genuine not-found does not use stale data. The English provider remains Free Dictionary API. Spanish uses the Spanish Wiktionary MediaWiki parse API and extracts only the Spanish language section, including its headword, available part of speech and definitions. Properly encode Unicode titles and follow supported redirects; no Spanish section is not found, while malformed, oversized, ambiguous, unusable or failed responses are unavailable. No provider markup is rendered as HTML.

`Dictionaries()` lists both records, `Download(id)` uses trusted catalog metadata, and progress includes dictionary ID. The Spanish package is [xxyzz's 20261018 es-es.tar.zst](https://github.com/xxyzz/wiktionary_stardict/releases/tag/20261018), publisher SHA-256 `0319285758d5d40022c3cc00c0e414b2d60883a1c1ae0bd53378789443c48676` ([release asset listing](https://github.com/xxyzz/wiktionary_stardict/releases/expanded_assets/20261018)). Before implementation is considered complete, verify exact bytes and archive format, metadata and content license, test sample lookups, and record the results; do not ship an unverified size or license. The verified archive is 23,446,485 bytes; its metadata identifies CC BY-SA 4.0 and a Spanish Wiktionary snapshot. Check source license and attribution requirements for both online and offline data; link to the entry, license and compiler as applicable and disclose extraction/format changes.

## Data & migration

Keep the current English `20261018/en-en.ifo` and English cache keys. Spanish gets a distinct dictionary ID, installation location, reader and provider cache ID; records retain language and provider so identical strings do not collide. No existing files are relocated or overwritten. Recover temporary installation data safely; an interrupted replacement keeps the last working dictionary. Cache remains disposable and is separate from installed data. Rollback leaves Spanish files on disk and restores normal English operation.

## Acceptance criteria

### Scenario: Both languages return results

**Given** both languages have definitions and valid attribution links for a valid query **When** the user searches **Then** both language-labelled groups are shown with source and license links, regardless of completion order.

### Scenario: Results arrive independently

**Given** one language completes while the other is pending **When** its result arrives **Then** it is visible immediately and the other remains pending without blocking the interface.

### Scenario: Invalid input (F1)

**Given** an empty, invalid UTF-8, overlong or control-character query **When** submitted **Then** validation is shown and no dictionary or network request is made.

### Scenario: One provider unavailable (F2)

**Given** one language is unavailable and the other succeeds **When** the search completes **Then** the successful definitions remain visible alongside a retry for the failed language.

### Scenario: Stale fallback (F2)

**Given** a language's offline data misses, its cache entry is stale and its online source fails **When** searched **Then** its saved definition and stale indicator appear, independently of the other language.

### Scenario: Not found and unavailable (F2, F3)

**Given** neither language yields a definition **When** one reports not found and the other unavailable **Then** show unavailable rather than claiming no definitions exist.

### Scenario: Neither language has an entry (F3)

**Given** both language providers report no match and no offline or fresh cache hit exists **When** searched **Then** show no definitions found and do not cache the misses.

### Scenario: Source response is malformed or unsafe (F2)

**Given** a provider returns oversized, malformed, ambiguous or hostile content **When** parsed **Then** report that language unavailable, retain any other successful result, and never execute or render its markup.

### Scenario: Spanish section and accented title

**Given** a Spanish Wiktionary page with a Spanish section and other language sections for an accented query **When** looked up **Then** only Spanish definitions are extracted, the title is correctly encoded and attribution links to the source.

### Scenario: Redirect and missing Spanish section (F3)

**Given** a supported redirect to a Spanish entry or a page without a Spanish section **When** looked up **Then** return the redirected Spanish definitions or not found respectively, never another language's definitions.

### Scenario: Offline hit and fallback

**Given** an installed Spanish dictionary has a matching entry and English has no offline match **When** searched **Then** Spanish uses offline data without a Spanish online request while English independently follows cache and online rules.

### Scenario: Cached identities and refresh

**Given** a query exists in both language caches and a Spanish entry has expired **When** searched **Then** entries remain isolated by provider and language and only the stale language is refreshed unless its offline dictionary answers.

### Scenario: Superseded search

**Given** a lookup is pending **When** a new query is submitted **Then** late results and retry responses for the prior query cannot replace or mix with current results.

### Scenario: User downloads Spanish

**Given** the Spanish dictionary is not installed **When** the user downloads the verified catalog package **Then** progress and completion are shown and a matching entry is available offline after restart without affecting English.

### Scenario: Invalid catalog ID or metadata error (F4)

**Given** an unknown ID or unavailable dictionary metadata **When** download is requested or the screen opens **Then** no untrusted package is downloaded, an error is shown and metadata can be retried.

### Scenario: Failed installation (F5)

**Given** transfer interruption, insufficient space, checksum mismatch, malformed dictionary or unsafe archive **When** installation fails **Then** temporary files are cleaned or recovered, existing working dictionaries remain usable and retry is offered.

### Scenario: No implicit download

**Given** the user has not requested a download **When** the app starts or searches **Then** no dictionary package is downloaded automatically.

### Scenario: Existing English data and rollback

**Given** an existing English install and cache **When** this feature is installed or rolled back **Then** those English data remain readable and no migration or deletion is required.

### Scenario: Legacy cached entry lacks attribution links

**Given** a fresh or stale legacy English cache entry lacks a valid source or license link **When** its definitions are displayed, including as a previous result after an error **Then** the definitions remain available, any valid link is shown, and an attribution-unavailable notice is displayed without inventing a link or requiring an online refresh solely to obtain one.

## Out of scope

Language selection, translation between languages, UI localization, automatic package updates, imported dictionaries, spelling suggestions, general inflection resolution, release publication and claims of untested platform support.

### Scenario: HTTP 502 is silent

**Given** an online dictionary endpoint returns HTTP 502 **When** a lookup completes **Then** use any stale entry, otherwise do not display an error for that language or cache the failure as a missing entry; if no result or other visible failure is present, show a neutral availability message.
