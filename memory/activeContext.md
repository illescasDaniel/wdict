# Active context

## Current focus

Spanish online and offline support is implemented and local `./wdict` rebuilt. To address test-quality review, the cache restart test now recreates services and memory caches against shared disk without subprocesses; real-data checks are integration-tagged and documented. Default race tests and both local real-data tests passed; independent test-quality regrade passed, noting the lack of an actual process restart test. Legacy attribution amendment was approved (`src:517058a14965c942a353`); HTTP 502 handling follows `src:fbb1669e08460c8e2421`.

Preparing v0.4.0 at the user's request: commit and push, verify CI, publish with the release workflow, and update the local Go-installed executable. v0.3.0 is the latest confirmed published release (`src:7b175f2d2ec3b891cb96`).

Launch experiments were removed at the user's request; the dictionary metadata error-and-retry fix remains (`src:951f2ed31a886ccf2478`).

## Blockers

Native appearance rendering remains to be verified by the user. The low-severity npm advisory, cross-distribution testing and bundled license review remain open.

## Next steps

Commit and push v0.4.0, verify CI for that commit, run the release workflow, and update the user's installed executable using go install. The user reports the silent 502 behavior works.

Run the installed wdict and verify appearance controls, persistence and system changes in the native window.
Review the dictionary metadata error fix in the release.
