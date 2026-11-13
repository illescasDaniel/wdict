import { render, fireEvent, screen, waitFor, cleanup } from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import Page from "./+page.svelte";
import { call, onLookupResult } from "../lib-api.js";

vi.mock("../lib-api.js", () => ({
  call: vi.fn(),
  onLookupResult: vi.fn(async () => () => {}),
}));

// S8 F4 copy comes from Go verbatim (adapters/voices_catalog.go).
const PythonCopy =
  "Python 3 is required for natural voices. Install Python 3 and retry.";

function pack(id, name, size, extra = {}) {
  return {
    ID: id,
    Kind: id === "piper-runtime" ? "runtime" : "voice",
    Name: name,
    Language: id === "es_ES-sharvard-medium" ? "es" : "en",
    SizeBytes: size,
    Source: { Name: "Example", URL: "", Attribution: "", LicenseURL: "" },
    Installed: false,
    Blocked: "",
    ...extra,
  };
}

const runtime = pack("piper-runtime", "Piper runtime", 120000000);
const alba = pack("en_GB-alba-medium", "en_GB · alba (medium)", 63000000);
const amy = pack("en_US-amy-medium", "en_US · amy (medium)", 63000000);
const sharvard = pack(
  "es_ES-sharvard-medium",
  "es_ES · sharvard (medium)",
  76000000,
);

const packs = [runtime, alba, amy, sharvard];

async function openDictionaries() {
  await fireEvent.click(
    screen.getByRole("button", { name: "Offline content" }),
  );
  await screen.findByText("Natural voices");
}

