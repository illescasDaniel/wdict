package core

import (
	"context"
	"time"
)

// Lookup is the inbound, user-initiated query. Implementations return a status
// for expected lookup outcomes; errors are reserved for cancellation or failures
// that cannot be represented by a result.
type Lookup interface {
	Lookup(ctx context.Context, request LookupRequest) (LookupResult, error)
}

// Dictionaries is the inbound interface for listing and installing curated data.
// The progress callback may be nil. An installation failure must be reported and
// leave the previously installed dictionary intact; retry is a new invocation.
type Dictionaries interface {
	List(ctx context.Context) ([]Dictionary, error)
	Download(ctx context.Context, id DictionaryID, progress func(DownloadProgress)) error
}

// DictionarySource reports a missing entry as found=false, err=nil; unavailable,
// invalid or unreadable sources return an error. Entries retain attribution.
type DictionarySource interface {
	Find(ctx context.Context, query string, language Language) (entries []Entry, found bool, err error)
}

// Cache represents one tier. Get reports a miss as found=false, err=nil; errors
// must not prevent lookup. Implementations own validation, limits and eviction.
// Put must not damage an existing entry if persistence fails.
type Cache interface {
	Get(ctx context.Context, key CacheKey) (record CacheRecord, found bool, err error)
	Put(ctx context.Context, record CacheRecord) error
}

// DictionaryCatalog exposes only reviewed catalog records and package identities.
type DictionaryCatalog interface {
	List(ctx context.Context) ([]Dictionary, error)
	Find(ctx context.Context, id DictionaryID) (dictionary Dictionary, found bool, err error)
}

// DictionaryInstaller owns trusted package resolution, bounded downloading,
// integrity and archive validation, cleanup and atomic installation.
type DictionaryInstaller interface {
	Install(ctx context.Context, id DictionaryID, progress func(DownloadProgress)) error
}

// Clock permits deterministic freshness and last-used decisions.
type Clock interface {
	Now() time.Time
}
