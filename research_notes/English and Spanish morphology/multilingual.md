# Multilingual morphology datasets

## Which data can map forms to lemmas offline?

### Takeaway
UniMorph Spanish and Wiktextract are promising structured sources, though neither has yet been benchmarked or evaluated against the exact installed wdict dictionaries. UniMorph has an explicit Spanish dataset with broad paradigms; Wiktextract is substantially larger but can produce form mappings for both languages from one source.

### Cited Findings
- The [UniMorph Spanish repository](https://github.com/unimorph/spa/blob/master/README.md) identifies English Wiktionary as its source, lists nouns, adjectives and verbs, and reports 51,395 noun, 12,700 adjective and 6,812 verb lemmas with 1,196,245 inflectional forms. These are repository-reported counts, not verified coverage or accuracy.
- The [UniMorph English repository](https://github.com/unimorph/eng) states its source as Wikipedia and its data license as CC BY-SA 3.0; the [project's dataset index](https://unimorph.github.io/) reports 115,523 forms and 22,765 paradigms for an English version, with verbs indicated. The age/version of these index counts must be verified against the file chosen. The source attribution is imprecise and should not be silently relabeled Wiktionary.
- UniMorph represents each form with a lemma and morphological feature bundle; relationships can be inverted into form-to-lemma mappings. [UniMorph project](https://unimorph.github.io/) and [UniMorph 4.0 paper](https://aclanthology.org/2022.lrec-1.89/).
- [Wiktextract's schema](https://github.com/tatuylonen/wiktextract/blob/master/README.md) contains language, headword and `forms` records with `form` and `tags`, including inflections and alternatives. Forms can be missing or contain a placeholder, so selection and cleanup are necessary. Its current [type definitions](https://github.com/tatuylonen/wiktextract/blob/master/src/wiktextract/extractor/en/type_utils.py) also expose sense-level `form_of` references.
- Kaikki's [English](https://kaikki.org/dictionary/English/) and [Spanish](https://kaikki.org/dictionary/Spanish/) pages report 1,396,111 and 775,464 distinct word forms respectively on their 2026-10-30 extraction. These are page word counts, not counts of accurate lemma mappings. The displayed data is postprocessed and may incorporate additional sources.

### Inferences
- Inverting paradigm data provides candidates for irregular forms as well as regular forms when included, but a form may have multiple lemmas and dictionary membership must be checked per language. A mapping is not a disambiguated meaning.
- Pinning a particular source snapshot and publishing reproducible conversion/checksums could make an optional small derived lookup index practical, subject to measurement and rights review.

### Gaps
- No actual sample mappings, ambiguity rate, recall, coverage against the selected StarDict snapshots, offline download/install costs, final compact-index size, or real-device latency were measured. Full compatibility and reliability are not established. Multiword correction requires a phrase/headword index and is not solved by morphology mappings alone.

## What licensing and practical distribution risks apply?

### Takeaway
The datasets' stated CC BY-SA 3.0 terms are not the extractor's MIT license. A distributable derived data package requires a concrete attribution, license and provenance plan; avoid claiming the entire MIT application or arbitrary merged Kaikki data is automatically cleared.

### Cited Findings
- Both [UniMorph Spanish](https://github.com/unimorph/spa/blob/master/README.md) and [UniMorph English](https://github.com/unimorph/eng) state CC BY-SA 3.0 for their data. Spanish explicitly identifies English Wiktionary; English says Wikipedia.
- The [Wiktextract software license](https://github.com/tatuylonen/wiktextract/blob/master/LICENSE) is MIT with separate conditions for certain test files. This software license does not grant MIT terms for extracted Wiktionary data. The applicable [Wikimedia Terms of Use](https://foundation.wikimedia.org/wiki/Policy:Terms_of_Use) and original material/attribution must be reviewed for a specific snapshot and derived package.
- Kaikki's [raw data page](https://kaikki.org/dictionary/rawdata.html) as of the 2026-10-01 dump lists all-language raw JSONL at 24.8 GB (2.9 GB compressed); its per-language postprocessed [English](https://kaikki.org/dictionary/English/) and [Spanish](https://kaikki.org/dictionary/Spanish/) downloads are listed at 3.2 GB and 1.0 GB respectively and described as deprecated. This is not the size of a generated morphology-only package.

### Inferences
- UniMorph is likely a simpler first feasibility candidate than processing the all-language raw dump, but selecting either remains contingent on actual file/version, checksums, accuracy, distribution and legal review. Existing StarDict package provenance and mapping lemma availability may diverge across source snapshots.

### Gaps
- No final compressed morphology-only sizes or distribution benchmarks were established. We did not verify a specific UniMorph revision, whether its attribution metadata is adequate for redistribution, or the licenses of any extra sources mixed into Kaikki's postprocessed data.
