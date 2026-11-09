package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/adrg/xdg"
	"github.com/illescasDaniel/wdict/internal/adapters"
	"github.com/illescasDaniel/wdict/internal/core"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/build
var assets embed.FS

// version is overridable via -ldflags "-X main.version=...".
var version = "dev"

// resolveVersion prefers the ldflags value, then the module version embedded
// by `go install ...@vX.Y.Z`, then "dev". A leading "v" is trimmed.
func resolveVersion(ldflag string, info *debug.BuildInfo, ok bool) string {
	if ldflag != "" && ldflag != "dev" {
		return ldflag
	}
	if ok && info != nil {
		if v := strings.TrimPrefix(info.Main.Version, "v"); v != "" && info.Main.Version != "(devel)" {
			return v
		}
	}
	return "dev"
}

func currentVersion() string {
	info, ok := debug.ReadBuildInfo()
	return resolveVersion(version, info, ok)
}

type clock struct{}

func (clock) Now() time.Time { return time.Now() }

type App struct {
	dictionaries     core.Dictionaries
	mu               sync.Mutex
	progress         core.DownloadProgress
	downloadActive   bool
	bilingual        core.BilingualLookup
	emit             func(core.LanguageOutcome)
	searchCancel     context.CancelFunc
	searchContext    context.Context
	searchNumber     uint64
	searchQuery      string
	attempts         map[core.Language]core.AttemptID
	attemptCancels   map[core.Language]context.CancelFunc
	attemptContexts  map[core.Language]context.Context
	alternatives     map[core.Language]core.Alternatives
	dictionaryIFO    map[core.Language]string
	emitAlternatives func(core.AlternativesResult)
	events           []lookupEvent
	dispatching      bool
}

type lookupEvent struct {
	searchID     core.SearchID
	query        string
	language     core.Language
	attempt      core.AttemptID
	outcome      *core.LanguageOutcome
	alternatives *core.AlternativesResult
}

// Version returns the application version.
func (a *App) Version() string { return currentVersion() }

