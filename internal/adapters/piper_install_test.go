package adapters

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/illescasDaniel/wdict/internal/core"
)

// fakeCmdRunner stands in for the python/pip runner (C2 seams). It recognises
// the venv, pip install and pip freeze invocations by their arguments,
// simulates the venv creation on disk, and records whether a marker from a
// previous venv invocation existed at each venv call (full-recreate check).
type fakeCmdRunner struct {
	calls          []string
	freezeOut      string
	venvErr        error
	pipErr         error
	venvDirExisted []bool
}

func (f *fakeCmdRunner) run(_ context.Context, name string, args ...string) ([]byte, error) {
	call := strings.Join(append([]string{name}, args...), " ")
	f.calls = append(f.calls, call)
	switch {
	case strings.Contains(call, " venv "):
		marker := filepath.Join(args[len(args)-1], ".venv-created")
		_, statErr := os.Stat(marker)
		f.venvDirExisted = append(f.venvDirExisted, statErr == nil)
		if f.venvErr != nil {
			return nil, f.venvErr
		}
		if err := os.MkdirAll(args[len(args)-1], 0700); err != nil {
			return nil, err
		}
		return nil, os.WriteFile(marker, nil, 0600)
	case strings.Contains(call, " install "):
		if f.pipErr != nil {
			return nil, f.pipErr
		}
		return nil, nil
	case strings.Contains(call, " freeze"):
		return []byte(f.freezeOut), nil
	}
	return nil, nil
}

