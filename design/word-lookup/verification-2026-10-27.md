# Provider and packaging verification — 2026-10-27

These checks are evidence for development, not release certification. The app now builds and smoke-launches on Linux. The observed live results can change.

## Online provider terms and attribution

- The [Free Dictionary API site](https://dictionaryapi.dev/) describes its service as free. Its [repository](https://github.com/meetDeveloper/freeDictionaryAPI) describes costs and requests donations. I did not find a formal service-level agreement, usage policy or documented quota; free does not guarantee availability, unlimited use, or future terms. The sample `hello` response included `x-ratelimit-limit: 450`, but it did not establish the window, policy or future limit.
- The provider maintainer's [2021 source announcement](https://github.com/meetDeveloper/freeDictionaryAPI/discussions/101) describes Wiktionary as the source and discusses commercial use. It is historical, not an agreement or proof of current terms for every returned item.
- Live responses on this date identified the entry source as Wiktionary and returned `license.name` = `CC BY-SA 3.0`, `license.url` = `https://creativecommons.org/licenses/by-sa/3.0`, and one or more `sourceUrls`. The response license metadata must be preserved per entry, alongside all source URLs. CC's [3.0 terms summary](https://creativecommons.org/licenses/by-sa/3.0/) calls for appropriate attribution, a license link and applicable share-alike terms; whether and how change notices apply depends on the use. [Wikimedia's terms](https://foundation.wikimedia.org/wiki/Policy:Terms_of_Use) describe a link to the page as one attribution method for contributor text. Content and licensing may differ between entries; do not fabricate missing metadata. English Wiktionary's [copyright page](https://en.wiktionary.org/wiki/Wiktionary:Copyrights) now describes CC BY-SA 4.0/GFDL for original text and separately licensed material; this differs from the API's 3.0 metadata. Check the exact source/version and all applicable notices rather than silently overriding it. This is not a legal clearance.
- The provider's repository [license file](https://github.com/meetDeveloper/freeDictionaryAPI/blob/master/LICENSE) is GPL v3; the API server source-code license is distinct from the license of response content. No conclusion about the license of this application follows merely from using the endpoint.
- The original adapter preserved only the first source URL and license name, discarding the license URL and additional sources. `core.Source` and the adapter now retain these fields, with a mapping test. The UI, cache validation and fallback for absent or varying metadata are not yet verified and must be completed before release.
- The selected offline catalog entry is the [xxyzz English Wiktionary StarDict release 20261018](https://github.com/xxyzz/wiktionary_stardict/releases/tag/20261018). Its published SHA-256 is `db0a650d0b707c82c6f495e8f9b6e122f9c4e10cb5551322591e35921e8b4765`; the 142,487,509-byte archive matched it. The included `.ifo` identifies a 2026-10-01 Wiktionary snapshot, CC BY-SA 4.0, compiled by xxyzz. The catalog is pinned; updates require a reviewed new release and checksum. Real `hello` and `look up` offline lookups succeeded, but broad phrase coverage and complete attribution/rights review remain open.

## Live phrase sample

Method: on 2026-10-27 use `curl` with a 12-second timeout and URL-encoded input against `https://api.dictionaryapi.dev/api/v2/entries/en/<query>`. Each result below is one request, not a performance, reliability or coverage study. A first Python urllib run received 403 for every query; subsequent curl requests succeeded, so clients and network conditions should be tested independently. Do not interpret those initial 403s as missing words.

| HTTP status | Queries |
| --- | --- |
| 200 | `hello`; `look up`; `give up`; `take off`; `put up with`; `break down`; `turn on`; `make up`; `look forward to`; `kick the bucket` |
| 404 | `get along with`; `run out of`; `a stitch in time` |

This is 8 of 10 chosen phrasal-verb expressions, plus one word and two idioms/proverbs. Success verifies a nonempty API response for those queries, not definition quality or comprehensive phrase support. On other networks and at other times results may differ. A 404 means the API returned no definitions for the exact expression; it does not prove the phrase is absent from every source.

## SvelteKit + Wails Linux smoke test

- Host: CachyOS Linux x86_64; Go `1.27.1-X:nodwarf5`; installed GTK3 `3.24.52` and WebKit2GTK `2.54.1` (4.1 API). `pkg-config` did not find GTK4/WebKitGTK 6.0 together. No globally installed package or Wails CLI was used.
- In `/tmp/wdict-verify`, using the project-local `frontend/node_modules`, built a minimal prerendered SvelteKit page with `@sveltejs/adapter-static` 3.0.10 and SvelteKit 2.70.3. `npm run build` produced static HTML and JS.
- A disposable Go module pinned Wails `v3.0.0`, embedded the generated `build` directory via `embed.FS` and `fs.Sub`, and served it with Wails `application.BundledAssetFileServer`. `go build -buildvcs=false -tags gtk3` succeeded. A five-second launch on the desktop succeeded; Wails logged requests for `/` and the generated SvelteKit JS files before timed shutdown. The sandboxed GUI launch could not access the display or D-Bus; the authorized outside-sandbox launch was used instead.
- This verifies a minimal production-mode asset build and native window path on this machine, not the actual project's app, frontend-to-Go service bindings, interaction, high-DPI/accessibility, startup target, installable package, or macOS/Windows. No screen-level visual assertion was made. There is no Node server at runtime in this test.
- [Wails platform status](https://v3.wails.io/status/) documents GTK4/WebKitGTK 6.0 as the default, and GTK3/WebKit2GTK 4.1 as a legacy v3.0.x option to be removed in v3.1. The [Linux packaging guide](https://v3.wails.io/guides/build/linux/) distinguishes runtime dependencies. Verify default GTK4, v3.1 migration, distributable packaging and each other OS separately. The [SvelteKit static adapter documentation](https://svelte.dev/docs/kit/adapter-static) describes prerendering and static output.

## User display check

On 2026-10-27, the user reported that text and controls were readable at their usual high-DPI scaling, keyboard focus was visible, and the window remained usable while a lookup was pending. They also reported duplicate download errors and unreadable `bword://` links in offline results; those issues are being addressed separately. This is a user-reported check on their Linux display, not cross-platform verification.
