package adapters

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// AudioFetch downloads a recording once, validates it at the edge and caches
// the bytes on disk keyed by the URL hash (C4). It implements
// core.RecordingSource.
type AudioFetch struct {
	Dir        string // cache directory, usually UserCacheDir/wdict/audio
	Client     *http.Client
	CacheBytes int64 // total cache size cap; default 50 MiB

	cacheInit sync.Once
	cache     audioDiskCache
	now       func() time.Time // default time.Now; replaceable in tests
}

const (
	maxAudioBytes     = 2 * 1024 * 1024
	audioFetchTimeout = 10 * time.Second
)

// disk returns the shared audio cache bound to this fetcher's fields.
func (f *AudioFetch) disk() *audioDiskCache {
	f.cacheInit.Do(func() { f.cache = audioDiskCache{Dir: f.Dir, CacheBytes: f.CacheBytes, now: f.now} })
	return &f.cache
}

// Fetch serves a cached recording when present, otherwise downloads the URL
// under the C4 caps (https with a host, 10 s timeout, 2 MiB body cap) and
// caches it. Failed, oversized or invalid downloads are never cached.
func (f *AudioFetch) Fetch(ctx context.Context, audioURL string) ([]byte, error) {
	if !validAudioURL(audioURL) {
		return nil, fmt.Errorf("invalid audio URL")
	}
	if data, ok := f.disk().get(audioURL); ok {
		return data, nil
	}
	data, err := f.download(ctx, audioURL)
	if err != nil {
		return nil, err
	}
	if err := f.disk().put(audioURL, data); err != nil {
		return nil, fmt.Errorf("caching audio: %w", err)
	}
	return data, nil
}

// validAudioURL is validSourceURL restricted to https: recording bytes are
// fetched, so the URL must name an https host.
func validAudioURL(raw string) bool {
	parsed, err := url.Parse(raw)
	return err == nil && validSourceURL(raw) && parsed.Scheme == "https"
}

func (f *AudioFetch) download(ctx context.Context, audioURL string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, audioFetchTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, audioURL, nil)
	if err != nil {
		return nil, fmt.Errorf("audio request: %w", err)
	}
	request.Header.Set("User-Agent", "wdict/0.4 (https://github.com/illescasDaniel/wdict)")
	client := f.Client
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("audio download: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("audio download: status %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxAudioBytes+1))
	if err != nil || len(data) > maxAudioBytes {
		return nil, errors.New("audio download: body over the size cap or unreadable")
	}
	if len(data) == 0 {
		return nil, errors.New("audio download: empty body")
	}
	return data, nil
}
