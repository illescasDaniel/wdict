package adapters

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/illescasDaniel/wdict/internal/core"
)

// fakePiperRunner stands in for the piper subprocess (C4 seams). When wav is
// set it writes the bytes to the command's -f output path, emulating piper.
type fakePiperRunner struct {
	calls int
	name  string
	args  []string
	stdin string
	err   error
	wav   []byte
}

func (f *fakePiperRunner) run(_ context.Context, name string, args []string, stdin string) error {
	f.calls++
	f.name, f.args, f.stdin = name, args, stdin
	if f.wav != nil && len(args) > 0 {
		return os.WriteFile(args[len(args)-1], f.wav, 0600)
	}
	return f.err
}

// fakeWav builds a minimal audible (or, with a small peak, near-silent) RIFF
// WAV for the guard tests.
func fakeWav(peak int) []byte {
	data := make([]byte, 48)
	copy(data, "RIFF")
	binary.LittleEndian.PutUint32(data[4:], 40)
	copy(data[8:], "WAVE")
	copy(data[12:], "fmt ")
	binary.LittleEndian.PutUint32(data[16:], 16)
	binary.LittleEndian.PutUint16(data[20:], 1) // PCM
	binary.LittleEndian.PutUint16(data[22:], 1) // mono
	binary.LittleEndian.PutUint32(data[24:], 22050)
	binary.LittleEndian.PutUint32(data[28:], 44100)
	binary.LittleEndian.PutUint16(data[32:], 2)
	binary.LittleEndian.PutUint16(data[34:], 16)
	copy(data[36:], "data")
	binary.LittleEndian.PutUint32(data[40:], 4)
	binary.LittleEndian.PutUint16(data[44:], uint16(peak))
	binary.LittleEndian.PutUint16(data[46:], uint16(int16(-peak)))
	return data
}

func newTestPiper(t *testing.T, runner *fakePiperRunner, cacheDir string) *PiperVoices {
	t.Helper()
	dir := t.TempDir()
	pretendRuntime(t, dir)
	pretendVoice(t, dir, "en_GB-alba-medium")
	return &PiperVoices{Dir: dir, CacheDir: cacheDir, runner: runner.run}
}

func TestGivenVoicePackInstalledWhenSpokenThenTheSubprocessUsesModelPhones(t *testing.T) {
	// given: the adapter sends one phone per Unicode scalar to a Python helper
	// that merges the model's configured vowel clusters and rejects unknowns.
	runner := &fakePiperRunner{wav: fakeWav(2000)}
	p := newTestPiper(t, runner, t.TempDir())
	// when
	data, err := p.Speak(context.Background(), core.English, "nˈɒlɪdʒ")
	// then
	if err != nil || string(data) != string(fakeWav(2000)) {
		t.Fatalf("data len=%d err=%v", len(data), err)
	}
	if runner.calls != 1 {
		t.Fatalf("runner calls=%d", runner.calls)
	}
	if runner.name != filepath.Join(p.Dir, "venv", "bin", "python3") {
		t.Fatalf("name=%q", runner.name)
	}
	if len(runner.args) != 4 || runner.args[0] != "-c" ||
		!strings.Contains(runner.args[1], "voice.config.vowel_clusters") ||
		runner.args[2] != filepath.Join(p.Dir, "voices", "en_GB-alba-medium", "en_GB-alba-medium.onnx") ||
		!strings.HasSuffix(runner.args[3], ".wav") {
		t.Fatalf("args=%q", runner.args)
	}
	if runner.stdin != "[\"n\",\"ˈ\",\"ɒ\",\"l\",\"ɪ\",\"d\",\"ʒ\"]" {
		t.Fatalf("stdin=%q", runner.stdin)
	}
	// no shell
	if runner.name == "sh" || runner.name == "bash" || containsAll(runner.args, "sh", "-c") {
		t.Fatalf("shell involved: %q %q", runner.name, runner.args)
	}
	// Cache key includes the synthesis pipeline version.
	key := fmt.Sprintf("piper:v2:%s:%s", "en_GB-alba-medium", "nˈɒlɪdʒ")
	sum := sha256.Sum256([]byte(key))
	if _, err := os.Stat(filepath.Join(p.CacheDir, hex.EncodeToString(sum[:]))); err != nil {
		t.Fatalf("cache miss after synthesis: %v", err)
	}
}

