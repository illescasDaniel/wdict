package adapters

import (
	"archive/tar"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/flock"
	stardict "github.com/ianlewis/go-stardict"
	"github.com/klauspost/compress/zstd"
	"wdict/internal/core"
)

const dictionaryURL = "https://github.com/xxyzz/wiktionary_stardict/releases/download/20261018/en-en.tar.zst"
const dictionarySHA = "db0a650d0b707c82c6f495e8f9b6e122f9c4e10cb5551322591e35921e8b4765"
const dictionarySize int64 = 142487509
const dictionaryVersion = "20261018"
const dictionaryID = "wiktionary-en-20261018"

var DictionarySourceInfo = core.Source{ID: dictionaryID, Name: "Wiktionary English-English", URL: "https://en.wiktionary.org/", Attribution: "Wiktionary contributors, CC BY-SA 4.0; compiled by xxyzz; definitions extracted and formatting changed by wdict", LicenseURL: "https://creativecommons.org/licenses/by-sa/4.0/", AdditionalURLs: []string{"https://xxyzz.github.io/wiktionary_stardict/"}}

type Installer struct {
	Dir             string
	Client          *http.Client
	mu              sync.Mutex
	packageOverride *packageMetadata // injected only by package tests
}

type packageMetadata struct {
	URL, SHA string
	Size     int64
}

func (i *Installer) metadata() packageMetadata {
	if i.packageOverride != nil {
		return *i.packageOverride
	}
	return packageMetadata{URL: dictionaryURL, SHA: dictionarySHA, Size: dictionarySize}
}

func (i *Installer) CurrentIFO() string { return filepath.Join(i.Dir, dictionaryVersion, "en-en.ifo") }

// Recover removes interrupted temporary installations only when no other
// process owns the shared download lock. It never touches installed data.
func (i *Installer) Recover() error {
	if _, err := os.Stat(i.Dir); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	lock := flock.New(filepath.Join(i.Dir, ".download.lock"))
	locked, err := lock.TryLock()
	if err != nil {
		return err
	}
	if !locked {
		return nil
	}
	defer lock.Unlock()
	paths, err := filepath.Glob(filepath.Join(i.Dir, ".install-*"))
	if err != nil {
		return err
	}
	for _, name := range paths {
		if err := os.RemoveAll(name); err != nil {
			return err
		}
	}
	backup := filepath.Join(i.Dir, ".old-"+dictionaryVersion)
	if _, err := os.Stat(i.CurrentIFO()); errors.Is(err, os.ErrNotExist) {
		if _, err := os.Stat(backup); err == nil {
			if err := os.Rename(backup, filepath.Join(i.Dir, dictionaryVersion)); err != nil {
				return err
			}
		}
	}
	return nil
}
func (i *Installer) List(_ context.Context) ([]core.Dictionary, error) {
	err := validateInstalled(i.Dir)
	return []core.Dictionary{{ID: dictionaryID, Name: "Wiktionary English-English", Language: "en", Source: DictionarySourceInfo, Version: dictionaryVersion, License: "CC BY-SA 4.0", DownloadBytes: dictionarySize, Installed: err == nil}}, nil
}

