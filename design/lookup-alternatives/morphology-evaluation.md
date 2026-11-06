# Morphology data and storage evaluation — 2026-11-06

This feasibility experiment complements [measurement.md](measurement.md), which identifies the pinned UniMorph files, installed dictionary hashes, machine and four actual-lookup examples. It is not a linguistic quality evaluation or a production storage implementation. The English and Spanish files are respectively UniMorph commits `66e0e9e8e2dcd196da081a25a48e5c1fe3d8b49b` and `b9655efb0e5ce0f99809badaa7f4e93017ab8430` with the SHA-256 hashes in that report.

## Incremental candidates

The installed dictionaries already resolve sampled forms such as English `went` and Spanish `fui`; existing aliases reduce the value of adding a second morphology source. The preceding probe verified four incremental pairs through the actual application lookup, but not their linguistic validity or frequency.

To avoid only inspecting alphabetical examples, a separate Python probe selected 20 unique potential extra pairs per language with deterministic seed `64061`. It used approximate NFD/case folding, removed combining marks and punctuation, and kept pairs with a letter-containing form absent from both installed index and synonym names but a lemma name present. It sampled **pairs**, not words in representative usage; it does not check app definitions, tags or linguistic correctness. The sample comprised:

| English form → proposed lemma | Spanish form → proposed lemma |
|---|---|
| protoselenides → protoselenide | negociándose → negociar |
| manœuvred → manœuvre | entreabrámoslos → entreabrir |
| biorelevances → biorelevance | amortizadla → amortizar |
| yewies → yewy | reglaméntalas → reglamentar |
| portyankis → portyanki | sofisticándolos → sofisticar |
| hadtals → hadtal | mascándole → mascar |
| phosphoenoylpyruvates → phosphoenoylpyruvate | desabrámosla → desabrir |
| dephosphatisations → dephosphatisation | traspillarlas → traspillar |
| syses → sys | lancémoslas → lanzar |
| stereofusions → stereofusion | cantéenlo → cantear |
| hexaploidisations → hexaploidisation | peritarlas → peritar |
| kazatskis → kazatski | ordeñémosla → ordeñar |
| cotwals → cotwal | secuenciándola → secuenciar |
| algodystrophies → algodystrophy | descongestionarlos → descongestionar |
| endotypings → endotyping | sobrevolándolas → sobrevolar |
| hotes → hote | egréselos → egresar |
| pyridoxylations → pyridoxylation | azogarnos → azogar |
| dodecacarbonates → dodecacarbonate | taladle → talar |
| twoses → twos | acordémoslo → acordar |
| tabbyhoods → tabbyhood | computarte → computar |

The many long Spanish clitic forms suggest a potentially useful gap, but the list also illustrates why counts alone are not enough: many examples are uncommon or require linguistic review. The English sample is largely uncommon plurals. None of these 40 pairs was independently assessed by a linguist; absence from the simplified name set does not guarantee a miss in the app. Morphology is ambiguous: the pinned English file maps `better` to `good`, `well`, `known`, and `matching`; Spanish `casas` maps to both `casa` and `casar`. Dictionary membership alone cannot determine the intended sense.

## Compact storage experiment

A throwaway Python/SQLite experiment stored distinct original UTF-8 `(form, lemma)` pairs in a `WITHOUT ROWID` table whose composite primary key supports exact form lookup. It discarded malformed/empty fields and identical form/lemma pairs, retained ambiguous alternatives, and omitted morphological tags. No dictionary-overlap filter was applied; therefore both counts include entries already resolved by the installed dictionary. An SQLite index without tags also cannot support a later part-of-speech or feature filter. The resulting files are experimental and not suitable for distribution without a package, integrity, attribution and trust-boundary design.

| Data | Raw rows | Invalid/empty | Same-form rows excluded | Distinct pairs | SQLite | gzip -9 | Build time | gzip time |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| English | 652,477 | 0 | 427,927 | 192,833 | 4,853,760 B | 1,657,529 B | 0.48 s | 0.31 s |
| Spanish | 1,196,245 | 21 | 57,607 | 839,713 | 21,876,736 B | 5,221,763 B | 1.45 s | 1.41 s |

