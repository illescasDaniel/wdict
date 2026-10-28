package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/illescasDaniel/wdict/internal/core"
)

// OnlineDictionary accesses the fixed English Free Dictionary API endpoint.
type OnlineDictionary struct{ Client *http.Client }

func (p OnlineDictionary) Find(ctx context.Context, query string, language core.Language) ([]core.Entry, bool, error) {
	if language != "en" {
		return nil, false, fmt.Errorf("unsupported language")
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.dictionaryapi.dev/api/v2/entries/en/"+url.PathEscape(query), nil)
	if err != nil {
		return nil, false, err
	}
	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, false, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return nil, false, nil
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, false, fmt.Errorf("provider status %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024+1))
	if err != nil || len(data) > 2*1024*1024 {
		return nil, false, fmt.Errorf("invalid provider body")
	}
	var items []struct {
		Word       string   `json:"word"`
		SourceURLs []string `json:"sourceUrls"`
		License    struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"license"`
		Meanings []struct {
			PartOfSpeech string `json:"partOfSpeech"`
			Definitions  []struct {
				Definition string `json:"definition"`
			} `json:"definitions"`
		} `json:"meanings"`
	}
	if json.Unmarshal(data, &items) != nil || len(items) == 0 {
		return nil, false, fmt.Errorf("invalid provider response")
	}
	entries := make([]core.Entry, 0, len(items))
	for _, item := range items {
		entry := core.Entry{Headword: item.Word, Language: language, Source: core.Source{ID: "free-dictionary-api", Name: "Free Dictionary API", Attribution: item.License.Name, LicenseURL: item.License.URL}}
		if len(item.SourceURLs) > 0 {
			entry.Source.URL = item.SourceURLs[0]
			entry.Source.AdditionalURLs = append([]string(nil), item.SourceURLs[1:]...)
		}
		for _, m := range item.Meanings {
			meaning := core.Meaning{PartOfSpeech: m.PartOfSpeech}
			for _, d := range m.Definitions {
				if strings.TrimSpace(d.Definition) != "" {
					meaning.Definitions = append(meaning.Definitions, d.Definition)
				}
			}
			if len(meaning.Definitions) > 0 {
				entry.Meanings = append(entry.Meanings, meaning)
			}
		}
		if strings.TrimSpace(entry.Headword) != "" && len(entry.Meanings) > 0 {
			entries = append(entries, entry)
		}
	}
	if len(entries) == 0 {
		return nil, false, fmt.Errorf("provider response has no definitions")
	}
	return entries, true, nil
}