func validateInstalled(dir string) error {
	base := filepath.Join(dir, dictionaryVersion)
	for _, name := range []string{"en-en.ifo", "en-en.idx", "en-en.dict.dz", "en-en.syn"} {
		info, err := os.Stat(filepath.Join(base, name))
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() || info.Size() == 0 {
			return errors.New("invalid dictionary file")
		}
	}
	d, err := stardict.Open(filepath.Join(base, "en-en.ifo"), nil)
	if err != nil {
		return err
	}
	return d.Close()
}
func (i *Installer) Download(ctx context.Context, id core.DictionaryID, progress func(core.DownloadProgress)) error {
	if id != dictionaryID {
		return errors.New("unknown dictionary")
	}
	if !i.mu.TryLock() {
		return errors.New("download already in progress")
	}
	defer i.mu.Unlock()
	packageInfo := i.metadata()
	report := func(s core.DownloadState, n int64) {
		if progress != nil {
			progress(core.DownloadProgress{DictionaryID: id, State: s, BytesRead: n, TotalBytes: packageInfo.Size})
		}
	}
	if err := os.MkdirAll(i.Dir, 0700); err != nil {
		return err
	}
	processLock := flock.New(filepath.Join(i.Dir, ".download.lock"))
	locked, err := processLock.TryLock()
	if err != nil {
		return err
	}
	if !locked {
		return errors.New("download already in progress")
	}
	defer processLock.Unlock()
	temp, err := os.MkdirTemp(i.Dir, ".install-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, packageInfo.URL, nil)
	if err != nil {
		return err
	}
	client := i.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Minute}
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}
	file, err := os.Create(filepath.Join(temp, "archive"))
	if err != nil {
		return err
	}
	h := sha256.New()
	report(core.DownloadInProgress, 0)
	buf := make([]byte, 128*1024)
	var n int64
	for {
		nr, er := resp.Body.Read(buf)
		if nr > 0 {
			n += int64(nr)
			if n > packageInfo.Size {
				file.Close()
				return errors.New("archive exceeds expected size")
			}
			if _, err = io.MultiWriter(file, h).Write(buf[:nr]); err != nil {
				file.Close()
				return err
			}
			report(core.DownloadInProgress, n)
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			file.Close()
			return er
		}
	}
	if err = file.Close(); err != nil {
		return err
	}
	if n != packageInfo.Size || hex.EncodeToString(h.Sum(nil)) != packageInfo.SHA {
		return errors.New("dictionary checksum verification failed")
	}
	report(core.DownloadInstalling, n)
	if err = extractDictionary(filepath.Join(temp, "archive"), temp); err != nil {
		return err
	}
	d, err := stardict.Open(filepath.Join(temp, "en-en.ifo"), nil)
	if err != nil {
		return fmt.Errorf("invalid dictionary: %w", err)
	}
	entries, err := d.Search("hello")
	if err != nil || len(entries) == 0 || strings.TrimSpace(entries[0].Data().String()) == "" {
		_ = d.Close()
		return errors.New("dictionary contents could not be validated")
	}
	if err = d.Close(); err != nil {
		return err
	}
	if err = os.Remove(filepath.Join(temp, "archive")); err != nil {
		return err
	}
	// Immutable version directories make the initial installation atomic. Existing
	// valid copies are retained instead of being replaced during a retry.
	target := filepath.Join(i.Dir, dictionaryVersion)
	if validateInstalled(i.Dir) == nil {
		report(core.DownloadComplete, n)
		return nil
	}
	backup := filepath.Join(i.Dir, ".old-"+dictionaryVersion)
	if _, statErr := os.Stat(target); statErr == nil {
		if _, err = os.Stat(backup); err == nil {
			return errors.New("dictionary recovery directory already exists")
		}
		if err = os.Rename(target, backup); err != nil {
			return err
		}
		defer func() {
			if _, e := os.Stat(target); os.IsNotExist(e) {
				_ = os.Rename(backup, target)
			}
		}()
	}
	if err = os.Rename(temp, target); err != nil {
		return err
	}
	_ = os.RemoveAll(backup)
	report(core.DownloadComplete, n)
	return nil
}
func extractDictionary(archive, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	z, err := zstd.NewReader(f, zstd.WithDecoderMaxMemory(512<<20))
	if err != nil {
		return err
	}
	defer z.Close()
	t := tar.NewReader(z)
	wanted := map[string]bool{"en-en.ifo": false, "en-en.idx": false, "en-en.dict.dz": false, "en-en.syn": false}
	var total int64
	for {
		h, e := t.Next()
		if e == io.EOF {
			break
		}
		if e != nil {
			return e
		}
		name := strings.TrimPrefix(h.Name, "./")
		clean := path.Clean(name)
		if strings.HasPrefix(name, "/") || clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(name, "\\") {
			return errors.New("unsafe archive path")
		}
		if h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeDir {
			return errors.New("unsupported archive entry")
		}
		if h.Size < 0 {
			return errors.New("invalid archive size")
		}
		total += h.Size
		if total > 1024<<20 {
			return errors.New("archive exceeds extraction limit")
		}
		if _, ok := wanted[clean]; !ok {
			continue
		}
		if wanted[clean] {
			return errors.New("duplicate dictionary file")
		}
		wanted[clean] = true
		out, e := os.OpenFile(filepath.Join(dest, clean), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			return e
		}
		_, e = io.Copy(out, t)
		closeErr := out.Close()
		if e != nil {
			return e
		}
		if closeErr != nil {
			return closeErr
		}
	}
	for _, ok := range wanted {
		if !ok {
			return errors.New("dictionary archive missing required file")
		}
	}
	return nil
}

func (i *Installer) Find(ctx context.Context, id core.DictionaryID) (core.Dictionary, bool, error) {
	list, err := i.List(ctx)
	if err != nil {
		return core.Dictionary{}, false, err
	}
	for _, d := range list {
		if d.ID == id {
			return d, true, nil
		}
	}
	return core.Dictionary{}, false, nil
}
func (i *Installer) Install(ctx context.Context, id core.DictionaryID, progress func(core.DownloadProgress)) error {
	return i.Download(ctx, id, progress)
}
