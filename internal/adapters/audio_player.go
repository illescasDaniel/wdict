package adapters

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// nativePlayers lists the host WAV players in preference order.
var nativePlayers = []struct {
	name string
	args []string
}{
	{"pw-play", nil},
	{"paplay", nil},
	{"aplay", []string{"-q"}},
	{"ffplay", []string{"-nodisp", "-autoexit", "-loglevel", "quiet"}},
}

// defaultMaxPlayback bounds one native playback so a stuck player is killed.
const defaultMaxPlayback = 30 * time.Second

// NativeAudioPlayer plays a WAV with the first host player found on PATH
// (pw-play, paplay, aplay, ffplay), so playback does not depend on the
// webview's GStreamer plugins (absent from the AppImage). The players are host
// binaries and are never bundled.
type NativeAudioPlayer struct {
	MaxDuration time.Duration // zero = 30 s
}

// Available reports whether any host player is on PATH.
func (p *NativeAudioPlayer) Available() bool {
	return len(p.found()) > 0
}

func (p *NativeAudioPlayer) found() []string {
	var paths []string
	for _, player := range nativePlayers {
		if path, err := exec.LookPath(player.name); err == nil {
			paths = append(paths, path)
		}
	}
	return paths
}

// Play writes the WAV to a private temporary file and plays it, blocking until
// the player exits or ctx is cancelled (the player is killed). A player that
// fails on its own hands over to the next one; cancellation and the maximum
// duration do not.
func (p *NativeAudioPlayer) Play(ctx context.Context, wav []byte) error {
	paths := p.found()
	if len(paths) == 0 {
		return errors.New("no native audio player (pw-play, paplay, aplay, ffplay) found")
	}
	file, err := os.CreateTemp("", "wdict-*.wav") // mode 0600
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, err = file.Write(wav)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	limit := p.MaxDuration
	if limit <= 0 {
		limit = defaultMaxPlayback
	}
	ctx, cancel := context.WithTimeout(ctx, limit)
	defer cancel()
	var failures []string
	for _, path := range paths {
		name := baseName(path)
		args := []string{file.Name()}
		for _, player := range nativePlayers {
			if player.name == name {
				args = append(append([]string{}, player.args...), file.Name())
			}
		}
		cmd := exec.CommandContext(ctx, path, args...)
		cmd.WaitDelay = time.Second
		cmd.Env = piperEnv(os.Environ())
		out, runErr := cmd.CombinedOutput()
		if runErr == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		failures = append(failures, fmt.Sprintf("%s: %v: %s", name, runErr, strings.TrimSpace(string(out))))
	}
	return errors.New(strings.Join(failures, "; "))
}

func baseName(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	return path
}
