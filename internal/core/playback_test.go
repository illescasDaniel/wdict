package core

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

// Verbatim failure copies (C6). The tests assert them byte for byte.
const (
	copyNoSource    = "No recording, and eSpeak-NG is not installed. Install eSpeak-NG (or use the AppImage build) to hear this transcription."
	copyNoSynthesis = "Could not synthesise this pronunciation."
)

type fakeRecordings struct {
	data []byte
	err  error
	urls []string
}

func (f *fakeRecordings) Fetch(_ context.Context, audioURL string) ([]byte, error) {
	f.urls = append(f.urls, audioURL)
	return f.data, f.err
}

// voiceCall records one synthesis attempt: phonemes marks the mnemonic mode
// (C5, revision 3); text marks the plain-text headword fallback (F4).
type voiceCall struct {
	phonemes bool
	voice    string
	text     string
}

type fakeVoices struct {
	available bool
	outputs   map[voiceCall][]byte
	errs      map[voiceCall]error
	calls     []voiceCall
}

func (f *fakeVoices) Available() bool { return f.available }
func (f *fakeVoices) SpeakPhonemes(_ context.Context, voice, mnemonics string) ([]byte, error) {
	call := voiceCall{phonemes: true, voice: voice, text: mnemonics}
	return f.speak(call)
}
func (f *fakeVoices) SpeakText(_ context.Context, voice, text string) ([]byte, error) {
	call := voiceCall{phonemes: false, voice: voice, text: text}
	return f.speak(call)
}

func (f *fakeVoices) speak(call voiceCall) ([]byte, error) {
	f.calls = append(f.calls, call)
	if err, ok := f.errs[call]; ok {
		return nil, err
	}
	if data, ok := f.outputs[call]; ok {
		return data, nil
	}
	return nil, errNoFakeOutput{call}
}

type errNoFakeOutput struct{ call voiceCall }

func (e errNoFakeOutput) Error() string {
	return "no fake output for " + e.call.voice + "/" + e.call.text
}

func service(recordings *fakeRecordings, voices *fakeVoices) *PronunciationAudioService {
	s := &PronunciationAudioService{Voices: voices}
	if recordings != nil {
		s.Recordings = recordings
	}
	return s
}

const wavBytes = "WAVDATA"

func voiceResult(voice, text string) map[voiceCall][]byte {
	return map[voiceCall][]byte{{voice: voice, text: text}: []byte(wavBytes)}
}

func phonemeResult(voice, mnemonics string) map[voiceCall][]byte {
	return map[voiceCall][]byte{{phonemes: true, voice: voice, text: mnemonics}: []byte(wavBytes)}
}

func TestVoiceForAccentMapsAccentNamesToEspeakVoices(t *testing.T) {
	for name, want := range map[string]string{
		"General American":       "en-us",
		"US":                     "en-us",
		"Received Pronunciation": "en-gb",
		"UK":                     "en-gb",
		"Australia":              "en-au",
		"New Zealand":            "en-nz",
		"Canada":                 "en-ca",
		"Ireland":                "en-ie",
		"South Africa":           "en-za",
		"general american":       "en-us", // case-insensitive
		"CANADA":                 "en-ca",
		"UK, dialectal":          "en-gb", // qualifiers ignored
		"US, Canada, dialectal":  "en-us",
	} {
		t.Run(name, func(t *testing.T) {
			if got := VoiceForAccent(English, name); got != want {
				t.Fatalf("got %q", got)
			}
		})
	}
}

func TestVoiceForAccentMapsFlagLabelsToEspeakVoices(t *testing.T) {
	for flags, want := range map[string]string{
		"🇺🇸": "en-us",
		"🇬🇧": "en-gb",
		"🇦🇺": "en-au",
		"🇳🇿": "en-nz",
		"🇨🇦": "en-ca",
		"🇮🇪": "en-ie",
		"🇿🇦": "en-za",
	} {
		t.Run(flags, func(t *testing.T) {
			if got := VoiceForAccent(English, flags); got != want {
				t.Fatalf("got %q", got)
			}
		})
	}
	if got := VoiceForAccent(English, "🇺🇸🇬🇧"); got != "en-us" { // generic → language default
		t.Fatalf("generic got %q", got)
	}
}

