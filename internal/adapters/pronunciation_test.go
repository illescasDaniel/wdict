package adapters

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	stardict "github.com/ianlewis/go-stardict"
	"github.com/ianlewis/go-stardict/idx"
	"github.com/illescasDaniel/wdict/internal/core"
)

func fixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "pronunciation", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// countingClient answers every request with body and counts the requests.
func countingClient(body string, requests *atomic.Int32) *http.Client {
	return &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		requests.Add(1)
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
}

func findOnlineEnglish(t *testing.T, body string) []core.Entry {
	t.Helper()
	var requests atomic.Int32
	entries, found, err := (OnlineDictionary{Client: countingClient(body, &requests)}).Find(context.Background(), "word", "en")
	if err != nil || !found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	return entries
}

func findOnlineSpanish(t *testing.T, body string) []core.Entry {
	t.Helper()
	var requests atomic.Int32
	entries, found, err := (SpanishOnlineDictionary{Client: countingClient(body, &requests)}).Find(context.Background(), "palabra", core.Spanish)
	if err != nil || !found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	return entries
}

func texts(ps []core.Pronunciation) []string {
	var out []string
	for _, p := range ps {
		out = append(out, strings.TrimSpace(p.Label+" "+p.Text))
	}
	return out
}

const englishItem = `[{"word":"word","phonetic":%s,"phonetics":%s,"sourceUrls":["https://en.wiktionary.org/wiki/word"],"license":{"name":"CC BY-SA 3.0","url":"https://creativecommons.org/licenses/by-sa/3.0"},"meanings":[{"partOfSpeech":"noun","definitions":[{"definition":"A unit."}]}]}]`

func englishJSON(phonetic, phonetics string) string {
	return fmt.Sprintf(englishItem, phonetic, phonetics)
}

func TestGivenOnlineHelloFixtureWhenLookedUpThenAudioOnlyItemIsSkippedAndTextsKeptInOrder(t *testing.T) {
	// given
	body := fixture(t, "online_en_hello.json")
	// when
	entries := findOnlineEnglish(t, body)
	// then
	// C2: each text keeps the audio supplied with the same item; the first
	// item has no text, so its audio is not used.
	want := []core.Pronunciation{
		{Text: "/həˈləʊ/", AudioURL: "https://api.dictionaryapi.dev/media/pronunciations/en/hello-uk.mp3"},
		{Text: "/həˈloʊ/"},
	}
	if len(entries) != 1 || !reflect.DeepEqual(entries[0].Pronunciations, want) {
		t.Fatalf("got %+v", entries)
	}
}

func TestGivenOnlineReadFixtureWhenLookedUpThenEachEntryKeepsItsOwnPronunciations(t *testing.T) {
	// given
	body := fixture(t, "online_en_read.json")
	// when
	entries := findOnlineEnglish(t, body)
	// then
	if len(entries) != 2 {
		t.Fatalf("entries=%d", len(entries))
	}
	if got := texts(entries[0].Pronunciations); !reflect.DeepEqual(got, []string{"/ɹiːd/", "/ɹid/"}) {
		t.Fatalf("first=%v", got)
	}
	if got := texts(entries[1].Pronunciations); !reflect.DeepEqual(got, []string{"/ɹɛd/"}) {
		t.Fatalf("second=%v", got)
	}
	want := []core.Pronunciation{
		{Text: "/ɹiːd/", AudioURL: "https://api.dictionaryapi.dev/media/pronunciations/en/read-1-uk.mp3"},
		{Text: "/ɹid/", AudioURL: "https://api.dictionaryapi.dev/media/pronunciations/en/read-1-us.ogg"},
	}
	if !reflect.DeepEqual(entries[0].Pronunciations, want) || entries[1].Pronunciations[0].AudioURL != "https://api.dictionaryapi.dev/media/pronunciations/en/read-2-us.mp3" {
		t.Fatalf("got %+v", entries)
	}
}

func TestGivenEnglishAudioPairedWithItsTextWhenLookedUpThenFirstNonEmptyAudioPerTextIsKept(t *testing.T) {
	// given (C2): first non-empty audio per text, in item order
	body := englishJSON(`null`, `[
		{"text":"/a/","audio":""},
		{"text":"/a/","audio":"https://example.test/first.mp3"},
		{"text":"/a/","audio":"https://example.test/second.mp3"},
		{"text":"/b/","audio":"https://example.test/b.mp3"}
	]`)
	// when
	entries := findOnlineEnglish(t, body)
	// then
	want := []core.Pronunciation{
		{Text: "/a/", AudioURL: "https://example.test/first.mp3"},
		{Text: "/b/", AudioURL: "https://example.test/b.mp3"},
	}
	if !reflect.DeepEqual(entries[0].Pronunciations, want) {
		t.Fatalf("got %+v", entries[0].Pronunciations)
	}
}

