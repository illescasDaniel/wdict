# Lookup alternatives feasibility probe — 2026-11-06

This is a small feasibility probe, not a benchmark of a production suggestion implementation or a comprehensive linguistic evaluation. Tests ran on Linux 7.3.1-1-cachyos, Intel i9-13900HX, with existing 20261018 offline dictionaries and likely warm OS page cache. No system cache was cleared. Results vary by hardware and load.

## Existing dictionaries and synonym mappings

The installed English StarDict `.ifo` advertises 800,017 headwords and 1,166,471 synonym records; Spanish advertises 123,686 and 1,052,756. The index files measured 15,465,283 and 2,240,634 bytes respectively, with synonym files 18,734,561 and 16,430,366 bytes. Input SHA-256:

| File | SHA-256 |
|---|---|
| `en-en.idx` | `225bec42c5f8759a6aece33c1b02dce54e893e9283c3987fe9f39312e889de90` |
| `en-en.syn` | `6b4db7ab9a097ac1a4fb6a6065449d37854998678ddf09a6e5779d760c0885b2` |
| `es-es.idx` | `2aeaf8323c620cd4e645710fecaec026225938d3038dc7540285275edfdf969b` |
| `es-es.syn` | `5f18d14a1c24e0909c32ae112614b80287bd369fabfcf71e17fe0b419687f24f` |

A direct scan of synonym records and their big-endian original-word indexes yielded these **sample mappings** (duplicates omitted):

| Language | Input | Linked index headwords |
|---|---|---|
| English | walking | walk |
| English | went | wend, ween, gan, go |
| English | children | childer, child |
| English | ran | rin, run |
| English | better | bettor, well, good, had better |
| Spanish | hablando | hablar |
| Spanish | fui | irse, ir, ser |
| Spanish | casas | casar, casarse, casa |
| Spanish | fueron | irse, ir, ser |
| Spanish | comiendo | comer |
| Spanish | mejores | mejor, mejorarse, mejorar |

These mappings do not establish all entries are grammatically suitable for the user's intended sense. The existing `scanDictionary` in `internal/adapters/offline.go` already matches synonym words and resolves their target indexes during ordinary offline lookup. If a mapped entry yields usable definitions, the lookup is `found`, so the approved miss-only alternatives flow will not run. This sample suggests testing **incremental coverage on actual misses**, not simply counting forms in a separate morphology dataset. It does not prove all listed examples yield displayable definitions or that these mappings cover the language.

English `look forward to` was present in `.idx`; the typo `look foward to` was absent from both English files. The Spanish phrase `dar a luz` was absent from both files as exact bytes in this small probe; it is not evidence of any other package's coverage.

## Scan timings

A temporary Python parser read each entire file into memory and iterated its NUL-terminated headwords and fixed-width binary records while matching eight sample strings per record. Two sequential scans in one process took 1.10–1.30 s for English `.idx`, 1.20–1.33 s for English `.syn`, 0.13–0.20 s for Spanish `.idx`, and 1.31–1.51 s for Spanish `.syn`. Python process peak resident memory after all scans was 49,136 KiB (includes interpreter and allocations; not memory of an optimized implementation). These timings **do not calculate edit distance** and are not the application latency. English had 159,436 index and 384,935 synonym entries containing ASCII spaces; Spanish had 8,969 and 2,653.

A separate temporary compiled C probe scanned both files sequentially and calculated ordinary Levenshtein distance at most 2 for byte strings containing spaces and within two **bytes** of the ASCII query length. On the English dictionary it found `look forward to` and took 236.7 and 247.2 ms for two fresh processes, processing 1,966,488 records and evaluating 263,958 candidates. Peak RSS was 1,728 and 1,924 KiB as reported by `getrusage`. The same ASCII English query over Spanish files took 56.4 ms (1,176,442 records, 4,708 candidate comparisons); that is not a representative Spanish or Unicode query. This C prototype does not implement Unicode normalization, correct rune-distance, cancellation, ranking, Go runtime behavior, concurrent app load, malformed input handling, cold disk or all required limits. Its timings cannot establish an acceptable end-to-end deadline.

## Reproduction

