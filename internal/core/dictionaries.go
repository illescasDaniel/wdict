package core

import (
	"context"
	"errors"
)

// DictionaryService limits installation requests to entries in the trusted catalog.
type DictionaryService struct {
	Catalog   DictionaryCatalog
	Installer DictionaryInstaller
}

func (s *DictionaryService) List(ctx context.Context) ([]Dictionary, error) {
	return s.Catalog.List(ctx)
}
func (s *DictionaryService) Download(ctx context.Context, id DictionaryID, progress func(DownloadProgress)) error {
	_, ok, err := s.Catalog.Find(ctx, id)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("unknown dictionary")
	}
	return s.Installer.Install(ctx, id, progress)
}
