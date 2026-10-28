import { readFileSync } from 'node:fs';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';

export default { plugins: [sveltekit({ adapter: adapter(), paths: { relative: false }, version: { name: readFileSync(new URL('../VERSION', import.meta.url), 'utf8').trim(), pollInterval: 0 } })] };
