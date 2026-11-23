package core

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestGivenThreeLanguagesWhenFrenchIsSlowThenOthersCompleteIndependently(t *testing.T) {
	// Given French is delayed while English and Spanish are available (S1).
	release := make(chan struct{})
	t.Cleanup(func() {
		select {
		case <-release:
		default:
			close(release)
		}
	})
	outcomes := make(chan LanguageOutcome, 3)
	frenchStarted := make(chan struct{})
	source := languageSource(func(ctx context.Context, q string, l Language) ([]Entry, bool, error) {
		if l == French {
			close(frenchStarted)
			select {
			case <-release:
			case <-ctx.Done():
				return nil, false, ctx.Err()
			}
		}
		return []Entry{{Headword: q, Language: l, Meanings: []Meaning{{Definitions: []string{"definition"}}}}}, true, nil
	})
	svc := BilingualService{Sources: bilingualResolver{English: {Online: source}, Spanish: {Online: source}, French: {Online: source}}, Clock: fakeClock{now: time.Now()}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- svc.Lookup(ctx, BilingualRequest{ID: "1", Query: "bonjour"}, func(o LanguageOutcome) { outcomes <- o })
	}()
	// When both quick sources complete, French must still be pending.
	seen := map[Language]bool{}
	for len(seen) < 2 {
		select {
		case o := <-outcomes:
			if o.Result.Status != Found || o.Language == French {
				t.Fatalf("unexpected result: %+v", o)
			}
			seen[o.Language] = true
		case <-ctx.Done():
			t.Fatal("fast languages did not complete independently")
		}
	}
	// Then lookup must not complete until French has answered: French is
	// known to be blocked on release once frenchStarted is closed.
	select {
	case <-frenchStarted:
	case <-ctx.Done():
		t.Fatal("French lookup never started")
	}
	select {
	case err := <-done:
		t.Fatalf("lookup returned before French completed: %v", err)
	default:
	}
	close(release)
	select {
	case o := <-outcomes:
		if o.Language != French || o.Result.Status != Found {
			t.Fatalf("French result: %+v", o)
		}
	case <-ctx.Done():
		t.Fatal("French result did not arrive after release")
	}
	if err := <-done; err != nil {
		t.Fatalf("lookup: %v", err)
	}
}

func TestGivenFrenchRetryWhenRequestedThenOtherLanguagesAreNotQueried(t *testing.T) {
	// Given a resolver tracking each language (S4).
	var mu sync.Mutex
	calls := map[Language]int{}
	source := languageSource(func(_ context.Context, _ string, l Language) ([]Entry, bool, error) {
		mu.Lock()
		defer mu.Unlock()
		calls[l]++
		return nil, false, nil
	})
	svc := BilingualService{Sources: bilingualResolver{English: {Online: source}, Spanish: {Online: source}, French: {Online: source}}, Clock: fakeClock{now: time.Now()}}
	var got LanguageOutcome
	// When retrying French.
	err := svc.RetryLanguage(context.Background(), BilingualRequest{ID: "1", Query: "bonjour"}, French, func(o LanguageOutcome) { got = o })
	// Then only French is looked up and delivered.
	if err != nil || calls[French] != 1 || calls[English] != 0 || calls[Spanish] != 0 || got.Language != French || got.Result.Status != NotFound {
		t.Fatalf("calls=%v got=%+v err=%v", calls, got, err)
	}
}

func TestGivenFrenchPluralsWhenAlternativesFoundThenOnlyIndexedBasesAppear(t *testing.T) {
	// Given nonrecursive French plural rules and a controlled index (S12).
	for _, tc := range []struct{ query, base string }{{"chats", "chat"}, {"chevaux", "cheval"}, {"bijoux", "bijou"}} {
		t.Run(tc.query, func(t *testing.T) {
			idx := &fakeAlternativesIndex{words: []string{tc.base, tc.base}}
			got, err := alternativesForTest(idx).FindAlternatives(context.Background(), AlternativesRequest{Query: tc.query, Language: French})
			if err != nil || got.Status != AlternativesReady || !reflect.DeepEqual(got.BaseForms, []string{tc.base}) {
				t.Fatalf("got=%+v err=%v", got, err)
			}
		})
	}
}

