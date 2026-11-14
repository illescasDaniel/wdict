package core

import (
	"context"
	"log"
	"strings"
)

// PronunciationAudioService resolves a play request into audio bytes or a
// failure copy (C6): recording → Piper (natural voice, optional) → IPA
// mnemonics → headword → verbatim copy.
type PronunciationAudioService struct {
	Recordings RecordingSource  // optional; nil = never fetch
	Piper      PiperVoice       // optional; nil = never synthesise with Piper
	Voices     VoiceSynthesiser // nil = never synthesise with eSpeak
}

// Play implements the fallback chain. It is stateless: a newer play wins on
// the frontend side only (S12). Expected failures come back as results with a
// verbatim copy; errors are never returned because the method has no error
// result at all.
func (s *PronunciationAudioService) Play(ctx context.Context, request PlayRequest) PlayResult {
	// 1. Recording path: fetch only when not skipped and a source is wired.
	//    The adapter validates the URL at the edge; a rejected URL behaves
	//    exactly like a failed fetch here (S10).
	if !request.SkipRecording && request.AudioURL != "" && s.Recordings != nil {
		if data, err := s.Recordings.Fetch(ctx, request.AudioURL); err == nil && len(data) > 0 {
			return PlayResult{Audio: audioDataURI("audio/mpeg", data), Source: PlayRecording}
		}
	}
	// 2. Piper path (natural voices, C5 of the piper-voices page): only when
	//    the runtime and the language's voice are installed. The adapter
	//    wraps the sanitised IPA in [[…]]; a failure falls through to the
	//    eSpeak path unchanged (F2). Piper problems are logged for
	//    diagnostics and never shown as errors. Non-phonemic notation
	//    (syllable dots, combining marks like the d͡ʒ tie, nasal tildes, the
	//    ɫ̩ syllabicity mark) is stripped first: the voices drop or mis-speak
	//    it — spoken dots read as extra "a"-like sounds.
	ipa := stripNonPhonemic(StripPronunciationDelimiters(request.Text))
	if s.Piper != nil && s.Piper.Available(request.Language) {
		if data, err := s.Piper.Speak(ctx, request.Language, ipa); err == nil && len(data) > 0 {
			return PlayResult{Audio: audioDataURI("audio/wav", data), Source: PlayPiper, Voice: s.Piper.Voice(request.Language)}
		} else {
			log.Printf("piper: falling back to eSpeak: %v", err)
		}
	}
	// 3. Voice path.
	if s.Voices == nil || !s.Voices.Available() {
		return PlayResult{Message: CopyNoSource}
	}
	voice := VoiceForAccent(request.Language, accentTitle(request))
	// Voice path (C5, revision 3): the shown IPA is spoken as eSpeak mnemonic
	// phonemes when every symbol maps (S2); otherwise — or when that
	// synthesis fails — the headword is spoken as plain text (S7). Both
	// results stay in the espeak source (C3). Non-phonemic notation is
	// stripped before conversion so tie-barred affricates (d͡ʒ), syllable
	// dots and nasalised vowels (ẽ) convert instead of falling back.
	if data, ok := s.speakMnemonics(ctx, request.Language, voice, ipa); ok {
		return PlayResult{Audio: audioDataURI("audio/wav", data), Source: PlayVoice, Voice: voice}
	}
	if data, ok := s.synthesise(ctx, voice, request.Headword); ok {
		return PlayResult{Audio: audioDataURI("audio/wav", data), Source: PlayVoice, Voice: voice}
	}
	return PlayResult{Message: CopyNoSynthesis}
}

// speakMnemonics converts the shown transcription into mnemonic phonemes and
// speaks them; it reports false when the transcription is not fully mappable
// or the synthesis fails. The input arrives delimiter-stripped and free of
// non-phonemic notation (see Play).
func (s *PronunciationAudioService) speakMnemonics(ctx context.Context, language Language, voice, ipa string) ([]byte, bool) {
	mnemonics, ok := IPAToMnemonics(language, ipa)
	if !ok {
		return nil, false
	}
	data, err := s.Voices.SpeakPhonemes(ctx, voice, mnemonics)
	if err != nil || len(data) == 0 {
		return nil, false
	}
	return data, true
}

// synthesise speaks one plain-text input (the headword fallback, F4) and
// reports whether the output is playable.
func (s *PronunciationAudioService) synthesise(ctx context.Context, voice, text string) ([]byte, bool) {
	if strings.TrimSpace(text) == "" {
		return nil, false
	}
	data, err := s.Voices.SpeakText(ctx, voice, text)
	if err != nil || len(data) == 0 {
		return nil, false
	}
	return data, true
}

// accentTitle picks the label text that names the accent: the original label
// when the display label is a flag run, otherwise the label itself.
func accentTitle(request PlayRequest) string {
	if title := strings.TrimSpace(request.LabelTitle); title != "" {
		return title
	}
	return strings.TrimSpace(request.Label)
}
