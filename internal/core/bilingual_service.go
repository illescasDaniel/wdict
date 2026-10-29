package core

import (
	"context"
	"errors"
	"sync"
)

// BilingualService coordinates separate language lookups using shared caches.
type BilingualService struct {
	Sources      SourceResolver
	Memory, Disk Cache
	Clock        Clock
}

func (s *BilingualService) Lookup(ctx context.Context, request BilingualRequest, deliver func(LanguageOutcome)) error {
	var wg sync.WaitGroup
	for _, language := range []Language{English, Spanish} {
		wg.Add(1)
		go func(l Language) { defer wg.Done(); _ = s.RetryLanguage(ctx, request, l, deliver) }(language)
	}
	wg.Wait()
	return ctx.Err()
}
func (s *BilingualService) RetryLanguage(ctx context.Context, request BilingualRequest, language Language, deliver func(LanguageOutcome)) error {
	if language != English && language != Spanish {
		return errors.New("unsupported language")
	}
	// Validate before resolving any sources.
	probe := LookupService{}
	result, err := probe.validateOnly(request.Query)
	if err != nil {
		return err
	}
	if result.Status == InvalidQuery {
		if ctx.Err() == nil {
			deliver(LanguageOutcome{SearchID: request.ID, Query: request.Query, Language: language, Result: result})
		}
		return ctx.Err()
	}
	sources, err := s.Sources.Resolve(language)
	if err != nil || sources.Online == nil {
		result = LookupResult{Status: SourceUnavailable}
	} else {
		service := LookupService{Offline: sources.Offline, Online: sources.Online, Memory: s.Memory, Disk: s.Disk, Clock: s.Clock, Provider: sources.Provider}
		result, err = service.Lookup(ctx, LookupRequest{Query: request.Query, Language: language})
		if err != nil {
			return err
		}
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	deliver(LanguageOutcome{SearchID: request.ID, Query: request.Query, Language: language, Result: result})
	return nil
}
