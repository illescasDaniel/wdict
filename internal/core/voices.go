package core

import (
	"context"
	"errors"
)

// VoicePackID identifies one catalog entry of the natural-voices feature: the
// Piper runtime or one voice pack.
type VoicePackID string

// VoicePackKind distinguishes the runtime entry from voice entries.
type VoicePackKind string

const (
	VoicePackRuntime VoicePackKind = "runtime"
	VoicePackVoice   VoicePackKind = "voice"
)

// VoicePack is one natural-voices catalog entry as shown in the Voices
// section. Blocked carries the user-facing copy when downloading is not
// possible (F4); empty means downloadable. URLs and checksums stay in the
// adapter, as with dictionaries.
type VoicePack struct {
	ID        VoicePackID
	Kind      VoicePackKind
	Name      string   // "Piper runtime", "en_GB · alba (medium)"
	Language  Language // empty for the runtime
	SizeBytes int64    // exact primary artifact size
	Source    Source
	Installed bool
	Blocked   string // user-facing copy when the download cannot start
}

// VoiceCatalog exposes the pinned natural-voices entries and their installed
// state. A missing entry is found=false, err=nil.
type VoiceCatalog interface {
	List(ctx context.Context) ([]VoicePack, error)
	Find(ctx context.Context, id VoicePackID) (pack VoicePack, found bool, err error)
}

// VoiceInstaller installs one catalog pack with bounded downloading, integrity
// checks and atomic installation. A failure leaves the previous pack intact.
type VoiceInstaller interface {
	Install(ctx context.Context, id VoicePackID, progress func(DownloadProgress)) error
}

// VoiceService limits voice-pack requests to entries in the trusted catalog.
type VoiceService struct {
	Catalog   VoiceCatalog
	Installer VoiceInstaller
}

func (s *VoiceService) List(ctx context.Context) ([]VoicePack, error) {
	return s.Catalog.List(ctx)
}

func (s *VoiceService) Download(ctx context.Context, id VoicePackID, progress func(DownloadProgress)) error {
	_, ok, err := s.Catalog.Find(ctx, id)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("unknown voice pack")
	}
	return s.Installer.Install(ctx, id, progress)
}
