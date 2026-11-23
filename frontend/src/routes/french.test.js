import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import Page from './+page.svelte';
import { chooseLanguage, chooseAppearance, openAppearance } from './menu-helpers.js';
import { resources } from '../language.js';
import { call, onLookupResult } from '../lib-api.js';

vi.mock('../lib-api.js', () => ({
	call: vi.fn(),
	onLookupResult: vi.fn(async () => () => {})
}));

const source = {
	Name: 'Wiktionary', Attribution: 'Contributors',
	URL: 'https://example.com/entry', LicenseURL: 'https://example.com/license'
};
const entry = (headword, definition, extra = {}) => ({
	Headword: headword, Meanings: [{ PartOfSpeech: 'noun', Definitions: [definition] }],
	Source: source, ...extra
});
let deliver;
let searchNumber;

beforeEach(() => {
	localStorage.clear();
	searchNumber = 0;
	call.mockReset();
	onLookupResult.mockImplementation(async (handler) => { deliver = handler; return () => {}; });
	call.mockImplementation(async (method) => {
		if (method === 'Lookup') return `search-${++searchNumber}`;
		if (method === 'Dictionaries' || method === 'ListVoices') return [];
		return {};
	});
});
afterEach(() => {
	cleanup();
	localStorage.clear();
	document.documentElement.lang = 'en';
});

async function search(word) {
	await fireEvent.input(screen.getByLabelText('Word or phrase'), { target: { value: word } });
	await fireEvent.click(screen.getByRole('button', { name: 'Look up' }));
	await waitFor(() => expect(call).toHaveBeenCalledWith('Lookup', word));
}
function send(id, language, result) {
	deliver({ SearchID: id, Language: language, Result: result });
}
const found = (value, origin = 'online') => ({ Status: 'found', Origin: origin, Entries: [value] });
const before = (a, b) => Boolean(a.compareDocumentPosition(b) & Node.DOCUMENT_POSITION_FOLLOWING);

