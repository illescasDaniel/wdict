package adapters

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"wdict/internal/core"
)

type MemoryCache struct {
	mu      sync.Mutex
	entries map[core.CacheKey]core.CacheRecord
	Limit   int
}

func (c *MemoryCache) Get(_ context.Context, k core.CacheKey) (core.CacheRecord, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.entries[k]
	return v, ok, nil
}
func (c *MemoryCache) Put(_ context.Context, r core.CacheRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.entries == nil {
		c.entries = make(map[core.CacheKey]core.CacheRecord)
	}
	limit := c.Limit
	if limit <= 0 {
		limit = 128
	}
	c.entries[r.Key] = r
	for len(c.entries) > limit {
		var oldest core.CacheKey
		var value core.CacheRecord
		first := true
		for k, v := range c.entries {
			if first || v.LastUsed.Before(value.LastUsed) {
				oldest = k
				value = v
				first = false
			}
		}
		delete(c.entries, oldest)
	}
	return nil
}

type DiskCache struct {
	Dir    string
	Limit  int
	mu     sync.Mutex
	rename func(string, string) error
}
type diskRecord struct {
	Version int
	Record  core.CacheRecord
}

func cacheName(k core.CacheKey) string {
	b, _ := json.Marshal(k)
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:]) + ".json"
}

const maxCacheRecordBytes = 2 * 1024 * 1024

var errCacheRecordTooLarge = errors.New("cache record too large")

func readCacheFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, maxCacheRecordBytes+1))
	if err != nil {
		return nil, err
	}
	if len(b) > maxCacheRecordBytes {
		return nil, errCacheRecordTooLarge
	}
	return b, nil
}
func (c *DiskCache) Get(_ context.Context, k core.CacheKey) (core.CacheRecord, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var zero core.CacheRecord
	b, err := readCacheFile(filepath.Join(c.Dir, cacheName(k)))
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, errCacheRecordTooLarge) {
		return zero, false, nil
	}
	if err != nil {
		return zero, false, err
	}
	var item diskRecord
	if len(b) > 2*1024*1024 || json.Unmarshal(b, &item) != nil || item.Version != 1 || item.Record.Key != k || !validCacheEntries(item.Record.Entries, k.Language) || item.Record.FetchedAt.IsZero() {
		return zero, false, nil
	}
	return item.Record, true, nil
}

func validCacheEntries(entries []core.Entry, language core.Language) bool {
	if len(entries) == 0 {
		return false
	}
	for _, entry := range entries {
		if strings.TrimSpace(entry.Headword) == "" || entry.Language != language || len(entry.Meanings) == 0 {
			return false
		}
		for _, meaning := range entry.Meanings {
			if len(meaning.Definitions) == 0 {
				return false
			}
			for _, definition := range meaning.Definitions {
				if strings.TrimSpace(definition) == "" {
					return false
				}
			}
		}
	}
	return true
}
func (c *DiskCache) Put(_ context.Context, r core.CacheRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := os.MkdirAll(c.Dir, 0700); err != nil {
		return err
	}
	b, err := json.Marshal(diskRecord{Version: 1, Record: r})
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(c.Dir, ".cache-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	rename := c.rename
	if rename == nil {
		rename = os.Rename
	}
	if err = rename(name, filepath.Join(c.Dir, cacheName(r.Key))); err != nil {
		return err
	}
	limit := c.Limit
	if limit <= 0 {
		limit = 1000
	}
	files, _ := filepath.Glob(filepath.Join(c.Dir, "*.json"))
	if len(files) <= limit {
		return nil
	}
	type candidate struct {
		path string
		rec  core.CacheRecord
	}
	var candidates []candidate
	for _, path := range files {
		data, e := readCacheFile(path)
		if e != nil {
			continue
		}
		var item diskRecord
		if json.Unmarshal(data, &item) == nil && item.Version == 1 && filepath.Base(path) == cacheName(item.Record.Key) && !item.Record.FetchedAt.IsZero() && !item.Record.LastUsed.IsZero() && validCacheEntries(item.Record.Entries, item.Record.Key.Language) {
			candidates = append(candidates, candidate{path, item.Record})
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].rec.LastUsed.Before(candidates[j].rec.LastUsed) })
	for len(candidates) > limit {
		_ = os.Remove(candidates[0].path)
		candidates = candidates[1:]
	}
	return nil
}
