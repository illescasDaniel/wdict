# English offline morphology datasets

## Which data can map English forms to lemmas?

### Takeaway
UniMorph English is a concrete, downloadable morphology candidate and Wiktextract is a broader alternative; neither has been evaluated against wdict's own dictionary or a representative accuracy set. A dataset can map multiple possible lemmas per surface form, so lookup should preserve ambiguity and verify that each suggested lemma exists in the installed dictionary.

### Cited Findings
- [UniMorph English](https://github.com/unimorph/eng) publishes an English data file, identifies its source as Wikipedia and declares CC BY-SA 3.0. The [data file page](https://github.com/unimorph/eng/blob/master/eng) reports 17.2 MB. The project's [schema description](https://unimorph.github.io/) describes linking a lemma and inflected form with morphological features.
- [Wiktextract paper](https://aclanthology.org/2022.lrec-1.140/) describes structured entries with `word`, `forms` and sense-level `form_of` fields. Its [datasets example](https://kaikki.org/dictionary/English/meaning/d/da/datasets.html) explicitly maps `datasets` to `dataset`. Its [multiword example](https://kaikki.org/dictionary/English/meaning/l/le/learner%27s%20licenses.html) maps `learner's licenses` to `learner's license`.
- [Kaikki raw downloads](https://kaikki.org/dictionary/rawdata.html) report the 2026-10-01 enwiktionary dump extracted as 24.8 GB JSONL or 2.9 GB gzipped, covering hundreds of languages. The English-only postprocessed download is reported as 3.2 GB, but explicitly deprecated for processing; [English dataset page](https://kaikki.org/dictionary/English/) directs processors to raw data. Size of an extracted, compressed form-to-lemma index has not been measured.
- [Wiktextract paper](https://aclanthology.org/2022.lrec-1.140/) documents extraction and linguistic fields, but this does not provide a wdict-specific accuracy guarantee.

### Inferences
- UniMorph is an attractive experiment because it is already structured as morphology and substantially smaller than the general raw Wiktextract dump. It needs a reverse form-to-lemma index and validation against the app's installed headwords; a mapping alone does not establish that the dictionary contains its target.
- Raw Wiktextract is an alternative for a separately built downloadable index with richer provenance and multiple part-of-speech analyses, at the cost of a heavier extraction pipeline.

### Gaps
- UniMorph's raw file could not be fetched through the browsing tool (timeout); actual rows, counts, irregular examples, freshness and accuracy were not independently verified. Do not claim complete irregular coverage or production readiness.
- No corpus-based evaluation of recalls, false positives, index size or scan latency for either candidate was performed. Existing dictionary may already contain some inflected entries; measure incremental value on actual downloaded StarDict data.

## Can wdict distribute the data?

### Takeaway
The listed licenses permit reuse subject to attribution and share-alike obligations, but exact distribution and derivative-data treatment require review; they do not establish an automatic license change to unrelated application code.

### Cited Findings
- UniMorph English declares [CC BY-SA 3.0](https://github.com/unimorph/eng); the [Creative Commons deed](https://creativecommons.org/licenses/by-sa/3.0/) states attribution, link to license, indication of changes and share-alike conditions for adaptations.
- [Kaikki English edition index](https://kaikki.org/dictionary/) states Wiktionary-extracted data is available under CC BY-SA and GFDL. [Wikimedia terms](https://foundation.wikimedia.org/wiki/Policy:Terms_of_Use) govern attribution and reuse of Wikimedia-hosted textual content; verify the exact snapshot and source licensing notices at packaging time.

### Inferences
- Keep the morphology data separately identified with source, revision, license text and transformation notes. Verify target-specific share-alike and attribution obligations before redistributing an index; treat compatibility as a licensing review rather than a foregone conclusion.

### Gaps
- This investigation is not a legal determination; no version-specific source-level provenance or build recipe for an actual distributable derived index was audited.
