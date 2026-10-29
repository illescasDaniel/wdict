// Package core defines the dictionary application's framework-independent contracts.
package core

import "time"

type Language string
type DictionaryID string
type ProviderID string

type Source struct {
	ID             string
	Name           string
	URL            string
	Attribution    string
	LicenseURL     string
	AdditionalURLs []string
}

type Meaning struct {
	PartOfSpeech string
	Definitions  []string
}

type Entry struct {
	Headword string
	Language Language
	Meanings []Meaning
	Source   Source
}

type LookupRequest struct {
	Query    string
	Language Language
}

type LookupStatus string

const (
	Found             LookupStatus = "found"
	InvalidQuery      LookupStatus = "invalid_query"
	NotFound          LookupStatus = "not_found"
	SourceUnavailable LookupStatus = "source_unavailable"
	SilentUnavailable LookupStatus = "silent_unavailable"
)

type ResultOrigin string

const (
	OriginOffline ResultOrigin = "offline"
	OriginMemory  ResultOrigin = "memory_cache"
	OriginDisk    ResultOrigin = "disk_cache"
	OriginOnline  ResultOrigin = "online"
)

type LookupResult struct {
	Status    LookupStatus
	Entries   []Entry
	Origin    ResultOrigin
	FetchedAt time.Time // Set for online and cached results.
	Stale     bool
}

type CacheKey struct {
	Provider        ProviderID
	Language        Language
	NormalizedQuery string
}

type CacheRecord struct {
	Key       CacheKey
	Entries   []Entry
	FetchedAt time.Time
	LastUsed  time.Time
}

// Dictionary describes trusted catalog metadata; package URLs and checksums are
// resolved within the download adapter and never supplied by the frontend.
type Dictionary struct {
	ID            DictionaryID
	Name          string
	Language      Language
	Source        Source
	Version       string
	License       string
	DownloadBytes int64
	Installed     bool
}

type DownloadState string

const (
	DownloadPending    DownloadState = "pending"
	DownloadInProgress DownloadState = "downloading"
	DownloadInstalling DownloadState = "installing"
	DownloadComplete   DownloadState = "complete"
	DownloadFailed     DownloadState = "failed"
)

type DownloadProgress struct {
	DictionaryID DictionaryID
	State        DownloadState
	BytesRead    int64
	TotalBytes   int64
	Message      string
}
