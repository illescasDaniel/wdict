package adapters

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/illescasDaniel/wdict/internal/core"
)

// alternativesFixture creates real StarDict index and synonym records.
// Synonyms map into idx by zero-based record number.
func alternativesFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	ifo := filepath.Join(dir, "dictionary.ifo")
	var idx, syn bytes.Buffer
	for _, word := range []string{"café", "look forward to", "walk"} {
		idx.WriteString(word)
		idx.WriteByte(0)
		_ = binary.Write(&idx, binary.BigEndian, uint32(0))
		_ = binary.Write(&idx, binary.BigEndian, uint32(1))
	}
	for _, alias := range []struct {
		word   string
		target uint32
	}{
		{"cafés", 0}, {"walking", 2},
	} {
		syn.WriteString(alias.word)
		syn.WriteByte(0)
		_ = binary.Write(&syn, binary.BigEndian, alias.target)
	}
	data := fmt.Sprintf("StarDict's dict ifo file\nversion=3.0.0\nbookname=Fixture\nwordcount=3\nsynwordcount=2\nidxfilesize=%d\n", idx.Len())
	for name, content := range map[string][]byte{
		"dictionary.ifo": []byte(data), "dictionary.idx": idx.Bytes(), "dictionary.syn": syn.Bytes(), "dictionary.dict": []byte("x"),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), content, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return ifo
}

func TestGivenInstalledHeadwordIndexWhenIteratedThenWordsPhrasesAndSynonymsAreReturned(t *testing.T) {
	// given
	ifo := alternativesFixture(t)
	var index core.HeadwordIndex = &StarDictHeadwordIndex{IFO: ifo, Language: core.English}
	var got []string

	// when
	err := index.IterateHeadwords(context.Background(), core.English, func(word string) error {
		got = append(got, word)
		return nil
	})

	// then
	want := []string{"café", "look forward to", "walk", "cafés", "walking"}
	sort.Strings(got)
	sort.Strings(want)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("headwords=%q, want %q; err=%v", got, want, err)
	}
}

func TestGivenIndexWithSynonymsWhenCandidateVerifiedThenOriginalSpellingIsReturned(t *testing.T) {
	// given
	var index core.HeadwordIndex = &StarDictHeadwordIndex{IFO: alternativesFixture(t), Language: core.English}

	// when
	for _, tc := range []struct {
		query string
		want  string
		found bool
	}{{"café", "café", true}, {"walking", "walking", true}, {"look forward to", "look forward to", true}, {"nonexistent", "", false}} {
		got, found, err := index.VerifyHeadword(context.Background(), tc.query, core.English)
		// then
		if err != nil || got != tc.want || found != tc.found {
			t.Errorf("query=%q got=(%q,%v,%v) want=(%q,%v,nil)", tc.query, got, found, err, tc.want, tc.found)
		}
	}
}

func TestGivenTruncatedIndexWhenIteratedThenErrorIsReturned(t *testing.T) {
	// given
	ifo := alternativesFixture(t)
	if err := os.WriteFile(strings.TrimSuffix(ifo, ".ifo")+".idx", []byte("word\x00\x00"), 0600); err != nil {
		t.Fatal(err)
	}
	var index core.HeadwordIndex = &StarDictHeadwordIndex{IFO: ifo, Language: core.English}

	// when
	err := index.IterateHeadwords(context.Background(), core.English, func(string) error { return nil })

	// then
	if err == nil {
		t.Fatal("truncated index was accepted as complete")
	}
}

