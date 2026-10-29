import {
  render,
  fireEvent,
  screen,
  waitFor,
  cleanup,
} from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import Page from "./+page.svelte";
import { call, onLookupResult } from "../lib-api.js";
import userEvent from "@testing-library/user-event";
vi.mock("../lib-api.js", () => ({
  call: vi.fn(),
  onLookupResult: vi.fn(async () => () => {}),
}));
const dictionary = {
  ID: "one",
  Name: "English dictionary",
  DownloadBytes: 1000000,
  Installed: false,
  License: "CC BY-SA",
  Source: {
    AdditionalURLs: ["https://example.com"],
    LicenseURL: "https://example.com/license",
  },
};
const entry = {
  Headword: "hello",
  Meanings: [{ PartOfSpeech: "noun", Definitions: ["a greeting"] }],
  Source: {
    Name: "Example",
    Attribution: "Example contributors",
    URL: "https://example.com",
    LicenseURL: "https://example.com/license",
  },
};
let resultHandler;
let fakeSearch = 0;
function mockBackend(implementation) {
  call.mockImplementation(async (method, ...args) => {
    const id = String(++fakeSearch);
    const value = await implementation(method, ...args);
    if (
      method === "Lookup" &&
      value &&
      typeof value === "object" &&
      value.Status
    ) {
      queueMicrotask(() => {
        resultHandler?.({
          SearchID: id,
          Query: args[0],
          Language: "en",
          Result: value,
        });
        resultHandler?.({
          SearchID: id,
          Query: args[0],
          Language: "es",
          Result: {
            Status:
              value.Status === "invalid_query" ? "invalid_query" : "not_found",
          },
        });
      });
      return id;
    }
    return value;
  });
}
beforeEach(() => {
  resultHandler = undefined;
  onLookupResult.mockImplementation(async (handler) => {
    resultHandler = handler;
    return () => {};
  });
  call.mockReset();
  mockBackend(async (method) =>
    method === "Dictionaries"
      ? [dictionary]
      : method === "Progress"
        ? {}
        : null,
  );
});
afterEach(() => cleanup());
describe("lookup and offline dictionary UI", () => {
  it.each([false, true])(
    "given_legacy_cache_when_displayed_then_missing_attribution_is_noted (stale=%s)",
    async (stale) => {
      const legacy = {
        ...entry,
        Source: { ...entry.Source, LicenseURL: "javascript:alert(1)" },
      };
      mockBackend(async (method) =>
        method === "Lookup"
          ? {
              Status: "found",
              Entries: [legacy],
              Origin: "disk_cache",
              Stale: stale,
            }
          : method === "Dictionaries"
            ? [dictionary]
            : {},
      );
      render(Page);
      await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
      await screen.findByText("a greeting");
      expect(
        screen.getByText(
          "Source or license information unavailable for this cached entry.",
        ),
      ).toBeTruthy();
      expect(
        screen.getByRole("link", { name: "Source" }).getAttribute("href"),
      ).toBe("https://example.com/");
      expect(screen.queryByRole("link", { name: "License" })).toBeNull();
    },
  );
  it("given_legacy_cache_when_next_search_fails_then_previous_result_retains_attribution_notice", async () => {
    let failed = false;
    const legacy = { ...entry, Source: { ...entry.Source, URL: "" } };
    mockBackend(async (method) =>
      method === "Lookup"
        ? failed
          ? { Status: "source_unavailable" }
          : { Status: "found", Entries: [legacy], Origin: "disk_cache" }
        : method === "Dictionaries"
          ? [dictionary]
          : {},
    );
    render(Page);
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await screen.findByText("a greeting");
    failed = true;
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await screen.findByText(/Previous result for/);
    expect(screen.getByText("a greeting")).toBeTruthy();
    expect(
      screen.getByText(
        "Source or license information unavailable for this cached entry.",
      ),
    ).toBeTruthy();
    expect(
      screen.getByRole("link", { name: "License" }).getAttribute("href"),
    ).toBe("https://example.com/license");
    expect(screen.queryByRole("link", { name: "Source" })).toBeNull();
  });
  it("given_app_start_when_window_opens_then_no_requests_are_sent_and_network_notice_is_shown", async () => {
    // given
    // when
    render(Page);
    await screen.findByText(/cache miss|saved entry cannot answer/i);
    // then
    expect(
      call.mock.calls.some(
        ([name]) => name === "Lookup" || name === "Download",
      ),
    ).toBe(false);
  });
  it("given_both_languages_when_results_arrive_then_they_are_shown_independently", async () => {
    mockBackend(async (method) =>
      method === "Dictionaries"
        ? [dictionary]
        : method === "Progress"
          ? {}
          : method === "Lookup"
            ? "search-1"
            : null,
    );
    render(Page);
    await fireEvent.input(screen.getByLabelText(/word or phrase/i), {
      target: { value: "casa" },
    });
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await waitFor(() => expect(call).toHaveBeenCalledWith("Lookup", "casa"));
    resultHandler({
      SearchID: "search-1",
      Query: "casa",
      Language: "es",
      Result: {
        Status: "found",
        Origin: "offline",
        Entries: [
          {
            ...entry,
            Headword: "casa",
            Meanings: [{ Definitions: ["Una vivienda"] }],
          },
        ],
      },
    });
    await screen.findByText("Una vivienda");
    expect(screen.getByText(/Searching English/)).toBeTruthy();
    resultHandler({
      SearchID: "search-1",
      Query: "casa",
      Language: "en",
      Result: { Status: "source_unavailable" },
    });
    await screen.findByRole("button", { name: "Retry English" });
    expect(screen.getByText(/English online lookup unavailable/)).toBeTruthy();
    expect(screen.getByText("Una vivienda")).toBeTruthy();
  });
  it("given_spanish_found_and_english_missing_when_results_arrive_then_no_error_is_shown", async () => {
    mockBackend(async (method) =>
      method === "Dictionaries"
        ? [dictionary]
        : method === "Progress"
          ? {}
          : method === "Lookup"
            ? "diccionario"
            : null,
    );
    render(Page);
    await fireEvent.input(screen.getByLabelText(/word or phrase/i), {
      target: { value: "diccionario" },
    });
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await waitFor(() =>
      expect(call).toHaveBeenCalledWith("Lookup", "diccionario"),
    );
    resultHandler({
      SearchID: "diccionario",
      Language: "en",
      Result: { Status: "not_found" },
    });
    resultHandler({
      SearchID: "diccionario",
      Language: "es",
      Result: {
        Status: "found",
        Origin: "offline",
        Entries: [
          {
            ...entry,
            Headword: "diccionario",
            Language: "es",
            Meanings: [{ Definitions: ["Libro de palabras"] }],
          },
        ],
      },
    });
    await screen.findByText("Libro de palabras");
    expect(screen.queryByText(/English online lookup unavailable/)).toBeNull();
    expect(screen.queryByRole("button", { name: "Retry English" })).toBeNull();
  });
  it("given_silent_502_and_spanish_result_when_looking_up_then_no_english_error_is_shown", async () => {
    mockBackend(async (method) =>
      method === "Dictionaries"
        ? [dictionary]
        : method === "Progress"
          ? {}
          : method === "Lookup"
            ? "silent-502"
            : null,
    );
    render(Page);
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await waitFor(() => expect(call).toHaveBeenCalledWith("Lookup", ""));
    resultHandler({
      SearchID: "silent-502",
      Language: "en",
      Result: { Status: "silent_unavailable" },
    });
    resultHandler({
      SearchID: "silent-502",
      Language: "es",
      Result: {
        Status: "found",
        Origin: "offline",
        Entries: [
          {
            ...entry,
            Headword: "diccionario",
            Language: "es",
            Meanings: [{ Definitions: ["Libro de palabras"] }],
          },
        ],
      },
    });
    await screen.findByText("Libro de palabras");
    expect(screen.queryByText(/English online lookup unavailable/)).toBeNull();
    expect(screen.queryByRole("button", { name: "Retry English" })).toBeNull();
  });
  it("given_only_silent_errors_when_search_finishes_then_no_failure_or_false_not_found_is_shown", async () => {
    mockBackend(async (method) =>
      method === "Dictionaries"
        ? [dictionary]
        : method === "Progress"
          ? {}
          : method === "Lookup"
            ? "both-502"
            : null,
    );
    render(Page);
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await waitFor(() => expect(call).toHaveBeenCalledWith("Lookup", ""));
    resultHandler({
      SearchID: "both-502",
      Language: "en",
      Result: { Status: "silent_unavailable" },
    });
    resultHandler({
      SearchID: "both-502",
      Language: "es",
      Result: { Status: "silent_unavailable" },
    });
    await screen.findByText("No definitions available right now.");
    expect(screen.queryByText(/No entry found/)).toBeNull();
    expect(screen.queryByText(/online lookup unavailable/)).toBeNull();
  });
  it("given_retry_pending_when_other_language_returns_then_both_remain_usable", async () => {
    mockBackend(async (method) =>
      method === "Dictionaries"
        ? [dictionary]
        : method === "Progress"
          ? {}
          : method === "Lookup"
            ? "retry-test"
            : method === "RetryLanguage"
              ? new Promise(() => {})
              : null,
    );
    render(Page);
    await fireEvent.input(screen.getByLabelText(/word or phrase/i), {
      target: { value: "casa" },
    });
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await waitFor(() => expect(call).toHaveBeenCalledWith("Lookup", "casa"));
    resultHandler({
      SearchID: "retry-test",
      Language: "en",
      Result: { Status: "source_unavailable" },
    });
    await fireEvent.click(
      await screen.findByRole("button", { name: "Retry English" }),
    );
    resultHandler({
      SearchID: "retry-test",
      Language: "es",
      Result: {
        Status: "found",
        Origin: "offline",
        Entries: [
          {
            ...entry,
            Language: "es",
            Headword: "casa",
            Meanings: [{ Definitions: ["vivienda"] }],
          },
        ],
      },
    });
    await screen.findByText("vivienda");
    expect(screen.getByText(/Searching English/)).toBeTruthy();
  });
  it("given_two_found_outcomes_when_search_completes_then_both_languages_are_labelled", async () => {
    mockBackend(async (method) =>
      method === "Dictionaries"
        ? [dictionary]
        : method === "Progress"
          ? {}
          : method === "Lookup"
            ? "both"
            : null,
    );
    render(Page);
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await waitFor(() => expect(call).toHaveBeenCalledWith("Lookup", ""));
    resultHandler({
      SearchID: "both",
      Language: "en",
      Result: {
        Status: "found",
        Origin: "online",
        Entries: [{ ...entry, Language: "en" }],
      },
    });
    resultHandler({
      SearchID: "both",
      Language: "es",
      Result: {
        Status: "found",
        Origin: "offline",
        Entries: [
          {
            ...entry,
            Language: "es",
            Headword: "hola",
            Meanings: [{ Definitions: ["saludo"] }],
          },
        ],
      },
    });
    await screen.findByText("saludo");
    expect(screen.getByText(/English · hello/)).toBeTruthy();
    expect(screen.getByText(/Spanish · hola/)).toBeTruthy();
    expect(screen.getAllByRole("link", { name: "Source" })).toHaveLength(2);
    expect(screen.getAllByRole("link", { name: "License" })).toHaveLength(2);
  });
  it("given_one_missing_and_one_unavailable_when_search_completes_then_unavailable_is_shown", async () => {
    mockBackend(async (method) =>
      method === "Dictionaries"
        ? [dictionary]
        : method === "Progress"
          ? {}
          : method === "Lookup"
            ? "missing"
            : null,
    );
    render(Page);
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await waitFor(() => expect(call).toHaveBeenCalledWith("Lookup", ""));
    resultHandler({
      SearchID: "missing",
      Language: "en",
      Result: { Status: "not_found" },
    });
    resultHandler({
      SearchID: "missing",
      Language: "es",
      Result: { Status: "source_unavailable" },
    });
    await screen.findByText(/The dictionary is unavailable/);
    expect(screen.queryByText(/No entry found/)).toBeNull();
  });
  it("given_listener_is_pending_when_search_submitted_then_waits_for_registration", async () => {
    let activate;
    onLookupResult.mockImplementation(
      (handler) =>
        new Promise((resolve) => {
          activate = () => {
            resultHandler = handler;
            resolve(() => {});
          };
        }),
    );
    mockBackend(async (method) =>
      method === "Dictionaries"
        ? [dictionary]
        : method === "Progress"
          ? {}
          : method === "Lookup"
            ? "ready"
            : null,
    );
    render(Page);
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    expect(call.mock.calls.some(([method]) => method === "Lookup")).toBe(false);
    activate();
    await waitFor(() => expect(call).toHaveBeenCalledWith("Lookup", ""));
  });
  it("given_new_search_when_old_event_arrives_then_it_is_ignored", async () => {
    let n = 0;
    mockBackend(async (method) =>
      method === "Dictionaries"
        ? [dictionary]
        : method === "Progress"
          ? {}
          : method === "Lookup"
            ? String(++n)
            : null,
    );
    render(Page);
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await waitFor(() => expect(n).toBe(1));
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await waitFor(() => expect(n).toBe(2));
    resultHandler({
      SearchID: "1",
      Language: "en",
      Result: { Status: "found", Origin: "online", Entries: [entry] },
    });
    resultHandler({
      SearchID: "2",
      Language: "es",
      Result: {
        Status: "found",
        Origin: "online",
        Entries: [
          {
            ...entry,
            Headword: "hola",
            Meanings: [{ Definitions: ["saludo"] }],
          },
        ],
      },
    });
    await screen.findByText("saludo");
    expect(screen.queryByText("a greeting")).toBeNull();
  });
  it("given_failed_retry_when_call_rejects_then_unavailable_returns", async () => {
    mockBackend(async (method) =>
      method === "Dictionaries"
        ? [dictionary]
        : method === "Progress"
          ? {}
          : method === "Lookup"
            ? "retry-fail"
            : method === "RetryLanguage"
              ? Promise.reject(Error("failed"))
              : null,
    );
    render(Page);
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await waitFor(() => expect(call).toHaveBeenCalledWith("Lookup", ""));
    resultHandler({
      SearchID: "retry-fail",
      Language: "en",
      Result: { Status: "source_unavailable" },
    });
    resultHandler({
      SearchID: "retry-fail",
      Language: "es",
      Result: { Status: "not_found" },
    });
    await fireEvent.click(
      await screen.findByRole("button", { name: "Retry English" }),
    );
    await screen.findByText(/The dictionary is unavailable/);
    expect(screen.queryByText("Searching…")).toBeNull();
  });
  it("given_retry_from_previous_search_when_it_arrives_then_it_is_ignored", async () => {
    let next = 0;
    mockBackend(async (method) =>
      method === "Dictionaries"
        ? [dictionary]
        : method === "Progress"
          ? {}
          : method === "Lookup"
            ? String(++next)
            : method === "RetryLanguage"
              ? new Promise(() => {})
              : null,
    );
    render(Page);
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await waitFor(() => expect(next).toBe(1));
    resultHandler({
      SearchID: "1",
      Language: "en",
      Result: { Status: "source_unavailable" },
    });
    await fireEvent.click(
      await screen.findByRole("button", { name: "Retry English" }),
    );
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await waitFor(() => expect(next).toBe(2));
    resultHandler({
      SearchID: "1",
      Language: "en",
      Result: { Status: "found", Origin: "online", Entries: [entry] },
    });
    resultHandler({
      SearchID: "2",
      Language: "es",
      Result: {
        Status: "found",
        Origin: "online",
        Entries: [
          {
            ...entry,
            Headword: "hola",
            Meanings: [{ Definitions: ["saludo"] }],
          },
        ],
      },
    });
    await screen.findByText("saludo");
    expect(screen.queryByText("a greeting")).toBeNull();
  });
  it("given_metadata_failure_when_dictionary_opened_then_error_can_be_retried", async () => {
    // given
    let failed = true;
    mockBackend(async (method) => {
      if (method === "Dictionaries" && failed) throw Error("unavailable");
      return method === "Dictionaries" ? [dictionary] : {};
    });
    render(Page);
    // when
    await fireEvent.click(
      screen.getByRole("button", { name: "Offline dictionaries" }),
    );
    await screen.findByText("Unable to load dictionary information.");
    failed = false;
    await fireEvent.click(
      screen.getByRole("button", {
        name: "Retry loading dictionary information",
      }),
    );
    // then
    await screen.findByText("English dictionary");
    expect(
      screen.queryByText("Unable to load dictionary information."),
    ).toBeNull();
  });
  it("given_successful_lookup_when_next_entry_is_missing_then_previous_result_remains", async () => {
    // given
    mockBackend(async (method) =>
      method === "Dictionaries"
        ? [dictionary]
        : method === "Progress"
          ? {}
          : method === "Lookup"
            ? { Status: "found", Entries: [entry], Origin: "online" }
            : null,
    );
    // when
    render(Page);
    await fireEvent.input(screen.getByLabelText(/word or phrase/i), {
      target: { value: "hello" },
    });
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await screen.findByText("a greeting");
    mockBackend(async (method) =>
      method === "Lookup"
        ? { Status: "not_found" }
        : method === "Dictionaries"
          ? [dictionary]
          : {},
    );
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await screen.findByText(/No entry found/);
    // then
    expect(screen.getByText("a greeting")).toBeTruthy();
  });
  it("given_pending_lookup_when_newer_lookup_completes_then_old_response_is_ignored", async () => {
    // given
    let resolveFirst;
    mockBackend((method) =>
      method === "Dictionaries"
        ? Promise.resolve([dictionary])
        : method === "Progress"
          ? Promise.resolve({})
          : method === "Lookup"
            ? new Promise((r) => {
                if (!resolveFirst) resolveFirst = r;
                else r({ Status: "found", Entries: [entry], Origin: "online" });
              })
            : Promise.resolve(),
    );
    // when
    render(Page);
    let input = screen.getByLabelText(/word or phrase/i);
    await fireEvent.input(input, { target: { value: "first" } });
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await fireEvent.input(input, { target: { value: "second" } });
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await screen.findByText("a greeting");
    resolveFirst({ Status: "not_found" });
    // then
    await waitFor(() =>
      expect(screen.queryByText(/No entry found/)).toBeNull(),
    );
  });
  it("given_download_failure_when_user_starts_install_then_retry_is_available", async () => {
    // given
    mockBackend(async (method) =>
      method === "Dictionaries"
        ? [dictionary]
        : method === "Progress"
          ? {}
          : method === "Download"
            ? Promise.reject(Error("network"))
            : null,
    );
    // when
    render(Page);
    await fireEvent.click(
      screen.getByRole("button", { name: "Offline dictionaries" }),
    );
    await screen.findByText("English dictionary");
    await fireEvent.click(screen.getByRole("button", { name: "Download" }));
    await screen.findByText(/Couldn’t install/);
    expect(
      await screen.findByRole("button", { name: "Try again" }),
    ).toBeTruthy();
    // then
    expect(call.mock.calls.some(([name]) => name === "Download")).toBe(true);
  });
  it("given_previous_result_when_lookup_fails_then_error_and_previous_result_are_shown", async () => {
    // given
    let status = "found";
    mockBackend(async (method) =>
      method === "Dictionaries"
        ? [dictionary]
        : method === "Progress"
          ? {}
          : method === "Lookup"
            ? status === "found"
              ? { Status: status, Entries: [entry], Origin: "online" }
              : { Status: status }
            : null,
    );
    // when
    render(Page);
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await screen.findByText("a greeting");
    status = "invalid_query";
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await screen.findByText(/Enter a word or phrase of up to/);
    status = "source_unavailable";
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await screen.findByText(/dictionary is unavailable/);
    // then
    expect(screen.getByText("a greeting")).toBeTruthy();
  });
  it("given_stale_cache_when_provider_fails_then_saved_date_and_attribution_are_shown", async () => {
    // given
    mockBackend(async (method) =>
      method === "Dictionaries"
        ? [dictionary]
        : method === "Progress"
          ? {}
          : method === "Lookup"
            ? {
                Status: "found",
                Entries: [entry],
                Origin: "disk_cache",
                Stale: true,
                FetchedAt: "2026-10-01T00:00:00Z",
              }
            : null,
    );
    // when
    render(Page);
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await screen.findByText(/Stale saved result/);
    // then
    expect(screen.getByText(/Example contributors/)).toBeTruthy();
  });
  it("given_download_in_progress_when_user_searches_then_lookup_works_and_install_completes", async () => {
    // given
    let finish;
    let installed = false;
    let downloading = false;
    mockBackend((method) => {
      if (method === "Dictionaries")
        return Promise.resolve([{ ...dictionary, Installed: installed }]);
      if (method === "Progress")
        return Promise.resolve(
          downloading
            ? { State: "downloading", BytesRead: 500000, TotalBytes: 1000000 }
            : {},
        );
      if (method === "Download") {
        downloading = true;
        return new Promise((resolve) => {
          finish = () => {
            installed = true;
            downloading = false;
            resolve();
          };
        });
      }
      if (method === "Lookup")
        return Promise.resolve({
          Status: "found",
          Entries: [entry],
          Origin: "online",
        });
    });
    // when
    render(Page);
    await fireEvent.click(
      screen.getByRole("button", { name: "Offline dictionaries" }),
    );
    await screen.findByText("English dictionary");
    await fireEvent.click(screen.getByRole("button", { name: "Download" }));
    expect(
      screen.getByRole("progressbar", { name: "Dictionary download" }),
    ).toBeTruthy();
    await fireEvent.click(screen.getByRole("button", { name: "← Look up" }));
    await fireEvent.input(screen.getByLabelText(/word or phrase/i), {
      target: { value: "hello" },
    });
    await fireEvent.submit(
      screen.getByLabelText(/word or phrase/i).closest("form"),
    );
    await screen.findByText("a greeting");
    finish();
    await fireEvent.click(
      screen.getByRole("button", { name: "Offline dictionaries" }),
    );
    // then
    await screen.findByText("Ready to use offline.");
  });
  it("given_pending_download_when_progress_response_is_empty_then_download_remains_active", async () => {
    // given
    let finish;
    mockBackend((method) => {
      if (method === "Dictionaries") return Promise.resolve([dictionary]);
      if (method === "Progress") return Promise.resolve({});
      if (method === "Download")
        return new Promise((resolve) => {
          finish = resolve;
        });
      return Promise.resolve();
    });
    // when
    render(Page);
    await fireEvent.click(
      screen.getByRole("button", { name: "Offline dictionaries" }),
    );
    await screen.findByText("English dictionary");
    await fireEvent.click(screen.getByRole("button", { name: "Download" }));
    await fireEvent.click(screen.getByRole("button", { name: "← Look up" }));
    await fireEvent.click(
      screen.getByRole("button", { name: "Offline dictionaries" }),
    );
    await waitFor(() =>
      expect(
        screen.getByRole("progressbar", { name: "Dictionary download" }),
      ).toBeTruthy(),
    );
    expect(screen.queryByRole("button", { name: "Download" })).toBeNull();
    // then
    expect(
      call.mock.calls.filter(([name]) => name === "Download"),
    ).toHaveLength(1);
    finish();
  });
  it("given_provider_markup_when_result_is_rendered_then_it_remains_plain_text", async () => {
    // given
    const malicious = {
      ...entry,
      Meanings: [{ Definitions: ["<img src=x onerror=alert(1)>"] }],
    };
    mockBackend(async (method) =>
      method === "Dictionaries"
        ? [dictionary]
        : method === "Progress"
          ? {}
          : method === "Lookup"
            ? { Status: "found", Entries: [malicious], Origin: "online" }
            : null,
    );
    // when
    render(Page);
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await screen.findByText("<img src=x onerror=alert(1)>");
    // then
    expect(document.querySelector("img")).toBeNull();
  });
  it("given_pending_lookup_when_using_keyboard_then_navigation_remains_available", async () => {
    // given
    let finish;
    mockBackend((method) =>
      method === "Dictionaries"
        ? Promise.resolve([dictionary])
        : method === "Progress"
          ? Promise.resolve({})
          : method === "Lookup"
            ? new Promise((resolve) => {
                finish = resolve;
              })
            : Promise.resolve(),
    );
    const user = userEvent.setup();
    // when
    render(Page);
    const input = screen.getByLabelText(/word or phrase/i);
    expect(document.activeElement).toBe(input);
    await user.type(input, "hello{Enter}");
    await screen.findByText("Searching…");
    await user.tab();
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Look up" }),
    );
    await user.tab({ shift: true });
    await user.tab({ shift: true });
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Offline dictionaries" }),
    );
    await user.keyboard("{Enter}");
    // then
    await screen.findByText("English dictionary");
    finish({ Status: "found", Entries: [entry], Origin: "online" });
  });
});

