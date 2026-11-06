# Bilingual spelling ranking probe — 2026-11-06

This is an exploratory, manually constructed sample, not a representative distribution of actual user searches. The fixture choices and intended answers were authored before running the existing prototype; no weights or scoring were tuned on these cases. The table reports one sequential warm-cache run on the Intel i9-13900HX reference machine. Each intended answer was checked as an exact `.idx` or `.syn` record. The query's approximate folded membership is also checked; these cases would generally not trigger the approved miss-only path, though a headword need not yield a usable definition. Search language, dictionary version, current app folding, and offline availability all matter. See [measurement.md](measurement.md) for dictionary hashes and prototype limitations.

The untouched prototype uses NFD, case folding, accent and punctuation removal; Unicode optimal-string-alignment distance ≤2; then exact adjacent-transposition preference, lowercase preference and lexical order; it returns five. It is a proof-of-concept with unbounded record parsing, a two-second deadline, 64-rune input cap and 2.1-million-record cap, and is not production code. Its string display is preserved below. Each expected target is present in the corresponding installed index or synonym file, but synonym hits are not independently proven to yield usable definitions. The Python folded membership check is approximate, not an actual application lookup.

| Lang | Input → intended | Approx. folded query present | Top five in rank order | Target rank | Time ms |
|---|---|---|---|---:|---:|
| en | `helo` → `hello` | yes | `delo`, `elo`, `h'lo`, `halo`, `halo-` | — | 600.8 |
| en | `wrld` → `world` | no | `wald`, `weld`, `wild`, `wld`, `wold` | — | 391.3 |
| en | `recieve` → `receive` | yes | `receive`, `decieve`, `relieve`, `resieve`, `believe` | 1 | 661.8 |
| en | `becuase` → `because` | yes | `because`, `becurse`, `b'cause`, `beacause`, `beacuse` | 1 | 661.6 |
| en | `freind` → `friend` | yes | `-friend`, `friend`, `Friend`, `feind`, `frein` | 2 | 575.4 |
| en | `langauge` → `language` | yes | `language`, `langage`, `langauges`, `e-language`, `langate` | 1 | 747.4 |
| en | `dictionry` → `dictionary` | no | `dictionary`, `Dictionary`, `dicktionary`, `diction`, `dictional` | 1 | 875.1 |
| en | `computre` → `computer` | no | `computer`, `compute`, `combure`, `commute`, `commuter` | 1 | 1232.3 |
| en | `watre` → `water` | no | `water`, `Water`, `-ware`, `catre`, `ware` | 1 | 755.1 |
| en | `mispell` → `misspell` | yes | `dispell`, `mis-sell`, `mis-spell`, `miscell.`, `misell` | — | 1086.2 |
| en | `give upp` → `give up` | no | `give up`, `gave up`, `give dap`, `give lip`, `given up` | 1 | 368.4 |
| en | `break dowm` → `break down` | no | `break down`, `break downs`, `break room`, `breaks down` | 1 | 545.0 |
| en | `take offf` → `take off` | no | `take off`, `take offs`, `bake off`, `make off`, `rake off` | 1 | 423.4 |
| es | `holaa` → `hola` | no | `hola`, `holea`, `holeá`, `holgá`, `hollá` | 1 | 372.7 |
| es | `casaa` → `casa` | no | `casa`, `casaba`, `casaca`, `casad`, `casada` | 1 | 370.5 |
| es | `perrro` → `perro` | no | `peorro`, `peroro`, `peroró`, `perreo`, `perrero` | — | 667.1 |
| es | `gatto` → `gato` | no | `atto-`, `gasto`, `gastó`, `gateo`, `gateó` | — | 477.1 |
| es | `comidáa` → `comida` | no | `comida`, `comidan`, `comidas`, `comidea`, `comideá` | 1 | 1110.5 |
| es | `libroo` → `libro` | no | `libreo`, `libreó`, `libro`, `libros`, `libró` | 3 | 791.5 |
| es | `escueal` → `escuela` | no | `escuela`, `escualo`, `escuda`, `escudad`, `escudan` | 1 | 1112.9 |
| es | `corazno` → `corazón` | no | `corazón`, `acorazo`, `acorazó`, `brazno`, `braznó` | 1 | 1114.4 |
| es | `palaabra` → `palabra` | no | `palabra`, `alabara`, `alabará`, `alambra`, `alambrá` | 1 | 1354.4 |
| es | `diicionario` → `diccionario` | no | `diccionario`, `diccionarios`, `Wikcionario` | 1 | 1665.8 |
| es | `por favro` → `por favor` | no | `por favor` | 1 | 112.7 |
| es | `buenos diass` → `buenos días` | no | `buenos dias`, `buenos días` | 2 | 128.3 |
| es | `a pesra de` → `a pesar de` | no | `a pesar de` | 1 | 109.1 |

