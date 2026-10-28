<script>
  import { onMount } from "svelte";
  import { call } from "../lib-api.js";
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
      if (!downloadActive || nextProgress?.State === "downloading" || nextProgress?.State === "installing")
        progress = nextProgress;
    } catch (e) {
      error = "Unable to load dictionary information.";
    }
  }
  onMount(() => {
    queryInput?.focus();
    refresh();
    let timer = setInterval(() => {
      if (progress?.State === "downloading" || progress?.State === "installing")
        refresh();
    }, 500);
    return () => clearInterval(timer);
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
            {#if entry.Source.Attribution} · {entry.Source.Attribution}{/if}{#if safeLink(entry.Source.URL)}
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
</main>

<style>
  :global(*) {
    box-sizing: border-box;
  }
  :global(body) {
    margin: 0;
    background: #eef2ef;
    color: #21352a;
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
    border-bottom: 1px solid #cad7cd;
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
    color: #50685a;
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
    color: #607467;
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
    border: 1px solid #aabdaf;
    border-radius: 7px;
    padding: 11px 12px;
    font: inherit;
    background: white;
  }
  button {
    font: inherit;
    font-weight: 650;
    border: 1px solid #245d42;
    border-radius: 7px;
    padding: 10px 15px;
    background: #245d42;
    color: white;
    cursor: pointer;
  }
  .link {
    border: 0;
    padding: 0;
    background: none;
    color: #245d42;
    text-align: left;
  }
  a {
    color: #245d42;
  }
  button:focus-visible,
  input:focus-visible,
  a:focus-visible {
    outline: 3px solid #174633;
    outline-offset: 3px;
  }
  .hint,
  .source {
    font-size: 13px;
    color: #50685a;
  }
  .notice {
    background: #e3eee5;
    border-radius: 8px;
    padding: 14px;
    margin-top: 18px;
  }
  .notice p {
    margin: 5px 0;
  }
  article {
    background: white;
    border: 1px solid #d2e0d5;
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
    background: #edf3ee;
    border-radius: 20px;
    padding: 4px 9px;
    white-space: nowrap;
  }
  .source {
    border-top: 1px solid #e5eee7;
    padding-top: 12px;
    margin-bottom: 0;
  }
  li {
    white-space: pre-wrap;
    margin-bottom: 8px;
  }
  .error {
    background: #fff1ea;
    color: #713d27;
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