func pretendRuntime(t *testing.T, dir string) {
	t.Helper()
	python := filepath.Join(dir, "venv", "bin", "python3")
	if err := os.MkdirAll(filepath.Dir(python), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(python, []byte("#!/bin/sh\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.txt"), []byte("piper-tts==1.8.0\n"), 0600); err != nil {
		t.Fatal(err)
	}
}

func pretendVoice(t *testing.T, dir, id string) {
	t.Helper()
	voice := filepath.Join(dir, "voices", id)
	if err := os.MkdirAll(voice, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{id + ".onnx", id + ".onnx.json"} {
		if err := os.WriteFile(filepath.Join(voice, name), []byte("data"), 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGivenCatalogWhenListedThenRuntimeAndThreeVoicesArePinned(t *testing.T) {
	// (C1/C6): the catalog is code-only; entries carry names, languages,
	// exact sizes and attribution.
	i := &VoicePackInstaller{Dir: t.TempDir()}
	list, err := i.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 4 {
		t.Fatalf("entries=%d", len(list))
	}
	runtime, voices := list[0], list[1:]
	if runtime.ID != piperRuntimeID || runtime.Kind != core.VoicePackRuntime || runtime.Name != "Piper runtime" || runtime.Language != "" {
		t.Fatalf("runtime entry %+v", runtime)
	}
	if runtime.SizeBytes != 34131442 {
		t.Fatalf("runtime size %d", runtime.SizeBytes)
	}
	if runtime.Source.LicenseURL == "" || runtime.Source.Attribution == "" {
		t.Fatalf("runtime source %+v", runtime.Source)
	}
	if len(voices) != 3 {
		t.Fatalf("voice entries %d", len(voices))
	}
	want := []struct {
		id       string
		name     string
		language core.Language
		size     int64
	}{
		{"en_GB-alba-medium", "en_GB · alba (medium)", core.English, 63221984},
		{"en_US-amy-medium", "en_US · amy (medium)", core.English, 63221984},
		{"es_ES-sharvard-medium", "es_ES · sharvard (medium)", core.Spanish, 76753841},
	}
	for n, v := range voices {
		if v.ID != core.VoicePackID(want[n].id) || v.Kind != core.VoicePackVoice || v.Name != want[n].name || v.Language != want[n].language || v.SizeBytes != want[n].size {
			t.Fatalf("entry %d = %+v", n, v)
		}
		if v.Source.URL == "" || v.Source.Attribution == "" || v.Source.LicenseURL == "" {
			t.Fatalf("entry %d source %+v", n, v.Source)
		}
	}
}

func TestGivenDiskStateWhenListedThenInstalledAndBlockedReflectIt(t *testing.T) {
	// (C6): installed state comes from the disk; voice entries are blocked
	// until the runtime is installed (C7).
	i := &VoicePackInstaller{Dir: t.TempDir()}
	list, err := i.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if list[0].Installed {
		t.Fatal("runtime reported installed on an empty dir")
	}
	for _, v := range list[1:] {
		if v.Installed || v.Blocked != "Download the Piper runtime first." {
			t.Fatalf("voice entry %+v", v)
		}
	}
	// when: runtime and one voice are installed on the disk
	pretendRuntime(t, i.Dir)
	pretendVoice(t, i.Dir, "en_GB-alba-medium")
	list, err = i.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// then
	if !list[0].Installed || list[0].Blocked != "" {
		t.Fatalf("runtime entry %+v", list[0])
	}
	if !list[1].Installed || list[1].Blocked != "" {
		t.Fatalf("alba entry %+v", list[1])
	}
	if list[2].Installed || list[2].Blocked != "" {
		t.Fatalf("amy entry %+v", list[2])
	}
}

func TestGivenPythonMissingWhenListedThenEveryEntryIsBlockedWithTheF4Copy(t *testing.T) {
	// (S8/F4): the copy is shown verbatim; playback is unaffected.
	i := &VoicePackInstaller{Dir: t.TempDir(), lookPath: func(string) (string, error) { return "", errors.New("not found") }}
	list, err := i.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	const copy = "Python 3 is required for natural voices. Install Python 3 and retry."
	for n, pack := range list {
		if pack.Installed || pack.Blocked != copy {
			t.Fatalf("entry %d = %+v", n, pack)
		}
	}
}

func TestGivenPythonMissingWhenRuntimeInstallAttemptedThenTheF4CopyIsReturned(t *testing.T) {
	// (S8/F4): the download cannot start, and the message is the user copy.
	i := &VoicePackInstaller{Dir: t.TempDir(), lookPath: func(string) (string, error) { return "", errors.New("not found") }}
	err := i.Install(context.Background(), piperRuntimeID, nil)
	if err == nil || err.Error() != "Python 3 is required for natural voices. Install Python 3 and retry." {
		t.Fatalf("err=%v", err)
	}
}

func TestGivenPythonWhenRuntimeInstalledThenVenvPipAndManifestRun(t *testing.T) {
	// (S5): venv → pip install (pinned version) → pip freeze manifest.
	dir := t.TempDir()
	runner := &fakeCmdRunner{freezeOut: "piper-tts==1.8.0\nnumpy==2.0.0\n"}
	i := &VoicePackInstaller{Dir: dir, runner: runner.run}
	var last core.DownloadProgress
	err := i.Install(context.Background(), piperRuntimeID, func(p core.DownloadProgress) { last = p })
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"python3 -m venv " + filepath.Join(dir, "venv"),
		filepath.Join(dir, "venv", "bin", "python3") + " -m pip install piper-tts==1.8.0",
		filepath.Join(dir, "venv", "bin", "python3") + " -m pip freeze",
	}
	if len(runner.calls) != len(want) {
		t.Fatalf("calls %v", runner.calls)
	}
	for n, call := range want {
		if runner.calls[n] != call {
			t.Fatalf("call %d = %q, want %q", n, runner.calls[n], call)
		}
	}
	manifest, err := os.ReadFile(filepath.Join(dir, "manifest.txt"))
	if err != nil || string(manifest) != runner.freezeOut {
		t.Fatalf("manifest=%q err=%v", manifest, err)
	}
	if last.State != core.DownloadComplete {
		t.Fatalf("progress %+v", last)
	}
}

func TestGivenHealthyRuntimeWhenRuntimeInstalledThenItIsANoOp(t *testing.T) {
	// (S5): an already-installed healthy runtime repeats nothing.
	dir := t.TempDir()
	pretendRuntime(t, dir)
	runner := &fakeCmdRunner{}
	i := &VoicePackInstaller{Dir: dir, runner: runner.run}
	if err := i.Install(context.Background(), piperRuntimeID, nil); err != nil {
		t.Fatal(err)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("runner called: %v", runner.calls)
	}
}

func TestGivenFailedRuntimeInstallWhenRetriedThenTheVenvIsRecreatedFromScratch(t *testing.T) {
	// (S6): no partial runtime is reused; the retry deletes the venv first.
	dir := t.TempDir()
	runner := &fakeCmdRunner{pipErr: errors.New("network down")}
	i := &VoicePackInstaller{Dir: dir, runner: runner.run}
	if err := i.Install(context.Background(), piperRuntimeID, nil); err == nil {
		t.Fatal("failed install accepted")
	}
	runner.pipErr = nil
	// when: retry
	if err := i.Install(context.Background(), piperRuntimeID, nil); err != nil {
		t.Fatal(err)
	}
	// then: the marker created by the first attempt was gone before the
	// second venv invocation (full recreate), and both attempts ran venv
	if len(runner.venvDirExisted) != 2 {
		t.Fatalf("venv calls %v", runner.calls)
	}
	for n, existed := range runner.venvDirExisted {
		if existed {
			t.Fatalf("venv call %d reused a previous venv", n)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "manifest.txt")); err != nil {
		t.Fatal(err)
	}
}

func TestGivenVenvCreationFailsWhenRuntimeInstallAttemptedThenTheF4CopyIsReturned(t *testing.T) {
	// (S10/F4): python3 exists but venv creation fails (python3-venv missing).
	runner := &fakeCmdRunner{venvErr: errors.New("ensurepip is missing")}
	i := &VoicePackInstaller{Dir: t.TempDir(), runner: runner.run}
	err := i.Install(context.Background(), piperRuntimeID, nil)
	if err == nil || err.Error() != "Python 3 is required for natural voices. Install Python 3 and retry." {
		t.Fatalf("err=%v", err)
	}
}

func voiceTestInstaller(t *testing.T, dir string, onnx, meta []byte, sha, metaSHA string, capBytes int64) (*VoicePackInstaller, *int) {
	t.Helper()
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if strings.HasSuffix(r.URL.Path, ".json") {
			_, _ = w.Write(meta)
			return
		}
		_, _ = w.Write(onnx)
	}))
	t.Cleanup(server.Close)
	entry := voiceCatalog[1] // en_GB-alba-medium shape; URLs and checksums replaced
	entry.url = server.URL + "/en_GB-alba-medium.onnx"
	entry.sha = sha
	entry.extraURL = server.URL + "/en_GB-alba-medium.onnx.json"
	entry.extraSHA = metaSHA
	i := &VoicePackInstaller{Dir: dir, Client: server.Client(), VoiceCapBytes: capBytes, packOverride: map[core.VoicePackID]voicePackEntry{entry.id: entry}}
	return i, &requests
}

func TestGivenVoiceFilesWhenInstalledThenBothAreVerifiedAndInstalledAtomically(t *testing.T) {
	// (S7): both files are checksummed, then installed atomically; no
	// temporary directories remain.
	dir := t.TempDir()
	pretendRuntime(t, dir)
	onnx, meta := []byte("ONNXBYTES"), []byte("{\"audio\":{}}")
	sha := sha256.Sum256(onnx)
	metaSHA := sha256.Sum256(meta)
	i, requests := voiceTestInstaller(t, dir, onnx, meta, hex.EncodeToString(sha[:]), hex.EncodeToString(metaSHA[:]), 0)
	var last core.DownloadProgress
	err := i.Install(context.Background(), core.VoicePackID("en_GB-alba-medium"), func(p core.DownloadProgress) { last = p })
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "voices", "en_GB-alba-medium", "en_GB-alba-medium.onnx"))
	if err != nil || string(got) != string(onnx) {
		t.Fatalf("onnx=%q err=%v", got, err)
	}
	got, err = os.ReadFile(filepath.Join(dir, "voices", "en_GB-alba-medium", "en_GB-alba-medium.onnx.json"))
	if err != nil || string(got) != string(meta) {
		t.Fatalf("json=%q err=%v", got, err)
	}
	entries, _ := os.ReadDir(filepath.Join(dir, "voices"))
	if len(entries) != 1 {
		t.Fatalf("voices dir = %v", entries)
	}
	if *requests != 2 {
		t.Fatalf("requests=%d", *requests)
	}
	if last.State != core.DownloadComplete {
		t.Fatalf("progress %+v", last)
	}
	// (C6): the list now reports the voice installed.
	list, err := i.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !list[1].Installed {
		t.Fatalf("alba entry %+v", list[1])
	}
}

