package core

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	maxPronunciations    = 3
	maxPronunciationLen  = 64
	maxAccentLabelLen    = 32  // limit on the label as shown (after flag mapping)
	maxRawAccentLabelLen = 128 // limit on the source label before mapping
)

// NormalizePronunciations applies C2 (validate), C3 (order, merge, cap 3) and
// C4 (delimiters, flag labels, LabelTitle). It is idempotent. It never returns an error: bad items are dropped.
func NormalizePronunciations(raw []Pronunciation) []Pronunciation {
	var out []Pronunciation
	seen := make(map[string]bool, maxPronunciations)
	for _, item := range raw {
		text, ok := cleanPronunciationText(item.Text, maxPronunciationLen)
		if !ok {
			continue
		}
		if !strings.HasPrefix(text, "/") && !strings.HasPrefix(text, "[") {
			text = "/" + text + "/"
			if utf8.RuneCountInString(text) > maxPronunciationLen {
				continue
			}
		}
		if seen[text] {
			continue
		}
		seen[text] = true
		label, title := normalizeAccentLabel(item.Label, item.LabelTitle)
		out = append(out, Pronunciation{Text: text, Label: label, LabelTitle: title})
		if len(out) == maxPronunciations {
			break
		}
	}
	return out
}

// cleanPronunciationText trims and collapses whitespace, then rejects empty,
// too long, markup-like or control-character text.
func cleanPronunciationText(s string, limit int) (string, bool) {
	s = strings.Join(strings.Fields(s), " ")
	if s == "" || utf8.RuneCountInString(s) > limit || !utf8.ValidString(s) {
		return "", false
	}
	for _, r := range s {
		if r == '<' || r == '>' || unicode.IsControl(r) {
			return "", false
		}
	}
	return s, true
}

// accentFlags maps known accent names (lower case) to flags (C4).
var accentFlags = map[string]string{
	"general american":       "🇺🇸",
	"us":                     "🇺🇸",
	"received pronunciation": "🇬🇧",
	"uk":                     "🇬🇧",
	"australia":              "🇦🇺",
	"new zealand":            "🇳🇿",
	"canada":                 "🇨🇦",
	"ireland":                "🇮🇪",
	"south africa":           "🇿🇦",
}

// normalizeAccentLabel validates a label and maps it to flags when every
// comma-separated part is known; the original text then becomes the title.
// A label that already is a run of known flags is kept with its title.
func normalizeAccentLabel(rawLabel, rawTitle string) (label, title string) {
	label, ok := cleanPronunciationText(rawLabel, maxRawAccentLabelLen)
	if !ok {
		return "", ""
	}
	if isFlagRun(label) {
		if title, ok = cleanPronunciationText(rawTitle, maxRawAccentLabelLen); !ok {
			title = ""
		}
		if isGeneric(label) {
			return "", ""
		}
		return limitShownLabel(label, title)
	}
	var flags strings.Builder
	for _, part := range strings.Split(label, ",") {
		name := strings.ToLower(strings.TrimSpace(part))
		if accentQualifiers[name] {
			continue
		}
		flag, known := accentFlags[name]
		if !known {
			return limitShownLabel(label, "")
		}
		flags.WriteString(flag)
	}
	if flags.Len() == 0 {
		return limitShownLabel(label, "")
	}
	if isGeneric(flags.String()) {
		return "", ""
	}
	return limitShownLabel(flags.String(), label)
}

// accentQualifiers are words that are not accents and are ignored when mapping (C4, R4).
var accentQualifiers = map[string]bool{"dialectal": true, "obsolete": true}

// isGeneric reports whether flags include both UK and US: the common form, shown with no label.
func isGeneric(flags string) bool {
	return strings.Contains(flags, "🇬🇧") && strings.Contains(flags, "🇺🇸")
}

// limitShownLabel blanks a label (and its title) whose displayed form is over
// maxAccentLabelLen runes.
func limitShownLabel(label, title string) (string, string) {
	if utf8.RuneCountInString(label) > maxAccentLabelLen {
		return "", ""
	}
	return label, title
}

// isFlagRun reports whether s is made only of known accent flags.
func isFlagRun(s string) bool {
	runes := []rune(s)
	if len(runes) == 0 || len(runes)%2 != 0 {
		return false
	}
	for i := 0; i < len(runes); i += 2 {
		known := false
		for _, flag := range accentFlags {
			if string(runes[i:i+2]) == flag {
				known = true
				break
			}
		}
		if !known {
			return false
		}
	}
	return true
}
