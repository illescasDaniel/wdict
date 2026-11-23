package core

import (
	"context"
	"log"
	"strings"
	"sync"
)

// PronunciationAudioService resolves a play request into audio bytes or a
// failure copy (C6): Piper (natural voice, optional) → IPA mnemonics →
// headword → verbatim copy. Recordings are never downloaded: a request's
// AudioURL is ignored.
type PronunciationAudioService struct {
	Piper  PiperVoice       // optional; nil = never synthesise with Piper
	Voices VoiceSynthesiser // nil = never synthesise with eSpeak
	Player AudioPlayer      // optional; nil or unavailable = hand the audio to the webview

	mu         sync.Mutex
	stopActive context.CancelFunc // cancels the playback of the latest Play
}

// Stop cancels the playback in progress, if any.
func (s *PronunciationAudioService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopActive != nil {
		s.stopActive()
		s.stopActive = nil
	}
}

// begin cancels the previous playback and returns the context of this one.
// It is independent of the caller's synthesis deadline: audio may outlast it.
func (s *PronunciationAudioService) begin() context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopActive != nil {
		s.stopActive()
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.stopActive = cancel
	return ctx
}

// Play implements the fallback chain. A newer play (or Stop) cancels the native
// playback of the previous one; Played results mean the audio already ran. Expected failures come back as results with a
// verbatim copy; errors are never returned because the method has no error
// result at all.
func (s *PronunciationAudioService) Play(ctx context.Context, request PlayRequest) PlayResult {
	playCtx := s.begin()
	result, wav := s.resolve(ctx, request)
	if wav == nil {
		return result
	}
	if s.Player != nil && s.Player.Available() {
		err := s.Player.Play(playCtx, wav)
		if err == nil || playCtx.Err() != nil {
			// finished, or stopped by a newer play / Stop: nothing left to do
			result.Played = true
			return result
		}
		log.Printf("native audio: falling back to the webview: %v", err)
	}
	result.Audio = audioDataURI("audio/wav", wav)
	return result
}

// resolve synthesises the request: the result carries the source and voice (or
// the failure copy) and wav holds the audio, nil on failure.
func (s *PronunciationAudioService) resolve(ctx context.Context, request PlayRequest) (result PlayResult, wav []byte) {
	if request.Mode == PlaySpelling {
		if s.Voices == nil || !s.Voices.Available() {
			return PlayResult{Message: CopyNoSource}, nil
		}
		voice := VoiceForAccent(request.Language, "")
		if data, ok := s.synthesise(ctx, voice, request.Headword); ok {
			return PlayResult{Source: PlayVoice, Voice: voice}, data
		}
		return PlayResult{Message: CopyNoSynthesis}, nil
	}
	// 1. Piper path (natural voices, C5 of the piper-voices page): only when
	//    the runtime and the language's voice are installed. Piper consumes
	//    raw phones, so normalize syllable-initial stress and notation while
	//    preserving phonemic diacritics. The adapter validates the phones
	//    against the installed model and clusters its diphthongs. Any failure
	//    falls through to the eSpeak path (F2).
	ipa := StripPronunciationDelimiters(request.Text)
	if s.Piper != nil && s.Piper.Available(request.Language) {
		if piperIPA, ok := NormalizePiperIPA(request.Language, ipa); ok {
			if data, err := s.Piper.Speak(ctx, request.Language, piperIPA); err == nil && len(data) > 0 {
				return PlayResult{Source: PlayPiper, Voice: s.Piper.Voice(request.Language)}, data
			} else {
				log.Printf("piper: falling back to eSpeak: %v", err)
			}
		}
	}
	// eSpeak's mnemonic syntax cannot represent all IPA diacritics, so only
	// this fallback strips marks before converting the transcription.
	if request.Language != French {
		ipa = stripNonPhonemic(ipa)
	}
	// 2. Voice path.
	if s.Voices == nil || !s.Voices.Available() {
		return PlayResult{Message: CopyNoSource}, nil
	}
	voice := VoiceForAccent(request.Language, accentTitle(request))
	// Voice path (C5, revision 3): the shown IPA is spoken as eSpeak mnemonic
	// phonemes when every symbol maps (S2); otherwise — or when that
	// synthesis fails — the headword is spoken as plain text (S7). Both
	// results stay in the espeak source (C3). Non-phonemic notation is
	// stripped before conversion so tie-barred affricates (d͡ʒ), syllable
	// dots and nasalised vowels (ẽ) convert instead of falling back.
	if data, ok := s.speakMnemonics(ctx, request.Language, voice, ipa); ok {
		return PlayResult{Source: PlayVoice, Voice: voice}, data
	}
	if data, ok := s.synthesise(ctx, voice, request.Headword); ok {
		return PlayResult{Source: PlayVoice, Voice: voice}, data
	}
	return PlayResult{Message: CopyNoSynthesis}, nil
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
