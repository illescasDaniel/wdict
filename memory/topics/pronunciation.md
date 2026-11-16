# Pronunciation
Updated: 2026-11-16. Related: [lookup-and-ranking](lookup-and-ranking.md), [releases-and-packaging](releases-and-packaging.md)

## Now
- IPA is read from the offline Wiktionary data and shown as a line with flag-emoji accent labels, first card per word only ([ipa-flag-labels-first-card-only](../notes/ipa-flag-labels-first-card-only.md)); extraction takes only `IPA:` list items ([ipa-extraction-ipa-list-items-only](../notes/ipa-extraction-ipa-list-items-only.md)). A missing IPA is never borrowed across capitalization ([ipa-not-borrowed-across-capitalization](../notes/ipa-not-borrowed-across-capitalization.md)).
- Play buttons speak through two engines: Piper neural voices (preferred) and eSpeak-NG, bundled in the AppImage as a separate process ([audio-pronunciation-feature](../notes/audio-pronunciation-feature.md), [espeak-gpl-separate-process](../notes/espeak-gpl-separate-process.md), [espeak-bundled-in-appimage](../notes/espeak-bundled-in-appimage.md)). The Piper runtime budget is 250 MB ([piper-runtime-budget-250mb](../notes/piper-runtime-budget-250mb.md)).
- The shown IPA is normalized for the engine in `NormalizePiperIPA`: notation stripped ([piper-ipa-notation-normalization](../notes/piper-ipa-notation-normalization.md)), default stress injected and untrained phones folded or rejected ([piper-default-stress-and-untrained-phones](../notes/piper-default-stress-and-untrained-phones.md)), initial schwa to ɐ ([piper-initial-schwa-folds-to-a-turned](../notes/piper-initial-schwa-folds-to-a-turned.md)), unknown diacritics dropped but unknown base phones abort ([piper-unknown-diacritics-dropped-base-phones-abort](../notes/piper-unknown-diacritics-dropped-base-phones-abort.md)), British əʊ kept split ([piper-gb-diphthong-kept-split](../notes/piper-gb-diphthong-kept-split.md)).
- Missing pronunciations are partly upstream data ([upstream-dictionary-data-gaps](../notes/upstream-dictionary-data-gaps.md)).

## How it got here
- 2026-11-11 flags and first-card rule chosen ([ipa-flag-labels-first-card-only](../notes/ipa-flag-labels-first-card-only.md)).
- 2026-11-13 audio added; SSML phoneme speech was silent, so it spoke the headword, then the shown IPA through a converter ([espeak-ssml-silent-plain-text-headword](../notes/espeak-ssml-silent-plain-text-headword.md), [ipa-to-espeak-mnemonic-converter](../notes/ipa-to-espeak-mnemonic-converter.md)).
- 2026-11-14 repairs from user reports (vegetable/table, appearance, Sábado, eight): [piper-model-input-must-match-phonemizer](../notes/piper-model-input-must-match-phonemizer.md), [piper-schwa-upsilon-fold-to-ou](../notes/piper-schwa-upsilon-fold-to-ou.md) (later wrong, see [piper-gb-diphthong-kept-split](../notes/piper-gb-diphthong-kept-split.md)), [vits-synthesis-stochastic-cache-permanent](../notes/vits-synthesis-stochastic-cache-permanent.md).

## Open
Listening quality and native-window playback are user checks only ([verification-off-screen-and-silent](../notes/verification-off-screen-and-silent.md), [real-window-checks-left-open](../notes/real-window-checks-left-open.md)).