The intended candidate appeared in the top five for 21/26 fixtures overall, and 17/20 fixtures whose query was absent under the approximate folding check. The three miss-path failures are `wrld` → `world`, `perrro` → `perro`, and `gatto` → `gato`. Six other queries already have approximately folded matches and are not credible confirmed misses without an actual lookup test. No fixture timed out. The maximum displayed run was 1,665.8 ms, not a cold-cache or low-end-device bound. With hand-picked fixtures and only one sample per query, these results cannot support a general quality threshold, estimate expected user accuracy, or establish the two-second policy as a production guarantee.

A deterministic OSA/transposition/lowercase/lexical ordering is a reproducible baseline, but should not be selected as the final policy on this evidence: common corrections still lose ties, accents can produce near-duplicate results, and lowercase is not a frequency signal. Independently curated and split English and Spanish typo sets, including no-valid-answer and ambiguous cases, should establish top-five recall, misleading-suggestion rate, and confidence before an acceptance threshold is proposed. Any frequency source would require a license and provenance review; no such data was used here.

## Reproduction

The source prototype is `/tmp/wdict-unicode-osa-priority.go`; `/tmp/wdict-rank-gen.py` replaces only its main function with the frozen cases in `/tmp/wdict-rank-fixtures.json`, producing `/tmp/wdict-ranking-fixtures.go`. All paths are temporary and may be removed by the system. From the repo root:

```sh
python3 /tmp/wdict-rank-gen.py
GOCACHE=/tmp/wdict-codex-go-cache go run /tmp/wdict-ranking-fixtures.go
python3 /tmp/wdict-rank-membership.py
python3 /tmp/wdict-ranking-fold-check.py
```

Source hashes:

- `/tmp/wdict-unicode-osa-priority.go`: `1631431bd44d85523f93b498ccb736bab2d1e33f80fc588456ada395267abe26`
- `/tmp/wdict-rank-fixtures.json`: `37e35bc4693c89b26bac40002ef5444fd5710401f2728bac1b8aec75a1113281`
- `/tmp/wdict-rank-gen.py`: `803560633c0deb3d14ae04bc1d4919e3ea682eb688691a356103a8d2bf89f7c6`
- `/tmp/wdict-ranking-fixtures.go`: `d4cd27b4f892ede6cb6c938406a0185815d17f861578526533a3bef8b62076a3`
- `/tmp/wdict-rank-membership.py`: `3d4486462da0e8bcfaced9cac965412cf5151cc5f926e3f675036cac05847219`
- `/tmp/wdict-ranking-fold-check.py`: `400de1f48d7cf37fab773df3e0504d2081d0e5f053b13aa4d4210afe8dd16025`

## Follow-up: repeated-letter and initial-letter preference

After inspecting the baseline failures, a **post hoc** variant retained distance and adjacent-transposition priority, then preferred a candidate that differs by exactly one inserted or deleted repeated letter, then a matching initial folded rune, then lowercase and lexical order. The variant was run against the same frozen fixtures, so this is development-set feedback, not an independent accuracy estimate. This rule may promote incorrect alternatives; it is not an approved or validated ranking policy. Both algorithms share the original prototype's correctness, time and resource limitations.

