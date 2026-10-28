# Appearance

## Metadata

UI design: [wireframe](../../wireframes/appearance.html). Design approved 2026-10-28; user authorized automatic approval of spec and architecture on 2026-10-28.

## Design decisions

- Success criteria: the footer offers System, Light and Dark and applies the selected palette throughout both screens. System is the default and follows changes to the device setting.
- Failure handling: invalid or unreadable saved preference defaults to System; if saving fails, the selection works for the current session.
- Performance & resource budget: appearance changes immediately, without network access or additional dependencies.
- Trust boundary: local storage is untrusted; accept only the three known values. No new network or filesystem access.
- Compatibility & rollout: an absent preference defaults to System; existing dictionary and cache data remain unchanged. The preference is stored independently in browser local storage.

## Triggers & routing

On startup restore a valid preference or use System. A footer selection changes appearance immediately and is saved when possible. System tracks device changes. Both screens share the control.

## Visual & UI rules

Follow the approved wireframe: labeled Appearance footer with monitor, sun and moon icon choices, accessible names, selected state and visible keyboard focus. Keep the footer usable at narrow widths. Light and dark palettes cover text, controls, notices, cards and errors.

## Acceptance criteria

### Scenario: default and system changes
**Given** no valid saved preference, **When** the app opens or the device appearance changes, **Then** System is selected and the palette follows the device.

### Scenario: explicit selection and restore
**Given** a user selects Light or Dark, **When** the device setting changes or the app reopens, **Then** the selected appearance remains active.

### Scenario: unavailable storage
**Given** preference storage fails, **When** the user changes appearance, **Then** the current session still changes and remains usable.

### Scenario: footer navigation
**Given** either screen, **When** the user selects a mode using mouse or keyboard, **Then** the named selected control reflects it and the same choice remains on the other screen.

## Out of scope

Custom themes, changes to dictionary data, and syncing preferences across devices.