func TestGivenEnglishAudioWithUnexpectedShapeWhenLookedUpThenTextIsKeptWithoutURL(t *testing.T) {
	// given (C2): unexpected shapes yield no URL, never an error
	body := englishJSON(`null`, `[{"text":"/a/","audio":5},{"text":"/b/","audio":{"url":"https://example.test/x.mp3"}}]`)
	// when
	entries := findOnlineEnglish(t, body)
	// then
	want := []core.Pronunciation{{Text: "/a/"}, {Text: "/b/"}}
	if !reflect.DeepEqual(entries[0].Pronunciations, want) {
		t.Fatalf("got %+v", entries[0].Pronunciations)
	}
}

func TestGivenOnlineResponseWithOnlyTopLevelPhoneticWhenLookedUpThenItIsUsed(t *testing.T) {
	// given
	body := englishJSON(`"/wɜːd/"`, `[{"audio":"https://example.test/a.mp3"},{"text":"","audio":""}]`)
	// when
	entries := findOnlineEnglish(t, body)
	// then
	if got := texts(entries[0].Pronunciations); !reflect.DeepEqual(got, []string{"/wɜːd/"}) {
		t.Fatalf("got=%v", got)
	}
}

func TestGivenOnlineResponseWithItemTextWhenLookedUpThenTopLevelPhoneticIsIgnored(t *testing.T) {
	// given
	body := englishJSON(`"/top/"`, `[{"text":"/item/"}]`)
	// when
	entries := findOnlineEnglish(t, body)
	// then
	if got := texts(entries[0].Pronunciations); !reflect.DeepEqual(got, []string{"/item/"}) {
		t.Fatalf("got=%v", got)
	}
}

func TestGivenOnlineResponseWithoutPronunciationWhenLookedUpThenListIsEmptyAndDefinitionsRemain(t *testing.T) {
	// given
	body := englishJSON(`null`, `[]`)
	// when
	entries := findOnlineEnglish(t, body)
	// then
	if entries[0].Pronunciations != nil || entries[0].Meanings[0].Definitions[0] != "A unit." {
		t.Fatalf("got %+v", entries[0])
	}
}

func TestGivenOnlineResponseWithInvalidValuesWhenLookedUpThenOnlyValidValueRemains(t *testing.T) {
	// given
	body := englishJSON(`null`, `[{"text":""},{"text":"/<b>x</b>/"},{"text":"/a\u0000b/"},{"text":"/`+strings.Repeat("a", 64)+`/"},{"text":"/ok/"}]`)
	// when
	entries := findOnlineEnglish(t, body)
	// then
	if got := texts(entries[0].Pronunciations); !reflect.DeepEqual(got, []string{"/ok/"}) {
		t.Fatalf("got=%v", got)
	}
}

func TestGivenOnlineResponseWithManyVariantsWhenLookedUpThenFirstThreeDistinctAreKept(t *testing.T) {
	// given
	body := englishJSON(`null`, `[{"text":"/a/"},{"text":"/b/"},{"text":"/a/"},{"text":"/c/"},{"text":"/d/"},{"text":"/e/"}]`)
	// when
	entries := findOnlineEnglish(t, body)
	// then
	if got := texts(entries[0].Pronunciations); !reflect.DeepEqual(got, []string{"/a/", "/b/", "/c/"}) {
		t.Fatalf("got=%v", got)
	}
}

func TestGivenOnlineResponseWithUnexpectedPronunciationShapeWhenLookedUpThenStillFoundWithoutPronunciations(t *testing.T) {
	for name, parts := range map[string][2]string{
		"phonetics is a string":     {`null`, `"oops"`},
		"phonetics is an object":    {`null`, `{"text":"/a/"}`},
		"items are numbers":         {`null`, `[1,2]`},
		"text is a number":          {`null`, `[{"text":5}]`},
		"phonetic is an object":     {`{"a":1}`, `[]`},
		"phonetic is a number":      {`7`, `null`},
		"text is an array of texts": {`null`, `[{"text":["/a/"]}]`},
	} {
		t.Run(name, func(t *testing.T) {
			// given
			body := englishJSON(parts[0], parts[1])
			// when
			entries := findOnlineEnglish(t, body)
			// then
			if entries[0].Pronunciations != nil || len(entries[0].Meanings) != 1 {
				t.Fatalf("got %+v", entries[0])
			}
		})
	}
}

func TestGivenOnlineEnglishLookupWhenFoundThenExactlyOneRequestIsMade(t *testing.T) {
	// given
	var requests atomic.Int32
	client := countingClient(fixture(t, "online_en_hello.json"), &requests)
	// when
	_, found, err := (OnlineDictionary{Client: client}).Find(context.Background(), "hello", "en")
	// then
	if err != nil || !found || requests.Load() != 1 {
		t.Fatalf("found=%v err=%v requests=%d", found, err, requests.Load())
	}
}

