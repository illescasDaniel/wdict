package adapters

import (
	"archive/tar"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func archiveFile(t *testing.T, entries map[string]byte) string {
	t.Helper()
	var b bytes.Buffer
	z, err := zstd.NewWriter(&b)
	if err != nil {
		t.Fatal(err)
	}
	w := tar.NewWriter(z)
	for name, kind := range entries {
		h := &tar.Header{Name: name, Mode: 0600, Typeflag: kind}
		if kind == tar.TypeReg {
			h.Size = 1
		}
		if err = w.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if kind == tar.TypeReg {
			if _, err = w.Write([]byte("x")); err != nil {
				t.Fatal(err)
			}
		}
	}
	w.Close()
	z.Close()
	p := filepath.Join(t.TempDir(), "data.zst")
	if err = os.WriteFile(p, b.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return p
}
func TestGivenUnsafeArchivesWhenExtractedThenTheyAreRejected(t *testing.T) {
	// given
	for _, entries := range []map[string]byte{{"../outside": tar.TypeReg}, {"en-en.ifo": tar.TypeSymlink}, {"en-en.ifo": tar.TypeReg}} {
		// when
		// then
		if err := extractDictionary(archiveFile(t, entries), t.TempDir()); err == nil {
			t.Fatal("accepted unsafe or incomplete archive")
		}
	}
}
func TestGivenRequiredFilesWhenExtractedThenTheyAreInstalled(t *testing.T) {
	// given
	entries := map[string]byte{"./en-en.ifo": tar.TypeReg, "./en-en.idx": tar.TypeReg, "./en-en.dict.dz": tar.TypeReg, "./en-en.syn": tar.TypeReg}
	dest := t.TempDir()
	// when
	if err := extractDictionary(archiveFile(t, entries), dest); err != nil {
		t.Fatal(err)
	}
	// then
	for name := range entries {
		if _, err := os.Stat(filepath.Join(dest, name)); err != nil {
			t.Fatal(err)
		}
	}
}
func TestGivenUnknownDictionaryWhenDownloadingThenItIsRejected(t *testing.T) {
	// given
	i := Installer{Dir: t.TempDir()}
	// when
	// then
	if err := i.Download(context.Background(), "unknown", nil); err == nil {
		t.Fatal("accepted unknown catalog id")
	}
}