func TestVoiceForAccentUsesLanguageDefaults(t *testing.T) {
	// (S3): no label and an unknown accent part give the language default.
	if got := VoiceForAccent(English, ""); got != "en-us" {
		t.Fatalf("english default got %q", got)
	}
	if got := VoiceForAccent(Spanish, ""); got != "es" {
		t.Fatalf("spanish default got %q", got)
	}
	if got := VoiceForAccent(English, "Northumbria"); got != "en-us" {
		t.Fatalf("unknown accent got %q", got)
	}
}

func TestIPAToMnemonicsCoversEveryTableEntry(t *testing.T) {
	for _, language := range []Language{English, Spanish} {
		table, ok := MnemonicTable(language)
		if !ok {
			t.Fatalf("no mnemonic table for %s", language)
		}
		for ipa, want := range table {
			t.Run(string(language)+"/"+ipa, func(t *testing.T) {
				got, ok := IPAToMnemonics(language, ipa)
				if !ok || got != want {
					t.Fatalf("got %q (%v), want %q", got, ok, want)
				}
			})
		}
	}
}

func TestIPAToMnemonicsMatchesLongestFirst(t *testing.T) {
	// (C5, revision 3): two-symbol entries win over their prefixes.
	for ipa, want := range map[string]string{
		"ɜː": "3:", // before ɜ→3
		"əʊ": "oU", // not @U
		"eɪ": "eI", // not EI
		"ɑː": "A:", // bare ɑ has no entry: the pair is required
	} {
		t.Run(ipa, func(t *testing.T) {
			got, ok := IPAToMnemonics(English, ipa)
			if !ok || got != want {
				t.Fatalf("got %q (%v), want %q", got, ok, want)
			}
		})
	}
}

// engineCanonicalWords pairs Wiktionary-style IPA with the engine's own
// canonical mnemonic stream (`espeak-ng -v en-gb -q -x <word>`, espeak-ng
// 1.52, this machine, 2026-11-13). They encode the vowels the engine really
// uses — the per-entry echo oracle cannot catch a wrong-but-parseable
// phoneme, which is exactly how ɒ→Q shipped sounding wrong. The comparison
// ignores stress markers: the engine inserts its default stress even when
// the IPA has none (stress from IPA marks is covered by the dedicated
// stress tests), and words whose canonical diverges from the shown IPA
// through engine allophony (e.g. cinco's ŋ canonicalised to n) are excluded.
func TestIPAToMnemonicsMatchesTheEngineCanonical(t *testing.T) {
	normalize := func(s string) string { return strings.ReplaceAll(s, "'", "") }
	for _, tc := range []struct {
		language Language
		ipa      string
		want     string
	}{
		{English, "/ˈnɒl.ɪdʒ/", "n'0lIdZ"}, // knowledge
		{English, "/kæt/", "k'at"},         // cat
		{English, "/ˈθɔːt/", "T'O:t"},      // thought
		{English, "/strʌt/", "str'Vt"},     // strut
		{English, "/nɜːs/", "n'3:s"},       // nurse
		{English, "/kɪt/", "k'It"},         // kit
		{English, "/fliːs/", "fl'i:s"},     // fleece
		{English, "/drɛs/", "dr'Es"},       // dress
		{English, "/fʊt/", "f'Ut"},         // foot
		{English, "/ɡuːs/", "g'u:s"},       // goose
		{English, "/tʃiːz/", "tS'i:z"},     // cheese
		{English, "/dʒʌdʒ/", "dZ'VdZ"},     // judge
		{English, "/ˈθɪn/", "T'In"},        // thin
		{English, "/ðɪs/", "D'Is"},         // this
		{English, "/ˈsɪŋ/", "s'IN"},        // sing
		{English, "/ˈʃɪp/", "S'Ip"},        // ship
		{English, "/ˈvɪʒən/", "v'IZ@n"},    // vision
		{English, "/ˈkwaɪ.ə/", "kw'aI@"},   // choir
		{English, "/həˈləʊ/", "h@l'oU"},    // hello
		{English, "/deɪ/", "d'eI"},         // day
		{English, "/kɔɪn/", "k'OIn"},       // coin
		{English, "/haʊs/", "h'aUs"},       // house
		{English, "/skwɛə/", "skw'e@"},     // square
		{English, "/kjʊə/", "kj'U@"},       // cure
		{Spanish, "[ˈola]", "'ola"},        // hola
	} {
		t.Run(string(tc.language)+"/"+tc.ipa, func(t *testing.T) {
			got, ok := IPAToMnemonics(tc.language, tc.ipa)
			if !ok || normalize(got) != normalize(tc.want) {
				t.Fatalf("got %q (%v), want %q", got, ok, tc.want)
			}
		})
	}
}

