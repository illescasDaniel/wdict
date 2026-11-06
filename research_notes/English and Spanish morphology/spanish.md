# Spanish offline morphology

## Which dataset can map forms to lemmas?

### Takeaway
UniMorph Spanish is a strong candidate to evaluate: its published dataset contains Spanish noun, adjective and verb paradigms and is explicitly licensed CC BY-SA 3.0. Its measured size, quality, overlap with wdict's particular Spanish dictionary, and installation method still need testing before selection.

### Cited Findings
- [UniMorph Spanish](https://github.com/unimorph/spa) says its `spa` data comes from English Wiktionary and includes 51,395 noun lemmas, 12,700 adjective lemmas, 6,812 verb lemmas and 1,196,245 inflectional forms. These are project-reported statistics, not independently checked. The file is [shown as 48 MB](https://github.com/unimorph/spa/blob/master/spa) before any local compression or index.
- [UniMorph's project description](https://unimorph.github.io/) defines a lemma, inflected form and morphological feature bundle representation. A reverse index could map an input form to potentially multiple lemmas. More than one lemma is necessary: [Wiktextract's Spanish `fui` entry](https://kaikki.org/dictionary/Spanish/meaning/f/fu/fui.html) shows both `ir` and `ser`; this does not independently confirm those rows in the UniMorph file.
- [UniMorph's Spanish README](https://github.com/unimorph/spa) states CC BY-SA 3.0. [Creative Commons](https://creativecommons.org/licenses/by-sa/3.0/) says reuse requires appropriate credit and a license link; adaptations have share-alike terms. It allows commercial reuse on those conditions.

### Inferences
- A separately attributed morphology data file, preserving source/version/license and providing a reversible mapping, is technically plausible with the MIT application; do not assert unconditional legal compatibility or that a derived index can be relicensed MIT. Review exact distribution, attribution and share-alike implications. It would require a new data artifact or installation process; it is not already the downloaded Spanish StarDict.
- Verify returned lemmas exist in the installed Spanish dictionary and show multiple candidates when ambiguous. Missing matches remain normal and do not prove misspelling.

### Gaps
- No end-to-end reliability benchmark, representative irregular verb checks, accent-folding tests, coverage overlap with the pinned Spanish Wiktionary archive, compressed sizes, build cost, checksum/pinning, or license counsel assessment were performed. Confirm precise revision and provenance/attribution before distribution; counts may not correspond to the current file.

## What alternatives deserve consideration?

### Takeaway
Wiktextract exposes richer Wiktionary form relations but raw extraction is substantially larger; Spanish Hunspell resources target spelling and may introduce separate engine and licensing concerns.

### Cited Findings
- [Kaikki Spanish](https://kaikki.org/dictionary/Spanish/index.html) reports 775,464 distinct forms at its snapshot and a deprecated postprocessed 1.0 GB JSONL download; [raw English-Wiktionary extraction](https://kaikki.org/dictionary/rawdata.html) is 24.8 GB uncompressed or 2.9 GB compressed across all languages. The Spanish page records an October 2026 dump; the precise cost of generating a compact morphology subset is unknown.
- [Spanish `fui` raw and postprocessed example](https://kaikki.org/dictionary/Spanish/meaning/f/fu/fui.html) shows structured `form_of` for `ir` and `ser`. [Wiktextract](https://github.com/tatuylonen/wiktextract/blob/master/README.md) documents its extraction and downloadable JSONL data. This demonstrates one example, not reliable coverage of the whole vocabulary; identify data-specific rights separately from the extractor's code license.
- [RLA-ES](https://github.com/sbosio/rla-es) builds Spanish spellchecking dictionaries for many regional variants and explicitly notes variant coverage can be incomplete. Its [license file](https://github.com/sbosio/rla-es/blob/master/LICENSE.md) offers GPL 3+, LGPL 3+, or MPL 1.1+ for its dictionaries. [Hunspell](https://github.com/hunspell/hunspell) is an analyzer and spellchecker under its own LGPL/GPL/MPL tri-license; dataset licensing is separate.

### Inferences
- UniMorph is the simpler first evaluation for explicit form-to-lemma pairs; RLA-ES could merit separate spelling evaluation, but a spellchecking lexicon is not by itself evidence that its flags can reliably return dictionary lemmas. Additional engine integration, cross-platform packaging and each artifact's license need review.

### Gaps
- No data files were downloaded and no examples were executed. No compatibility conclusion for RLA-ES or Wiktextract has been reached, and no claim of complete Spanish morphology is justified.
