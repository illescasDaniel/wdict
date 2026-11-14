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

// NormalizePiperIPA converts display IPA into the raw phone sequence Piper
// expects: syllable stress moves from the syllable onset to its vowel and a
// transcription without any stress mark gains the default primary stress
// eSpeak's own output always carries — per word of a multi-word stream (eight
// → 'eIt, to say nothing of → t@ s'eI n'VTIN Vv; verified against espeak-ng
// 1.52) — Wiktionary omits stress on monosyllables and phrase words, but the
// voices were trained on eSpeak's stressed shapes and render unstressed feeds
// weakly or garbled. Syllable dots, affix hyphens, affricate tie bars and
// word-join ties (‿) are removed, whitespace feeds as one space, nasal vowels
// are decomposed without losing their phonemic tilde, precomposed umlaut
// vowels fold to their base, the Indian-English retroflex series folds to the
// nearest standard phone, and syllabic sonorants become the model's
// schwa-plus-sonorant sequence. Phones no engine of the language speaks are
// rejected so playback falls back instead of synthesising untrained noise.
func NormalizePiperIPA(language Language, ipa string) (string, bool) {
	if language != English && language != Spanish {
		return "", false
	}
	runes := []rune(StripPronunciationDelimiters(ipa))
	// English word-initial unstressed ə becomes ɐ (post-0.8.0 fix, user
	// report: "appearance" /əˈpiɹ.əns/ loses its start): both pinned voices
	// are trained on eSpeak's text-mode output, whose English dictionaries
	// write the initial prefix vowel as ɐ (about → ɐbˈaʊt, appearance →
	// ɐpˈiəɹəns, occasion → əkˈeɪʒən). The models render a word-initial ə
	// as near-silence — verified against the installed runtime, where the
	// ə-variant's audio is the vowel-less audio shifted ~80 ms later —
	// which reads as the word missing its start. ɐ is the same unstressed
	// central vowel to the ear; a mid-word ə is common in training and
	// stays as written.
	// GOAT diphthong: the engine realises əʊ as oʊ in every position (its own
	// mnemonic table folds əʊ → oU, espeak-ng 1.52 verified), and the voices'
	// cluster list merges oʊ as one trained phone while ə+ʊ stays an untrained
	// split pair (oak /əʊk/ fed ɐʊk before this fold). Runs before the
	// word-initial rule below so the diphthong is folded, not split.
	if language == English {
		for i := 0; i+1 < len(runes); i++ {
			if runes[i] == 'ə' && runes[i+1] == 'ʊ' {
				runes[i] = 'o'
			}
		}
	}
	if language == English && len(runes) > 0 && runes[0] == 'ə' {
		runes[0] = 'ɐ'
	}
	var out strings.Builder
	out.Grow(len(ipa))
	pendingStress := rune(0)
	injectStress := !strings.ContainsRune(string(runes), 'ˈ') && !strings.ContainsRune(string(runes), 'ˌ')
	if injectStress {
		// No stress mark anywhere: eSpeak's text-mode output — the voices'
		// training data — always marks an isolated word's stress (eight →
		// 'eIt, cat → k'at, up → 'Vp) even where Wiktionary's transcription
		// omits it (/eɪt/), and stresses every word of a stream. Feed the
		// canonical shape: default primary stress on each word's first
		// vowel. Unstressed feeds render weakly (the /eɪt/ report lost its
		// diphthong) and short unstressed sequences can collapse to noise
		// (the /eʈ/ report).
		pendingStress = 'ˈ'
	}
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch r {
		case '.', '͡', '͜', '-', '‿':
			// ‿ is the liaison/word-join tie (wɪ‿ðə): notation the models
			// never saw — the joined sounds flow as one prosodic word
			continue
		case 'ⁿ':
			// superscript nasal-release marker (Wednesday /ˈwɛdⁿnz.deɪ/):
			// spoken as n — no voice map carries the marker itself — unless a
			// real nasal already follows, where it is redundant
			if i+1 < len(runes) && strings.ContainsRune("nmŋ", runes[i+1]) {
				continue
			}
			out.WriteRune('n')
		case 'ˈ', 'ˌ':
			if pendingStress != 0 {
				return "", false
			}
			pendingStress = r
			continue
		case 'ã', 'ẽ', 'ĩ', 'õ', 'ũ':
			if pendingStress != 0 {
				out.WriteRune(pendingStress)
				pendingStress = 0
			}
			base := r
			switch r {
			case 'ã':
				base = 'a'
			case 'ẽ':
				base = 'e'
			case 'ĩ':
				base = 'i'
			case 'õ':
				base = 'o'
			case 'ũ':
				base = 'u'
			}
			out.WriteRune(base)
			out.WriteRune('̃')
		case 'ä', 'ö', 'ü', 'ë', 'ï':
			// precomposed umlaut vowels are centralisation qualifiers, not
			// separate sounds (out /ˈäʊ̯t/): fold to the base vowel, unlike
			// nasality the model inventory has no mark for them
			if pendingStress != 0 {
				out.WriteRune(pendingStress)
				pendingStress = 0
			}
			switch r {
			case 'ä':
				r = 'a'
			case 'ö':
				r = 'o'
			case 'ü':
				r = 'u'
			case 'ë':
				r = 'e'
			case 'ï':
				r = 'i'
			}
			out.WriteRune(r)
		case 'ʈ', 'ɖ', 'ʂ', 'ʐ', 'ɳ', 'ɭ', 'ɽ':
			// Indian-English retroflex series (eight /eʈ/): the voice maps
			// carry the symbols but no English voice was trained on them —
			// they synthesise noise. Fold to the nearest standard phone.
			switch r {
			case 'ʈ':
				r = 't'
			case 'ɖ':
				r = 'd'
			case 'ʂ':
				r = 's'
			case 'ʐ':
				r = 'z'
			case 'ɳ':
				r = 'n'
			case 'ɭ':
				r = 'l'
			case 'ɽ':
				r = 'ɾ'
			}
			out.WriteRune(r)
		case 'l', 'ɫ', 'n', 'm':
			if i+1 < len(runes) && runes[i+1] == '̩' {
				if pendingStress != 0 {
					out.WriteRune(pendingStress)
					pendingStress = 0
				}
				out.WriteRune('ə')
				out.WriteRune(r)
				i++
				continue
			}
			out.WriteRune(r)
		default:
			if unicode.IsSpace(r) {
				// multi-word IPA (fat finger, /aɪ piː siː siː/) is real
				// dictionary data and the training text has word spaces; any
				// whitespace feeds as one space and arms the next word's
				// default stress (eSpeak's phrase output stresses every word)
				out.WriteRune(' ')
				if injectStress {
					pendingStress = 'ˈ'
				}
				continue
			}
			if !piperSpeakable(language, r) {
				return "", false // no engine of this language speaks it: fall back
			}
			if pendingStress != 0 && isIPAVowel(r) {
				out.WriteRune(pendingStress)
				pendingStress = 0
			}
			out.WriteRune(r)
		}
	}
	if pendingStress != 0 {
		return "", false
	}
	return out.String(), true
}

