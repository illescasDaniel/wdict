package core

import (
	"context"
	"errors"
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

// ErrSilentUnavailable marks a temporary provider failure that is not shown as an error.
var ErrSilentUnavailable = errors.New("provider temporarily unavailable")

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

// Voices is the inbound interface for listing and installing the natural
// voices packs. The progress callback may be nil. An installation failure
// must be reported and leave the previously installed pack intact.
type Voices interface {
	List(ctx context.Context) ([]VoicePack, error)
	Download(ctx context.Context, id VoicePackID, progress func(DownloadProgress)) error
}

// PiperVoice synthesises speech from the shown IPA with a neural voice when
// the pack for the language is installed. Available reports whether the
// runtime and a voice for the language are installed; when false, Speak is
// never called and the caller falls back to eSpeak. Voice names the short
// voice id (e.g. "alba") used for the language, for the Piper result label;
// empty when no voice is installed.
type PiperVoice interface {
	Available(language Language) bool
	Voice(language Language) string
	Speak(ctx context.Context, language Language, ipa string) (data []byte, err error)
}

// Clock permits deterministic freshness and last-used decisions.
type Clock interface {
	Now() time.Time
}

// RecordingSource fetches a recording: cache first, then a bounded download.
// Implementations validate the URL at the edge and own the cache; a miss or a
// failed fetch returns an error and the caller falls back to the voice path.
type RecordingSource interface {
	Fetch(ctx context.Context, audioURL string) (data []byte, err error)
}

// VoiceSynthesiser speaks an eSpeak voice, either mnemonic phonemes or plain
// text (C5, revision 3: core converts the shown IPA to mnemonics and falls
// back to the headword). Available reports whether the eSpeak binary exists;
// when false, neither method is called and the caller shows the
// install-guidance copy.
type VoiceSynthesiser interface {
	Available() bool
	SpeakPhonemes(ctx context.Context, voice, mnemonics string) (data []byte, err error)
	SpeakText(ctx context.Context, voice, text string) (data []byte, err error)
}
