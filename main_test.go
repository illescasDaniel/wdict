package main

import (
	"context"
	"errors"
	"github.com/illescasDaniel/wdict/internal/adapters"
	"github.com/illescasDaniel/wdict/internal/core"
	"testing"
)

type startupProvider struct{}

func TestGivenDisplaySizeWhenChoosingInitialWindowThenReferenceProportionsFit(t *testing.T) {
	for _, tc := range []struct{ width, height, wantWidth, wantHeight int }{
		{2560, 1600, 1467, 1029},
		{1280, 800, 734, 515},
		{1920, 1080, 990, 695},
		{0, 0, 1040, 730},
	} {
		width, height := initialWindowSize(tc.width, tc.height)
		if width != tc.wantWidth || height != tc.wantHeight {
			t.Errorf("%dx%d: got %dx%d, want %dx%d", tc.width, tc.height, width, height, tc.wantWidth, tc.wantHeight)
		}
	}
}

func (startupProvider) Find(context.Context, string, core.Language) ([]core.Entry, bool, error) {
	return []core.Entry{{Headword: "hello"}}, true, nil
}
func TestGivenCacheDirectoryUnavailableWhenStartingThenLookupStillWorks(t *testing.T) {
	// given
	disk := diskCache(func() (string, error) { return "", errors.New("unavailable") })
	if disk != nil {
		t.Fatal("expected disk cache disabled")
	}
	lookup := core.LookupService{Online: startupProvider{}, Memory: &adapters.MemoryCache{}, Disk: disk, Clock: clock{}}
	// when
	r, err := lookup.Lookup(context.Background(), core.LookupRequest{Query: "hello", Language: "en"})
	// then
	if err != nil || r.Status != core.Found {
		t.Fatalf("result=%+v err=%v", r, err)
	}
}

type blockingDictionaries struct {
	started chan struct{}
	release chan struct{}
}

func (d *blockingDictionaries) List(context.Context) ([]core.Dictionary, error) { return nil, nil }
func (d *blockingDictionaries) Download(_ context.Context, id core.DictionaryID, report func(core.DownloadProgress)) error {
	close(d.started)
	<-d.release
	report(core.DownloadProgress{DictionaryID: id, State: core.DownloadComplete})
	return nil
}
func TestGivenDownloadPendingWhenRequestedTwiceThenProgressIsNotFailed(t *testing.T) {
	// given
	d := &blockingDictionaries{started: make(chan struct{}), release: make(chan struct{})}
	app := &App{dictionaries: d}
	done := make(chan error, 1)
	go func() { done <- app.Download("one") }()
	<-d.started
	if p := app.Progress(); p.State != core.DownloadInProgress {
		t.Fatalf("progress=%+v", p)
	}
	// when
	if err := app.Download("one"); err != nil {
		t.Fatal(err)
	}
	// then
	if p := app.Progress(); p.State != core.DownloadInProgress {
		t.Fatalf("progress=%+v", p)
	}
	close(d.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if p := app.Progress(); p.State != core.DownloadComplete {
		t.Fatalf("progress=%+v", p)
	}
}
