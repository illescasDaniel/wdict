# Decisions

Append dated decisions with context, rationale and source references. Add a new entry when a decision is reversed.

## 2026-11-04 — Footer and interface language design

The user approved the fixed-footer and English/Spanish interface-language wireframe (`src:43ce9da020fb43286fd5`).
The user also agreed to the proposed spec decisions, including formatting app-generated dates in the selected interface language (`src:b100cb1de6d73aa010df`).
The user approved the spec (`src:cb4e45d81b39aa662285`).
The user chose a dedicated JavaScript internationalization library (`src:a99625ad49fc3a442cfc`) and approved the i18next architecture (`src:24f6db6819cfaaa41e94`). The architecture approval note remains pending screening.

## 2026-10-29 — Legacy cache attribution

The user approved preserving legacy cached definitions while showing a notice if a source or license link is missing (`src:517058a14965c942a353`). A durable note was queued but remains unscreened because the decision provider was unavailable.

## 2026-10-29 — Silent HTTP 502 responses

The user asked to treat HTTP 502 from the online dictionary lookup endpoints as silent errors for now (`src:fbb1669e08460c8e2421`). Retain stale fallback, do not cache failures as missing entries, and show a neutral message when no results are available.

## 2026-10-29 — Automatic bilingual lookup

Search both English and Spanish for each query without a language selector, as requested by the user (`src:34b97b8117fcb33761f4`); design approved (`src:58170218c984fa66a08b`); specification decisions approved (`src:e86c89718081e601095e`); spec approved (`src:6e99afe45172b4ac8823`); architecture approved (`src:dc4bf98b737711e4b089`).

## 2026-10-29 — Launch experiments

Experiments showed no demonstrated launch improvement and the user saw no difference from the placeholder; remove the experiments and retain the independent dictionary error fix (`src:92330b20cd5d33849f20`, `src:951f2ed31a886ccf2478`).

## 2026-10-27 — Dictionary direction

User approved the initial design with changes in chat on 2026-10-26: Linux first, aim for macOS and Windows, English and Spanish with same-language definitions, downloadable data where feasible, online cache as an option, and interest in inflections and suggestions. Tkinter is unsuitable for the user’s high-DPI setup; GUI stack and scope remain open. See `design/word-lookup/tech-stack-and-decisions.md`.

## 2026-10-27 — Wails and SvelteKit

User selected Wails with Go and asked to use SvelteKit. Wails v3 provides a Svelte starter but removed the built-in SvelteKit starter. Proposed approach is to adapt a pinned v3 starter using SvelteKit's static adapter; integration and packaging remain untested. See `design/word-lookup/tech-stack-and-decisions.md`.

## 2026-10-27 — First release scope

User agreed the five proposed decisions and clarified that offline dictionaries, response cache and both languages are nice to have in v1; inflections and spelling suggestions are deferred. The draft uses English online lookup as the required baseline and keeps provider interfaces ready for future sources. Source: user statement captured as `src:78ae2237acccde7265bf`.

## 2026-10-27 — Cache in v1

User explicitly requested both in-memory and on-disk caching to avoid unnecessary requests. This supersedes the earlier nice-to-have status for caching; the spec proposes freshness, limits and stale fallback for review. Spanish and downloaded offline dictionaries remain desirable but not required.

## 2026-10-27 — Offline dictionary distribution

User chose an in-app download for optional offline dictionaries, with go-stardict as the reader candidate; users need not install StarDict separately. Offline support remains desirable rather than mandatory for v1. Source: `src:901c3d41cff158d1b9e4`.

## 2026-10-27 — Required v1 dictionary download and spec approval

User approved the spec with the explicit change that in-app download of a curated dictionary is required for v1. This supersedes the previous optional-v1 scope; installing a dictionary remains optional for each user. Source: `src:d9ba2309b87de6de7a1c`.

## 2026-10-27 — Architecture approval and dependency management

User approved the download wireframe and architecture, and requested package managers for both Go and frontend dependencies with no global installations. Use Go modules and project-local npm. Source: `src:7451c617c3b69a94fb37`.

## 2026-10-27 — English offline dictionary candidate

