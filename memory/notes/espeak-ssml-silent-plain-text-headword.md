---
name: espeak-ssml-silent-plain-text-headword
description: SSML <phoneme> rendered digitally silent WAVs; speech was changed to the plain headword with an audibility check
type: incident
date: 2026-11-13
status: current
supersedes: 
superseded_by: 
source: 
---
After the user reported "i hear no sound", ffmpeg volumedetect showed the approved SSML mechanism produced max -91 dB
against -5.2 dB for plain text. With chat approval ("Speak the headword") criterion C5 was amended to plain-text headword
speech, and the real-binary integration test gained a peak >= 1000 audibility check that would have caught it.

**How to apply:** do not trust a returned WAV; assert audibility. Later revision 3 speaks the shown IPA instead, see
[ipa-to-espeak-mnemonic-converter](ipa-to-espeak-mnemonic-converter.md).
