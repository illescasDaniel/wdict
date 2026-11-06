package core

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

type fakeAlternativesIndex struct {
	words        []string
	verification map[string]string
	scanErr      error
	scanCalls    int
	verifyCalls  int
}

func (f *fakeAlternativesIndex) VerifyHeadword(ctx context.Context, candidate string, language Language) (string, bool, error) {
	f.verifyCalls++
	if err := ctx.Err(); err != nil {
		return "", false, err
	}
	word, ok := f.verification[candidate]
	return word, ok, nil
}
func (f *fakeAlternativesIndex) IterateHeadwords(ctx context.Context, language Language, visit func(string) error) error {
	f.scanCalls++
	for _, word := range f.words {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := visit(word); err != nil {
			return err
		}
	}
	return f.scanErr
}

func alternativesForTest(index HeadwordIndex) Alternatives { return &AlternativesService{Index: index} }

func TestGivenInflectionWhenAlternativesFoundThenOnlyVerifiedBasesAreReturned(t *testing.T) {
	// Given a miss whose regular forms include two verified candidates.
	idx := &fakeAlternativesIndex{words: []string{"walk", "walke"}, verification: map[string]string{"walk": "walk", "walke": "walke"}}
	req := AlternativesRequest{SearchID: "s1", Attempt: 2, Query: "walked", Language: English}
	// When alternatives are found.
	got, err := alternativesForTest(idx).FindAlternatives(context.Background(), req)
	// Then both candidates are preserved and the attempt identity survives.
	if err != nil || got.Status != AlternativesReady {
		t.Fatalf("got %+v, err %v", got, err)
	}
	if got.SearchID != req.SearchID || got.Attempt != req.Attempt || got.Query != req.Query || got.Language != req.Language {
		t.Fatalf("identity changed: %+v", got)
	}
	if !reflect.DeepEqual(got.BaseForms, []string{"walk", "walke"}) {
		t.Fatalf("bases = %q", got.BaseForms)
	}
}

func TestGivenSpanishPluralWhenBaseIsNotInDictionaryThenItIsNotSuggested(t *testing.T) {
	// Given a Spanish plural and an index containing no possible base.
	idx := &fakeAlternativesIndex{}
	// When finding alternatives.
	got, err := alternativesForTest(idx).FindAlternatives(context.Background(), AlternativesRequest{Query: "casas", Language: Spanish})
	// Then unverified generated strings are not returned.
	if err != nil || got.Status != AlternativesReady || len(got.BaseForms) != 0 || len(got.Spelling) != 0 {
		t.Fatalf("got %+v, err %v", got, err)
	}
}

func TestGivenSpellingErrorsWhenAlternativesFoundThenExpectedWordsAndPhrasesRankWithinFive(t *testing.T) {
	cases := []struct {
		query    string
		language Language
		words    []string
		expected string
	}{
		{"wrods", English, []string{"prods", "rods", "words", "wolds", "wods", "cords", "words"}, "words"},
		{"arbl", Spanish, []string{"Arbe", "arbol", "arel", "árbol", "arél", "Arbo"}, "árbol"},
		{"perrro", Spanish, []string{"peorro", "peroro", "peroró", "perreo", "perrero", "perro"}, "perro"},
		{"gatto", Spanish, []string{"atto-", "gasto", "gastó", "gateo", "gateó", "gato"}, "gato"},
		{"look foward to", English, []string{"look forward to", "looks forward to", "look", "forward"}, "look forward to"},
	}
	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			// Given dictionary headwords and an absent query.
			idx := &fakeAlternativesIndex{words: tc.words, verification: map[string]string{}}
			for _, word := range tc.words {
				idx.verification[word] = word
			}
			// When a complete suggestion search runs.
			got, err := alternativesForTest(idx).FindAlternatives(context.Background(), AlternativesRequest{Query: tc.query, Language: tc.language})
			// Then the expected correction is within the five distinct suggestions.
			if err != nil || got.Status != AlternativesReady {
				t.Fatalf("got %+v, err %v", got, err)
			}
			if len(got.Spelling) > 5 {
				t.Fatalf("too many suggestions: %q", got.Spelling)
			}
			found := false
			unique := map[string]bool{}
			for _, word := range got.Spelling {
				if word == tc.expected {
					found = true
				}
				if unique[word] {
					t.Fatalf("duplicate %q", word)
				}
				unique[word] = true
			}
			if !found {
				t.Fatalf("expected %q in %q", tc.expected, got.Spelling)
			}
		})
	}
}