func TestVowelMnemonicsStayInsideTheTables(t *testing.T) {
	// the stress anchoring treats these values as vowels; every one of them
	// must be a table value, or stress would anchor to nothing
	for _, language := range []Language{English, Spanish} {
		table, ok := MnemonicTable(language)
		if !ok {
			t.Fatalf("no mnemonic table for %s", language)
		}
		values := make(map[string]bool, len(table))
		for _, mnemonic := range table {
			values[mnemonic] = true
		}
		for mnemonic := range ipaVowelMnemonics[language] {
			if !values[mnemonic] {
				t.Fatalf("%s: vowel mnemonic %q is not in the table", language, mnemonic)
			}
		}
	}
}

func TestIPAToMnemonicsPlacesStressBeforeTheStressedVowel(t *testing.T) {
	// (C5, revision 3, engine-canonical style): the marker anchors to the
	// next vowel, never to an onset consonant (word → w'3:d, hola → 'ola)
	for ipa, want := range map[string]string{
		"ˈwɝːd": "w'3:d",
		"ˌwɝːd": "w'3:d", // secondary stress uses the same marker
		"ˈola":  "'ola",
	} {
		t.Run(ipa, func(t *testing.T) {
			language := English
			if want == "'ola" {
				language = Spanish
			}
			got, ok := IPAToMnemonics(language, ipa)
			if !ok || got != want {
				t.Fatalf("got %q (%v), want %q", got, ok, want)
			}
		})
	}
}

func TestIPAToMnemonicsDropsSyllableDots(t *testing.T) {
	if got, ok := IPAToMnemonics(English, "bʌt.ə"); !ok || got != "bVt@" {
		t.Fatalf("got %q (%v)", got, ok)
	}
	if got, ok := IPAToMnemonics(Spanish, "ˈpla.ʝa"); !ok || got != "pl'aJa" {
		t.Fatalf("got %q (%v)", got, ok)
	}
}

func TestIPAToMnemonicsCollapsesWhitespace(t *testing.T) {
	// multi-word IPA keeps one space; runs collapse and edges trim.
	if got, ok := IPAToMnemonics(English, "ˈhɛloʊ  ˈwɜːld"); !ok || got != "h'EloU w'3:ld" {
		t.Fatalf("got %q (%v)", got, ok)
	}
	if got, ok := IPAToMnemonics(English, "\t a  b \n"); !ok || got != "a b" {
		t.Fatalf("got %q (%v)", got, ok)
	}
}

func TestIPAToMnemonicsMergesTheLengthMark(t *testing.T) {
	// ː merges with the preceding vowel mnemonic (C5, revision 3).
	for ipa, want := range map[string]string{
		"iː": "i:", "uː": "u:", "ɜː": "3:", "ɑː": "A:", "ɔː": "O:", // table pairs
		"ɝː":   "3:",   // already long: the mark is consumed
		"ɛː":   "E:",   // generic merge on a table vowel
		"wɝ:d": "w3:d", // ASCII colon behaves like ː
	} {
		t.Run(ipa, func(t *testing.T) {
			got, ok := IPAToMnemonics(English, ipa)
			if !ok || got != want {
				t.Fatalf("got %q (%v), want %q", got, ok, want)
			}
		})
	}
	if _, ok := IPAToMnemonics(English, "tː"); ok {
		t.Fatal("ː merged with a consonant")
	}
	if _, ok := IPAToMnemonics(English, "ːd"); ok {
		t.Fatal("bare ː converted")
	}
}

func TestIPAToMnemonicsRejectsUnmappableSymbols(t *testing.T) {
	// (S7): any symbol outside the language's table fails the conversion, so
	// the caller falls back to the headword.
	for _, tc := range []struct {
		language Language
		ipa      string
	}{
		{English, "ø"}, // not in the en table (combining marks are stripped before conversion)
		{English, "β"}, // Spanish-only symbol
		{Spanish, "ɑː"},
		{Spanish, "ø"},
		{English, ""},
		{English, "   "},
		{English, "..."}, // everything dropped → nothing to speak
		{English, "/"},
	} {
		t.Run(string(tc.language)+"/"+tc.ipa, func(t *testing.T) {
			if got, ok := IPAToMnemonics(tc.language, tc.ipa); ok {
				t.Fatalf("converted %q to %q", tc.ipa, got)
			}
		})
	}
	if _, ok := IPAToMnemonics(Language("de"), "aɪ"); ok {
		t.Fatal("language without a table converted")
	}
}

