package adapters

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestGivenPhraseWhenLookingUpThenEscapedAndDefinitionsMapped(t *testing.T) {
	// given
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.EscapedPath() != "/api/v2/entries/en/look%20up" {
			t.Errorf("path=%s", r.URL.EscapedPath())
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`[{"word":"look up","sourceUrls":["https://en.wiktionary.org/wiki/look_up","https://en.wiktionary.org/wiki/lookup"],"license":{"name":"CC BY-SA 3.0","url":"https://creativecommons.org/licenses/by-sa/3.0"},"meanings":[{"partOfSpeech":"verb","definitions":[{"definition":"To search for information."}]}]}]`)), Header: make(http.Header)}, nil
	})}
	// when
	entries, found, err := (OnlineDictionary{Client: client}).Find(context.Background(), "look up", "en")
	// then
	if err != nil || !found || len(entries) != 1 || entries[0].Meanings[0].Definitions[0] != "To search for information." || entries[0].Source.LicenseURL != "https://creativecommons.org/licenses/by-sa/3.0" || len(entries[0].Source.AdditionalURLs) != 1 {
		t.Fatalf("entries=%+v found=%v err=%v", entries, found, err)
	}
}
func TestGivenMalformedProviderResponseWhenLookingUpThenUnavailable(t *testing.T) {
	// given
	client := &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{}`)), Header: make(http.Header)}, nil
	})}
	// when
	_, _, err := (OnlineDictionary{Client: client}).Find(context.Background(), "test", "en")
	// then
	if err == nil {
		t.Fatal("expected invalid response error")
	}
}

func TestGivenProviderStatusWhenLookingUpThenMapsResponses(t *testing.T) {
	// given
	for _, tc := range []struct {
		status             int
		wantFound, wantErr bool
	}{{201, true, false}, {404, false, false}, {429, false, true}, {503, false, true}} {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			client := &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(`[{"word":"test","meanings":[{"definitions":[{"definition":"a test"}]}]}]`)), Header: make(http.Header)}, nil
			})}
			// when
			_, found, err := (OnlineDictionary{Client: client}).Find(context.Background(), "test", "en")
			// then
			if found != tc.wantFound || (err != nil) != tc.wantErr {
				t.Fatalf("found=%v err=%v", found, err)
			}
		})
	}
}
func TestGivenUnicodeAndDelimitersWhenLookingUpThenDestinationIsFixed(t *testing.T) {
	// given
	for _, q := range []string{"a/b?c#d", "niño y café", "../test"} {
		client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Host != "api.dictionaryapi.dev" || r.URL.RawQuery != "" || r.URL.Fragment != "" || r.URL.EscapedPath() != "/api/v2/entries/en/"+url.PathEscape(q) {
				t.Errorf("unsafe URL: %s", r.URL)
			}
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
		})}
		// when
		_, _, err := (OnlineDictionary{Client: client}).Find(context.Background(), q, "en")
		// then
		if err != nil {
			t.Fatal(err)
		}
	}
}
func TestGivenOversizedResponseWhenLookingUpThenUnavailable(t *testing.T) {
	// given
	client := &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(strings.Repeat("x", 2*1024*1024+1))), Header: make(http.Header)}, nil
	})}
	// when
	_, _, err := (OnlineDictionary{Client: client}).Find(context.Background(), "word", "en")
	// then
	if err == nil {
		t.Fatal("accepted oversized response")
	}
}