func (a *App) Lookup(query string) (string, error) {
	a.mu.Lock()
	if a.searchCancel != nil {
		a.searchCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.searchCancel, a.searchContext = cancel, ctx
	a.searchNumber++
	a.searchQuery = query
	id := core.SearchID(fmt.Sprint(a.searchNumber))
	a.attempts = map[core.Language]core.AttemptID{core.English: 1, core.Spanish: 1}
	a.attemptCancels = make(map[core.Language]context.CancelFunc)
	a.attemptContexts = make(map[core.Language]context.Context)
	for _, language := range []core.Language{core.English, core.Spanish} {
		child, stop := context.WithCancel(ctx)
		a.attemptContexts[language], a.attemptCancels[language] = child, stop
	}
	a.mu.Unlock()
	go func() {
		_ = a.bilingual.Lookup(ctx, core.BilingualRequest{ID: id, Query: query}, func(outcome core.LanguageOutcome) { a.deliverLookup(outcome, 1) })
	}()
	return string(id), nil
}
func (a *App) RetryLanguage(id, query string, language core.Language) error {
	if language != core.English && language != core.Spanish {
		return fmt.Errorf("unsupported language")
	}
	a.mu.Lock()
	if a.searchContext == nil || core.SearchID(fmt.Sprint(a.searchNumber)) != core.SearchID(id) || query != a.searchQuery {
		a.mu.Unlock()
		return nil
	}
	if cancel := a.attemptCancels[language]; cancel != nil {
		cancel()
	}
	ctx, cancel := context.WithCancel(a.searchContext)
	a.attemptCancels[language], a.attemptContexts[language] = cancel, ctx
	a.attempts[language]++
	attempt := a.attempts[language]
	a.mu.Unlock()
	return a.bilingual.RetryLanguage(ctx, core.BilingualRequest{ID: core.SearchID(id), Query: query}, language, func(outcome core.LanguageOutcome) { a.deliverLookup(outcome, attempt) })
}

// enqueueLookupEvent is called with a.mu held. One worker delivers events in
// order without holding that mutex, since Wails hooks can re-enter App.
func (a *App) enqueueLookupEvent(event lookupEvent) {
	kept := a.events[:0]
	for _, pending := range a.events {
		if a.lookupEventCurrent(pending) {
			kept = append(kept, pending)
		}
	}
	a.events = append(kept, event)
	if !a.dispatching {
		a.dispatching = true
		go a.dispatchLookupEvents()
	}
}

// lookupEventCurrent requires a.mu.
func (a *App) lookupEventCurrent(event lookupEvent) bool {
	return core.SearchID(fmt.Sprint(a.searchNumber)) == event.searchID &&
		a.searchQuery == event.query && a.attempts[event.language] == event.attempt
}

func (a *App) dispatchLookupEvents() {
	for {
		a.mu.Lock()
		if len(a.events) == 0 {
			a.dispatching = false
			a.mu.Unlock()
			return
		}
		event := a.events[0]
		a.events[0] = lookupEvent{}
		a.events = a.events[1:]
		current := a.lookupEventCurrent(event)
		emit, emitAlternatives := a.emit, a.emitAlternatives
		a.mu.Unlock()
		if !current {
			continue
		}
		if event.outcome != nil && emit != nil {
			emit(*event.outcome)
		}
		if event.alternatives != nil && emitAlternatives != nil {
			emitAlternatives(*event.alternatives)
		}
	}
}

// deliverLookup associates each event with the attempt started by the caller.
// The search and attempt checks prevent earlier retries from publishing results.
func (a *App) deliverLookup(outcome core.LanguageOutcome, attempt core.AttemptID) {
	a.mu.Lock()
	if core.SearchID(fmt.Sprint(a.searchNumber)) != outcome.SearchID || a.searchQuery != outcome.Query || a.attempts[outcome.Language] != attempt {
		a.mu.Unlock()
		return
	}
	outcome.Attempt = attempt
	ctx := a.attemptContexts[outcome.Language]
	a.enqueueLookupEvent(lookupEvent{searchID: outcome.SearchID, query: outcome.Query, language: outcome.Language, attempt: attempt, outcome: &outcome})
	if outcome.Result.Status != core.NotFound || ctx == nil {
		a.mu.Unlock()
		return
	}
	finder := a.alternatives[outcome.Language]
	path := a.dictionaryIFO[outcome.Language]
	a.mu.Unlock()

	go func() {
		req := core.AlternativesRequest{SearchID: outcome.SearchID, Attempt: attempt, Query: outcome.Query, Language: outcome.Language}
		result := core.AlternativesResult{SearchID: req.SearchID, Attempt: req.Attempt, Query: req.Query, Language: req.Language, Status: core.AlternativesUnavailable}
		// Stat only the known application catalog path. If no dictionary is
		// installed, publish unavailable without opening or scanning an index.
		if finder != nil && path != "" {
			if _, err := os.Stat(path); err == nil {
				var findErr error
				result, findErr = finder.FindAlternatives(ctx, req)
				if findErr != nil {
					return
				}
			}
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		if ctx.Err() == nil && core.SearchID(fmt.Sprint(a.searchNumber)) == req.SearchID && a.searchQuery == req.Query && a.attempts[req.Language] == req.Attempt {
			a.enqueueLookupEvent(lookupEvent{searchID: req.SearchID, query: req.Query, language: req.Language, attempt: req.Attempt, alternatives: &result})
		}
	}()
}
func (a *App) Dictionaries() ([]core.Dictionary, error) {
	return a.dictionaries.List(context.Background())
}
func (a *App) Progress() core.DownloadProgress { a.mu.Lock(); defer a.mu.Unlock(); return a.progress }
func (a *App) Download(id core.DictionaryID) error {
	a.mu.Lock()
	if a.downloadActive {
		a.mu.Unlock()
		return nil
	}
	a.downloadActive = true
	a.progress = core.DownloadProgress{DictionaryID: id, State: core.DownloadInProgress}
	a.mu.Unlock()
	err := a.dictionaries.Download(context.Background(), id, func(p core.DownloadProgress) { a.mu.Lock(); a.progress = p; a.mu.Unlock() })
	a.mu.Lock()
	a.downloadActive = false
	if err != nil {
		a.progress.State = core.DownloadFailed
		a.progress.Message = err.Error()
	}
	a.mu.Unlock()
	return err
}
func diskCache(dir func() (string, error)) core.Cache {
	root, err := dir()
	if err != nil || root == "" {
		return nil
	}
	return &adapters.DiskCache{Dir: filepath.Join(root, "wdict", "responses")}
}

// initialWindowSize keeps the reference proportions while fitting the available display.
func initialWindowSize(width, height int) (int, int) {
	if width <= 0 || height <= 0 {
		return 1040, 730
	}
	const refWidth, refHeight = 1467, 1029
	scale := min(float64(width)/2560, float64(height)/1600)
	return int(float64(refWidth)*scale + 0.5), int(float64(refHeight)*scale + 0.5)
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println("wdict " + currentVersion())
		return
	}
	installer := &adapters.Installer{Dir: filepath.Join(xdg.DataHome, "wdict", "dictionaries")}
	spanish := &adapters.Installer{Dir: installer.Dir, Spanish: true}
	go func() {
		for _, i := range []*adapters.Installer{installer, spanish} {
			if err := i.Recover(); err != nil {
				log.Printf("dictionary recovery: %v", err)
			}
		}
	}()
	catalog := &adapters.Catalog{English: installer, Spanish: spanish}
	memory := &adapters.MemoryCache{}
	disk := diskCache(os.UserCacheDir)
	bilingual := &core.BilingualService{Sources: sourceResolver{english: core.LanguageSources{Offline: &adapters.OfflineDictionary{IFO: installer.CurrentIFO(), Source: adapters.DictionarySourceInfo}, Online: adapters.OnlineDictionary{}, Provider: "free-dictionary-api"}, spanish: core.LanguageSources{Offline: &adapters.OfflineDictionary{IFO: spanish.CurrentIFO(), Source: adapters.SpanishDictionarySourceInfo, Language: core.Spanish}, Online: adapters.SpanishOnlineDictionary{}, Provider: "spanish-wiktionary"}}, Memory: memory, Disk: disk, Clock: clock{}}
	appService := &App{
		dictionaries:  &core.DictionaryService{Catalog: catalog, Installer: catalog},
		bilingual:     bilingual,
		dictionaryIFO: map[core.Language]string{core.English: installer.CurrentIFO(), core.Spanish: spanish.CurrentIFO()},
		alternatives: map[core.Language]core.Alternatives{
			core.English: &core.AlternativesService{Index: &adapters.StarDictHeadwordIndex{IFO: installer.CurrentIFO(), Language: core.English}},
			core.Spanish: &core.AlternativesService{Index: &adapters.StarDictHeadwordIndex{IFO: spanish.CurrentIFO(), Language: core.Spanish}},
		},
	}

	root, err := fs.Sub(assets, "frontend/build")
	if err != nil {
		log.Fatal(err)
	}
	app := application.New(application.Options{Name: "wdict", Description: "Look up words and phrases", Services: []application.Service{application.NewService(appService)}, Assets: application.AssetOptions{Handler: application.BundledAssetFileServer(root)}})
	appService.emit = func(outcome core.LanguageOutcome) { app.Event.Emit("lookup:result", outcome) }
	appService.emitAlternatives = func(result core.AlternativesResult) { app.Event.Emit("lookup-alternatives", result) }
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{Title: "wdict", URL: "/", Width: 1040, Height: 730, MinWidth: 320, MinHeight: 400})
	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		if screen := app.Screen.GetPrimary(); screen != nil {
			width, height := initialWindowSize(screen.WorkArea.Width, screen.WorkArea.Height)
			window.SetSize(width, height)
			window.Center()
		}
	})
	if err = app.Run(); err != nil {
		log.Fatal(err)
	}
}

type sourceResolver struct{ english, spanish core.LanguageSources }

func (r sourceResolver) Resolve(language core.Language) (core.LanguageSources, error) {
	switch language {
	case core.English:
		return r.english, nil
	case core.Spanish:
		return r.spanish, nil
	default:
		return core.LanguageSources{}, fmt.Errorf("unsupported language")
	}
}