func isIPAVowel(r rune) bool {
	// Cover the IPA vowel symbols in the pinned Piper/espeak inventory, not
	// only the common English/Spanish subset. Stress must move past an onset
	// to the first nucleus even when the transcription uses less common vowels.
	return strings.ContainsRune("iyɨʉɯuɪʏʊeøɘɵɤoəɛœɜɞʌɔæɐaɶɑɒɚɝᵻ", r)
}

// piperNotation lists notation every normalized Piper stream may carry. The
// symbols are in the voice maps or handled by the adapter's diacritic drop.
var piperNotation = map[rune]bool{'ˈ': true, 'ˌ': true, 'ː': true, '(': true, ')': true}

// piperExtras lists phones the pinned voices demonstrably speak that sit
// outside the eSpeak mnemonic tables: ɐ (eSpeak's own word-initial unstressed
// vowel), ʰ aspiration, ʍ (which), x (loch), ʔ (glottal stop) and ᵻ — all in
// the installed voice maps and exercised by real dictionary transcriptions.
// Anything else beyond the mnemonic tables is exotic (retroflexes are folded
// before this check; clicks and implosives are not) and must fall back: the
// maps carry every IPA symbol, so the model would happily synthesise
// untrained noise for them (the eight /eʈ/ report).
var piperExtras = map[Language]map[rune]bool{
	English: {'ɐ': true, 'ʰ': true, 'ʍ': true, 'x': true, 'ʔ': true, 'ᵻ': true},
	Spanish: {},
}

// piperTableRunes expands the per-language IPA→mnemonic table keys into the
// set of base phones the language's engines speak.
var piperTableRunes = func() map[Language]map[rune]bool {
	out := make(map[Language]map[rune]bool, len(ipaMnemonics))
	for language, table := range ipaMnemonics {
		runes := make(map[rune]bool, len(table)*2)
		for key := range table {
			for _, r := range key {
				runes[r] = true
			}
		}
		out[language] = runes
	}
	return out
}()

// piperSpeakable reports whether the language's engines speak one rune of a
// normalized transcription: notation, combining diacritics (the adapter drops
// those the model lacks), the mnemonic tables' phones, or a documented extra.
func piperSpeakable(language Language, r rune) bool {
	if piperNotation[r] || unicode.Is(unicode.Mn, r) {
		return true
	}
	if extras := piperExtras[language]; extras[r] {
		return true
	}
	runes := piperTableRunes[language]
	return runes[r]
}

// stripNonPhonemic prepares IPA for eSpeak's mnemonic converter. eSpeak does
// not accept IPA combining notation in its mnemonic input, so syllable dots,
// combining marks and nasalization marks are omitted only on that fallback
// path. The Piper path preserves phonemic diacritics through NormalizePiperIPA.
func stripNonPhonemic(text string) string {
	runes := []rune(text)
	var out strings.Builder
	out.Grow(len(text))
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if (r == 'l' || r == 'ɫ' || r == 'n' || r == 'm') &&
			i+1 < len(runes) && runes[i+1] == '̩' {
			out.WriteRune('ə')
			out.WriteRune(r)
			i++
			continue
		}
		if r == '.' || r == '͡' || r == '͜' || unicode.Is(unicode.Mn, r) {
			continue
		}
		switch r {
		case 'ã':
			r = 'a'
		case 'ẽ':
			r = 'e'
		case 'ĩ':
			r = 'i'
		case 'õ':
			r = 'o'
		case 'ũ':
			r = 'u'
		}
		out.WriteRune(r)
	}
	return out.String()
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
		// Indian-English retroflex series: fold to the nearest standard
		// phone (matches NormalizePiperIPA, whose Piper feed folds the same)
		"ʈ": "t", "ɖ": "d", "ʂ": "s", "ʐ": "z", "ɳ": "n", "ɭ": "l", "ɽ": "R2",
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
