# Inflected forms and spelling suggestions

## Metadata

- **Feature:** Offer offline possible base forms and spelling corrections, including phrases, when an English or Spanish lookup misses.
- **Wireframe:** [lookup alternatives](../../wireframes/lookup-alternatives.html), approved 2026-11-06 (chat).
- **System design:** [DESIGN.md](../../design/lookup-alternatives/DESIGN.md), approved 2026-11-06 (`src:c34610a226566b0738a4`).
- **Design decisions confirmed:** 2026-11-06 (`src:de40e77c3089de7f6007`).
- **Spec approved:** 2026-11-06 (`src:665fce98e49af19f680d`).
- **Related:** [word lookup spec](../word-lookup/SPEC.md), [Spanish support spec](../spanish-support/SPEC.md).

## Partial amendment — approved 2026-11-06

The user approved this amendment after measurement exposed conflicting failure behavior and weak spelling ranking (`src:8a02e642d2cb6cd1bb7f`). Its initial dataset and resource placeholders were resolved by the later final and memory-limit amendments.

- **Missing dictionary:** after a confirmed miss, do not start an index scan if the relevant dictionary is absent. Show that suggestions require that offline dictionary; retain the lookup result. If installed data is unreadable, show suggestions unavailable.
- **Optional morphology data:** when that data is absent or corrupt, use bounded regular-form fallback for base forms and continue spelling and phrase searches against the installed dictionary. A failure to complete the required dictionary scan produces unavailable and no partial ranking. The exact fallback rules require a further measured amendment before implementation.
- **Suggestion limits:** use a two-second per-language deadline and a 64-Unicode-scalar query limit for suggestions. A longer otherwise valid query still gets its normal lookup, with suggestions unavailable; it is not an invalid lookup. A timeout produces no partial ranking. These are approved policy limits, not guaranteed completion times: the warm-cache prototype took 116–666 ms, with cold-disk and slower-machine behavior unverified. Existing lookup's 256-rune validation is unchanged.
- **Ranking quality:** replace the previous distance-then-alphabetical-only rule. Evaluate a curated bilingual set covering words, transpositions, accents and phrases. Candidate ranking must put `words` in the top five for `wrods`, `árbol` in the top five for `arbl`, and `look forward to` in the top five for `look foward to`, when the respective headwords exist. Ranking must be deterministic, and its chosen method and any source/license for frequency data require a further amendment after broader evaluation. A five-example prototype is not sufficient proof of overall quality.
- **Dataset and resources:** no release dataset is selected. Evaluate actual misses, correctness, source/attribution and packaging before selection. The current small probe suggests Spanish UniMorph may add coverage while English adds less; neither is proven reliable. This placeholder is superseded by the later final and memory-limit amendments.

### Proposed amendment scenarios

- **Given** a confirmed miss without the relevant installed dictionary **When** suggestions would start **Then** no index scan starts, the existing result remains and the UI explains suggestions need the dictionary.
- **Given** a confirmed miss with an installed dictionary and missing or corrupt optional morphology data **When** suggestions run **Then** bounded regular-form fallback and spelling search remain available.
- **Given** a valid 65–256-rune query **When** lookup completes **Then** the ordinary lookup remains valid while suggestions are unavailable without scanning.
- **Given** the expected headword exists **When** `wrods`, `arbl` or `look foward to` is searched in the appropriate language **Then** the corresponding correction appears among at most five suggestions when the scan completes within budget.

## Final amendment — approved 2026-11-06 (`src:6dce82b0c9d386831cec`)

This approved amendment completes the remaining choices for the initial release, superseding the original ranking, dataset and resource placeholders where they conflict. Evidence: [morphology evaluation](../../design/lookup-alternatives/morphology-evaluation.md), [ranking development](../../design/lookup-alternatives/ranking-evaluation.md), [independent holdout](../../design/lookup-alternatives/holdout-evaluation.md). The samples are small and do not establish general accuracy or cold-machine performance.

