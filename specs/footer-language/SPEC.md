# Fixed footer and interface language

## Metadata

- **Feature:** Keep the footer visible and let users choose English or Spanish for the interface.
- **Wireframe:** [Footer and interface language](../../wireframes/footer-language.html), approved 2026-11-04 (chat; `src:43ce9da020fb43286fd5`).
- **Design decisions approved:** 2026-11-04 (chat; `src:b100cb1de6d73aa010df`).
- **Spec approved:** 2026-11-04 (chat; `src:cb4e45d81b39aa662285`).
- **Related:** Existing bilingual lookup and appearance features remain in effect.

## Design decisions

### Success criteria

- On lookup and offline dictionary views, the footer remains visible at the bottom while content scrolls. All content and keyboard focus remain reachable at desktop and narrow sizes.
- English and Español switch all app-owned interface text, including current statuses and errors, accessible names, titles and document language, on the next render. Preserve query, results, current view and ongoing download state.
- Display app-generated dates using the selected interface language. Dictionary headwords, definitions and source-provided content retain their supplied language. Each lookup still searches both dictionaries.

### Failure handling

- An absent, invalid or unreadable saved preference falls back to English. A failed write keeps the selected language for the current session without blocking or adding a notification.
- Existing lookup, metadata and download failures and retry behavior remain available and are displayed in the selected language, including messages already visible when switching.

### Performance & resource budget

- Bundle translations locally; switching language causes no network or backend request and adds no polling. Store one small preference independently of other settings. Target an update on the next UI render, with under 100 ms under normal conditions as a goal, not a hard guarantee. Do not change lookup or download performance requirements.

### Trust boundary

- Treat persisted locale as untrusted: accept only `en` and `es`, falling back to English otherwise. The bundled translation catalog is trusted. Preserve existing escaping and link validation for untrusted dictionary content and URLs. Switching languages sends no query or content to a service.

### Compatibility & rollout

- Existing users start in English. Store the preference independently; preserve appearance, dictionary and cache data. No API change or dictionary/cache migration is required. Older versions may ignore the new preference; rollback does not require data migration. Rebuild and include embedded frontend assets.

## Triggers & routing

- **Entry:** Open either existing view and select English or Español in the footer.
- **Exit:** The current view remains open in the chosen language, with the selection restored on subsequent launches when storage is available.

## Visual & UI rules

- Follow the approved wireframe: a footer at the bottom of the window, with Language/Idioma and the existing Appearance/Apariencia controls; content scrolls independently without being hidden under it. Narrow layouts wrap controls as shown.
- Use a labelled native dropdown with options named English and Español; keep the existing appearance control and its keyboard operation. Translate labels, buttons, placeholders, messages, progress, attribution labels, titles and screen-reader text. Do not ship reviewer-only notes.
- Update document language and app-generated date formatting with the selected language. The cached-result fetch date currently shown is included.

## Acceptance criteria

### Scenario: Default language

**Given** no saved language preference **When** the app starts **Then** the interface is English and the selector shows English.

### Scenario: Switch language on either view

**Given** either view is open **When** the user selects Español and then English **Then** app-owned visible text, current status and error messages, accessible names, titles and document language update on each switch without navigation or an extra request.

### Scenario: Preserve active state

**Given** a query, results, current view or an ongoing dictionary download **When** the language changes **Then** those states remain intact and progress continues.

### Scenario: Remember the preference

**Given** the user selects Español and storage works **When** the app restarts **Then** Español is selected and the interface is Spanish.

### Scenario: Invalid or unavailable preference

**Given** a saved value other than `en` or `es`, or a storage read failure **When** the app starts **Then** it safely uses English and remains usable.

### Scenario: Preference write fails

**Given** preference storage cannot be written **When** the user selects Español **Then** the interface switches for the current session without blocking or a new notification.

### Scenario: Localized dates

**Given** a cached result with a displayed fetch date **When** the interface language changes **Then** the date is formatted according to the selected interface language.

### Scenario: Dictionary content and bilingual lookup

**Given** either interface language **When** a word is searched **Then** both English and Spanish dictionaries are checked and returned headwords, definitions and source-provided content remain unchanged.

### Scenario: Scrolling and narrow windows

**Given** long content on either view at desktop or narrow window sizes **When** the user scrolls or navigates by keyboard **Then** the footer remains visible and all content and focusable controls can be reached without being obscured.

### Scenario: Appearance controls

**Given** either interface language **When** the user operates the appearance radiogroup with keyboard or pointer **Then** its existing selection, focus behavior and persistence continue to work and its labels are translated.

### Scenario: Existing errors and retries

**Given** a lookup, dictionary metadata or download error **When** the user switches language or retries **Then** the current app-owned error and retry controls appear in the selected language and existing retry behavior remains usable.

### Scenario: Untrusted dictionary content

**Given** dictionary content or links containing unsafe input **When** results are displayed in either interface language **Then** existing escaping and link validation continue to apply.

## Out of scope

- Translating dictionary definitions or source-provided content; changing which dictionaries are searched.
- New dictionary providers, appearance modes, backend APIs or cache migrations.
