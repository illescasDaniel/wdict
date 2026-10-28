package adapters

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"

	stardict "github.com/ianlewis/go-stardict"
	"github.com/ianlewis/go-stardict/dict"
	"github.com/ianlewis/go-stardict/idx"
	"github.com/ianlewis/go-stardict/syn"
	"github.com/illescasDaniel/wdict/internal/core"
	"golang.org/x/net/html"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// OfflineDictionary opens the dictionary lazily and scans its compact indexes.
type OfflineDictionary struct {
	IFO        string
	Source     core.Source
	mu         sync.Mutex
	dictionary *stardict.Stardict
	cache      map[string][]core.Entry
	cacheOrder []string
}

func (d *OfflineDictionary) Find(ctx context.Context, query string, language core.Language) ([]core.Entry, bool, error) {
	if language != "en" {
		return nil, false, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	if _, err := os.Stat(d.IFO); os.IsNotExist(err) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	key := foldWord(query)
	if key == "" {
		key = "\x00" + query
	}
	if result, ok := d.cache[key]; ok {
		d.remember(key, result)
		return result, true, nil
	}
	if d.dictionary == nil {
		var err error
		d.dictionary, err = stardict.Open(filepath.Clean(d.IFO), nil)
		if err != nil {
			return nil, false, err
		}
	}
	matches, err := scanDictionary(ctx, d.dictionary, d.IFO, query)
	if err != nil {
		return nil, false, err
	}
	dictionary, err := d.dictionary.Dict()
	if err != nil {
		return nil, false, err
	}
	var entries []core.Entry
	for _, m := range matches {
		if err := ctx.Err(); err != nil {
			return nil, false, err
		}
		var meanings []core.Meaning
		word, err := dictionary.Word(m)
		if err != nil {
			return nil, false, err
		}
		for _, data := range word.Data {
			if data.Type == dict.HTMLType {
				meanings = append(meanings, htmlMeanings(string(data.Data))...)
			} else if text := strings.TrimSpace(data.String()); text != "" {
				meanings = append(meanings, core.Meaning{Definitions: []string{text}})
			}
		}
		if len(meanings) > 0 {
			source := d.Source
			if source.ID == dictionaryID {
				source.URL = "https://en.wiktionary.org/wiki/" + url.PathEscape(strings.ReplaceAll(m.Word, " ", "_"))
			}
			entries = append(entries, core.Entry{Headword: m.Word, Language: language, Meanings: meanings, Source: source})
		}
	}
	if len(entries) == 0 {
		return nil, false, nil
	}
	d.remember(key, entries)
	return entries, true, nil
}

func (d *OfflineDictionary) remember(key string, entries []core.Entry) {
	if d.cache == nil {
		d.cache = make(map[string][]core.Entry)
	}
	for i, k := range d.cacheOrder {
		if k == key {
			d.cacheOrder = append(d.cacheOrder[:i], d.cacheOrder[i+1:]...)
			break
		}
	}
	d.cache[key] = entries
	d.cacheOrder = append(d.cacheOrder, key)
	if len(d.cacheOrder) > 128 {
		delete(d.cache, d.cacheOrder[0])
		d.cacheOrder = d.cacheOrder[1:]
	}
}

// foldWord mirrors the default StarDict case, whitespace, accent and punctuation folding.
func foldWord(word string) string {
	if ascii(word) {
		var b strings.Builder
		space := false
		started := false
		for _, c := range word {
			if unicode.IsSpace(c) {
				if started {
					space = true
				}
				continue
			}
			if space {
				b.WriteByte(' ')
				space = false
			}
			started = true
			if unicode.IsPunct(c) {
				continue
			}
			if c >= 'A' && c <= 'Z' {
				c += 'a' - 'A'
			}
			b.WriteRune(c)
		}
		return b.String()
	}
	word = cases.Fold().String(norm.NFD.String(word))
	word = strings.Join(strings.Fields(word), " ")
	word = strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) || unicode.IsPunct(r) {
			return -1
		}
		return r
	}, word)
	return norm.NFC.String(word)
}

