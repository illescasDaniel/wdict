<script>
  import { onMount } from "svelte";
  import { call } from "../lib-api.js";
  import {
    readAppearance,
    saveAppearance,
    resolveAppearance,
  } from "../appearance.js";
  let appearance = "system";
  let systemDark = false;
  function applyAppearance() {
    document.documentElement.dataset.theme = resolveAppearance(
      appearance,
      systemDark,
    );
  }
  function modeKey(event) {
    const values = ["system", "light", "dark"];
    if (
      !["ArrowLeft", "ArrowRight", "ArrowUp", "ArrowDown"].includes(event.key)
    )
      return;
    event.preventDefault();
    const step = event.key === "ArrowLeft" || event.key === "ArrowUp" ? -1 : 1;
    const next =
      values[
        (values.indexOf(appearance) + step + values.length) % values.length
      ];
    selectAppearance(next);
    event.currentTarget.parentElement
      .querySelector(`[aria-label="${next[0].toUpperCase() + next.slice(1)}"]`)
      .focus();
  }
  function selectAppearance(value) {
    appearance = value;
    applyAppearance();
    try {
      saveAppearance(window.localStorage, value);
    } catch {}
  }
  let queryInput;
  let query = "",
    page = "lookup",
    busy = false,
    status = "",
    entries = [],
    origin = "",
    fetched = "",
    stale = false,
    dictionary = null,
    progress = null,
    downloadActive = false,
    refreshSeq = 0,
    error = "",
    seq = 0;
  async function refresh() {
    const id = ++refreshSeq;
    try {
      const nextDictionary = (await call("Dictionaries"))[0];
      const nextProgress = await call("Progress");
      if (id !== refreshSeq) return;
      dictionary = nextDictionary;
      if (
        !downloadActive ||
        nextProgress?.State === "downloading" ||
        nextProgress?.State === "installing"
      )
        progress = nextProgress;
    } catch (e) {
      error = "Unable to load dictionary information.";
    }
  }
  onMount(() => {
    try {
      appearance = readAppearance(window.localStorage);
    } catch {
      appearance = "system";
    }
    const media = window.matchMedia?.("(prefers-color-scheme: dark)");
    systemDark = media?.matches ?? false;
    const onAppearanceChange = (event) => {
      systemDark = event.matches;
      applyAppearance();
    };
    media?.addEventListener("change", onAppearanceChange);
    applyAppearance();
    queryInput?.focus();
    refresh();
    let timer = setInterval(() => {
      if (progress?.State === "downloading" || progress?.State === "installing")
        refresh();
    }, 500);
    return () => {
      clearInterval(timer);
      media?.removeEventListener("change", onAppearanceChange);
    };
  });
  async function lookup() {
    const id = ++seq;
    busy = true;
    status = "Searching…";
    try {
      let result = await call("Lookup", query);
      if (id !== seq) return;
      busy = false;
      if (result.Status === "found") {
        entries = result.Entries;
        origin = result.Origin;
        fetched = result.FetchedAt;
        stale = result.Stale;
        status = "";
      } else
        status =
          {
            invalid_query: "Enter a word or phrase of up to 256 characters.",
            not_found: "No entry found. Try another word or phrase.",
            source_unavailable:
              "The dictionary is unavailable. Check your connection and try again.",
          }[result.Status] || "Unable to complete the lookup.";
    } catch (e) {
      if (id === seq) {
        busy = false;
        status = "The dictionary is unavailable. Please try again.";
      }
    }
  }
  async function download() {
    if (downloadActive) return;
    downloadActive = true;
    error = "";
    progress = {
      State: "downloading",
      BytesRead: 0,
      TotalBytes: dictionary.DownloadBytes,
    };
    try {
      await call("Download", dictionary.ID);
      downloadActive = false;
      await refresh();
    } catch (e) {
      downloadActive = false;
      error =
        "Couldn’t install the dictionary. Check your connection and available storage, then try again.";
      await refresh();
    }
  }
  function safeLink(url) {
    try {
      let u = new URL(url);
      return u.protocol === "https:" || u.protocol === "http:" ? u.href : "";
    } catch {
      return "";
    }
  }
