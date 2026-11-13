import { render, fireEvent, screen, cleanup } from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import Page from "./+page.svelte";
import { call, onLookupResult } from "../lib-api.js";
import { createPlaybackController } from "./playback.js";

vi.mock("../lib-api.js", () => ({
  call: vi.fn(),
  onLookupResult: vi.fn(async () => () => {}),
}));

const CopyNoSource =
  "No recording, and eSpeak-NG is not installed. Install eSpeak-NG (or use the AppImage build) to hear this transcription.";
const CopyNoSynthesis = "Could not synthesise this pronunciation.";

const flush = () => new Promise((resolve) => setTimeout(resolve, 0));

function fakeAudioFactory() {
  const created = [];
  const factory = () => {
    const element = {
      src: "",
      listeners: {},
      paused: true,
      addEventListener(type, handler) {
        (element.listeners[type] ||= []).push(handler);
      },
      play() {
        element.paused = false;
        return Promise.resolve();
      },
      pause() {
        element.paused = true;
      },
      emit(type) {
        for (const handler of element.listeners[type] || [])
          handler({ type, target: element });
      },
    };
    created.push(element);
    return element;
  };
  factory.created = created;
  return factory;
}

function track(controller) {
  const states = [];
  controller.subscribe((state) => states.push(state));
  return {
    states,
    current: () => states[states.length - 1],
  };
}

const baseRequest = {
  Text: "/wɝːd/",
  Label: "🇺🇸",
  LabelTitle: "General American",
  Language: "en",
  AudioURL: "",
  SkipRecording: false,
  Headword: "word",
};

