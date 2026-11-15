package core

import (
	"golang.org/x/text/cases"
	"strings"
)

var properPOS = map[string]bool{"proper noun": true, "nombre propio": true, "sustantivo propio": true, "sustantivo propio masculino": true, "sustantivo propio femenino": true}
var ordinaryPOS = func() map[string]bool {
	m := map[string]bool{}
	for _, p := range []string{"noun", "verb", "adjective", "adverb", "pronoun", "determiner", "article", "preposition", "conjunction", "interjection", "numeral", "particle", "sustantivo", "sustantivo masculino", "sustantivo femenino", "sustantivo masculino y femenino", "verbo", "adjetivo", "adverbio", "pronombre", "determinante", "artículo", "preposición", "conjunción", "interjección", "partícula"} {
		m[p] = true
	}
	return m
}()

func rankPOS(e Entry, fold cases.Caser) int {
	if len(e.Meanings) == 0 {
		return 0
	}
	ordinary := false
	for _, m := range e.Meanings {
		p := fold.String(strings.Join(strings.Fields(m.PartOfSpeech), " "))
		if ordinaryPOS[p] {
			ordinary = true
		} else if !properPOS[p] {
			return 0
		}
	}
	if ordinary {
		return 1
	}
	return 2
}

// RankEntries orders a copy of candidates for this query without mutating entries.
func RankEntries(query string, entries []Entry) []Entry {
	query = strings.TrimSpace(query)
	key := FoldWord(query)
	fold := cases.Fold()
	caseQuery := fold.String(query)
	var tiers [4][]Entry
	for _, e := range entries {
		tier := 3
		switch {
		case e.Headword == query:
			tier = 0
		case fold.String(e.Headword) == caseQuery:
			tier = 1
		case key != "" && FoldWord(e.Headword) == key:
			tier = 2
		}
		tiers[tier] = append(tiers[tier], e)
	}
	// Keep neutral slots fixed while stably partitioning known records.
	var known []Entry
	classes := make([]int, len(tiers[1]))
	for i, e := range tiers[1] {
		classes[i] = rankPOS(e, fold)
	}
	for i, e := range tiers[1] {
		if classes[i] == 1 {
			known = append(known, e)
		}
	}
	for i, e := range tiers[1] {
		if classes[i] == 2 {
			known = append(known, e)
		}
	}
	i := 0
	for j := range tiers[1] {
		if classes[j] != 0 {
			tiers[1][j] = known[i]
			i++
		}
	}
	out := make([]Entry, 0, len(entries))
	for _, tier := range tiers {
		out = append(out, tier...)
	}
	return out
}