The temporary probes are `/tmp/wdict-scan-feasibility.py`, `/tmp/wdict-syn-probe.py` and `/tmp/wdict-phrase-probe.c` on this machine. Their SHA-256 values are respectively `1d9e9e2200d18d926f4c6b711292ead342dfb908a30ed126a52c171870e1d851`, `66a55f12ebda1c79d963fa4e8a63fa5093be641b1e8c3a1c6f8feb1cc6e43a47` and `bbf056577c6cd6fda5e1af4c7233eb4c08b6eb26a3cd594b873e3f94557aa7d7`. They are disposable, not project assets. Run:

```sh
python3 /tmp/wdict-scan-feasibility.py
python3 /tmp/wdict-syn-probe.py
cc -O2 /tmp/wdict-phrase-probe.c -o /tmp/wdict-phrase-probe
/tmp/wdict-phrase-probe "$HOME/.local/share/wdict/dictionaries/20261018/en-en.idx" "$HOME/.local/share/wdict/dictionaries/20261018/en-en.syn"
/tmp/wdict-phrase-probe "$HOME/.local/share/wdict/dictionaries/es-20261018/es-es.idx" "$HOME/.local/share/wdict/dictionaries/es-20261018/es-es.syn"
```

The Python probes parse `.idx` as NUL-terminated UTF-8 headword followed by two four-byte fields, and `.syn` as a headword followed by a four-byte big-endian original index. The C probe uses byte-wise lengths and edit distance; it intentionally is not a Unicode correctness test.

## UniMorph dataset evaluation

The sandbox initially failed DNS access. A subsequent authorized network request retrieved pinned public files into `/tmp`:

| Language | Git revision | Raw file size | Rows | SHA-256 |
|---|---|---:|---:|---|
| English | `66e0e9e8e2dcd196da081a25a48e5c1fe3d8b49b` | 18,022,905 bytes | 652,477 | `20a191cefdc7cad6fa74b00f49d6f658684f17b14541aae372e5a3d5a8c15c67` |
| Spanish | `b9655efb0e5ce0f99809badaa7f4e93017ab8430` | 50,335,761 bytes | 1,196,245 | `7a165d4b44eb078d23964f3eee3813f5140c5df0ffab0143c806c562e67f8771` |

Full-file offline overlap was approximated by taking unique form–lemma pairs, case folding, decomposing accents, removing combining marks and Unicode punctuation, and comparing to the similarly folded installed `.idx` and `.syn` headwords. Only missing forms containing a letter with a present lemma were counted. English yielded 569,251 unique folded forms, 565,244 present in dictionary index or aliases, and **1,158 potential additional forms** linked to a present lemma. Spanish yielded 805,789 unique folded forms, 348,155 present, and **346,605 potential additional forms** linked to a present lemma (348,451 pairs). These are candidate counts, **not demonstrated correct or actual application misses**: Unicode handling is only an approximation of Go's `foldWord`, an index name may not yield displayable definitions, and raw morphological labels and forms may be erroneous or implausible. Counts ignore frequency and do not measure user-facing value. Python held multiple full sets; the exploratory process reported 386,220 and 409,044 KiB peak RSS and 8.64 s and 10.27 s elapsed, respectively; that is not a compact index measurement or app budget.

All six Spanish examples in the previous table were present in both UniMorph and the dictionary. In English, `walking` maps to `walk`, `went` to `go`, but `children` had no UniMorph row in this file and `ran` returned `ran` and `rin`, not `run`. UniMorph `better` included `better`, `good`, `well`, `known` and `matching`, illustrating that blindly presenting all mappings can produce dubious suggestions. English sample possible incremental pairs include `abernathyites` → `abernathyite`; Spanish examples include `abacoradas` → `abacorar` and `abajadla` → `abajar`. These have not been checked for linguistic accuracy, actual app misses or useful definitions. At least one Spanish raw record had an empty form; input validation is required.

Reproduce overlap with `python3 /tmp/wdict-unimorph-filtered-eval.py` while the temporary files remain. The script uses the full downloaded data and installed dictionary indexes. Recreate downloads from `https://raw.githubusercontent.com/unimorph/eng/66e0e9e8e2dcd196da081a25a48e5c1fe3d8b49b/eng` and `https://raw.githubusercontent.com/unimorph/spa/b9655efb0e5ce0f99809badaa7f4e93017ab8430/spa`, verify the SHA-256 values above, then run the probe. Dataset license and redistribution terms still need review for the exact intended package; see the [research report](../../reports/English%20and%20Spanish%20morphology.md).

## Unresolved budgets

