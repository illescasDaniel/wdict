package core

import (
	"context"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// LookupService coordinates local data, the two response-cache tiers and online lookup.
type LookupService struct {
	Offline, Online DictionarySource
	Memory, Disk    Cache
	Clock           Clock
	Provider        ProviderID
}

func (s *LookupService) Lookup(ctx context.Context, request LookupRequest) (LookupResult, error) {
	query := strings.TrimSpace(request.Query)
	if query == "" || !utf8.ValidString(request.Query) || utf8.RuneCountInString(query) > 256 || strings.IndexFunc(request.Query, unicode.IsControl) >= 0 {
		return LookupResult{Status: InvalidQuery}, nil
	}
	if err := ctx.Err(); err != nil {
		return LookupResult{}, err
	}
	if s.Offline != nil {
		entries, found, _ := s.Offline.Find(ctx, query, request.Language)
		if err := ctx.Err(); err != nil {
			return LookupResult{}, err
		}
		if found && len(entries) > 0 {
			return LookupResult{Status: Found, Entries: entries, Origin: OriginOffline}, nil
		}
	}
	key := CacheKey{Provider: s.Provider, Language: request.Language, NormalizedQuery: norm.NFC.String(cases.Fold().String(norm.NFC.String(query)))}
	now := s.Clock.Now()
	var stale *CacheRecord
	var staleOrigin ResultOrigin
	for _, tier := range []struct {
		cache  Cache
		origin ResultOrigin
	}{{s.Memory, OriginMemory}, {s.Disk, OriginDisk}} {
		if tier.cache == nil {
			continue
		}
		rec, found, err := tier.cache.Get(ctx, key)
		if err != nil || !found || len(rec.Entries) == 0 || rec.FetchedAt.IsZero() {
			continue
		}
		if now.Sub(rec.FetchedAt) < 7*24*time.Hour && !rec.FetchedAt.After(now) {
			rec.LastUsed = now
			_ = tier.cache.Put(ctx, rec)
			if tier.origin == OriginDisk && s.Memory != nil {
				_ = s.Memory.Put(ctx, rec)
			}
			return LookupResult{Status: Found, Entries: rec.Entries, Origin: tier.origin, FetchedAt: rec.FetchedAt}, nil
		}
		if !rec.FetchedAt.After(now) && (stale == nil || rec.FetchedAt.After(stale.FetchedAt)) {
			copy := rec
			stale = &copy
			staleOrigin = tier.origin
		}
	}
	entries, found, err := s.Online.Find(ctx, query, request.Language)
	if ctx.Err() != nil {
		return LookupResult{}, ctx.Err()
	}
	if err != nil || (found && len(entries) == 0) {
		if stale != nil {
			return LookupResult{Status: Found, Entries: stale.Entries, Origin: staleOrigin, FetchedAt: stale.FetchedAt, Stale: true}, nil
		}
		return LookupResult{Status: SourceUnavailable}, nil
	}
	if !found {
		return LookupResult{Status: NotFound}, nil
	}
	rec := CacheRecord{Key: key, Entries: entries, FetchedAt: now, LastUsed: now}
	if s.Memory != nil {
		_ = s.Memory.Put(ctx, rec)
	}
	if s.Disk != nil {
		_ = s.Disk.Put(ctx, rec)
	}
	return LookupResult{Status: Found, Entries: entries, Origin: OriginOnline, FetchedAt: now}, nil
}