func TestGivenSavedSpanishResponsesWhenLookedUpThenOnlyTheSpanishValueIsReturned(t *testing.T) {
	// C2: the phonos audio button is resolved to an https file URL.
	for word, want := range map[string]core.Pronunciation{
		"hola":  {Text: "[ˈola]", AudioURL: "https://upload.wikimedia.org/wikipedia/commons/transcoded/a/a2/Es-hola.oga/Es-hola.oga.mp3"},
		"casa":  {Text: "[ˈkasa]", AudioURL: "https://upload.wikimedia.org/wikipedia/commons/transcoded/8/81/Es-am-lat-casa.ogg/Es-am-lat-casa.ogg.mp3"},
		"perro": {Text: "[ˈpero]", AudioURL: "https://upload.wikimedia.org/wikipedia/commons/transcoded/2/28/LL-Q1321_%28spa%29-Rodelar-perro.wav/LL-Q1321_%28spa%29-Rodelar-perro.wav.mp3"},
		"agua":  {Text: "[ˈaɣ̞wa]", AudioURL: "https://upload.wikimedia.org/wikipedia/commons/transcoded/9/98/Es-am-lat-agua.ogg/Es-am-lat-agua.ogg.mp3"},
		"leer":  {Text: "[leˈeɾ]"},
	} {
		t.Run(word, func(t *testing.T) {
			// given
			body := fixture(t, "online_es_"+word+".json")
			// when
			entries := findOnlineSpanish(t, body)
			// then
			if len(entries) != 1 || !reflect.DeepEqual(entries[0].Pronunciations, []core.Pronunciation{want}) {
				t.Fatalf("got %+v", entries[0].Pronunciations)
			}
			if len(entries[0].Meanings) == 0 {
				t.Fatal("definitions missing")
			}
		})
	}
}

func spanishPage(sections ...string) string {
	return `{"parse":{"title":"palabra","text":` + jsonString(strings.Join(sections, "")) + `}}`
}

