package core

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// fakePlayer records native playbacks; block makes Play wait for ctx
// cancellation, like a real player still producing sound.
type fakePlayer struct {
	mu        sync.Mutex
	available bool
	err       error
	block     bool
	played    [][]byte
	started   chan struct{}
}

func (f *fakePlayer) Available() bool { return f.available }
func (f *fakePlayer) Play(ctx context.Context, wav []byte) error {
	f.mu.Lock()
	f.played = append(f.played, wav)
	f.mu.Unlock()
	if f.started != nil {
		f.started <- struct{}{}
	}
	if f.block {
		<-ctx.Done()
		return ctx.Err()
	}
	return f.err
}

func (f *fakePlayer) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.played)
}

func nativeService(voices *fakeVoices, player *fakePlayer) *PronunciationAudioService {
	return &PronunciationAudioService{Voices: voices, Player: player}
}

var wordRequest = PlayRequest{Text: "/wɝːd/", Language: English, Headword: "word"}

func TestGivenNativePlayerWhenPlayedThenAudioIsPlayedNativelyWithoutDataURI(t *testing.T) {
	// given
	voices := &fakeVoices{available: true, outputs: phonemeResult("en-us", "w3:d")}
	player := &fakePlayer{available: true}
	// when
	result := nativeService(voices, player).Play(context.Background(), wordRequest)
	// then
	if !result.Played || result.Audio != "" || result.Source != PlayVoice || result.Voice != "en-us" || result.Message != "" {
		t.Fatalf("result %+v", result)
	}
	if player.count() != 1 || string(player.played[0]) != wavBytes {
		t.Fatalf("played %q", player.played)
	}
}

func TestGivenNoNativePlayerWhenPlayedThenAudioDataURIIsReturned(t *testing.T) {
	// given
	voices := &fakeVoices{available: true, outputs: phonemeResult("en-us", "w3:d")}
	player := &fakePlayer{available: false}
	// when
	result := nativeService(voices, player).Play(context.Background(), wordRequest)
	// then
	if result.Played || result.Audio == "" || player.count() != 0 {
		t.Fatalf("result %+v, played %d", result, player.count())
	}
}

func TestGivenPlayerFailsWhenPlayedThenFallsBackToAudioDataURI(t *testing.T) {
	// given
	voices := &fakeVoices{available: true, outputs: phonemeResult("en-us", "w3:d")}
	player := &fakePlayer{available: true, err: errors.New("no sound server")}
	// when
	result := nativeService(voices, player).Play(context.Background(), wordRequest)
	// then
	if result.Played || result.Audio == "" || result.Source != PlayVoice {
		t.Fatalf("result %+v", result)
	}
}

func TestGivenSynthesisFailsWhenPlayedThenNothingIsPlayedAndMessageIsReturned(t *testing.T) {
	// given
	voices := &fakeVoices{available: true}
	player := &fakePlayer{available: true}
	// when
	result := nativeService(voices, player).Play(context.Background(), wordRequest)
	// then
	if result.Played || result.Message != CopyNoSynthesis || player.count() != 0 {
		t.Fatalf("result %+v", result)
	}
}

func TestGivenSynthesisDeadlineWhenPlaybackRunsThenItIsNotCutByTheSynthesisContext(t *testing.T) {
	// given: a player that outlives the synthesis context
	voices := &fakeVoices{available: true, outputs: phonemeResult("en-us", "w3:d")}
	ctx, cancel := context.WithCancel(context.Background())
	player := &fakePlayer{available: true}
	player.started = make(chan struct{}, 1)
	s := nativeService(voices, player)
	// when: the request context is cancelled once playback began
	done := make(chan PlayResult)
	go func() { done <- s.Play(ctx, wordRequest) }()
	<-player.started
	cancel()
	result := <-done
	// then
	if !result.Played {
		t.Fatalf("result %+v", result)
	}
}

func TestGivenPlaybackRunningWhenANewPlayStartsThenThePreviousIsCancelled(t *testing.T) {
	// given
	voices := &fakeVoices{available: true, outputs: phonemeResult("en-us", "w3:d")}
	player := &fakePlayer{available: true, block: true, started: make(chan struct{}, 2)}
	s := nativeService(voices, player)
	first := make(chan PlayResult)
	go func() { first <- s.Play(context.Background(), wordRequest) }()
	<-player.started
	// when
	second := make(chan PlayResult)
	go func() { second <- s.Play(context.Background(), wordRequest) }()
	// then: the first returns promptly (cancelled) while the second plays
	select {
	case r := <-first:
		if !r.Played || r.Audio != "" {
			t.Fatalf("first %+v", r)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("first playback was not cancelled")
	}
	<-player.started
	s.Stop()
	select {
	case r := <-second:
		if !r.Played {
			t.Fatalf("second %+v", r)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Stop did not end the playback")
	}
}

func TestGivenNoPlaybackWhenStoppedThenNothingHappens(t *testing.T) {
	// given
	s := nativeService(&fakeVoices{}, &fakePlayer{available: true})
	// when / then
	s.Stop()
}
