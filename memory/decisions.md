# Decisions

Append dated decisions with context, rationale and source references. Add a new entry when a decision is reversed.

## 2026-10-29 — Legacy cache attribution

The user approved preserving legacy cached definitions while showing a notice if a source or license link is missing (`src:517058a14965c942a353`). A durable note was queued but remains unscreened because the decision provider was unavailable.

## 2026-10-29 — Silent HTTP 502 responses

The user asked to treat HTTP 502 from the online dictionary lookup endpoints as silent errors for now (`src:fbb1669e08460c8e2421`). Retain stale fallback, do not cache failures as missing entries, and show a neutral message when no results are available.

## 2026-10-29 — Automatic bilingual lookup

Search both English and Spanish for each query without a language selector, as requested by the user (`src:34b97b8117fcb33761f4`); design approved (`src:58170218c984fa66a08b`); specification decisions approved (`src:e86c89718081e601095e`); spec approved (`src:6e99afe45172b4ac8823`); architecture approved (`src:dc4bf98b737711e4b089`).

## 2026-10-29 — Launch experiments

Experiments showed no demonstrated launch improvement and the user saw no difference from the placeholder; remove the experiments and retain the independent dictionary error fix (`src:92330b20cd5d33849f20`, `src:951f2ed31a886ccf2478`).

## 2026-10-27 — Dictionary direction

User approved the initial design with changes in chat on 2026-10-26: Linux first, aim for macOS and Windows, English and Spanish with same-language definitions, downloadable data where feasible, online cache as an option, and interest in inflections and suggestions. Tkinter is unsuitable for the user’s high-DPI setup; GUI stack and scope remain open. See `design/word-lookup/tech-stack-and-decisions.md`.

## 2026-10-27 — Wails and SvelteKit

User selected Wails with Go and asked to use SvelteKit. Wails v3 provides a Svelte starter but removed the built-in SvelteKit starter. Proposed approach is to adapt a pinned v3 starter using SvelteKit's static adapter; integration and packaging remain untested. See `design/word-lookup/tech-stack-and-decisions.md`.

## 2026-10-27 — First release scope

User agreed the five proposed decisions and clarified that offline dictionaries, response cache and both languages are nice to have in v1; inflections and spelling suggestions are deferred. The draft uses English online lookup as the required baseline and keeps provider interfaces ready for future sources. Source: user statement captured as `src:78ae2237acccde7265bf`.

## 2026-10-27 — Cache in v1

User explicitly requested both in-memory and on-disk caching to avoid unnecessary requests. This supersedes the earlier nice-to-have status for caching; the spec proposes freshness, limits and stale fallback for review. Spanish and downloaded offline dictionaries remain desirable but not required.

## 2026-10-27 — Offline dictionary distribution

User chose an in-app download for optional offline dictionaries, with go-stardict as the reader candidate; users need not install StarDict separately. Offline support remains desirable rather than mandatory for v1. Source: `src:901c3d41cff158d1b9e4`.

## 2026-10-27 — Required v1 dictionary download and spec approval

User approved the spec with the explicit change that in-app download of a curated dictionary is required for v1. This supersedes the previous optional-v1 scope; installing a dictionary remains optional for each user. Source: `src:d9ba2309b87de6de7a1c`.

## 2026-10-27 — Architecture approval and dependency management

User approved the download wireframe and architecture, and requested package managers for both Go and frontend dependencies with no global installations. Use Go modules and project-local npm. Source: `src:7451c617c3b69a94fb37`.

## 2026-10-27 — English offline dictionary candidate

Selected the xxyzz 20261018 English-English Wiktionary StarDict release as the initial catalog candidate. Its published checksum matches the downloaded archive and its metadata states CC BY-SA 4.0; real offline word and phrase lookups succeeded. Source evidence: `src:1425520ad562497ff9d1`. Attribution and content rights still need a full release review, and cold index construction is slow.

## 2026-10-27 — Practical attribution

User prefers a quick, pragmatic review and source/license attribution in the README and app, rather than prolonged rights research. Source: `src:7277bd055edce9e9b478`.

- 2026-10-27: Use the pinned Wails v3.0.0 AppImage generator for GTK 3 packaging because it bundles and relocates WebKit helpers. Its linuxdeploy continuous download is not pinned; cross-distribution verification and license review remain before release.

## 2026-10-27 — Contributor task runner

Use Task as a pinned Go tool to expose setup, build, test and AppImage commands without a global Task installation. Its shared Go module dependencies may also update application dependencies.

## 2026-10-27 — Publication direction

Use version 0.1.0 and MIT for the project source, retaining separate third-party and dictionary content notices. Build Linux AppImages in CI; defer macOS and Windows binaries and the first release until platform testing.

## 2026-10-28 — CI and releases

Run routine checks without generating AppImages; a separately dispatched workflow builds the AppImage and creates its tag and GitHub release. Preserve the deferred first release until macOS and Windows testing (source `src:a8efe69c530e6bb4f0af`).

## 2026-10-28 — Audit and first release

User approved failing npm audit for moderate or higher findings while displaying low findings, and explicitly requested the v0.1.0 release after CI and Go installation work (`src:18d6de4a6d7390db8ac3`).

## 2026-10-28 — Cache freshness

Keep online results fresh for 365 days before attempting a refresh; retain stale fallback when the provider is unavailable. User requested v0.2.0 and installation after publication (`src:bc2bcf9382dacdc66077`).

## 2026-10-28 — Appearance

Use System by default with footer icons for System, Light and Dark; remember the selection and follow system changes. The user approved the design and authorized automatic spec and architecture approval for this feature (`src:f0c8c5907b9be0a2eada`).
