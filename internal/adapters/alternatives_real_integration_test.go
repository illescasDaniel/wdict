//go:build integration

package adapters

import (
	"context"
	"os"
	"runtime"
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