- **Morphology:** ship no UniMorph data initially. Existing dictionary synonyms already cover many forms; additional data has not been established as sufficiently reliable and redistributable. On a confirmed miss, generate nonrecursive candidate bases. English: remove `s`, `es`, `ed`, or `ing` if a nonempty stem remains; for `ed` and `ing`, also try stem plus `e`. Spanish: remove `s` or `es` if a nonempty stem remains. Deduplicate and return only bases that verify against installed headwords; label each possible. These are deliberately incomplete and can yield unrelated words. No separate morphology download or persistent index is needed. Further datasets remain a future amendment.
- **Spelling and phrases:** normalize for comparison using Unicode case folding and canonical decomposition, removing combining marks and punctuation, while preserving display forms. Compare single-word queries to single-word index entries and phrases to phrase entries using Unicode optimal-string-alignment distance at most two. Sort by distance, then prefer exact single adjacent transposition, then exact insertion/deletion of one repeated letter, then matching initial folded rune, then lowercase display forms, then deterministic lexical order. Remove exact duplicate display strings and show at most five. Synonym headwords are eligible; verification still applies. This ranking is an initial baseline, not a guarantee that the intended word is offered. The independent hand-made holdout produced the intended candidate in the top five on 20 approximately missing inputs; this is not a population accuracy estimate.
- **Resource limits:** keep the approved two-second per-language deadline and 64-rune suggestion query limit. Read at most 128 MiB of dictionary index data per language search, reject malformed or oversized records rather than exceed it, and create no persistent suggestion files. The original 64 MiB hard-cap requirement is superseded by the approved memory-limit amendment below. The read and time limits are caps, not measured performance guarantees; the warm prototype used about 13 MiB process peak RSS but its incremental application memory and cold-disk behavior are unknown. If a cap or deadline prevents a complete scan, show suggestions unavailable with no partial ranking. Exact lookup is not blocked or limited by these suggestion caps.
- **Release checks:** exercise each design failure branch, candidate verification, deterministic ordering, malformed data and cancellation; compare at least the recorded frozen typo fixtures including `wrods`, `arbl`, `perrro`, `gatto` and `look foward to`, and test the caps and independence of concurrent ordinary lookup. Measure first and repeat latency, memory and cancellation on the available reference machine, and disclose cold-disk and other-machine gaps. No release is justified if the implementation cannot meet its caps.

### Proposed final amendment scenarios

- **Given** a miss with no optional morphology data **When** the relevant installed index is readable **Then** the specified language rules and spelling search both run without network access.
- **Given** a rule proposes a string absent from the installed index **When** candidates are verified **Then** it is not shown as a base form.
- **Given** `perrro` or `gatto` misses and its intended Spanish headword is installed **When** the complete search finishes within the budget **Then** `perro` or `gato`, respectively, appears in the top five.
- **Given** an oversized or malformed index record, exhausted read/memory budget or a deadline **When** a scan cannot complete **Then** it stops safely, returns no partial ranking and leaves ordinary lookup usable.

## Memory-limit amendment — approved 2026-11-06 (`src:bb20bbd7cfc98225d3b0`)

Implementation revealed that a strict 64 MiB cap on *additional live Go heap* cannot be enforced reliably for one feature using the shared garbage-collected heap: concurrent runtime activity and garbage collection affect the measurement, and the current interfaces cannot account for all allocations. The following would replace that single hard-cap requirement; all other approved limits remain unchanged.

- Bound feature-owned retained state and input buffers, keep only the five best spelling candidates, and do not materialize the dictionary indexes in memory. Continue to enforce the two-second deadline, 64-rune query limit and 128 MiB index-read limit.
- Treat 64 MiB additional live heap as a **measurement target**, not a hard per-feature runtime limit. Measure concurrent searches using sampled Go heap on the reference machine; investigate and fix any observed increase above the target before release. The measurement is process-wide and sampling can miss peaks; it does not guarantee behavior on every system or under unrelated concurrent workloads.
- The corrected real-dictionary probe observed about 3 MB sampled shared-heap increase in three warm-cache runs. The English phrase completed twice and timed out once while Spanish ran concurrently; a timeout retains the existing result and shows suggestions unavailable, without partial ranking. Cold-disk and slower-machine performance remain unverified.

