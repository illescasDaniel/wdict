# Changelog

## [0.9.0] - 2026-11-23

- Add French: definitions from French Wiktionary, an optional offline French dictionary, a French interface, and French pronunciation with an optional Piper voice and an eSpeak fallback.
- After a French search finds nothing, suggest possible base forms (such as plurals) and spelling matches from the installed French dictionary.
- Search English, Spanish and French independently, so one language failing no longer hides the others.
- Fix searches with leading or trailing spaces, tabs or newlines being rejected: surrounding whitespace is now trimmed.
- Show the pronunciation "no recording" and "could not synthesise" messages in Spanish when the interface is in Spanish.

## [0.8.3] - 2026-11-15

- Fix the UI freezing (and no sound in the AppImage) when playing a pronunciation: sound now plays through the system audio player (`pw-play`, `paplay`, `aplay` or `ffplay`) instead of the web view, and the original MP3 recordings are no longer downloaded.
- Move the scrollbar to the window edge and remove the duplicate scrollbar.
- Show the unavailable-pronunciation control as one compact line, and show records with identical pronunciation together.
- Stop stylesheet text from appearing in Spanish definitions.

## [0.8.2] - 2026-11-15

- Show the source attribution once per card, shortened.

## [0.8.1] - 2026-11-15

- Fix pronunciation voices dropping the start of words like `appearance` or `eight` and speaking noise for unsupported sounds (for example the Indian-English `ʈ`): transcriptions without stress marks now keep their vowel, and unsupported phones speak the headword instead.

## [0.8.0] - 2026-11-14

- Play pronunciations aloud using the displayed pronunciation, with optional natural neural voices from Piper.

## [0.7.1] - 2026-11-13

- Fix duplicated offline results: the same entry no longer appears once per matching synonym alias when the dictionary lists aliases repeatedly or an alias matches the word itself (for example `dogfood`).

## [0.7.0] - 2026-11-11

- Show IPA pronunciations under the headword from the offline English and Spanish dictionaries and the online sources. Accent names appear as flags (with the original name on hover), the line shows on the first card of a word only, and a generic pronunciation has no flag.

## [0.6.1] - 2026-11-09

- Add `wdict --version` (and `-v`) and show the app version in the footer.

## [0.6.0] - 2026-11-06

- Suggest offline alternatives, inflected forms and spelling corrections (including phrase headwords) when a lookup misses.
- Order lookup results by the selected interface language.

## [0.5.0] - 2026-11-04

- Keep the footer visible and add English and Spanish interface language selection.
- Localize application dates and translated interface text.

## [0.4.0] - 2026-10-29

- Search English and Spanish together, with independent online results and downloadable Spanish offline dictionary support.
- Keep older cached entries available and indicate when source or license links are missing.
- Treat online provider HTTP 502 responses as silent temporary failures, using stale cached definitions when available.
- Reject ambiguous Spanish dictionary responses and improve dictionary installation and recovery coverage.

## [0.3.0] - 2026-10-28

- Add System, Light and Dark appearance options in the footer; System follows the device setting.
- Remember the selected appearance between launches.

## [0.2.0] - 2026-10-28

- Keep online dictionary results fresh for 365 days, with stale fallback when the provider is unavailable.
- Focus the search field on startup and clarify Linux installation options.
- Verify cached results persist across app restarts.

## [0.1.0] - 2026-10-28

- Initial Linux-first desktop dictionary with online lookup and optional offline English Wiktionary data.
- Linux AppImage packaging. macOS and Windows support remains to be tested.