func TestIPAToMnemonicsStripsTheDelimiters(t *testing.T) {
	if got, ok := IPAToMnemonics(English, "/wɝːd/"); !ok || got != "w3:d" {
		t.Fatalf("got %q (%v)", got, ok)
	}
	if got, ok := IPAToMnemonics(Spanish, "[ˈola]"); !ok || got != "'ola" {
		t.Fatalf("got %q (%v)", got, ok)
	}
}

func TestStripNonPhonemic(t *testing.T) {
	// notation, not sound: syllable dots, combining marks (d͡ʒ tie, nasal
	// tildes, syllabicity); stress and length stay
	for in, want := range map[string]string{
		"ˈvɛd͡ʒ.tə.bəl": "ˈvɛdʒtəbəl", // vegetable: tie + dot stripped
		"[ẽnˈrike]":     "[enˈrike]",  // Enrique: nasal tilde stripped
		"bʌtn̩":          "bʌtn",       // syllabicity mark stripped
		"ˈbʌt.ən":       "ˈbʌtən",     // dots stripped
		"wɜːd":          "wɜːd",       // stress/length and base symbols stay
		"":              "",
	} {
		if got := stripNonPhonemic(in); got != want {
			t.Fatalf("%q: got %q", in, got)
		}
	}
}

func TestGivenTieBarredTranscriptionWhenPlayedThenTheConversionSucceeds(t *testing.T) {
	// (post-handoff fixes, sdd/audio-pronunciation.html): the real offline
	// "vegetable" items carry the d͡ʒ tie and syllable dots — they must be
	// spoken via the mnemonic conversion, not the headword fallback
	voices := &fakeVoices{available: true, outputs: phonemeResult("en-gb", "v'EdZt@b@l")}
	s := service(nil, voices)
	result := s.Play(context.Background(), PlayRequest{
		Text: "/ˈvɛd͡ʒ.tə.bəl/", Label: "🇬🇧", LabelTitle: "Received Pronunciation", Language: English, Headword: "vegetable",
	})
	want := "data:audio/wav;base64," + base64.StdEncoding.EncodeToString([]byte(wavBytes))
	if result.Audio != want || result.Source != PlayVoice {
		t.Fatalf("got %+v", result)
	}
	if len(voices.calls) != 1 || voices.calls[0] != (voiceCall{phonemes: true, voice: "en-gb", text: "v'EdZt@b@l"}) {
		t.Fatalf("voice calls %+v", voices.calls)
	}
}

func TestGivenNasalisedTranscriptionWhenPlayedThenTheConversionSucceeds(t *testing.T) {
	// the real offline "Enrique" item is [ẽnˈrike] — the nasal tilde is
	// notation; the conversion now speaks it instead of falling back
	voices := &fakeVoices{available: true, outputs: phonemeResult("es", "enr'ike")}
	s := service(nil, voices)
	result := s.Play(context.Background(), PlayRequest{
		Text: "[ẽnˈrike]", Language: Spanish, Headword: "Enrique",
	})
	if result.Audio == "" || result.Source != PlayVoice {
		t.Fatalf("got %+v", result)
	}
	if len(voices.calls) != 1 || voices.calls[0] != (voiceCall{phonemes: true, voice: "es", text: "enr'ike"}) {
		t.Fatalf("voice calls %+v", voices.calls)
	}
}

func TestStripPronunciationDelimiters(t *testing.T) {
	for in, want := range map[string]string{
		"/wɝːd/":     "wɝːd",
		"[ˈola]":     "ˈola",
		"wɝːd":       "wɝːd",
		"/wɜːd]":     "/wɜːd]", // unmatched delimiters are kept
		"  /ˈola/  ": "ˈola",
	} {
		if got := StripPronunciationDelimiters(in); got != want {
			t.Fatalf("%q: got %q", in, got)
		}
	}
}

