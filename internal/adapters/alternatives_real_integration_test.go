//go:build integration

package adapters

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/illescasDaniel/wdict/internal/core"
)

// TestGivenRealIndexesWhenSearchingConcurrentlyThenAlternativesAreBounded is an
// opt-in measurement. It does not flush OS caches or isolate other processes.
func TestGivenRealIndexesWhenSearchingConcurrentlyThenAlternativesAreBounded(t *testing.T) {
	english, spanish := os.Getenv("WDICT_TEST_EN_IFO"), os.Getenv("WDICT_TEST_ES_IFO")
	if english == "" || spanish == "" {
		t.Skip("set WDICT_TEST_EN_IFO and WDICT_TEST_ES_IFO to installed StarDict metadata paths")
	}
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	var peak uint64
	done := make(chan struct{})
	var sampled sync.WaitGroup
	sampled.Add(1)
	go func() {
		defer sampled.Done()
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				var s runtime.MemStats
				runtime.ReadMemStats(&s)
				if s.HeapAlloc > peak {
					peak = s.HeapAlloc
				}
			}
		}
	}()
	cases := []struct {
		path, query, expected string
		language              core.Language
	}{{english, "look foward to", "look forward to", core.English}, {spanish, "arbl", "árbol", core.Spanish}}
	var wg sync.WaitGroup
	results := make([]core.AlternativesResult, len(cases))
	errs := make([]error, len(cases))
	started := time.Now()
	for i := range cases {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			svc := core.AlternativesService{Index: &StarDictHeadwordIndex{IFO: cases[i].path, Language: cases[i].language}}
			results[i], errs[i] = svc.FindAlternatives(context.Background(), core.AlternativesRequest{Query: cases[i].query, Language: cases[i].language})
		}(i)
	}
	wg.Wait()
	elapsed := time.Since(started)
	close(done)
	sampled.Wait()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	if after.HeapAlloc > peak {
		peak = after.HeapAlloc
	}
	t.Logf("concurrent elapsed=%s heap baseline=%d sampled_peak=%d delta=%d bytes", elapsed, before.HeapAlloc, peak, int64(peak)-int64(before.HeapAlloc))
	for i, c := range cases {
		if errs[i] != nil {
			t.Fatalf("%s: %v", c.language, errs[i])
		}
		got := results[i]
		t.Logf("%s status=%s candidates=%q", c.language, got.Status, got.Spelling)
		switch got.Status {
		case core.AlternativesReady:
			found := false
			for _, w := range got.Spelling {
				if w == c.expected {
					found = true
				}
			}
			if !found {
				t.Errorf("%s: expected %q in %q", c.language, c.expected, got.Spelling)
			}
		case core.AlternativesTimedOut:
			if len(got.BaseForms) != 0 || len(got.Spelling) != 0 {
				t.Errorf("%s: timeout returned partial candidates", c.language)
			}
		default:
			t.Errorf("%s: unexpected status %s", c.language, got.Status)
		}
		if len(got.Spelling) > 5 {
			t.Errorf("%s: too many suggestions: %s", c.language, strings.Join(got.Spelling, ", "))
		}
	}
}

// TestGivenRealFrenchIndexWhenSearchingThenAlternativesArriveWithinDeadline is an
// opt-in measurement against the installed French index (about 545k headwords
// plus 1.7M synonyms). Set WDICT_TEST_FR_IFO to the fr-fr.ifo path, or leave it
// unset to use an installed copy under the default app data directory.
func TestGivenRealFrenchIndexWhenSearchingThenAlternativesArriveWithinDeadline(t *testing.T) {
	ifo := os.Getenv("WDICT_TEST_FR_IFO")
	if ifo == "" {
		if home, err := os.UserHomeDir(); err == nil {
			matches, _ := filepath.Glob(filepath.Join(home, ".local", "share", "wdict", "dictionaries", "fr-*", "fr-fr.ifo"))
			if len(matches) > 0 {
				sort.Strings(matches)
				ifo = matches[len(matches)-1]
			}
		}
	}
	if ifo == "" {
		t.Skip("French dictionary not installed: set WDICT_TEST_FR_IFO to the installed fr-fr.ifo path")
	}
	if _, err := os.Stat(ifo); err != nil {
		t.Skipf("French dictionary not available: %v", err)
	}
	cases := []struct {
		name, query, expected string
		base                  bool
	}{
		{"plural form", "chevaux", "cheval", true},
		{"misspelling", "ecolle", "école", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			svc := core.AlternativesService{Index: &StarDictHeadwordIndex{IFO: ifo, Language: core.French}}
			started := time.Now()
			got, err := svc.FindAlternatives(context.Background(), core.AlternativesRequest{Query: c.query, Language: core.French})
			elapsed := time.Since(started)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("query=%q elapsed=%s status=%s base_forms=%q spelling=%q", c.query, elapsed, got.Status, got.BaseForms, got.Spelling)
			if got.Status != core.AlternativesReady {
				t.Fatalf("status = %s, want %s", got.Status, core.AlternativesReady)
			}
			if elapsed > 2*time.Second {
				t.Errorf("elapsed %s exceeds the 2s deadline", elapsed)
			}
			words := got.Spelling
			if c.base {
				words = got.BaseForms
			}
			found := false
			for _, w := range words {
				if w == c.expected {
					found = true
				}
			}
			if !found {
				t.Errorf("expected %q in %q", c.expected, words)
			}
			if len(got.Spelling) > 5 {
				t.Errorf("too many suggestions: %s", strings.Join(got.Spelling, ", "))
			}
		})
	}
}
