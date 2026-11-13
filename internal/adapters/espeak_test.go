package adapters

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type fakeRunner struct {
	name string
	args []string
	data []byte
	err  error
}

func (f *fakeRunner) run(_ context.Context, name string, args []string) ([]byte, error) {
	f.name, f.args = name, args
	if f.err != nil {
		return nil, f.err
	}
	return f.data, nil
}

func newFakeEspeak(data []byte) (*EspeakVoices, *fakeRunner) {
	runner := &fakeRunner{data: data}
	return &EspeakVoices{runner: runner.run}, runner
}

func TestGivenPlainTextWhenSpokenThenNoPhonemeElementIsUsed(t *testing.T) {
	// given (S2/C5, amended): the voice path speaks the headword as plain
	// text, with markup characters escaped in the text node (the SSML
	// phoneme mechanism is withdrawn — it renders silence on espeak-ng 1.52)
	espeak, runner := newFakeEspeak([]byte("RIFF...."))
	// when
	data, err := espeak.SpeakText(context.Background(), "en-gb", `a<b>&"c`)
	// then
	if err != nil || string(data) != "RIFF...." {
		t.Fatalf("data=%q err=%v", data, err)
	}
	want := []string{"-v", "en-gb", "-m", `<speak>a&lt;b&gt;&amp;&quot;c</speak>`, "--stdout"}
	if strings.Join(runner.args, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("args=%q", runner.args)
	}
	// empty voice or text starts no process
	empty, emptyRunner := newFakeEspeak(nil)
	if _, err := empty.SpeakText(context.Background(), "", "word"); err == nil {
		t.Fatal("empty voice accepted")
	}
	if _, err := empty.SpeakText(context.Background(), "en-gb", "  "); err == nil {
		t.Fatal("empty text accepted")
	}
	if emptyRunner.name != "" {
		t.Fatalf("runner called: %v %v", emptyRunner.name, emptyRunner.args)
	}
}

func TestGivenMnemonicStreamWhenSpokenThenRawTextModeIsUsed(t *testing.T) {
	// given (S2/C5, revision 3): mnemonic input works as plain text on
	// espeak-ng 1.52 — the SSML markup mode (-m) renders silence for phoneme
	// input, so the phonemes are passed raw, wrapped in [[…]]
	espeak, runner := newFakeEspeak([]byte("RIFF...."))
	// when
	data, err := espeak.SpeakPhonemes(context.Background(), "en-gb", "w3:d")
	// then
	if err != nil || string(data) != "RIFF...." {
		t.Fatalf("data=%q err=%v", data, err)
	}
	want := []string{"-v", "en-gb", "[[w3:d]]", "--stdout"} // no -m
	if strings.Join(runner.args, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("args=%q", runner.args)
	}
	espeak, runner = newFakeEspeak([]byte("RIFF...."))
	if _, err := espeak.SpeakPhonemes(context.Background(), "es", "n^ l^ J^ R2"); err != nil {
		t.Fatal(err)
	}
	want = []string{"-v", "es", "[[n^ l^ J^ R2]]", "--stdout"}
	if strings.Join(runner.args, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("args=%q", runner.args)
	}
}

func TestGivenUnsupportedMnemonicsWhenSpokenThenNoProcessIsStarted(t *testing.T) {
	// (C5, revision 3): defense in depth — only the characters the core
	// converter can emit pass the whitelist; everything else is rejected
	// before the subprocess starts
	for _, tc := range []struct{ voice, mnemonics string }{
		{"", "w3:d"},            // no voice
		{"en-gb", ""},           // no mnemonics
		{"en-gb", "   "},        // whitespace only
		{"en-gb", "w3:d$(x)"},   // shell-like input
		{"en-gb", "w3:d;ls"},    // separator
		{"en-gb", "w3:d\nw3:d"}, // newline
		{"en-gb", "w3:d|ls"},    // pipe
	} {
		espeak, runner := newFakeEspeak([]byte("RIFF...."))
		if _, err := espeak.SpeakPhonemes(context.Background(), tc.voice, tc.mnemonics); err == nil {
			t.Fatalf("accepted %q/%q", tc.voice, tc.mnemonics)
		}
		if runner.name != "" {
			t.Fatalf("runner called for %q/%q: %v %v", tc.voice, tc.mnemonics, runner.name, runner.args)
		}
	}
}

