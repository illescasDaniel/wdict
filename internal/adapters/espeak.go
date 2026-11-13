package adapters

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// EspeakVoices synthesises speech with the system espeak-ng binary (C5).
// It implements core.VoiceSynthesiser.
type EspeakVoices struct {
	Binary  string        // default "espeak-ng"
	Timeout time.Duration // default 5s; the subprocess is killed at the deadline

	lookPath func(string) (string, error)                                          // default exec.LookPath; replaceable in tests
	runner   func(ctx context.Context, name string, args []string) ([]byte, error) // default runDefault; replaceable in tests
}

const (
	espeakDefaultBinary = "espeak-ng"
	espeakKillTimeout   = 5 * time.Second
	maxWAVBytes         = 4 * 1024 * 1024
)

func (e *EspeakVoices) binary() string {
	if e.Binary != "" {
		return e.Binary
	}
	return espeakDefaultBinary
}

func (e *EspeakVoices) timeout() time.Duration {
	if e.Timeout > 0 {
		return e.Timeout
	}
	return espeakKillTimeout
}

// Available reports whether the eSpeak binary exists on the PATH.
func (e *EspeakVoices) Available() bool {
	look := e.lookPath
	if look == nil {
		look = exec.LookPath
	}
	_, err := look(e.binary())
	return err == nil
}

// SpeakText speaks plain text (C5, revision 3: the headword fallback of the
// voice path) and returns the WAV bytes. Errors are expected outcomes for the
// core fallback chain (F4).
func (e *EspeakVoices) SpeakText(ctx context.Context, voice, text string) ([]byte, error) {
	if strings.TrimSpace(voice) == "" {
		return nil, errors.New("espeak: no voice")
	}
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("espeak: no pronunciation text")
	}
	return e.run(ctx, voice, "-m", espeakTextSSML(text))
}

// espeakMnemonicCharset matches exactly the characters core.IPAToMnemonics can
// emit (defense in depth: the subprocess input is whitelisted at the edge).
var espeakMnemonicCharset = regexp.MustCompile(`^[A-Za-z0-9:'@^#\[\] ]+$`)

// SpeakPhonemes speaks eSpeak mnemonic phonemes (C5, revision 3) and returns
// the WAV bytes. The input is passed as raw text wrapped in [[…]] — without
// the -m flag, which renders silence for phoneme input on the real engine.
// The mnemonics are whitelisted at the edge; anything the core converter
// cannot emit is rejected before the subprocess starts.
func (e *EspeakVoices) SpeakPhonemes(ctx context.Context, voice, mnemonics string) ([]byte, error) {
	if strings.TrimSpace(voice) == "" {
		return nil, errors.New("espeak: no voice")
	}
	if strings.TrimSpace(mnemonics) == "" {
		return nil, errors.New("espeak: no phoneme mnemonics")
	}
	if !espeakMnemonicCharset.MatchString(mnemonics) {
		return nil, errors.New("espeak: mnemonics contain unsupported characters")
	}
	return e.run(ctx, voice, "[["+mnemonics+"]]")
}

// run executes one synthesis command with the shared runner and output rules.
func (e *EspeakVoices) run(ctx context.Context, voice string, args ...string) ([]byte, error) {
	run := e.runner
	if run == nil {
		run = e.runDefault
	}
	full := append(append([]string{"-v", voice}, args...), "--stdout")
	data, err := run(ctx, e.binary(), full)
	if err == nil && len(data) == 0 {
		return nil, errors.New("espeak: no output") // no audible output is an F4 failure
	}
	return data, err
}

// espeakTextSSML builds the SSML input for plain text: the escaped text is
// read as words, with no phoneme element.
func espeakTextSSML(text string) string {
	return "<speak>" + xmlEscape(text) + "</speak>"
}

// xmlEscape escapes markup characters for SSML attributes and text nodes.
func xmlEscape(s string) string {
	return strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	).Replace(s)
}

// runDefault runs the synthesiser command with the kill timeout and the WAV
// output cap.
func (e *EspeakVoices) runDefault(ctx context.Context, name string, args []string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, e.timeout())
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(stderr.String()))
	}
	out, readErr := io.ReadAll(io.LimitReader(pipe, maxWAVBytes+1))
	if len(out) > maxWAVBytes {
		cancel() // abort at the cap: kill the still-running subprocess before waiting
	}
	waitErr := cmd.Wait()
	if readErr != nil {
		return nil, fmt.Errorf("%s: reading output: %w", name, readErr)
	}
	if len(out) > maxWAVBytes {
		return nil, fmt.Errorf("%s: output too large (over %d bytes)", name, maxWAVBytes)
	}
	if waitErr != nil {
		return nil, fmt.Errorf("%s: %w: %s", name, waitErr, strings.TrimSpace(stderr.String()))
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s: no output", name)
	}
	if !bytes.HasPrefix(out, []byte("RIFF")) {
		return nil, fmt.Errorf("%s: output is not WAV", name)
	}
	return out, nil
}
