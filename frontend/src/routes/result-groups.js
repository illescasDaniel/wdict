export function pronunciationItems(entry) {
  return (
    Array.isArray(entry.Pronunciations) ? entry.Pronunciations : []
  ).filter((item) => item && item.Text);
}

function pronunciationSignature(entry) {
  return JSON.stringify(
    pronunciationItems(entry).map((item) => [
      item.Label ?? "",
      item.LabelTitle ?? "",
      item.Text,
      item.AudioURL ?? "",
    ]),
  );
}

// Consecutive records of one group with an identical pronunciation list (or
// none at all) share one pronunciation row; indexes stay the group's indexes.
export function clusterByPronunciation(group) {
  const clusters = [];
  let lastSignature = null;
  for (const [position, record] of group.records.entries()) {
    const signature = pronunciationSignature(record);
    if (signature !== lastSignature) {
      clusters.push({ records: [], indexes: [] });
      lastSignature = signature;
    }
    const cluster = clusters[clusters.length - 1];
    cluster.records.push(record);
    cluster.indexes.push(group.indexes[position]);
  }
  return clusters;
}

export function groupRecords(entries) {
  const groups = [];
  const byKey = new Map();
  for (const [index, entry] of entries.entries()) {
    const source = entry.Source;
    const key =
      source?.ID?.trim() && source?.URL?.trim()
        ? JSON.stringify([
            entry.Headword,
            entry.Language,
            source.ID,
            source.Name,
            source.URL,
            source.Attribution,
            source.LicenseURL,
            source.AdditionalURLs,
          ])
        : null;
    let group = key === null ? null : byKey.get(key);
    if (!group) {
      group = {
        headword: entry.Headword,
        language: entry.Language,
        records: [],
        indexes: [],
      };
      groups.push(group);
      if (key !== null) byKey.set(key, group);
    }
    group.records.push(entry);
    group.indexes.push(index);
  }
  return groups;
}
