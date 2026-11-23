package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/illescasDaniel/wdict/internal/core"
	"golang.org/x/net/html"
)

type frenchRoundTrip func(*http.Request) (*http.Response, error)

func (f frenchRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestGivenFrenchAndOtherSectionsWhenParsedThenOnlyFrenchDefinitionsAppear(t *testing.T) {
	const page = `<h2 id="Français"><span class="sectionlangue" id="fr">Français</span></h2><h3><span class="titredef">Nom commun</span></h3><p><span class="API">\ʃa\</span></p><ol><li>Un félin<ul><li>example ignored</li></ul><script>unsafe</script></li></ol><h2 id="Anglais"><span class="sectionlangue" id="en">Anglais</span></h2><h3><span class="titredef">Noun</span></h3><ol><li>wrong language</li></ol>`
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	meanings, pron, sections, _ := frenchContent(doc, false)
	if sections != 1 || len(meanings) != 1 || meanings[0].Definitions[0] != "Un félin" || len(pron) != 1 || pron[0].Text != "/ʃa/" {
		t.Fatalf("sections=%d meanings=%+v pronunciation=%+v", sections, meanings, pron)
	}
}

func TestGivenFrenchProviderMissingTitleWhenLookedUpThenNotFound(t *testing.T) {
	client := &http.Client{Transport: frenchRoundTrip(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "fr.wiktionary.org" {
			t.Fatalf("unexpected host: %s", r.URL.Host)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"error":{"code":"missingtitle"}}`)), Header: make(http.Header)}, nil
	})}
	_, found, err := (FrenchOnlineDictionary{Client: client}).Find(context.Background(), "missing", core.French)
	if found || err != nil {
		t.Fatalf("found=%v err=%v", found, err)
	}
}

func TestGivenFrenchOfflineDefinitionsWhenParsedThenExamplesAreExcluded(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(`<section><h4><span class="titredef">Interjection</span></h4><p><span class="API">\bɔ̃.ʒuʁ\</span></p><ol><li>Salutation<ul><li>example</li></ul></li></ol><h4>Notes</h4><ol><li>not a definition</li></ol></section>`))
	if err != nil {
		t.Fatal(err)
	}
	meanings, pron, _, _ := frenchContent(doc, true)
	if len(meanings) != 1 || meanings[0].Definitions[0] != "Salutation" || len(pron) != 1 {
		t.Fatalf("meanings=%+v pron=%+v", meanings, pron)
	}
}

func TestGivenDedicatedFrenchPronunciationWhenParsedThenOnlyHeadwordIPAIsCollected(t *testing.T) {
	const page = `<h2 id="Français"><span class="sectionlangue" id="fr">Français</span></h2>
 <h3><span class="titredef">Nom commun</span></h3><ol><li>Salutation</li></ol>
 <h3 id="Prononciation"><span class="titrepron">Prononciation</span></h3>
 <ul><li>Rime avec <a title="Annexe:Rimes"><span class="API">\uʁ\</span></a>.</li>
 <li><a title="Annexe:Prononciation/français"><span class="API">\bɔ̃.ʒuʁ\</span></a></li>
 <li><span class="audio-ipa"><span class="API">[bɔ̃.ʒuʁ]</span></span></li></ul>
 <h3><span class="titresyno">Synonymes</span></h3><ul><li><a title="Annexe:Prononciation/français"><span class="API">\salut\</span></a></li></ul>
 <h2 id="Anglais"><span class="sectionlangue">Anglais</span></h2><h3><span class="titrepron">Prononciation</span></h3><ul><li><a title="Annexe:Prononciation/français"><span class="API">\wrong\</span></a></li></ul>`
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	meanings, pron, sections, _ := frenchContent(doc, false)
	if sections != 1 || len(meanings) != 1 || len(pron) != 2 || pron[0].Text != "/bɔ̃.ʒuʁ/" || pron[1].Text != "/bɔ̃.ʒuʁ/" {
		t.Fatalf("sections=%d meanings=%+v pronunciations=%+v", sections, meanings, pron)
	}
}

func TestGivenFrenchProviderMarkupWhenParsedThenOnlyPlainFrenchDefinitionsAreReturned(t *testing.T) {
	page := `<h2 id="Français"><span class="sectionlangue" id="fr">Français</span></h2><h3><span class="titredef">Nom commun</span></h3><ol><li>Un <a href="javascript:alert(1)">chat</a><script>bad script</script><ul><li>example only</li></ul></li></ol>`
	response, _ := json.Marshal(map[string]any{"parse": map[string]string{"title": "chat", "text": page}})
	client := &http.Client{Transport: frenchRoundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(response))), Header: make(http.Header)}, nil
	})}
	entries, found, err := (FrenchOnlineDictionary{Client: client}).Find(context.Background(), "chat", core.French)
	if err != nil || !found || len(entries) != 1 || len(entries[0].Meanings) != 1 || entries[0].Meanings[0].Definitions[0] != "Un chat" || entries[0].Source.URL != "https://fr.wiktionary.org/wiki/chat" {
		t.Fatalf("entries=%+v found=%v err=%v", entries, found, err)
	}
}

func TestGivenAmbiguousFrenchSectionsWhenParsedThenUnavailable(t *testing.T) {
	page := `<h2 id="Français"><span class="sectionlangue" id="fr">Français</span></h2><h3><span class="titredef">Nom commun</span></h3><ol><li>un</li></ol><h2 id="Français"><span class="sectionlangue" id="fr">Français</span></h2>`
	response, _ := json.Marshal(map[string]any{"parse": map[string]string{"title": "test", "text": page}})
	client := &http.Client{Transport: frenchRoundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(response))), Header: make(http.Header)}, nil
	})}
	_, found, err := (FrenchOnlineDictionary{Client: client}).Find(context.Background(), "test", core.French)
	if err == nil || found {
		t.Fatalf("found=%v err=%v", found, err)
	}
}

func TestGivenFrenchProviderResponsesWhenLookedUpThenFailuresAreClassified(t *testing.T) {
	otherOnly, _ := json.Marshal(map[string]any{"parse": map[string]string{"title": "word", "text": `<h2 id="Anglais"><span class="sectionlangue" id="en">Anglais</span></h2><h3><span class="titredef">Noun</span></h3><ol><li>wrong language</li></ol>`}})
	noMarkers, _ := json.Marshal(map[string]any{"parse": map[string]string{"title": "word", "text": `<p>unrecognized page</p><h3><span class="titredef">Noun</span></h3><ol><li>no language heading</li></ol>`}})
	oversized := `{"parse":{"title":"chat","text":"` + strings.Repeat("a", 2*1024*1024+1) + `"}}`
	tests := []struct {
		name       string
		status     int
		body       string
		wantFound  bool
		wantErr    bool
		wantSilent bool
	}{
		{name: "bad gateway is silently unavailable", status: http.StatusBadGateway, body: "error", wantErr: true, wantSilent: true},
		{name: "other non-200 status is an error", status: http.StatusInternalServerError, body: "error", wantErr: true},
		{name: "not found status is an error", status: http.StatusNotFound, body: "error", wantErr: true},
		{name: "malformed JSON is an error", status: http.StatusOK, body: `{"parse":`, wantErr: true},
		{name: "body over the size cap is an error", status: http.StatusOK, body: oversized, wantErr: true},
		{name: "only other-language sections is a reliable miss", status: http.StatusOK, body: string(otherOnly)},
		{name: "page without language markers is an error", status: http.StatusOK, body: string(noMarkers), wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			client := &http.Client{Transport: frenchRoundTrip(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body)), Header: make(http.Header)}, nil
			})}
			// when
			entries, found, err := (FrenchOnlineDictionary{Client: client}).Find(context.Background(), "word", core.French)
			// then
			if found != tc.wantFound || (err != nil) != tc.wantErr || errors.Is(err, core.ErrSilentUnavailable) != tc.wantSilent || len(entries) != 0 {
				t.Fatalf("entries=%+v found=%v err=%v", entries, found, err)
			}
		})
	}
}
