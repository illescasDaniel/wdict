import { describe, it, expect, vi } from "vitest";
import fs from "node:fs";
import { groupRecords, clusterByPronunciation } from "./result-groups.js";

describe("record grouping", () => {
  it("given separated matching records when grouped then sections stay distinct in first card", () => {
    const source = {
      ID: "one",
      URL: "https://example.com",
      Name: "dictionary",
      AdditionalURLs: [],
    };
    const a = {
      Headword: "Car",
      Language: "en",
      Source: source,
      Meanings: [{ Definitions: ["language"] }],
    };
    const b = { ...a, Headword: "car" };
    const c = { ...a, Meanings: [{ Definitions: ["surname"] }] };
    const snapshot = structuredClone([a, b, c]);
    expect(groupRecords([a, b, c]).map((g) => g.records)).toEqual([
      [a, c],
      [b],
    ]);
    expect([a, b, c]).toEqual(snapshot);
  });
  it("given missing or different source fields when grouped then records remain separate", () => {
    const a = {
      Headword: "Car",
      Language: "en",
      Source: { ID: "a", URL: "https://a", Name: "A" },
    };
    expect(
      groupRecords([
        a,
        { ...a, Source: { ...a.Source, Name: "B" } },
        { ...a, Source: { ID: "a" } },
      ]),
    ).toHaveLength(3);
  });
  it("given records when grouped then grouping makes no network filesystem or audio calls", () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");
    const readSpy = vi.spyOn(fs, "readFileSync");
    const audioSpy = vi.fn();
    vi.stubGlobal("Audio", audioSpy);
    try {
      const records = [
        {
          Headword: "Car",
          Language: "en",
          Source: { ID: "one", URL: "https://example.org" },
          Meanings: [],
        },
      ];
      expect(groupRecords(records)[0].records).toEqual(records);
      expect(fetchSpy).not.toHaveBeenCalled();
      expect(readSpy).not.toHaveBeenCalled();
      expect(audioSpy).not.toHaveBeenCalled();
    } finally {
      fetchSpy.mockRestore();
      readSpy.mockRestore();
      vi.unstubAllGlobals();
    }
  });
});

describe("pronunciation clustering", () => {
  const ipa = (text, label = "🇬🇧") => ({ Text: text, Label: label });
  const rec = (name, pronunciations) => ({
    Headword: "car",
    Meanings: [{ Definitions: [name] }],
    Pronunciations: pronunciations,
  });
  it("given_consecutive_identical_pronunciations_when_clustered_then_records_share_one_cluster_with_group_indexes", () => {
    // given
    const a = rec("a", [ipa("/kɑː/")]);
    const b = rec("b", [ipa("/kɑː/")]);
    const c = rec("c", [ipa("/kɑʊ/")]);
    const group = { records: [a, b, c], indexes: [4, 5, 6] };
    // when
    const clusters = clusterByPronunciation(group);
    // then
    expect(clusters).toEqual([
      { records: [a, b], indexes: [4, 5] },
      { records: [c], indexes: [6] },
    ]);
  });
  it("given_records_without_pronunciation_when_clustered_then_they_cluster_together_and_apart_from_others", () => {
    // given
    const a = rec("a", undefined);
    const b = rec("b", [{ Text: "" }]);
    const c = rec("c", [ipa("/x/")]);
    const d = rec("d", [ipa("/x/", "🇺🇸")]);
    // when
    const clusters = clusterByPronunciation({
      records: [a, b, c, d],
      indexes: [0, 1, 2, 3],
    });
    // then
    expect(clusters.map((x) => x.indexes)).toEqual([[0, 1], [2], [3]]);
  });
  it("given_non_consecutive_identical_pronunciations_when_clustered_then_order_is_preserved", () => {
    // given
    const p = [ipa("/a/")];
    // when
    const clusters = clusterByPronunciation({
      records: [rec("1", p), rec("2", [ipa("/b/")]), rec("3", p)],
      indexes: [0, 1, 2],
    });
    // then
    expect(clusters).toHaveLength(3);
  });
});