func ascii(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 128 {
			return false
		}
	}
	return true
}

func foldedMatch(word, key string) bool {
	if !ascii(word) {
		return foldWord(word) == key
	}
	i := 0
	started, space := false, false
	for j := 0; j < len(word); j++ {
		c := word[j]
		if c == ' ' || (c >= '\t' && c <= '\r') {
			if started {
				space = true
			}
			continue
		}
		if space {
			if i >= len(key) || key[i] != ' ' {
				return false
			}
			i++
			space = false
		}
		started = true
		if unicode.IsPunct(rune(c)) {
			continue
		}
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if i >= len(key) || key[i] != c {
			return false
		}
		i++
	}
	return i == len(key)
}

// scanDictionary reads the original StarDict indexes without constructing a
// multi-million-entry in-memory search index. Alias positions refer to .idx.
func scanDictionary(ctx context.Context, d *stardict.Stardict, path, query string) ([]*idx.Word, error) {
	key := foldWord(query)
	matches := func(word string) bool {
		if key == "" {
			return word == query
		}
		return foldedMatch(word, key)
	}
	aliases := make(map[uint32]int)
	file, err := syn.Open(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if err == nil {
		s, err := syn.NewScanner(file)
		if err != nil {
			file.Close()
			return nil, err
		}
		for count := 0; s.Scan(); count++ {
			if count%8192 == 0 {
				if err := ctx.Err(); err != nil {
					s.Close()
					return nil, err
				}
			}
			w := s.Word()
			if matches(w.Word) {
				aliases[w.OriginalWordIndex]++
			}
		}
		err = s.Err()
		closeErr := s.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			return nil, closeErr
		}
	}
	s, err := d.IndexScanner()
	if err != nil {
		return nil, err
	}
	defer s.Close()
	var results []*idx.Word
	for i := uint32(0); s.Scan(); i++ {
		if i%8192 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
		}
		w := s.Word()
		if matches(w.Word) {
			results = append(results, w)
		}
		for n := aliases[i]; n > 0; n-- {
			results = append(results, w)
		}
		delete(aliases, i)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if len(aliases) != 0 {
		return nil, errors.New("invalid synonym index")
	}
	return results, nil
}

// htmlMeanings extracts the entry's ordered definition list, without exposing
// link destinations or treating dictionary markup as executable content.
func htmlMeanings(raw string) []core.Meaning {
	doc, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		return nil
	}
	var meanings []core.Meaning
	var heading string
	var visit func(*html.Node, bool)
	visit = func(n *html.Node, nestedSection bool) {
		if n.Type == html.ElementNode && n.Data == "section" {
			if nestedSection {
				return
			}
			nestedSection = true
		}
		if n.Type == html.ElementNode && (n.Data == "h3" || n.Data == "h4") {
			heading = nodeText(n)
		}
		if n.Type == html.ElementNode && n.Data == "ol" {
			var defs []string
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				if child.Type == html.ElementNode && child.Data == "li" {
					if text := nodeText(child); text != "" {
						defs = append(defs, text)
					}
				}
			}
			if len(defs) > 0 {
				meanings = append(meanings, core.Meaning{PartOfSpeech: heading, Definitions: defs})
			}
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c, nestedSection)
		}
	}
	visit(doc, false)
	if len(meanings) == 0 {
		if text := htmlFallbackText(doc); text != "" {
			meanings = append(meanings, core.Meaning{Definitions: []string{text}})
		}
	}
	return meanings
}

func htmlFallbackText(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
			return
		}
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
		if n.Type == html.ElementNode && (n.Data == "p" || n.Data == "li" || n.Data == "br") {
			b.WriteString(" ")
		}
	}
	walk(n)
	return strings.Join(strings.Fields(b.String()), " ")
}

func nodeText(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style" || n.Data == "dl" || n.Data == "ol" || n.Data == "ul") {
			return
		}
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(b.String()), " ")
}
