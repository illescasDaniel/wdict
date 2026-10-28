package adapters

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/illescasDaniel/wdict/internal/core"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	stardict "github.com/ianlewis/go-stardict"
)

func TestGivenWiktionaryHTMLWhenExtractedThenDefinitionsUseLinkText(t *testing.T) {
	// given
	raw := `<section><h4>Verb</h4><ul><li>IPA: test</li></ul><p>AI forms</p><ol><li>(informal) To <a href="bword://use">use</a> <a href="bword://artificial_intelligence">artificial intelligence</a>.<dl><dd>Example</dd></dl></li><li>Another &amp; meaning.</li></ol><section><h4>Alternative forms</h4><ol><li>not a definition</li></ol></section></section>`
	want := []core.Meaning{{PartOfSpeech: "Verb", Definitions: []string{"(informal) To use artificial intelligence.", "Another & meaning."}}}
	// when
	// then
	if got := htmlMeanings(raw); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestGivenPunctuationEntriesWhenScannedThenAllExactEntriesAreFound(t *testing.T) {
	// given
	dir := t.TempDir()
	path := filepath.Join(dir, "dictionary.ifo")
	var index bytes.Buffer
	for i := 0; i < 130; i++ {
		index.WriteString("?\x00")
		_ = binary.Write(&index, binary.BigEndian, uint32(i))
		_ = binary.Write(&index, binary.BigEndian, uint32(1))
	}
	ifo := fmt.Sprintf("StarDict's dict ifo file\nversion=3.0.0\nbookname=Fixture\nwordcount=130\nidxfilesize=%d\n", index.Len())
	if err := os.WriteFile(path, []byte(ifo), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "dictionary.idx"), index.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	d, err := stardict.Open(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	// when
	matches, err := scanDictionary(context.Background(), d, path, "?")
	// then
	if err != nil || len(matches) != 130 {
		t.Fatalf("matches=%d err=%v", len(matches), err)
	}
	other, err := scanDictionary(context.Background(), d, path, "!")
	if err != nil || len(other) != 0 {
		t.Fatalf("other=%d err=%v", len(other), err)
	}
	var synonyms bytes.Buffer
	synonyms.WriteString("hi\x00")
	_ = binary.Write(&synonyms, binary.BigEndian, uint32(999))
	if err := os.WriteFile(filepath.Join(dir, "dictionary.syn"), synonyms.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := scanDictionary(context.Background(), d, path, "hi"); err == nil {
		t.Fatal("invalid synonym target accepted")
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := scanDictionary(cancelled, d, path, "?"); err != context.Canceled {
		t.Fatalf("cancelled scan: %v", err)
	}
}
func TestGivenHTMLWithoutDefinitionListWhenExtractedThenReadableTextIsPreserved(t *testing.T) {
	// given
	// when
	got := htmlMeanings(`<p>A <a href="bword://greeting">greeting</a>.</p><script>ignore me</script>`)
	// then
	if len(got) != 1 || got[0].Definitions[0] != "A greeting." {
		t.Fatalf("got %#v", got)
	}
}

func TestGivenWordsWithCaseAccentsAndPunctuationWhenFoldedThenEquivalentFormsMatch(t *testing.T) {
	// given
	for _, tc := range []struct{ word, query, want string }{{"HELLO", "hello", "hello"}, {"cafÉ", "cafe", "cafe"}, {"look   up", "look up", "look up"}, {"A.I.", "ai", "ai"}, {"! x", "! x", " x"}, {"Straße", "strasse", "strasse"}} {
		// when
		key := foldWord(tc.query)
		// then
		if key != tc.want || !foldedMatch(tc.word, key) || foldWord(tc.word) != tc.want {
			t.Errorf("word=%q query=%q word_fold=%q query_fold=%q", tc.word, tc.query, foldWord(tc.word), key)
		}
	}
}

func TestGivenDifferentASCIIWordsWhenMatchedThenFastMatcherAgreesWithFolding(t *testing.T) {
	// given
	words := []string{"abc", "ABC", "a-b", "a b", "  A.B ", "! a", "a! b", "a\tb", "a_b", "a+b", ""}
	// when
	for _, word := range words {
		for _, query := range words {
			key := foldWord(query)
			// then
			if got, want := foldedMatch(word, key), foldWord(word) == key; got != want {
				t.Errorf("word=%q query=%q got=%v want=%v", word, query, got, want)
			}
		}
	}
}
