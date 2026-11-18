package adapters

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// stubBinary writes an executable shell script named name into dir.
func stubBinary(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// linkTools symlinks host tools the stubs need into dir, so PATH can be dir
// alone and no real audio player is ever found.
func linkTools(t *testing.T, dir string, tools ...string) {
	t.Helper()
	for _, tool := range tools {
		path, err := exec.LookPath(tool)
		if err != nil {
			t.Skipf("%s not available", tool)
		}
		if err := os.Symlink(path, filepath.Join(dir, tool)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGivenNoPlayerOnPathWhenCheckedThenUnavailableAndPlayFails(t *testing.T) {
	// given
	t.Setenv("PATH", t.TempDir())
	p := &NativeAudioPlayer{}
	// when / then
	if p.Available() {
		t.Fatal("available without a player")
	}
	if err := p.Play(context.Background(), []byte("RIFF")); err == nil {
		t.Fatal("expected an error")
	}
}

func TestGivenPlayersOnPathWhenPlayedThenPreferredPlayerGetsA0600WavFileThatIsRemoved(t *testing.T) {
	// given: pw-play and aplay both exist; pw-play must win
	dir := t.TempDir()
	log := filepath.Join(dir, "log")
	linkTools(t, dir, "stat", "cat")
	stubBinary(t, dir, "pw-play", `echo "pw $1" > "`+log+`"
mode=$(stat -c %a "$1" 2>/dev/null) || mode=$(stat -f %Lp "$1") || exit 1
printf '%s\n' "$mode" >> "`+log+`"
cat "$1" >> "`+log+`"`)
	stubBinary(t, dir, "aplay", `echo aplay > "`+log+`"`)
	t.Setenv("PATH", dir)
	p := &NativeAudioPlayer{}
	// when
	if !p.Available() {
		t.Fatal("not available")
	}
	if err := p.Play(context.Background(), []byte("WAVDATA")); err != nil {
		t.Fatal(err)
	}
	// then
	out, _ := os.ReadFile(log)
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) != 3 || !strings.HasPrefix(lines[0], "pw ") || lines[1] != "600" || lines[2] != "WAVDATA" {
		t.Fatalf("log %q", out)
	}
	file := strings.TrimPrefix(lines[0], "pw ")
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Fatalf("temp file %s not removed: %v", file, err)
	}
}

func TestGivenFirstPlayerFailsWhenPlayedThenNextPlayerIsTried(t *testing.T) {
	// given
	dir := t.TempDir()
	log := filepath.Join(dir, "log")
	stubBinary(t, dir, "pw-play", `exit 1`)
	stubBinary(t, dir, "paplay", `echo paplay > "`+log+`"`)
	t.Setenv("PATH", dir)
	// when
	err := (&NativeAudioPlayer{}).Play(context.Background(), []byte("x"))
	// then
	out, _ := os.ReadFile(log)
	if err != nil || strings.TrimSpace(string(out)) != "paplay" {
		t.Fatalf("err=%v log=%q", err, out)
	}
}

func TestGivenAfplayOnPathWhenPlayedThenOnlyDarwinUsesIt(t *testing.T) {
	dir := t.TempDir()
	wavDir := filepath.Join(dir, "audio files;test")
	if err := os.Mkdir(wavDir, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TMPDIR", wavDir)
	linkTools(t, dir, "cat")
	log := filepath.Join(dir, "log")
	stubBinary(t, dir, "afplay", `[ "$#" -eq 1 ] || exit 2
printf '%s\n' "$1" > "`+log+`"
cat "$1" >> "`+log+`"`)
	t.Setenv("PATH", dir)
	player := &NativeAudioPlayer{}
	wantAvailable := runtime.GOOS == "darwin"
	if player.Available() != wantAvailable {
		t.Fatalf("available = %v, want %v", player.Available(), wantAvailable)
	}
	err := player.Play(context.Background(), []byte("WAVDATA"))
	if !wantAvailable {
		if err == nil {
			t.Fatal("afplay unexpectedly ran outside Darwin")
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	path, data, _ := strings.Cut(string(out), "\n")
	if filepath.Dir(path) != wavDir || filepath.Ext(path) != ".wav" || data != "WAVDATA" {
		t.Fatalf("afplay received %q with data %q", path, data)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("temp file %s not removed: %v", path, err)
	}
}

func TestGivenAfplayFailsOnDarwinWhenPlayedThenFallbackRuns(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("requires Darwin")
	}
	dir := t.TempDir()
	log := filepath.Join(dir, "log")
	stubBinary(t, dir, "afplay", `echo afplay >> "`+log+`"
exit 1`)
	stubBinary(t, dir, "pw-play", `echo pw-play >> "`+log+`"`)
	t.Setenv("PATH", dir)
	if err := (&NativeAudioPlayer{}).Play(context.Background(), []byte("WAVDATA")); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "afplay\npw-play\n" {
		t.Fatalf("player order = %q", out)
	}
}

func TestGivenAllPlayersFailOnDarwinWhenPlayedThenErrorsAreReportedAndWavIsRemoved(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("requires Darwin")
	}
	dir := t.TempDir()
	wavDir := t.TempDir()
	stubBinary(t, dir, "afplay", `echo "could not open audio output" >&2
exit 2`)
	stubBinary(t, dir, "pw-play", `echo "could not connect to PipeWire" >&2
exit 3`)
	t.Setenv("PATH", dir)
	t.Setenv("TMPDIR", wavDir)

	err := (&NativeAudioPlayer{}).Play(context.Background(), []byte("WAVDATA"))
	if err == nil ||
		!strings.Contains(err.Error(), "afplay: exit status 2: could not open audio output") ||
		!strings.Contains(err.Error(), "pw-play: exit status 3: could not connect to PipeWire") {
		t.Fatalf("playback error = %v, want both player failures", err)
	}
	files, err := os.ReadDir(wavDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatalf("temporary directory contains %v after playback failed", files)
	}
}

func TestGivenAplayOnlyWhenPlayedThenQuietFlagIsPassed(t *testing.T) {
	// given
	dir := t.TempDir()
	log := filepath.Join(dir, "log")
	stubBinary(t, dir, "aplay", `echo "$1" > "`+log+`"`)
	t.Setenv("PATH", dir)
	// when
	err := (&NativeAudioPlayer{}).Play(context.Background(), []byte("x"))
	// then
	out, _ := os.ReadFile(log)
	if err != nil || strings.TrimSpace(string(out)) != "-q" {
		t.Fatalf("err=%v log=%q", err, out)
	}
}

func TestGivenLongPlaybackWhenContextCancelledThenPlayerIsKilledPromptly(t *testing.T) {
	// given
	dir := t.TempDir()
	linkTools(t, dir, "sleep")
	player := "pw-play"
	if runtime.GOOS == "darwin" {
		player = "afplay"
	}
	ready := filepath.Join(dir, "ready")
	stubBinary(t, dir, player, `echo ready > "`+ready+`" && exec sleep 30`)
	t.Setenv("PATH", dir)
	wavDir := t.TempDir()
	t.Setenv("TMPDIR", wavDir)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	// when
	go func() { done <- (&NativeAudioPlayer{}).Play(ctx, []byte("x")) }()
	deadline := time.Now().Add(3 * time.Second)
	for {
		_, err := os.Stat(ready)
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			t.Fatal(err)
		}
		if time.Now().After(deadline) {
			t.Fatal("player did not signal startup")
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	// then
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("playback error = %v, want context canceled", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("player not killed")
	}
	files, err := os.ReadDir(wavDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatalf("temporary directory contains %v after cancellation", files)
	}
}

func TestGivenPlayerOverrunsMaxDurationWhenPlayedThenItIsKilled(t *testing.T) {
	// given
	dir := t.TempDir()
	linkTools(t, dir, "sleep")
	stubBinary(t, dir, "pw-play", `exec sleep 30`)
	t.Setenv("PATH", dir)
	p := &NativeAudioPlayer{MaxDuration: 200 * time.Millisecond}
	start := time.Now()
	// when
	err := p.Play(context.Background(), []byte("x"))
	// then
	if err == nil || time.Since(start) > 3*time.Second {
		t.Fatalf("err=%v after %s", err, time.Since(start))
	}
}

func TestGivenAppImageEnvironmentWhenPlayedThenPlayerSeesSanitisedLibraryPath(t *testing.T) {
	// given
	dir := t.TempDir()
	log := filepath.Join(dir, "log")
	stubBinary(t, dir, "pw-play", `echo "[$LD_LIBRARY_PATH]" > "`+log+`"`)
	t.Setenv("PATH", dir)
	t.Setenv("APPDIR", "/tmp/.mount_x")
	t.Setenv("LD_LIBRARY_PATH", "/tmp/.mount_x/usr/lib:/usr/local/lib")
	// when
	err := (&NativeAudioPlayer{}).Play(context.Background(), []byte("x"))
	// then
	out, _ := os.ReadFile(log)
	if err != nil || strings.TrimSpace(string(out)) != "[/usr/local/lib]" {
		t.Fatalf("err=%v log=%q", err, out)
	}
}