## Design decisions

### Success criteria

- On a language's confirmed miss, show possible base forms and up to five spelling candidates found in its installed dictionary. Correct multiword phrases when an appropriate phrase headword exists, for example `look foward to` → `look forward to`. Examples depend on installed data and are not claims about its coverage.
- Keep original outcomes visible until the user selects an alternative. Selecting it starts a fresh bilingual lookup. Preserve multiple valid base forms and do not claim an index hit proves a morphological relation.
- Evaluate UniMorph for each language first; measure actual coverage against the shipped dictionary snapshots, accuracy on regular, irregular and ambiguous examples, performance, size and distribution terms. If unsuitable or unreliable, use partial bounded regular-form rules for the affected language. Record the evidence and attribution.

### Failure handling

- Invalid input stops before alternatives. A provider failure retains the existing unavailable or silent-unavailable behavior; it is never labelled a confirmed miss. Missing or unreadable dictionary or suggestion index and a timed-out scan leave existing outcomes intact and show suggestions unavailable. An empty result retains the ordinary not-found state.
- A new search or language retry invalidates previous work for that search or language. Cancellation discards stale events. No partial top-N results are published after timeout. A later search or retry can try again.

### Performance & resource budget

- Alternatives run asynchronously and independently of exact lookup. Enforce a cancellable time limit and bounded query length, work and output. At most five spelling candidates per language. No additional network requests occur during lookup.
- **Superseded:** the final and memory-limit amendments above set the operative resource limits and measurement requirements.

### Trust boundary

- Queries, downloaded dictionaries and morphology/index data are untrusted. Validate query encoding and length before scans, parse data defensively within limits, render candidates as text, and only return candidates verified against the relevant installed dictionary. Never use candidate text as a filesystem path or executable command.
- Lookup may read local installed data; it may not send additional suggestion queries over the network. Preserve license, attribution, source revision and transformations for distributed data.

### Compatibility & rollout

- Preserve existing dictionaries, caches and their formats. Any morphology index is optional, separately replaceable, and must not make existing dictionaries unusable by older versions. Missing or removed optional data cannot lose user data. Rollback to an older app preserves existing dictionaries and caches.
- Release normally only after coverage, resource measurements and distribution/license review pass. The final amendment selects no additional morphology dataset or package for this release.

## Triggers & routing

The existing lookup starts independently for English and Spanish. Only a `not_found` outcome for a language starts its alternatives search and only if its offline dictionary is installed. `found`, invalid and unavailable outcomes do not start it. Results are associated with the current search ID and language retry generation. A candidate click submits its displayed headword through the existing bilingual lookup.

## Visual & UI rules

- Follow the approved wireframe: separate language sections, labelled possible base forms and spelling suggestions, searching, no results, unavailable and narrow-window states. Review-only notes do not ship.
- Do not alter or hide an exact result in the other language. Display candidate strings as text and provide keyboard-operable buttons. Use the selected interface language for labels and status text.
- Present suggestion loading and unavailable separately from the ordinary lookup. Candidate selection changes the query only as part of starting the new lookup.

## Interfaces & contracts

`Lookup(query)` and existing language outcomes remain compatible. A separate `lookup-alternatives` event contains `SearchID`, `Query`, `Language`, `Status` (`ready`, `unavailable`, `timed_out`), `BaseForms` and `Spelling`. A retry generation or equivalent identity prevents a prior attempt for the same search ID and language from replacing a newer attempt. `ready` with empty arrays means no candidates, not a statement that the query is invalid. `unavailable` and `timed_out` contain no partial ranking. The consumer ignores events for stale search IDs or attempts. Selecting a suggestion uses the existing lookup contract, not a new external endpoint.

