import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import Page from './+page.svelte';
import { call, onLookupResult, onLookupAlternatives } from '../lib-api.js';

vi.mock('../lib-api.js', () => ({
	call: vi.fn(),
	onLookupResult: vi.fn(),
	onLookupAlternatives: vi.fn()
}));

const englishEntry = {
	Headword: 'gift',
	Meanings: [{ PartOfSpeech: 'noun', Definitions: ['Something given as a present.'] }],
	Source: { Name: 'Dictionary', URL: 'https://example.com', LicenseURL: 'https://example.com/license' }
};
let deliverLookup;
let deliverAlternatives;
let nextSearch;
beforeEach(() => {
	localStorage.clear();
	nextSearch = 0;
	deliverLookup = undefined;
	deliverAlternatives = undefined;
	call.mockReset();
	onLookupResult.mockImplementation(async (handler) => { deliverLookup = handler; return () => {}; });
	onLookupAlternatives.mockImplementation(async (handler) => { deliverAlternatives = handler; return () => {}; });
	call.mockImplementation(async (method) => {
		if (method === 'Dictionaries') return [];
		if (method === 'Progress') return {};
		if (method === 'Lookup') return `search-${++nextSearch}`;
		return null;
	});
});
afterEach(() => {
	cleanup();
	vi.restoreAllMocks();
	localStorage.clear();
	document.documentElement.lang = 'en';
});

async function startLookup(query) {
	render(Page);
	await waitFor(() => expect(deliverLookup).toBeTypeOf('function'));
	await fireEvent.input(screen.getByLabelText('Word or phrase'), { target: { value: query } });
	await fireEvent.click(screen.getByRole('button', { name: 'Look up' }));
	await waitFor(() => expect(call).toHaveBeenCalledWith('Lookup', query));
	await waitFor(() => expect(deliverAlternatives).toBeTypeOf('function'));
}
const lookup = (language, status, attempt = 1, searchID = 'search-1') => ({
	SearchID: searchID, Query: 'gift', Language: language, Attempt: attempt,
	Result: status === 'found' ? { Status: status, Entries: [englishEntry], Origin: 'offline' } : { Status: status }
});
const alternatives = (values = {}) => ({
	SearchID: 'search-1', Query: 'gift', Language: 'es', Attempt: 1,
	Status: 'ready', BaseForms: [], Spelling: [], ...values
});

