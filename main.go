package main

import (
	"context"
	"embed"
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
	lookup         *core.LookupService
	dictionaries   core.Dictionaries
	mu             sync.Mutex
	progress       core.DownloadProgress
	downloadActive bool
}

func (a *App) Lookup(query string) (core.LookupResult, error) {
	return a.lookup.Lookup(context.Background(), core.LookupRequest{Query: query, Language: "en"})
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
	go func() {
		if err := installer.Recover(); err != nil {
			log.Printf("dictionary recovery: %v", err)
		}
	}()
	appService := &App{dictionaries: &core.DictionaryService{Catalog: installer, Installer: installer}, lookup: &core.LookupService{Offline: &adapters.OfflineDictionary{IFO: installer.CurrentIFO(), Source: adapters.DictionarySourceInfo}, Online: adapters.OnlineDictionary{}, Memory: &adapters.MemoryCache{}, Disk: diskCache(os.UserCacheDir), Clock: clock{}, Provider: "free-dictionary-api"}}
	root, err := fs.Sub(assets, "frontend/build")
	if err != nil {
		log.Fatal(err)
	}
	app := application.New(application.Options{Name: "wdict", Description: "Look up words and phrases", Services: []application.Service{application.NewService(appService)}, Assets: application.AssetOptions{Handler: application.BundledAssetFileServer(root)}})
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