Selected the xxyzz 20261018 English-English Wiktionary StarDict release as the initial catalog candidate. Its published checksum matches the downloaded archive and its metadata states CC BY-SA 4.0; real offline word and phrase lookups succeeded. Source evidence: `src:1425520ad562497ff9d1`. Attribution and content rights still need a full release review, and cold index construction is slow.

## 2026-10-27 — Practical attribution

User prefers a quick, pragmatic review and source/license attribution in the README and app, rather than prolonged rights research. Source: `src:7277bd055edce9e9b478`.

- 2026-10-27: Use the pinned Wails v3.0.0 AppImage generator for GTK 3 packaging because it bundles and relocates WebKit helpers. Its linuxdeploy continuous download is not pinned; cross-distribution verification and license review remain before release.

## 2026-10-27 — Contributor task runner

Use Task as a pinned Go tool to expose setup, build, test and AppImage commands without a global Task installation. Its shared Go module dependencies may also update application dependencies.

## 2026-10-27 — Publication direction

Use version 0.1.0 and MIT for the project source, retaining separate third-party and dictionary content notices. Build Linux AppImages in CI; defer macOS and Windows binaries and the first release until platform testing.

## 2026-10-28 — CI and releases

Run routine checks without generating AppImages; a separately dispatched workflow builds the AppImage and creates its tag and GitHub release. Preserve the deferred first release until macOS and Windows testing (source `src:a8efe69c530e6bb4f0af`).

## 2026-10-28 — Audit and first release

User approved failing npm audit for moderate or higher findings while displaying low findings, and explicitly requested the v0.1.0 release after CI and Go installation work (`src:18d6de4a6d7390db8ac3`).

## 2026-10-28 — Cache freshness

Keep online results fresh for 365 days before attempting a refresh; retain stale fallback when the provider is unavailable. User requested v0.2.0 and installation after publication (`src:bc2bcf9382dacdc66077`).

## 2026-10-28 — Appearance

Use System by default with footer icons for System, Light and Dark; remember the selection and follow system changes. The user approved the design and authorized automatic spec and architecture approval for this feature (`src:f0c8c5907b9be0a2eada`).

## 2026-11-06 — Lookup alternatives direction

Show alternatives only after a language misses, require the corresponding offline dictionary, investigate suitable morphology datasets before falling back to partial regular-form coverage, and include multiword spelling corrections (`src:6f16694188c7fc53cc4a`). The user approved the full design (`src:c34610a226566b0738a4`).

The user likes the lookup-alternatives wireframe and requested planning a separate search over dictionary phrase headwords for typo correction (`src:23d2651f2c57ed43c790`). 

The user confirmed the five proposed lookup-alternatives specification decisions, including measurement before numerical resource budgets and dataset selection (`src:de40e77c3089de7f6007`).

The user approved the lookup-alternatives spec (`src:665fce98e49af19f680d`). Before implementation, resolve its missing-dictionary state and optional morphology-index fallback alongside measured resource budgets and dataset selection.

On 2026-11-06 the user approved the lookup-alternatives core architecture (`src:fd9eb7586fb1e9d78b2b`); the associated memory note is pending screening.

On 2026-11-06 the user approved the partial lookup-alternatives spec amendment (`src:8a02e642d2cb6cd1bb7f`); the memory note is pending screening.

On 2026-11-06 the user approved the final lookup-alternatives amendment: existing dictionaries, bounded inflection rules, spelling ranking and resource caps (`src:6dce82b0c9d386831cec`).

On 2026-11-06 the user approved treating 64 MiB as a measured shared-heap target rather than a strict per-feature heap cap (`src:bb20bbd7cfc98225d3b0`).

## 2026-11-06 — Lookup alternatives implementation choices

Implemented without a morphology dataset: UniMorph was investigated but no suitable distribution was found for the first release, so bounded regular-form rules (suffix stripping for English and Spanish) are used instead, as approved by the final amendment. Bounded Unicode optimal-string-alignment distance (OSA ≤ 2, which counts adjacent transpositions as one edit) with deterministic tie-breaking ranks spelling candidates; only headwords verified in the installed index are returned. The 64 MiB heap cap is treated as a measured shared-heap target rather than a hard per-feature limit (`src:bb20bbd7cfc98225d3b0`). `timed_out` and `unavailable` alternatives statuses are shown with distinct UI messages. No persistent derived index is written.

