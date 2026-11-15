package adapters

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/illescasDaniel/wdict/internal/core"
	"golang.org/x/net/html"
)

func TestGivenMultipleLanguagesWhenSpanishParsedThenOnlySpanishDefinitionsAreReturned(t *testing.T) {
	body := `{"parse":{"title":"niño","text":"<h2 id=\"Español\">Español</h2><h4>Sustantivo</h4><dl><dt>1</dt><dd>Persona joven.<ul><li>nota</li></ul></dd></dl><h2 id=\"Italiano\">Italiano</h2><dl><dt>1</dt><dd>Otra lengua</dd></dl>"}}`
	p := SpanishOnlineDictionary{Client: &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("page") != "niño" {
			t.Fatal("query not preserved")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
	entries, found, err := p.Find(context.Background(), "niño", core.Spanish)
	if err != nil || !found || len(entries) != 1 || len(entries[0].Meanings) != 1 || entries[0].Meanings[0].Definitions[0] != "Persona joven." {
		t.Fatalf("%+v %v %v", entries, found, err)
	}
}
func TestGivenMissingPageWhenSpanishLookupThenNotFound(t *testing.T) {
	p := SpanishOnlineDictionary{Client: &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"error":{"code":"missingtitle"}}`)), Header: make(http.Header)}, nil
	})}}
	_, found, err := p.Find(context.Background(), "absent", core.Spanish)
	if found || err != nil {
		t.Fatalf("found %v error %v", found, err)
	}
}

func TestGivenUnrecognizedPageWhenSpanishLookupThenUnavailable(t *testing.T) {
	body := `{"parse":{"title":"casa","text":"<p>unrecognized page</p>"}}`
	p := SpanishOnlineDictionary{Client: &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
	_, found, err := p.Find(context.Background(), "casa", core.Spanish)
	if found || err == nil {
		t.Fatalf("found=%v err=%v", found, err)
	}
}

func TestGivenDuplicateSpanishSectionsWhenLookupThenUnavailable(t *testing.T) {
	body := `{"parse":{"title":"casa","text":"<h2 id=\"Español\">Español</h2><dl><dt>1</dt><dd>Una casa.</dd></dl><h2 id=\"Español\">Español</h2><dl><dt>1</dt><dd>Otra casa.</dd></dl>"}}`
	p := SpanishOnlineDictionary{Client: &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
	entries, found, err := p.Find(context.Background(), "casa", core.Spanish)
	if found || err == nil || len(entries) != 0 {
		t.Fatalf("entries=%+v found=%v err=%v", entries, found, err)
	}
}
func TestGivenOnlyOtherLanguageWhenSpanishLookupThenNotFound(t *testing.T) {
	body := `{"parse":{"title":"word","text":"<h2 id=\"Italiano\"><span class=\"headline-lang\" id=\"it\"></span>Italiano</h2>"}}`
	p := SpanishOnlineDictionary{Client: &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
	_, found, err := p.Find(context.Background(), "word", core.Spanish)
	if found || err != nil {
		t.Fatalf("found=%v err=%v", found, err)
	}
}

func TestGivenRedirectedSpanishPageWhenLookedUpThenCanonicalSourceIsUsed(t *testing.T) {
	body := `{"parse":{"title":"niño","text":"<h2 id=\"Español\">Español</h2><h4>Sustantivo</h4><dl><dt>1</dt><dd>Persona joven.</dd></dl>"}}`
	p := SpanishOnlineDictionary{Client: &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("redirects") != "1" {
			t.Fatal("redirects disabled")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
	entries, found, err := p.Find(context.Background(), "niña", core.Spanish)
	if err != nil || !found || entries[0].Headword != "niño" || !strings.Contains(entries[0].Source.URL, "ni%C3%B1o") {
		t.Fatalf("entries=%+v found=%v err=%v", entries, found, err)
	}
}

func TestGivenChangedSpanishHeadingWhenParsedThenSourceIsUnavailable(t *testing.T) {
	body := `{"parse":{"title":"casa","text":"<h2 id=\"es\"><span class=\"headline-lang\" id=\"es\"></span>Español</h2>"}}`
	p := SpanishOnlineDictionary{Client: &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
	_, found, err := p.Find(context.Background(), "casa", core.Spanish)
	if found || err == nil {
		t.Fatalf("found=%v err=%v", found, err)
	}
}

func TestGivenBadGatewayWhenEitherLanguageIsQueriedThenFailureIsSilent(t *testing.T) {
	client := &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusBadGateway, Body: io.NopCloser(strings.NewReader("error")), Header: make(http.Header)}, nil
	})}
	for _, tc := range []struct {
		name     string
		source   core.DictionarySource
		language core.Language
	}{{"English", OnlineDictionary{Client: client}, core.English}, {"Spanish", SpanishOnlineDictionary{Client: client}, core.Spanish}} {
		t.Run(tc.name, func(t *testing.T) {
			_, found, err := tc.source.Find(context.Background(), "diccionario", tc.language)
			if found || !errors.Is(err, core.ErrSilentUnavailable) {
				t.Fatalf("found=%v err=%v", found, err)
			}
		})
	}
}

func TestGivenStyleInsideDefinitionWhenSpanishParsedThenCSSIsNotInDefinition(t *testing.T) {
	// given
	body := `{"parse":{"title":"Car","text":"<h2 id=\"Español\">Español</h2><h4><style data-mw-deduplicate=\"x\">.h{color:red}</style>Sustantivo propio</h4><dl><dt>1</dt><dd><style data-mw-deduplicate=\"TemplateStyles:r1\">.mw-parser-output .definicion-impropia{font-style:italic}</style>Apellido.<script>var leak=1</script><noscript>nojs</noscript></dd></dl>"}}`
	p := SpanishOnlineDictionary{Client: &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
	// when
	entries, found, err := p.Find(context.Background(), "Car", core.Spanish)
	// then
	if err != nil || !found || len(entries) != 1 || len(entries[0].Meanings) != 1 {
		t.Fatalf("%+v %v %v", entries, found, err)
	}
	m := entries[0].Meanings[0]
	if m.Definitions[0] != "Apellido." || m.PartOfSpeech != "Sustantivo propio" {
		t.Fatalf("leaked markup: %q / %q", m.Definitions[0], m.PartOfSpeech)
	}
}

func TestGivenStyleInPronunciationCellWhenValuesReadThenCSSIsNotPronunciation(t *testing.T) {
	// given
	doc, err := html.Parse(strings.NewReader(`<table><tr><td><style>.x{color:red}</style>kar</td></tr></table>`))
	if err != nil {
		t.Fatal(err)
	}
	var cell *html.Node
	var find func(*html.Node)
	find = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "td" {
			cell = n
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			find(c)
		}
	}
	find(doc)
	// when
	values := spanishCellValues(cell)
	// then
	if len(values) != 1 || values[0].Text != "kar" {
		t.Fatalf("%+v", values)
	}
}
