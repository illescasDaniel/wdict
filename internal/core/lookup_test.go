package core

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeSource struct {
	entries []Entry
	found   bool
	err     error
	calls   int
}

func (s *fakeSource) Find(_ context.Context, _ string, _ Language) ([]Entry, bool, error) {
	s.calls++
	return s.entries, s.found, s.err
}

type fakeCache struct {
	record CacheRecord
	found  bool
	err    error
	puts   int
}

func (c *fakeCache) Get(_ context.Context, _ CacheKey) (CacheRecord, bool, error) {
	return c.record, c.found, c.err
}
func (c *fakeCache) Put(_ context.Context, record CacheRecord) error {
	c.puts++
	if c.err != nil {
		return c.err
	}
	c.record = record
	c.found = true
	return nil
}

type fakeClock struct{ now time.Time }

func (c fakeClock) Now() time.Time { return c.now }
func TestGivenOfflineEntryWhenLookingUpThenNoOnlineRequest(t *testing.T) {
	// given
	now := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)
	local := &fakeSource{found: true, entries: []Entry{{Headword: "look up"}}}
	online := &fakeSource{}
	s := LookupService{Offline: local, Online: online, Memory: &fakeCache{}, Disk: &fakeCache{}, Clock: fakeClock{now}, Provider: "online"}
	// when
	r, err := s.Lookup(context.Background(), LookupRequest{Query: "look up", Language: "en"})
	// then
	if err != nil || r.Status != Found || r.Origin != OriginOffline || online.calls != 0 {
		t.Fatalf("result=%+v err=%v calls=%d", r, err, online.calls)
	}
}
func TestGivenExpiredCacheWhenProviderFailsThenStaleResult(t *testing.T) {
	// given
	now := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)
	cache := &fakeCache{found: true, record: CacheRecord{Entries: []Entry{{Headword: "hello"}}, FetchedAt: now.Add(-8 * 24 * time.Hour)}}
	online := &fakeSource{err: errors.New("offline")}
	s := LookupService{Offline: &fakeSource{}, Online: online, Memory: cache, Disk: &fakeCache{}, Clock: fakeClock{now}, Provider: "online"}
	// when
	r, err := s.Lookup(context.Background(), LookupRequest{Query: "hello", Language: "en"})
	// then
	if err != nil || !r.Stale || r.Status != Found {
		t.Fatalf("result=%+v err=%v", r, err)
	}
}
func TestGivenInvalidQueryWhenLookingUpThenNoSourcesAreCalled(t *testing.T) {
	// given
	source := &fakeSource{}
	s := LookupService{Offline: source, Online: source, Memory: &fakeCache{}, Disk: &fakeCache{}, Clock: fakeClock{time.Now()}}
	// when
	r, err := s.Lookup(context.Background(), LookupRequest{Query: "\n", Language: "en"})
	// then
	if err != nil || r.Status != InvalidQuery || source.calls != 0 {
		t.Fatalf("result=%+v err=%v calls=%d", r, err, source.calls)
	}
}

