package adapters

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofrs/flock"
	"github.com/ianlewis/go-dictzip"
	"github.com/illescasDaniel/wdict/internal/core"
	"github.com/klauspost/compress/zstd"
)

// fixtureArchive builds a real one-entry StarDict package without network access.
func fixtureArchive(t *testing.T) []byte { return fixtureArchiveLanguage(t, "en-en", "hello") }
func fixtureArchiveLanguage(t *testing.T, base, word string) []byte {
	t.Helper()
	definition := []byte("a greeting")
	var idx bytes.Buffer
	idx.WriteString(word + "\x00")
	binary.Write(&idx, binary.BigEndian, uint32(0))
	binary.Write(&idx, binary.BigEndian, uint32(len(definition)))
	var syn bytes.Buffer
	syn.WriteString("hi\x00")
	binary.Write(&syn, binary.BigEndian, uint32(0))
	var dict bytes.Buffer
	dz, err := dictzip.NewWriter(&dict)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = dz.Write(definition); err != nil {
		t.Fatal(err)
	}
	if err = dz.Close(); err != nil {
		t.Fatal(err)
	}
	ifo := fmt.Sprintf("StarDict's dict ifo file\nversion=3.0.0\nbookname=Fixture\nwordcount=1\nsynwordcount=1\nidxfilesize=%d\nsametypesequence=m\n", idx.Len())
	files := map[string][]byte{base + ".ifo": []byte(ifo), base + ".idx": idx.Bytes(), base + ".syn": syn.Bytes(), base + ".dict.dz": dict.Bytes()}
	var output bytes.Buffer
	z, err := zstd.NewWriter(&output)
	if err != nil {
		t.Fatal(err)
	}
	w := tar.NewWriter(z)
	for name, data := range files {
		if err = w.WriteHeader(&tar.Header{Name: name, Mode: 0600, Size: int64(len(data))}); err != nil {
			t.Fatal(err)
		}
		if _, err = w.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}
func fixtureInstaller(t *testing.T, dir string, data []byte) *Installer {
	t.Helper()
	sum := sha256.Sum256(data)
	return &Installer{Dir: dir, packageOverride: &packageMetadata{URL: "https://example.test/dictionary", Size: int64(len(data)), SHA: hex.EncodeToString(sum[:])}, Client: &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(data)), Header: make(http.Header)}, nil
	})}}
}
func TestGivenCuratedPackageWhenInstalledThenProgressAndOfflineLookupPersistAfterRestart(t *testing.T) {
	// given
	archive := fixtureArchive(t)
	dir := t.TempDir()
	i := fixtureInstaller(t, dir, archive)
	var states []core.DownloadState
	// when
	if err := i.Download(context.Background(), dictionaryID, func(p core.DownloadProgress) { states = append(states, p.State) }); err != nil {
		t.Fatal(err)
	}
	// then
	if len(states) < 3 || states[0] != core.DownloadInProgress || states[len(states)-2] != core.DownloadInstalling || states[len(states)-1] != core.DownloadComplete {
		t.Fatalf("progress=%v", states)
	}
	restarted := fixtureInstaller(t, dir, archive)
	listed, err := restarted.List(context.Background())
	if err != nil || !listed[0].Installed {
		t.Fatalf("list=%+v err=%v", listed, err)
	}
	source := OfflineDictionary{IFO: restarted.CurrentIFO(), Source: DictionarySourceInfo}
	entries, found, err := source.Find(context.Background(), "hello", "en")
	if err != nil || !found || len(entries) == 0 || entries[0].Meanings[0].Definitions[0] != "a greeting" {
		t.Fatalf("entries=%+v found=%v err=%v", entries, found, err)
	}
	if entries[0].Source.URL != "https://en.wiktionary.org/wiki/hello" || entries[0].Source.LicenseURL == "" {
		t.Fatal("missing entry attribution")
	}
	if !strings.Contains(entries[0].Source.Attribution, "formatting changed") {
		t.Fatal("missing change notice")
	}
	// when
	aliased, found, err := source.Find(context.Background(), "hi", "en")
	// then
	if err != nil || !found || len(aliased) == 0 || aliased[0].Headword != "hello" {
		t.Fatalf("synonym lookup: entries=%+v found=%v err=%v", aliased, found, err)
	}
}

type failingReader struct{ sent bool }

