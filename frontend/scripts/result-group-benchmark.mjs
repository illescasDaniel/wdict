// Measures JS preparation only; excludes DOM creation and rendering in both paths.
// Baseline replicates HEAD's direct #each entries traversal; current path groups
// then traverses records. Use --expose-gc; heap deltas are retained memory,
// not allocation counts, and may include runtime noise.
import { performance } from "node:perf_hooks";
import { groupRecords } from "../src/routes/result-groups.js";
import os from "node:os";
const source = {
  ID: "dict",
  Name: "Dictionary",
  URL: "https://example.org",
  Attribution: "CC",
  LicenseURL: "https://example.org/license",
  AdditionalURLs: [],
};
let sink = 0;
function baseline(entries) {
  for (const entry of entries) sink += entry.Headword.length;
}
function grouped(entries) {
  for (const group of groupRecords(entries))
    for (const entry of group.records) sink += entry.Headword.length;
}
function measure(fn, entries, iterations) {
  for (let i = 0; i < 100; i++) fn(entries);
  global.gc?.();
  const before = process.memoryUsage().heapUsed;
  const start = performance.now();
  for (let i = 0; i < iterations; i++) fn(entries);
  const ms = performance.now() - start;
  global.gc?.();
  const retainedBytes = process.memoryUsage().heapUsed - before;
  return {
    ms: Number(ms.toFixed(3)),
    usPerCall: Number(((ms * 1000) / iterations).toFixed(3)),
    retainedBytes,
  };
}
console.log(
  JSON.stringify({
    node: process.version,
    platform: `${os.platform()} ${os.release()} ${os.arch()}`,
    cpu: os.cpus()[0]?.model,
    methodology:
      "HEAD direct entry traversal versus groupRecords plus traversal; JS only, no DOM; 100 warmups; forced GC; retained heap delta is not allocations",
  }),
);
for (const size of [10, 100, 1000]) {
  const entries = Array.from({ length: size }, (_, i) => ({
    Headword: `word${Math.floor(i / 2)}`,
    Language: "en",
    Source: source,
    Meanings: [],
  }));
  const iterations = size === 1000 ? 1000 : 10000;
  console.log(
    JSON.stringify({
      size,
      iterations,
      baseline: measure(baseline, entries, iterations),
      grouped: measure(grouped, entries, iterations),
    }),
  );
}
if (sink === 0) throw new Error("empty benchmark");
