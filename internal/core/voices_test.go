package core

import (
	"context"
	"errors"
	"testing"
)

type fakeVoicePacks struct {
	listed []VoicePack
	err    error

	found   bool
	pack    VoicePack
	errFind error

	installedIDs []VoicePackID
	progressSeen []DownloadProgress
}

func (f *fakeVoicePacks) List(context.Context) ([]VoicePack, error) { return f.listed, f.err }
func (f *fakeVoicePacks) Find(context.Context, VoicePackID) (VoicePack, bool, error) {
	return f.pack, f.found, f.errFind
}
func (f *fakeVoicePacks) Install(_ context.Context, id VoicePackID, progress func(DownloadProgress)) error {
	f.installedIDs = append(f.installedIDs, id)
	if progress != nil {
		p := DownloadProgress{DictionaryID: DictionaryID(id), State: DownloadInProgress}
		f.progressSeen = append(f.progressSeen, p)
		progress(p)
	}
	return nil
}

func TestVoiceServiceListDelegatesToCatalog(t *testing.T) {
	// (C6): the service lists exactly what the catalog reports.
	pack := VoicePack{ID: "piper-runtime", Kind: VoicePackRuntime, Name: "Piper runtime"}
	catalog := &fakeVoicePacks{listed: []VoicePack{pack}}
	service := &VoiceService{Catalog: catalog, Installer: &fakeVoicePacks{}}
	// when
	list, err := service.List(context.Background())
	// then
	if err != nil || len(list) != 1 || list[0].ID != pack.ID {
		t.Fatalf("list=%v err=%v", list, err)
	}
}

func TestVoiceServiceDownloadRejectsUnknownPacks(t *testing.T) {
	// (C1): only catalog entries are downloadable.
	catalog := &fakeVoicePacks{}
	installer := &fakeVoicePacks{}
	service := &VoiceService{Catalog: catalog, Installer: installer}
	// when
	err := service.Download(context.Background(), "not-in-catalog", nil)
	// then
	if err == nil {
		t.Fatal("unknown pack accepted")
	}
	if len(installer.installedIDs) != 0 {
		t.Fatalf("installer called: %v", installer.installedIDs)
	}
}

func TestVoiceServiceDownloadInstallsCatalogPacks(t *testing.T) {
	// (C2/C3): a catalog id reaches the installer, with the progress callback.
	id := VoicePackID("en_GB-alba-medium")
	catalog := &fakeVoicePacks{found: true, pack: VoicePack{ID: id}}
	installer := &fakeVoicePacks{}
	service := &VoiceService{Catalog: catalog, Installer: installer}
	seen := 0
	// when
	err := service.Download(context.Background(), id, func(p DownloadProgress) { seen++ })
	// then
	if err != nil {
		t.Fatal(err)
	}
	if len(installer.installedIDs) != 1 || installer.installedIDs[0] != id {
		t.Fatalf("installer got %v", installer.installedIDs)
	}
	if seen != 1 || len(installer.progressSeen) != 1 {
		t.Fatalf("progress callbacks=%d recorded=%d", seen, len(installer.progressSeen))
	}
	if installer.progressSeen[0].DictionaryID != DictionaryID(id) {
		t.Fatalf("progress carries %q", installer.progressSeen[0].DictionaryID)
	}
}

func TestVoiceServiceDownloadPropagatesCatalogErrors(t *testing.T) {
	// given: the catalog lookup itself fails
	catalog := &fakeVoicePacks{errFind: errors.New("catalog unavailable")}
	service := &VoiceService{Catalog: catalog, Installer: &fakeVoicePacks{}}
	// when
	err := service.Download(context.Background(), "piper-runtime", nil)
	// then
	if err == nil {
		t.Fatal("catalog error swallowed")
	}
}
