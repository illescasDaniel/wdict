package adapters

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
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
	stubBinary(t, dir, "pw-play", `echo "pw $1" > "`+log+`"; stat -c %a "$1" >> "`+log+`"; cat "$1" >> "`+log+`"`)
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
	stubBinary(t, dir, "pw-play", `exec sleep 30`)
	t.Setenv("PATH", dir)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	// when
	go func() { done <- (&NativeAudioPlayer{}).Play(ctx, []byte("x")) }()
	time.Sleep(200 * time.Millisecond)
	cancel()
	// then
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected a cancellation error")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("player not killed")
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