func TestGivenRecordingURLWhenPlayedThenRecordingBytesComeBackAsDataURI(t *testing.T) {
	// given (S1)
	rec := &fakeRecordings{data: []byte("MP3DATA")}
	voices := &fakeVoices{available: true}
	s := service(rec, voices)
	// when
	result := s.Play(context.Background(), PlayRequest{
		Text: "/wɝːd/", Label: "🇺🇸", LabelTitle: "General American", Language: English,
		AudioURL: "https://example.test/a.mp3", Headword: "word",
	})
	// then
	want := "data:audio/mpeg;base64," + base64.StdEncoding.EncodeToString([]byte("MP3DATA"))
	if result.Audio != want || result.Source != PlayRecording || result.Message != "" {
		t.Fatalf("got %+v", result)
	}
	if len(voices.calls) != 0 {
		t.Fatalf("voice path used: %+v", voices.calls)
	}
}

func TestGivenNoURLAndNoSynthesiserWhenPlayedThenTheF1CopyIsShown(t *testing.T) {
	// given (S4, F1)
	voices := &fakeVoices{available: false}
	s := service(nil, voices)
	// when
	result := s.Play(context.Background(), PlayRequest{Text: "/wɝːd/", Language: English, Headword: "word"})
	// then
	if result.Audio != "" || result.Message != copyNoSource {
		t.Fatalf("got %+v", result)
	}
	if len(voices.calls) != 0 {
		t.Fatalf("synthesis attempted without a synthesiser: %+v", voices.calls)
	}
}

func TestGivenMappableTranscriptionWhenPlayedThenTheMnemonicsAreSpoken(t *testing.T) {
	// given (S2, revision 3): the shown IPA is fully covered by the table
	voices := &fakeVoices{available: true, outputs: phonemeResult("en-us", "w3:d")}
	s := service(nil, voices)
	// when
	result := s.Play(context.Background(), PlayRequest{
		Text: "/wɝːd/", Label: "🇺🇸", LabelTitle: "General American", Language: English, Headword: "word",
	})
	// then: eSpeak speaks the converted stream with the accent voice
	want := "data:audio/wav;base64," + base64.StdEncoding.EncodeToString([]byte(wavBytes))
	if result.Audio != want || result.Source != PlayVoice || result.Message != "" {
		t.Fatalf("got %+v", result)
	}
	if result.Voice != "en-us" {
		t.Fatalf("voice %q", result.Voice)
	}
	if len(voices.calls) != 1 || voices.calls[0] != (voiceCall{phonemes: true, voice: "en-us", text: "w3:d"}) {
		t.Fatalf("voice calls %+v", voices.calls)
	}
}

func TestGivenFailedRecordingWhenPlayedThenTheAccentVoicePlaysInstead(t *testing.T) {
	// given (S5, F2, revision 3): the mappable transcription is spoken as
	// mnemonic phonemes
	rec := &fakeRecordings{err: context.DeadlineExceeded}
	voices := &fakeVoices{available: true, outputs: phonemeResult("en-us", "w3:d")}
	s := service(rec, voices)
	// when
	result := s.Play(context.Background(), PlayRequest{
		Text: "/wɝːd/", Label: "🇺🇸", LabelTitle: "General American", Language: English,
		AudioURL: "https://example.test/a.mp3", Headword: "word",
	})
	// then
	want := "data:audio/wav;base64," + base64.StdEncoding.EncodeToString([]byte(wavBytes))
	if result.Audio != want || result.Source != PlayVoice || result.Message != "" {
		t.Fatalf("got %+v", result)
	}
	if result.Voice != "en-us" {
		t.Fatalf("voice %q", result.Voice)
	}
	if len(voices.calls) != 1 || voices.calls[0] != (voiceCall{phonemes: true, voice: "en-us", text: "w3:d"}) {
		t.Fatalf("voice calls %+v", voices.calls)
	}
}

func TestGivenUnmappableTranscriptionWhenPlayedThenTheHeadwordIsSpoken(t *testing.T) {
	// given (S7, revision 3): ø is outside the table even after the
	// non-phonemic strip (the /ˈbʌtn̩/ syllabic case now CONVERTS: the mark
	// is notation, not sound)
	voices := &fakeVoices{available: true, outputs: voiceResult("en-us", "button")}
	s := service(nil, voices)
	// when
	result := s.Play(context.Background(), PlayRequest{
		Text: "/bʌtø/", Label: "🇺🇸", LabelTitle: "General American", Language: English, Headword: "button",
	})
	// then: the headword is spoken as plain text, still as the espeak source
	want := "data:audio/wav;base64," + base64.StdEncoding.EncodeToString([]byte(wavBytes))
	if result.Audio != want || result.Source != PlayVoice || result.Message != "" {
		t.Fatalf("got %+v", result)
	}
	if result.Voice != "en-us" {
		t.Fatalf("voice %q", result.Voice)
	}
	if len(voices.calls) != 1 || voices.calls[0] != (voiceCall{voice: "en-us", text: "button"}) {
		t.Fatalf("voice calls %+v", voices.calls)
	}
}

