import { describe, expect, it } from "vitest";
import { compactMeanings } from "./senses.js";

const meaning = (pos, token, ...definitions) => ({
  PartOfSpeech: pos,
  SenseGroup: token,
  Definitions: definitions,
});

describe("compactMeanings", () => {
  it("given_six_confirmed_senses_when_compacted_then_order_and_duplicates_are_preserved_without_mutation", () => {
    // given
    const input = Array.from({ length: 6 }, (_, i) =>
      Object.freeze(
        meaning(
          "Sustantivo masculino",
          "g1",
          i === 5 ? "uno" : ["uno", "dos", "tres", "cuatro", "cinco"][i],
        ),
      ),
    );
    input.forEach((m) => Object.freeze(m.Definitions));
    Object.freeze(input);
    const original = JSON.stringify(input);
    // when
    const result = compactMeanings(input);
    // then
    expect(result).toHaveLength(1);
    expect(result[0].Definitions).toEqual([
      "uno",
      "dos",
      "tres",
      "cuatro",
      "cinco",
      "uno",
    ]);
    expect(JSON.stringify(input)).toBe(original);
  });
  it("given_different_or_missing_evidence_when_compacted_then_senses_remain_separate", () => {
    // given
    const input = [
      meaning("noun", "a", "1"),
      meaning("noun", "b", "2"),
      meaning("noun", "", "3"),
      meaning("noun", undefined, "4"),
      meaning(" ", "a", "5"),
      meaning("", "a", "6"),
      meaning("noun", "a", "7"),
      meaning("verb", "a", "8"),
      meaning("noun", "a", "9"),
      meaning(" noun", "a", "10"),
      meaning("noun", " ", "11"),
      meaning("noun", " ", "12"),
    ];
    // when
    // then
    expect(compactMeanings(input).map((m) => m.Definitions)).toEqual(
      input.map((m) => m.Definitions),
    );
  });
  it("given_multi_definition_blocks_and_later_runs_when_compacted_then_content_and_boundaries_are_preserved", () => {
    // given
    const input = [
      meaning("noun", "a", "é", "example"),
      meaning("noun", "a", "é"),
      meaning("verb", "a", "verb"),
      meaning("noun", "a", "later"),
    ];
    // when
    // then
    expect(compactMeanings(input).map((m) => m.Definitions)).toEqual([
      ["é", "example", "é"],
      ["verb"],
      ["later"],
    ]);
  });
});
