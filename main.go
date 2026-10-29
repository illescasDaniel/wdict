package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
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

type clock struct{}

func (clock) Now() time.Time { return time.Now() }

type App struct {
	dictionaries   core.Dictionaries
	mu             sync.Mutex
	progress       core.DownloadProgress
	downloadActive bool
	bilingual      core.BilingualLookup
	emit           func(core.LanguageOutcome)
	searchCancel   context.CancelFunc
	searchContext  context.Context
	searchNumber   uint64
	searchQuery    string
}

func (a *App) Lookup(query string) (string, error) {
	a.mu.Lock()
	if a.searchCancel != nil {
		a.searchCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.searchCancel = cancel
	a.searchContext = ctx
	a.searchNumber++
	a.searchQuery = query
	id := core.SearchID(fmt.Sprint(a.searchNumber))
	a.mu.Unlock()
	go func() { _ = a.bilingual.Lookup(ctx, core.BilingualRequest{ID: id, Query: query}, a.emit) }()
	return string(id), nil
}
func (a *App) RetryLanguage(id, query string, language core.Language) error {
	a.mu.Lock()
	if a.searchContext == nil || core.SearchID(fmt.Sprint(a.searchNumber)) != core.SearchID(id) || query != a.searchQuery {
		a.mu.Unlock()
		return nil
	}
	ctx := a.searchContext
	a.mu.Unlock()
	return a.bilingual.RetryLanguage(ctx, core.BilingualRequest{ID: core.SearchID(id), Query: query}, language, a.emit)
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
	appService := &App{dictionaries: &core.DictionaryService{Catalog: catalog, Installer: catalog}, bilingual: bilingual}

	root, err := fs.Sub(assets, "frontend/build")
	if err != nil {
		log.Fatal(err)
	}
	app := application.New(application.Options{Name: "wdict", Description: "Look up words and phrases", Services: []application.Service{application.NewService(appService)}, Assets: application.AssetOptions{Handler: application.BundledAssetFileServer(root)}})
	appService.emit = func(outcome core.LanguageOutcome) { app.Event.Emit("lookup:result", outcome) }
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
