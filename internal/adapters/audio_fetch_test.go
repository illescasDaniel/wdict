package adapters

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func audioServer(t *testing.T, status int, body string, requests *int) *httptest.Server {
	t.Helper()
	// TLS so that the served URLs are https, which validAudioURL requires.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests != nil {
			*requests++
		}
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(server.Close)
	return server
}

func audioFetch(t *testing.T, server *httptest.Server, cacheBytes int64) *AudioFetch {
	t.Helper()
	return &AudioFetch{Dir: t.TempDir(), Client: server.Client(), CacheBytes: cacheBytes}
}

func cacheFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func TestGivenValidRecordingURLWhenFetchedThenBytesAreCachedAndReplayedOffline(t *testing.T) {
	// given (S1, S9)
	requests := 0
	server := audioServer(t, 200, "MP3BYTES", &requests)
	fetch := audioFetch(t, server, 0)
	url := server.URL + "/a.mp3"
	// when
	data, err := fetch.Fetch(context.Background(), url)
	// then
	if err != nil || string(data) != "MP3BYTES" {
		t.Fatalf("data=%q err=%v", data, err)
	}
	if requests != 1 {
		t.Fatalf("requests=%d", requests)
	}
	if files := cacheFiles(t, fetch.Dir); len(files) != 1 {
		t.Fatalf("cache files=%v", files)
	}
	// when (S9): replay with the server gone
	server.CloseClientConnections()
	fetch.Client = &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})}
	data, err = fetch.Fetch(context.Background(), url)
	// then
	if err != nil || string(data) != "MP3BYTES" {
		t.Fatalf("offline replay: data=%q err=%v", data, err)
	}
}

func TestGivenFailedDownloadWhenFetchedThenNothingIsCached(t *testing.T) {
	// given (S9)
	server := audioServer(t, 500, "nope", nil)
	fetch := audioFetch(t, server, 0)
	// when
	if _, err := fetch.Fetch(context.Background(), server.URL+"/a.mp3"); err == nil {
		t.Fatal("failed download accepted")
	}
	// then
	if files := cacheFiles(t, fetch.Dir); len(files) != 0 {
		t.Fatalf("cache files=%v", files)
	}
}

func TestGivenUnreachableServerWhenFetchedThenAnErrorIsReturned(t *testing.T) {
	// given (S5, adapter level): a connection that is refused, not cached
	fetch := &AudioFetch{Dir: t.TempDir(), Client: &http.Client{}}
	// when
	if _, err := fetch.Fetch(context.Background(), "https://127.0.0.1:1/a.mp3"); err == nil {
		t.Fatal("refused connection accepted")
	}
	// then
	if files := cacheFiles(t, fetch.Dir); len(files) != 0 {
		t.Fatalf("cache files=%v", files)
	}
}

func TestGivenInvalidRecordingURLWhenFetchedThenNoRequestIsMade(t *testing.T) {
	// given (S10): https and a host are required; a hit here fails the test
	requests := 0
	server := audioServer(t, 200, "MP3BYTES", &requests)
	fetch := audioFetch(t, server, 0)
	for name, url := range map[string]string{
		"http not https":  "http://example.test/a.mp3",
		"no host":         "https:///a.mp3",
		"scheme relative": "//example.test/a.mp3",
		"not a url":       "not a url",
		"empty":           "",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := fetch.Fetch(context.Background(), url); err == nil {
				t.Fatal("invalid URL accepted")
			}
		})
	}
	if requests != 0 {
		t.Fatalf("requests=%d", requests)
	}
	if files := cacheFiles(t, fetch.Dir); len(files) != 0 {
		t.Fatalf("cache files=%v", files)
	}
}

func TestGivenOversizedRecordingWhenFetchedThenTheDownloadIsAbortedAndNothingIsCached(t *testing.T) {
	// given (S11)
	server := audioServer(t, 200, strings.Repeat("x", 2*1024*1024+1), nil)
	fetch := audioFetch(t, server, 0)
	// when
	if _, err := fetch.Fetch(context.Background(), server.URL+"/big.mp3"); err == nil {
		t.Fatal("oversized body accepted")
	}
	// then
	if files := cacheFiles(t, fetch.Dir); len(files) != 0 {
		t.Fatalf("cache files=%v", files)
	}
}

func TestGivenExactlyTheCapSizedRecordingWhenFetchedThenItIsCached(t *testing.T) {
	// given: the cap is inclusive
	body := strings.Repeat("x", 2*1024*1024)
	server := audioServer(t, 200, body, nil)
	fetch := audioFetch(t, server, 0)
	// when
	data, err := fetch.Fetch(context.Background(), server.URL+"/a.mp3")
	// then
	if err != nil || len(data) != len(body) {
		t.Fatalf("len=%d err=%v", len(data), err)
	}
	if files := cacheFiles(t, fetch.Dir); len(files) != 1 {
		t.Fatalf("cache files=%v", files)
	}
}

func TestGivenCacheBudgetWhenExceededThenTheOldestLastUsedFileIsEvicted(t *testing.T) {
	// given (C4): 50 MiB budget; the test shrinks it
	requests := 0
	server := audioServer(t, 200, strings.Repeat("x", 1024), &requests)
	fetch := audioFetch(t, server, 2*1024)
	first := server.URL + "/one.mp3"
	if _, err := fetch.Fetch(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	// age the first entry beyond the others
	path := filepath.Join(fetch.Dir, cacheFiles(t, fetch.Dir)[0])
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	// when
	for _, name := range []string{"/two.mp3", "/three.mp3"} {
		if _, err := fetch.Fetch(context.Background(), server.URL+name); err != nil {
			t.Fatal(err)
		}
	}
	// then: three 1 KiB files over a 3 KiB budget evict the oldest last-used
	files := cacheFiles(t, fetch.Dir)
	if len(files) != 2 {
		t.Fatalf("cache files=%v", files)
	}
	if _, err := fetch.Fetch(context.Background(), first); err != nil || requests != 4 {
		t.Fatalf("evicted entry was not refetched: err=%v requests=%d", err, requests)
	}
}

func TestGivenEmptyBodyWhenFetchedThenNothingIsCached(t *testing.T) {
	server := audioServer(t, 200, "", nil)
	fetch := audioFetch(t, server, 0)
	if _, err := fetch.Fetch(context.Background(), server.URL+"/a.mp3"); err == nil {
		t.Fatal("empty body accepted")
	}
	if files := cacheFiles(t, fetch.Dir); len(files) != 0 {
		t.Fatalf("cache files=%v", files)
	}
}