func TestGivenInvalidSynonymReferenceWhenIteratedThenErrorIsReturned(t *testing.T) {
	// given
	ifo := alternativesFixture(t)
	var syn bytes.Buffer
	syn.WriteString("alias\x00")
	_ = binary.Write(&syn, binary.BigEndian, uint32(999))
	if err := os.WriteFile(strings.TrimSuffix(ifo, ".ifo")+".syn", syn.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	var index core.HeadwordIndex = &StarDictHeadwordIndex{IFO: ifo, Language: core.English}

	// when
	err := index.IterateHeadwords(context.Background(), core.English, func(string) error { return nil })

	// then
	if err == nil {
		t.Fatal("synonym with out-of-range target was accepted")
	}
}

func TestGivenOversizedIndexWhenIteratedThenReadBudgetStopsScan(t *testing.T) {
	// given
	ifo := alternativesFixture(t)
	f, err := os.OpenFile(strings.TrimSuffix(ifo, ".ifo")+".idx", os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(128<<20 + 1); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	var index core.HeadwordIndex = &StarDictHeadwordIndex{IFO: ifo, Language: core.English}

	// when
	err = index.IterateHeadwords(context.Background(), core.English, func(string) error { return nil })

	// then
	if err == nil {
		t.Fatal("index beyond the 128 MiB read budget was accepted")
	}
}

func TestGivenCancelledScanWhenIteratedThenCancellationIsReturned(t *testing.T) {
	// given
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var index core.HeadwordIndex = &StarDictHeadwordIndex{IFO: alternativesFixture(t), Language: core.English}
	calls := 0

	// when
	err := index.IterateHeadwords(ctx, core.English, func(string) error { calls++; return nil })

	// then
	if !errors.Is(err, context.Canceled) || calls != 0 {
		t.Fatalf("err=%v callbacks=%d", err, calls)
	}
}

func TestGivenCallbackFailureWhenIteratedThenErrorStopsTheScan(t *testing.T) {
	// given
	var index core.HeadwordIndex = &StarDictHeadwordIndex{IFO: alternativesFixture(t), Language: core.English}
	stop := errors.New("stop")
	calls := 0

	// when
	err := index.IterateHeadwords(context.Background(), core.English, func(string) error {
		calls++
		return stop
	})

	// then
	if !errors.Is(err, stop) || calls != 1 {
		t.Fatalf("err=%v callbacks=%d", err, calls)
	}
}

// TestGivenIndexRecordAtBufferBoundaryWhenIteratedThenHeadwordIsPreserved
// checks that reading the fixed payload cannot overwrite the previous name.
func TestGivenIndexRecordAtBufferBoundaryWhenIteratedThenHeadwordIsPreserved(t *testing.T) {
	// given
	ifo := alternativesFixture(t)
	word := strings.Repeat("a", alternativesMaxRecordBytes-2)
	var records bytes.Buffer
	records.WriteString(word)
	records.WriteByte(0)
	_ = binary.Write(&records, binary.BigEndian, uint64(1))
	records.WriteString("other")
	records.WriteByte(0)
	_ = binary.Write(&records, binary.BigEndian, uint64(1))
	if err := os.WriteFile(strings.TrimSuffix(ifo, ".ifo")+".idx", records.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(strings.TrimSuffix(ifo, ".ifo")+".syn", nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ifo, []byte("StarDict's dict ifo file\nversion=3.0.0\nbookname=Fixture\nwordcount=2\n"), 0600); err != nil {
		t.Fatal(err)
	}
	index := &StarDictHeadwordIndex{IFO: ifo, Language: core.English}
	var got []string

	// when
	err := index.IterateHeadwords(context.Background(), core.English, func(s string) error {
		got = append(got, s)
		return nil
	})

	// then
	if err != nil || !reflect.DeepEqual(got, []string{word, "other"}) {
		t.Fatalf("headwords corrupted at buffer refill: %q, err=%v", got, err)
	}
}

func TestGivenWholeRecordTruncationWhenIteratedThenErrorIsReturned(t *testing.T) {
	for _, tc := range []struct {
		name      string
		extension string
		content   []byte
	}{
		{name: "empty index", extension: ".idx", content: nil},
		{name: "one complete index record", extension: ".idx", content: append([]byte("café\x00"), make([]byte, 8)...)},
		{name: "empty synonym index", extension: ".syn", content: nil},
		{name: "one complete synonym record", extension: ".syn", content: append([]byte("cafés\x00"), make([]byte, 4)...)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// given
			ifo := alternativesFixture(t)
			base := strings.TrimSuffix(ifo, ".ifo")
			if err := os.WriteFile(base+tc.extension, tc.content, 0600); err != nil {
				t.Fatal(err)
			}
			if tc.extension == ".idx" {
				if err := os.WriteFile(base+".syn", nil, 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(ifo, []byte("StarDict\x27s dict ifo file\nversion=3.0.0\nwordcount=3\nsynwordcount=0\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			index := &StarDictHeadwordIndex{IFO: ifo, Language: core.English}

			// when
			err := index.IterateHeadwords(context.Background(), core.English, func(string) error { return nil })

			// then
			if err == nil {
				t.Fatal("whole-record truncation was accepted as a complete scan")
			}
		})
	}
}

func TestGivenMissingOrCorruptDefinitionsWhenIteratedThenUnavailable(t *testing.T) {
	for _, corrupt := range []bool{false, true} {
		t.Run(fmt.Sprintf("corrupt_dictzip=%v", corrupt), func(t *testing.T) {
			// given
			ifo := alternativesFixture(t)
			base := strings.TrimSuffix(ifo, ".ifo")
			if err := os.Rename(base+".dict", base+".saved"); err != nil {
				t.Fatal(err)
			}
			if corrupt {
				if err := os.WriteFile(base+".dict.dz", []byte("not dictzip"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			index := &StarDictHeadwordIndex{IFO: ifo, Language: core.English}
			calls := 0

			// when
			err := index.IterateHeadwords(context.Background(), core.English, func(string) error { calls++; return nil })

			// then
			if err == nil || calls != 0 {
				t.Fatalf("err=%v calls=%d", err, calls)
			}
		})
	}
}
