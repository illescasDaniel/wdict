# Now
Updated: 2026-11-23

## Doing
- Nothing in progress. French support (lookup, interface, Piper and eSpeak voice, morphology and spelling alternatives) is committed. Gate passes. Last independent grade was NEEDS WORK; findings G4 to G9 were fixed afterwards and not re-graded. See `sdd/french-support.html`.

## Done recently
- Released 0.9.0 (French support, folded in the held 0.8.4 fixes); v0.9.0 tag and GitHub release with the Linux AppImage. The npm audit hold cleared (0 vulnerabilities).
- README simplified on main; detailed build, testing, voices, sources and release guidance moved to docs. Awaiting user review before any push.
- macOS support and unsigned app packaging merged to main; Linux and macOS CI passed and uploaded the Apple silicon ZIP.

## Next
- (User) Review the French support verification report; remaining gaps: real-window checks, older-binary rollback, French pronunciation quality, no re-grade after the G4-G9 fixes. French alternatives now take about 0.5-0.7 s on the real index (fold fast path); slower hardware has less margin under the 2 s deadline.
- The user has verified that the native macOS GUI and audible playback work. Cross-distribution AppImage and bundled-licence checks remain open.
