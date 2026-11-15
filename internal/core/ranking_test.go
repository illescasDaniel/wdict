package core

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestGivenCaseVariantsWhenLookupThenExactCaseWinsWithoutMutatingCache(t *testing.T) {
	// given
	entries := []Entry{{Headword: "Car", Meanings: []Meaning{{PartOfSpeech: "Proper noun"}}}, {Headword: "car", Meanings: []Meaning{{PartOfSpeech: "Noun"}}}}
	cache := &fakeCache{found: true, record: CacheRecord{Entries: entries, FetchedAt: time.Now()}}
	s := LookupService{Memory: cache, Online: &fakeSource{}, Clock: fakeClock{time.Now()}}
	// when
	got, err := s.Lookup(context.Background(), LookupRequest{Query: "car", Language: English})
	// then
	if err != nil || got.Entries[0].Headword != "car" || !reflect.DeepEqual(cache.record.Entries, entries) {
		t.Fatalf("got %+v, cache %+v, error %v", got.Entries, cache.record.Entries, err)
	}
}

func BenchmarkLookupCandidateOrdering(b *testing.B) {
	entries := make([]Entry, 100)
	for i := range entries {
		entries[i] = Entry{Headword: "Car", Meanings: []Meaning{{PartOfSpeech: "Proper noun"}}}
	}
	entries[99].Headword = "car"
	s := LookupService{Offline: &fakeSource{entries: entries, found: true}}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = s.Lookup(context.Background(), LookupRequest{Query: "car", Language: English})
	}
}

func TestGivenAmbiguousCaseWhenRankedThenNeutralSlotsRemainAndOrdinaryPrecedesProper(t *testing.T) {
	// given
	entries := []Entry{{Headword: "Car", Meanings: []Meaning{{PartOfSpeech: "Proper noun"}}}, {Headword: "cAr", Meanings: []Meaning{{PartOfSpeech: "unknown"}}}, {Headword: "car", Meanings: []Meaning{{PartOfSpeech: "Noun"}}}, {Headword: "cAR", Meanings: []Meaning{{PartOfSpeech: "Proper noun"}}}}
	// when
	got := RankEntries("CAR", entries)
	// then
	if got[0].Headword != "car" || got[1].Headword != "cAr" || got[2].Headword != "Car" || got[3].Headword != "cAR" {
		t.Fatalf("unexpected order: %+v", got)
	}
}

func TestGivenFourMatchTiersWhenRankedThenStrongestDirectMatchFirst(t *testing.T) {
	// given
	entries := []Entry{{Headword: "other"}, {Headword: "-ano"}, {Headword: "Año"}, {Headword: "año"}}
	// when
	got := RankEntries("año", entries)
	// then
	want := []string{"año", "Año", "-ano", "other"}
	for i := range got {
		if got[i].Headword != want[i] {
			t.Fatalf("got %+v", got)
		}
	}
}

func TestGivenSpellingModeWhenPlayingThenOnlyTextVoiceIsUsed(t *testing.T) {
	// given
	voices := &fakeVoices{available: true, outputs: voiceResult("es", "Car")}
	recordings := &fakeRecordings{data: []byte("recording")}
	s := service(recordings, voices)
	// when
	got := s.Play(context.Background(), PlayRequest{Mode: PlaySpelling, Language: Spanish, Headword: "Car", AudioURL: "https://example.com/x", Text: "/kɑː/", Label: "UK"})
	// then
	if got.Source != PlayVoice || got.Voice != "es" || len(recordings.urls) != 0 || len(voices.calls) != 1 || voices.calls[0].phonemes || voices.calls[0].text != "Car" {
		t.Fatalf("got %+v calls %+v recordings %+v", got, voices.calls, recordings.urls)
	}
}

