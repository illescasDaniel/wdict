//go:build integration

package adapters

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/illescasDaniel/wdict/internal/core"
)

func TestGivenRealSpanishArchiveWhenInstalledThenDefinitionsWorkOffline(t *testing.T) {
	archive := os.Getenv("WDICT_TEST_SPANISH_ARCHIVE")
	if archive == "" {
		t.Skip("set WDICT_TEST_SPANISH_ARCHIVE to the publisher archive")
	}
	installer := &Installer{Dir: t.TempDir(), Spanish: true, Client: &http.Client{Transport: archiveTransport{archive}}}
	if err := installer.Download(context.Background(), installer.id(), nil); err != nil {
		t.Fatal(err)
	}
	d := OfflineDictionary{IFO: installer.CurrentIFO(), Source: SpanishDictionarySourceInfo, Language: core.Spanish}
	entries, found, err := d.Find(context.Background(), "casa", core.Spanish)
	if err != nil || !found || len(entries) == 0 || len(entries[0].Meanings) < 2 {
		t.Fatalf("entries=%+v found=%v err=%v", entries, found, err)
	}
	if !strings.Contains(entries[0].Meanings[0].Definitions[0], "vivienda") {
		t.Fatalf("wrong definition: %+v", entries[0].Meanings[0])
	}
}
