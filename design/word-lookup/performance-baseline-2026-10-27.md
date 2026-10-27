# Linux performance baseline — 2026-10-27

## Environment and method

Reference machine: CachyOS Linux 7.3.0-3-cachyos, x86-64, Intel Core i9-13900HX (32 logical CPUs), 31 GiB RAM, Go 1.27.1-X:nodwarf5, Wails 3.0.0 using GTK3 and WebKit2GTK 4.1 (WebKitGTK 2.54.1). Measurements were made on a running desktop with other software active; CPU governor reported 51% scaling at the time. No attempt was made to clear OS caches. Results are a local baseline, not guarantees for other systems.

The spec's targets are a usable window within 2 seconds and local operations within 250 ms, excluding cold OS cache. There is no agreed numeric memory or CPU ceiling. See `specs/word-lookup/SPEC.md`.

## Native startup

Launched the built `./wdict` three times with empty, temporary `XDG_DATA_HOME` and `XDG_CACHE_HOME`, so no offline dictionary or existing responses were loaded. A Python observer used monotonic time from process creation to the Wails asset-server log reporting the root HTML and SvelteKit page JavaScript requests; sampling resolution was approximately 50 ms. These are **asset delivery proxies**, not a measurement of a painted, interactive window. Each instance ran for four seconds and was then terminated. The desktop was visible, but no automated screenshot or interaction established the exact ready time.

| Run | Root HTML request | Page JS request | Sum of observed process RSS at ~3 s | Cumulative observed CPU at ~3 s |
| --- | ---: | ---: | ---: | ---: |
| 1 | 255 ms | 255 ms | 398.3 MiB | 0.38 s |
| 2 | 208 ms | 208 ms | 397.0 MiB | 0.35 s |
| 3 | 240 ms | 240 ms | 404.9 MiB | 0.36 s |

The observer sampled the app and its then-visible descendants through `/proc`; three processes were visible at 3 seconds. Summed RSS double-counts shared memory and can omit processes not parented under the app, so it is not unique memory consumption. CPU is the sum of observed process CPU ticks at that instant, not whole-system CPU percentage or an exact total including already-exited children. Startup with an installed dictionary and actual first-interactive timing are not measured here. These samples do not establish the 2-second usable-window target.

## Local lookup and cache

A separate Go harness called the production adapters with the verified English snapshot extracted under `/tmp/wdict-dictionary`; the process was fresh and its in-memory index empty. The files had been used earlier, so OS cache was likely warm. The first call constructs the StarDict in-memory index. Subsequent calls reused the same `OfflineDictionary` instance. These are backend durations and do not include Wails messaging or rendering.

| Operation | Time |
| --- | ---: |
| First offline `hello` (fresh process, likely warm OS cache) | 8.681 s |
| Second offline `hello` | 0.930 ms |
| Subsequent offline `look up` | 0.363 ms |
| Subsequent offline missing word | 0.085 ms |
| Memory cache `Get`, mean / maximum over 1,000 reads | 0.067 µs / 1.269 µs |
| Disk cache `Get`, one-entry cache, mean / maximum over 1,000 reads | 9.126 µs / 20.89 µs |

The standalone lookup process used 37.942 CPU seconds in user mode and 0.791 in system mode, with 336,824 KiB peak RSS across its lifetime, measured using `getrusage(RUSAGE_CHILDREN)` by its parent. CPU time can exceed wall time because the process uses multiple cores. Go-reported live heap after the first lookup was 160.8 MiB; it is not total process memory. This is a separate harness, not the GUI's peak memory.

A second harness generated 1,000 valid small disk-cache records in a temporary directory, then called the production `LookupService` 20 times with memory cache and offline source disabled, including each disk hit's access-time rewrite and eviction scan. Mean was **10.73 ms**, maximum **13.82 ms**, with a likely warm OS cache. This is a synthetic full-cache test, not a representative sample of large records or the UI round trip.

## Interpretation

- The observed first offline lookup **misses the 250 ms target**: index creation takes 8.681 seconds even though the OS cache was likely warm. Repeated backend lookups and the synthetic full disk-cache hits in these samples were below 250 ms, but that does not prove all local operations meet the target.
- The startup asset requests arrived well before 2 seconds in these three runs; actual window usability was not timed, so the startup target remains unverified.
- More representative measurements should include real interactive ready/lookup timing, larger cache records and cold OS cache, plus macOS and Windows. The user's prior live check confirms that the window remained usable during a pending lookup; it did not time startup.

## Subsequent scan-based lookup experiment

After the baseline, the offline adapter was changed to scan the StarDict `.syn` and `.idx` files with the reader's public scanners instead of building its full in-memory index. It folds case, accents, punctuation and whitespace for matching and caches up to 128 successful offline queries in memory. In three fresh standalone processes on the same machine, with likely warm OS cache, the first `hello` took 142.6, 151.1 and 133.9 ms; a repeat took 5.1, 3.1 and 7.5 µs. Other uncached examples took 126.6–174.5 ms. These are backend calls, not end-to-end UI latency or cold-disk results, and timings may vary on other hardware. A real-archive integration check compared titles and multiplicities for `hello`, `look up`, `AI` and `café` against the original go-stardict search; a small fixture checks an alias. These examples are not exhaustive proof of equivalence across all entries or normalization cases.
