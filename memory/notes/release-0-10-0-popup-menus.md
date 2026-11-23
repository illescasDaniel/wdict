---
name: release-0-10-0-popup-menus
description: 0.10.0 published; footer language and appearance selectors are popup menus
type: decision
date: 2026-11-23
status: current
supersedes: 
superseded_by: 
source: 
---
The footer language select and appearance radio buttons became popup menus (`role=menu`, `menuitemradio`, `.menu-trigger`) to save space; done as a small tweak without SDD gates, at the user's choice. Tests pick options through `frontend/src/routes/menu-helpers.js`; browser scripts must open the popup before clicking `[data-mode]`. 0.10.0 released via the dispatched workflow and the system install updated (`wdict --version` prints 0.10.0).
**Open:** the browser check scripts were never run (no Chrome); `sdd/appearance` and `sdd/footer-language` specs still describe the old controls.
