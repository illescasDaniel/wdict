# Technology options and proposed spec decisions

The initial design was approved in chat on 2026-10-26 with these changes: prioritize Linux, aim for macOS and Windows, focus on English and Spanish with definitions in the same language as the searched word, obtain dictionaries by download where feasible, consider online lookup with persistent cache, and aim for inflected forms and spelling suggestions. Tkinter is ruled out due to high-DPI concerns. The original diagrams and wireframe represent the earlier local-only proposal and need revision once the remaining decisions are agreed.

## GUI options

| Option | Strengths | Tradeoffs |
|---|---|---|
| **PySide6 + Qt Widgets** | Python throughout, native desktop controls, Qt's high-DPI support, mature threading and packaging tools. Good fit for a compact search window. | Qt adds a substantial runtime and packaging footprint. Check LGPL obligations for the final distribution. Startup and memory must be measured on target hardware. |
| **pywebview + Python + HTML/CSS/JS** | HTML styling and Python backend, system webview rather than bundling a browser, simple JS-to-Python bridge. | Additional frontend language and bridge. Linux needs GTK/WebKit or Qt dependencies, and rendering and packaging differ by platform. Webview choice affects size and display behavior. |
| **Wails + Go + HTML/CSS/JS** | Compiled Go backend and system webview, Go-to-JavaScript bindings; a strong fit if Go and web UI are preferred. | Two languages and a webview bridge; Linux GTK/WebKit runtime compatibility varies by distro. Wails v3 is newly stable and its v3.1 removes the GTK3 legacy option. Measure startup and scaling. |
| **Tauri + Rust + HTML/CSS/JS** | System webview and compiled Rust backend; attractive if small distribution and a web UI are priorities. | Rust toolchain and a second backend language; retaining Python adds an integration and distribution burden. Platform webview dependencies remain. |

Qt documents [high-DPI behavior](https://doc.qt.io/qtforpython-6.10/overviews/qtdoc-highdpi.html), [desktop deployment](https://doc.qt.io/qtforpython-6/deployment/index.html), and [licensing](https://doc.qt.io/qtforpython-6/). [pywebview](https://pywebview.flowrl.com/guide/) documents platform renderers and [dependencies](https://pywebview.flowrl.com/guide/installation); [Tauri](https://v2.tauri.app/concept/process-model/) documents its platform webviews. No startup time or package-size comparison has been measured for this project.

## Data feasibility

[Kaikki's English-Wiktionary extracts](https://kaikki.org/dictionary/rawdata.html) and [Spanish-Wiktionary extract](https://kaikki.org/eswiktionary/rawdata.html) are downloadable, with glosses in their respective edition languages. A downloaded edition can contain entries for many languages: the importer must filter by both edition and entry language. The Spanish extract page currently lists about 98 MB compressed and 1.1 GB uncompressed for all entries; sizes change with releases. English data is considerably larger. Coverage, available form links, correctness, attribution, licensing and practical import times still need hands-on evaluation. A usable multilingual online definition API has not yet been selected or verified; do not assume that a cache replaces an unvisited offline dictionary.

## Proposed decisions for the spec

1. **Success:** a user can search English or Spanish headwords and phrases, see same-language definitions and attribution, and repeat previously successful lookups offline. An installed dictionary permits new offline searches. Support forms and offer suggestions where data supports them, with representative tests rather than promising complete linguistic coverage.
2. **Failure handling:** start even when no dictionary or network is available. Show whether no entry exists or a source is unavailable; preserve the last good result and installed dictionary on download/import failure. Retry on user request. If offline and a matching cached response exists, display it with its source and saved date.
3. **Performance and resources:** propose a usable search window within 2 seconds of launch and local results within 250 ms on a documented reference Linux machine, excluding initial import and cold OS cache; benchmark before treating these as guarantees. Perform import and network work off the UI thread. No arbitrary dictionary size limit; report disk requirements before download and allow cancellation.
4. **Trust boundary:** treat search text, downloaded data, API responses and chosen paths as untrusted. Queries use parameterized database access. Network lookup transmits the query only when the user enables it; show the source. Cache and dictionaries stay in the user's application data directory, and replacements occur only after validation.
5. **Compatibility and rollout:** target Linux first, design for macOS and Windows but validate separately before claiming support. Store versioned local dictionary and cache formats separately from the application. No migration from an existing application is needed. Introduce online access as opt-in; permit reverting to local-only lookup without deleting data.

Open decisions: decide whether an installed offline dictionary is required for v1 or cache-only is an acceptable fallback; decide whether both languages, inflection and suggestions are v1 requirements or staged goals. Select and verify an online source before specifying its contract. Confirm or adjust all five proposed decisions above. Once these choices settle, update the design diagrams and wireframe states for each accepted flow before the spec.

## Wails assessment (2026-10-27)

Wails is a credible alternative to PySide6: use Go for lookup, SQLite, downloads and cache, and a small local HTML/CSS/JS frontend. Its [v3 desktop release](https://wails.io/blog/wails-v3-released/) is newly stable, so pin the version and evaluate the actual target Linux distribution. The [v3 platform contract](https://wails.io/status/) requires GTK4 and WebKitGTK 6.0 by default; its GTK3 / WebKit2GTK 4.1 compatibility option is slated for removal in v3.1. Windows needs WebView2. Verify runtime availability, startup, high-DPI and accessibility locally before committing. The user subsequently selected Wails with Go and SvelteKit; see the decision below.

## Selected stack and starter research (2026-10-27)

The user selected Wails, Go and SvelteKit. [Wails v3 includes a Svelte starter](https://v3.wails.io/contributing/template-system/) but its [changelog](https://wails.io/changelog/) records removal of the built-in SvelteKit starter. Svelte and SvelteKit are distinct; a Svelte starter alone does not satisfy the selected stack. The older Wails community listing includes SvelteKit starters, but their compatibility with v3 is unverified.

Proposed integration for implementation after the SDD gates: start with a pinned Wails v3 Svelte starter and adapt its frontend to SvelteKit, using [adapter-static](https://svelte.dev/docs/kit/adapter-static) for packaged frontend assets. Use Go services and Wails bindings for lookup, persistence and network access. Determine prerendering versus SPA fallback from actual routing needs and validate both development and packaged builds. SvelteKit server-only endpoints and SSR that require a Node server are not assumed to work in the packaged app. Verify the existing frontend build and asset paths before configuring the adapter. No template has been installed, built or evaluated locally.