Build timing ends before `VACUUM`; gzip timing follows it. The two datasets were processed sequentially in one Python process; its cumulative peak RSS was 25,304 KiB after English and 26,456 KiB after Spanish. This includes interpreter, SQLite and input buffering, does not measure compressed archive extraction, and is not incremental app memory or a cold-machine measurement. Timings are one warm-file run on the machine in measurement.md, with durability and journal disabled **for this disposable build only**. SQLite is not currently chosen as an app dependency or package format.

In a separate read-only SQLite/Python process, first lookups for five sample forms took 0.019–0.140 ms each, and 10,000 sequential lookups over selected sorted forms took 52.3 ms English and 77.4 ms Spanish. Process peak RSS after both languages was 19,184 KiB; Python and its 10,000-form list are included. This tests warm exact byte equality only: it does not cover app Unicode folding, rebuilding, actual dictionary verification, startup, cancellation, concurrent exact lookup, cold disk or Go integration. Under the app's folding rules, a separately normalized key/index and a policy preserving multiple original spellings are required, with additional size and behavior to measure. Build and query figures do not prove a two-second end-to-end budget.

## Candidate and fallback recommendation

- **English:** do not ship the pinned file solely on current evidence. Its estimated 1,158 folded extra forms are a small fraction of the 569,251 forms and the file has questionable candidate mappings. Existing synonyms already provide many irregular results; representative frequency-weighted benefit and correctness are unknown.
- **Spanish:** continue evaluating the pinned file as a possible independently versioned package. The approximate 346,605 extra forms with existing lemma names and the sample of clitic forms make it a stronger candidate than English, but do not establish accuracy, usefulness or release readiness. A stratified, independently checked set of common forms and clitics, actual app misses and candidate quality is needed; preserve multiple candidate lemmas and identify them as possible. The experimental compact storage establishes feasibility only for an exact-key mapping, not a complete safe design.
- **Fallback for both:** a conservative, deterministic starting proposal is to try only a bounded list of literal transformations, verify every resulting headword in the appropriate installed dictionary, and show results as *possible*. English: remove `s`, `es`, `ed` or `ing` when a nonempty stem remains; for `ed` and `ing`, also try appending `e`. Spanish: remove final `s` or `es` when a nonempty stem remains. Deduplicate generated strings, do not recursively apply rules, and leave irregular or unhandled inflections to dictionary aliases. These deliberately incomplete rules can produce unrelated valid words, so membership does not prove morphology; no quantitative accuracy or coverage is claimed. Exact rules are proposed for review, not approved product behavior. A representative fixture set and explicit work/time limits are required before implementation.

## Rights and reproducibility

The [English](https://github.com/unimorph/eng) and [Spanish](https://github.com/unimorph/spa) data repositories declare CC BY-SA 3.0; their upstream descriptions and [license text](https://creativecommons.org/licenses/by-sa/3.0/) should be reviewed for the actual derived artifact and distribution method. Record source revision, license, attribution and transformation; do not infer that a data grant changes wdict's MIT code license. This is not a legal clearance and does not resolve contributor-specific provenance questions. The prior [research report](../../reports/English%20and%20Spanish%20morphology.md) describes its source limitations.

Reproduce the local probes while pinned data and the installed dictionaries are available:

```sh
python3 /tmp/wdict-morph-sample2.py
python3 /tmp/wdict-morph-eval2.py
python3 /tmp/wdict-morph-query2.py
```

Temporary-script SHA-256 values: sample `097e930fe83f91625765cec7f39611dedd3062aa7982a09b15eb56ef6166272d`, build `671eca9bbdbdff23216468561ea336a933352a93d66f749b1f8fe24351998c3b`, query `2dbf8efddc6914dd533d66fa60931cf6428643e1d0f46a7e1e50d8c74f127e2d`. Scripts and outputs under `/tmp` are disposable. The build script now refuses to overwrite existing temporary SQLite outputs; move those files to trash before rerunning. The measured version was identical apart from its overwrite behavior. Dataset hashes and dictionary index hashes should be checked as documented in [measurement.md](measurement.md) before comparing results.