describe("appearance", () => {
  let media, listener;
  beforeEach(() => {
    localStorage.clear();
    media = {
      matches: false,
      addEventListener: vi.fn((_, fn) => {
        listener = fn;
      }),
      removeEventListener: vi.fn(),
    };
    vi.stubGlobal(
      "matchMedia",
      vi.fn(() => media),
    );
  });
  afterEach(() => {
    vi.unstubAllGlobals();
    localStorage.clear();
  });
  it("given_no_valid_preference_when_system_changes_then_system_tracks_device", async () => {
    // given
    localStorage.setItem("wdict-appearance", "invalid");
    // when
    render(Page);
    // then
    expect(
      screen
        .getByRole("radio", { name: "System" })
        .getAttribute("aria-checked"),
    ).toBe("true");
    expect(document.documentElement.dataset.theme).toBe("light");
    listener({ matches: true });
    expect(document.documentElement.dataset.theme).toBe("dark");
  });
  it("given_explicit_selection_when_system_changes_and_app_reopens_then_selection_is_restored", async () => {
    // given
    render(Page);
    // when
    await fireEvent.click(screen.getByRole("radio", { name: "Dark" }));
    listener({ matches: false });
    // then
    expect(document.documentElement.dataset.theme).toBe("dark");
    cleanup();
    render(Page);
    expect(
      screen.getByRole("radio", { name: "Dark" }).getAttribute("aria-checked"),
    ).toBe("true");
    expect(document.documentElement.dataset.theme).toBe("dark");
    await fireEvent.click(screen.getByRole("radio", { name: "Light" }));
    expect(document.documentElement.dataset.theme).toBe("light");
    await fireEvent.click(screen.getByRole("radio", { name: "System" }));
    listener({ matches: true });
    expect(document.documentElement.dataset.theme).toBe("dark");
  });
  it("given_unreadable_storage_when_app_opens_then_system_is_selected", () => {
    // given
    const read = vi
      .spyOn(Storage.prototype, "getItem")
      .mockImplementation(() => {
        throw Error("denied");
      });
    // when
    render(Page);
    // then
    expect(
      screen
        .getByRole("radio", { name: "System" })
        .getAttribute("aria-checked"),
    ).toBe("true");
    read.mockRestore();
  });
  it("given_storage_failure_when_mode_changes_then_session_remains_usable", async () => {
    // given
    const save = vi
      .spyOn(Storage.prototype, "setItem")
      .mockImplementation(() => {
        throw Error("denied");
      });
    render(Page);
    // when
    await fireEvent.click(screen.getByRole("radio", { name: "Dark" }));
    // then
    expect(document.documentElement.dataset.theme).toBe("dark");
    save.mockRestore();
  });
  it("given_footer_when_keyboard_selects_mode_and_navigates_then_selection_remains", async () => {
    // given
    const user = userEvent.setup();
    render(Page);
    // when
    screen.getByRole("radio", { name: "System" }).focus();
    await user.keyboard("{ArrowLeft}");
    await user.click(
      screen.getByRole("button", { name: "Offline dictionaries" }),
    );
    // then
    expect(
      screen.getByRole("radio", { name: "Dark" }).getAttribute("aria-checked"),
    ).toBe("true");
    expect(document.documentElement.dataset.theme).toBe("dark");
  });
});
