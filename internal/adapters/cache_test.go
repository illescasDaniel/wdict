package adapters

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"wdict/internal/core"
)

func TestGivenExistingDiskEntriesWhenReplacementRenameFailsThenEntriesRemain(t *testing.T) {
	// given
	ctx := context.Background()
	dir := t.TempDir()
	now := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)
	record := func(word, definition string) core.CacheRecord {
		return core.CacheRecord{Key: core.CacheKey{Provider: "p", Language: "en", NormalizedQuery: word}, FetchedAt: now, LastUsed: now, Entries: []core.Entry{{Headword: word, Language: "en", Meanings: []core.Meaning{{Definitions: []string{definition}}}}}}
	}
	old, unrelated := record("word", "old definition"), record("other", "other definition")
	cache := &DiskCache{Dir: dir, Limit: 2}
	before := make(map[core.CacheKey][]byte)
	for _, entry := range []core.CacheRecord{old, unrelated} {
		if err := cache.Put(ctx, entry); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(dir, cacheName(entry.Key)))
		if err != nil {
			t.Fatal(err)
		}
		before[entry.Key] = data
	}
	failure := errors.New("rename failed")
	cache.rename = func(_, _ string) error { return failure }
	// when
	if err := cache.Put(ctx, record("word", "new definition")); !errors.Is(err, failure) {
		t.Fatalf("expected rename failure, got %v", err)
	}
	// then
	for _, entry := range []core.CacheRecord{old, unrelated} {
		data, err := os.ReadFile(filepath.Join(dir, cacheName(entry.Key)))
		if err != nil || !bytes.Equal(data, before[entry.Key]) {
			t.Fatalf("entry %v changed: %v", entry.Key, err)
		}
		got, found, err := (&DiskCache{Dir: dir}).Get(ctx, entry.Key)
		if err != nil || !found || !reflect.DeepEqual(got, entry) {
			t.Fatalf("entry %v: got=%+v found=%v err=%v", entry.Key, got, found, err)
		}
	}
	temps, err := filepath.Glob(filepath.Join(dir, ".cache-*"))
	if err != nil || len(temps) != 0 {
		t.Fatalf("temporary files=%v err=%v", temps, err)
	}
}

