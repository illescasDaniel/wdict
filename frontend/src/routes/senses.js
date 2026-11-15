export function compactMeanings(meanings) {
  const groups = [];
  let previous;
  for (const meaning of meanings) {
    const token = meaning.SenseGroup;
    const pos = meaning.PartOfSpeech;
    if (
      typeof token === "string" &&
      token.trim() &&
      typeof pos === "string" &&
      pos.trim() &&
      previous &&
      previous.SenseGroup === token &&
      previous.PartOfSpeech === pos
    ) {
      for (const definition of meaning.Definitions) {
        groups[groups.length - 1].Definitions.push(definition);
      }
    } else {
      groups.push({ ...meaning, Definitions: [...meaning.Definitions] });
    }
    previous = meaning;
  }
  return groups;
}
