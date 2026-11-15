package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/illescasDaniel/wdict/internal/core"
)

// PiperVoices synthesises speech from model-aligned IPA with the installed
// Piper voice pack (C4). It implements core.PiperVoice: the venv's Python loads
// the voice, merges configured vowel clusters, drops diacritics the model
// inventory lacks (the base phone keeps the sound) while still rejecting
// unknown base phones, and writes a WAV that is audibility-guarded and cached.
type PiperVoices struct {
	Dir      string        // <data>/wdict/piper root: the venv and the voices tree
	CacheDir string        // audio cache directory; empty = no cache
	Timeout  time.Duration // default 10s; the subprocess is killed at the deadline

	runner        func(ctx context.Context, name string, args []string, stdin string) error // default runPiperDefault; replaceable in tests
	testTracePath string                                                                    // optional helper trace for installed-runtime integration tests
	cacheInit     sync.Once
	cache         audioDiskCache
}

const piperKillTimeout = 10 * time.Second

const piperSynthesisScript = `import json
import sys
import unicodedata
import wave

import numpy as np
from piper import PiperVoice, SynthesisConfig

voice = PiperVoice.load(sys.argv[1])
phones = json.load(sys.stdin)
clusters = voice.config.vowel_clusters or set()
max_cluster = max((len(cluster) for cluster in clusters), default=1)
merged = []
i = 0
while i < len(phones):
    match = None
    for size in range(min(max_cluster, len(phones) - i), 1, -1):
        candidate = tuple(phones[i:i + size])
        if candidate in clusters:
            match = candidate
            break
    if match is None:
        merged.append(phones[i])
        i += 1
    else:
        merged.append("".join(match))
        i += len(match)

# A diacritic (Unicode Mn) outside the model's inventory is dropped: the
# base phone before it carries the sound (Wiktionary marks Spanish
# lenition as [saˈβ̞að̞o]; the models have β and ð but no ̞, and the
# retraction marker ̠ is likewise absent). A missing base phone still
# aborts — silently dropping it would corrupt the word.
supported = []
for phone in merged:
    if phone in voice.config.phoneme_id_map:
        supported.append(phone)
    elif unicodedata.category(phone) != "Mn":
        raise ValueError("unsupported Piper phoneme(s): " + repr([phone]))
merged = supported

syn_config = SynthesisConfig()
ids = voice.phonemes_to_ids(merged)
if len(sys.argv) > 3:
    with open(sys.argv[3], "w", encoding="utf-8") as trace:
        json.dump({"phones": merged, "ids": ids}, trace)
audio = voice.phoneme_ids_to_audio(ids, syn_config)
if syn_config.normalize_audio:
    peak = np.max(np.abs(audio))
    audio = np.zeros_like(audio) if peak < 1e-8 else audio / peak
if syn_config.volume != 1.0:
    audio = audio * syn_config.volume
audio = np.clip(audio, -1.0, 1.0).astype(np.float32)
with wave.open(sys.argv[2], "wb") as wav:
    wav.setframerate(voice.config.sample_rate)
    wav.setsampwidth(2)
    wav.setnchannels(1)
    wav.writeframes(np.clip(audio * 32767.0, -32767.0, 32767.0).astype(np.int16).tobytes())
`

// disk returns the shared audio cache; nil when no cache directory is set.
func (p *PiperVoices) disk() *audioDiskCache {
	if p.CacheDir == "" {
		return nil
	}
	p.cacheInit.Do(func() { p.cache = audioDiskCache{Dir: p.CacheDir} })
	return &p.cache
}

func (p *PiperVoices) timeout() time.Duration {
	if p.Timeout > 0 {
		return p.Timeout
	}
	return piperKillTimeout
}

func (p *PiperVoices) venvPython() string {
	return filepath.Join(p.Dir, "venv", "bin", "python3")
}

// voiceFor returns the first installed catalog voice for the language (C5:
// one curated voice per language; catalog order decides between accents).
func (p *PiperVoices) voiceFor(language core.Language) *voicePackEntry {
	for n := range voiceCatalog {
		e := &voiceCatalog[n]
		if e.kind == core.VoicePackVoice && e.language == language && p.voiceInstalled(e.id) {
			return e
		}
	}
	return nil
}

func (p *PiperVoices) voiceInstalled(id core.VoicePackID) bool {
	dir := filepath.Join(p.Dir, "voices", string(id))
	for _, name := range []string{string(id) + ".onnx", string(id) + ".onnx.json"} {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
			return false
		}
	}
	return true
}

// runtimeInstalled reports a healthy venv (C6: a corrupted runtime counts as
// not installed and Available stays false).
func (p *PiperVoices) runtimeInstalled() bool {
	info, err := os.Stat(p.venvPython())
	return err == nil && info.Mode().IsRegular() && info.Size() > 0
}

// Available reports whether the runtime and a voice for the language are
// installed; only then does the chain call Speak.
func (p *PiperVoices) Available(language core.Language) bool {
	return p.runtimeInstalled() && p.voiceFor(language) != nil
}

// Voice names the short voice id used for the language (e.g. "alba"), for the
// Piper result label; empty when no voice is installed.
func (p *PiperVoices) Voice(language core.Language) string {
	if e := p.voiceFor(language); e != nil {
		return e.short
	}
	return ""
}

