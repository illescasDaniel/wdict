package adapters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGivenCacheBudgetWhenExceededThenTheOldestLastUsedFileIsEvicted(t *testing.T) {
	// given: a 2 KiB budget and three 1 KiB entries
	cache := &audioDiskCache{Dir: t.TempDir(), CacheBytes: 2 * 1024}
	body := []byte(strings.Repeat("x", 1024))
	if err := cache.put("one", body); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(cache.path("one"), old, old); err != nil {
		t.Fatal(err)
	}
	// when
	for _, key := range []string{"two", "three"} {
		if err := cache.put(key, body); err != nil {
			t.Fatal(err)
		}
	}
	// then: the oldest last-used entry is gone, the newer ones remain
	if _, err := os.Stat(cache.path("one")); err == nil {
		t.Fatal("oldest entry was not evicted")
	}
	for _, key := range []string{"two", "three"} {
		if _, ok := cache.get(key); !ok {
			t.Fatalf("entry %q was evicted", key)
		}
	}
	entries, err := os.ReadDir(filepath.Clean(cache.Dir))
	if err != nil || len(entries) != 2 {
		t.Fatalf("entries=%v err=%v", entries, err)
	}
}