describe('French support', () => {
	it('given_french_interface_when_restarted_then_language_and_catalog_are_restored', async () => {
		// given
		render(Page);
		// when
		await chooseLanguage('fr');
		await waitFor(() => expect(document.documentElement.lang).toBe('fr'));
		cleanup();
		render(Page);
		// then
		await waitFor(() => expect(document.documentElement.lang).toBe('fr'));
		expect(screen.getByRole('button', { name: /^Langue: Français/ })).toBeTruthy();
		expect(resources.fr.translation).toBeTruthy();
		expect(Object.keys(resources.fr.translation).sort()).toEqual(Object.keys(resources.en.translation).sort());
		for (const key of Object.keys(resources.en.translation)) {
			const params = (message) => [...message.matchAll(/{{\s*([^{}]+?)\s*}}/g)].map((m) => m[1]).sort();
			expect(resources.fr.translation[key]).toBeTruthy();
			expect(params(resources.fr.translation[key])).toEqual(params(resources.en.translation[key]));
		}
	});

	it('given_french_cached_definition_when_interface_changes_then_date_is_french_and_definition_unchanged', async () => {
		// given
		render(Page);
		await search('bonjour');
		const fetched = '2026-10-01T12:00:00Z';
		send('search-1', 'fr', { ...found(entry('bonjour', 'formule de salutation')), Origin: 'disk_cache', Stale: true, FetchedAt: fetched });
		send('search-1', 'en', { Status: 'not_found' });
		send('search-1', 'es', { Status: 'not_found' });
		await screen.findByText('formule de salutation');
		// when
		await chooseLanguage('fr');
		// then
		await waitFor(() => expect(document.documentElement.lang).toBe('fr'));
		expect(screen.getByText('formule de salutation')).toBeTruthy();
		expect(screen.getByText(new RegExp(new Date(fetched).toLocaleDateString('fr').replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))).toBeTruthy();
	});

	it('given_three_languages_when_results_arrive_separately_then_each_is_shown_and_french_interface_orders_first', async () => {
		// given
		render(Page);
		await search('cafe');
		// when
		send('search-1', 'fr', found(entry('café', 'boisson française')));
		await screen.findByText('boisson française');
		expect(screen.queryByText('English definition')).toBeNull();
		send('search-1', 'es', found(entry('café', 'bebida española')));
		await screen.findByText('bebida española');
		send('search-1', 'en', found(entry('cafe', 'English definition')));
		await screen.findByText('English definition');
		await chooseLanguage('fr');
		// then
		await waitFor(() => expect(document.documentElement.lang).toBe('fr'));
		const french = screen.getByText('boisson française');
		const english = screen.getByText('English definition');
		const spanish = screen.getByText('bebida española');
		expect(before(french, english)).toBe(true);
		expect(before(english, spanish)).toBe(true);
		expect(call.mock.calls.filter(([method]) => method === 'Lookup')).toHaveLength(1);
	});

	it('given_previous_results_when_interface_language_changes_then_previous_groups_follow_language_order', async () => {
		// given
		render(Page);
		await search('cafe');
		send('search-1', 'fr', found(entry('café', 'définition française', { Language: 'fr' })));
		send('search-1', 'es', found(entry('café', 'definición española', { Language: 'es' })));
		send('search-1', 'en', found(entry('cafe', 'English definition', { Language: 'en' })));
		await screen.findByText('English definition');
		// when
		await search('missing');
		for (const language of ['en', 'es', 'fr']) send('search-2', language, { Status: 'source_unavailable' });
		await screen.findAllByText(/Previous result/);
		// then
		const french = screen.getByText('définition française');
		const english = screen.getByText('English definition');
		const spanish = screen.getByText('definición española');
		expect(before(english, spanish)).toBe(true);
		expect(before(spanish, french)).toBe(true);
		await chooseLanguage('fr');
		await waitFor(() => expect(document.documentElement.lang).toBe('fr'));
		expect(before(screen.getByText('définition française'), screen.getByText('English definition'))).toBe(true);
		expect(before(screen.getByText('English definition'), screen.getByText('definición española'))).toBe(true);
	});

	it('given_french_unavailable_when_retrying_then_other_results_remain_and_only_french_is_retried', async () => {
		// given
		render(Page);
		await search('bonjour');
		send('search-1', 'en', found(entry('bonjour', 'English result')));
		send('search-1', 'es', { Status: 'not_found' });
		send('search-1', 'fr', { Status: 'source_unavailable' });
		await screen.findByText('English result');
		// when
		const retry = await screen.findByRole('button', { name: /retry.*french/i });
		await userEvent.click(retry);
		// then
		expect(call).toHaveBeenCalledWith('RetryLanguage', 'search-1', 'bonjour', 'fr');
		expect(screen.getByText('English result')).toBeTruthy();
		expect(call.mock.calls.filter(([method]) => method === 'Lookup')).toHaveLength(1);
	});

	it('given_french_interface_when_language_names_appear_in_sentences_then_they_are_lowercase_and_titles_stay_capitalised', async () => {
		// given
		const dictionary = {
			ID: 'es-es', Name: 'Spanish dictionary', Language: 'es', Installed: false,
			DownloadBytes: 1000, License: 'CC BY-SA 4.0',
			Source: { AdditionalURLs: ['https://example.com'], LicenseURL: 'https://example.com/license' }
		};
		call.mockImplementation(async (method) => {
			if (method === 'Lookup') return `search-${++searchNumber}`;
			if (method === 'Dictionaries') return [dictionary];
			if (method === 'ListVoices') return [];
			return {};
		});
		render(Page);
		await search('bonjour');
		send('search-1', 'en', { Status: 'source_unavailable' });
		send('search-1', 'es', { Status: 'not_found' });
		await screen.findByRole('button', { name: 'Retry English' });
		// when
		await chooseLanguage('fr');
		// then
		await waitFor(() => expect(document.documentElement.lang).toBe('fr'));
		expect(screen.getByRole('button', { name: 'Réessayer en anglais' })).toBeTruthy();
		expect(screen.getByText(/Recherche en ligne en anglais indisponible\./)).toBeTruthy();
		expect(screen.getByText('Recherche en français…')).toBeTruthy();
		expect(screen.getByRole('heading', { name: 'Espagnol' })).toBeTruthy();
		await userEvent.click(screen.getByRole('button', { name: 'Contenu hors ligne' }));
		expect(await screen.findByText(/Langue : définitions en espagnol/)).toBeTruthy();
	});

	it('given_english_and_spanish_interfaces_when_language_names_appear_in_sentences_then_they_are_unchanged', async () => {
		// given
		render(Page);
		await search('bonjour');
		send('search-1', 'en', { Status: 'source_unavailable' });
		send('search-1', 'es', { Status: 'not_found' });
		// when
		const english = await screen.findByRole('button', { name: 'Retry English' });
		await chooseLanguage('es');
		// then
		expect(english).toBeTruthy();
		await waitFor(() => expect(document.documentElement.lang).toBe('es'));
		expect(screen.getByRole('button', { name: 'Reintentar en Inglés' })).toBeTruthy();
		expect(screen.getByText('Buscando en Francés…')).toBeTruthy();
	});

	it('given_french_result_from_old_search_when_new_search_starts_then_old_result_cannot_replace_current', async () => {
		// given
		render(Page);
		await search('ancien');
		await search('nouveau');
		send('search-2', 'fr', found(entry('nouveau', 'nouvelle définition')));
		await screen.findByText('nouvelle définition');
		// when
		send('search-1', 'fr', found(entry('ancien', 'ancienne définition')));
		// then
		expect(screen.getByText('nouvelle définition')).toBeTruthy();
		expect(screen.queryByText('ancienne définition')).toBeNull();
	});

	it('given_french_dictionary_and_voice_when_downloaded_then_each_uses_its_own_catalog_id', async () => {
		// given
		const dictionary = {
			ID: 'fr-fr', Name: 'French dictionary', Language: 'fr', Installed: false,
			DownloadBytes: 86864423, License: 'CC BY-SA 4.0',
			Source: { AdditionalURLs: ['https://example.com'], LicenseURL: 'https://example.com/license' }
		};
		const voice = { ID: 'fr_FR-siwis-medium', Name: 'French voice', Language: 'fr', Installed: false, SizeBytes: 63221984 };
		call.mockImplementation(async (method) => {
			if (method === 'Dictionaries') return [dictionary];
			if (method === 'ListVoices') return [voice];
			return {};
		});
		render(Page);
		await userEvent.click(screen.getByRole('button', { name: 'Offline content' }));
		// when
		const dictionaryCard = (await screen.findByText('French dictionary')).closest('article');
		const voiceCard = (await screen.findByText('French voice')).closest('.voice-pack');
		expect(within(dictionaryCard).getByText(/Language: French definitions/)).toBeTruthy();
		await userEvent.click(within(dictionaryCard).getByRole('button', { name: 'Download' }));
		await userEvent.click(within(voiceCard).getByRole('button', { name: 'Download' }));
		// then
		expect(call).toHaveBeenCalledWith('Download', 'fr-fr');
		expect(call).toHaveBeenCalledWith('DownloadVoicePack', 'fr_FR-siwis-medium');
	});

	it('given_french_pronunciation_when_play_is_requested_then_french_language_and_ipa_are_sent', async () => {
		// given
		render(Page);
		await search('bonjour');
		send('search-1', 'fr', found(entry('bonjour', 'salutation', {
			Language: 'fr', Pronunciations: [{ Text: '/bɔ̃ʒuʁ/', Label: '🇫🇷', LabelTitle: 'French' }]
		})));
		const play = await screen.findByRole('button', { name: /play pronunciation.*french/i });
		// when
		await userEvent.click(play);
		// then
		await waitFor(() => expect(call).toHaveBeenCalledWith('PlayPronunciation', expect.objectContaining({
			Language: 'fr', Headword: 'bonjour', Text: '/bɔ̃ʒuʁ/'
		})));
	});
});
