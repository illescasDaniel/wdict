package core

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type bilingualResolver map[Language]LanguageSources

func (r bilingualResolver) Resolve(l Language) (LanguageSources, error) {
	s, ok := r[l]
	if !ok {
		return LanguageSources{}, errors.New("unsupported")
	}
	return s, nil
}

type languageSource func(context.Context, string, Language) ([]Entry, bool, error)

func (f languageSource) Find(c context.Context, q string, l Language) ([]Entry, bool, error) {
	return f(c, q, l)
}
func TestGivenIndependentLanguagesWhenOneIsSlowThenOtherArrivesFirst(t *testing.T) {
	release := make(chan struct{})
	started := make(chan Language, 2)
	outcomes := make(chan LanguageOutcome, 2)
	source := languageSource(func(ctx context.Context, q string, l Language) ([]Entry, bool, error) {
		started <- l
		if l == Spanish {
			select {
			case <-release:
			case <-ctx.Done():
				return nil, false, ctx.Err()
			}
		}
		return []Entry{{Headword: q, Language: l, Meanings: []Meaning{{Definitions: []string{"definition"}}}}}, true, nil
	})
	svc := BilingualService{Sources: bilingualResolver{English: {Online: source, Provider: "en"}, Spanish: {Online: source, Provider: "es"}}, Clock: fakeClock{now: time.Now()}}
	done := make(chan error, 1)
	go func() {
		done <- svc.Lookup(context.Background(), BilingualRequest{ID: "1", Query: "casa"}, func(o LanguageOutcome) { outcomes <- o })
	}()
	<-started
	<-started
	select {
	case o := <-outcomes:
		if o.Language != English || o.Result.Status != Found {
			t.Fatalf("unexpected: %+v", o)
		}
	case <-time.After(time.Second):
		t.Fatal("fast result blocked")
	}
	close(release)
	<-done
	if o := <-outcomes; o.Language != Spanish {
		t.Fatalf("unexpected: %+v", o)
	}
}
func TestGivenInvalidQueryWhenSearchingThenNoSourceIsCalled(t *testing.T) {
	var mu sync.Mutex
	calls := 0
	source := languageSource(func(context.Context, string, Language) ([]Entry, bool, error) {
		mu.Lock()
		calls++
		mu.Unlock()
		return nil, false, nil
	})
	svc := BilingualService{Sources: bilingualResolver{English: {Online: source}, Spanish: {Online: source}}, Clock: fakeClock{now: time.Now()}}
	var got []LanguageOutcome
	err := svc.Lookup(context.Background(), BilingualRequest{ID: "x", Query: " "}, func(o LanguageOutcome) { mu.Lock(); got = append(got, o); mu.Unlock() })
	if err != nil || calls != 0 || len(got) != 2 || got[0].Result.Status != InvalidQuery || got[1].Result.Status != InvalidQuery {
		t.Fatalf("err=%v calls=%d outcomes=%+v", err, calls, got)
	}
}

func TestGivenOfflineSpanishWhenSearchingThenEnglishStillUsesOnline(t *testing.T) {
	var mu sync.Mutex
	calls := map[Language]int{}
	online := languageSource(func(_ context.Context, q string, l Language) ([]Entry, bool, error) {
		mu.Lock()
		calls[l]++
		mu.Unlock()
		return []Entry{{Headword: q, Language: l, Meanings: []Meaning{{Definitions: []string{"online"}}}}}, true, nil
	})
	offline := languageSource(func(_ context.Context, q string, l Language) ([]Entry, bool, error) {
		return []Entry{{Headword: q, Language: l, Meanings: []Meaning{{Definitions: []string{"offline"}}}}}, true, nil
	})
	svc := BilingualService{Sources: bilingualResolver{English: {Online: online, Provider: "en"}, Spanish: {Offline: offline, Online: online, Provider: "es"}}, Clock: fakeClock{now: time.Now()}}
	got := map[Language]LookupResult{}
	err := svc.Lookup(context.Background(), BilingualRequest{ID: "1", Query: "casa"}, func(o LanguageOutcome) { mu.Lock(); got[o.Language] = o.Result; mu.Unlock() })
	if err != nil || calls[English] != 1 || calls[Spanish] != 0 || got[English].Origin != OriginOnline || got[Spanish].Origin != OriginOffline {
		t.Fatalf("calls=%v results=%+v err=%v", calls, got, err)
	}
}
func TestGivenRetryWhenLanguageSelectedThenOnlyThatLanguageIsQueried(t *testing.T) {
	calls := map[Language]int{}
	source := languageSource(func(_ context.Context, _ string, l Language) ([]Entry, bool, error) {
		calls[l]++
		return nil, false, nil
	})
	svc := BilingualService{Sources: bilingualResolver{English: {Online: source}, Spanish: {Online: source}}, Clock: fakeClock{now: time.Now()}}
	var got LanguageOutcome
	err := svc.RetryLanguage(context.Background(), BilingualRequest{ID: "abc", Query: "casa"}, Spanish, func(o LanguageOutcome) { got = o })
	if err != nil || calls[Spanish] != 1 || calls[English] != 0 || got.SearchID != "abc" || got.Result.Status != NotFound {
		t.Fatalf("calls=%v outcome=%+v err=%v", calls, got, err)
	}
}

