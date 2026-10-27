package adapters

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	stardict "github.com/ianlewis/go-stardict"
)

type archiveTransport struct{ path string }

func (a archiveTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	f, e := os.Open(a.path)
	if e != nil {
		return nil, e
	}
	return &http.Response{StatusCode: 200, Body: f, Header: make(http.Header)}, nil
}
func TestGivenRealArchiveWhenInstalledThenOfflineLookupsWork(t *testing.T) {
	// given
	archive := os.Getenv("WDICT_TEST_ARCHIVE")
	if archive == "" {
		t.Skip("set WDICT_TEST_ARCHIVE to test the publisher archive")
	}
	i := Installer{Dir: t.TempDir(), Client: &http.Client{Transport: archiveTransport{archive}}}
	// A corrupt prior installation must be replaced by a verified package.
	if err := os.MkdirAll(filepath.Join(i.Dir, dictionaryVersion), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(i.CurrentIFO(), []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	// when
	if err := i.Download(context.Background(), dictionaryID, nil); err != nil {
		t.Fatal(err)
	}
	// then
	if _, err := os.Stat(filepath.Join(i.Dir, dictionaryVersion, "en-en.ifo")); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(i.CurrentIFO())
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err = io.ReadAll(f); err != nil {
		t.Fatal(err)
	}
	offline := &OfflineDictionary{IFO: i.CurrentIFO(), Source: DictionarySourceInfo}
	for _, query := range []string{"hello", "look up"} {
		started := time.Now()
		entries, found, err := offline.Find(context.Background(), query, "en")
		if err != nil || !found || len(entries) == 0 {
			t.Fatalf("lookup %q: found=%v err=%v", query, found, err)
		}
		t.Logf("%s: %v, %d entries", query, time.Since(started), len(entries))
	}
	entries, found, err := offline.Find(context.Background(), "AI", "en")
	if err != nil || !found {
		t.Fatalf("AI: found=%v err=%v", found, err)
	}
	verb := false
	for _, entry := range entries {
		for _, meaning := range entry.Meanings {
			for _, definition := range meaning.Definitions {
				if strings.Contains(definition, "bword://") {
					t.Fatalf("link target leaked: %q", definition)
				}
				if meaning.PartOfSpeech == "Verb" && strings.Contains(definition, "use artificial intelligence") {
					verb = true
				}
			}
		}
	}
	if !verb {
		t.Fatal("AI verb definition not found")
	}
	// Compare representative exact, phrase, alias and Unicode searches with
	// the original go-stardict index rather than just asserting timing.
	reference, err := stardict.Open(i.CurrentIFO(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer reference.Close()
	for _, query := range []string{"hello", "look up", "AI", "café"} {
		old, err := reference.Search(query)
		if err != nil {
			t.Fatalf("reference %q: %v", query, err)
		}
		fast, err := scanDictionary(context.Background(), reference, i.CurrentIFO(), query)
		if err != nil {
			t.Fatalf("scan %q: %v", query, err)
		}
		counts := map[string]int{}
		for _, item := range old {
			counts[item.Title()]++
		}
		for _, item := range fast {
			counts[item.Word]--
		}
		for title, count := range counts {
			if count != 0 {
				t.Fatalf("%q: mismatch for %q: %d", query, title, count)
			}
		}
	}
}
