// Owns the pronunciation playback flow (C7): the App.PlayPronunciation
// request, newest-wins play state and the audio element lifecycle. The host
// normally plays natively: the bound call blocks until the sound ends and
// returns Played (no Audio), so the item stays "playing" until it resolves
// and StopPronunciation cancels it. Without native playback the result carries
// an Audio data URI that is assigned to one element per play; it is never
// exposed as a link or anywhere else.

export function createPlaybackController({
  call,
  createAudio = () => new Audio(),
}) {
  let seq = 0; // token of the newest play; older results are ignored
  let audio = null; // the single active audio element
  let state = { playing: null, message: null, voices: {} };
  const subscribers = new Set();

  function update(patch) {
    state = { ...state, ...patch };
    for (const subscriber of subscribers) subscriber(state);
  }

  function stopAudio() {
    if (!audio) return;
    const element = audio;
    audio = null;
    try {
      element.pause();
    } catch {}
  }

  function stopNative() {
    // Fire and forget: cancels any native playback; failures are irrelevant.
    try {
      Promise.resolve(call("StopPronunciation")).catch(() => {});
    } catch {}
  }

  async function request(token, key, item) {
    let result;
    try {
      result = await call("PlayPronunciation", item);
    } catch {
      // Unexpected internal error: no verbatim copy is available (C3); the
      // page shows its fallback copy for the empty text.
      result = { Audio: "", Message: "" };
    }
    if (token !== seq) return; // a newer play replaced this one
    if (result.Played || result.Audio) {
      // C7 hover: once the Go side resolved the voice, the item's hover text
      // names it ("eSpeak-NG · en-us") or, for Piper results (C5 of the
      // piper-voices page), the natural voice ("Piper · alba").
      if (result.Voice)
        update({
          voices: {
            ...state.voices,
            [key]: { Source: result.Source || "", Voice: result.Voice },
          },
        });
      if (result.Played) {
        update({ playing: null, message: null });
        return;
      }
      startAudio(token, key, result.Audio);
      return;
    }
    update({ playing: null, message: { key, text: result.Message || "" } });
  }

  function startAudio(token, key, src) {
    const element = createAudio();
    audio = element;
    element.addEventListener("ended", () => {
      if (audio === element && token === seq) update({ playing: null });
    });
    const failed = () => {
      if (audio !== element || token !== seq) return;
      audio = null;
      update({ playing: null, message: { key, text: "" } });
    };
    element.addEventListener("error", failed);
    element.src = src;
    update({ playing: key });
    const playing = element.play?.();
    if (playing?.catch) playing.catch(failed);
  }

  return {
    subscribe(run) {
      subscribers.add(run);
      run(state);
      return () => subscribers.delete(run);
    },
    play(key, item) {
      const wasPlaying = state.playing;
      const token = ++seq;
      stopAudio(); // newest wins: no queue, no double audio (S12)
      if (wasPlaying !== null) stopNative();
      if (wasPlaying === key) {
        // Clicking the playing item again stops it.
        update({ playing: null, message: null });
        return Promise.resolve();
      }
      update({ playing: key, message: null });
      return request(token, key, item);
    },
    reset() {
      ++seq;
      stopAudio();
      if (state.playing !== null) stopNative();
      update({ playing: null, message: null, voices: {} });
    },
  };
}
