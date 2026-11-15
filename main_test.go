package main

import (
	"context"
	"errors"
	"github.com/illescasDaniel/wdict/internal/adapters"
	"github.com/illescasDaniel/wdict/internal/core"
	"os"
	"path/filepath"
	"runtime/debug"
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
func TestGivenBundledEspeakLayoutWhenLocatedThenBinaryAndDataDirMatch(t *testing.T) {
	// (AppImage packaging, GPLv3 §5 aggregation): the bundled CLI and its
	// data are found in the layouts the build produces; each return value
	// is empty when that piece is absent.
	for _, tc := range []struct {
		name        string
		write       func(t *testing.T, dir string)
		exe         func(dir string) string
		wantBinary  func(dir string) string
		wantDataDir func(dir string) string
	}{
		{
			name: "appimage layout: executable and espeak-ng in usr/bin",
			write: func(t *testing.T, dir string) {
				bin := filepath.Join(dir, "usr", "bin")
				if err := os.MkdirAll(bin, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(bin, "espeak-ng"), []byte("#!/bin/sh\n"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(filepath.Join(dir, "usr", "share", "espeak-ng-data"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			exe:         func(dir string) string { return filepath.Join(dir, "usr", "bin", "wdict") },
			wantBinary:  func(dir string) string { return filepath.Join(dir, "usr", "bin", "espeak-ng") },
			wantDataDir: func(dir string) string { return filepath.Join(dir, "usr", "share", "espeak-ng-data") },
		},
		{
			name: "install layout: executable two directories below bin",
			write: func(t *testing.T, dir string) {
				if err := os.MkdirAll(filepath.Join(dir, "bin"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "bin", "espeak-ng"), []byte("#!/bin/sh\n"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(filepath.Join(dir, "share", "espeak-ng-data"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			exe:         func(dir string) string { return filepath.Join(dir, "lib", "wdict") },
			wantBinary:  func(dir string) string { return filepath.Join(dir, "bin", "espeak-ng") },
			wantDataDir: func(dir string) string { return filepath.Join(dir, "share", "espeak-ng-data") },
		},
		{
			name: "data directory without the binary: only the data path is reported",
			write: func(t *testing.T, dir string) {
				if err := os.MkdirAll(filepath.Join(dir, "usr", "share", "espeak-ng-data"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			exe:         func(dir string) string { return filepath.Join(dir, "usr", "bin", "wdict") },
			wantBinary:  func(string) string { return "" },
			wantDataDir: func(dir string) string { return filepath.Join(dir, "usr", "share", "espeak-ng-data") },
		},
		{
			name:        "flat directory without a bundle: nothing found",
			write:       func(*testing.T, string) {},
			exe:         func(dir string) string { return filepath.Join(dir, "wdict") },
			wantBinary:  func(string) string { return "" },
			wantDataDir: func(string) string { return "" },
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// given
			dir := t.TempDir()
			tc.write(t, dir)
			// when
			binary, dataDir := bundledEspeak(tc.exe(dir))
			// then
			if binary != tc.wantBinary(dir) {
				t.Errorf("binary = %q, want %q", binary, tc.wantBinary(dir))
			}
			if dataDir != tc.wantDataDir(dir) {
				t.Errorf("dataDir = %q, want %q", dataDir, tc.wantDataDir(dir))
			}
		})
	}
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

// blockingVoices keeps a voice download in flight so the single-download and
// progress state of the App methods can be observed.
type blockingVoices struct {
	started chan struct{}
	release chan struct{}
}

func (d *blockingVoices) List(context.Context) ([]core.VoicePack, error) { return nil, nil }
func (d *blockingVoices) Download(_ context.Context, id core.VoicePackID, report func(core.DownloadProgress)) error {
	close(d.started)
	<-d.release
	report(core.DownloadProgress{DictionaryID: core.DictionaryID(id), State: core.DownloadComplete})
	return nil
}

func TestGivenVoiceDownloadPendingWhenRequestedTwiceThenProgressIsNotFailed(t *testing.T) {
	// given: the voice-pack download state is kept separate from dictionary
	// downloads (C7 wiring): a second request while one is active neither
	// fails nor corrupts the progress shown
	v := &blockingVoices{started: make(chan struct{}), release: make(chan struct{})}
	app := &App{voices: v}
	done := make(chan error, 1)
	go func() { done <- app.DownloadVoicePack("piper-runtime") }()
	<-v.started
	if p := app.VoiceProgress(); p.State != core.DownloadInProgress {
		t.Fatalf("progress=%+v", p)
	}
	// when
	if err := app.DownloadVoicePack("piper-runtime"); err != nil {
		t.Fatal(err)
	}
	// then
	if p := app.VoiceProgress(); p.State != core.DownloadInProgress {
		t.Fatalf("progress=%+v", p)
	}
	close(v.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if p := app.VoiceProgress(); p.State != core.DownloadComplete || p.DictionaryID != core.DictionaryID("piper-runtime") {
		t.Fatalf("progress=%+v", p)
	}
}

func TestGivenBuildInfoWhenResolvingVersionThenPrefersLdflagsThenModuleVersion(t *testing.T) {
	info := func(v string) *debug.BuildInfo { return &debug.BuildInfo{Main: debug.Module{Version: v}} }
	for _, tc := range []struct {
		name, ldflag string
		info         *debug.BuildInfo
		ok           bool
		want         string
	}{
		{"ldflag set", "1.0.0", info("v2.0.0"), true, "1.0.0"},
		{"dev with module version", "dev", info("v1.2.3"), true, "1.2.3"},
		{"empty ldflag with module version", "", info("v1.2.3"), true, "1.2.3"},
		{"devel build", "dev", info("(devel)"), true, "dev"},
		{"empty module version", "dev", info(""), true, "dev"},
		{"no build info", "dev", nil, false, "dev"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveVersion(tc.ldflag, tc.info, tc.ok); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGivenVersionVariableWhenCallingAppVersionThenReturnsResolvedVersion(t *testing.T) {
	old := version
	defer func() { version = old }()
	version = "3.4.5"
	if got := (&App{}).Version(); got != "3.4.5" {
		t.Errorf("got %q, want 3.4.5", got)
	}
}
