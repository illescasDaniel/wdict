//go:build integration

package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/illescasDaniel/wdict/internal/core"
)

// Runs only when the caller supplies an existing local French model and Piper runtime.
func TestGivenFrenchPiperWhenIPAIsSpokenThenExpectedPhonesReachModel(t *testing.T) {
	runtime, model := os.Getenv("WDICT_PIPER_TEST_DIR"), os.Getenv("WDICT_FRENCH_MODEL")
	if runtime == "" || model == "" {
		t.Skip("set WDICT_PIPER_TEST_DIR and WDICT_FRENCH_MODEL")
	}
	dir := t.TempDir()
	if err := os.Symlink(filepath.Join(runtime, "venv"), filepath.Join(dir, "venv")); err != nil {
		t.Fatal(err)
	}
	const id = "fr_FR-siwis-medium"
	voiceDir := filepath.Join(dir, "voices", id)
	if err := os.MkdirAll(voiceDir, 0700); err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"", ".json"} {
		if err := os.Symlink(model+suffix, filepath.Join(voiceDir, id+".onnx"+suffix)); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		name, ipa string
		phones    []string
	}{
		{"nasal", "bɔ̃ʒuʁ", []string{"b", "ɔ", "̃", "ʒ", "ˈ", "u", "ʁ"}},
		{"palatal", "aɲo", []string{"a", "n", "j", "ˈ", "o"}},
		{"glide", "lɥi", []string{"l", "y", "ˈ", "i"}},
		{"phrase", "le.z‿a.mi", []string{"l", "e", "z", "a", "m", "i"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			normalized, ok := core.NormalizePiperIPA(core.French, tc.ipa)
			if !ok {
				t.Fatalf("rejected %q", tc.ipa)
			}
			trace := filepath.Join(t.TempDir(), "trace.json")
			p := &PiperVoices{Dir: dir, testTracePath: trace}
			wav, err := p.Speak(context.Background(), core.French, normalized)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.HasPrefix(wav, []byte("RIFF")) || wavPeak(wav) < 1000 {
				t.Fatal("invalid or silent audio")
			}
			b, err := os.ReadFile(trace)
			if err != nil {
				t.Fatal(err)
			}
			var got struct {
				Phones []string `json:"phones"`
				IDs    []int    `json:"ids"`
			}
			if err = json.Unmarshal(b, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Phones, tc.phones) {
				t.Fatalf("phones=%q want=%q ids=%v", got.Phones, tc.phones, got.IDs)
			}
			t.Logf("phones=%q ids=%v wav=%d bytes", got.Phones, got.IDs, len(wav))
		})
	}
}