func jsonString(s string) string {
	s = strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`).Replace(s)
	return `"` + s + `"`
}

const spanishDefinition = `<h4>Sustantivo</h4><dl><dt>1</dt><dd>Una palabra.</dd></dl>`

func pronunciationRow(label, value string) string {
	return `<table><tbody><tr><td><b>` + label + `</b> (<a href="/wiki/AFI">AFI</a>)</td><td>` + value + `</td></tr></tbody></table>`
}

func TestGivenOnlySpanishSectionHasNoRowWhenLookedUpThenOtherLanguagesRowsAreNotUsed(t *testing.T) {
	// given
	body := spanishPage(`<h2 id="Español">Español</h2>`, spanishDefinition, `<h2 id="Italiano">Italiano</h2>`, pronunciationRow("pronunciación", "/ita/"))
	// when
	entries := findOnlineSpanish(t, body)
	// then
	if entries[0].Pronunciations != nil {
		t.Fatalf("got %+v", entries[0].Pronunciations)
	}
}

func TestGivenOtherLanguageBeforeSpanishWhenLookedUpThenItsRowIsNotUsed(t *testing.T) {
	// given
	body := spanishPage(`<h2 id="Catalán">Catalán</h2>`, pronunciationRow("pronunciación", "[ˈo.la]"), `<h2 id="Español">Español</h2>`, spanishDefinition, pronunciationRow("pronunciación", "[ˈola]"))
	// when
	entries := findOnlineSpanish(t, body)
	// then
	if got := texts(entries[0].Pronunciations); !reflect.DeepEqual(got, []string{"[ˈola]"}) {
		t.Fatalf("got=%v", got)
	}
}

func TestGivenSpanishRowsWithOtherLabelsWhenLookedUpThenOnlyPronunciationRowIsUsed(t *testing.T) {
	// given
	body := spanishPage(`<h2 id="Español">Español</h2>`, spanishDefinition, pronunciationRow("silabación", "ho-la"), pronunciationRow("central", "[ˈx]"), pronunciationRow("pronunciación", `[ˈola] <span class="ext-phonos"><span>▶ audio</span></span><br />[ˈoɫa]<br />`))
	// when
	entries := findOnlineSpanish(t, body)
	// then
	if got := texts(entries[0].Pronunciations); !reflect.DeepEqual(got, []string{"[ˈola]", "[ˈoɫa]"}) {
		t.Fatalf("got=%v", got)
	}
}

func TestGivenSpanishPhonosMarkupWhenLookedUpThenTheAudioReferenceIsResolvedToHTTPS(t *testing.T) {
	for name, value := range map[string]string{
		"scheme relative":   `<span class="ext-phonos"><a role="button" href="//upload.example.test/w/x.ogg/x.ogg.mp3">▶</a></span>[ˈola]<br />`,
		"already https":     `<span class="ext-phonos"><a role="button" href="https://upload.example.test/w/x.ogg/x.ogg.mp3">▶</a></span>[ˈola]<br />`,
		"audio before text": `<span class="ext-phonos"><a role="button" href="//upload.example.test/w/x.ogg/x.ogg.mp3">▶</a></span> <br />[ˈola]<br />`,
	} {
		t.Run(name, func(t *testing.T) {
			// given
			body := spanishPage(`<h2 id="Español">Español</h2>`, spanishDefinition, pronunciationRow("pronunciación", value))
			// when
			entries := findOnlineSpanish(t, body)
			// then
			want := core.Pronunciation{Text: "[ˈola]", AudioURL: "https://upload.example.test/w/x.ogg/x.ogg.mp3"}
			if !reflect.DeepEqual(entries[0].Pronunciations, []core.Pronunciation{want}) {
				t.Fatalf("got %+v", entries[0].Pronunciations)
			}
		})
	}
}

func TestGivenSpanishPhonosWithUnusableLinkWhenLookedUpThenNoAudioURLIsSet(t *testing.T) {
	for name, value := range map[string]string{
		"relative link":  `<span class="ext-phonos"><a role="button" href="/wiki/Archivo:x.ogg">▶</a></span>[ˈola]<br />`,
		"plain http":     `<span class="ext-phonos"><a role="button" href="http://upload.example.test/x.mp3">▶</a></span>[ˈola]<br />`,
		"no link":        `<span class="ext-phonos"><span>▶ audio</span></span>[ˈola]<br />`,
		"nested no link": `<span class="ext-phonos"><span><span class="ext-phonos"><span>▶</span></span></span></span>[ˈola]<br />`,
	} {
		t.Run(name, func(t *testing.T) {
			// given
			body := spanishPage(`<h2 id="Español">Español</h2>`, spanishDefinition, pronunciationRow("pronunciación", value))
			// when
			entries := findOnlineSpanish(t, body)
			// then
			want := core.Pronunciation{Text: "[ˈola]"}
			if !reflect.DeepEqual(entries[0].Pronunciations, []core.Pronunciation{want}) {
				t.Fatalf("got %+v", entries[0].Pronunciations)
			}
		})
	}
}

func TestGivenTwoSpanishValuesWithOneAudioWhenLookedUpThenTheAudioStaysWithItsOwnValue(t *testing.T) {
	// given (C2): the second phonos button belongs to the second value; the
	// fixture for casa has the Latin-American button next to [ˈkasa]
	body := spanishPage(`<h2 id="Español">Español</h2>`, spanishDefinition, pronunciationRow("pronunciación",
		`[ˈkasa] <span class="ext-phonos"><a role="button" href="//upload.example.test/one.mp3">▶</a></span><br /> <span class="ext-phonos"><a role="button" href="//upload.example.test/two.mp3">▶</a></span><br />`))
	// when
	entries := findOnlineSpanish(t, body)
	// then
	want := []core.Pronunciation{{Text: "[ˈkasa]", AudioURL: "https://upload.example.test/one.mp3"}}
	if !reflect.DeepEqual(entries[0].Pronunciations, want) {
		t.Fatalf("got %+v", entries[0].Pronunciations)
	}
}

func TestGivenSpanishValueWithMarkupTextWhenLookedUpThenItIsDropped(t *testing.T) {
	// given
	body := spanishPage(`<h2 id="Español">Español</h2>`, spanishDefinition, pronunciationRow("pronunciación", `&lt;b&gt;x&lt;/b&gt;<br />[ˈola]`))
	// when
	entries := findOnlineSpanish(t, body)
	// then
	if got := texts(entries[0].Pronunciations); !reflect.DeepEqual(got, []string{"[ˈola]"}) {
		t.Fatalf("got=%v", got)
	}
}

func TestGivenSpanishRowWithoutValueCellWhenLookedUpThenStillFoundWithoutPronunciations(t *testing.T) {
	for name, row := range map[string]string{
		"no value cell":  `<table><tbody><tr><td><b>pronunciación</b> (AFI)</td></tr></tbody></table>`,
		"empty value":    pronunciationRow("pronunciación", ` `),
		"row not in tr":  `<div><b>pronunciación</b> (AFI) [ˈola]</div>`,
		"label only":     `<table><tbody><tr><td>pronunciación (AFI)</td><td></td></tr></tbody></table>`,
		"not an AFI row": `<table><tbody><tr><td><b>pronunciación</b></td><td>[ˈola]</td></tr></tbody></table>`,
	} {
		t.Run(name, func(t *testing.T) {
			// given
			body := spanishPage(`<h2 id="Español">Español</h2>`, spanishDefinition, row)
			// when
			entries := findOnlineSpanish(t, body)
			// then
			if entries[0].Pronunciations != nil || len(entries[0].Meanings) != 1 {
				t.Fatalf("got %+v", entries[0])
			}
		})
	}
}

func TestGivenSpanishLookupWhenFoundThenExactlyOneRequestIsMade(t *testing.T) {
	// given
	var requests atomic.Int32
	client := countingClient(fixture(t, "online_es_hola.json"), &requests)
	// when
	_, found, err := (SpanishOnlineDictionary{Client: client}).Find(context.Background(), "hola", core.Spanish)
	// then
	if err != nil || !found || requests.Load() != 1 {
		t.Fatalf("found=%v err=%v requests=%d", found, err, requests.Load())
	}
}

// writeHTMLDictionary builds an uncompressed StarDict (sametypesequence=h)
// whose index lists each (word, html) pair in order.
func writeHTMLDictionary(t *testing.T, pairs ...[2]string) string {
	t.Helper()
	dir := t.TempDir()
	var idx, data bytes.Buffer
	for _, p := range pairs {
		idx.WriteString(p[0] + "\x00")
		_ = binary.Write(&idx, binary.BigEndian, uint32(data.Len()))
		_ = binary.Write(&idx, binary.BigEndian, uint32(len(p[1])))
		data.WriteString(p[1])
	}
	ifo := fmt.Sprintf("StarDict's dict ifo file\nversion=3.0.0\nbookname=Fixture\nwordcount=%d\nidxfilesize=%d\nsametypesequence=h\n", len(pairs), idx.Len())
	path := filepath.Join(dir, "dictionary.ifo")
	for name, content := range map[string][]byte{"dictionary.ifo": []byte(ifo), "dictionary.idx": idx.Bytes(), "dictionary.dict": data.Bytes()} {
		if err := os.WriteFile(filepath.Join(dir, name), content, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

func findOffline(t *testing.T, language core.Language, word string, pairs ...[2]string) []core.Entry {
	t.Helper()
	d := &OfflineDictionary{IFO: writeHTMLDictionary(t, pairs...), Language: language, Source: core.Source{ID: "fixture"}}
	lookup := language
	if lookup == "" {
		lookup = "en"
	}
	entries, found, err := d.Find(context.Background(), word, lookup)
	if err != nil || !found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	return entries
}

func TestGivenOfflineHelloEntriesPerPartOfSpeechWhenLookedUpThenEachHasUSUKAndNorthumbriaOnly(t *testing.T) {
	// given
	pairs := [][2]string{
		{"hello", fixture(t, "offline_en_hello_interjection.html")},
		{"hello", fixture(t, "offline_en_hello_noun.html")},
		{"hello", fixture(t, "offline_en_hello_verb.html")},
	}
	// when
	entries := findOffline(t, "", "hello", pairs...)
	// then
	if len(entries) != 3 {
		t.Fatalf("entries=%d", len(entries))
	}
	want := []core.Pronunciation{
		{Text: "/hɛˈloʊ/", Label: "🇺🇸", LabelTitle: "General American"},
		{Text: "/həˈləʊ/", Label: "🇬🇧", LabelTitle: "Received Pronunciation"},
		{Text: "/haˈlaʊ/", Label: "Northumbria"},
	}
	if len(entries) < 1 {
		t.Fatal("no entries found for vegetable fixture")
	}
	for i, entry := range entries {
		if !reflect.DeepEqual(entry.Pronunciations, want) {
			t.Fatalf("entry %d: got %+v", i, entry.Pronunciations)
		}
	}
	if entries[0].Meanings[0].PartOfSpeech != "Interjection" {
		t.Fatalf("definitions changed: %+v", entries[0].Meanings)
	}
}

func TestGivenOfflineEnglishRepeatedAndOverflowingAccentsWhenLookedUpThenDuplicatesMergeAndCapApplies(t *testing.T) {
	// given
	item := func(accent, ipa string) string {
		return `<li><span class="usage-label-accent">` + accent + `</span> IPA: <span class="IPA nowrap">` + ipa + `</span><span class="IPA nowrap">/ignored/</span></li>`
	}
	page := `<section><h4>Noun</h4><ul>` + item("A", "/a/") + item("B", "/b/") + item("C2", "/a/") + item("C", "/c/") + item("D", "/d/") + `</ul><ol><li>Thing.</li></ol></section>`
	// when
	entries := findOffline(t, "", "word", [2]string{"word", page})
	// then
	want := []core.Pronunciation{{Text: "/a/", Label: "A"}, {Text: "/b/", Label: "B"}, {Text: "/c/", Label: "C"}}
	if !reflect.DeepEqual(entries[0].Pronunciations, want) {
		t.Fatalf("got %+v", entries[0].Pronunciations)
	}
}

func TestGivenOfflineEnglishUnlabelledIPAWhenLookedUpThenLabelIsEmpty(t *testing.T) {
	// given
	page := `<section><h4>Noun</h4><ul><li>IPA: <span class="IPA nowrap">/wɜːd/</span></li></ul><ol><li>Thing.</li></ol></section>`
	// when
	entries := findOffline(t, "", "word", [2]string{"word", page})
	// then
	if want := []core.Pronunciation{{Text: "/wɜːd/"}}; !reflect.DeepEqual(entries[0].Pronunciations, want) {
		t.Fatalf("got %+v", entries[0].Pronunciations)
	}
}

func TestGivenOfflineEnglishWithoutIPAWhenLookedUpThenListIsEmptyAndDefinitionsRemain(t *testing.T) {
	for name, page := range map[string]string{
		"no pronunciation":             `<section><h4>Verb</h4><ol><li>To do a thing.</li></ol></section>`,
		"only enPR":                    `<section><h4>Verb</h4><ul><li>enPR: <span class="AHD enPR">hĕ-lō'</span></li></ul><ol><li>To do a thing.</li></ol></section>`,
		"IPA outside a list item":      `<section><h4>Verb</h4><div><span class="IPA">/x/</span></div><ol><li>To do a thing.</li></ol></section>`,
		"IPA span is empty":            `<section><h4>Verb</h4><ul><li><span class="IPA"></span></li></ul><ol><li>To do a thing.</li></ol></section>`,
		"markup text in IPA":           `<section><h4>Verb</h4><ul><li><span class="IPA">&lt;b&gt;</span></li></ul><ol><li>To do a thing.</li></ol></section>`,
		"plain text entry, not HTML":   `plain`,
		"broken markup around the IPA": `<section><h4>Verb</h4><ul><li><span class="IPA"><ol><li>To do a thing.</li></ol>`,
	} {
		t.Run(name, func(t *testing.T) {
			// when
			entries := findOffline(t, "", "word", [2]string{"word", page})
			// then
			if entries[0].Pronunciations != nil || len(entries[0].Meanings) == 0 {
				t.Fatalf("got %+v", entries[0])
			}
		})
	}
}

