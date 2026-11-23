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

	"github.com/illescasDaniel/wdict/internal/core"
	"golang.org/x/net/html"
)

// FrenchOnlineDictionary reads only the French section of French Wiktionary.
type FrenchOnlineDictionary struct{ Client *http.Client }

func (p FrenchOnlineDictionary) Find(ctx context.Context, query string, language core.Language) ([]core.Entry, bool, error) {
	if language != core.French {
		return nil, false, errors.New("unsupported language")
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	params := url.Values{"action": {"parse"}, "page": {query}, "prop": {"text"}, "format": {"json"}, "formatversion": {"2"}, "redirects": {"1"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://fr.wiktionary.org/w/api.php?"+params.Encode(), nil)
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
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024+1))
	if err != nil || len(body) > 2*1024*1024 {
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
	if json.Unmarshal(body, &data) != nil {
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
	meanings, pronunciations, sections, headings := frenchContent(root, false)
	if sections > 1 {
		return nil, false, errors.New("ambiguous French sections")
	}
	if sections == 0 {
		if headings == 0 {
			return nil, false, errors.New("unrecognized language sections")
		}
		return nil, false, nil
	}
	if len(meanings) == 0 {
		return nil, false, errors.New("unrecognized French entry")
	}
	source := core.Source{ID: "french-wiktionary", Name: "Wiktionnaire", URL: "https://fr.wiktionary.org/wiki/" + url.PathEscape(strings.ReplaceAll(data.Parse.Title, " ", "_")), Attribution: "Wiktionary contributors; adapted by wdict", LicenseURL: "https://creativecommons.org/licenses/by-sa/4.0/"}
	return []core.Entry{{Headword: data.Parse.Title, Language: language, Pronunciations: core.NormalizePronunciations(pronunciations), Meanings: meanings, Source: source}}, true, nil
}

// frenchContent reads only definition lists introduced by a titredef heading;
// nested examples and non-content elements cannot become definitions.
func frenchContent(root *html.Node, offline bool) ([]core.Meaning, []core.Pronunciation, int, int) {
	active := offline
	part := ""
	pronunciationSection := false
	sections, headings := 0, 0
	var meanings []core.Meaning
	var pronunciations []core.Pronunciation
	hasClass := func(n *html.Node, name string) bool {
		for _, a := range n.Attr {
			if a.Key == "class" {
				for _, c := range strings.Fields(a.Val) {
					if c == name {
						return true
					}
				}
			}
		}
		return false
	}
	var title func(*html.Node) string
	title = func(n *html.Node) string {
		if hasClass(n, "titredef") {
			return strings.TrimSpace(spanishNodeText(n))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if v := title(c); v != "" {
				return v
			}
		}
		return ""
	}
	var plain func(*html.Node, *strings.Builder)
	plain = func(n *html.Node, b *strings.Builder) {
		if n.Type == html.ElementNode && (n.Data == "ul" || n.Data == "ol" || n.Data == "sup" || isNonContentElement(n)) {
			return
		}
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			plain(c, b)
		}
	}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "h2" && !offline {
				active = false
				part = ""
				pronunciationSection = false
				for _, a := range n.Attr {
					if a.Key == "id" && a.Val == "Français" {
						active = true
						sections++
					}
				}
				var languageMarker func(*html.Node) bool
				languageMarker = func(node *html.Node) bool {
					if hasClass(node, "sectionlangue") {
						return true
					}
					for child := node.FirstChild; child != nil; child = child.NextSibling {
						if languageMarker(child) {
							return true
						}
					}
					return false
				}
				if languageMarker(n) {
					headings++
				} else if active {
					sections++ // known heading without language marker is ambiguous
				}
				return
			}
			if active && ((!offline && n.Data == "h3") || (offline && n.Data == "h4")) {
				part = title(n)
				pronunciationSection = false
				var hasPronTitle func(*html.Node) bool
				hasPronTitle = func(node *html.Node) bool {
					if hasClass(node, "titrepron") {
						return true
					}
					for child := node.FirstChild; child != nil; child = child.NextSibling {
						if hasPronTitle(child) {
							return true
						}
					}
					return false
				}
				pronunciationSection = hasPronTitle(n)
				return
			}
			if active && n.Data == "span" && hasClass(n, "API") {
				relevant := part != ""
				if pronunciationSection {
					for ancestor := n.Parent; ancestor != nil; ancestor = ancestor.Parent {
						if hasClass(ancestor, "audio-ipa") {
							relevant = true
							break
						}
						if ancestor.Type == html.ElementNode && ancestor.Data == "a" {
							for _, a := range ancestor.Attr {
								if a.Key == "title" && a.Val == "Annexe:Prononciation/français" {
									relevant = true
									break
								}
							}
						}
						if ancestor.Type == html.ElementNode && ancestor.Data == "li" {
							break
						}
					}
				}
				if !relevant {
					return
				}
				v := strings.TrimSpace(spanishNodeText(n))
				v = strings.Trim(v, "\\/[] ")
				if v != "" {
					pronunciations = append(pronunciations, core.Pronunciation{Text: "/" + v + "/"})
				}
				return
			}
			if active && part != "" && n.Data == "ol" {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode && c.Data == "li" {
						var b strings.Builder
						for child := c.FirstChild; child != nil; child = child.NextSibling {
							plain(child, &b)
						}
						if v := strings.Join(strings.Fields(b.String()), " "); v != "" {
							meanings = append(meanings, core.Meaning{PartOfSpeech: part, Definitions: []string{v}})
						}
					}
				}
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	return meanings, pronunciations, sections, headings
}
