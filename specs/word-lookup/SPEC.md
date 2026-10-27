# Word lookup

## Metadata

- **Feature:** A fast desktop window for English word and phrase definitions using an online dictionary and a two-level cache.
- **Wireframe:** [word lookup](../../wireframes/word-lookup.html); initial design approved 2026-10-26, revised to incorporate decisions for spec review.
- **Download wireframe:** [offline dictionary download](../../wireframes/dictionary-download.html); approved 2026-10-27.
- **System design:** [design](../../design/word-lookup/DESIGN.md); initial design approved 2026-10-26.
- **Design decisions:** agreed 2026-10-27, with scope adjusted in chat.
- **Spec approval:** approved 2026-10-27, with in-app dictionary download required in v1.

## Design decisions

### Success criteria

A user can look up an English word or multiword expression and see available English definitions, parts of speech and source attribution. The provider's coverage determines which entries exist. Successful results are cached in memory and on disk so repeated lookups avoid network requests. An in-app downloadable offline dictionary is required in v1; the user can download a curated dictionary without a separate StarDict installation. Spanish is desirable but not required.

### Failure handling

The window opens even without network access. Invalid input, no entry and unavailable provider are distinguishable. Errors retain any previous successful result and offer retry by submitting again. If an expired cached entry exists and the provider fails or returns unusable data, show the stale entry with its saved date and a stale label. A disk cache failure must not prevent lookup: continue with memory and provider, and preserve a valid online result even if persistence fails. No dictionary download is required to use the app. An interrupted or failed download is reported and can be retried; it must not replace a working installed dictionary or block lookup.

### Performance & resource budget

Targets, to be verified on a documented Linux reference machine: usable window within 2 seconds and local operations within 250 ms, excluding cold OS cache; online response time depends on the provider. The window remains responsive throughout network calls. Use a bounded network timeout of 10 seconds. Keep at most 128 entries in memory and 1,000 entries on disk, evicting least recently used entries within each tier. These are proposed defaults for review, not measured requirements. CPU and memory: no numeric limit agreed; record measured baseline during testing.

### Trust boundary

Search input, cached records and provider responses are untrusted. Store the cache in the user application cache directory, not a user-chosen path; validate cached records before use and ignore corrupted or incompatible records. Validate query as nonempty after trimming, at most 256 Unicode characters, without control characters; retain internal spaces. URL-encode the query as one path segment. Limit response body to 2 MiB and validate required entry content. Render remote text as text, never executable HTML. Only Go's provider adapter accesses the provider; the frontend cannot choose the destination. A user-initiated lookup may send the query online after a cache miss or expiry; the UI discloses this. No automatic requests at startup.

### Compatibility & rollout

Linux first; macOS and Windows are aspirational until tested separately. English only is required for v1. Wails/Go and SvelteKit are selected, with static frontend output and a provider interface. There is no previous app data or public API to migrate. The disk cache has an explicit schema version and is disposable: unsupported versions are ignored and replaced, without affecting other application data. No backfill is needed; rollback may discard the cache. Previously cached entries and entries in an installed dictionary work without network access. An uncached entry absent from the installed dictionary requires the provider. In-app dictionary download is required in v1, although the user need not download it to use the app.

## Triggers & routing

The window opens ready for input. Clicking **Look up** or pressing Enter starts a lookup. Success, invalid input, no entry or failure appears in the same window. A later search supersedes any prior pending search.

## Visual & UI rules

Match the [wireframe](../../wireframes/word-lookup.html): single input, English label, lookup button and a notice that a cache miss may send the query to the provider. Show cached and stale results with their source and saved date. Show searching, result, invalid input, no entry and unavailable states. Include headword, available parts of speech, definitions and provider-supplied source attribution. Preserve the previous successful result when reporting an error. Support keyboard use, visible focus and high-DPI scaling. Review-only notes do not ship.

## Interfaces & contracts

The frontend invokes the Go lookup use case through Wails with query and language `en`. Results are `found`, `invalid_query`, `not_found`, or `source_unavailable`. The Go adapter calls the Free Dictionary API documented in the [design](../../design/word-lookup/DESIGN.md), maps valid definitions to the domain, and returns source attribution from the response. The lookup checks the installed dictionary for the selected language first. If absent, unavailable, invalid or missing the entry, it checks memory then disk, and queries the provider only on a miss or when an entry is older than 7 days. A successful online result replaces both cache tiers; not-found responses and errors are not cached. Stale entries are retained for fallback on source failure. The lookup may update cache state but does not alter source dictionary data. Provider terms and real-world phrase coverage must be verified before release; inability to satisfy them requires revisiting this spec.

## Data & migration

Store successful normalized entries and their attribution in a versioned local cache, keyed by provider ID, language and trimmed query with Unicode NFC normalization and case folding. Preserve internal spaces and distinguish different languages and providers. Each record has key, validated entry payload, fetched-at timestamp and last-used timestamp. Disk writes must be atomic or transactional; corrupted entries are ignored and must not destroy unrelated entries. No query history or backfill is required. The cache is disposable; future schema changes can rebuild it. Provider and cache adapters remain separate from the use case.

## In-app offline dictionary download

For v1, offer at least one curated English-to-English StarDict package through an in-app **Download** action; Spanish-to-Spanish remains desirable when a suitable dataset is available. Display dictionary name, language, source, download size, license and installed/not-installed state before download. Downloads start only on user action, show progress, success or an actionable failure, and permit retry. The window and normal lookup remain usable during download. No system package, server or separate StarDict application is required.