## 2026-11-06 — v0.6.0 release process

The Release workflow (`workflow_dispatch`, input version) validates `VERSION`, creates the tag and publishes the release, so tags are never pushed by hand. A hand-pushed `v0.6.0` was deleted before any release or proxy fetch and recreated by the workflow on the same commit `ad7cfee`. After a version bump, `frontend/build` must be rebuilt, not sed-edited, because its hash changes and CI checks it.

## 2026-11-09 — Version reporting

The version is `-ldflags -X main.version=$(cat VERSION)` when built via Task or the AppImage script; `go install ...@vX.Y.Z` falls back to build info (leading `v` trimmed); anything else reports `dev`. The footer label comes from the Go `App.Version()` binding, hidden when unavailable. User asked for this after finding no way to see the installed version (v0.6.0 had no flag).

## 2026-11-11: SDD workflow is now `sdd-html-feature`

User decision: features go through the `sdd-html-feature` skill (one page `sdd/<feature>.html`, global install) instead of `sdd-feature` (`specs/`, `wireframes/`, `design/`). Existing `specs/`, `wireframes/`, `design/` stay as history. Rerun the loadout with `--without sdd` (it would otherwise reinstall `sdd-feature` and its AGENTS.md block); the AGENTS.md SDD section no longer carries loadout markers. Loadout 0.6.0 was applied with `--force --without sdd --no-hooks` (Entire hooks left untouched).

## 2026-11-11: Phonetic pronunciation decisions

- Shown as a line under the headword. Accent labels become flag emoji (General American/US 🇺🇸, Received Pronunciation/UK 🇬🇧, Australia 🇦🇺, New Zealand 🇳🇿, Canada 🇨🇦, Ireland 🇮🇪, South Africa 🇿🇦); compound labels split at commas; if any part is unknown the whole label stays as written. The original text is kept in `Pronunciation.LabelTitle` (hover text and accessible name). User chose flags over "US"/"UK" text.
- The line is shown only on the first card of a word; a later card with the same headword (case-insensitive), same language and identical list shows none (offline English has one card per part of speech). User chose this over repeating it.
- Spec R3 (user: "map flags first and then check length"): the 32-rune limit applies to the label as shown, after mapping; the original may be up to 128 runes.
- Offline English extraction: only `<li>` whose own text starts with `IPA:`, skipping `<ol>` (definitions) and Rhymes items. Rejected: any `IPA`-class span (took rhyme and definition IPA, e.g. `diktat`).
- Old cached text labels are re-mapped to flags on disk-cache read (`DiskCache.Get` normalizes); no cache version bump. `MemoryCache` does not normalize.
- Verified 2026-11-11 by tests, QA in Chromium with a stub runtime and the saved fixtures in `internal/adapters/testdata/pronunciation`; not verified in the real window.

## 2026-11-13: Offline duplicate results fixed (alias handling)

The user reported identical duplicated result cards when searching `dogfood` (screenshot). Root cause, verified 2026-11-13 against the installed xxyzz/wiktionary_stardict 20261018 en-en dictionary: the `.syn` file lists matching aliases repeatedly (`dogfood` twice) and includes an alias that folds to the target word itself (`dog-food` → the `dogfood` index entry), while `scanDictionary` in `internal/adapters/offline.go` appended the index word once per matching alias — `Find("dogfood")` returned 6 entries in 3 identical pairs. Fix: each index entry is appended at most once (`matches(word) || aliases[i] > 0`, edit id `a3355d73`); alias targets still resolve to their original headword. Regression test `TestGivenRepeatedSynonymAliasesWhenScannedThenEachIndexEntryIsReturnedOnce` (`internal/adapters/offline_format_test.go`) failed before and passes after; quality gate `go tool task test` green the same day (svelte-check 0/0, 70 frontend tests, Go tests with gtk3 tag). StarDict data is untrusted and stays that way: `.idx` legitimately repeats ~45k headwords with distinct offsets (separate articles), so dedupe is per index position, not per headword. Shipped in v0.7.1 (tag `v0.7.1`, release run 43461376019) and reinstalled locally (`/home/daniel/go/bin/wdict` 0.7.1).

## 2026-11-13: Audio pronunciation decisions