func (f *failingReader) Read(p []byte) (int, error) {
	if !f.sent {
		f.sent = true
		return copy(p, []byte("partial")), nil
	}
	return 0, errors.New("interrupted")
}
func (*failingReader) Close() error { return nil }
func TestGivenWorkingDictionaryWhenDownloadFailsThenPreservedAndRetrySucceeds(t *testing.T) {
	// given
	data := fixtureArchive(t)
	dir := t.TempDir()
	good := fixtureInstaller(t, dir, data)
	if err := good.Download(context.Background(), dictionaryID, nil); err != nil {
		t.Fatal(err)
	}
	old, err := os.ReadFile(good.CurrentIFO())
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"checksum", "interrupted"} {
		t.Run(mode, func(t *testing.T) {
			i := fixtureInstaller(t, dir, data)
			if mode == "checksum" {
				i.packageOverride.SHA = string(bytes.Repeat([]byte("0"), 64))
			} else {
				i.Client.Transport = transportFunc(func(*http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: 200, Body: &failingReader{}, Header: make(http.Header)}, nil
				})
			}
			// when
			if err := i.Download(context.Background(), dictionaryID, nil); err == nil {
				t.Fatal("expected failure")
			}
			// then
			after, e := os.ReadFile(i.CurrentIFO())
			if e != nil || !bytes.Equal(old, after) {
				t.Fatal("working dictionary changed")
			}
			matches, _ := filepath.Glob(filepath.Join(dir, ".install-*"))
			if len(matches) != 0 {
				t.Fatal("temporary download remains")
			}
			local := OfflineDictionary{IFO: i.CurrentIFO()}
			_, found, e := local.Find(context.Background(), "hello", "en")
			if e != nil || !found {
				t.Fatalf("offline lookup failed: %v", e)
			}
			if err := good.Download(context.Background(), dictionaryID, nil); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestGivenCorruptInstalledDictionaryWhenRetrySucceedsThenReplaced(t *testing.T) {
	// given
	i := fixtureInstaller(t, t.TempDir(), fixtureArchive(t))
	if err := os.MkdirAll(filepath.Dir(i.CurrentIFO()), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(i.CurrentIFO(), []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	// when
	if err := i.Download(context.Background(), dictionaryID, nil); err != nil {
		t.Fatal(err)
	}
	// then
	list, err := i.List(context.Background())
	if err != nil || !list[0].Installed {
		t.Fatalf("list=%+v err=%v", list, err)
	}
}

func TestGivenInterruptedInstallWhenRecoveringThenTemporaryDataIsRemovedAndDictionaryRemains(t *testing.T) {
	// given
	dir := t.TempDir()
	i := fixtureInstaller(t, dir, fixtureArchive(t))
	if err := i.Download(context.Background(), dictionaryID, nil); err != nil {
		t.Fatal(err)
	}
	temporary, err := os.MkdirTemp(dir, ".install-")
	if err != nil {
		t.Fatal(err)
	}
	// when
	if err = i.Recover(); err != nil {
		t.Fatal(err)
	}
	// then
	if _, err = os.Stat(temporary); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temporary directory still exists: %v", err)
	}
	source := OfflineDictionary{IFO: i.CurrentIFO()}
	_, found, err := source.Find(context.Background(), "hello", "en")
	if err != nil || !found {
		t.Fatalf("installed dictionary lost: %v", err)
	}
}
func TestGivenAnotherActiveDownloadWhenRecoveringThenItsTemporaryDataIsNotRemoved(t *testing.T) {
	// given
	dir := t.TempDir()
	temporary, err := os.MkdirTemp(dir, ".install-")
	if err != nil {
		t.Fatal(err)
	}
	lock := flock.New(filepath.Join(dir, ".download.lock"))
	locked, err := lock.TryLock()
	if err != nil || !locked {
		t.Fatalf("lock: %v", err)
	}
	defer lock.Unlock()
	i := fixtureInstaller(t, dir, fixtureArchive(t))
	// when
	if err = i.Recover(); err != nil {
		t.Fatal(err)
	}
	// then
	if _, err = os.Stat(temporary); err != nil {
		t.Fatalf("active download removed: %v", err)
	}
}

func TestGivenInterruptedReplacementWhenRecoveringThenBackupIsRestored(t *testing.T) {
	// given
	dir := t.TempDir()
	i := fixtureInstaller(t, dir, fixtureArchive(t))
	if err := i.Download(context.Background(), dictionaryID, nil); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, dictionaryVersion)
	backup := filepath.Join(dir, ".old-"+dictionaryVersion)
	if err := os.Rename(target, backup); err != nil {
		t.Fatal(err)
	}
	// when
	if err := i.Recover(); err != nil {
		t.Fatal(err)
	}
	// then
	list, err := i.List(context.Background())
	if err != nil || !list[0].Installed {
		t.Fatalf("recovery failed: %+v %v", list, err)
	}
}

func TestGivenEnglishInstalledWhenSpanishInstallFailsAndRetriesThenBothSurviveRestart(t *testing.T) {
	dir := t.TempDir()
	en := fixtureInstaller(t, dir, fixtureArchive(t))
	if err := en.Download(context.Background(), en.id(), nil); err != nil {
		t.Fatal(err)
	}
	es := fixtureInstaller(t, dir, fixtureArchiveLanguage(t, "es-es", "casa"))
	es.Spanish = true
	bad := *es
	metadata := *es.packageOverride
	metadata.SHA = strings.Repeat("0", 64)
	bad.packageOverride = &metadata
	if err := bad.Download(context.Background(), bad.id(), nil); err == nil {
		t.Fatal("expected checksum failure")
	}
	if err := en.validateInstalled(); err != nil {
		t.Fatalf("English lost: %v", err)
	}
	if err := es.Download(context.Background(), es.id(), nil); err != nil {
		t.Fatal(err)
	}
	restartedEn := &Installer{Dir: dir}
	restartedEs := &Installer{Dir: dir, Spanish: true}
	if err := restartedEn.Recover(); err != nil {
		t.Fatal(err)
	}
	if err := restartedEs.Recover(); err != nil {
		t.Fatal(err)
	}
	catalog := Catalog{English: restartedEn, Spanish: restartedEs}
	list, err := catalog.List(context.Background())
	if err != nil || len(list) != 2 || !list[0].Installed || !list[1].Installed {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	spanish := OfflineDictionary{IFO: restartedEs.CurrentIFO(), Language: core.Spanish, Source: SpanishDictionarySourceInfo}
	entries, found, err := spanish.Find(context.Background(), "casa", core.Spanish)
	if err != nil || !found || len(entries) == 0 {
		t.Fatalf("entries=%v found=%v err=%v", entries, found, err)
	}
}