func TestGivenChecksumMismatchWhenVoiceInstalledThenNothingIsInstalled(t *testing.T) {
	// (S7/F3): a wrong checksum fails the download and installs nothing.
	dir := t.TempDir()
	pretendRuntime(t, dir)
	i, _ := voiceTestInstaller(t, dir, []byte("TAMPERED"), []byte("{}"), "0000", strings.Repeat("0", 64), 0)
	err := i.Install(context.Background(), core.VoicePackID("en_GB-alba-medium"), nil)
	if err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "voices", "en_GB-alba-medium")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("partial voice installed: %v", err)
	}
	entries, _ := os.ReadDir(filepath.Join(dir, "voices"))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".voice-") {
			t.Fatalf("temp dir left behind: %v", e.Name())
		}
	}
}

func TestGivenVoiceOverTheDownloadCapWhenInstalledThenTheDownloadAborts(t *testing.T) {
	// (C3): a per-file download cap aborts oversized bodies.
	dir := t.TempDir()
	pretendRuntime(t, dir)
	body := []byte(strings.Repeat("x", 32))
	sha := sha256.Sum256([]byte("unused"))
	i, _ := voiceTestInstaller(t, dir, body, []byte("{}"), hex.EncodeToString(sha[:]), strings.Repeat("0", 64), 8)
	if err := i.Install(context.Background(), core.VoicePackID("en_GB-alba-medium"), nil); err == nil {
		t.Fatal("over-cap download accepted")
	}
	if _, err := os.Stat(filepath.Join(dir, "voices", "en_GB-alba-medium")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("voice installed despite the cap: %v", err)
	}
}