func TestGivenMnemonicSynthesisFailureWhenPlayedThenTheHeadwordIsSpoken(t *testing.T) {
	// given (S7, revision 3): the phoneme synthesis fails, the headword retry
	// succeeds in the same accent voice
	voices := &fakeVoices{
		available: true,
		errs:      map[voiceCall]error{{phonemes: true, voice: "en-us", text: "w3:d"}: context.DeadlineExceeded},
		outputs:   voiceResult("en-us", "word"),
	}
	s := service(nil, voices)
	// when
	result := s.Play(context.Background(), PlayRequest{
		Text: "/wɝːd/", Label: "🇺🇸", LabelTitle: "General American", Language: English, Headword: "word",
	})
	// then
	want := "data:audio/wav;base64," + base64.StdEncoding.EncodeToString([]byte(wavBytes))
	if result.Audio != want || result.Source != PlayVoice || result.Voice != "en-us" {
		t.Fatalf("got %+v", result)
	}
	if len(voices.calls) != 2 ||
		voices.calls[0] != (voiceCall{phonemes: true, voice: "en-us", text: "w3:d"}) ||
		voices.calls[1] != (voiceCall{voice: "en-us", text: "word"}) {
		t.Fatalf("voice calls %+v", voices.calls)
	}
}

func TestGivenSynthesisFailureWhenPlayedThenTheF4CopyIsShown(t *testing.T) {
	// given (S7, F4, revision 3): both voice-path attempts fail
	rec := &fakeRecordings{err: context.DeadlineExceeded}
	voices := &fakeVoices{available: true, errs: map[voiceCall]error{
		{phonemes: true, voice: "en-us", text: "w3:d"}: context.DeadlineExceeded,
		{voice: "en-us", text: "word"}:                 context.DeadlineExceeded,
	}}
	s := service(rec, voices)
	// when
	result := s.Play(context.Background(), PlayRequest{
		Text: "/wɝːd/", Language: English, AudioURL: "https://example.test/a.mp3", Headword: "word",
	})
	// then
	if result.Audio != "" || result.Message != copyNoSynthesis {
		t.Fatalf("got %+v", result)
	}
	if len(voices.calls) != 2 {
		t.Fatalf("voice calls %+v", voices.calls)
	}
}

func TestGivenEmptyVoiceOutputWhenPlayedThenTheF4CopyIsShown(t *testing.T) {
	// given (S7): empty output counts as a failure on both attempts
	rec := &fakeRecordings{err: context.DeadlineExceeded}
	voices := &fakeVoices{
		available: true,
		outputs: map[voiceCall][]byte{
			{phonemes: true, voice: "en-us", text: "w3:d"}: nil,
		},
	}
	s := service(rec, voices)
	// when
	result := s.Play(context.Background(), PlayRequest{
		Text: "/wɝːd/", Language: English, AudioURL: "https://example.test/a.mp3", Headword: "word",
	})
	// then
	if result.Audio != "" || result.Message != copyNoSynthesis {
		t.Fatalf("got %+v", result)
	}
	if len(voices.calls) != 2 {
		t.Fatalf("voice calls %+v", voices.calls)
	}
}

func TestGivenSkipRecordingWhenPlayedThenTheRecordingIsNeverFetched(t *testing.T) {
	// given (S8, F5)
	rec := &fakeRecordings{data: []byte("MP3DATA")} // would succeed
	voices := &fakeVoices{available: true, outputs: phonemeResult("es", "'ola")}
	s := service(rec, voices)
	// when
	result := s.Play(context.Background(), PlayRequest{
		Text: "[ˈola]", Language: Spanish, AudioURL: "https://example.test/a.mp3",
		SkipRecording: true, Headword: "hola",
	})
	// then
	if result.Audio == "" || result.Source != PlayVoice {
		t.Fatalf("got %+v", result)
	}
	if len(rec.urls) != 0 {
		t.Fatalf("recording fetched despite SkipRecording: %v", rec.urls)
	}
	if len(voices.calls) != 1 || voices.calls[0] != (voiceCall{phonemes: true, voice: "es", text: "'ola"}) {
		t.Fatalf("voice calls %+v", voices.calls)
	}
}

