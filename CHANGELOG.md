# Changelog

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