func TestGivenFrenchIndexWhenFindingAlternativesThenAccentsAndPhraseBoundariesArePreserved(t *testing.T) {
	// Given a miss and French phrase and word headwords (S13).
	idx := &fakeAlternativesIndex{words: []string{"été", "etre", "être", "à bientôt", "bientôt"}}
	service := alternativesForTest(idx)
	first, e1 := service.FindAlternatives(context.Background(), AlternativesRequest{Query: "êtr", Language: French})
	second, e2 := service.FindAlternatives(context.Background(), AlternativesRequest{Query: "êtr", Language: French})
	if e1 != nil || e2 != nil || first.Status != AlternativesReady || !reflect.DeepEqual(first.Spelling, second.Spelling) || len(first.Spelling) > 5 {
		t.Fatalf("first=%+v second=%+v errors=%v/%v", first, second, e1, e2)
	}
	if !strings.Contains(strings.Join(first.Spelling, "|"), "être") {
		t.Fatalf("accented headword missing: %q", first.Spelling)
	}
	phrase, err := service.FindAlternatives(context.Background(), AlternativesRequest{Query: "a bientotx", Language: French})
	if err != nil || phrase.Status != AlternativesReady || !reflect.DeepEqual(phrase.Spelling, []string{"à bientôt"}) {
		t.Fatalf("phrase=%+v err=%v", phrase, err)
	}
}