The same installed index and query yield deterministic ordering. The original proposal ranked spelling candidates using bounded Unicode edit distance followed by deterministic tie-breaks; the approved final amendment above defines the ranking based on measured examples. Preserve display spelling and accents. The approved final and memory-limit amendments define normalization and numerical bounds.

## Data & migration

No changes to existing cache or dictionary formats are required. Any new optional derived index must carry a source revision, license and attribution, be independently replaceable, and not alter the original data. No additional derived index is included in the approved first release; a future one would require its own approved package and migration details. Rollback removes the new app's use of the optional index while leaving existing dictionary and cache data intact.

## Acceptance criteria

### Scenario: base form on a language miss

**Given** a confirmed English miss, an installed English dictionary and a supported inflected form whose possible base is present **When** alternatives complete **Then** the base is labelled as possible and can be selected without changing the original result first.

### Scenario: ambiguous base forms

**Given** a supported form with two valid candidate bases present in the relevant dictionary **When** alternatives complete **Then** both are retained as possible candidates rather than silently choosing one.

### Scenario: spelling correction for a phrase

**Given** an installed dictionary containing the phrase `look forward to` and a confirmed miss for `look foward to` **When** the bounded phrase search completes within budget **Then** the phrase is offered as a spelling suggestion.

### Scenario: language results remain independent

**Given** English finds a definition and Spanish misses **When** Spanish alternatives complete **Then** the English result stays visible and Spanish alternatives appear only in the Spanish section.

### Scenario: select an alternative

**Given** a displayed alternative **When** the user activates it **Then** a fresh bilingual lookup begins for that candidate and stale events from the previous search cannot change the new results.

### Scenario: invalid query (F1)

**Given** an empty, invalid UTF-8, control-containing or over-limit query **When** lookup is requested **Then** the existing invalid state is shown and no alternative index is scanned.

### Scenario: lookup source unavailable (F2)

**Given** a language's lookup ends in source unavailable or silent unavailable **When** its outcome is displayed **Then** existing failure behavior is preserved and no miss-based suggestion search is started for that language.

### Scenario: missing or corrupt index (F3)

**Given** a confirmed miss and its dictionary is absent, unreadable or corrupt **When** alternatives are requested **Then** suggestions are unavailable and existing outcomes remain intact. A subsequent search may retry.

### Scenario: work exceeds deadline (F4)

**Given** a confirmed miss and an alternatives scan exceeding its agreed deadline **When** it is cancelled **Then** no incomplete ranking is displayed and the ordinary lookup remains visible with suggestions unavailable.

### Scenario: stale search or retry (F5)

**Given** an alternatives search is in flight **When** a new search or same-language retry starts **Then** the old work is cancelled or discarded and any late event cannot replace the current state.

### Scenario: no candidates (F6)

**Given** a confirmed miss and no verified candidate **When** the scan finishes **Then** the original not-found state remains and no fabricated suggestion appears.

### Scenario: untrusted dictionary or index data

**Given** malformed or oversized index content or a candidate containing markup **When** it is processed **Then** work stays within agreed limits, invalid data is rejected safely, and any displayed candidate is text rather than executable markup.

### Scenario: candidate not present in dictionary

**Given** morphology data suggests a lemma absent from the corresponding installed dictionary **When** verification runs **Then** that lemma is not offered.

### Scenario: dataset not suitable

**Given** the dataset fails the agreed coverage, performance or license review for a language **When** the implementation approach is chosen **Then** documented partial regular-form rules are used for that language instead of shipping the unsuitable dataset.

### Scenario: optional index removed or rollback

**Given** existing dictionaries and caches and an optional morphology index **When** the index is removed or the app is rolled back **Then** existing data remains usable and no user data is lost.

## Out of scope

- Automatic correction or replacing the user's query without selection.
- Online suggestion services and network requests during candidate generation.
- Claims of complete inflection, spelling or phrase coverage; suggestions for phrases absent from the installed dictionary.
- Changes to dictionary definition content or existing cache retention.