Go handles a fixed, reviewed HTTPS package source; the frontend selects a catalog identifier, not an arbitrary URL or filesystem path. Download to temporary app-owned storage, enforce bounded archive and extracted sizes, verify a published checksum from trusted catalog metadata, reject unsafe archive paths, links and invalid or unsupported files, then install atomically into a per-user application data directory. A failed or interrupted installation cleans up temporary files and leaves any previously working dictionary intact. Installed dictionary data persists across restarts and is separate from the disposable response cache. Open lazily or asynchronously so a missing, corrupt or large dictionary does not prevent startup; do not automatically download or send queries at startup.

Use the pure-Go `github.com/ianlewis/go-stardict` reader as the selected candidate, subject to a compatibility, performance and license review. The selected dictionary's identity, source, version, content rights, attribution, size, update mechanism and phrase coverage must be verified before the feature ships; none is yet selected. Render dictionary content as untrusted text, and preserve source attribution. If a valid offline entry is found, show it and make no online request. Otherwise use the existing cache and online fallback; an offline read failure is not an online not-found result. The full offline dictionary is not subject to the response-cache TTL or LRU limits. Imported user files, background updates and a dictionary marketplace are out of scope.

### Download acceptance criteria

### Scenario: Download and install

**Given** a curated dictionary is available and not installed **When** the user selects Download and the package passes integrity and format checks **Then** progress and completion are shown and the dictionary is available after restart without a separate system installation.

### Scenario: Offline hit avoids network

**Given** a downloaded English dictionary contains the requested word or phrase **When** the user looks it up without network access **Then** its definition and source are displayed and no online request is made.

### Scenario: Offline miss or unavailable source falls back

**Given** the dictionary is absent, unreadable or has no entry **When** the user searches **Then** the app checks the response cache and online provider under the normal cache rules.

### Scenario: Download failure or invalid archive

**Given** a download fails, is interrupted, fails verification or contains unsafe or invalid files **When** installation ends **Then** the app reports an actionable failure, offers retry, keeps any previously installed dictionary usable and continues normal lookup.

### Scenario: Download is user initiated

**Given** the app has started and the user has not selected Download **When** they use the app **Then** no dictionary package is downloaded automatically, and no lookup is blocked by a download.

## Acceptance criteria

### Scenario: English word or phrase found

**Given** no installed dictionary entry and an available provider with an entry for the submitted English word or phrase **When** the user selects Look up **Then** the UI shows the returned headword, definitions, available parts of speech and source attribution.

### Scenario: Invalid input (F1)

**Given** an empty, whitespace-only, overlong or control-character query **When** the user submits it **Then** the UI requests valid input and no provider request is made.

### Scenario: No entry (F2)

**Given** no installed dictionary or fresh cached entry matches and a provider reports an English entry absent **When** the user submits that query **Then** the UI shows no entry found and retains the last successful result.

### Scenario: Source unavailable (F3)

**Given** a network failure, timeout, or non-success provider status other than not found **When** the user submits a valid query **Then** the UI reports the source unavailable, retains the last successful result and permits retry.

### Scenario: Malformed provider response (F4)

**Given** an oversized, malformed or unusable provider response **When** a lookup completes **Then** the UI reports the source unavailable, does not render untrusted markup and permits retry.

### Scenario: User controls network access

**Given** the app has just opened **When** the user has not submitted a lookup **Then** no provider request is made and the UI explains that a cache miss or expiry may send the query online.

### Scenario: Safe query encoding

**Given** a valid query containing spaces, Unicode or URL delimiters **When** it is submitted **Then** it is encoded as a single query path segment to the configured provider and cannot change the destination.

### Scenario: Superseded lookup

**Given** an earlier lookup is pending **When** a later lookup finishes first **Then** the earlier response cannot replace the later result or status.

### Scenario: Fresh memory cache hit

**Given** no installed dictionary entry matches and a valid matching entry in memory fetched less than 7 days ago **When** the user searches for it **Then** the cached entry and its source and saved date are shown and no provider request is made.

### Scenario: Fresh disk cache hit after restart

**Given** no installed dictionary entry matches and a valid matching unexpired disk entry and an empty memory cache **When** the user searches after restarting the app **Then** the entry is shown, promoted to memory, and no provider request is made.

### Scenario: Expired entry refresh

**Given** no installed dictionary entry matches and an entry at least 7 days old **When** the user searches and the provider succeeds **Then** the fresh result is shown and replaces the cached entry.

### Scenario: Stale fallback (F5)

**Given** no installed dictionary entry matches and an expired valid cached entry **When** the provider fails or returns an unusable response **Then** the cached result is displayed as stale with its saved date and source.

### Scenario: Disk cache unavailable or corrupt (F6)

**Given** the disk cache cannot be read or contains an invalid record **When** the user searches **Then** that record is ignored, memory and provider lookup remain available, and unrelated valid records are preserved.

### Scenario: Cache write fails (F7)

**Given** a successful provider response and a failed disk cache write **When** lookup completes **Then** the result is displayed and retained in memory without damaging an existing disk entry.

### Scenario: Failed lookup is not cached

**Given** a provider returns not found or an error **When** the user retries the same uncached query **Then** a new provider request is allowed.

### Scenario: Cache isolation and limits

**Given** entries with different provider IDs or languages and more entries than a tier limit **When** lookup and eviction occur **Then** keys do not collide and the least recently used entries are evicted within the affected tier.

### Scenario: Responsive and accessible window

**Given** a lookup is pending **When** the user operates the window using keyboard or a high-DPI display **Then** the window remains responsive, focus is visible, and controls and text remain usable.

## Out of scope

- Spanish support is desirable but not required for release.
- Inflected form resolution, spelling suggestions, and comprehensive morphology are deferred and recorded in project memory.
- Accounts, history, translation and a public application API.
