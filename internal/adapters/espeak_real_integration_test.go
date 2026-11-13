//go:build integration

package adapters

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/illescasDaniel/wdict/internal/core"
)

func TestGivenRealEspeakNGWhenSpokenThenAudibleRIFFWAVBytesComeBack(t *testing.T) {
	// given: the real binary, verified opt-in (C5 build check, amended
	// 2026-11-13: the synthesis must be AUDIBLE — the SSML phoneme mechanism
	// produced valid-length silent WAVs)
	if _, err := exec.LookPath("espeak-ng"); err != nil {
		t.Skip("espeak-ng not installed")
	}
	espeak := &EspeakVoices{}
	if !espeak.Available() {
		t.Fatal("Available is false although the binary exists")
	}
	// when
	data, err := espeak.SpeakText(context.Background(), "en-us", "word")
	// then
	if err != nil {
		t.Fatalf("speak: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("RIFF")) || len(data) > 4*1024*1024 {
		t.Fatalf("output len=%d head=%q", len(data), data[:min(4, len(data))])
	}
	if peak := wavPeak(data); peak < 1000 { // silence decodes to peak 0; audible speech is far above
		t.Fatalf("output is (near-)silent: peak amplitude %d over %d bytes", peak, len(data))
	}
}

func TestGivenRealEspeakNGWhenPhonemesAreSpokenThenAudibleRIFFWAVBytesComeBack(t *testing.T) {
	// given (C5, revision 3): mnemonic input works as plain text —
	// "[[w3:d]]" must produce an AUDIBLE WAV without the -m flag
	if _, err := exec.LookPath("espeak-ng"); err != nil {
		t.Skip("espeak-ng not installed")
	}
	espeak := &EspeakVoices{}
	// when
	data, err := espeak.SpeakPhonemes(context.Background(), "en-gb", "w3:d")
	// then
	if err != nil {
		t.Fatalf("speak phonemes: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("RIFF")) || len(data) > 4*1024*1024 {
		t.Fatalf("output len=%d head=%q", len(data), data[:min(4, len(data))])
	}
	if peak := wavPeak(data); peak < 1000 {
		t.Fatalf("mnemonic output is (near-)silent: peak amplitude %d over %d bytes", peak, len(data))
	}
}

func TestGivenRealEspeakNGWhenVerifyingTheMappingTableThenEveryEntryEchoes(t *testing.T) {
	// (C5, revision 3): the tables are sourced from the engine's own
	// `espeak-ng -v <voice> -q -x` output; this oracle keeps every entry
	// honest against engine changes. An entry that fails to echo must fail
	// the test.
	if _, err := exec.LookPath("espeak-ng"); err != nil {
		t.Skip("espeak-ng not installed")
	}
	for _, tc := range []struct {
		language core.Language
		voice    string
	}{
		{core.English, "en-gb"},
		{core.Spanish, "es"},
	} {
		table, ok := core.MnemonicTable(tc.language)
		if !ok {
			t.Fatalf("no mnemonic table for %s", tc.language)
		}
		for ipa, mnemonic := range table {
			t.Run(string(tc.language)+"/"+ipa, func(t *testing.T) {
				out, err := exec.Command("espeak-ng", "-v", tc.voice, "-q", "-x", "[["+mnemonic+"]]").Output()
				if err != nil {
					t.Fatalf("%q: %v", mnemonic, err)
				}
				// the engine may insert a default stress mark ' and canonicalise
				// the linking r as ';' (I@ → I;@); both are ignored
				echo := strings.NewReplacer("'", "", ";", "").Replace(strings.TrimSpace(string(out)))
				if echo != mnemonic {
					t.Fatalf("%q echoed %q", mnemonic, echo)
				}
			})
		}
	}
}

// wavPeak lives in piper.go now (production audibility guard, shared with the
// Piper integration checks).
