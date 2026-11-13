//go:build integration

package adapters

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/illescasDaniel/wdict/internal/core"
)

// TestGivenRealPipAndVoiceWhenInstalledThenPiperSynthesisesAudibleWAV runs the
// real C2/C3/C4 chain end to end: a real python3 venv with piper-tts pinned,
// the real checksummed voice download and a real synthesis through the
// audibility guard. Heavy: it downloads ≈120 MB of runtime and a 63 MB voice
// into a temp dir. Skipped without python3.
func TestGivenRealPipAndVoiceWhenInstalledThenPiperSynthesisesAudibleWAV(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not installed")
	}
	dir := t.TempDir()
	installer := &VoicePackInstaller{Dir: dir}
	ctx := context.Background()
	// when: the real runtime install (C2/S5)
	started := time.Now()
	if err := installer.Install(ctx, piperRuntimeID, nil); err != nil {
		t.Fatalf("runtime install: %v", err)
	}
	t.Logf("runtime install took %v", time.Since(started))
	if !installer.runtimeInstalled() {
		t.Fatal("runtime reported absent after a successful install")
	}
	manifest, err := os.ReadFile(installer.manifestPath())
	if err != nil || !strings.Contains(string(manifest), "piper-tts") {
		t.Fatalf("manifest=%q err=%v", manifest, err)
	}
	// when: the real voice download (C3/S7)
	started = time.Now()
	if err := installer.Install(ctx, core.VoicePackID("en_GB-alba-medium"), nil); err != nil {
		t.Fatalf("voice install: %v", err)
	}
	t.Logf("voice install took %v", time.Since(started))
	if !installer.voiceInstalled("en_GB-alba-medium") {
		t.Fatal("voice reported absent after a successful install")
	}
	// when: real synthesis through the adapter (C4/S1)
	piper := &PiperVoices{Dir: dir, CacheDir: filepath.Join(t.TempDir(), "audio")}
	if !piper.Available(core.English) {
		t.Fatal("Piper reported unavailable with the pack installed")
	}
	started = time.Now()
	data, err := piper.Speak(ctx, core.English, "ˈnɒl.ɪdʒ")
	if err != nil {
		t.Fatalf("speak: %v", err)
	}
	t.Logf("first synthesis took %v (%d bytes)", time.Since(started), len(data))
	if !bytes.HasPrefix(data, []byte("RIFF")) {
		t.Fatalf("output is not WAV: head=%q", data[:min(4, len(data))])
	}
	if peak := wavPeak(data); peak < 1000 {
		t.Fatalf("output is (near-)silent: peak amplitude %d over %d bytes", peak, len(data))
	}
	// then (S9): the replay comes from the cache
	again, err := piper.Speak(ctx, core.English, "ˈnɒl.ɪdʒ")
	if err != nil || !bytes.Equal(again, data) {
		t.Fatalf("replay: len=%d err=%v", len(again), err)
	}
	// (F2 shape): a non-installed language reports unavailable, never speaks
	if piper.Available(core.Spanish) {
		t.Fatal("spanish reported available with only an english voice")
	}
}