func containsAll(args []string, want ...string) bool {
	for _, w := range want {
		found := false
		for _, a := range args {
			if a == w {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestGivenCachedResultWhenSpokenThenNoSubprocessRuns(t *testing.T) {
	// given (S9): a result cached under sha256("piper:v2:<voice id>:<ipa>")
	runner := &fakePiperRunner{wav: fakeWav(2000)}
	cacheDir := t.TempDir()
	p := newTestPiper(t, runner, cacheDir)
	ipa := "nˈɒlɪdʒ"
	key := fmt.Sprintf("piper:v2:%s:%s", "en_GB-alba-medium", ipa)
	sum := sha256.Sum256([]byte(key))
	if err := os.WriteFile(filepath.Join(cacheDir, hex.EncodeToString(sum[:])), fakeWav(3000), 0600); err != nil {
		t.Fatal(err)
	}
	// when
	data, err := p.Speak(context.Background(), core.English, ipa)
	// then: served from cache, no subprocess, no cache write
	if err != nil || string(data) != string(fakeWav(3000)) {
		t.Fatalf("data len=%d err=%v", len(data), err)
	}
	if runner.calls != 0 {
		t.Fatalf("runner called: %v %v", runner.name, runner.args)
	}
}

func TestGivenLegacyCacheEntryWhenSpokenThenTheUpdatedPipelineSynthesisesAgain(t *testing.T) {
	// given: an old pipeline's cache file exists for the same normalized input
	runner := &fakePiperRunner{wav: fakeWav(2000)}
	cacheDir := t.TempDir()
	p := newTestPiper(t, runner, cacheDir)
	ipa := "nˈɒlɪdʒ"
	legacyKey := fmt.Sprintf("piper:%s:%s", "en_GB-alba-medium", ipa)
	legacySum := sha256.Sum256([]byte(legacyKey))
	if err := os.WriteFile(filepath.Join(cacheDir, hex.EncodeToString(legacySum[:])), fakeWav(3000), 0600); err != nil {
		t.Fatal(err)
	}
	// when
	data, err := p.Speak(context.Background(), core.English, ipa)
	// then
	if err != nil || string(data) != string(fakeWav(2000)) {
		t.Fatalf("data len=%d err=%v", len(data), err)
	}
	if runner.calls != 1 {
		t.Fatalf("runner calls=%d", runner.calls)
	}
}

func TestGivenFailedSynthesisWhenSpokenThenAnErrorIsReturnedAndNothingIsCached(t *testing.T) {
	// given (S4/C4): failed syntheses are never cached
	runner := &fakePiperRunner{err: errors.New("exit status 1")}
	p := newTestPiper(t, runner, t.TempDir())
	if _, err := p.Speak(context.Background(), core.English, "nˈɒlɪdʒ"); err == nil {
		t.Fatal("failed synthesis accepted")
	}
	entries, _ := os.ReadDir(p.CacheDir)
	if len(entries) != 0 {
		t.Fatalf("cache files=%v", entries)
	}
	// when: the subprocess recovers, the next play synthesises again
	runner.err = nil
	runner.wav = fakeWav(2000)
	if _, err := p.Speak(context.Background(), core.English, "nˈɒlɪdʒ"); err != nil {
		t.Fatal(err)
	}
	if runner.calls != 2 {
		t.Fatalf("runner calls=%d", runner.calls)
	}
}

func TestGivenNonWAVOrEmptyOrSilentOutputWhenSpokenThenAnErrorIsReturned(t *testing.T) {
	// (S4/C4 audibility guard): non-WAV, empty and silent outputs are errors.
	for name, tc := range map[string]struct {
		wav  []byte
		want string
	}{
		"not wav":      {wav: []byte("plain text, definitely not RIFF"), want: "not WAV"},
		"empty output": {wav: nil, want: ""},
		"silent":       {wav: fakeWav(5), want: "silent"},
	} {
		t.Run(name, func(t *testing.T) {
			runner := &fakePiperRunner{wav: tc.wav}
			p := newTestPiper(t, runner, t.TempDir())
			_, err := p.Speak(context.Background(), core.English, "nˈɒlɪdʒ")
			if err == nil {
				t.Fatal("invalid output accepted")
			}
			if tc.want != "" && !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err=%v", err)
			}
			entries, _ := os.ReadDir(p.CacheDir)
			if len(entries) != 0 {
				t.Fatalf("invalid output cached: %v", entries)
			}
		})
	}
}

func TestGivenNoVoiceForTheLanguageWhenSpokenThenNoSubprocessStarts(t *testing.T) {
	// (F1/C4): no installed voice, no process, an error the chain falls back on.
	runner := &fakePiperRunner{}
	dir := t.TempDir()
	pretendRuntime(t, dir)
	p := &PiperVoices{Dir: dir, CacheDir: t.TempDir(), runner: runner.run}
	if _, err := p.Speak(context.Background(), core.Spanish, "ˈola"); err == nil {
		t.Fatal("missing voice accepted")
	}
	if runner.calls != 0 {
		t.Fatalf("runner called: %v", runner.args)
	}
	if _, err := p.Speak(context.Background(), core.English, "  "); err == nil {
		t.Fatal("empty IPA accepted")
	}
	if runner.calls != 0 {
		t.Fatalf("runner called: %v", runner.args)
	}
}

func TestGivenLanguagesWhenVoicedThenTheFirstInstalledCatalogVoiceIsUsed(t *testing.T) {
	// (C5/C6): Available requires the runtime and a voice for the language;
	// the voice label is the short catalog name, first installed wins.
	dir := t.TempDir()
	p := &PiperVoices{Dir: dir}
	if p.Available(core.English) {
		t.Fatal("empty dir reported available")
	}
	pretendRuntime(t, dir)
	pretendVoice(t, dir, "en_US-amy-medium")
	if !p.Available(core.English) || p.Voice(core.English) != "amy" {
		t.Fatalf("english: available=%v voice=%q", p.Available(core.English), p.Voice(core.English))
	}
	if p.Available(core.Spanish) || p.Voice(core.Spanish) != "" {
		t.Fatalf("spanish: available=%v voice=%q", p.Available(core.Spanish), p.Voice(core.Spanish))
	}
	pretendVoice(t, dir, "es_ES-sharvard-medium")
	if !p.Available(core.Spanish) || p.Voice(core.Spanish) != "sharvard" {
		t.Fatalf("spanish: available=%v voice=%q", p.Available(core.Spanish), p.Voice(core.Spanish))
	}
	// catalog order: alba precedes amy for English
	pretendVoice(t, dir, "en_GB-alba-medium")
	if p.Voice(core.English) != "alba" {
		t.Fatalf("english voice=%q", p.Voice(core.English))
	}
	// (C6): a corrupted runtime (venv python missing) counts as not installed
	if err := os.Remove(filepath.Join(dir, "venv", "bin", "python3")); err != nil {
		t.Fatal(err)
	}
	if p.Available(core.English) {
		t.Fatal("corrupted runtime reported available")
	}
}

func TestGivenHangingPiperWhenSpokenThenItIsKilledAtTheTimeout(t *testing.T) {
	// given (C4): the subprocess is killed at 10 s; the test shrinks the deadline
	p := &PiperVoices{Timeout: 100 * time.Millisecond}
	start := time.Now()
	// when
	err := p.runDefault(context.Background(), "sh", []string{"-c", "sleep 10"}, "stdin text")
	// then
	if err == nil {
		t.Fatal("hanging subprocess produced output")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("took %v; subprocess was not killed at the timeout", elapsed)
	}
}

func TestGivenRealRunnerWhenTheProcessFailsThenStderrReachesTheError(t *testing.T) {
	// (C4 diagnostics): the runner's errors carry the subprocess stderr.
	p := &PiperVoices{}
	err := p.runDefault(context.Background(), "sh", []string{"-c", "echo piper-boom >&2; exit 3"}, "")
	if err == nil || !strings.Contains(err.Error(), "piper-boom") {
		t.Fatalf("err=%v", err)
	}
}
