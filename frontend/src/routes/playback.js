// Owns the pronunciation playback flow (C7): the App.PlayPronunciation
// request, newest-wins play state, the audio element lifecycle, and the F5
// audio-error fallback that re-requests the same item with SkipRecording.
// Audio data URIs are assigned to one element per play; they are never
// exposed as links or anywhere else.

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
    if (result.Audio) {
      // C7 hover: once the Go side resolved the voice, the item's hover text
      // names it ("eSpeak-NG · en-us") or, for Piper results (C5 of the
      // piper-voices page), the natural voice ("Piper · alba"); recordings
      // keep their source title.
      if (result.Voice)
        update({
          voices: {
            ...state.voices,
            [key]: { Source: result.Source || "", Voice: result.Voice },
          },
        });
      startAudio(token, key, item, result.Audio);
      return;
    }
    update({ playing: null, message: { key, text: result.Message || "" } });
  }

  function startAudio(token, key, item, src) {
    const element = createAudio();
    audio = element;
    element.addEventListener("ended", () => {
      if (audio === element && token === seq) update({ playing: null });
    });
    element.addEventListener("error", () => {
      if (audio !== element || token !== seq) return;
      audio = null;
      // F5: the recording decoded but is not playable; fall back to the
      // voice path immediately, once.
      if (item.SkipRecording)
        update({ playing: null, message: { key, text: "" } });
      else request(token, key, { ...item, SkipRecording: true });
    });
    element.src = src;
    update({ playing: key });
    const playing = element.play?.();
    if (playing?.catch)
      playing.catch(() => {
        if (audio !== element || token !== seq) return;
        audio = null;
        if (item.SkipRecording)
          update({ playing: null, message: { key, text: "" } });
        else request(token, key, { ...item, SkipRecording: true });
      });
  }

  return {
    subscribe(run) {
      subscribers.add(run);
      run(state);
      return () => subscribers.delete(run);
    },
    play(key, item) {
      const token = ++seq;
      stopAudio(); // newest wins: no queue, no double audio (S12)
      update({ playing: key, message: null });
      return request(token, key, item);
    },
    reset() {
      ++seq;
      stopAudio();
      update({ playing: null, message: null, voices: {} });
    },
  };
}
