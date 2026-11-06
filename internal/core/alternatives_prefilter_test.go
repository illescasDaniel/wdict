package core

import (
	"context"
	"strings"
	"testing"
)

func TestGivenUnicodeAndPunctuationWhenSearchingThenSafeLengthPrefilterRetainsMatches(t *testing.T) {
	// Given candidates whose raw byte length exceeds the folded query length.
	idx := &fakeAlternativesIndex{words: []string{"café", "ca-fé", "caffè"}}
	// When the query requires one folded character insertion.
	got, err := (&AlternativesService{Index: idx}).FindAlternatives(context.Background(), AlternativesRequest{Query: "cfe", Language: English})
	// Then both accent and punctuation variants remain eligible.
	if err != nil || got.Status != AlternativesReady {
		t.Fatalf("result=%+v err=%v", got, err)
	}
	for _, want := range []string{"café", "ca-fé"} {
		if !containsAlternative(got.Spelling, want) {
			t.Errorf("%q missing from %q", want, got.Spelling)
		}
	}
}

func TestGivenExpandingUnicodeFoldWhenSearchingThenDistanceScratchRemainsBounded(t *testing.T) {
	// Given a valid 64-rune query whose case fold expands to 128 runes.
	query := strings.Repeat("ß", 64)
	idx := &fakeAlternativesIndex{words: []string{strings.Repeat("s", 127) + "t"}}
	// When a candidate is scanned.
	got, err := (&AlternativesService{Index: idx}).FindAlternatives(context.Background(), AlternativesRequest{Query: query, Language: English})
	// Then matching finishes without overflowing the bounded scratch array.
	if err != nil || got.Status != AlternativesReady || len(got.Spelling) != 1 {
		t.Fatalf("result=%+v err=%v", got, err)
	}
}
