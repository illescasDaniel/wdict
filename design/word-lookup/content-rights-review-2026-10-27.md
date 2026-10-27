# Content rights review — 2026-10-27

This is a technical source-and-notice review, not a legal opinion or a guarantee of rights in every dictionary entry. It covers the current English sources only.

## Offline Wiktionary package

- The app downloads xxyzz's pinned [20261018 English-English StarDict archive](https://github.com/xxyzz/wiktionary_stardict/releases/tag/20261018), verifies its published SHA-256 and uses its `.ifo` metadata, which identifies an English Wiktionary snapshot dated 2026-10-01, CC BY-SA 4.0, and compiler xxyzz. The app does not distribute a copy of this archive itself; the user downloads it from the publisher.
- [English Wiktionary's copyright page](https://en.wiktionary.org/wiki/Wiktionary:Copyrights) describes licensing of its original text and warns that other material may have different rights. [Wikimedia's terms](https://foundation.wikimedia.org/wiki/Policy:Terms_of_Use) explain that a link to the reused page can attribute its contributors; imported text may require additional notices, and its page history and discussion may need review.
- [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/) permits sharing and adaptation with attribution, a license link and an indication of changes; share-alike applies to distributed adaptations. The app displays the entry's Wiktionary page URL, credits contributors and xxyzz, links the license and identifies its extraction/reformatting. The app selects definition text and strips HTML and other sections, so this notice is meaningful; it is not a complete analysis of whether every transformation creates an adaptation.
- **Unresolved:** The publisher's package-level notice does not prove the source and attribution of every entry or every imported passage. We have not compared the package's complete contents against per-page histories, notices and licenses, nor established whether the compiled package preserves every applicable upstream notice. If publishing or redistributing content at scale, investigate the upstream snapshot and its transformation/notice preservation and seek professional legal review as appropriate. Do not assert universal clearance.

## Online provider

- [Free Dictionary API](https://github.com/meetDeveloper/freeDictionaryAPI) describes the service and its Wiktionary data; its published response supplies license and source URLs. On 2026-10-27 sampled responses provided CC BY-SA 3.0 and Wiktionary page URLs. The app preserves the per-response license and URLs, and displays them; it does not silently replace the provider's 3.0 notice with 4.0.
- **Unresolved:** Current Wiktionary licensing and the sampled provider metadata use different version labels. We have not established each API result's exact source revision, any imported third-party notices, or the provider's terms for this particular planned distribution. A free API does not by itself establish rights for every response. Resolve provenance with authoritative information before claiming blanket clearance; retain the provider's original notices in the meantime.

## Scope and release implications

These obligations concern dictionary content and notices; no conclusion is made here about the license of the app's own code. Future dictionary packages, Spanish sources, media, and redistribution of extracted data need their own review. The user does not need to perform legal research to run the app locally. A public release that redistributes content or depends on blanket licensing assurances may warrant a qualified reviewer after the concrete provenance questions above are investigated.
