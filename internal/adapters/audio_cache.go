package adapters

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// audioDiskCache is the shared audio cache (Q4 of the piper-voices page: one
// cache, one policy): bytes stored on disk keyed by an opaque hash key,
// capped by a total-size budget with oldest-last-used eviction. AudioFetch
// (recordings) and PiperVoices (natural-voice WAVs) share it.
type audioDiskCache struct {
	Dir        string
	CacheBytes int64 // total cache size cap; default 50 MiB

	mu  sync.Mutex
	now func() time.Time // default time.Now; replaceable in tests
}

const defaultAudioCacheMax = 50 * 1024 * 1024

func (c *audioDiskCache) cacheMax() int64 {
	if c.CacheBytes > 0 {
		return c.CacheBytes
	}
	return defaultAudioCacheMax
}

func (c *audioDiskCache) timestamp() time.Time {
	if c.now != nil {
		return c.now()
	}
	return time.Now()
}

func (c *audioDiskCache) path(key string) string {
	hash := sha256.Sum256([]byte(key))
	return filepath.Join(c.Dir, hex.EncodeToString(hash[:]))
}

// get returns the cached bytes and refreshes the file's last-used time.
func (c *audioDiskCache) get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	path := c.path(key)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || info.Size() > maxWAVBytes {
		return nil, false
	}
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return nil, false
	}
	now := c.timestamp()
	_ = os.Chtimes(path, now, now) // last-used marker for eviction; best effort
	return data, true
}

// put writes the bytes atomically and evicts the oldest-last-used files
// until the cache directory fits the budget.
func (c *audioDiskCache) put(key string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := os.MkdirAll(c.Dir, 0700); err != nil {
		return err
	}
	temp, err := os.CreateTemp(c.Dir, ".audio-*")
	if err != nil {
		return err
	}
	name := temp.Name()
	defer os.Remove(name)
	if _, err = temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	if err = os.Rename(name, c.path(key)); err != nil {
		return err
	}
	c.evict(c.cacheMax())
	return nil
}

// evict removes the oldest-last-used cache files until the directory total
// fits the budget. It requires c.mu.
func (c *audioDiskCache) evict(limit int64) {
	entries, err := os.ReadDir(c.Dir)
	if err != nil {
		return
	}
	type entry struct {
		name  string
		size  int64
		mtime time.Time
	}
	var all []entry
	var total int64
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		all = append(all, entry{name: e.Name(), size: info.Size(), mtime: info.ModTime()})
		total += info.Size()
	}
	if total <= limit {
		return
	}
	sort.Slice(all, func(i, j int) bool { return all[i].mtime.Before(all[j].mtime) })
	for _, e := range all {
		if total <= limit {
			return
		}
		if err := os.Remove(filepath.Join(c.Dir, e.name)); err == nil {
			total -= e.size
		}
	}
}
