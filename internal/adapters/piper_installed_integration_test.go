//go:build integration

package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/illescasDaniel/wdict/internal/core"
)

// TestGivenInstalledPiperModelsWhenDictionaryIPAIsSpokenThenModelPhonesAreUsed
// uses only an existing runtime/model pack and never downloads.
func TestGivenInstalledPiperModelsWhenDictionaryIPAIsSpokenThenModelPhonesAreUsed(t *testing.T) {
	dir := os.Getenv("WDICT_PIPER_TEST_DIR")
	if dir == "" {
		t.Skip("set WDICT_PIPER_TEST_DIR to an installed Piper data directory")
	}
	piper := &PiperVoices{Dir: dir, CacheDir: filepath.Join(t.TempDir(), "audio")}
	cases := []struct {
		word       string
		language   core.Language
		ipa        string
		wantPhones []string
		wantIDs    []int
	}{
		{
			"vegetable", core.English, "ˈvɛd͡ʒ.tə.bəl",
			[]string{"v", "ˈ", "ɛ", "d", "ʒ", "t", "ə", "b", "ə", "l"},
			[]int{1, 0, 34, 0, 120, 0, 61, 0, 17, 0, 108, 0, 32, 0, 59, 0, 15, 0, 59, 0, 24, 0, 2},
		},
		{
			"table", core.English, "ˈteɪbl̩",
			[]string{"t", "ˈ", "eɪ", "b", "ə", "l"},
			[]int{1, 0, 32, 0, 120, 0, 164, 0, 15, 0, 59, 0, 24, 0, 2},
		},
		{
			"Enrique", core.Spanish, "ẽnˈrike",
			[]string{"e", "̃", "n", "r", "ˈ", "i", "k", "e"},
			[]int{1, 0, 18, 0, 141, 0, 26, 0, 30, 0, 120, 0, 21, 0, 23, 0, 18, 0, 2},
		},
	}
	for _, tc := range cases {
		t.Run(tc.word, func(t *testing.T) {
			if !piper.Available(tc.language) {
				t.Skipf("no installed voice for %s", tc.language)
			}
			ipa, ok := core.NormalizePiperIPA(tc.language, tc.ipa)
			if !ok {
				t.Fatalf("IPA %q was rejected", tc.ipa)
			}
			tracePath := filepath.Join(t.TempDir(), "phones.json")
			piper.testTracePath = tracePath
			data, err := piper.Speak(context.Background(), tc.language, ipa)
			if err != nil {
				t.Fatalf("speak %s (%s): %v", tc.word, ipa, err)
			}
			if !bytes.HasPrefix(data, []byte("RIFF")) || wavPeak(data) < 1000 {
				t.Fatalf("%s returned invalid or silent WAV (%d bytes)", tc.word, len(data))
			}
			var trace struct {
				Phones []string `json:"phones"`
				IDs    []int    `json:"ids"`
			}
			traceData, err := os.ReadFile(tracePath)
			if err != nil {
				t.Fatalf("read model phone trace: %v", err)
			}
			if err := json.Unmarshal(traceData, &trace); err != nil {
				t.Fatalf("unmarshal model phone trace: %v", err)
			}
			if strings.Join(trace.Phones, "\x00") != strings.Join(tc.wantPhones, "\x00") {
				t.Fatalf("model phones for %s = %q, want %q (ids %v)", tc.word, trace.Phones, tc.wantPhones, trace.IDs)
			}
			if !reflect.DeepEqual(trace.IDs, tc.wantIDs) {
				t.Fatalf("model phone IDs for %s = %v, want %v", tc.word, trace.IDs, tc.wantIDs)
			}
			t.Logf("%s model phones %q -> IDs %v", tc.word, trace.Phones, trace.IDs)
			cached, err := piper.Speak(context.Background(), tc.language, ipa)
			if err != nil || !bytes.Equal(cached, data) {
				t.Fatalf("%s cache replay differs: %v", tc.word, err)
			}
		})
	}
	t.Run("unsupported phone rejected", func(t *testing.T) {
		piper := &PiperVoices{Dir: dir, CacheDir: filepath.Join(t.TempDir(), "audio")}
		_, err := piper.Speak(context.Background(), core.English, "\ue000")
		if err == nil || !strings.Contains(err.Error(), "unsupported Piper phoneme") {
			t.Fatalf("unsupported phone error = %v", err)
		}
		entries, readErr := os.ReadDir(piper.CacheDir)
		if (readErr != nil && !os.IsNotExist(readErr)) || len(entries) != 0 {
			t.Fatalf("unsupported phone cache entries=%v err=%v", entries, readErr)
		}
	})
}
