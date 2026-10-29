package core

import "context"

// Supported lookup languages. Both are searched for a new user query.
const (
	English Language = "en"
	Spanish Language = "es"
)

// SearchID identifies one user search, including its per-language retries.
// The caller assigns a new ID for every new query and ignores older outcomes.
type SearchID string

// BilingualRequest starts a search in both supported languages. A retry is a
// separate RetryLanguage call and retains the original search ID and query.
type BilingualRequest struct {
	ID    SearchID
	Query string
}

// LanguageOutcome associates a completed per-language lookup with its search.
// Expected failures are represented by Result.Status, not by an error.
type LanguageOutcome struct {
	SearchID SearchID
	Query    string
	Language Language
	Result   LookupResult
}

// BilingualLookup delivers outcomes independently, without waiting for the
// other language. Callbacks may run concurrently and must be concurrency-safe.
// Lookup waits for both attempts to finish; cancellation stops pending work.
// An invalid query emits InvalidQuery outcomes without consulting any source.
// RetryLanguage repeats only the requested supported language for the same
// search, preserving the other outcome. Unsupported languages return an error.
// Callers cancel or ignore callbacks after a search is superseded; the ID also
// allows transport adapters to discard late events.
type BilingualLookup interface {
	Lookup(ctx context.Context, request BilingualRequest, deliver func(LanguageOutcome)) error
	RetryLanguage(ctx context.Context, request BilingualRequest, language Language, deliver func(LanguageOutcome)) error
}

// LanguageSources identifies the providers for one language. Provider keys
// the online cache and must remain stable for existing English cache entries.
// A nil Offline means no installed dictionary; Online is required.
type LanguageSources struct {
	Offline  DictionarySource
	Online   DictionarySource
	Provider ProviderID
}

// SourceResolver selects configured sources and cache identity for a supported
// language. It never derives endpoints or paths from a query or frontend input.
type SourceResolver interface {
	Resolve(language Language) (sources LanguageSources, err error)
}
