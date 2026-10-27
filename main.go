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
	"github.com/wailsapp/wails/v3/pkg/application"
	"wdict/internal/adapters"
	"wdict/internal/core"
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
	app.Window.NewWithOptions(application.WebviewWindowOptions{Title: "wdict", URL: "/", Width: 760, Height: 760, MinWidth: 320, MinHeight: 400})
	if err = app.Run(); err != nil {
		log.Fatal(err)
	}
}
