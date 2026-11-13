package adapters

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/illescasDaniel/wdict/internal/core"
)

// VoicePackInstaller installs the pinned natural-voices packs (C2/C3): the
// Piper runtime as a python3 venv with piper-tts pinned, and voice packs as
// two checksummed files installed atomically. It implements core.VoiceCatalog
// (list/find over the installed state) and core.VoiceInstaller.
type VoicePackInstaller struct {
	Dir           string // <data>/wdict/piper
	Python        string // python3 binary name; default "python3" (injectable for tests)
	Client        *http.Client
	VoiceCapBytes int64 // per-file download cap; default 150 MiB

	lookPath func(string) (string, error)                                           // default exec.LookPath; replaceable in tests
	runner   func(ctx context.Context, name string, args ...string) ([]byte, error) // default runCommand; replaceable in tests

	packOverride map[core.VoicePackID]voicePackEntry // injected only by package tests

	mu sync.Mutex
}

const (
	piperVenvTimeout    = 120 * time.Second
	piperPipTimeout     = 10 * time.Minute
	piperDefaultCap     = int64(150) << 20
	piperTTSRequirement = "piper-tts==1.8.0"
)

func (i *VoicePackInstaller) pythonBinary() string {
	if i.Python != "" {
		return i.Python
	}
	return "python3"
}

func (i *VoicePackInstaller) capBytes() int64 {
	if i.VoiceCapBytes > 0 {
		return i.VoiceCapBytes
	}
	return piperDefaultCap
}

func (i *VoicePackInstaller) venvDir() string { return filepath.Join(i.Dir, "venv") }
func (i *VoicePackInstaller) venvPython() string {
	return filepath.Join(i.Dir, "venv", "bin", "python3")
}
func (i *VoicePackInstaller) manifestPath() string { return filepath.Join(i.Dir, "manifest.txt") }
func (i *VoicePackInstaller) voicesDir() string    { return filepath.Join(i.Dir, "voices") }
func (i *VoicePackInstaller) voiceDir(id core.VoicePackID) string {
	return filepath.Join(i.voicesDir(), string(id))
}

// entry resolves a pack id against the trusted catalog; package tests may
// replace the record's URLs and checksums.
func (i *VoicePackInstaller) entry(id core.VoicePackID) (voicePackEntry, bool) {
	if i.packOverride != nil {
		if e, ok := i.packOverride[id]; ok {
			return e, true
		}
	}
	for _, e := range voiceCatalog {
		if e.id == id {
			return e, true
		}
	}
	return voicePackEntry{}, false
}

// probePython reports whether the python3 binary is on the PATH (F4 probe).
func (i *VoicePackInstaller) probePython() bool {
	look := i.lookPath
	if look == nil {
		look = exec.LookPath
	}
	_, err := look(i.pythonBinary())
	return err == nil
}

// runtimeInstalled reports a healthy runtime: the venv python and the
// dependency manifest are present (C6: a corrupted runtime counts as absent).
func (i *VoicePackInstaller) runtimeInstalled() bool {
	info, err := os.Stat(i.venvPython())
	if err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
		return false
	}
	info, err = os.Stat(i.manifestPath())
	return err == nil && info.Mode().IsRegular() && info.Size() > 0
}

// voiceInstalled reports both voice files present and non-empty (C3).
func (i *VoicePackInstaller) voiceInstalled(id core.VoicePackID) bool {
	dir := i.voiceDir(id)
	for _, name := range []string{string(id) + ".onnx", string(id) + ".onnx.json"} {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
			return false
		}
	}
	return true
}

// List reflects the installed and blocked states of the pinned catalog (C6).
func (i *VoicePackInstaller) List(_ context.Context) ([]core.VoicePack, error) {
	pythonOK := i.probePython()
	runtime := i.runtimeInstalled()
	packs := make([]core.VoicePack, 0, len(voiceCatalog))
	for _, e := range voiceCatalog {
		record, ok := i.entry(e.id)
		if !ok {
			continue
		}
		pack := core.VoicePack{ID: record.id, Kind: record.kind, Name: record.name, Language: record.language, SizeBytes: record.size, Source: record.source}
		switch record.kind {
		case core.VoicePackRuntime:
			pack.Installed = runtime
			if !pythonOK {
				pack.Blocked = piperPythonCopy
			}
		case core.VoicePackVoice:
			pack.Installed = i.voiceInstalled(record.id)
			if !pythonOK {
				pack.Blocked = piperPythonCopy
			} else if !runtime {
				pack.Blocked = piperRuntimeFirstCopy
			}
		}
		packs = append(packs, pack)
	}
	return packs, nil
}

// Find reports a missing pack as found=false, err=nil (core.VoiceCatalog).
func (i *VoicePackInstaller) Find(ctx context.Context, id core.VoicePackID) (core.VoicePack, bool, error) {
	list, err := i.List(ctx)
	if err != nil {
		return core.VoicePack{}, false, err
	}
	for _, pack := range list {
		if pack.ID == id {
			return pack, true, nil
		}
	}
	return core.VoicePack{}, false, nil
}

// Install routes one catalog pack to the runtime or voice installer. A
// failure never damages the previously installed packs (F3).
func (i *VoicePackInstaller) Install(ctx context.Context, id core.VoicePackID, progress func(core.DownloadProgress)) error {
	entry, ok := i.entry(id)
	if !ok {
		return errors.New("unknown voice pack")
	}
	if !i.mu.TryLock() {
		return errors.New("download already in progress")
	}
	defer i.mu.Unlock()
	switch entry.kind {
	case core.VoicePackRuntime:
		return i.installRuntime(ctx, id, progress)
	case core.VoicePackVoice:
		return i.installVoice(ctx, id, entry, progress)
	}
	return errors.New("unknown voice pack")
}

