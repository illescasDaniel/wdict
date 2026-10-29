//go:build integration

package adapters

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/illescasDaniel/wdict/internal/core"
)

func TestGivenRealSpanishAPIResponseWhenParsedThenDefinitionsAreExtracted(t *testing.T) {
	path := os.Getenv("WDICT_TEST_SPANISH_RESPONSE")
	if path == "" {
		t.Skip("set WDICT_TEST_SPANISH_RESPONSE to a captured API response for casa")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p := SpanishOnlineDictionary{Client: &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(b))), Header: make(http.Header)}, nil
	})}}
	entries, found, err := p.Find(context.Background(), "casa", core.Spanish)
	if err != nil || !found || len(entries) != 1 || len(entries[0].Meanings) < 2 || !strings.Contains(entries[0].Meanings[0].Definitions[0], "vivienda") {
		t.Fatalf("found=%v err=%v entries=%+v", found, err, entries)
	}
}
