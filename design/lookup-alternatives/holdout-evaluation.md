# Independent bilingual typo holdout — 2026-11-06

I froze 29 hand-created fixtures before inspecting the ranking implementation or prior fixtures (SHA-256 `c299f96ffc3b2ccef21fb07c3c4ec0c37bb0b157af8a0096938b10a3f5a7d073`). They contain English and Spanish ordinary words, adjacent transpositions, repeated letters, accents, phrases, two nonsense queries and ambiguous/valid-form cases. These are not sampled from user searches and no statistical population estimate is justified. All 27 expected headwords were confirmed present as exact index or synonym names, not necessarily as entries with usable definitions. One run used the existing post-hoc tuned Go prototype on the installed 20261018 dictionaries, likely warm caches on the reference machine. A parallel membership scan overlapped early execution; timings are exploratory and include contention. Candidate ordering and normalization are those of the prototype, not production implementation.

| Language | Input | Intended | Approximate folded query already present | Top five with edit distance | Time ms |
|---|---|---|---|---|---:|
| en | `appel` | `apple` | yes | `[{apple 1} {Apple 1} {Papel 1} {Appell 1} {'appen 1}]` | 687.3 |
| en | `bananna` | `banana` | no | `[{banana 1} {BANANA 1} {Banana 1} {bandanna 1} {Bahanna 1}]` | 810.4 |
| en | `tomorow` | `tomorrow` | no | `[{to-morrow 1} {tommorow 1} {tomorrow 1} {to‐morrow 1} {thorow 2}]` | 829.4 |
| en | `beautifull` | `beautiful` | yes | `[{beautiful 1} {beautifully 1} {beautifuls 1} {beastfull 2} {beautifool 2}]` | 838.1 |
| en | `teh` | `the` | yes | `[{the 1} {the- 1} {-eth 1} {eth 1} {eth- 1}]` | 361.9 |
| en | `peopel` | `people` | no | `[{-people 1} {people 1} {propel 1} {Pepel 1} {pee-pee 2}]` | 606.7 |
| en | `neccessary` | `necessary` | yes | `[{necessary 1} {Necessary 1} {necessar 2} {necessarry 2} {necessars 2}]` | 818.2 |
| en | `hous` | `house` | yes | `[{Huos 1} {ohus 1} {houss 1} {'hos 1} {hobs 1}]` | 442.6 |
| en | `bookk` | `book` | no | `[{book 1} {Book 1} {booke 1} {books 1} {booky 1}]` | 476.6 |
| en | `grren` | `green` | no | `[{gren 1} {Gren. 1} {graen 1} {green 1} {Garen 1}]` | 495.9 |
| en | `make upp` | `make up` | no | `[{make up 1} {make ups 1} {made up 2} {make use 2} {makes up 2}]` | 339.9 |
| en | `turn onn` | `turn on` | no | `[{turn on 1} {turn down 2} {turn in 2} {turn off 2} {turn out 2}]` | 346.1 |
| en | `look upp` | `look up` | no | `[{look up 1} {lock up 2} {look upon 2} {looks up 2} {loop up 2}]` | 340.2 |
| en | `zzqqxx` | none | no | `[]` | 662.2 |
| es | `mannana` | `mañana` | no | `[{mañana 1} {manzana 1} {macana 2} {malsana 2} {manaba 2}]` | 905.7 |
| es | `gracais` | `gracias` | no | `[{garcáis 1} {gracias 1} {grabáis 1} {gradáis 1} {gramáis 1}]` | 823.7 |
| es | `muchoo` | `mucho` | no | `[{mucho 1} {muchos 1} {macheo 2} {macheó 2} {macho 2}]` | 494.8 |
| es | `ninnos` | `niños` | no | `[{niños 1} {ñiños 1} {ninfos 1} {nanos 2} {neneos 2}]` | 490.9 |
| es | `famila` | `familia` | no | `[{familia 1} {fámula 1} {mamila 1} {Camila 1} {Ramila 1}]` | 493.7 |
| es | `mujerr` | `mujer` | no | `[{mujer 1} {majear 2} {mejer 2} {mejerá 2} {mejeré 2}]` | 489.0 |
| es | `amigoo` | `amigo` | no | `[{amigo 1} {amigó 1} {Amigo 1} {Amigó 1} {amigos 1}]` | 499.2 |
| es | `colorr` | `color` | no | `[{color 1} {colora 1} {colorar 1} {colore 1} {colorir 1}]` | 493.8 |
| es | `comerr` | `comer` | no | `[{comer 1} {comerá 1} {comeré 1} {cimera 2} {cimero 2}]` | 491.9 |
| es | `buenas nochess` | `buenas noches` | no | `[{buenas noches 1}]` | 95.8 |
| es | `por supuetso` | `por supuesto` | no | `[{por supuesto 1}]` | 100.3 |
| es | `hasta lueggo` | `hasta luego` | no | `[{hasta luego 1}]` | 101.2 |
| es | `zzqqxx` | none | no | `[]` | 490.7 |
| en | `cta` | `cat` | yes | `[{cat 1} {cat- 1} {CAT 1} {Cat 1} {TCA 1}]` | 366.3 |
| es | `papa` | `papá` | yes | `[{paca 1} {paga 1} {pagá 1} {paipa 1} {paja 1}]` | 288.1 |

Among 20 expected-answer cases with an approximately absent folded query, the expected answer appeared in the top five in all 20; 16 appeared first; `tomorrow` appeared third, `people` second, `green` fourth, and `gracias` second. Seven other expected-answer cases have an approximate existing match and therefore may not exercise miss-only suggestions. Both nonsense inputs returned no suggestions. The accented `papa` → `papá` example is ambiguous and folds to an existing query; it cannot test miss-only correction, and the prototype excludes distance-zero normalized candidates. `appel`, `beautifull`, `teh`, `neccessary`, `hous`, and `cta` likewise may already resolve in existing lookup, so their rankings are not counted as miss-path success. The output for `hous` did not include `house`; that is an important failure if this headword is present but fails to produce a definition in an actual lookup.

Approximate membership checks used case folding, NFD accent stripping and Unicode punctuation removal; they do not reproduce every detail of application matching and do not test actual definitions or network outcomes. Thus the 20 are potential misses, not proven app misses, and no assumption is made about whether a normal lookup reaches the suggestions path. The fixture set has no frequency weighting, is small, independently hand-authored and not fully representative. In particular the no-answer inputs do not show a general false-suggestion rate; common dictionary words, names and ambiguous senses may produce misleading results. The prototype makes no explicit safety guarantee for untrusted files and its 2 s and record caps are experimental. The run's elapsed times range from 95.8 to 905.7 ms, not cold or slow-machine bounds.

## Reproduction

The frozen inputs are listed in the table. While temporary files remain, run `GOCACHE=/tmp/wdict-codex-go-cache go run /tmp/wdict-independent-holdout.go`. The generated program uses the unchanged scanning and ranking functions from `/tmp/wdict-ranking-variant.go` and replaces only its main fixture loop. Fixture generation and complete output are in `/tmp/wdict-independent-holdout.json`, `/tmp/wdict-independent-holdout.go`, and `/tmp/wdict-independent-holdout-output.txt`. Expected results should not be considered a language-quality oracle or a release acceptance threshold.