describe("playback controller", () => {
  let backendResult;
  let controller;
  let audioFactory;
  let tracked;
  beforeEach(() => {
    backendResult = { Audio: "data:audio/wav;base64,AAAA", Message: "" };
    call.mockReset();
    call.mockImplementation(async () => backendResult);
    audioFactory = fakeAudioFactory();
    controller = createPlaybackController({
      call,
      createAudio: audioFactory,
    });
    tracked = track(controller);
  });

  it("given_click_when_backend_returns_audio_then_request_is_verbatim_and_playing_state_shows", async () => {
    // when
    await controller.play("item-1", baseRequest);
    // then
    expect(call).toHaveBeenCalledTimes(1);
    expect(call).toHaveBeenCalledWith("PlayPronunciation", baseRequest);
    expect(tracked.current().playing).toBe("item-1");
    expect(tracked.current().message).toBeNull();
    expect(audioFactory.created).toHaveLength(1);
    expect(audioFactory.created[0].src).toBe("data:audio/wav;base64,AAAA");
    expect(audioFactory.created[0].paused).toBe(false);
  });

  it("given_two_pending_plays_when_newest_resolves_then_newest_wins_with_no_queue_or_double_audio", async () => {
    // given
    const pending = [];
    call.mockImplementation(
      () =>
        new Promise((resolve) => {
          pending.push(resolve);
        }),
    );
    controller.play("item-a", baseRequest);
    controller.play("item-b", baseRequest);
    // when: B resolves first, then the older A resolves late
    pending[1](backendResult);
    await flush();
    pending[0](backendResult);
    await flush();
    // then
    expect(tracked.current().playing).toBe("item-b");
    expect(audioFactory.created).toHaveLength(1);
    expect(audioFactory.created[0].src).toBe("data:audio/wav;base64,AAAA");
    expect(audioFactory.created[0].paused).toBe(false);
  });

  it("given_playing_recording_when_newer_click_starts_then_older_audio_stops", async () => {
    // given
    await controller.play("item-a", baseRequest);
    const older = audioFactory.created[0];
    expect(older.paused).toBe(false);
    // when
    await controller.play("item-b", baseRequest);
    // then
    expect(older.paused).toBe(true);
    expect(audioFactory.created).toHaveLength(2);
    expect(tracked.current().playing).toBe("item-b");
  });

  it("given_audio_ends_when_playback_finishes_then_playing_state_clears", async () => {
    // given
    await controller.play("item-1", baseRequest);
    // when
    audioFactory.created[0].emit("ended");
    // then
    expect(tracked.current().playing).toBeNull();
  });

  it("given_recording_not_decodable_when_error_event_fires_then_voice_fallback_is_requested_immediately", async () => {
    // given
    const requests = [];
    call.mockImplementation(async (method, request) => {
      requests.push(request);
      return requests.length === 1
        ? { Audio: "data:audio/mpeg;base64,AAAA", Message: "" }
        : { Audio: "data:audio/wav;base64,BBBB", Message: "" };
    });
    await controller.play("item-1", baseRequest);
    // when
    audioFactory.created[0].emit("error");
    await vi.waitFor(() => expect(call).toHaveBeenCalledTimes(2));
    // then
    expect(requests[1]).toEqual({ ...baseRequest, SkipRecording: true });
    expect(tracked.current().message).toBeNull();
    expect(tracked.current().playing).toBe("item-1");
    expect(audioFactory.created[1].src).toBe("data:audio/wav;base64,BBBB");
  });

  it("given_fallback_also_fails_when_voice_path_returns_copy_then_verbatim_copy_is_shown", async () => {
    // given
    const requests = [];
    call.mockImplementation(async (method, request) => {
      requests.push(request);
      return requests.length === 1
        ? { Audio: "data:audio/mpeg;base64,AAAA", Message: "" }
        : { Audio: "", Message: CopyNoSynthesis };
    });
    await controller.play("item-1", baseRequest);
    // when
    audioFactory.created[0].emit("error");
    await vi.waitFor(() =>
      expect(tracked.current().message?.text).toBe(CopyNoSynthesis),
    );
    // then
    expect(requests[1].SkipRecording).toBe(true);
    expect(tracked.current().playing).toBeNull();
    expect(tracked.current().message.key).toBe("item-1");
  });

  it("given_voice_audio_also_not_decodable_when_error_fires_again_then_no_endless_fallback", async () => {
    // given: both attempts return audio the webview rejects
    call.mockImplementation(async () => ({
      Audio: "data:audio/mpeg;base64,AAAA",
      Message: "",
    }));
    await controller.play("item-1", baseRequest);
    audioFactory.created[0].emit("error");
    await vi.waitFor(() => expect(audioFactory.created).toHaveLength(2));
    // when
    audioFactory.created[1].emit("error");
    await vi.waitFor(() => expect(tracked.current().playing).toBeNull());
    // then
    expect(call).toHaveBeenCalledTimes(2);
    expect(tracked.current().message?.key).toBe("item-1");
  });

  it("given_backend_returns_no_audio_when_first_call_fails_then_copy_is_shown_without_retry", async () => {
    // given
    backendResult = { Audio: "", Message: CopyNoSource };
    // when
    await controller.play("item-1", baseRequest);
    // then
    expect(call).toHaveBeenCalledTimes(1);
    expect(tracked.current().playing).toBeNull();
    expect(tracked.current().message).toEqual({
      key: "item-1",
      text: CopyNoSource,
    });
  });

  it("given_call_rejects_when_unexpected_error_then_playing_clears_and_copy_stays_empty", async () => {
    // given
    call.mockImplementation(async () => {
      throw new Error("boom");
    });
    // when
    await controller.play("item-1", baseRequest);
    // then
    expect(tracked.current().playing).toBeNull();
    expect(tracked.current().message).toEqual({ key: "item-1", text: "" });
  });

  it("given_reset_when_new_lookup_starts_then_state_clears_and_audio_stops", async () => {
    // given
    await controller.play("item-1", baseRequest);
    // when
    controller.reset();
    // then
    expect(audioFactory.created[0].paused).toBe(true);
    expect(tracked.current().playing).toBeNull();
    expect(tracked.current().message).toBeNull();
  });
});

