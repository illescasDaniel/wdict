import { readFileSync, writeFileSync } from 'node:fs';

// SvelteKit's generated HTML contains whitespace-only lines.
const index = new URL('../build/index.html', import.meta.url);
writeFileSync(index, readFileSync(index, 'utf8').replace(/[\t ]+$/gm, ''));