func TestGivenSharedCachedCandidatesWhenSearchingEitherCaseThenEachQueryRanksIndependently(t *testing.T) {
	// given
	now := time.Now()
	original := []Entry{{Headword: "Car", Meanings: []Meaning{{PartOfSpeech: "Proper noun", Definitions: []string{"name"}}}}, {Headword: "car", Meanings: []Meaning{{PartOfSpeech: "Noun", Definitions: []string{"vehicle"}}}}}
	for _, origin := range []string{"offline", "memory", "disk", "stale", "online"} {
		t.Run(origin, func(t *testing.T) {
			cached := append([]Entry(nil), original...)
			before, err := json.Marshal(cached)
			if err != nil {
				t.Fatal(err)
			}
			source := &fakeSource{found: true, entries: cached}
			s := LookupService{Online: source, Clock: fakeClock{now}}
			switch origin {
			case "offline":
				s.Offline = source
			case "memory":
				s.Memory = &fakeCache{found: true, record: CacheRecord{Entries: cached, FetchedAt: now}}
			case "disk":
				s.Disk = &fakeCache{found: true, record: CacheRecord{Entries: cached, FetchedAt: now}}
			case "stale":
				s.Memory = &fakeCache{found: true, record: CacheRecord{Entries: cached, FetchedAt: now.Add(-366 * 24 * time.Hour)}}
				source.err = errors.New("offline")
				source.found = false
			}
			// when
			for _, query := range []string{"car", "Car", "Car", "car"} {
				got, err := s.Lookup(context.Background(), LookupRequest{Query: query, Language: English})
				if err != nil || len(got.Entries) != 2 || got.Entries[0].Headword != query {
					t.Fatalf("%s got %+v err %v", query, got, err)
				}
			}
			// then
			after, err := json.Marshal(cached)
			if err != nil {
				t.Fatal(err)
			}
			if string(after) != string(before) {
				t.Fatalf("cached nested content changed: before %s after %s", before, after)
			}
		})
	}
}

func TestGivenSpellingModeWhenPlayingEnglishThenDefaultVoiceBypassesRecordingAndPiper(t *testing.T) {
	// given
	voices := &fakeVoices{available: true, outputs: voiceResult("en-us", "Car")}
	rec := &fakeRecordings{data: []byte("rec")}
	piper := &fakePiper{output: []byte("piper")}
	s := piperService(rec, piper, voices)
	// when
	got := s.Play(context.Background(), PlayRequest{Mode: PlaySpelling, Language: English, Headword: "Car", Label: "UK", Text: "/kɑː/", AudioURL: "https://example.com/audio"})
	// then
	if got.Source != PlayVoice || got.Voice != "en-us" || len(rec.urls) != 0 || len(piper.calls) != 0 || len(voices.calls) != 1 || voices.calls[0] != (voiceCall{voice: "en-us", text: "Car"}) {
		t.Fatalf("result %+v, recording %+v piper %+v voice %+v", got, rec.urls, piper.calls, voices.calls)
	}
}

func TestGivenUnavailableSpellingVoiceWhenPlayingThenFailureCanBeRetried(t *testing.T) {
	// given
	voices := &fakeVoices{available: false}
	s := service(nil, voices)
	request := PlayRequest{Mode: PlaySpelling, Language: English, Headword: "Car"}
	// when
	missing := s.Play(context.Background(), request)
	voices.available = true
	failed := s.Play(context.Background(), request)
	voices.outputs = voiceResult("en-us", "Car")
	recovered := s.Play(context.Background(), request)
	// then
	if missing.Message != CopyNoSource || failed.Message != CopyNoSynthesis || recovered.Source != PlayVoice || len(voices.calls) != 2 {
		t.Fatalf("missing %+v failed %+v recovered %+v calls %+v", missing, failed, recovered, voices.calls)
	}
}

type cancelledSpellingVoice struct{ seen error }

func (v *cancelledSpellingVoice) Available() bool { return true }
func (v *cancelledSpellingVoice) SpeakPhonemes(context.Context, string, string) ([]byte, error) {
	return nil, errors.New("unexpected phoneme synthesis")
}
func (v *cancelledSpellingVoice) SpeakText(ctx context.Context, _ string, _ string) ([]byte, error) {
	v.seen = ctx.Err()
	return nil, ctx.Err()
}

func TestGivenCancelledSpellingRequestWhenSynthesisingThenCancellationReachesVoice(t *testing.T) {
	// given
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	v := &cancelledSpellingVoice{}
	s := PronunciationAudioService{Voices: v}
	// when
	got := s.Play(ctx, PlayRequest{Mode: PlaySpelling, Language: English, Headword: "Car"})
	// then
	if !errors.Is(v.seen, context.Canceled) || got.Audio != "" || got.Message != CopyNoSynthesis {
		t.Fatalf("seen %v result %+v", v.seen, got)
	}
}
