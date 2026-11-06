package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/illescasDaniel/wdict/internal/core"
)

type alternativeBilingualFake struct {
	initial chan func(core.LanguageOutcome)
	retry   chan func(core.LanguageOutcome)
}

func (f *alternativeBilingualFake) Lookup(_ context.Context, _ core.BilingualRequest, emit func(core.LanguageOutcome)) error {
	f.initial <- emit
	return nil
}
func (f *alternativeBilingualFake) RetryLanguage(_ context.Context, _ core.BilingualRequest, _ core.Language, emit func(core.LanguageOutcome)) error {
	f.retry <- emit
	return nil
}

type alternativeFinderFake struct {
	requests chan core.AlternativesRequest
	release  chan struct{}
}

func (f *alternativeFinderFake) FindAlternatives(ctx context.Context, req core.AlternativesRequest) (core.AlternativesResult, error) {
	f.requests <- req
	select {
	case <-ctx.Done():
		return core.AlternativesResult{}, ctx.Err()
	case <-f.release:
		return core.AlternativesResult{SearchID: req.SearchID, Query: req.Query, Language: req.Language, Attempt: req.Attempt, Status: core.AlternativesReady, Spelling: []string{"hello"}}, nil
	}
}
func TestGivenLanguageMissWhenDictionaryMissingThenUnavailableWithoutScan(t *testing.T) {
	// Given a search without the requested offline dictionary.
	fake := &alternativeBilingualFake{initial: make(chan func(core.LanguageOutcome), 1), retry: make(chan func(core.LanguageOutcome), 1)}
	results := make(chan core.LanguageOutcome, 2)
	alternatives := make(chan core.AlternativesResult, 2)
	finder := &alternativeFinderFake{requests: make(chan core.AlternativesRequest, 2), release: make(chan struct{})}
	app := &App{bilingual: fake, emit: func(o core.LanguageOutcome) { results <- o }, emitAlternatives: func(r core.AlternativesResult) { alternatives <- r }, alternatives: map[core.Language]core.Alternatives{core.English: finder}, dictionaryIFO: map[core.Language]string{core.English: filepath.Join(t.TempDir(), "missing.ifo")}}
	id, err := app.Lookup("helo")
	if err != nil {
		t.Fatal(err)
	}
	emit := <-fake.initial
	// When the English lookup reports a confirmed miss.
	emit(core.LanguageOutcome{SearchID: core.SearchID(id), Query: "helo", Language: core.English, Result: core.LookupResult{Status: core.NotFound}})
	// Then the existing result is retained and alternatives are unavailable without scanning.
	if out := <-results; out.Attempt != 1 {
		t.Fatalf("attempt: %v", out.Attempt)
	}
	select {
	case got := <-alternatives:
		if got.Status != core.AlternativesUnavailable || got.Attempt != 1 {
			t.Fatalf("result: %+v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("missing unavailable event")
	}
	select {
	case req := <-finder.requests:
		t.Fatalf("unexpected scan: %+v", req)
	default:
	}
}
func TestGivenLanguageRetryWhenOldResultsArriveThenOnlyLatestAttemptDelivered(t *testing.T) {
	// Given a pending initial miss and an installed dictionary.
	fake := &alternativeBilingualFake{initial: make(chan func(core.LanguageOutcome), 1), retry: make(chan func(core.LanguageOutcome), 1)}
	results := make(chan core.LanguageOutcome, 4)
	alternatives := make(chan core.AlternativesResult, 4)
	finder := &alternativeFinderFake{requests: make(chan core.AlternativesRequest, 4), release: make(chan struct{})}
	ifo := filepath.Join(t.TempDir(), "index.ifo")
	if err := os.WriteFile(ifo, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	app := &App{bilingual: fake, emit: func(o core.LanguageOutcome) { results <- o }, emitAlternatives: func(r core.AlternativesResult) { alternatives <- r }, alternatives: map[core.Language]core.Alternatives{core.English: finder}, dictionaryIFO: map[core.Language]string{core.English: ifo}}
	id, _ := app.Lookup("helo")
	initial := <-fake.initial
	initial(core.LanguageOutcome{SearchID: core.SearchID(id), Query: "helo", Language: core.English, Result: core.LookupResult{Status: core.NotFound}})
	<-results
	<-finder.requests
	// When retry starts, the earlier search is cancelled and its late outcome ignored.
	if err := app.RetryLanguage(id, "helo", core.English); err != nil {
		t.Fatal(err)
	}
	retry := <-fake.retry
	initial(core.LanguageOutcome{SearchID: core.SearchID(id), Query: "helo", Language: core.English, Result: core.LookupResult{Status: core.NotFound}})
	retry(core.LanguageOutcome{SearchID: core.SearchID(id), Query: "helo", Language: core.English, Result: core.LookupResult{Status: core.NotFound}})
	// Then the new attempt alone produces a result and alternatives.
	if out := <-results; out.Attempt != 2 {
		t.Fatalf("attempt: %v", out.Attempt)
	}
	if req := <-finder.requests; req.Attempt != 2 {
		t.Fatalf("request: %+v", req)
	}
	close(finder.release)
	select {
	case got := <-alternatives:
		if got.Attempt != 2 {
			t.Fatalf("stale: %+v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("missing alternatives")
	}
	select {
	case out := <-results:
		t.Fatalf("unexpected stale outcome: %+v", out)
	default:
	}
}

func TestGivenSynchronousEventHookWhenItReentersAppThenDeliveryDoesNotDeadlock(t *testing.T) {
	// Given an event callback that synchronously calls another App method.
	fake := &alternativeBilingualFake{initial: make(chan func(core.LanguageOutcome), 1), retry: make(chan func(core.LanguageOutcome), 1)}
	delivered := make(chan string, 2)
	app := &App{bilingual: fake}
	app.emit = func(core.LanguageOutcome) { _ = app.Progress(); delivered <- "lookup" }
	app.emitAlternatives = func(core.AlternativesResult) { _ = app.Progress(); delivered <- "alternatives" }
	id, err := app.Lookup("helo")
	if err != nil {
		t.Fatal(err)
	}
	emit := <-fake.initial
	// When a miss triggers the ordinary and missing-dictionary events.
	emit(core.LanguageOutcome{SearchID: core.SearchID(id), Query: "helo", Language: core.English, Result: core.LookupResult{Status: core.NotFound}})
	// Then both callbacks finish in order without waiting for the App mutex.
	for _, want := range []string{"lookup", "alternatives"} {
		select {
		case got := <-delivered:
			if got != want {
				t.Fatalf("got event %q, want %q", got, want)
			}
		case <-time.After(time.Second):
			t.Fatal("event callback deadlocked while re-entering App")
		}
	}
}

func TestGivenUnavailableLookupWhenDeliveredThenSuggestionFinderIsNotCalled(t *testing.T) {
	for _, status := range []core.LookupStatus{core.SourceUnavailable, core.SilentUnavailable} {
		t.Run(string(status), func(t *testing.T) {
			// Given an installed dictionary and a finder that records any call.
			fake := &alternativeBilingualFake{initial: make(chan func(core.LanguageOutcome), 1), retry: make(chan func(core.LanguageOutcome), 1)}
			results := make(chan core.LanguageOutcome, 1)
			alternatives := make(chan core.AlternativesResult, 1)
			finder := &alternativeFinderFake{requests: make(chan core.AlternativesRequest, 1), release: make(chan struct{})}
			ifo := filepath.Join(t.TempDir(), "index.ifo")
			if err := os.WriteFile(ifo, []byte("fixture"), 0600); err != nil {
				t.Fatal(err)
			}
			app := &App{bilingual: fake, emit: func(o core.LanguageOutcome) { results <- o }, emitAlternatives: func(r core.AlternativesResult) { alternatives <- r }, alternatives: map[core.Language]core.Alternatives{core.English: finder}, dictionaryIFO: map[core.Language]string{core.English: ifo}}
			id, err := app.Lookup("helo")
			if err != nil {
				t.Fatal(err)
			}
			emit := <-fake.initial
			// When the language lookup is unavailable rather than a confirmed miss.
			emit(core.LanguageOutcome{SearchID: core.SearchID(id), Query: "helo", Language: core.English, Result: core.LookupResult{Status: status}})
			// Then the original outcome is delivered, with no suggestion request or event.
			if out := <-results; out.Result.Status != status {
				t.Fatalf("status: %v", out.Result.Status)
			}
			select {
			case req := <-finder.requests:
				t.Fatalf("unexpected scan: %+v", req)
			default:
			}
			select {
			case got := <-alternatives:
				t.Fatalf("unexpected alternatives: %+v", got)
			default:
			}
		})
	}
}
