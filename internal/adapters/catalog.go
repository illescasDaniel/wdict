package adapters

import (
	"context"
	"errors"
	"github.com/illescasDaniel/wdict/internal/core"
)

// Catalog routes fixed IDs to independent installers.
type Catalog struct{ English, Spanish, French *Installer }

func (c *Catalog) List(ctx context.Context) ([]core.Dictionary, error) {
	a, e := c.English.List(ctx)
	if e != nil {
		return nil, e
	}
	b, e := c.Spanish.List(ctx)
	if e != nil {
		return nil, e
	}
	if c.French != nil {
		fr, err := c.French.List(ctx)
		if err != nil {
			return nil, err
		}
		b = append(b, fr...)
	}
	return append(a, b...), nil
}
func (c *Catalog) Find(ctx context.Context, id core.DictionaryID) (core.Dictionary, bool, error) {
	for _, i := range []*Installer{c.English, c.Spanish, c.French} {
		if i == nil {
			continue
		}
		d, ok, e := i.Find(ctx, id)
		if e != nil || ok {
			return d, ok, e
		}
	}
	return core.Dictionary{}, false, nil
}
func (c *Catalog) Install(ctx context.Context, id core.DictionaryID, progress func(core.DownloadProgress)) error {
	for _, i := range []*Installer{c.English, c.Spanish, c.French} {
		if i == nil {
			continue
		}
		if id == i.id() {
			return i.Install(ctx, id, progress)
		}
	}
	return errors.New("unknown dictionary")
}