func TestGivenOfflineMissWhenLookingUpThenUsesOnlineProvider(t *testing.T) {
	// given
	for _, offline := range []*fakeSource{{}, {err: errors.New("unavailable")}} {
		online := &fakeSource{found: true, entries: []Entry{{Headword: "word"}}}
		s := LookupService{Offline: offline, Online: online, Clock: fakeClock{time.Now()}}
		// when
		got, err := s.Lookup(context.Background(), LookupRequest{Query: "word", Language: "en"})
		// then
		if err != nil || got.Status != Found || online.calls != 1 {
			t.Fatalf("got=%+v err=%v", got, err)
		}
	}
}
func TestGivenCacheHitWhenLookingUpThenNoNetwork(t *testing.T) {
	// given
	now := time.Now()
	record := CacheRecord{Entries: []Entry{{Headword: "word"}}, FetchedAt: now.Add(-time.Hour)}
	for _, fromDisk := range []bool{false, true} {
		memory := &fakeCache{}
		disk := &fakeCache{}
		if fromDisk {
			disk.found = true
			disk.record = record
		} else {
			memory.found = true
			memory.record = record
		}
		online := &fakeSource{}
		s := LookupService{Online: online, Memory: memory, Disk: disk, Clock: fakeClock{now}}
		// when
		r, e := s.Lookup(context.Background(), LookupRequest{Query: "word", Language: "en"})
		// then
		if e != nil || r.Status != Found || online.calls != 0 || !r.FetchedAt.Equal(record.FetchedAt) || (fromDisk && memory.puts != 1) {
			t.Fatalf("result=%+v error=%v", r, e)
		}
	}
}
func TestGivenExpiredCacheWhenProviderSucceedsThenRefreshesCaches(t *testing.T) {
	// given
	now := time.Now()
	old := CacheRecord{Entries: []Entry{{Headword: "old"}}, FetchedAt: now.Add(-8 * 24 * time.Hour)}
	mem := &fakeCache{found: true, record: old}
	disk := &fakeCache{}
	provider := &fakeSource{found: true, entries: []Entry{{Headword: "new"}}}
	s := LookupService{Online: provider, Memory: mem, Disk: disk, Clock: fakeClock{now}}
	// when
	r, e := s.Lookup(context.Background(), LookupRequest{Query: "word", Language: "en"})
	// then
	if e != nil || r.Entries[0].Headword != "new" || mem.puts != 1 || disk.puts != 1 {
		t.Fatalf("result=%+v error=%v", r, e)
	}
}
func TestGivenDiskFailureWhenProviderSucceedsThenResultAndMemoryAreKept(t *testing.T) {
	// given
	now := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)
	old := CacheRecord{Entries: []Entry{{Headword: "old"}}, FetchedAt: now.Add(-8 * 24 * time.Hour)}
	mem := &fakeCache{}
	disk := &fakeCache{record: old, found: true}
	provider := &fakeSource{found: true, entries: []Entry{{Headword: "word"}}}
	// Fail only writes, so the service can read the expired disk record first.
	failingDisk := &writeFailCache{fakeCache: disk}
	s := LookupService{Online: provider, Memory: mem, Disk: failingDisk, Clock: fakeClock{now}}
	// when
	r, e := s.Lookup(context.Background(), LookupRequest{Query: "word", Language: "en"})
	// then
	if e != nil || r.Status != Found || len(r.Entries) != 1 || r.Entries[0].Headword != "word" || mem.puts != 1 || !mem.found || mem.record.Entries[0].Headword != "word" || failingDisk.puts != 1 || disk.record.Entries[0].Headword != "old" {
		t.Fatalf("result=%+v error=%v", r, e)
	}
	// when
	r, e = s.Lookup(context.Background(), LookupRequest{Query: "word", Language: "en"})
	// then
	if e != nil || r.Status != Found || r.Entries[0].Headword != "word" || provider.calls != 1 {
		t.Fatalf("cached result=%+v error=%v provider calls=%d", r, e, provider.calls)
	}
}

type writeFailCache struct {
	*fakeCache
	puts int
}

func (c *writeFailCache) Put(_ context.Context, _ CacheRecord) error {
	c.puts++
	return errors.New("disk write failed")
}
func TestGivenFailedLookupWhenRetriedThenNotCached(t *testing.T) {
	// given
	for _, provider := range []*fakeSource{{}, {err: errors.New("network")}} {
		mem := &fakeCache{}
		disk := &fakeCache{}
		s := LookupService{Online: provider, Memory: mem, Disk: disk, Clock: fakeClock{time.Now()}}
		for i := 0; i < 2; i++ {
			// when
			r, e := s.Lookup(context.Background(), LookupRequest{Query: "word", Language: "en"})
			// then
			if e != nil || r.Status == Found {
				t.Fatalf("result=%+v error=%v", r, e)
			}
		}
		if provider.calls != 2 || mem.puts != 0 || disk.puts != 0 {
			t.Fatal("failure cached")
		}
	}
}
func TestGivenInvalidQueriesWhenLookingUpThenNoNetwork(t *testing.T) {
	// given
	for _, q := range []string{"", "  ", string(make([]byte, 257)), "a\nb", "hello\n", "\thello", string(make([]rune, 257))} {
		provider := &fakeSource{}
		s := LookupService{Online: provider, Clock: fakeClock{time.Now()}}
		// when
		r, e := s.Lookup(context.Background(), LookupRequest{Query: q, Language: "en"})
		// then
		if e != nil || r.Status != InvalidQuery || provider.calls != 0 {
			t.Fatalf("query %q result=%+v error=%v", q, r, e)
		}
	}
}