func TestGivenOfflineEnglishRhymeItemAndDefinitionIPAWhenLookedUpThenOnlyIPAItemsOfTheEntryAreUsed(t *testing.T) {
	// given
	rhymes := `<ul><li><span>Rhymes: </span><span class="IPA nowrap">-ɑːt</span>` +
		`<ul><li><span class="usage-label-accent">UK</span><span> </span>IPA<span>:</span><span> </span><span class="IPA nowrap">/ˈdɪktæt/</span></li></ul></li>` +
		`<li><span class="usage-label-accent">US</span><span> </span>IPA<span>:</span><span> </span><span class="IPA nowrap">/dɪkˈtɑt/</span></li></ul>`
	for name, page := range map[string]string{
		"rhyme item with nested IPA item": `<section><h4>Noun</h4>` + rhymes + `<ol><li>A decree.</li></ol></section>`,
		"IPA span inside a definition":    `<section><h4>Noun</h4>` + rhymes + `<ol><li>Sounds like <span class="IPA">/t/</span>.</li></ol></section>`,
	} {
		t.Run(name, func(t *testing.T) {
			// when
			entries := findOffline(t, "", "diktat", [2]string{"diktat", page})
			// then
			want := []core.Pronunciation{{Text: "/ˈdɪktæt/", Label: "🇬🇧", LabelTitle: "UK"}, {Text: "/dɪkˈtɑt/", Label: "🇺🇸", LabelTitle: "US"}}
			if !reflect.DeepEqual(entries[0].Pronunciations, want) {
				t.Fatalf("got %+v", entries[0].Pronunciations)
			}
		})
	}
	t.Run("definition-only IPA gives no pronunciation", func(t *testing.T) {
		// given
		page := `<section><h4>Suffix</h4><ol><li>Sounds like <span class="IPA">/t/</span>.</li></ol></section>`
		// when
		entries := findOffline(t, "", "-t", [2]string{"-t", page})
		// then
		if entries[0].Pronunciations != nil {
			t.Fatalf("got %+v", entries[0].Pronunciations)
		}
	})
}