Do not set a universal deadline, memory or disk budget from this probe. Measure a Unicode-correct Go candidate generator on both real datasets, including representative Spanish accents, single words and long phrases, cold and warm caches, cancellation, and concurrent exact lookup on target machines. Evaluate at least first and repeat latency, peak additional memory, index and downloaded sizes, number of candidates evaluated and whether top-five results are useful. The existing 256-rune query validator and approved five-spelling-candidate output limit are known current/spec constraints, not measured performance budgets. Select numerical limits and a dataset only after that evidence and license review.

## Follow-up: actual offline lookups and Unicode Go probe

A disposable test supplied through the Go `-overlay` flag invoked the existing `OfflineDictionary.Find` against the installed dictionaries. The following results were observed on this machine; each pair contains a UniMorph candidate form and its proposed base:

| Language | Form → base | Actual `Find` outcome |
|---|---|---|
| English | `abernathyites` → `abernathyite` | form not found; base found (`abernathyite`) |
| English | `abietines` → `abietine` | form not found; base found with headword `abietin` (synonym resolution) |
| Spanish | `abacoradas` → `abacorar` | form not found; base found (`abacorar`) |
| Spanish | `abajadla` → `abajar` | form not found; base found (`abajar`) |

English `went` and `children`, and Spanish `fui` and `hablando`, produced existing found entries. The four incremental examples establish actual behavior for those inputs, but are an extremely small, unweighted sample and do not verify the correctness of the morphological relationships, prevalence, definition quality or candidate ranking. Their base lookup times were approximately 235–268 ms for the English cases and 660–750 ms for Spanish on likely warm data; this includes the existing offline scan and definition extraction, not the additional finder.

A separate disposable Go program streamed both `.idx` and `.syn`, normalized Unicode with NFD, case folding and mark/punctuation removal, then used rune-based Levenshtein distance ≤2 and ranked by distance followed by lexical order. It rejected query lengths above 64 runes, capped a scan at 2,100,000 records, checked a two-second context deadline every 1,024 records, and returned at most five distinct candidates. It filtered candidate length and phrase-vs-single-word before distance calculation. The following are two successive runs per query in one process, with an explicit GC before each run; disk caches were likely warm:

| Query | Time range | Records | Distance comparisons | Top candidates |
|---|---:|---:|---:|---|
| English `look foward to` | 602–619 ms | 1,966,488 | 266,895 | `look forward to`, `looks forward to` |
| English `wrods` | 661–666 ms | 1,966,488 | 457,381 | `Prods`, `RRODs`, `RRoDs`, `WODs`, `WROS` |
| Spanish `a contnuación` | 117–120 ms | 1,176,442 | 4,526 | `a continuación` |
| Spanish `a calzón qutado` | 116 ms | 1,176,442 | 4,990 | `a calzón quitado` |
| Spanish `arbl` | 393–406 ms | 1,176,442 | 96,608 | `Arbe`, `Arbel`, `Arbil`, `Arbo`, `Arel` |

Process peak RSS over all runs was 14,328 KiB; this is a process-wide peak, not additive application memory. Allocations per scan ranged from 32–33 MB for the two Spanish phrases to 243 MB for the English phrase, 326 MB for `wrods`, and 203 MB for `arbl` (total allocation, not peak live memory). No operation timed out in this sample. Critically, simple edit-distance/lexical ranking did **not** place the intuitive corrections `words` and `árbol` in the top five; the prototype is not evidence of acceptable suggestion quality. The record cap is only slightly above the installed English count and may reject future dictionaries.

This probe is **not production-safe**: it has no bounded record-byte parser, uses a rudimentary length prefilter that can exclude valid Unicode normalizations, and stores all matching candidates before truncating. Its time and record caps are experimental settings, not recommended hard budgets. It lacks cold-disk, concurrent-load, low-end hardware, cancellation-stress and malformed-data tests. A two-second proposed cap seems sufficient for these five warm examples on this machine, but is not validated for realistic worst cases or user expectations; do not adopt it as a release threshold without additional evidence and agreement. Candidate ranking needs a separate measured quality decision.

Reproduce while temporary files are present:

```sh
GOCACHE=/tmp/wdict-codex-go-cache go test -overlay /tmp/wdict-feasibility-overlay.json -tags gtk3 ./internal/adapters -run '^TestFeasibilityLookup$' -count=1 -v
GOCACHE=/tmp/wdict-codex-go-cache go build -o /tmp/wdict-unicode-probe /tmp/wdict-unicode-probe.go
/tmp/wdict-unicode-probe
```

