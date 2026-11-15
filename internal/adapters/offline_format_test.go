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
func TestGivenRepeatedSynonymAliasesWhenScannedThenEachIndexEntryIsReturnedOnce(t *testing.T) {
	// given
	dir := t.TempDir()
	path := filepath.Join(dir, "dictionary.ifo")
	words := []string{"dog food", "dog food", "dogfood"}
	var index bytes.Buffer
	for offset, word := range words {
		index.WriteString(word + "\x00")
		_ = binary.Write(&index, binary.BigEndian, uint32(offset))
		_ = binary.Write(&index, binary.BigEndian, uint32(1))
	}
	// The real English dictionary repeats aliases and lists an alias that folds
	// to the target word itself ("dog-food" → "dogfood"), which duplicated results.
	var synonyms bytes.Buffer
	for _, alias := range []struct {
		word   string
		target uint32
	}{{"dog-food", 2}, {"dog-food", 0}, {"dogfood", 0}, {"dogfood", 1}, {"dogfood", 2}} {
		synonyms.WriteString(alias.word + "\x00")
		_ = binary.Write(&synonyms, binary.BigEndian, alias.target)
	}
	ifo := fmt.Sprintf("StarDict's dict ifo file\nversion=3.0.0\nbookname=Fixture\nwordcount=3\nsynwordcount=5\nidxfilesize=%d\n", index.Len())
	if err := os.WriteFile(path, []byte(ifo), 0600); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string][]byte{
		"dictionary.idx": index.Bytes(), "dictionary.syn": synonyms.Bytes(),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), content, 0600); err != nil {
			t.Fatal(err)
		}
	}
	d, err := stardict.Open(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	// when
	matches, err := scanDictionary(context.Background(), d, path, "dogfood")
	// then
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, m := range matches {
		got = append(got, m.Word)
	}
	// The exact headword ranks first since the tiering fix; the duplicate
	// "dog food" entries follow, still reported once each (c8977f6).
	want := []string{"dogfood", "dog food", "dog food"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("headwords=%q, want %q", got, want)
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

// synAlias is one .syn record pointing at a zero-based .idx position.
type synAlias struct {
	word   string
	target uint32
}

// rankingFixture writes a StarDict index with the given headwords in raw
// order, plus optional synonym redirects, and returns the .ifo path.
func rankingFixture(t *testing.T, words []string, aliases []synAlias) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "dictionary.ifo")
	var index, synonyms bytes.Buffer
	for offset, word := range words {
		index.WriteString(word + "\x00")
		_ = binary.Write(&index, binary.BigEndian, uint32(offset))
		_ = binary.Write(&index, binary.BigEndian, uint32(1))
	}
	for _, alias := range aliases {
		synonyms.WriteString(alias.word + "\x00")
		_ = binary.Write(&synonyms, binary.BigEndian, alias.target)
	}
	ifo := fmt.Sprintf("StarDict's dict ifo file\nversion=3.0.0\nbookname=Fixture\nwordcount=%d\nsynwordcount=%d\nidxfilesize=%d\n", len(words), len(aliases), index.Len())
	if err := os.WriteFile(path, []byte(ifo), 0600); err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{"dictionary.idx": index.Bytes()}
	if len(aliases) > 0 {
		files["dictionary.syn"] = synonyms.Bytes()
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), content, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

// TestGivenMixedMatchKindsWhenScannedThenExactHeadwordRanksFirst checks that
// exact headword cards precede folded matches, which precede alias redirects,
// with the raw index order preserved inside each tier: the app shows the
// pronunciation line of the first card only (año/thought/sing reports).
func TestGivenMixedMatchKindsWhenScannedThenExactHeadwordRanksFirst(t *testing.T) {
	for _, tc := range []struct {
		name    string
		words   []string
		aliases []synAlias
		query   string
		want    []string
	}{
		{
			// The raw order interleaves every tier: "año" and "Año" are exact
			// matches (case only), while "ano", "anó" and the "-ano" suffix
			// card only match once diacritics and punctuation are folded, so
			// they must not precede the queried headword's [ˈaɲo] card.
			name:  "exact beats folded",
			words: []string{"ano", "-ano", "año", "anó", "Año"},
			query: "año",
			want:  []string{"año", "Año", "ano", "-ano", "anó"},
		},
		{
			// "thought" sits last in the raw index but must come first; the
			// think and thinking cards follow through the .syn redirect, and
			// the alias that points at the queried word itself does not
			// duplicate it.
			name:    "exact beats alias",
			words:   []string{"think", "thinking", "thought"},
			aliases: []synAlias{{"thought", 0}, {"thought", 1}, {"thought", 2}},
			query:   "thought",
			want:    []string{"thought", "think", "thinking"},
		},
		{
			// The censored aliases "s****ing" and "s—ing" fold to "sing" and
			// redirect to the "shit" card, which must not outrank the queried
			// headword.
			name:    "folded alias redirects rank last",
			words:   []string{"shit", "sing"},
			aliases: []synAlias{{"s****ing", 0}, {"s—ing", 0}},
			query:   "sing",
			want:    []string{"sing", "shit"},
		},
		{
			// An alias redirect still returns results when nothing matches
			// exactly or after folding (the quiso → querer shape).
			name:    "alias-only match",
			words:   []string{"querer"},
			aliases: []synAlias{{"quiso", 0}},
			query:   "quiso",
			want:    []string{"querer"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// given
			path := rankingFixture(t, tc.words, tc.aliases)
			d, err := stardict.Open(path, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer d.Close()
			// when
			matches, err := scanDictionary(context.Background(), d, path, tc.query)
			// then
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, m := range matches {
				got = append(got, m.Word)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("headwords=%q, want %q", got, tc.want)
			}
		})
	}
}