</script>

<svelte:head><title>wdict</title></svelte:head>
<main>
  <header>
    <strong>wdict</strong>{#if page === "lookup"}<button
        class="link"
        onclick={() => {
          page = "dictionaries";
          refresh();
        }}>Offline dictionaries</button
      >{:else}<button class="link" onclick={() => (page = "lookup")}
        >← Look up</button
      >{/if}
  </header>
  {#if page === "lookup"}
    <h1>Look up a word</h1>
    <p>Find English definitions for words and phrases.</p>
    <form
      onsubmit={(e) => {
        e.preventDefault();
        lookup();
      }}
    >
      <label for="query">Word or phrase <span>English</span></label>
      <div class="search">
        <input
          id="query"
          bind:this={queryInput}
          bind:value={query}
          placeholder="Enter a word or phrase"
          autocomplete="off"
        /><button type="submit">Look up</button>
      </div>
    </form>
    <p class="hint">
      If an offline dictionary or saved entry cannot answer your search, it may
      be sent to the online dictionary.
    </p>
    {#if status}<div class="notice" role="status">{status}</div>{/if}
    {#if entries.length}{#each entries as entry}<article>
          <h2>{entry.Headword}</h2>
          {#if origin !== "offline"}<p class="hint">
              {stale
                ? "Stale saved result"
                : origin === "online"
                  ? "Online result"
                  : "Saved result"}{fetched
                ? " · " + new Date(fetched).toLocaleDateString()
                : ""}
            </p>{/if}{#each entry.Meanings as meaning}<section>
              {#if meaning.PartOfSpeech}<h3>{meaning.PartOfSpeech}</h3>{/if}
              <ol>
                {#each meaning.Definitions as definition}<li>
                    {definition}
                  </li>{/each}
              </ol>
            </section>{/each}
          <p class="source">
            Source: {entry.Source.Name}
            {#if entry.Source.Attribution}
              · {entry.Source.Attribution}{/if}{#if safeLink(entry.Source.URL)}
              · <a
                href={safeLink(entry.Source.URL)}
                target="_blank"
                rel="noreferrer">Source</a
              >{/if}{#if safeLink(entry.Source.LicenseURL)}
              · <a
                href={safeLink(entry.Source.LicenseURL)}
                target="_blank"
                rel="noreferrer">License</a
              >{/if}{#each entry.Source.AdditionalURLs || [] as url}{#if safeLink(url)}
                · <a href={safeLink(url)} target="_blank" rel="noreferrer"
                  >More information</a
                >{/if}{/each}
          </p>
        </article>{/each}{:else if !busy}<div class="notice">
        <strong>Look up words offline</strong>
        <p>Download an English dictionary for offline definitions.</p>
        <button
          class="link"
          onclick={() => {
            page = "dictionaries";
            refresh();
          }}>View offline dictionaries →</button
        >
      </div>{/if}
  {:else}<h1>Offline dictionaries</h1>
    <p>Download a dictionary to look up words without a connection.</p>
    {#if dictionary}<article>
        <div class="title">
          <h2>{dictionary.Name}</h2>
          <span class="badge"
            >{dictionary.Installed
              ? "Installed"
              : progress?.State === "installing"
                ? "Installing"
                : progress?.State === "downloading"
                  ? "Downloading"
                  : "Not installed"}</span
          >
        </div>
        <p class="hint">
          Language: English definitions<br />Source:
          <a
            href={dictionary.Source.AdditionalURLs[0]}
            target="_blank"
            rel="noreferrer">Wiktionary · xxyzz</a
          ><br />License:
          <a
            href={dictionary.Source.LicenseURL}
            target="_blank"
            rel="noreferrer">{dictionary.License}</a
          >
          · Download: {(dictionary.DownloadBytes / 1000000).toFixed(0)} MB
        </p>
        {#if progress?.State === "downloading"}<progress
            aria-label="Dictionary download"
            max={progress.TotalBytes}
            value={progress.BytesRead}
          ></progress>
          <p class="hint">
            {(progress.BytesRead / 1000000).toFixed(1)} MB of {(
              progress.TotalBytes / 1000000
            ).toFixed(0)} MB
          </p>{:else if progress?.State === "installing"}<p class="notice">
            Checking and installing the dictionary…
          </p>{:else if error}<button onclick={download}>Try again</button
          >{:else if dictionary.Installed}<p class="notice">
            Ready to use offline.
          </p>{:else}<button onclick={download}>Download</button
          >{/if}{#if error}<p class="error" role="alert">{error}</p>{/if}
      </article>
      <p class="hint">
        You can continue looking up words while it downloads.
      </p>{:else}<p role="status">Loading dictionary information…</p>{/if}{/if}
  <footer>
    <span>Appearance</span>
    <div class="modes" role="radiogroup" aria-label="Appearance">
      <button
        type="button"
        class="mode"
        aria-label="System"
        title="System"
        role="radio"
        aria-checked={appearance === "system"}
        tabindex={appearance === "system" ? 0 : -1}
        onkeydown={modeKey}
        onclick={() => selectAppearance("system")}
        ><svg viewBox="0 0 24 24" aria-hidden="true"
          ><rect x="3" y="4" width="18" height="13" rx="2" /><path
            d="M8 21h8M12 17v4"
          /></svg
        ></button
      >
      <button
        type="button"
        class="mode"
        aria-label="Light"
        title="Light"
        role="radio"
        aria-checked={appearance === "light"}
        tabindex={appearance === "light" ? 0 : -1}
        onkeydown={modeKey}
        onclick={() => selectAppearance("light")}
        ><svg viewBox="0 0 24 24" aria-hidden="true"
          ><circle cx="12" cy="12" r="4" /><path
            d="M12 2v2M12 20v2M4.93 4.93l1.42 1.42M17.65 17.65l1.42 1.42M2 12h2M20 12h2M4.93 19.07l1.42-1.42M17.65 6.35l1.42-1.42"
          /></svg
        ></button
      >
      <button
        type="button"
        class="mode"
        aria-label="Dark"
        title="Dark"
        role="radio"
        aria-checked={appearance === "dark"}
        tabindex={appearance === "dark" ? 0 : -1}
        onkeydown={modeKey}
        onclick={() => selectAppearance("dark")}
        ><svg viewBox="0 0 24 24" aria-hidden="true"
          ><path
            d="M20.5 14.2A8.5 8.5 0 0 1 9.8 3.5a8.5 8.5 0 1 0 10.7 10.7Z"
          /></svg
        ></button
      >
    </div>
  </footer>
</main>

<style>
  :global(html) {
    --bg: #eef2ef;
    --text: #21352a;
    --muted: #50685a;
    --border: #cad7cd;
    --input-border: #aabdaf;
    --surface: white;
    --accent: #245d42;
    --button-text: white;
    --notice: #e3eee5;
    --badge: #edf3ee;
    --rule: #e5eee7;
    --error-bg: #fff1ea;
    --error-text: #713d27;
    --focus: #174633;
  }
  :global(html[data-theme="dark"]) {
    color-scheme: dark;
    --bg: #17221b;
    --text: #e4eee7;
    --muted: #b2c5b7;
    --border: #415548;
    --input-border: #607a69;
    --surface: #202f25;
    --accent: #a5d6b4;
    --button-text: #172a1e;
    --notice: #25372b;
    --badge: #344a3b;
    --rule: #415548;
    --error-bg: #4b3027;
    --error-text: #ffd6c2;
    --focus: #a5d6b4;
  }
  footer {
    margin-top: 40px;
    padding-top: 14px;
    border-top: 1px solid var(--border);
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    color: var(--muted);
    font-size: 13px;
  }
  .modes {
    display: flex;
    gap: 2px;
    padding: 3px;
    border: 1px solid var(--border);
    border-radius: 9px;
  }
  button.mode {
    width: 36px;
    height: 34px;
    padding: 0;
    display: grid;
    place-items: center;
    border: 0;
    border-radius: 6px;
    background: transparent;
    color: var(--muted);
  }
  button.mode[aria-checked="true"] {
    background: var(--surface);
    color: var(--text);
    box-shadow: 0 1px 3px #0003;
  }
  .mode svg {
    width: 19px;
    height: 19px;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.8;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
  :global(*) {
    box-sizing: border-box;
  }
  :global(body) {
    margin: 0;
    background: var(--bg);
    color: var(--text);
    font:
      16px/1.55 system-ui,
      sans-serif;
  }
  main {
    max-width: 760px;
    margin: auto;
    padding: 28px;
  }
  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid var(--border);
    padding-bottom: 16px;
    margin-bottom: 30px;
  }
  header strong {
    font-size: 21px;
  }
  h1 {
    font-size: 27px;
    line-height: 1.2;
    margin: 0 0 8px;
  }
  h2 {
    font-size: 21px;
    margin: 0;
  }
  h3 {
    font-size: 15px;
    color: var(--muted);
    font-style: italic;
  }
  p {
    margin: 8px 0 16px;
  }
  form {
    margin-top: 26px;
  }
  label {
    display: block;
    font-weight: 650;
    margin-bottom: 6px;
  }
  label span {
    color: var(--muted);
    font-weight: 400;
    font-size: 13px;
  }
  .search {
    display: flex;
    gap: 8px;
  }
  input {
    flex: 1;
    min-width: 0;
    border: 1px solid var(--input-border);
    border-radius: 7px;
    padding: 11px 12px;
    font: inherit;
    background: var(--surface);
  }
  button {
    font: inherit;
    font-weight: 650;
    border: 1px solid var(--accent);
    border-radius: 7px;
    padding: 10px 15px;
    background: var(--accent);
    color: var(--button-text);
    cursor: pointer;
  }
  .link {
    border: 0;
    padding: 0;
    background: none;
    color: var(--accent);
    text-align: left;
  }
  a {
    color: var(--accent);
  }
  button:focus-visible,
  input:focus-visible,
  a:focus-visible {
    outline: 3px solid var(--focus);
    outline-offset: 3px;
  }
  .hint,
  .source {
    font-size: 13px;
    color: var(--muted);
  }
  .notice {
    background: var(--notice);
    border-radius: 8px;
    padding: 14px;
    margin-top: 18px;
  }
  .notice p {
    margin: 5px 0;
  }
  article {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 10px;
    margin-top: 20px;
    padding: 20px;
  }
  .title {
    display: flex;
    align-items: start;
    justify-content: space-between;
    gap: 10px;
  }
  .badge {
    font-size: 12px;
    background: var(--badge);
    border-radius: 20px;
    padding: 4px 9px;
    white-space: nowrap;
  }
  .source {
    border-top: 1px solid var(--rule);
    padding-top: 12px;
    margin-bottom: 0;
  }
  li {
    white-space: pre-wrap;
    margin-bottom: 8px;
  }
  .error {
    background: var(--error-bg);
    color: var(--error-text);
    padding: 12px;
    border-radius: 8px;
  }
  progress {
    width: 100%;
    accent-color: #307a53;
  }
  @media (max-width: 460px) {
    main {
      padding: 18px;
    }
    .search {
      flex-wrap: wrap;
    }
    .search input {
      min-width: 180px;
    }
  }
</style>
