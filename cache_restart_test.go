package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/illescasDaniel/wdict/internal/adapters"
	"github.com/illescasDaniel/wdict/internal/core"
)

type restartClock struct{ now time.Time }

func (c restartClock) Now() time.Time { return c.now }

type restartProvider struct {
	calls int
	fail  bool
}

func (p *restartProvider) Find(_ context.Context, query string, language core.Language) ([]core.Entry, bool, error) {
	p.calls++
	if p.fail {
		return nil, false, errors.New("provider unavailable")
	}
	return []core.Entry{{Headword: query, Language: language, Meanings: []core.Meaning{{Definitions: []string{"a definition"}}}}}, true, nil
}

func TestGivenOnlineLookupWhenAppRestartsThenDiskCacheIsReused(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("XDG_CACHE_HOME isolation is Linux-specific")
	}
	if phase := os.Getenv("WDICT_CACHE_RESTART_PHASE"); phase != "" {
		testRestartPhase(t, phase)
		return
	}

	// Each invocation is a separate process with its own memory cache.
	cacheDir := t.TempDir()
	for _, phase := range []string{"online", "fresh", "stale"} {
		cmd := exec.Command(os.Args[0], "-test.run=^TestGivenOnlineLookupWhenAppRestartsThenDiskCacheIsReused$")
		cmd.Env = append(os.Environ(), "WDICT_CACHE_RESTART_PHASE="+phase, "XDG_CACHE_HOME="+cacheDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s process failed: %v\n%s", phase, err, output)
		}
	}
}

func testRestartPhase(t *testing.T, phase string) {
	t.Helper()
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	provider := &restartProvider{fail: phase != "online"}
	if phase == "fresh" {
		now = now.Add(364 * 24 * time.Hour)
	} else if phase == "stale" {
		now = now.Add(365 * 24 * time.Hour)
	}
	cache := diskCache(os.UserCacheDir)
	if cache == nil {
		t.Fatal("disk cache unavailable")
	}
	// The same wiring and cache location are used by the desktop app.
	app := &App{lookup: &core.LookupService{Online: provider, Memory: &adapters.MemoryCache{}, Disk: cache, Clock: restartClock{now}, Provider: "free-dictionary-api"}}
	result, err := app.Lookup("word")
	if err != nil || result.Status != core.Found || len(result.Entries) != 1 || result.Entries[0].Headword != "word" {
		t.Fatalf("%s: result=%+v error=%v", phase, result, err)
	}
	if result.Entries[0].Meanings[0].Definitions[0] != "a definition" {
		t.Fatalf("%s: wrong definition: %+v", phase, result.Entries)
	}
	switch phase {
	case "online":
		if result.Origin != core.OriginOnline || provider.calls != 1 {
			t.Fatalf("initial lookup: origin=%s provider calls=%d", result.Origin, provider.calls)
		}
		files, err := filepath.Glob(filepath.Join(os.Getenv("XDG_CACHE_HOME"), "wdict", "responses", "*.json"))
		if err != nil || len(files) != 1 {
			t.Fatalf("disk cache files=%v error=%v", files, err)
		}
	case "fresh":
		if result.Origin != core.OriginDisk || result.Stale || provider.calls != 0 {
			t.Fatalf("fresh lookup: result=%+v provider calls=%d", result, provider.calls)
		}
	case "stale":
		if result.Origin != core.OriginDisk || !result.Stale || provider.calls != 1 {
			t.Fatalf("stale lookup: result=%+v provider calls=%d", result, provider.calls)
		}
	default:
		t.Fatalf("unknown test phase %q", phase)
	}
}
