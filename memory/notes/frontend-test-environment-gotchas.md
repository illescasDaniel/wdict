---
name: frontend-test-environment-gotchas
description: Frontend test traps: jsdom has no media play(), no media error event, route tests need the lib-api mock, and Svelte whitespace
type: gotcha
date: 2026-11-13
status: current
supersedes: 
superseded_by: 
source: 
---
Verified 2026-11-13 in this repo: jsdom's `HTMLMediaElement.play()` is unimplemented (inject an audio factory);
testing-library `fireEvent` has no media `error` event (dispatch `new Event("error")`); new route test files need the
`vi.mock("../lib-api.js")` block or they fail with `call.mockReset is not a function`; and newline-separated whitespace text
nodes inside `{#each}` or `<p>` break exact-match `textContent` tests (Svelte whitespace rule), so keep those on one line.
Playback later moved to the host audio player (v0.8.3), but the test setup is unchanged.