const source = {
  Name: "Example",
  Attribution: "Example contributors",
  URL: "https://example.com",
  LicenseURL: "https://example.com/license",
};
function entryWith(pronunciations, headword = "hello") {
  return {
    Headword: headword,
    Meanings: [{ PartOfSpeech: "noun", Definitions: ["a greeting"] }],
    Source: source,
    Pronunciations: pronunciations,
  };
}
class FakeAudio {
  constructor() {
    FakeAudio.instances.push(this);
    this.src = "";
    this.listeners = {};
    this.paused = true;
  }
  addEventListener(type, handler) {
    (this.listeners[type] ||= []).push(handler);
  }
  play() {
    this.paused = false;
    return Promise.resolve();
  }
  pause() {
    this.paused = true;
  }
  emit(type) {
    for (const handler of this.listeners[type] || [])
      handler({ type, target: this });
  }
}
FakeAudio.instances = [];

describe("play button on the pronunciation line", () => {
  let resultHandler;
  let lookups;
  let playResult;
  function mockBackend(lookup) {
    let id = 0;
    call.mockImplementation(async (method, ...args) => {
      if (method === "Lookup") {
        const searchID = String(++id);
        const result = lookup(args[0], ++lookups);
        queueMicrotask(() => {
          resultHandler?.({
            SearchID: searchID,
            Query: args[0],
            Language: "en",
            Result: result,
          });
          resultHandler?.({
            SearchID: searchID,
            Query: args[0],
            Language: "es",
            Result: { Status: "not_found" },
          });
        });
        return searchID;
      }
      if (method === "PlayPronunciation") return playResult(args[0]);
      return method === "Dictionaries" ? [] : {};
    });
  }
  async function lookUp() {
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await screen.findByText("a greeting");
  }
  const playButton = (name) => screen.getByRole("button", { name });
  beforeEach(() => {
    localStorage.clear();
    lookups = 0;
    resultHandler = undefined;
    playResult = () => ({ Audio: "", Message: "" });
    FakeAudio.instances = [];
    vi.stubGlobal("Audio", FakeAudio);
    onLookupResult.mockImplementation(async (handler) => {
      resultHandler = handler;
      return () => {};
    });
    call.mockReset();
  });
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("given_items_with_and_without_recording_when_rendered_then_hover_and_accessible_names_match_their_source", async () => {
    // given
    mockBackend(() => ({
      Status: "found",
      Origin: "online",
      Entries: [
        entryWith([
          {
            Text: "/wɜːd/",
            Label: "🇬🇧",
            LabelTitle: "Received Pronunciation",
            AudioURL: "https://example.com/audio.mp3",
          },
          { Text: "/wɝːd/", Label: "🇺🇸", LabelTitle: "General American" },
        ]),
      ],
    }));
    render(Page);
    // when
    await lookUp();
    // then
    const recording = playButton("Play pronunciation — Received Pronunciation");
    expect(recording.getAttribute("title")).toBe(
      "Recording · Example · Example contributors",
    );
    const voice = playButton("Play pronunciation — General American");
    expect(voice.getAttribute("title")).toBe("eSpeak-NG");
    expect(recording.hasAttribute("disabled")).toBe(false);
    expect(voice.hasAttribute("disabled")).toBe(false);
    // the idle line display is unchanged by the buttons
    expect(
      screen.getByRole("group", { name: "Pronunciation" }).textContent.trim(),
    ).toBe("🇬🇧 /wɜːd/ · 🇺🇸 /wɝːd/");
  });

  it("given_item_without_attribution_when_rendered_then_hover_names_source_only", async () => {
    // given
    mockBackend(() => ({
      Status: "found",
      Origin: "online",
      Entries: [
        {
          ...entryWith([{ Text: "/wɜːd/", AudioURL: "https://example.com/a.mp3" }]),
          Source: {
            Name: "Example",
            URL: "https://example.com",
            LicenseURL: "https://example.com/license",
          },
        },
      ],
    }));
    render(Page);
    await lookUp();
    // then
    expect(playButton("Play pronunciation").getAttribute("title")).toBe(
      "Recording · Example",
    );
  });
  it("given_click_when_requesting_play_then_app_receives_the_full_contract", async () => {
    // given
    playResult = () => ({ Audio: "data:audio/wav;base64,AAAA", Message: "" });
    mockBackend(() => ({
      Status: "found",
      Origin: "online",
      Entries: [
        entryWith([
          {
            Text: "/wɜːd/",
            Label: "🇬🇧",
            LabelTitle: "Received Pronunciation",
            AudioURL: "https://example.com/audio.mp3",
          },
        ]),
      ],
    }));
    render(Page);
    await lookUp();
    // when
    await fireEvent.click(
      playButton("Play pronunciation — Received Pronunciation"),
    );
    // then
    expect(call).toHaveBeenCalledWith("PlayPronunciation", {
      Text: "/wɜːd/",
      Label: "🇬🇧",
      LabelTitle: "Received Pronunciation",
      Language: "en",
      AudioURL: "https://example.com/audio.mp3",
      SkipRecording: false,
      Headword: "hello",
    });
  });

  it("given_voice_result_with_voice_when_it_resolves_then_hover_names_the_voice", async () => {
    // given (C7 hover upgrade)
    playResult = () => ({
      Audio: "data:audio/wav;base64,AAAA",
      Source: "espeak",
      Voice: "en-us",
      Message: "",
    });
    mockBackend(() => ({
      Status: "found",
      Origin: "online",
      Entries: [entryWith([{ Text: "/wɝːd/", Label: "🇺🇸", LabelTitle: "General American" }])],
    }));
    render(Page);
    await lookUp();
    // when
    await fireEvent.click(playButton("Play pronunciation — General American"));
    // then
    expect(
      playButton("Play pronunciation — General American").getAttribute("title"),
    ).toBe("eSpeak-NG · en-us");
  });

  it("given_piper_result_with_voice_when_it_resolves_then_hover_names_piper_and_the_voice", async () => {
    // given (C5 hover: Piper results name the Piper voice, not eSpeak)
    playResult = () => ({
      Audio: "data:audio/wav;base64,AAAA",
      Source: "piper",
      Voice: "alba",
      Message: "",
    });
    mockBackend(() => ({
      Status: "found",
      Origin: "online",
      Entries: [entryWith([{ Text: "/wɝːd/", Label: "🇺🇸", LabelTitle: "General American" }])],
    }));
    render(Page);
    await lookUp();
    // when
    await fireEvent.click(playButton("Play pronunciation — General American"));
    // then
    expect(
      playButton("Play pronunciation — General American").getAttribute("title"),
    ).toBe("Piper · alba");
  });

  it("given_play_result_with_audio_when_it_arrives_then_data_uri_plays_with_playing_state", async () => {
    // given
    playResult = () => ({ Audio: "data:audio/wav;base64,AAAA", Message: "" });
    mockBackend(() => ({
      Status: "found",
      Origin: "online",
      Entries: [entryWith([{ Text: "/wɝːd/", Label: "🇺🇸" }])],
    }));
    render(Page);
    await lookUp();
    // when
    await fireEvent.click(playButton("Play pronunciation"));
    // then
    expect(FakeAudio.instances).toHaveLength(1);
    expect(FakeAudio.instances[0].src).toBe("data:audio/wav;base64,AAAA");
    expect(FakeAudio.instances[0].paused).toBe(false);
    expect(playButton("Play pronunciation").className).toMatch(/playing/);
    // no audio URI ever becomes a link
    expect(
      document.querySelectorAll('a[href^="data:audio"]').length,
    ).toBe(0);
  });

  it("given_nothing_can_play_when_backend_returns_f1_copy_then_copy_shows_inline_and_button_stays_enabled", async () => {
    // given
    playResult = () => ({ Audio: "", Message: CopyNoSource });
    mockBackend(() => ({
      Status: "found",
      Origin: "online",
      Entries: [entryWith([{ Text: "/wɝːd/", Label: "🇺🇸" }])],
    }));
    render(Page);
    await lookUp();
    // when
    await fireEvent.click(playButton("Play pronunciation"));
    await screen.findByText(CopyNoSource);
    // then
    const button = playButton("Play pronunciation");
    expect(button.hasAttribute("disabled")).toBe(false);
    const line = screen.getByRole("group", { name: "Pronunciation" });
    expect(line.textContent).toContain(CopyNoSource);
    // no plugin-install guidance is ever invented
    expect(document.body.textContent).not.toMatch(/plugin/i);
  });

  it("given_recording_not_decodable_when_audio_errors_then_voice_fallback_is_requested_for_the_same_item", async () => {
    // given
    playResult = () => ({ Audio: "data:audio/mpeg;base64,AAAA", Message: "" });
    mockBackend(() => ({
      Status: "found",
      Origin: "online",
      Entries: [
        entryWith([
          {
            Text: "/wɜːd/",
            Label: "🇬🇧",
            LabelTitle: "Received Pronunciation",
            AudioURL: "https://example.com/audio.mp3",
          },
        ]),
      ],
    }));
    render(Page);
    await lookUp();
    await fireEvent.click(
      playButton("Play pronunciation — Received Pronunciation"),
    );
    // when
    FakeAudio.instances[0].emit("error");
    // then
    await vi.waitFor(() => {
      const plays = call.mock.calls.filter(([m]) => m === "PlayPronunciation");
      expect(plays).toHaveLength(2);
      expect(plays[1][1]).toEqual({
        Text: "/wɜːd/",
        Label: "🇬🇧",
        LabelTitle: "Received Pronunciation",
        Language: "en",
        AudioURL: "https://example.com/audio.mp3",
        SkipRecording: true,
        Headword: "hello",
      });
    });
  });

  it("given_fallback_fails_too_when_voice_path_returns_copy_then_verbatim_copy_is_shown", async () => {
    // given
    let attempts = 0;
    playResult = () => {
      attempts += 1;
      return attempts === 1
        ? { Audio: "data:audio/mpeg;base64,AAAA", Message: "" }
        : { Audio: "", Message: CopyNoSynthesis };
    };
    mockBackend(() => ({
      Status: "found",
      Origin: "online",
      Entries: [entryWith([{ Text: "/wɝːd/", Label: "🇺🇸" }])],
    }));
    render(Page);
    await lookUp();
    await fireEvent.click(playButton("Play pronunciation"));
    // when
    FakeAudio.instances[0].emit("error");
    // then
    await screen.findByText(CopyNoSynthesis);
    expect(screen.queryByText(CopyNoSource)).toBeNull();
  });

  it("given_rapid_clicks_on_two_items_when_both_land_then_newest_play_wins", async () => {
    // given
    playResult = () => ({ Audio: "data:audio/wav;base64,AAAA", Message: "" });
    mockBackend(() => ({
      Status: "found",
      Origin: "online",
      Entries: [
        entryWith([{ Text: "/wɝːd/", Label: "🇺🇸" }, { Text: "/həˈləʊ/" }]),
      ],
    }));
    render(Page);
    await lookUp();
    const buttons = screen.getAllByRole("button", {
      name: "Play pronunciation",
    });
    // when
    await fireEvent.click(buttons[0]);
    await fireEvent.click(buttons[1]);
    // then: the newest play owns the single active audio element
    expect(FakeAudio.instances).toHaveLength(2);
    expect(FakeAudio.instances[0].paused).toBe(true);
    expect(FakeAudio.instances[1].paused).toBe(false);
    expect(buttons[0].className).not.toMatch(/playing/);
    expect(buttons[1].className).toMatch(/playing/);
  });

  it("given_audio_ends_when_playback_completes_then_playing_state_clears", async () => {
    // given
    playResult = () => ({ Audio: "data:audio/wav;base64,AAAA", Message: "" });
    mockBackend(() => ({
      Status: "found",
      Origin: "online",
      Entries: [entryWith([{ Text: "/wɝːd/", Label: "🇺🇸" }])],
    }));
    render(Page);
    await lookUp();
    await fireEvent.click(playButton("Play pronunciation"));
    // when
    FakeAudio.instances[0].emit("ended");
    // then
    await flush();
    expect(playButton("Play pronunciation").className).not.toMatch(/playing/);
  });
});
