package core

import (
	"encoding/base64"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// PlayRequest carries one transcription item from the pronunciation line.
type PlayRequest struct {
	Text          string // transcription as shown, e.g. "/wɝːd/" (display/keying only)
	Label         string // flags or source label as shown
	LabelTitle    string // original accent label, e.g. "General American"
	Language      Language
	AudioURL      string
	SkipRecording bool   // set by the frontend's audio-error fallback (F5)
	Headword      string // plain headword spoken by the voice path (C5)
}

// PlaySource names which path produced the audio.
type PlaySource string

const (
	PlayRecording PlaySource = "recording"
	PlayVoice     PlaySource = "espeak"
	PlayPiper     PlaySource = "piper"
)

// PlayResult is audio to play or the verbatim failure copy to show. Expected
// failures are results, never errors.
type PlayResult struct {
	Audio   string // data:audio/wav or data:audio/mpeg URI; empty on failure
	Source  PlaySource
	Voice   string // eSpeak voice id used, for voice results
	Message string // verbatim failure copy when Audio is empty
}

// Verbatim failure copies (C6). Shown inline near the item; the button stays.
const (
	// CopyNoSource is the F1/F3 copy: nothing can play.
	CopyNoSource = "No recording, and eSpeak-NG is not installed. Install eSpeak-NG (or use the AppImage build) to hear this transcription."
	// CopyNoSynthesis is the F4 final copy: the voice path failed entirely.
	CopyNoSynthesis = "Could not synthesise this pronunciation."
)

// accentVoices maps the accent names NormalizePronunciations maps to flags
// (C4) to eSpeak voice ids (C5). The ids are verified against the real
// espeak-ng binary during the build; a missing per-accent voice falls back to
// the language default in the synthesiser.
var accentVoices = map[string]string{
	"general american":       "en-us",
	"us":                     "en-us",
	"received pronunciation": "en-gb",
	"uk":                     "en-gb",
	"australia":              "en-au",
	"new zealand":            "en-nz",
	"canada":                 "en-ca",
	"ireland":                "en-ie",
	"south africa":           "en-za",
}

// voiceForFlag reverses the accentFlags mapping for labels shown as flags.
var voiceForFlag = func() map[string]string {
	reverse := make(map[string]string, len(accentFlags))
	for name, flag := range accentFlags {
		if voice, ok := accentVoices[name]; ok {
			reverse[flag] = voice
		}
	}
	return reverse
}()

// languageVoices are the approved language default voices (C5).
var languageVoices = map[Language]string{
	English: "en-us",
	Spanish: "es",
}

// VoiceForAccent maps an accent label to an eSpeak voice (C5): the same
// case-insensitive accent names NormalizePronunciations maps to flags, with a
// language default when the label is empty, unknown, or has an unknown part.
// Labels shown as flags (with the original name as the title) also map, by
// reversing the flag table; a generic UK+US run uses the language default.
func VoiceForAccent(language Language, accentTitle string) (voice string) {
	title := strings.Join(strings.Fields(accentTitle), " ")
	if title == "" {
		return defaultVoice(language)
	}
	if isFlagRun(title) {
		if isGeneric(title) {
			return defaultVoice(language)
		}
		runes := []rune(title)
		return voiceForFlag[string(runes[0:2])]
	}
	for _, part := range strings.Split(title, ",") {
		name := strings.ToLower(strings.TrimSpace(part))
		if accentQualifiers[name] {
			continue
		}
		if mapped, ok := accentVoices[name]; ok {
			return mapped
		}
		return defaultVoice(language) // one unknown part: the whole label is unmapped
	}
	return defaultVoice(language)
}

func defaultVoice(language Language) string {
	if voice, ok := languageVoices[language]; ok {
		return voice
	}
	return "en-us"
}

// StripPronunciationDelimiters removes one matched pair of surrounding
// pronunciation delimiters (/…/ or […]) from a transcription as shown.
func StripPronunciationDelimiters(text string) string {
	text = strings.TrimSpace(text)
	if len(text) < 2 {
		return text
	}
	lead, tail := text[0], text[len(text)-1]
	if (lead == '/' && tail == '/') || (lead == '[' && tail == ']') {
		return text[1 : len(text)-1]
	}
	return text
}

// ipaMnemonics holds the per-language IPA→eSpeak mnemonic phoneme tables (C5,
// revision 3). Every entry is sourced from the engine's own
// `espeak-ng -v <voice> -q -x` output on espeak-ng 1.52 and parse-verified
// through the same oracle by the integration build; the emitted characters
// stay inside letters, :, ', @, ^, digits and space (the adapter whitelists
// that set). Two-symbol keys are matched longest-first.
var ipaMnemonics = map[Language]map[string]string{
	English: {
		// vowels, longest first in effect: length marks and r-colours
		"ɑː": "A:", "ɔː": "O:", "ɜː": "3:", "iː": "i:", "uː": "u:",
		"ɝ": "3:", "ɚ": "3r", "ɜ": "3",
		// diphthongs
		"aɪ": "aI", "eɪ": "eI", "ɔɪ": "OI", "aʊ": "aU", "əʊ": "oU", "oʊ": "oU",
		"ɪə": "i@", "ʊə": "U@", "eə": "e@", "ɛə": "e@",
		// monophthongs — the LOT vowel is the engine's "0" (canonical for
		// hot/lot/not/knowledge); "Q" parses but is a different, far quieter
		// phoneme and sounds wrong
		"æ": "a", "a": "a", "e": "E", "ɛ": "E", "ɪ": "I", "ʊ": "U", "ʌ": "V",
		"ə": "@", "ɔ": "O", "ɒ": "0", "i": "i", "o": "o", "u": "u",
		// consonants
		"tʃ": "tS", "dʒ": "dZ", "θ": "T", "ð": "D", "ŋ": "N", "ʃ": "S", "ʒ": "Z",
		"ɡ": "g", "g": "g", "ɹ": "r", "r": "r", "ɫ": "l", "l": "l", "ɾ": "R2",
		"p": "p", "b": "b", "t": "t", "d": "d", "k": "k", "f": "f", "v": "v",
		"s": "s", "z": "z", "m": "m", "n": "n", "h": "h", "w": "w", "j": "j",
	},
	Spanish: {
		// palatals and lenited fricatives (engine mnemonics)
		"ʝ": "J", "j": "J^", "y": "J^", "ʎ": "l^", "ɲ": "n^",
		"β": "B", "ð": "D", "θ": "T", "ɣ": "Q",
		// consonants
		"tʃ": "tS", "dʒ": "dZ", "ʃ": "S", "ʒ": "Z",
		"ɡ": "g", "g": "g", "ɾ": "R2", "r": "r", "ŋ": "N", // ŋ: nasal before /k/ (cinco [ˈθiŋko])
		"p": "p", "b": "b", "t": "t", "d": "d", "k": "k", "f": "f", "v": "v",
		"s": "s", "z": "z", "m": "m", "n": "n", "l": "l", "w": "w", "x": "x",
		// vowels
		"a": "a", "e": "e", "i": "i", "o": "o", "u": "u",
	},
}

// MnemonicTable returns a copy of the language's IPA→mnemonic table so the
// real-binary build check can verify every entry against the engine; false
// when the language has no table.
func MnemonicTable(language Language) (map[string]string, bool) {
	table, ok := ipaMnemonics[language]
	if !ok {
		return nil, false
	}
	out := make(map[string]string, len(table))
	for key, mnemonic := range table {
		out[key] = mnemonic
	}
	return out, true
}

// vowelMnemonics lists the mnemonic values that are vowels per language: the
// converter anchors the ' stress marker to the next vowel (the engine's own
// canonical style, e.g. word → w'3:d), so consonants must not receive it.
// Kept beside the tables; the unit tests assert every value exists in the
// language's table.
var ipaVowelMnemonics = map[Language]map[string]bool{
	English: {
		"a": true, "A:": true, "E": true, "I": true, "i": true, "i:": true,
		"U": true, "u": true, "u:": true, "V": true, "@": true, "O": true,
		"O:": true, "0": true, "3": true, "3:": true, "3r": true,
		"aI": true, "eI": true, "OI": true, "aU": true, "oU": true,
		"i@": true, "U@": true, "e@": true, "o": true,
	},
	Spanish: {
		"a": true, "e": true, "i": true, "o": true, "u": true,
	},
}

// IPAToMnemonics converts a shown IPA transcription into eSpeak mnemonic
// phonemes (C5, revision 3): delimiters stripped, syllable dots dropped, ˈ/ˌ
// become the ' stress marker anchored to the next vowel (the engine's
// canonical style: word → w'3:d, hola → 'ola), whitespace runs collapse to one
// space, and a following ː lengthens the preceding vowel mnemonic. It reports
// false when any symbol is outside the language's table: the caller then
// speaks the headword instead (S7).
func IPAToMnemonics(language Language, ipa string) (string, bool) {
	table := ipaMnemonics[language]
	if table == nil {
		return "", false
	}
	vowels := ipaVowelMnemonics[language]
	// longest keys first: aɪ before a, ɜː before ɜ, ɪə before ɪ
	keys := make([]string, 0, len(table))
	for key := range table {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return utf8.RuneCountInString(keys[i]) > utf8.RuneCountInString(keys[j])
	})
	var pieces []string // mnemonics and explicit " " separators, joined empty
	stress := false     // pending ˈ/ˌ → ' before the next vowel mnemonic
	text := strings.Join(strings.Fields(StripPronunciationDelimiters(ipa)), " ")
	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])
		switch {
		case r == 'ˈ' || r == 'ˌ':
			stress = true
			i += size
			continue
		case r == '.':
			i += size
			continue
		case unicode.IsSpace(r):
			pieces = append(pieces, " ")
			i += size
			continue
		}
		matched := ""
		for _, key := range keys {
			if strings.HasPrefix(text[i:], key) {
				matched = table[key]
				i += len(key)
				break
			}
		}
		if matched == "" {
			return "", false // symbol outside the table: the caller falls back
		}
		// ː (or the ASCII variant) merges with the preceding vowel mnemonic
		if r, size := utf8.DecodeRuneInString(text[i:]); r == 'ː' || r == ':' {
			last, _ := utf8.DecodeLastRuneInString(matched)
			switch {
			case last == ':': // already long (ɝː → 3:): the mark is consumed
			case strings.ContainsRune("aeiouAEIOU@", last):
				matched += ":"
			default:
				return "", false
			}
			i += size
		}
		if stress && vowels[matched] {
			matched = "'" + matched
			stress = false
		}
		pieces = append(pieces, matched)
	}
	if stress { // trailing stress with no vowel after it: keep it on the last mnemonic
		for i := len(pieces) - 1; i >= 0; i-- {
			if pieces[i] != " " {
				pieces[i] = "'" + pieces[i]
				break
			}
		}
	}
	out := strings.Join(pieces, "")
	if strings.TrimSpace(out) == "" {
		return "", false // nothing speakable was left
	}
	return out, true
}

// audioDataURI encodes audio bytes as a data URI for the webview (C3).
func audioDataURI(mime string, data []byte) string {
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
}
