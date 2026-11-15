package core

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
	"strings"
	"unicode"
)

// foldWord mirrors the default StarDict case, whitespace, accent and punctuation folding.
func FoldWord(word string) string {
	if asciiWord(word) {
		var b strings.Builder
		space := false
		started := false
		for _, c := range word {
			if unicode.IsSpace(c) {
				if started {
					space = true
				}
				continue
			}
			if space {
				b.WriteByte(' ')
				space = false
			}
			started = true
			if unicode.IsPunct(c) {
				continue
			}
			if c >= 'A' && c <= 'Z' {
				c += 'a' - 'A'
			}
			b.WriteRune(c)
		}
		return b.String()
	}
	word = cases.Fold().String(norm.NFD.String(word))
	word = strings.Join(strings.Fields(word), " ")
	word = strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) || unicode.IsPunct(r) {
			return -1
		}
		return r
	}, word)
	return norm.NFC.String(word)
}

func asciiWord(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 128 {
			return false
		}
	}
	return true
}