- Feature "Play pronunciations aloud" (`sdd/audio-pronunciation.html`): design, spec and architecture approved in chat; Q1/Q2/Q4 answered in the user's inline review comments (codec failure → eSpeak fallback, not plugin messaging; English default voice en-us, Spanish es; bundle eSpeak-NG if its license allows); Q3 defaulted to per-item play buttons at design approval ("design approved", "spec approved", "architecture approved").
- Verify amendments approved with "Okay let's go with your recommendations": C6 gains the translated playback-failed copy for the webview double-failure case ("Pronunciation playback failed. Try again." / "No se pudo reproducir la pronunciación. Vuelve a intentarlo."); C7 hover wording is "source and attribution" (the sources supply attribution strings, not a license name field). No code change needed.
- G4 bundling decision (license review 2026-11-13, web sources): eSpeak-NG is GPLv3; the app calls the **unmodified standalone CLI** as a separate process, which upstream maintainers call the legally correct way, and GPL-3.0 §5 makes the AppImage an aggregate — the GPL does not extend to the MIT app. Conditions when packaging: ship the GPLv3 COPYING text, name espeak-ng, and provide a source offer (upstream repo + pinned version). Do not link against libespeak-ng (that would combine works). Packaging implementation still pending.
- Build additions carried in the Build tab: PlayRequest.Headword (F4 retry data), PlayResult.Voice (C7 hover upgrade "eSpeak-NG · voice" after a voice play), F5 fallback capped at one attempt.

- IPA→mnemonic converter facts (verified 2026-11-13, espeak-ng 1.52.0, delegated coder + oracle tests): es ɣ→**Q** (`[[G]]` is invalid in es; the engine's own "luego" is `lu'eQo`); bare `[[A]]` is invalid in en (ɑ needs the ɑː→A: pair); every vowel mnemonic accepts `:` lengthening so a generic ː merge is safe; `I@` echoes `I;@`; isolated voiced consonants synthesize silence but are audible in vowel context, so audibility tests must use full streams; `-q -x` echo is the parse oracle; mnemonic input `[[…]]` is audible WITHOUT `-m` in raw text mode. Full tables live in `internal/core/playback.go`; 94/94 entries oracle-verified; revision 3 of `sdd/audio-pronunciation.html` (user: the exact transcription matters → converter approved).
- **Speech-notation facts (verified 2026-11-14, root cause of the "a \<word\>" and Enrique reports):** Wiktionary IPA uses the tie-barred affricates `d͡ʒ`/`t͡ʃ` (U+0361 between the letters — a plain-digraph table misses them), syllable dots inside transcriptions, syllabicity marks (ɫ̩) and PRECOMPOSED nasal vowels (ẽ = U+1EBD, category Ll — a Unicode-Mn strip alone misses it). Piper voices: the `.` token (id 10) is a real ~0.15 s segment inside words (with isolated schwas this reads as "a \<word\>"), the tie and ẽ are silently dropped (Enrique → "ŋrike"); ˈ and ː resolve fine. Fix: `stripNonPhonemic` (core) strips dots + Mn marks and maps ẽĩõũã → plain vowels before both Piper and the mnemonic converter. es_ES-sharvard map HAS r/ɾ/ŋ; dictionaryapi.dev audio URLs return Cloudflare HTML to plain clients (recordings fail → voice fallback).

## 2026-11-11: Phonetic pronunciation, revision R4 (found with "vegetable")


- The qualifier words "dialectal" and "obsolete" are ignored when mapping an accent label to flags (other non-accent words such as "rhotic" and merger names stay as text). A label that maps to both 🇬🇧 and 🇺🇸 is the generic form and shows no label. User: "Approved, start the build. Yes ignore the obsolete too". Reason: the real offline entry lists "Received Pronunciation, General American, Canada", "UK, dialectal", "US, Canada, dialectal"; the earlier rule flagged the common form and left the other two as text.
- Source counts (my own scan of the installed English dictionary 2026-11-11, not independently checked): about 4,500 UK+US labels, "dialectal" 252 uses, "obsolete" 346.
- Real fixture: `internal/adapters/testdata/pronunciation/offline_en_vegetable.html`.
- Open: the approved 5-run median performance report was not re-measured for R3/R4 (grader G8); the user has not yet accepted or asked for it.