func TestGivenPhonemeRunnerFailureOrEmptyOutputWhenSpokenThenAnErrorIsReturned(t *testing.T) {
	// (F4 conditions observed by the core service)
	espeak, _ := newFakeEspeak(nil)
	if _, err := espeak.SpeakPhonemes(context.Background(), "en-us", "w3:d"); err == nil {
		t.Fatal("empty output accepted")
	}
	runner := &fakeRunner{err: errors.New("exit status 1")}
	espeak = &EspeakVoices{runner: runner.run}
	if _, err := espeak.SpeakPhonemes(context.Background(), "en-us", "w3:d"); err == nil {
		t.Fatal("runner error swallowed")
	}
}

func TestGivenEmptyVoiceOrTextWhenSpokenThenNoProcessIsStarted(t *testing.T) {
	espeak, runner := newFakeEspeak(nil)
	if _, err := espeak.SpeakText(context.Background(), "", "word"); err == nil {
		t.Fatal("empty voice accepted")
	}
	if _, err := espeak.SpeakText(context.Background(), "en-us", "  "); err == nil {
		t.Fatal("empty text accepted")
	}
	if runner.name != "" {
		t.Fatalf("runner called: %v %v", runner.name, runner.args)
	}
}

func TestGivenRunnerFailureOrEmptyOutputWhenSpokenThenAnErrorIsReturned(t *testing.T) {
	// (F4 conditions observed by the core service)
	espeak, _ := newFakeEspeak(nil)
	if _, err := espeak.SpeakText(context.Background(), "en-us", "word"); err == nil {
		t.Fatal("empty output accepted")
	}
	runner := &fakeRunner{err: errors.New("exit status 1")}
	espeak = &EspeakVoices{runner: runner.run}
	if _, err := espeak.SpeakText(context.Background(), "en-us", "word"); err == nil {
		t.Fatal("runner error swallowed")
	}
}

func TestGivenCustomBinaryNameWhenSpokenThenItIsUsed(t *testing.T) {
	espeak, runner := newFakeEspeak([]byte("RIFF...."))
	espeak.Binary = "espeak-ng-custom"
	if _, err := espeak.SpeakText(context.Background(), "en-gb", "word"); err != nil {
		t.Fatal(err)
	}
	if runner.name != "espeak-ng-custom" {
		t.Fatalf("binary=%q", runner.name)
	}
}

func TestAvailableUsesLookPath(t *testing.T) {
	// given
	failing := &EspeakVoices{lookPath: func(string) (string, error) { return "", exec.ErrNotFound }}
	succeeding := &EspeakVoices{lookPath: func(string) (string, error) { return "/usr/bin/espeak-ng", nil }}
	// then
	if failing.Available() {
		t.Fatal("unavailable binary reported available")
	}
	if !succeeding.Available() {
		t.Fatal("available binary reported unavailable")
	}
}

func TestGivenRunnerOutputOverTheWAVCapWhenRunThenItAbortsAtTheCap(t *testing.T) {
	// given (C5): the real runner reads subprocess output through the cap
	e := &EspeakVoices{}
	// when: 5 MiB of output from a real helper process
	_, err := e.runDefault(context.Background(), "sh", []string{"-c", "head -c 5000000 /dev/zero"})
	// then
	if err == nil || !strings.Contains(err.Error(), "too large") {
		t.Fatalf("over-cap output accepted: %v", err)
	}
}

func TestGivenRunnerOutputThatIsNotWAVWhenRunThenItIsRejected(t *testing.T) {
	e := &EspeakVoices{}
	if _, err := e.runDefault(context.Background(), "sh", []string{"-c", "head -c 100 /dev/zero"}); err == nil {
		t.Fatal("non-WAV output accepted")
	}
}

func TestGivenHangingSubprocessWhenRunThenItIsKilledAtTheTimeout(t *testing.T) {
	// given (C5): subprocess killed at 5 s; the test shrinks the deadline
	e := &EspeakVoices{Timeout: 100 * time.Millisecond}
	start := time.Now()
	// when
	_, err := e.runDefault(context.Background(), "sh", []string{"-c", "sleep 10"})
	// then
	if err == nil {
		t.Fatal("hanging subprocess produced output")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("took %v; subprocess was not killed at the timeout", elapsed)
	}
}