func TestGivenNonHTTPSRecordingURLWhenPlayedThenTheURLIsTreatedAsAbsent(t *testing.T) {
	// given (S10): the fake RecordingSource emulates the adapter edge and
	// rejects anything that is not https with a host
	rec := &fakeRecordings{err: errInvalidURL{}}
	voices := &fakeVoices{available: true, outputs: phonemeResult("en-us", "w3:d")}
	s := service(rec, voices)
	// when
	result := s.Play(context.Background(), PlayRequest{
		Text: "/wɝːd/", Language: English, AudioURL: "http://example.test/a.mp3", Headword: "word",
	})
	// then
	if result.Audio == "" || result.Source != PlayVoice {
		t.Fatalf("got %+v", result)
	}
}

type errInvalidURL struct{}

func (errInvalidURL) Error() string { return "invalid audio URL" }

func TestGivenOvercapRecordingWhenPlayedThenTheVoiceFallbackPlays(t *testing.T) {
	// given (S11): the adapter aborts an oversized download, which the service
	// observes as a failed fetch
	rec := &fakeRecordings{err: errTooLarge{}}
	voices := &fakeVoices{available: true, outputs: phonemeResult("en-us", "w3:d")}
	s := service(rec, voices)
	// when
	result := s.Play(context.Background(), PlayRequest{
		Text: "/wɝːd/", Language: English, AudioURL: "https://example.test/big.mp3", Headword: "word",
	})
	// then
	if result.Audio == "" || result.Source != PlayVoice {
		t.Fatalf("got %+v", result)
	}
}

type errTooLarge struct{}

func (errTooLarge) Error() string { return "audio over the size cap" }

func TestGivenRecordingUnavailableWhenPlayedThenResultIsNeverAnError(t *testing.T) {
	// The chain always terminates in audio or copy; Play returns no error by signature,
	// so assert the result contract shape for the recorded path (C3).
	rec := &fakeRecordings{data: []byte("MP3DATA")}
	voices := &fakeVoices{available: true}
	s := service(rec, voices)
	result := s.Play(context.Background(), PlayRequest{
		Text: "/wɝːd/", Language: English, AudioURL: "https://example.test/a.mp3", Headword: "word",
	})
	if (result.Audio != "" && result.Message != "") || (result.Audio == "" && result.Message == "") {
		t.Fatalf("result must be audio or copy: %+v", result)
	}
	if !strings.HasPrefix(result.Audio, "data:audio/") && result.Audio != "" {
		t.Fatalf("audio is not a data URI: %q", result.Audio)
	}
}

// piperCall records one Piper synthesis attempt (C5, piper-voices page).
type piperCall struct {
	language Language
	ipa      string
}

type fakePiper struct {
	available func(Language) bool
	name      func(Language) string
	output    []byte
	err       error
	calls     []piperCall
}

func (f *fakePiper) Available(language Language) bool {
	if f.available != nil {
		return f.available(language)
	}
	return true
}
func (f *fakePiper) Voice(language Language) string {
	if f.name != nil {
		return f.name(language)
	}
	return "alba"
}
func (f *fakePiper) Speak(_ context.Context, language Language, ipa string) ([]byte, error) {
	f.calls = append(f.calls, piperCall{language: language, ipa: ipa})
	return f.output, f.err
}

func piperService(rec *fakeRecordings, piper *fakePiper, voices *fakeVoices) *PronunciationAudioService {
	s := &PronunciationAudioService{Voices: voices, Piper: piper}
	if rec != nil {
		s.Recordings = rec
	}
	return s
}

func TestGivenVoicePackInstalledWhenPlayedThenPiperSpeaksTheStrippedIPA(t *testing.T) {
	// given (S1): runtime + language voice installed, no recording
	piper := &fakePiper{output: []byte(wavBytes)}
	voices := &fakeVoices{available: true, outputs: phonemeResult("en-us", "w3:d")}
	s := piperService(nil, piper, voices)
	// when
	result := s.Play(context.Background(), PlayRequest{
		Text: "/ˈnɒl.ɪdʒ/", Language: English, Headword: "knowledge",
	})
	// then: Piper is called with the shown transcription, delimiters stripped
	if len(piper.calls) != 1 || piper.calls[0].ipa != "ˈnɒlɪdʒ" || piper.calls[0].language != English {
		t.Fatalf("piper calls %+v", piper.calls)
	}
	if len(voices.calls) != 0 {
		t.Fatalf("eSpeak called despite a Piper result: %+v", voices.calls)
	}
	if result.Source != "piper" || result.Audio == "" || result.Voice != "alba" {
		t.Fatalf("result %+v", result)
	}
	if !strings.HasPrefix(result.Audio, "data:audio/wav;base64,") {
		t.Fatalf("audio %q", result.Audio)
	}
}