func TestGivenSameIndexWhenFindingAlternativesTwiceThenOrderingIsDeterministic(t *testing.T) {
	// Given candidates in an arbitrary order with distinct display spellings.
	idx := &fakeAlternativesIndex{words: []string{"arél", "árbol", "Arbo", "arbol", "arel"}, verification: map[string]string{"arél": "arél", "árbol": "árbol", "Arbo": "Arbo", "arbol": "arbol", "arel": "arel"}}
	req := AlternativesRequest{Query: "arbl", Language: Spanish}
	// When repeating the same search.
	first, e1 := alternativesForTest(idx).FindAlternatives(context.Background(), req)
	second, e2 := alternativesForTest(idx).FindAlternatives(context.Background(), req)
	// Then output order is identical.
	if e1 != nil || e2 != nil || !reflect.DeepEqual(first.Spelling, second.Spelling) {
		t.Fatalf("first %q (%v), second %q (%v)", first.Spelling, e1, second.Spelling, e2)
	}
}

func TestGivenOverlongSuggestionQueryWhenSearchingThenIndexIsNotScanned(t *testing.T) {
	// Given a 65-rune query which remains valid for ordinary lookup.
	idx := &fakeAlternativesIndex{}
	// When alternatives are requested.
	got, err := alternativesForTest(idx).FindAlternatives(context.Background(), AlternativesRequest{Query: strings.Repeat("a", 65), Language: English})
	// Then suggestions are unavailable without touching the index.
	if err != nil || got.Status != AlternativesUnavailable || idx.scanCalls != 0 || idx.verifyCalls != 0 {
		t.Fatalf("got %+v, err %v, scan=%d verify=%d", got, err, idx.scanCalls, idx.verifyCalls)
	}
}

func TestGivenInvalidQueryWhenSearchingThenIndexIsNotScanned(t *testing.T) {
	for _, query := range []string{"", "hello\n", string([]byte{0xff})} {
		t.Run(query, func(t *testing.T) {
			// Given an invalid query.
			idx := &fakeAlternativesIndex{}
			// When alternatives are requested.
			_, err := alternativesForTest(idx).FindAlternatives(context.Background(), AlternativesRequest{Query: query, Language: English})
			// Then the request is rejected without index access.
			if idx.scanCalls != 0 || idx.verifyCalls != 0 {
				t.Fatalf("err=%v scan=%d verify=%d", err, idx.scanCalls, idx.verifyCalls)
			}
		})
	}
}

func TestGivenUnreadableIndexWhenSearchingThenUnavailableHasNoPartialCandidates(t *testing.T) {
	// Given an index that fails after yielding a candidate.
	idx := &fakeAlternativesIndex{words: []string{"words"}, verification: map[string]string{"words": "words"}, scanErr: errors.New("corrupt index")}
	// When alternatives are requested.
	got, err := alternativesForTest(idx).FindAlternatives(context.Background(), AlternativesRequest{Query: "wrods", Language: English})
	// Then no partial results escape.
	if err != nil || got.Status != AlternativesUnavailable || len(got.Spelling) != 0 || len(got.BaseForms) != 0 {
		t.Fatalf("got %+v, err %v", got, err)
	}
}

func TestGivenCancelledRequestWhenSearchingThenNoIndexIsScanned(t *testing.T) {
	// Given a caller-cancelled search.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	idx := &fakeAlternativesIndex{}
	// When alternatives are requested.
	_, err := alternativesForTest(idx).FindAlternatives(ctx, AlternativesRequest{Query: "wrods", Language: English})
	// Then cancellation is returned and no work is started.
	if !errors.Is(err, context.Canceled) || idx.scanCalls != 0 || idx.verifyCalls != 0 {
		t.Fatalf("err=%v scan=%d verify=%d", err, idx.scanCalls, idx.verifyCalls)
	}
}

// TestGivenTimedOutScanWhenSearchingThenNoPartialRankingIsReturned covers F4
// without sleeping by returning the deadline error from the fake index.
func TestGivenTimedOutScanWhenSearchingThenNoPartialRankingIsReturned(t *testing.T) {
	// Given a scan that times out after yielding a candidate.
	idx := &fakeAlternativesIndex{words: []string{"words"}, verification: map[string]string{"words": "words"}, scanErr: context.DeadlineExceeded}
	// When finding alternatives.
	got, err := alternativesForTest(idx).FindAlternatives(context.Background(), AlternativesRequest{Query: "wrods", Language: English})
	// Then no partial candidate ranking is returned.
	if err != nil || got.Status != AlternativesTimedOut || len(got.Spelling) != 0 || len(got.BaseForms) != 0 {
		t.Fatalf("got %+v, err %v", got, err)
	}
}

func TestGivenNoDictionaryWhenSearchingThenSuggestionsAreUnavailable(t *testing.T) {
	// Given no installed index for this language.
	// When alternatives are requested.
	got, err := alternativesForTest(nil).FindAlternatives(context.Background(), AlternativesRequest{Query: "wrods", Language: English})
	// Then suggestions are unavailable without a scan.
	if err != nil || got.Status != AlternativesUnavailable || len(got.Spelling) != 0 || len(got.BaseForms) != 0 {
		t.Fatalf("got %+v, err %v", got, err)
	}
}