func TestGivenRuntimeMissingWhenVoiceInstallAttemptedThenItIsRefused(t *testing.T) {
	// (C7): voice entries are blocked until the runtime exists.
	dir := t.TempDir()
	i := &VoicePackInstaller{Dir: dir}
	err := i.Install(context.Background(), core.VoicePackID("en_GB-alba-medium"), nil)
	if err == nil || err.Error() != "Download the Piper runtime first." {
		t.Fatalf("err=%v", err)
	}
}

func TestGivenInstalledVoiceWhenInstalledAgainThenItIsANoOp(t *testing.T) {
	// (C3 idempotence): both files present = installed; no re-download.
	dir := t.TempDir()
	pretendRuntime(t, dir)
	pretendVoice(t, dir, "en_GB-alba-medium")
	i, requests := voiceTestInstaller(t, dir, []byte("new"), []byte("{}"), strings.Repeat("f", 64), strings.Repeat("f", 64), 0)
	if err := i.Install(context.Background(), core.VoicePackID("en_GB-alba-medium"), nil); err != nil {
		t.Fatal(err)
	}
	if *requests != 0 {
		t.Fatalf("requests=%d", *requests)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "voices", "en_GB-alba-medium", "en_GB-alba-medium.onnx")); string(got) != "data" {
		t.Fatalf("installed files were replaced: %q", got)
	}
}

func TestGivenUnknownPackWhenInstalledThenItIsRejected(t *testing.T) {
	i := &VoicePackInstaller{Dir: t.TempDir()}
	if err := i.Install(context.Background(), "not-a-pack", nil); err == nil {
		t.Fatal("unknown pack accepted")
	}
}
