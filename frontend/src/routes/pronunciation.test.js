import { render, fireEvent, screen, cleanup } from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import Page from "./+page.svelte";
import { call, onLookupResult } from "../lib-api.js";
vi.mock("../lib-api.js", () => ({
  call: vi.fn(),
  onLookupResult: vi.fn(async () => () => {}),
}));
const source = {
  Name: "Example",
  Attribution: "Example contributors",
  URL: "https://example.com",
  LicenseURL: "https://example.com/license",
};
function entryWith(pronunciations, headword = "hello") {
  const entry = {
    Headword: headword,
    Meanings: [{ PartOfSpeech: "noun", Definitions: ["a greeting"] }],
    Source: source,
  };
  if (pronunciations !== undefined) entry.Pronunciations = pronunciations;
  return entry;
}
let resultHandler;
let lookups;
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
    return method === "Dictionaries" ? [] : {};
  });
}
async function lookUp() {
  await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
  await screen.findByText("a greeting");
}
beforeEach(() => {
  localStorage.clear();
  lookups = 0;
  resultHandler = undefined;
  onLookupResult.mockImplementation(async (handler) => {
    resultHandler = handler;
    return () => {};
  });
  call.mockReset();
});
afterEach(() => cleanup());
describe("pronunciation line", () => {
  it("given_labelled_and_unlabelled_pronunciations_when_displayed_then_line_is_under_title_above_meanings", async () => {
    // given
    mockBackend(() => ({
      Status: "found",
      Origin: "online",
      Entries: [
        entryWith([
          { Text: "/hɛˈloʊ/", Label: "🇺🇸", LabelTitle: "General American" },
          {
            Text: "/həˈləʊ/",
            Label: "🇬🇧",
            LabelTitle: "Received Pronunciation",
          },
          { Text: "[ˈola]", Label: "" },
        ]),
      ],
    }));
    render(Page);
    // when
    await lookUp();
    // then
    const line = screen.getByRole("group", { name: "Pronunciation" });
    expect(line.tagName).toBe("P");
    expect(line.textContent.trim()).toBe("🇺🇸 /hɛˈloʊ/ · 🇬🇧 /həˈləʊ/ · [ˈola]");
    const title = screen.getByRole("heading", { level: 2, name: /hello/ });
    const part = screen.getByRole("heading", { level: 3 });
    expect(
      title.compareDocumentPosition(line) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
    expect(
      line.compareDocumentPosition(part) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
  });
  it.each([undefined, null, []])(
    "given_no_pronunciations_when_displayed_then_no_line (%s)",
    async (value) => {
      // given
      mockBackend(() => ({
        Status: "found",
        Origin: "online",
        Entries: [entryWith(value)],
      }));
      render(Page);
      // when
      await lookUp();
      // then
      expect(screen.queryByRole("group", { name: "Pronunciation" })).toBeNull();
    },
  );
  it("given_markup_in_text_when_displayed_then_it_is_shown_as_text", async () => {
    // given
    mockBackend(() => ({
      Status: "found",
      Origin: "online",
      Entries: [entryWith([{ Text: "<b>x</b>", Label: "<i>US</i>" }])],
    }));
    render(Page);
    // when
    await lookUp();
    // then
    const line = screen.getByRole("group", { name: "Pronunciation" });
    expect(line.querySelector("b, i")).toBeNull();
    expect(line.textContent.trim()).toBe("<i>US</i> <b>x</b>");
  });
  it("given_spanish_interface_when_displayed_then_accessible_name_is_spanish", async () => {
    // given
    mockBackend(() => ({
      Status: "found",
      Origin: "online",
      Entries: [entryWith([{ Text: "[ˈola]" }])],
    }));
    render(Page);
    await fireEvent.change(screen.getByRole("combobox"), {
      target: { value: "es" },
    });
    // when
    await fireEvent.click(screen.getByRole("button", { name: "Buscar" }));
    await screen.findByText("a greeting");
    // then
    expect(
      screen.getByRole("group", { name: "Pronunciación" }).textContent.trim(),
    ).toBe("[ˈola]");
  });
  it("given_previous_result_after_failed_search_when_displayed_then_line_is_shown", async () => {
    // given
    mockBackend((_, n) =>
      n === 1
        ? {
            Status: "found",
            Origin: "online",
            Entries: [
              entryWith([
                {
                  Text: "/hɛˈloʊ/",
                  Label: "🇺🇸",
                  LabelTitle: "General American",
                },
              ]),
            ],
          }
        : { Status: "source_unavailable" },
    );
    render(Page);
    await lookUp();
    // when
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await screen.findByText(/Previous result for/);
    // then
    expect(
      screen.getByRole("group", { name: "Pronunciation" }).textContent.trim(),
    ).toBe("🇺🇸 /hɛˈloʊ/");
  });
  it("given_flag_with_label_title_when_displayed_then_hover_text_and_accessible_name_are_the_original", async () => {
    // given
    mockBackend(() => ({
      Status: "found",
      Origin: "online",
      Entries: [
        entryWith([
          { Text: "/hɛˈloʊ/", Label: "🇺🇸", LabelTitle: "General American" },
          { Text: "/haˈlaʊ/", Label: "Northumbria" },
        ]),
      ],
    }));
    render(Page);
    // when
    await lookUp();
    // then
    const flag = screen.getByRole("img", { name: "General American" });
    expect(flag.getAttribute("title")).toBe("General American");
    expect(flag.textContent).toBe("🇺🇸");
    expect(screen.queryByRole("img", { name: "Northumbria" })).toBeNull();
    expect(
      screen.getByRole("group", { name: "Pronunciation" }).textContent.trim(),
    ).toBe("🇺🇸 /hɛˈloʊ/ · Northumbria /haˈlaʊ/");
  });
  it("given_repeated_and_differing_lists_when_displayed_then_only_first_card_of_identical_list_shows_line", async () => {
    // given
    const same = [
      { Text: "/hɛˈloʊ/", Label: "🇺🇸", LabelTitle: "General American" },
    ];
    mockBackend(() => ({
      Status: "found",
      Origin: "offline",
      Entries: [
        entryWith(same, "hello"),
        entryWith(structuredClone(same), "Hello"),
        entryWith(structuredClone(same), "hello"),
        entryWith([{ Text: "/ɹiːd/" }], "read"),
        entryWith([{ Text: "/ɹiːd/" }], "read"),
        entryWith([{ Text: "/ɹɛd/" }], "read"),
      ],
    }));
    render(Page);
    // when
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await screen.findAllByText("a greeting");
    // then
    const lines = screen
      .getAllByRole("group", { name: "Pronunciation" })
      .map((line) => line.textContent.trim());
    expect(lines).toEqual(["🇺🇸 /hɛˈloʊ/", "/ɹiːd/", "/ɹɛd/"]);
  });
  it("given_same_headword_and_list_in_two_languages_when_shown_as_previous_result_then_both_lines_are_shown", async () => {
    // given
    const same = [{ Text: "/ˈola/" }];
    mockBackend((_, n) =>
      n === 1
        ? {
            Status: "found",
            Origin: "online",
            Entries: [
              { ...entryWith(same, "hola"), Language: "en" },
              { ...entryWith(structuredClone(same), "hola"), Language: "es" },
            ],
          }
        : { Status: "source_unavailable" },
    );
    render(Page);
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await screen.findAllByText("a greeting");
    // when
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    await screen.findAllByText(/Previous result for/);
    // then
    const lines = screen
      .getAllByRole("group", { name: "Pronunciation" })
      .map((line) => line.textContent.trim());
    expect(lines).toEqual(["/ˈola/", "/ˈola/"]);
  });
});