type bilingualCache struct {
	mu    sync.Mutex
	items map[CacheKey]CacheRecord
}

func (c *bilingualCache) Get(_ context.Context, k CacheKey) (CacheRecord, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.items[k]
	return v, ok, nil
}
func (c *bilingualCache) Put(_ context.Context, r CacheRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.items == nil {
		c.items = make(map[CacheKey]CacheRecord)
	}
	c.items[r.Key] = r
	return nil
}
func TestGivenLanguageCachesWhenOnlySpanishIsStaleThenOnlySpanishRefreshes(t *testing.T) {
	now := time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC)
	cache := &bilingualCache{items: map[CacheKey]CacheRecord{}}
	for _, l := range []Language{English, Spanish} {
		key := CacheKey{Provider: ProviderID(l), Language: l, NormalizedQuery: "casa"}
		age := time.Hour
		if l == Spanish {
			age = 366 * 24 * time.Hour
		}
		cache.items[key] = CacheRecord{Key: key, FetchedAt: now.Add(-age), Entries: []Entry{{Headword: "casa", Language: l, Meanings: []Meaning{{Definitions: []string{"saved"}}}}}}
	}
	var mu sync.Mutex
	calls := map[Language]int{}
	source := languageSource(func(_ context.Context, q string, l Language) ([]Entry, bool, error) {
		mu.Lock()
		calls[l]++
		mu.Unlock()
		return []Entry{{Headword: q, Language: l, Meanings: []Meaning{{Definitions: []string{"fresh"}}}}}, true, nil
	})
	svc := BilingualService{Sources: bilingualResolver{English: {Online: source, Provider: "en"}, Spanish: {Online: source, Provider: "es"}}, Memory: cache, Clock: fakeClock{now: now}}
	outcomes := map[Language]LookupResult{}
	err := svc.Lookup(context.Background(), BilingualRequest{ID: "1", Query: "casa"}, func(o LanguageOutcome) { mu.Lock(); outcomes[o.Language] = o.Result; mu.Unlock() })
	if err != nil || calls[English] != 0 || calls[Spanish] != 1 || outcomes[English].Origin != OriginMemory || outcomes[Spanish].Origin != OriginOnline {
		t.Fatalf("calls=%v outcomes=%+v err=%v", calls, outcomes, err)
	}
}

func TestGivenSpanishStaleAndEnglishFoundWhenSpanishProviderFailsThenBothRemainAvailable(t *testing.T) {
	now := time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC)
	key := CacheKey{Provider: "es", Language: Spanish, NormalizedQuery: "casa"}
	cache := &bilingualCache{items: map[CacheKey]CacheRecord{key: {Key: key, FetchedAt: now.Add(-366 * 24 * time.Hour), Entries: []Entry{{Headword: "casa", Language: Spanish, Meanings: []Meaning{{Definitions: []string{"saved Spanish"}}}}}}}}
	online := languageSource(func(_ context.Context, q string, l Language) ([]Entry, bool, error) {
		if l == Spanish {
			return nil, false, errors.New("offline")
		}
		return []Entry{{Headword: q, Language: l, Meanings: []Meaning{{Definitions: []string{"English"}}}}}, true, nil
	})
	svc := BilingualService{Sources: bilingualResolver{English: {Online: online, Provider: "en"}, Spanish: {Online: online, Provider: "es"}}, Memory: cache, Clock: fakeClock{now: now}}
	var mu sync.Mutex
	out := map[Language]LookupResult{}
	err := svc.Lookup(context.Background(), BilingualRequest{ID: "1", Query: "casa"}, func(o LanguageOutcome) { mu.Lock(); out[o.Language] = o.Result; mu.Unlock() })
	if err != nil || out[English].Status != Found || out[Spanish].Status != Found || !out[Spanish].Stale || out[Spanish].Entries[0].Meanings[0].Definitions[0] != "saved Spanish" {
		t.Fatalf("out=%+v err=%v", out, err)
	}
}
