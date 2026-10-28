# Decisions

Append dated decisions with context, rationale and source references. Add a new entry when a decision is reversed.

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