describe("natural voices section (C7)", () => {
  beforeEach(() => {
    localStorage.clear();
    call.mockReset();
    onLookupResult.mockImplementation(async (handler) => {
      handler;
      return () => {};
    });
  });
  afterEach(() => cleanup());

  it("given_uninstalled_packs_when_section_renders_then_names_sizes_and_download_buttons_are_shown", async () => {
    // given (S5–S8 UI parts, C7)
    call.mockImplementation(async (method) =>
      method === "Dictionaries"
        ? []
        : method === "ListVoices"
          ? packs
          : {},
    );
    // when
    render(Page);
    await openDictionaries();
    // then
    expect(screen.getByText("Piper runtime")).toBeTruthy();
    expect(screen.getByText("en_GB · alba (medium)")).toBeTruthy();
    expect(screen.getByText("en_US · amy (medium)")).toBeTruthy();
    expect(screen.getByText("es_ES · sharvard (medium)")).toBeTruthy();
    expect(screen.getByText(/120 MB/)).toBeTruthy();
    expect(screen.getAllByText(/63 MB/)).toHaveLength(2);
    expect(screen.getByText(/76 MB/)).toBeTruthy();
    expect(screen.getAllByRole("button", { name: "Download" })).toHaveLength(4);
  });

  it("given_python_missing_when_section_renders_then_downloads_are_disabled_with_the_verbatim_f4_copy", async () => {
    // given (S8 UI part, F4)
    call.mockImplementation(async (method) =>
      method === "Dictionaries"
        ? []
        : method === "ListVoices"
          ? packs.map((p) => ({ ...p, Blocked: PythonCopy }))
          : {},
    );
    // when
    render(Page);
    await openDictionaries();
    // then: no download can start and the Go copy is shown verbatim
    for (const button of screen.getAllByRole("button", { name: "Download" }))
      expect(button.hasAttribute("disabled")).toBe(true);
    expect(screen.getAllByText(PythonCopy).length).toBeGreaterThanOrEqual(1);
    expect(call.mock.calls.some(([name]) => name === "DownloadVoicePack")).toBe(
      false,
    );
  });

  it("given_voice_click_when_user_downloads_a_pack_then_download_voice_pack_is_called_with_the_pack_id", async () => {
    // given (S5/S7 UI part)
    call.mockImplementation(async (method) =>
      method === "Dictionaries"
        ? []
        : method === "ListVoices"
          ? packs
          : method === "DownloadVoicePack"
            ? new Promise(() => {})
            : {},
    );
    render(Page);
    await openDictionaries();
    // when
    await fireEvent.click(
      screen.getAllByRole("button", { name: "Download" })[1],
    );
    // then
    expect(call).toHaveBeenCalledWith(
      "DownloadVoicePack",
      "en_GB-alba-medium",
    );
  });

  it("given_download_in_progress_when_the_pack_downloads_then_progress_is_shown_and_the_button_disappears", async () => {
    // given (S5 UI part)
    let finish;
    let installed = false;
    call.mockImplementation(async (method) => {
      if (method === "Dictionaries") return [];
      if (method === "ListVoices")
        return packs.map((p) =>
          p.ID === "en_GB-alba-medium" ? { ...p, Installed: installed } : p,
        );
      if (method === "VoiceProgress")
        return {
          DictionaryID: "en_GB-alba-medium",
          State: "downloading",
          BytesRead: 10000000,
          TotalBytes: 63000000,
        };
      if (method === "DownloadVoicePack")
        return new Promise((resolve) => {
          finish = () => {
            installed = true;
            resolve();
          };
        });
      return {};
    });
    render(Page);
    await openDictionaries();
    // when
    await fireEvent.click(
      screen.getAllByRole("button", { name: "Download" })[1],
    );
    // then: progress mirrors the dictionary-download pattern
    expect(
      screen.getByRole("progressbar", { name: "Natural voices download" }),
    ).toBeTruthy();
    await screen.findByText(/10\.0 MB of 63 MB/);
    expect(screen.getAllByRole("button", { name: "Download" })).toHaveLength(3);
    // completing the download shows the installed tag without a button
    finish();
    await screen.findByText("Installed");
    const tags = screen.getAllByText("Installed");
    expect(tags).toHaveLength(1);
    expect(screen.getAllByRole("button", { name: "Download" })).toHaveLength(3);
  });

  it("given_download_failure_when_the_install_fails_then_failed_state_with_retry_is_shown", async () => {
    // given (S6/S7 UI part, F3)
    let attempts = 0;
    call.mockImplementation(async (method) => {
      if (method === "Dictionaries") return [];
      if (method === "ListVoices") return packs;
      if (method === "DownloadVoicePack") {
        attempts += 1;
        return attempts === 1 ? Promise.reject(Error("network")) : {};
      }
      return {};
    });
    render(Page);
    await openDictionaries();
    await fireEvent.click(
      screen.getAllByRole("button", { name: "Download" })[1],
    );
    // then: the voice-specific download-failure copy and retry appear
    await screen.findByText(
      "Couldn’t install the voice pack. Check your connection and available storage, then try again.",
    );
    await fireEvent.click(await screen.findByRole("button", { name: "Try again" }));
    await waitFor(() => expect(attempts).toBe(2));
  });

  it("given_runtime_failure_when_the_runtime_install_fails_then_retry_targets_the_runtime_pack", async () => {
    // given (S6 UI part)
    const requested = [];
    call.mockImplementation(async (method, ...args) => {
      if (method === "Dictionaries") return [];
      if (method === "ListVoices") return packs;
      if (method === "DownloadVoicePack") {
        requested.push(args[0]);
        return Promise.reject(Error("pip"));
      }
      return {};
    });
    render(Page);
    await openDictionaries();
    // when
    await fireEvent.click(screen.getAllByRole("button", { name: "Download" })[0]);
    await screen.findByRole("button", { name: "Try again" });
    await fireEvent.click(screen.getByRole("button", { name: "Try again" }));
    // then
    expect(requested).toEqual(["piper-runtime", "piper-runtime"]);
  });

  it("given_installed_pack_when_section_renders_then_installed_tag_is_shown_without_a_button", async () => {
    // given (S5 UI part)
    call.mockImplementation(async (method) =>
      method === "Dictionaries"
        ? []
        : method === "ListVoices"
          ? packs.map((p) => ({ ...p, Installed: true }))
          : {},
    );
    // when
    render(Page);
    await openDictionaries();
    // then
    expect(screen.getAllByText("Installed")).toHaveLength(4);
    expect(screen.queryByRole("button", { name: "Download" })).toBeNull();
  });

  it("given_voice_downloads_blocked_by_missing_runtime_when_section_renders_then_the_runtime_first_copy_is_verbatim", async () => {
    // given (C7: voice rows disabled until the runtime exists)
    call.mockImplementation(async (method) =>
      method === "Dictionaries"
        ? []
        : method === "ListVoices"
          ? packs.map((p) =>
              p.ID === "piper-runtime"
                ? p
                : { ...p, Blocked: "Download the Piper runtime first." },
            )
          : {},
    );
    render(Page);
    await openDictionaries();
    // then
    const buttons = screen.getAllByRole("button", { name: "Download" });
    expect(buttons).toHaveLength(4);
    expect(buttons[0].hasAttribute("disabled")).toBe(false);
    for (const button of buttons.slice(1))
      expect(button.hasAttribute("disabled")).toBe(true);
    expect(screen.getAllByText("Download the Piper runtime first.").length).toBe(3);
  });
});