| Lang | Input → intended | Variant top five | Intended rank | Time ms |
|---|---|---|---:|---:|
| en | `helo` → `hello` | `hello`, `h'lo`, `halo`, `halo-`, `he'll` | 1 | 524.6 |
| en | `wrld` → `world` | `wald`, `weld`, `wild`, `wld`, `wold` | — | 535.9 |
| en | `recieve` → `receive` | `receive`, `relieve`, `resieve`, `decieve`, `re-cite` | 1 | 927.5 |
| en | `becuase` → `because` | `because`, `becurse`, `b'cause`, `beacause`, `beacuse` | 1 | 1050.8 |
| en | `freind` → `friend` | `-friend`, `friend`, `Friend`, `feind`, `frein` | 2 | 830.0 |
| en | `langauge` → `language` | `language`, `langage`, `langauges`, `langate`, `langrage` | 1 | 1002.1 |
| en | `dictionry` → `dictionary` | `dictionary`, `Dictionary`, `dicktionary`, `diction`, `dictional` | 1 | 1074.3 |
| en | `computre` → `computer` | `computer`, `compute`, `combure`, `commute`, `commuter` | 1 | 951.6 |
| en | `watre` → `water` | `water`, `Water`, `-ware`, `ware`, `warre` | 1 | 590.9 |
| en | `mispell` → `misspell` | `mis-spell`, `misspell`, `mis-sell`, `miscell.`, `misell` | 2 | 833.6 |
| en | `give upp` → `give up` | `give up`, `gave up`, `give dap`, `give lip`, `given up` | 1 | 299.6 |
| en | `break dowm` → `break down` | `break down`, `break downs`, `break room`, `breaks down` | 1 | 464.5 |
| en | `take offf` → `take off` | `take off`, `take offs`, `take huff`, `taken off`, `takes off` | 1 | 361.9 |
| es | `holaa` → `hola` | `hola`, `holá`, `holea`, `holeá`, `holgá` | 1 | 522.4 |
| es | `casaa` → `casa` | `casa`, `casá`, `casaba`, `casaca`, `casad` | 1 | 483.2 |
| es | `perrro` → `perro` | `perro`, `peorro`, `peroro`, `peroró`, `perreo` | 1 | 651.9 |
| es | `gatto` → `gato` | `gato`, `gasto`, `gastó`, `gateo`, `gateó` | 1 | 521.3 |
| es | `comidáa` → `comida` | `comida`, `comidan`, `comidas`, `comidea`, `comideá` | 1 | 849.6 |
| es | `libroo` → `libro` | `libro`, `libró`, `libreo`, `libreó`, `libros` | 1 | 653.9 |
| es | `escueal` → `escuela` | `escuela`, `escualo`, `escuda`, `escudad`, `escudan` | 1 | 853.7 |
| es | `corazno` → `corazón` | `corazón`, `codazo`, `colazo`, `corado`, `coralino` | 1 | 900.5 |
| es | `palaabra` → `palabra` | `palabra`, `palabras`, `palabrea`, `palabreá`, `palabro` | 1 | 1093.7 |
| es | `diicionario` → `diccionario` | `diccionario`, `diccionarios`, `Wikcionario` | 1 | 1430.2 |
| es | `por favro` → `por favor` | `por favor` | 1 | 123.1 |
| es | `buenos diass` → `buenos días` | `buenos dias`, `buenos días` | 2 | 113.6 |
| es | `a pesra de` → `a pesar de` | `a pesar de` | 1 | 100.7 |

On these same examples the variant ranked the intended answer in the top five for 25/26 total and 19/20 approximate miss-path cases, versus 21/26 and 17/20 for the baseline. It recovered `perrro` → `perro` and `gatto` → `gato` at rank 1; `wrld` → `world` remained absent from the top five. Baseline-non-miss cases `helo` and `mispell` also moved into the top five. No intended answer from the baseline top five fell out on this set. The extra rules were chosen after seeing failures, so comparing these numbers overstates likely generalization; no significance or validated acceptance threshold follows. The maximum measured variant fixture time was 1,430.2 ms in this warm run, not a production limit.

The original five probe examples were rerun with the variant: `wrods` → `words`, `look foward to` → `look forward to`, `a contnuación` → `a continuación`, and `a calzón qutado` → `a calzón quitado` each ranked first; `arbl` → `árbol` ranked fourth. This confirms preservation only for those examples, not all transpositions. A fresh independently constructed bilingual holdout, including misses and cases where the repeated letter is intentional, remains necessary before deciding whether to use these rules.

Reproduce with the existing temporary fixtures and prototype:

```sh
python3 /tmp/wdict-rank-variant-gen.py
GOCACHE=/tmp/wdict-codex-go-cache go run /tmp/wdict-ranking-variant.go
GOCACHE=/tmp/wdict-codex-go-cache go run /tmp/wdict-rank-variant-original.go
```

- `/tmp/wdict-rank-variant-gen.py`: `b33471992f7523af294fde6e689909c1dc7766899d23557ff8d2531731343521`
- `/tmp/wdict-ranking-variant.go`: `7b1b4494755cbdc1b805ae0d3522fa8b3a6291e12f3675617d5e8158491da300`
- `/tmp/wdict-rank-variant-original.go`: `7573f7bec6088772ae40383b29cda57ec49d323bda343fa3f16cc24b44c9e603`