func TestGivenDiskEntryWhenReopenedThenEntrySurvives(t *testing.T) {
	// given
	dir := t.TempDir()
	now := time.Now()
	key := core.CacheKey{Provider: "p", Language: "en", NormalizedQuery: "hello"}
	record := core.CacheRecord{Key: key, FetchedAt: now, LastUsed: now, Entries: []core.Entry{{Headword: "hello", Language: "en", Meanings: []core.Meaning{{Definitions: []string{"a greeting"}}}}}}
	if err := (&DiskCache{Dir: dir}).Put(context.Background(), record); err != nil {
		t.Fatal(err)
	}
	// when
	got, found, err := (&DiskCache{Dir: dir}).Get(context.Background(), key)
	// then
	if err != nil || !found || got.Entries[0].Headword != "hello" {
		t.Fatalf("got=%+v found=%v err=%v", got, found, err)
	}
}
func TestGivenCorruptDiskEntryWhenReadThenIsIgnored(t *testing.T) {
	// given
	dir := t.TempDir()
	key := core.CacheKey{Provider: "p", Language: "en", NormalizedQuery: "hello"}
	if err := os.WriteFile(filepath.Join(dir, cacheName(key)), []byte("invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	// when
	_, found, err := (&DiskCache{Dir: dir}).Get(context.Background(), key)
	// then
	if err != nil || found {
		t.Fatalf("found=%v err=%v", found, err)
	}
}
func TestGivenMemoryCacheFullWhenWritingThenLeastRecentlyUsedIsRemoved(t *testing.T) {
	// given
	c := &MemoryCache{Limit: 1}
	ctx := context.Background()
	a := core.CacheRecord{Key: core.CacheKey{NormalizedQuery: "a"}, LastUsed: time.Now()}
	b := core.CacheRecord{Key: core.CacheKey{NormalizedQuery: "b"}, LastUsed: a.LastUsed.Add(time.Second)}
	_ = c.Put(ctx, a)
	// when
	_ = c.Put(ctx, b)
	// then
	_, found, _ := c.Get(ctx, a.Key)
	if found {
		t.Fatal("oldest entry still present")
	}
}

func TestGivenInvalidCachePayloadWhenReadThenIgnored(t *testing.T) {
	// given
	dir := t.TempDir()
	key := core.CacheKey{Provider: "p", Language: "en", NormalizedQuery: "bad"}
	c := &DiskCache{Dir: dir}
	r := core.CacheRecord{Key: key, FetchedAt: time.Now(), Entries: []core.Entry{{Headword: "bad", Language: "en"}}}
	if err := c.Put(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	// when
	_, found, err := c.Get(context.Background(), key)
	// then
	if found || err != nil {
		t.Fatalf("found=%v err=%v", found, err)
	}
}

func TestGivenDifferentProvidersAndLanguagesWhenCachedThenEntriesRemainIsolated(t *testing.T) {
	// given
	c := &DiskCache{Dir: t.TempDir(), Limit: 3}
	ctx := context.Background()
	now := time.Now()
	for _, key := range []core.CacheKey{{Provider: "p", Language: "en", NormalizedQuery: "word"}, {Provider: "q", Language: "en", NormalizedQuery: "word"}, {Provider: "p", Language: "es", NormalizedQuery: "word"}} {
		r := core.CacheRecord{Key: key, FetchedAt: now, LastUsed: now, Entries: []core.Entry{{Headword: string(key.Provider) + string(key.Language), Language: key.Language, Meanings: []core.Meaning{{Definitions: []string{"definition"}}}}}}
		// when
		if err := c.Put(ctx, r); err != nil {
			t.Fatal(err)
		}
		// then
		got, found, err := c.Get(ctx, key)
		if err != nil || !found || got.Entries[0].Headword != string(key.Provider)+string(key.Language) {
			t.Fatalf("got=%+v found=%v err=%v", got, found, err)
		}
	}
}
func TestGivenDiskCacheFullWhenWritingThenLeastRecentlyUsedIsRemoved(t *testing.T) {
	// given
	c := &DiskCache{Dir: t.TempDir(), Limit: 1}
	ctx := context.Background()
	now := time.Now()
	// when
	for j, word := range []string{"old", "new"} {
		key := core.CacheKey{Provider: "p", Language: "en", NormalizedQuery: word}
		r := core.CacheRecord{Key: key, FetchedAt: now, LastUsed: now.Add(time.Duration(j) * time.Second), Entries: []core.Entry{{Headword: word, Language: "en", Meanings: []core.Meaning{{Definitions: []string{"definition"}}}}}}
		if err := c.Put(ctx, r); err != nil {
			t.Fatal(err)
		}
	}
	// then
	_, found, err := c.Get(ctx, core.CacheKey{Provider: "p", Language: "en", NormalizedQuery: "old"})
	if err != nil || found {
		t.Fatalf("found=%v err=%v", found, err)
	}
}

func TestGivenOversizedCacheRecordWhenReadOrEvictThenIgnoredAndOtherEntriesRemain(t *testing.T) {
	// given
	dir := t.TempDir()
	c := &DiskCache{Dir: dir, Limit: 3}
	ctx := context.Background()
	bad := core.CacheKey{Provider: "p", Language: "en", NormalizedQuery: "bad"}
	f, err := os.Create(filepath.Join(dir, cacheName(bad)))
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Truncate(100 * 1024 * 1024); err != nil {
		t.Fatal(err)
	}
	f.Close()
	// when
	_, found, err := c.Get(ctx, bad)
	// then
	if err != nil || found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	now := time.Now()
	key := core.CacheKey{Provider: "p", Language: "en", NormalizedQuery: "good"}
	record := core.CacheRecord{Key: key, FetchedAt: now, LastUsed: now, Entries: []core.Entry{{Headword: "good", Language: "en", Meanings: []core.Meaning{{Definitions: []string{"definition"}}}}}}
	if err = c.Put(ctx, record); err != nil {
		t.Fatal(err)
	}
	_, found, err = c.Get(ctx, key)
	if err != nil || !found {
		t.Fatalf("valid record missing: found=%v err=%v", found, err)
	}
}

func TestGivenCorruptNewerRecordWhenEvictingThenValidRecordIsPreserved(t *testing.T) {
	// given
	dir := t.TempDir()
	c := &DiskCache{Dir: dir, Limit: 2}
	ctx := context.Background()
	now := time.Now()
	valid := func(word string, at time.Time) core.CacheRecord {
		return core.CacheRecord{Key: core.CacheKey{Provider: "p", Language: "en", NormalizedQuery: word}, FetchedAt: now, LastUsed: at, Entries: []core.Entry{{Headword: word, Language: "en", Meanings: []core.Meaning{{Definitions: []string{"definition"}}}}}}
	}
	old := valid("old", now)
	bad := valid("bad", now.Add(2*time.Hour))
	bad.Entries[0].Meanings = nil
	newer := valid("new", now.Add(time.Hour))
	// when
	for _, r := range []core.CacheRecord{old, bad, newer} {
		if err := c.Put(ctx, r); err != nil {
			t.Fatal(err)
		}
	}
	// then
	_, found, err := c.Get(ctx, old.Key)
	if err != nil || !found {
		t.Fatalf("valid record removed: found=%v err=%v", found, err)
	}
}
