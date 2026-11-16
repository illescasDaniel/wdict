---
name: espeak-gpl-separate-process
description: eSpeak-NG may be bundled because it runs as an unmodified separate CLI process; never link libespeak-ng
type: decision
date: 2026-11-13
status: current
supersedes: 
superseded_by: 
source: 
---
Licence review (2026-11-13, web sources): eSpeak-NG is GPLv3. Calling the unmodified standalone CLI as a separate process is
what upstream calls correct, and GPL-3.0 section 5 makes the AppImage an aggregate, so the GPL does not extend to the MIT app.
Packaging conditions: ship the GPLv3 COPYING text, name espeak-ng, and give a source offer (upstream repo and pinned
version). Linking libespeak-ng would combine works and is not allowed.