describe('lookup alternatives', () => {
	it('given_one_language_found_when_other_misses_then_alternatives_preserve_the_definition', async () => {
		// given
		await startLookup('gift');
		deliverLookup(lookup('en', 'found'));
		deliverLookup(lookup('es', 'not_found'));
		await screen.findByText('Something given as a present.');
		// when
		deliverAlternatives(alternatives({ BaseForms: ['regalo'], Spelling: ['gato'] }));
		// then
		expect(await screen.findByRole('button', { name: 'regalo' })).toBeTruthy();
		expect(screen.getByRole('button', { name: 'gato' })).toBeTruthy();
		expect(screen.getByText('Something given as a present.')).toBeTruthy();
		expect(screen.getByText(/possible base forms/i)).toBeTruthy();
		expect(screen.getByText(/did you mean/i)).toBeTruthy();
	});

	it('given_a_phrase_suggestion_when_selected_then_a_fresh_bilingual_lookup_starts', async () => {
		// given
		await startLookup('look foward to');
		deliverLookup({ ...lookup('en', 'not_found'), Query: 'look foward to' });
		deliverAlternatives(alternatives({ Query: 'look foward to', Language: 'en', Spelling: ['look forward to'] }));
		const suggestion = await screen.findByRole('button', { name: 'look forward to' });
		// when
		await userEvent.click(suggestion);
		// then
		await waitFor(() => expect(call).toHaveBeenCalledWith('Lookup', 'look forward to'));
		expect(call.mock.calls.filter(([method]) => method === 'Lookup')).toHaveLength(2);
	});

	it('given_no_candidates_when_alternatives_are_ready_then_the_not_found_state_remains', async () => {
		// given
		await startLookup('gift');
		deliverLookup(lookup('es', 'not_found'));
		expect(await screen.findByText('No definition found.')).toBeTruthy();
		// when
		deliverAlternatives(alternatives({ BaseForms: [], Spelling: [] }));
		// then
		expect(await screen.findByText('No suggestions found.')).toBeTruthy();
		expect(screen.getByText('No definition found.')).toBeTruthy();
		expect(screen.queryByText('Possible base forms')).toBeNull();
		expect(screen.queryByText('Did you mean?')).toBeNull();
	});

	it('given_suggestions_unavailable_when_other_language_succeeds_then_result_is_preserved', async () => {
		// given
		await startLookup('gift');
		deliverLookup(lookup('en', 'found'));
		deliverLookup(lookup('es', 'not_found'));
		// when
		deliverAlternatives(alternatives({ Status: 'unavailable' }));
		// then
		expect(await screen.findByText(/suggestions.*unavailable/i)).toBeTruthy();
		expect(screen.getByText('Something given as a present.')).toBeTruthy();
	});

	it('given_timed_out_alternatives_when_displayed_then_shows_distinct_message', async () => {
		// given
		await startLookup('gift');
		deliverLookup(lookup('es', 'not_found'));
		// when
		deliverAlternatives(alternatives({ Status: 'timed_out' }));
		// then – must show timed-out message, not the generic unavailable message
		expect(await screen.findByText(/timed out/i)).toBeTruthy();
		expect(screen.queryByText(/unavailable right now/i)).toBeNull();
	});

	it('given_source_unavailable_when_lookup_outcome_arrives_then_no_alternatives_start_for_that_language', async () => {
		// given – F2: a provider failure should not be treated as a confirmed miss
		await startLookup('gift');
		deliverLookup(lookup('en', 'found'));
		// when
		deliverLookup(lookup('es', 'source_unavailable'));
		// then – alternatives object for 'es' must be absent (null / undefined), not 'loading'
		await screen.findByText(/retry.*spanish/i);
		expect(screen.queryByText(/checking possible alternatives/i)).toBeNull();
		expect(screen.queryByText(/suggestions/i)).toBeNull();
		// deliver a stale alternatives event for the same search — it must be ignored
		deliverAlternatives(alternatives({ Language: 'es', Spelling: ['gift'] }));
		await new Promise((r) => setTimeout(r, 50));
		expect(screen.queryByRole('button', { name: 'gift' })).toBeNull();
	});

	it('given_a_new_search_when_an_old_alternatives_event_arrives_then_it_is_ignored', async () => {
		// given
		await startLookup('gift');
		deliverLookup(lookup('es', 'not_found'));
		await fireEvent.input(screen.getByLabelText('Word or phrase'), { target: { value: 'new' } });
		await fireEvent.click(screen.getByRole('button', { name: 'Look up' }));
		await waitFor(() => expect(call).toHaveBeenCalledWith('Lookup', 'new'));
		// when
		deliverAlternatives(alternatives({ Spelling: ['outdated'] }));
		deliverLookup({ ...lookup('es', 'not_found', 1, 'search-2'), Query: 'new' });
		deliverAlternatives(alternatives({ SearchID: 'search-2', Query: 'new', Spelling: ['newer'] }));
		// then
		expect(await screen.findByRole('button', { name: 'newer' })).toBeTruthy();
		expect(screen.queryByRole('button', { name: 'outdated' })).toBeNull();
	});

	it('given_a_language_retry_when_an_old_attempt_arrives_then_only_new_attempt_is_shown', async () => {
		// given
		await startLookup('gift');
		deliverLookup(lookup('es', 'source_unavailable'));
		const retry = await screen.findByRole('button', { name: /retry.*spanish/i });
		await userEvent.click(retry);
		await waitFor(() => expect(call).toHaveBeenCalledWith('RetryLanguage', 'search-1', 'gift', 'es'));
		deliverLookup(lookup('es', 'not_found', 2));
		// when
		deliverAlternatives(alternatives({ Attempt: 1, Spelling: ['obsolete'] }));
		deliverAlternatives(alternatives({ Attempt: 2, Spelling: ['current'] }));
		// then
		expect(await screen.findByRole('button', { name: 'current' })).toBeTruthy();
		expect(screen.queryByRole('button', { name: 'obsolete' })).toBeNull();
	});

	it('given_untrusted_candidate_when_shown_then_it_is_text_not_markup', async () => {
		// given
		await startLookup('gift');
		deliverLookup(lookup('es', 'not_found'));
		const candidate = '<img src=x onerror=alert(1)>';
		// when
		deliverAlternatives(alternatives({ Spelling: [candidate] }));
		// then
		const button = await screen.findByRole('button', { name: candidate });
		expect(button.textContent).toBe(candidate);
		expect(button.querySelector('img')).toBeNull();
	});
});
