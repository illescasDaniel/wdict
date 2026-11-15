package adapters

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/illescasDaniel/wdict/internal/core"
	"golang.org/x/net/html"
)

func senseGroups(t *testing.T, meanings []core.Meaning) []string {
	t.Helper()
	b, err := json.Marshal(meanings)
	if err != nil {
		t.Fatal(err)
	}
	var wire []struct{ SenseGroup string }
	if err := json.Unmarshal(b, &wire); err != nil {
		t.Fatal(err)
	}
	groups := make([]string, len(wire))
	for i := range wire {
		groups[i] = wire[i].SenseGroup
	}
	return groups
}
func parseSenses(t *testing.T, raw string) []core.Meaning {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := spanishMeanings(doc, true)
	return got
}
func TestGivenInstalledTouchWhenExtractedThenSixConsecutiveSensesShareEvidence(t *testing.T) {
	raw, err := os.ReadFile("testdata/compact-senses/offline_es_touch.html")
	if err != nil {
		t.Fatal(err)
	}
	got := parseSenses(t, string(raw))
	groups := senseGroups(t, got)
	want := []string{"Cada una de las líneas laterales del campo de juego.", "Relación sentimental carente de compromiso.", "Cantidad pequeña, en especial de tiempo.", "Sello, carácter distintivo que una persona imprime en las cosas.", "Operación estética pequeña.", "Comentario breve, digresión."}
	if len(got) != len(want) {
		t.Fatalf("meanings=%+v", got)
	}
	for i := range got {
		if got[i].PartOfSpeech != "Sustantivo masculino" || !reflect.DeepEqual(got[i].Definitions, []string{want[i]}) {
			t.Fatalf("meaning %d=%+v", i, got[i])
		}
		if groups[i] == "" || groups[i] != groups[0] {
			t.Fatalf("groups=%q", groups)
		}
	}
}
func TestGivenBoundariesWhenExtractedThenRunsAreIsolated(t *testing.T) {
	cases := []struct{ name, between string }{
		{"repeated heading", "<h4>noun</h4>"}, {"section", "<section></section>"}, {"other heading", "<h3>context</h3><h4>noun</h4>"}, {"content", "<p>context</p>"}, {"unfamiliar container", "<div>context</div>"}, {"unpaired number", "<dl><dt>2</dt></dl>"}, {"empty candidate", "<dl><dt>2</dt><dd> </dd></dl>"}, {"pronunciation", "<p>pronunciación: /x/</p>"}, {"etymology", "<h3>Etimología</h3><h4>noun</h4>"}, {"language", "<h2 id=Español>Español</h2><h4>noun</h4>"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := "<section><h4>noun</h4><dl><dt>1</dt><dd>one</dd></dl>" + tc.between + "<dl><dt>3</dt><dd>three</dd></dl></section>"
			got := parseSenses(t, raw)
			groups := senseGroups(t, got)
			if len(groups) != 2 || groups[0] == "" || groups[1] == groups[0] {
				t.Fatalf("meanings=%+v groups=%q", got, groups)
			}
		})
	}
}
func TestGivenAdjacentListsWhenExtractedThenTheyShareEvidence(t *testing.T) {
	got := parseSenses(t, "<section><h4>noun</h4><dl><dt>1</dt><dd>one</dd></dl> \n <dl><dt>2</dt><dd>two</dd></dl></section>")
	groups := senseGroups(t, got)
	if len(groups) != 2 || groups[0] == "" || groups[0] != groups[1] {
		t.Fatalf("groups=%q", groups)
	}
}
func TestGivenMissingHeadingWhenExtractedThenNoEvidenceIsSupplied(t *testing.T) {
	for _, heading := range []string{"", "<h4>  </h4>"} {
		got := parseSenses(t, "<section>"+heading+"<dl><dt>1</dt><dd>one</dd><dt>2</dt><dd>two</dd></dl></section>")
		for _, g := range senseGroups(t, got) {
			if g != "" {
				t.Fatalf("groups=%q", senseGroups(t, got))
			}
		}
	}
}

func TestGivenMultipleOfflinePayloadsWhenExtractedThenTokensDoNotCollide(t *testing.T) {
	raw := "<section><h4>noun</h4><dl><dt>1</dt><dd>one</dd><dt>2</dt><dd>two</dd></dl></section>"
	path := writeHTMLDictionary(t, [2]string{"word", raw + "\x00" + raw})
	ifo, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, []byte(strings.Replace(string(ifo), "sametypesequence=h\n", "sametypesequence=hh\n", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	d := &OfflineDictionary{IFO: path, Language: core.Spanish}
	entries, found, err := d.Find(context.Background(), "word", core.Spanish)
	if err != nil || !found || len(entries) != 1 {
		t.Fatalf("found=%v entries=%+v err=%v", found, entries, err)
	}
	groups := senseGroups(t, entries[0].Meanings)
	if len(groups) != 4 || groups[0] == "" || groups[0] != groups[1] || groups[2] != groups[3] || groups[0] == groups[2] {
		t.Fatalf("groups=%q", groups)
	}
}

func TestGivenSameSpanishPayloadWhenOnlineAndOfflineThenEvidenceAndStringsMatch(t *testing.T) {
	raw, err := os.ReadFile("testdata/compact-senses/offline_es_touch.html")
	if err != nil {
		t.Fatal(err)
	}
	offline := parseSenses(t, string(raw))
	onlineDoc, err := html.Parse(strings.NewReader(`<h2 id="Español">Español</h2>` + string(raw)))
	if err != nil {
		t.Fatal(err)
	}
	online, sections := spanishMeanings(onlineDoc, false)
	if sections != 1 || !reflect.DeepEqual(online, offline) {
		t.Fatalf("sections=%d online=%+v offline=%+v", sections, online, offline)
	}
}

func TestGivenSyntheticSectionFixtureWhenExtractedThenEachHeadingAndSectionIsIsolated(t *testing.T) {
	raw, err := os.ReadFile("testdata/compact-senses/spanish_boundaries.html")
	if err != nil {
		t.Fatal(err)
	}
	groups := senseGroups(t, parseSenses(t, string(raw)))
	if len(groups) != 3 || groups[0] == "" || groups[1] == "" || groups[2] == "" || groups[0] == groups[1] || groups[1] == groups[2] || groups[0] == groups[2] {
		t.Fatalf("groups=%q", groups)
	}
}

func TestGivenNoActualSectionWhenExtractedThenEvidenceIsUnknown(t *testing.T) {
	got := parseSenses(t, "<h4>noun</h4><dl><dt>1</dt><dd>one</dd><dt>2</dt><dd>two</dd></dl>")
	groups := senseGroups(t, got)
	if len(groups) != 2 || groups[0] != "" || groups[1] != "" {
		t.Fatalf("groups=%q", groups)
	}
}