`/tmp/wdict-unicode-probe.go` SHA-256: `e860227fda2e02416703b6b626377070be08eac984368002f60c64f39673ff98`; `/tmp/wdict-feasibility-test.txt`: `7de0a2bbcf427066106a77c449e0be64d3ff995555bd3900af1ad8237c62b5ee`; `/tmp/wdict-feasibility-overlay.json`: `29054fa5a7609d4c85f2c3da41352e7b3effc845b7aa27375783a2d6ade67211`. The overlay maps a nonexistent package-local test path to the temporary test and makes no repository source change.

## Follow-up: ranking and simultaneous scans

Two disposable variants of the Unicode Go probe evaluated adjacent character transpositions using optimal-string-alignment distance (OSA), still after NFD/case/mark/punctuation folding. They retain a maximum distance of two and the same input/record/deadline limits and unsafe prototype caveats above. Their tie-breakers are test hypotheses, not a choice approved in the spec:

1. OSA distance, then prefer all-lowercase display forms, then lexical order. This improved `arbl`: `árbol` appeared fourth, while `arbol` appeared first. `wrods` still returned `brods`, `prods`, `rods`, `trods`, `wods` and did not show `words` in its top five: OSA makes the transposition distance one, but many other single-edit candidates tie.
2. OSA distance, then prefer candidates that differ by exactly one adjacent transposition after normalization, then lowercase display forms, then lexical order. This returned **`words` first** for `wrods`; the `arbl` order remained `arbol`, `arel`, `arél`, `árbol`, `Arbe`. Both `look foward to` → `look forward to` and `a contnuación` → `a continuación` appeared first, as did `a calzón qutado` → `a calzón quitado`. These examples are not a representative quality benchmark: prioritizing transpositions may also promote incorrect results, case is not a popularity signal, `arbol` and `árbol` remain separate entries, and no frequency or language-specific disambiguation is used. Ranking needs broader evaluation; an observed top-five success should not be read as reliable correction.

The second variant, measured twice per query sequentially on the same machine and likely warm caches, took 749–770 ms for the English phrase, 726–744 ms for `wrods`, 117–122 ms for `a contnuación`, 123–131 ms for `a calzón qutado`, and 396–416 ms for `arbl`. The process-wide maximum RSS reported by `getrusage` was 13,800 KiB; it includes the Go runtime and all earlier sequential runs and is not incremental app memory. Total allocation per scan ranged from 32.6 MB for the Spanish phrase to 354.9 MB for `wrods`.

To assess contention, a separate process started the English phrase and Spanish `arbl` scans concurrently in two goroutines, three successive rounds, with an explicit GC between rounds and likely warm pages. Their elapsed times were 851/926/931 ms for English and 465/522/530 ms for Spanish; wall time for each pair was 851–931 ms. The process-wide `getrusage` maximum RSS reached 13,096 KiB, and total allocation was approximately 486.4–486.5 MB per pair. This runs two suggestion scans together; it does **not** simulate concurrent exact dictionary lookup, production UI, persistent index creation, other workloads, cold disk, CPU-constrained hardware, or all queries. `getrusage` maximum RSS is cumulative for the whole process, not a peak delta attributable to this feature. No reliable production deadline or memory limit follows from this short probe.

Reproduce while temporary files exist:

```sh
GOCACHE=/tmp/wdict-codex-go-cache go build -o /tmp/wdict-unicode-osa-probe /tmp/wdict-unicode-osa-probe.go
/tmp/wdict-unicode-osa-probe
GOCACHE=/tmp/wdict-codex-go-cache go build -o /tmp/wdict-unicode-osa-priority /tmp/wdict-unicode-osa-priority.go
/tmp/wdict-unicode-osa-priority
GOCACHE=/tmp/wdict-codex-go-cache go build -o /tmp/wdict-unicode-concurrent /tmp/wdict-unicode-concurrent.go
/tmp/wdict-unicode-concurrent
```

Source SHA-256s: `/tmp/wdict-unicode-osa-probe.go` `2763edd51f394d396865825c3716c09de36ac5f2122561fd09868a1159e8a97f`; `/tmp/wdict-unicode-osa-priority.go` `1631431bd44d85523f93b498ccb736bab2d1e33f80fc588456ada395267abe26`; `/tmp/wdict-unicode-concurrent.go` `21f319696bb26bf54352c49b67d94debc0970ac74fcb31a3a91d3f963adea11e`.
