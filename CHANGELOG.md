# Changelog

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