func TestGivenPackMissingWhenPlayedThenTheEspeakPathRunsUnchanged(t *testing.T) {
	// given (S2/F1): no pack for the language — Piper is never started
	piper := &fakePiper{available: func(Language) bool { return false }, output: []byte(wavBytes)}
	voices := &fakeVoices{available: true, outputs: phonemeResult("en-us", "w3:d")}
	s := piperService(nil, piper, voices)
	// when
	result := s.Play(context.Background(), PlayRequest{
		Text: "/wɝːd/", Language: English, Headword: "word",
	})
	// then: today's eSpeak behaviour, no Piper call, no message
	if len(piper.calls) != 0 {
		t.Fatalf("piper called without a pack: %+v", piper.calls)
	}
	if result.Source != PlayVoice || result.Audio == "" {
		t.Fatalf("result %+v", result)
	}
}

func TestGivenPiperFailsWhenPlayedThenTheEspeakPathProducesAudio(t *testing.T) {
	// given (S4/F2): the subprocess fails; the chain falls back
	piper := &fakePiper{err: errors.New("piper exited 1")}
	voices := &fakeVoices{available: true, outputs: phonemeResult("en-us", "w3:d")}
	s := piperService(nil, piper, voices)
	// when
	result := s.Play(context.Background(), PlayRequest{
		Text: "/wɝːd/", Language: English, Headword: "word",
	})
	// then
	if len(piper.calls) != 1 {
		t.Fatalf("piper calls %+v", piper.calls)
	}
	if result.Source != PlayVoice || result.Audio == "" || result.Message != "" {
		t.Fatalf("result %+v", result)
	}
}

func TestGivenRecordingAvailableWhenPlayedThenPiperIsNotStarted(t *testing.T) {
	// given (S3): a recording still wins over any installed voice
	rec := &fakeRecordings{data: []byte("MP3DATA")}
	piper := &fakePiper{output: []byte(wavBytes)}
	voices := &fakeVoices{available: true}
	s := piperService(rec, piper, voices)
	// when
	result := s.Play(context.Background(), PlayRequest{
		Text: "/wɝːd/", Language: English, AudioURL: "https://example.test/a.mp3", Headword: "word",
	})
	// then
	if result.Source != PlayRecording || result.Audio == "" {
		t.Fatalf("result %+v", result)
	}
	if len(piper.calls) != 0 || len(voices.calls) != 0 {
		t.Fatalf("piper=%+v voices=%+v", piper.calls, voices.calls)
	}
}

func TestGivenSkipRecordingWhenPlayedThenPiperIsStillTriedBeforeEspeak(t *testing.T) {
	// given (S11): recordings are skipped, Piper (installed) is not
	piper := &fakePiper{output: []byte(wavBytes)}
	voices := &fakeVoices{available: true, outputs: phonemeResult("en-us", "w3:d")}
	s := piperService(&fakeRecordings{data: []byte("MP3DATA")}, piper, voices)
	// when
	result := s.Play(context.Background(), PlayRequest{
		Text: "/wɝːd/", Language: English, AudioURL: "https://example.test/a.mp3",
		SkipRecording: true, Headword: "word",
	})
	// then
	if len(piper.calls) != 1 || result.Source != "piper" {
		t.Fatalf("calls %+v result %+v", piper.calls, result)
	}
}

func TestGivenNoPiperFieldWhenPlayedThenBehaviourIsExactlyTodayS(t *testing.T) {
	// (S2 regression guard): Piper nil = the feature is absent entirely.
	voices := &fakeVoices{available: true, outputs: phonemeResult("en-us", "w3:d")}
	s := service(nil, voices)
	result := s.Play(context.Background(), PlayRequest{Text: "/wɝːd/", Language: English, Headword: "word"})
	if result.Source != PlayVoice || result.Audio == "" {
		t.Fatalf("result %+v", result)
	}
}