func TestGivenOfflineSpanishFixturesWhenLookedUpThenAFIValueKeepsBracketsWithoutLabel(t *testing.T) {
	for word, want := range map[string][]core.Pronunciation{
		"hola":  {{Text: "[ˈola]"}},
		"casa":  {{Text: "[ˈkasa]"}},
		"perro": {{Text: "[ˈpero]"}},
	} {
		t.Run(word, func(t *testing.T) {
			// given
			pairs := [][2]string{{word, fixture(t, "offline_es_"+word+".html")}}
			// when
			entries := findOffline(t, core.Spanish, word, pairs...)
			// then
			if !reflect.DeepEqual(entries[0].Pronunciations, want) {
				t.Fatalf("got %+v", entries[0].Pronunciations)
			}
			if len(entries[0].Meanings) == 0 {
				t.Fatal("definitions missing")
			}
		})
	}
}

func TestGivenOfflineSpanishWithoutAFIItemWhenLookedUpThenListIsEmpty(t *testing.T) {
	for name, list := range map[string]string{
		"no pronunciation":           `<ul><li><b>silabación</b>: <span>ho-la</span></li></ul>`,
		"IPA in another item":        `<ul><li><b>rima</b>: <span class="IPA">[x]</span></li></ul>`,
		"AFI item without IPA class": `<ul><li><b>pronunciación</b> (AFI): [ˈola]</li></ul>`,
	} {
		t.Run(name, func(t *testing.T) {
			// given
			page := `<section lang="es"><h4>Interjección</h4>` + list + `<dl><dt>1</dt><dd>Saludo.</dd></dl></section>`
			// when
			entries := findOffline(t, core.Spanish, "hola", [2]string{"hola", page})
			// then
			if entries[0].Pronunciations != nil || len(entries[0].Meanings) != 1 {
				t.Fatalf("got %+v", entries[0])
			}
		})
	}
}