func TestGivenFrenchIndexFailureWhenFindingAlternativesThenNoPartialRankingIsReturned(t *testing.T) {
	// Given a French index whose scan times out.
	idx := &fakeAlternativesIndex{words: []string{"chat"}, scanErr: context.DeadlineExceeded}
	// When finding alternatives.
	got, err := alternativesForTest(idx).FindAlternatives(context.Background(), AlternativesRequest{Query: "chats", Language: French})
	// Then the lookup reports a timeout without partial rankings.
	if err != nil || got.Status != AlternativesTimedOut || len(got.BaseForms) != 0 || len(got.Spelling) != 0 {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestGivenFrenchIPAWhenPiperNormalizesThenNasalityAndExplicitStressRemain(t *testing.T) {
	// Given French nasal vowels and an explicit stress marker (S11).
	got, ok := NormalizePiperIPA(French, "bɔ̃ʒˈuʁ")
	if !ok || !strings.Contains(got, "̃") || !strings.Contains(got, "ˈu") {
		t.Fatalf("normalized=%q ok=%v", got, ok)
	}
}

func TestGivenFrenchIPAWhenESpeakPronouncesThenNasalVowelIsNotDiscarded(t *testing.T) {
	// Given an available French eSpeak voice that accepts a nasal phoneme (S11).
	voices := &fakeVoices{available: true, outputs: phonemeResult("fr-fr", "bO~")}
	result := service(voices).Play(context.Background(), PlayRequest{Language: French, Text: "/bɔ̃/", Headword: "bon"})
	if result.Source != PlayVoice || len(voices.calls) != 1 || !voices.calls[0].phonemes || !strings.Contains(voices.calls[0].text, "~") {
		t.Fatalf("result=%+v calls=%+v", result, voices.calls)
	}
}

func TestGivenFrenchProviderFailureWhenOtherLanguagesSucceedThenResultsAreIndependent(t *testing.T) {
	// Given a failing French source and working English and Spanish sources.
	var mu sync.Mutex
	source := languageSource(func(_ context.Context, q string, l Language) ([]Entry, bool, error) {
		if l == French {
			return nil, false, errors.New("offline")
		}
		return []Entry{{Headword: q, Language: l, Meanings: []Meaning{{Definitions: []string{"definition"}}}}}, true, nil
	})
	svc := BilingualService{Sources: bilingualResolver{English: {Online: source}, Spanish: {Online: source}, French: {Online: source}}, Clock: fakeClock{now: time.Now()}}
	out := map[Language]LookupResult{}
	// When looking up a word in all languages.
	err := svc.Lookup(context.Background(), BilingualRequest{ID: "1", Query: "bonjour"}, func(o LanguageOutcome) { mu.Lock(); defer mu.Unlock(); out[o.Language] = o.Result })
	// Then English and Spanish are found and only French is unavailable.
	if err != nil || len(out) != 3 || out[English].Status != Found || out[Spanish].Status != Found || out[French].Status != SourceUnavailable {
		t.Fatalf("out=%+v err=%v", out, err)
	}
}

func TestGivenInvalidQueryWhenThreeLanguagesSearchThenNoSourceRuns(t *testing.T) {
	// Given invalid input and three sources that track calls (S3).
	var mu sync.Mutex
	calls := 0
	source := languageSource(func(context.Context, string, Language) ([]Entry, bool, error) {
		mu.Lock()
		defer mu.Unlock()
		calls++
		return nil, false, nil
	})
	svc := BilingualService{Sources: bilingualResolver{English: {Online: source}, Spanish: {Online: source}, French: {Online: source}}}
	var out []LanguageOutcome
	// When searching with a blank query.
	err := svc.Lookup(context.Background(), BilingualRequest{ID: "1", Query: "\t "}, func(o LanguageOutcome) { mu.Lock(); defer mu.Unlock(); out = append(out, o) })
	// Then every language reports an invalid query and no source runs.
	if err != nil || calls != 0 || len(out) != 3 {
		t.Fatalf("calls=%d outcomes=%+v err=%v", calls, out, err)
	}
	for _, o := range out {
		if o.Result.Status != InvalidQuery {
			t.Fatalf("outcome=%+v", o)
		}
	}
}

func TestGivenStaleFrenchEntryWhenProviderFailsThenFrenchStaleAndOtherResultsRemain(t *testing.T) {
	// Given a stale French cache entry and failed French online source (S4, S8).
	now := time.Date(2026, 11, 21, 0, 0, 0, 0, time.UTC)
	key := CacheKey{Provider: "fr", Language: French, NormalizedQuery: "bonjour"}
	cache := &bilingualCache{items: map[CacheKey]CacheRecord{key: {Key: key, FetchedAt: now.Add(-366 * 24 * time.Hour), Entries: []Entry{{Headword: "bonjour", Language: French, Meanings: []Meaning{{Definitions: []string{"salutation"}}}}}}}}
	source := languageSource(func(_ context.Context, q string, l Language) ([]Entry, bool, error) {
		if l == French {
			return nil, false, errors.New("unavailable")
		}
		return []Entry{{Headword: q, Language: l, Meanings: []Meaning{{Definitions: []string{"definition"}}}}}, true, nil
	})
	svc := BilingualService{Sources: bilingualResolver{English: {Online: source, Provider: "en"}, Spanish: {Online: source, Provider: "es"}, French: {Online: source, Provider: "fr"}}, Memory: cache, Clock: fakeClock{now: now}}
	var mu sync.Mutex
	out := map[Language]LookupResult{}
	// When looking up the word in all languages.
	err := svc.Lookup(context.Background(), BilingualRequest{ID: "1", Query: "bonjour"}, func(o LanguageOutcome) { mu.Lock(); defer mu.Unlock(); out[o.Language] = o.Result })
	// Then French serves the stale entry and the other languages are found.
	if err != nil || out[English].Status != Found || out[Spanish].Status != Found || out[French].Status != Found || !out[French].Stale {
		t.Fatalf("out=%+v err=%v", out, err)
	}
}

func TestGivenFrenchIPAWhenNormalizedThenStressAndTrainedSubstitutionsArePreserved(t *testing.T) {
	// Given French IPA inputs and an unsupported phone.
	// When normalizing them for Piper.
	// Then stress and trained substitutions are applied and the unsupported phone is rejected.
	for _, tc := range []struct{ input, want string }{
		{"bɔ̃ʒuʁ", "bɔ̃ʒˈuʁ"},
		{"ˈbɔ̃", "bˈɔ̃"},
		{"aɲo", "anjˈo"},
		{"lɥi", "lyˈi"},
		{"le.z‿a.mi", "lezami"},
	} {
		got, ok := NormalizePiperIPA(French, tc.input)
		if !ok || got != tc.want {
			t.Errorf("%q = %q, %t; want %q", tc.input, got, ok, tc.want)
		}
	}
	if got, ok := NormalizePiperIPA(French, "ǀa"); ok {
		t.Fatalf("unsupported phone accepted: %q", got)
	}
}

func TestGivenFrenchNasalVowelsWhenConvertedThenDistinctMnemonicsRemain(t *testing.T) {
	// Given the four French nasal vowels.
	// When converting them to eSpeak mnemonics.
	// Then each keeps its own distinct mnemonic.
	for _, tc := range []struct{ ipa, want string }{{"ɑ̃", "A~"}, {"ɛ̃", "E~"}, {"ɔ̃", "O~"}, {"œ̃", "W~"}} {
		got, ok := IPAToMnemonics(French, tc.ipa)
		if !ok || got != tc.want {
			t.Errorf("%q = %q %t", tc.ipa, got, ok)
		}
	}
}

func TestGivenAmbiguousFrenchBaseWhenAlternativesFoundThenAccentedVariantsRemain(t *testing.T) {
	// Given an index with an accented and an unaccented headword for one plural.
	idx := &fakeAlternativesIndex{words: []string{"côte", "cote", "côte"}}
	// When finding alternatives for the plural.
	got, err := alternativesForTest(idx).FindAlternatives(context.Background(), AlternativesRequest{Query: "cotes", Language: French})
	// Then both variants remain as base forms, without duplicates.
	if err != nil || got.Status != AlternativesReady || !reflect.DeepEqual(got.BaseForms, []string{"cote", "côte"}) {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestGivenFrenchPronunciationWhenSynthesisFailsThenFallsBackInOrder(t *testing.T) {
	// Given a French request and progressively failing synthesis backends.
	// When playing it, then each case falls back in order: Piper, IPA voice, headword voice, copy.
	req := PlayRequest{Language: French, Text: "/bɔ̃/", Headword: "bon"}
	for _, tc := range []struct {
		name       string
		piper      []byte
		voices     map[voiceCall][]byte
		want       PlaySource
		message    string
		voiceCalls int
	}{
		{"piper", []byte(wavBytes), nil, PlayPiper, "", 0},
		{"ipa", nil, phonemeResult("fr-fr", "bO~"), PlayVoice, "", 1},
		{"headword", nil, voiceResult("fr-fr", "bon"), PlayVoice, "", 2},
		{"failure", nil, nil, "", CopyNoSynthesis, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &fakePiper{output: tc.piper}
			v := &fakeVoices{available: true, outputs: tc.voices}
			got := piperService(p, v).Play(context.Background(), req)
			if got.Source != tc.want || got.Message != tc.message || len(p.calls) != 1 || len(v.calls) != tc.voiceCalls {
				t.Fatalf("result=%+v piper=%+v voices=%+v", got, p.calls, v.calls)
			}
			if tc.voiceCalls > 0 && (!v.calls[0].phonemes || v.calls[0].voice != "fr-fr" || v.calls[0].text != "bO~") {
				t.Fatalf("IPA call=%+v", v.calls[0])
			}
			if tc.voiceCalls > 1 && (v.calls[1].phonemes || v.calls[1].text != "bon") {
				t.Fatalf("headword call=%+v", v.calls[1])
			}
		})
	}
}

func TestGivenFreshFrenchCacheWhenRetriedThenOtherLanguageCacheDoesNotCollide(t *testing.T) {
	// Given fresh English and French cache records for the same query.
	now := time.Date(2026, 11, 21, 0, 0, 0, 0, time.UTC)
	cache := &bilingualCache{items: map[CacheKey]CacheRecord{}}
	for _, lang := range []Language{English, French} {
		key := CacheKey{Provider: ProviderID(lang), Language: lang, NormalizedQuery: "chat"}
		cache.items[key] = CacheRecord{Key: key, FetchedAt: now, Entries: []Entry{{Headword: "chat", Language: lang, Meanings: []Meaning{{Definitions: []string{string(lang)}}}}}}
	}
	calls := 0
	source := languageSource(func(context.Context, string, Language) ([]Entry, bool, error) { calls++; return nil, false, nil })
	svc := BilingualService{Sources: bilingualResolver{French: {Online: source, Provider: "fr"}}, Memory: cache, Clock: fakeClock{now: now}}
	var got LanguageOutcome
	// When retrying French.
	if err := svc.RetryLanguage(context.Background(), BilingualRequest{ID: "1", Query: "chat"}, French, func(o LanguageOutcome) { got = o }); err != nil {
		t.Fatal(err)
	}
	// Then the French cache record is served without calling the source.
	if calls != 0 || got.Result.Origin != OriginMemory || len(got.Result.Entries) != 1 || got.Result.Entries[0].Meanings[0].Definitions[0] != "fr" {
		t.Fatalf("calls=%d result=%+v", calls, got)
	}
}
