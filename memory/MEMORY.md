# Memory

Code at HEAD wins for current behaviour; notes win for why and history; newest current note wins; open `source` sessions for detail.

Converted 2026-11-16 from the agemem project memory (the old files and `.agemem/` were deleted the same day; git history keeps them).

## Topics
- [Pronunciation](topics/pronunciation.md) 21 notes: IPA display, audio playback, eSpeak and Piper voices
- [Lookup and ranking](topics/lookup-and-ranking.md) 16 notes on offline dictionaries, alternatives, ranking and cache
- [Releases and packaging](topics/releases-and-packaging.md) 10 notes: release workflow, AppImage, versions, CI

## Decisions
- [French support includes automatic lookup, a French interface, a French Piper voice and bounded morphology alternatives](notes/french-language-support-scope.md) 2026-11-21
- [AppImage is built with the pinned Wails v3.0.0 generator plus scripts/build-appimage.sh; l](notes/appimage-wails3-generator.md) 2026-10-27
- [wdict is a Linux-first desktop dictionary built with Wails v3, Go and SvelteKit, English a](notes/dictionary-app-direction.md) 2026-10-27
- [Offline dictionaries are xxyzz Wiktionary StarDict releases (20261018); first lookup uses](notes/offline-dictionary-xxyzz-stardict.md) 2026-10-27
- [v1 requires in-memory and on-disk caching and an in-app download of a curated offline dict](notes/v1-offline-download-and-cache-required.md) 2026-10-27
- [Appearance defaults to System with footer icons for System, Light and Dark; the choice is](notes/appearance-system-default.md) 2026-10-28
- [Online results stay fresh for 365 days; a stale result is used when the provider fails](notes/cache-freshness-365-days.md) 2026-10-28
- [CI and release fail on moderate or higher npm audit findings and only display low ones](notes/npm-audit-fails-on-moderate.md) 2026-10-28
- [0.8.4 release held: npm audit fails on picomatch with no patched version yet](notes/picomatch-advisory-blocks-0-8-4.md) 2026-11-18
- [0.9.0 published with French support; rebuild frontend/build after a version bump](notes/release-0-9-0-published.md) 2026-11-23
- [Releases come from a manually dispatched workflow that creates the tag; never push v* tags](notes/release-via-dispatched-workflow.md) 2026-10-28
- [v0.1.0 was released Linux-only; macOS and Windows binaries stay deferred](notes/v0-1-0-linux-first-release.md) 2026-10-28
- [Startup experiments (prerender, window geometry, deferred metadata) showed no gain and wer](notes/launch-experiments-rejected.md) 2026-10-29
- [Legacy cached definitions are kept and shown with a notice when their source or licence li](notes/legacy-cache-attribution-notice.md) 2026-10-29
- [HTTP 502 from the online endpoints is a silent error: keep stale results, do not cache a m](notes/online-failures-keep-stale-cache.md) 2026-10-29
- [Interface language uses the i18next library; app-generated dates follow the interface lang](notes/i18n-library-i18next.md) 2026-11-04
- [The 64 MiB heap figure for lookup alternatives is a measured shared-heap target, not a per](notes/alternatives-heap-target-not-hard-cap.md) 2026-11-06
- [Alternatives appear only after a miss, need the offline dictionary, and use bounded rules](notes/lookup-alternatives-design-choices.md) 2026-11-06
- [wdict --version comes from ldflags main.version, then the go install module version, else](notes/version-reporting-ldflags.md) 2026-11-09
- [Offline English IPA comes only from <li> items whose own text starts with 'IPA:', skipping](notes/ipa-extraction-ipa-list-items-only.md) 2026-11-11
- [Pronunciation line shows flag emoji for accents, only on the first card of a word, label l](notes/ipa-flag-labels-first-card-only.md) 2026-11-11
- [Features go through the sdd-html-feature skill (one page sdd/<feature>.html); the loadout](notes/sdd-html-feature-workflow.md) 2026-11-11
- [Pronunciations play through a host audio player, not the webview audio element](notes/audio-native-host-player.md) 2026-11-15
- [Audio pronunciation: per-item play buttons; en-us/es default voices; codec failure falls b](notes/audio-pronunciation-feature.md) 2026-11-13
- [eSpeak-NG may be bundled because it runs as an unmodified separate CLI process; never link](notes/espeak-gpl-separate-process.md) 2026-11-13
- [The voice path speaks the shown IPA through a pure IPAToMnemonics converter, oracle-checke](notes/ipa-to-espeak-mnemonic-converter.md) 2026-11-13
- [Inject default primary stress before each word's first vowel and fold or reject phones the](notes/piper-default-stress-and-untrained-phones.md) 2026-11-14
- [Piper runtime disk budget is at most 250 MB installed; the measured piper-tts 1.8.0 runtim](notes/piper-runtime-budget-250mb.md) 2026-11-14
- [The Piper script drops missing Unicode-Mn diacritics but still aborts on unknown base phon](notes/piper-unknown-diacritics-dropped-base-phones-abort.md) 2026-11-14
- [The AppImage injects the unmodified espeak-ng CLI and its data dir; main.go sets ESPEAK_DA](notes/espeak-bundled-in-appimage.md) 2026-11-15
- [Never fill a missing IPA for one capitalization from another (Car vs car, Polish vs polish](notes/ipa-not-borrowed-across-capitalization.md) 2026-11-15
- [Offline results are ranked in tiers: exact raw headword, then folded match, then alias; in](notes/offline-exact-headword-ranking.md) 2026-11-15
- [British əʊ is kept as separate ə and ʊ for en_GB-alba; only the US voice uses clustered oʊ](notes/piper-gb-diphthong-kept-split.md) 2026-11-15
- [Real-window checks stay open unless doable without new installs (no Xvfb; the user runs Wa](notes/real-window-checks-left-open.md) 2026-11-15
- [Result ranking: case-aware tiers, source-card grouping, labeled spelling reading; Design,](notes/result-ranking-card-grouping-design.md) 2026-11-15
- [Compact senses are grouped in the frontend from an optional Meaning.SenseGroup, not by mer](notes/sense-grouping-presentation-only.md) 2026-11-15

## Preferences
- [User wants a quick licensing review with source and licence attribution in the README and](notes/attribution-pragmatic-review.md) 2026-10-27
- [Dependencies come from Go modules and project-local npm with no global installs; Task runs](notes/dependencies-project-local.md) 2026-10-27
- [After a release is published, update the user's installed wdict with go install and confir](notes/update-system-install-after-release.md) 2026-10-28
- [Each search checks English and Spanish automatically; the user does not want a language se](notes/bilingual-lookup-without-selector.md) 2026-10-29
- [Agent verification must stay off screen and silent (no GUI windows, no audio out); the use](notes/verification-off-screen-and-silent.md) 2026-11-14

## Incidents
- [Suggestion search was CPU-bound in alternativeDistance and hit its 2 s deadline; a banded](notes/alternatives-distance-cpu-hotspot.md) 2026-11-06
- [SSML <phoneme> rendered digitally silent WAVs; speech was changed to the plain headword wi](notes/espeak-ssml-silent-plain-text-headword.md) 2026-11-13
- [Repeated .syn aliases made scanDictionary return each entry several times (dogfood); fixed](notes/offline-duplicate-results-aliases.md) 2026-11-13
- [English word-initial schwa was near-silent in Piper; fold it to ɐ. Mid-word ə stays](notes/piper-initial-schwa-folds-to-a-turned.md) 2026-11-14

## Gotchas
- [Frontend test traps: jsdom has no media play(), no media error event, route tests need the](notes/frontend-test-environment-gotchas.md) 2026-11-13
- [go vet reports a copylocks finding in installer_lifecycle_test.go:273 that is not from new](notes/go-vet-integration-copylocks-preexisting.md) 2026-11-13
- [In T3 Code with opencode-provider models, html_render attachments never display; paste the](notes/html-render-silent-failure-opencode-provider.md) 2026-11-13
- [OpenCode ignores .mcp.json; project MCP servers must be listed in opencode.json](notes/opencode-needs-its-own-mcp-config.md) 2026-11-13
- [Wiktionary IPA uses tie bars, dots, syllabicity marks and precomposed nasal vowels that Pi](notes/piper-ipa-notation-normalization.md) 2026-11-14
- [Raw [[IPA]] for Piper skips vowel-cluster merging; inspect actual model phones, not just a](notes/piper-model-input-must-match-phonemizer.md) 2026-11-14
- [Piper VITS output is stochastic per run and cached for good; a garbled draw replays foreve](notes/vits-synthesis-stochastic-cache-permanent.md) 2026-11-14
- [spanishMeanings drops dt labels and nested lists, so labels like 'Rugby' or 'Uso: coloquia](notes/spanish-senses-drop-dt-labels.md) 2026-11-15
- [Some pronunciation gaps are upstream data (quiso, many common headwords); the 'sing -> thi](notes/upstream-dictionary-data-gaps.md) 2026-11-15

## History (superseded)
- [First draft of v1 scope: offline dictionaries, response cache and Spanish were n](notes/v1-scope-offline-and-cache-optional.md) 2026-10-27 → superseded by v1-offline-download-and-cache-required
- [Plan of 2026-10-27: hold the first release until macOS and Windows binaries were](notes/first-release-deferred-until-all-platforms.md) 2026-10-27 → superseded by v0-1-0-linux-first-release
- [əʊ was folded to oʊ everywhere because only aɪ aʊ eɪ oʊ ɔɪ were believed trained](notes/piper-schwa-upsilon-fold-to-ou.md) 2026-11-14 → superseded by piper-gb-diphthong-kept-split
