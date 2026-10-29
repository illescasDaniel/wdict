package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/illescasDaniel/wdict/internal/core"
	"golang.org/x/net/html"
)

// SpanishOnlineDictionary extracts definitions from the Spanish Wiktionary API.
type SpanishOnlineDictionary struct{ Client *http.Client }

func (p SpanishOnlineDictionary) Find(ctx context.Context, query string, language core.Language) ([]core.Entry, bool, error) {
	if language != core.Spanish {
		return nil, false, errors.New("unsupported language")
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	params := url.Values{"action": {"parse"}, "page": {query}, "prop": {"text"}, "format": {"json"}, "formatversion": {"2"}, "redirects": {"1"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://es.wiktionary.org/w/api.php?"+params.Encode(), nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", "wdict/0.4 (https://github.com/illescasDaniel/wdict)")
	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusBadGateway {
		return nil, false, core.ErrSilentUnavailable
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("provider status %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024+1))
	if err != nil || len(b) > 2*1024*1024 {
		return nil, false, errors.New("invalid provider body")
	}
	var data struct {
		Parse struct {
			Title string `json:"title"`
			Text  string `json:"text"`
		} `json:"parse"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if json.Unmarshal(b, &data) != nil {
		return nil, false, errors.New("invalid provider response")
	}
	if data.Error.Code == "missingtitle" {
		return nil, false, nil
	}
	if data.Error.Code != "" || data.Parse.Title == "" || data.Parse.Text == "" {
		return nil, false, errors.New("invalid provider response")
	}
	root, err := html.Parse(strings.NewReader(data.Parse.Text))
	if err != nil {
		return nil, false, err
	}
	meanings, sections := spanishMeanings(root, false)
	if sections > 1 {
		return nil, false, errors.New("ambiguous Spanish sections")
	}
	if sections == 0 {
		if hasSpanishMarker(root) || !hasLanguageHeading(root) {
			return nil, false, errors.New("unrecognized language sections")
		}
		return nil, false, nil
	}
	if len(meanings) == 0 {
		return nil, false, errors.New("unrecognized Spanish entry")
	}
	source := core.Source{ID: "spanish-wiktionary", Name: "Wikcionario", URL: "https://es.wiktionary.org/wiki/" + url.PathEscape(strings.ReplaceAll(data.Parse.Title, " ", "_")), Attribution: "Wiktionary contributors; definitions extracted and formatting changed by wdict", LicenseURL: "https://creativecommons.org/licenses/by-sa/4.0/"}
	return []core.Entry{{Headword: data.Parse.Title, Language: language, Meanings: meanings, Source: source}}, true, nil
}

// spanishMeanings extracts numbered definition lists from Spanish source HTML.
func spanishMeanings(root *html.Node, offline bool) ([]core.Meaning, int) {
	active, numbered := offline, false
	sections := 0
	part := ""
	var meanings []core.Meaning
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "h2" {
				active = false
				for _, a := range n.Attr {
					if a.Key == "id" && a.Val == "Español" {
						active = true
						sections++
					}
				}
			}
			if active && n.Data == "h4" {
				part = strings.TrimSpace(spanishNodeText(n))
				numbered = false
			}
			if active && n.Data == "dt" {
				t := strings.TrimSpace(spanishNodeText(n))
				numbered = len(t) > 0 && unicode.IsDigit([]rune(t)[0])
			}
			if active && n.Data == "dd" && numbered {
				var b strings.Builder
				var plain func(*html.Node)
				plain = func(x *html.Node) {
					if x.Type == html.ElementNode && (x.Data == "ul" || x.Data == "ol" || x.Data == "sup") {
						return
					}
					if x.Type == html.TextNode {
						b.WriteString(x.Data)
					}
					for c := x.FirstChild; c != nil; c = c.NextSibling {
						plain(c)
					}
				}
				plain(n)
				v := strings.TrimSpace(b.String())
				if v != "" {
					meanings = append(meanings, core.Meaning{PartOfSpeech: part, Definitions: []string{v}})
				}
				numbered = false
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}
	visit(root)
	return meanings, sections
}

func spanishNodeText(n *html.Node) string {
	var b strings.Builder
	var f func(*html.Node)
	f = func(x *html.Node) {
		if x.Type == html.TextNode {
			b.WriteString(x.Data)
		}
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return b.String()
}

// hasLanguageHeading recognizes MediaWiki language sections without treating
// arbitrary or changed page markup as proof that Spanish is absent.
func hasLanguageHeading(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "h2" {
		var containsLanguage func(*html.Node) bool
		containsLanguage = func(x *html.Node) bool {
			for _, a := range x.Attr {
				if a.Key == "class" {
					for _, c := range strings.Fields(a.Val) {
						if c == "headline-lang" {
							return true
						}
					}
				}
			}
			for c := x.FirstChild; c != nil; c = c.NextSibling {
				if containsLanguage(c) {
					return true
				}
			}
			return false
		}
		if containsLanguage(n) {
			return true
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if hasLanguageHeading(c) {
			return true
		}
	}
	return false
}

// A Spanish marker under an unexpected heading is an unsupported structure.
func hasSpanishMarker(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "span" {
		var id, class string
		for _, a := range n.Attr {
			if a.Key == "id" {
				id = a.Val
			}
			if a.Key == "class" {
				class = a.Val
			}
		}
		if id == "es" {
			for _, c := range strings.Fields(class) {
				if c == "headline-lang" {
					return true
				}
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if hasSpanishMarker(c) {
			return true
		}
	}
	return false
}