func (i *VoicePackInstaller) report(progress func(core.DownloadProgress), id core.VoicePackID, state core.DownloadState, read, total int64) {
	if progress != nil {
		progress(core.DownloadProgress{DictionaryID: core.DictionaryID(id), State: state, BytesRead: read, TotalBytes: total})
	}
}

// installRuntime creates the venv, installs the pinned piper-tts version and
// records the dependency manifest (C2). A failed install is never partially
// reused: every attempt recreates the venv from scratch (S6/S10).
func (i *VoicePackInstaller) installRuntime(ctx context.Context, id core.VoicePackID, progress func(core.DownloadProgress)) error {
	if !i.probePython() {
		return errors.New(piperPythonCopy)
	}
	i.report(progress, id, core.DownloadInProgress, 0, piperWheelSize)
	if i.runtimeInstalled() {
		i.report(progress, id, core.DownloadComplete, piperWheelSize, piperWheelSize)
		return nil
	}
	if err := os.MkdirAll(i.Dir, 0700); err != nil {
		return err
	}
	// full recreate: delete any previous (possibly partial) venv first
	if err := os.RemoveAll(i.venvDir()); err != nil {
		return err
	}
	venvCtx, cancel := context.WithTimeout(ctx, piperVenvTimeout)
	defer cancel()
	if _, err := i.run(venvCtx, i.pythonBinary(), "-m", "venv", i.venvDir()); err != nil {
		// venv creation failure (e.g. missing python3-venv on Debian) shows
		// the F4 copy; nothing partial is reused (S10)
		return errors.New(piperPythonCopy)
	}
	pipCtx, cancel := context.WithTimeout(ctx, piperPipTimeout)
	defer cancel()
	if _, err := i.run(pipCtx, i.venvPython(), "-m", "pip", "install", piperTTSRequirement); err != nil {
		return err
	}
	freeze, err := i.run(pipCtx, i.venvPython(), "-m", "pip", "freeze")
	if err != nil {
		return err
	}
	if err := writeFileAtomically(i.manifestPath(), freeze); err != nil {
		return err
	}
	i.report(progress, id, core.DownloadComplete, piperWheelSize, piperWheelSize)
	return nil
}

// installVoice downloads both pinned files, verifies their SHA-256 and
// installs them atomically (C3). The previous pack stays intact on failure.
func (i *VoicePackInstaller) installVoice(ctx context.Context, id core.VoicePackID, entry voicePackEntry, progress func(core.DownloadProgress)) error {
	if !i.runtimeInstalled() {
		return errors.New(piperRuntimeFirstCopy)
	}
	i.report(progress, id, core.DownloadInProgress, 0, entry.size)
	if i.voiceInstalled(id) {
		i.report(progress, id, core.DownloadComplete, entry.size, entry.size)
		return nil
	}
	if err := os.MkdirAll(i.voicesDir(), 0700); err != nil {
		return err
	}
	temp, err := os.MkdirTemp(i.voicesDir(), ".voice-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	onnx := filepath.Join(temp, string(id)+".onnx")
	if err := i.fetchVoiceFile(ctx, entry.url, entry.sha, onnx, entry.size, id, progress); err != nil {
		return err
	}
	meta := filepath.Join(temp, string(id)+".onnx.json")
	if err := i.fetchVoiceFile(ctx, entry.extraURL, entry.extraSHA, meta, entry.extraSize, id, progress); err != nil {
		return err
	}
	i.report(progress, id, core.DownloadInstalling, entry.size, entry.size)
	target := i.voiceDir(id)
	if _, err := os.Stat(target); err == nil {
		// replace an existing (invalid) pack only after both files verified
		if err := os.RemoveAll(target); err != nil {
			return err
		}
	}
	if err := os.Rename(temp, target); err != nil {
		return err
	}
	i.report(progress, id, core.DownloadComplete, entry.size, entry.size)
	return nil
}

// fetchVoiceFile downloads one pinned file with the per-file cap, verifies
// the SHA-256 and writes the bytes to dest.
func (i *VoicePackInstaller) fetchVoiceFile(ctx context.Context, url, sha, dest string, total int64, id core.VoicePackID, progress func(core.DownloadProgress)) error {
	capBytes := i.capBytes()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	client := i.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Minute}
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("voice download returned HTTP %d", response.StatusCode)
	}
	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	hash := sha256.New()
	buf := make([]byte, 128*1024)
	var read int64
	for {
		nr, er := response.Body.Read(buf)
		if nr > 0 {
			read += int64(nr)
			if read > capBytes {
				file.Close()
				return fmt.Errorf("voice download exceeds the %d byte cap", capBytes)
			}
			if _, err = io.MultiWriter(file, hash).Write(buf[:nr]); err != nil {
				file.Close()
				return err
			}
			i.report(progress, id, core.DownloadInProgress, read, total)
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			file.Close()
			return er
		}
	}
	if err = file.Close(); err != nil {
		return err
	}
	if hex.EncodeToString(hash.Sum(nil)) != sha {
		return errors.New("voice download checksum verification failed")
	}
	return nil
}

// run executes one installer command with the injected runner seam.
func (i *VoicePackInstaller) run(ctx context.Context, name string, args ...string) ([]byte, error) {
	if i.runner != nil {
		return i.runner(ctx, name, args...)
	}
	return runCommand(ctx, name, args...)
}

// runCommand runs one subprocess, returning stdout; errors carry stderr.
func runCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

// writeFileAtomically writes bytes to path via a same-directory temp file.
func writeFileAtomically(path string, data []byte) error {
	temp, err := os.CreateTemp(filepath.Dir(path), ".manifest-*")
	if err != nil {
		return err
	}
	name := temp.Name()
	defer os.Remove(name)
	if _, err = temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}