// Speak synthesises one normalized IPA transcription with the language's voice
// (C4): cache first, then the model-aware subprocess; only audible RIFF WAV
// output is cached and returned. Failures return errors and the core chain
// falls back to eSpeak (F2); failed syntheses are never cached.
func (p *PiperVoices) Speak(ctx context.Context, language core.Language, ipa string) ([]byte, error) {
	entry := p.voiceFor(language)
	if entry == nil {
		return nil, errors.New("piper: no voice installed for the language")
	}
	if strings.TrimSpace(ipa) == "" {
		return nil, errors.New("piper: no pronunciation text")
	}
	key := fmt.Sprintf("piper:v3:%s:%s", entry.id, ipa)
	if cache := p.disk(); cache != nil {
		if data, ok := cache.get(key); ok {
			return data, nil
		}
	}
	temp, err := os.CreateTemp("", "piper-*.wav")
	if err != nil {
		return nil, err
	}
	wavPath := temp.Name()
	if err = temp.Close(); err != nil {
		return nil, err
	}
	defer os.Remove(wavPath)
	run := p.runner
	if run == nil {
		run = p.runDefault
	}
	phones := make([]string, 0, len(ipa))
	for _, r := range ipa {
		phones = append(phones, string(r))
	}
	input, err := json.Marshal(phones)
	if err != nil {
		return nil, fmt.Errorf("encoding Piper phones: %w", err)
	}
	args := []string{"-c", piperSynthesisScript, p.voicePath(entry.id), wavPath}
	if p.testTracePath != "" {
		args = append(args, p.testTracePath)
	}
	if err = run(ctx, p.venvPython(), args, string(input)); err != nil {
		return nil, err
	}
	info, err := os.Stat(wavPath)
	if err != nil {
		return nil, fmt.Errorf("piper: no output produced: %w", err)
	}
	if info.Size() > maxWAVBytes {
		return nil, fmt.Errorf("piper: output too large (over %d bytes)", maxWAVBytes)
	}
	data, err := os.ReadFile(wavPath)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("piper: no output")
	}
	if !bytes.HasPrefix(data, []byte("RIFF")) {
		return nil, errors.New("piper: output is not WAV")
	}
	if peak := wavPeak(data); peak < 1000 { // silence decodes to peak 0; audible speech is far above
		return nil, fmt.Errorf("piper: output is (near-)silent: peak amplitude %d over %d bytes", peak, len(data))
	}
	if cache := p.disk(); cache != nil {
		if err := cache.put(key, data); err != nil {
			return nil, fmt.Errorf("caching audio: %w", err)
		}
	}
	return data, nil
}

func (p *PiperVoices) voicePath(id core.VoicePackID) string {
	return filepath.Join(p.Dir, "voices", string(id), string(id)+".onnx")
}

// runDefault runs the piper subprocess with the kill timeout: the IPA text is
// written to stdin (which is closed after the write) and the WAV goes to the
// -f output path.
func (p *PiperVoices) runDefault(ctx context.Context, name string, args []string, stdin string) error {
	ctx, cancel := context.WithTimeout(ctx, p.timeout())
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.WaitDelay = p.timeout()
	cmd.Env = piperEnv(os.Environ())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin) // EOF closes the pipe after the text
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

// wavPeak returns the WAV's peak absolute 16-bit sample amplitude: the
// audibility guard shared with the eSpeak integration checks.
func wavPeak(data []byte) int {
	peak := 0
	for i := 44; i+1 < len(data); i += 2 {
		sample := int(int16(uint16(data[i]) | uint16(data[i+1])<<8))
		if sample < 0 {
			sample = -sample
		}
		if sample > peak {
			peak = sample
		}
	}
	return peak
}

// piperEnv returns the environment for the Piper Python subprocess. Inside an
// AppImage the launcher points LD_LIBRARY_PATH (and may set PYTHONHOME,
// PYTHONPATH and LD_PRELOAD) at the bundled GTK/WebKit libraries; the venv's
// interpreter is the host's python3 and must resolve its own libraries, so
// those AppDir-relative settings are dropped for it. Outside an AppImage
// (APPDIR unset) the environment is returned unchanged.
func piperEnv(environ []string) []string {
	appDir := ""
	for _, kv := range environ {
		if value, ok := strings.CutPrefix(kv, "APPDIR="); ok {
			appDir = strings.TrimRight(value, "/")
		}
	}
	if appDir == "" {
		return environ
	}
	out := make([]string, 0, len(environ))
	for _, kv := range environ {
		name, value, _ := strings.Cut(kv, "=")
		switch name {
		case "PYTHONHOME", "PYTHONPATH", "LD_PRELOAD":
			continue
		case "LD_LIBRARY_PATH":
			var keep []string
			for _, dir := range strings.Split(value, ":") {
				if dir != "" && dir != appDir && !strings.HasPrefix(dir, appDir+"/") {
					keep = append(keep, dir)
				}
			}
			if len(keep) == 0 {
				continue
			}
			kv = "LD_LIBRARY_PATH=" + strings.Join(keep, ":")
		}
		out = append(out, kv)
	}
	return out
}