func TestGivenOfflineWordLookedUpTwiceWhenIndexIsGoneThenOneScanServesBothLookups(t *testing.T) {
	// given
	scans := 0
	original := scanIndex
	scanIndex = func(ctx context.Context, d *stardict.Stardict, path, query string) ([]*idx.Word, error) {
		scans++
		return original(ctx, d, path, query)
	}
	t.Cleanup(func() { scanIndex = original })
	d := &OfflineDictionary{IFO: writeHTMLDictionary(t, [2]string{"hola", fixture(t, "offline_es_hola.html")}), Language: core.Spanish, Source: core.Source{ID: "fixture"}}
	first, _, err := d.Find(context.Background(), "hola", core.Spanish)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(strings.TrimSuffix(d.IFO, ".ifo") + ".idx"); err != nil {
		t.Fatal(err)
	}
	// when
	second, found, err := d.Find(context.Background(), "hola", core.Spanish)
	// then
	if err != nil || !found || !reflect.DeepEqual(first, second) || len(second[0].Pronunciations) != 1 {
		t.Fatalf("found=%v err=%v first=%+v second=%+v", found, err, first, second)
	}
	if scans != 1 {
		t.Fatalf("index scans=%d, want 1 for one uncached lookup", scans)
	}
}

func diskRecordFor(key core.CacheKey, entries ...core.Entry) core.CacheRecord {
	now := time.Now()
	return core.CacheRecord{Key: key, FetchedAt: now, LastUsed: now, Entries: entries}
}

func TestGivenRecordWrittenBeforeFeatureWhenReadFromDiskThenItIsValidWithoutPronunciations(t *testing.T) {
	// given
	dir := t.TempDir()
	key := core.CacheKey{Provider: "p", Language: "en", NormalizedQuery: "hello"}
	record := diskRecordFor(key, core.Entry{Headword: "hello", Language: "en", Meanings: []core.Meaning{{Definitions: []string{"a greeting"}}}})
	cache := &DiskCache{Dir: dir}
	if err := cache.Put(context.Background(), record); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, cacheName(key)))
	if err != nil {
		t.Fatal(err)
	}
	// when
	got, found, err := cache.Get(context.Background(), key)
	// then
	if err != nil || !found || got.Entries[0].Pronunciations != nil {
		t.Fatalf("found=%v err=%v got=%+v", found, err, got)
	}
	if strings.Contains(string(raw), "Pronunciations") || !strings.HasPrefix(string(raw), `{"Version":1,`) {
		t.Fatalf("record=%s", raw)
	}
}

