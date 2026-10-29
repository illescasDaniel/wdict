# Active context

## Current focus

v0.3.0 is published with a Linux AppImage and the local Go-installed executable is updated and verified (`src:7b175f2d2ec3b891cb96`).

Launch experiments were removed at the user's request; the dictionary metadata error-and-retry fix remains (`src:951f2ed31a886ccf2478`). The local `./wdict` is rebuilt with the fix; the installed executable remains unchanged.

## Blockers

Native appearance rendering remains to be verified by the user. The low-severity npm advisory, cross-distribution testing and bundled license review remain open.

## Next steps

Run the installed wdict and verify appearance controls, persistence and system changes in the native window.
Review the dictionary metadata error fix in the next release; no new version is being published now.