func TestGivenRecordWithPronunciationsWhenRoundTrippedThroughMemoryAndDiskThenUnchanged(t *testing.T) {
	// given
	key := core.CacheKey{Provider: "p", Language: "en", NormalizedQuery: "hello"}
	pronunciations := []core.Pronunciation{{Text: "/hɛˈloʊ/", Label: "🇺🇸", LabelTitle: "US"}, {Text: "/həˈləʊ/"}}
	record := diskRecordFor(key, core.Entry{Headword: "hello", Language: "en", Pronunciations: pronunciations, Meanings: []core.Meaning{{Definitions: []string{"a greeting"}}}})
	disk, memory := &DiskCache{Dir: t.TempDir()}, &MemoryCache{}
	// when
	if err := disk.Put(context.Background(), record); err != nil {
		t.Fatal(err)
	}
	if err := memory.Put(context.Background(), record); err != nil {
		t.Fatal(err)
	}
	fromDisk, diskFound, diskErr := (&DiskCache{Dir: disk.Dir}).Get(context.Background(), key)
	fromMemory, memoryFound, memoryErr := memory.Get(context.Background(), key)
	// then
	if diskErr != nil || memoryErr != nil || !diskFound || !memoryFound {
		t.Fatalf("disk=%v/%v memory=%v/%v", diskFound, diskErr, memoryFound, memoryErr)
	}
	if !reflect.DeepEqual(fromDisk.Entries[0].Pronunciations, pronunciations) || !reflect.DeepEqual(fromMemory.Entries[0].Pronunciations, pronunciations) {
		t.Fatalf("disk=%+v memory=%+v", fromDisk.Entries[0].Pronunciations, fromMemory.Entries[0].Pronunciations)
	}
}

func TestGivenRecordWithInvalidPronunciationsWhenReadFromDiskThenValuesAreDroppedAndRecordKept(t *testing.T) {
	// given
	dir := t.TempDir()
	key := core.CacheKey{Provider: "p", Language: "en", NormalizedQuery: "hello"}
	bad := []core.Pronunciation{{Text: ""}, {Text: "/<b>x</b>/"}, {Text: "/a\x01/"}, {Text: "/" + strings.Repeat("a", 64) + "/"}, {Text: "/ok/", Label: "<i>"}}
	record := diskRecordFor(key, core.Entry{Headword: "hello", Language: "en", Pronunciations: bad, Meanings: []core.Meaning{{Definitions: []string{"a greeting"}}}})
	if err := (&DiskCache{Dir: dir}).Put(context.Background(), record); err != nil {
		t.Fatal(err)
	}
	// when
	got, found, err := (&DiskCache{Dir: dir}).Get(context.Background(), key)
	// then
	want := []core.Pronunciation{{Text: "/ok/"}}
	if err != nil || !found || !reflect.DeepEqual(got.Entries[0].Pronunciations, want) {
		t.Fatalf("found=%v err=%v got=%+v", found, err, got.Entries)
	}
}

func TestGivenOldCachedRecordWithTextLabelsWhenReadFromDiskThenLabelsAreRemappedToFlags(t *testing.T) {
	// given
	dir := t.TempDir()
	key := core.CacheKey{Provider: "p", Language: "en", NormalizedQuery: "hello"}
	old := []core.Pronunciation{{Text: "/a/", Label: "US"}, {Text: "/b/", Label: "Received Pronunciation"}, {Text: "/c/", Label: "Northumbria"}}
	record := diskRecordFor(key, core.Entry{Headword: "hello", Language: "en", Pronunciations: old, Meanings: []core.Meaning{{Definitions: []string{"a greeting"}}}})
	if err := (&DiskCache{Dir: dir}).Put(context.Background(), record); err != nil {
		t.Fatal(err)
	}
	// when
	got, found, err := (&DiskCache{Dir: dir}).Get(context.Background(), key)
	// then
	want := []core.Pronunciation{{Text: "/a/", Label: "🇺🇸", LabelTitle: "US"}, {Text: "/b/", Label: "🇬🇧", LabelTitle: "Received Pronunciation"}, {Text: "/c/", Label: "Northumbria"}}
	if err != nil || !found || !reflect.DeepEqual(got.Entries[0].Pronunciations, want) {
		t.Fatalf("found=%v err=%v got=%+v", found, err, got.Entries)
	}
}

func TestGivenOfflineVegetableFixtureWhenLookedUpThenGenericHasNoLabelAndQualifiersAreIgnored(t *testing.T) {
	// given
	pair := [2]string{"vegetable", fixture(t, "offline_en_vegetable.html")}
	// when
	entries := findOffline(t, "", "vegetable", pair)
	// then
	want := []core.Pronunciation{
		{Text: "/ˈvɛd͡ʒ.tə.bəl/"},
		{Text: "/ˈvɛd͡ʒ.ə.tə.bəl/", Label: "🇬🇧", LabelTitle: "UK, dialectal"},
		{Text: "/ˈvɛt͡ʃ.tə.bəl/", Label: "🇺🇸🇨🇦", LabelTitle: "US, Canada, dialectal"},
	}
	for i, entry := range entries {
		if !reflect.DeepEqual(entry.Pronunciations, want) {
			t.Fatalf("entry %d: got %+v", i, entry.Pronunciations)
		}
	}
}
