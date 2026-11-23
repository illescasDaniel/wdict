import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte';
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

const dictionary = {
	ID: 'one', Name: 'English dictionary', DownloadBytes: 1000000, Installed: false,
	License: 'CC BY-SA',
	Source: { AdditionalURLs: ['https://example.com'], LicenseURL: 'https://example.com/license' }
};
const entry = {
	Headword: 'hello',
	Meanings: [{ PartOfSpeech: 'noun', Definitions: ['a greeting'] }],
	Source: { Name: 'Example', Attribution: 'Example contributors', URL: 'https://example.com', LicenseURL: 'https://example.com/license' }
};
let deliver;
beforeEach(() => {
	localStorage.clear();
	call.mockReset();
	onLookupResult.mockImplementation(async (handler) => { deliver = handler; return () => {}; });
	call.mockImplementation(async (method) => {
		if (method === 'Dictionaries') return [dictionary];
		if (method === 'Progress') return {};
		if (method === 'Lookup') return 'search-1';
		return null;
	});
});
afterEach(() => {
	cleanup();
	vi.restoreAllMocks();
	localStorage.clear();
	document.documentElement.lang = 'en';
});

describe('interface language', () => {
	it('given_no_preference_when_started_then_english_is_selected', () => {
		// given
		render(Page);
		// when
		const trigger = screen.getByRole('button', { name: 'Language: English' });
		// then
		expect(trigger.textContent).toBe('English');
		expect(screen.getByRole('heading', { name: 'Look up a word' })).toBeTruthy();
		expect(document.documentElement.lang).toBe('en');
	});

	it('given_lookup_when_switching_both_ways_then_text_accessibility_and_input_update_without_calls', async () => {
		// given
		render(Page);
		const input = screen.getByLabelText('Word or phrase');
		await fireEvent.input(input, { target: { value: 'hello' } });
		const callsBefore = call.mock.calls.length;
		// when
		await chooseLanguage('es');
		// then
		await screen.findByRole('heading', { name: 'Buscar una palabra' });
		expect(screen.getByLabelText('Palabra o expresión').value).toBe('hello');
		expect(screen.getByRole('button', { name: 'Apariencia: Sistema' }).title).toBe('Apariencia');
		expect(document.documentElement.lang).toBe('es');
		expect(call.mock.calls.length).toBe(callsBefore);
		await chooseLanguage('en');
		await screen.findByRole('heading', { name: 'Look up a word' });
		expect(screen.getByLabelText('Word or phrase').value).toBe('hello');
		expect(document.documentElement.lang).toBe('en');
	});

	it('given_spanish_selected_when_restarted_then_preference_is_restored', async () => {
		// given
		render(Page);
		await chooseLanguage('es');
		await screen.findByRole('heading', { name: 'Buscar una palabra' });
		// when
		cleanup();
		render(Page);
		// then
		await screen.findByRole('heading', { name: 'Buscar una palabra' });
		expect(screen.getByRole('button', { name: /^Idioma: Español/ })).toBeTruthy();
	});

	it('given_unreadable_preference_when_started_then_english_is_usable', () => {
		// given
		vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => { throw Error('denied'); });
		// when
		render(Page);
		// then
		expect(screen.getByRole('button', { name: /^Language: English/ })).toBeTruthy();
		expect(screen.getByRole('heading', { name: 'Look up a word' })).toBeTruthy();
	});

	it('given_write_failure_when_switching_then_spanish_works_in_session', async () => {
		// given
		vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => { throw Error('denied'); });
		render(Page);
		// when
		await chooseLanguage('es');
		// then
		await screen.findByRole('heading', { name: 'Buscar una palabra' });
		expect(screen.getByRole('button', { name: /^Idioma: Español/ })).toBeTruthy();
	});

	it('given_cached_result_when_language_changes_then_date_and_labels_change_but_content_remains', async () => {
		// given
		render(Page);
		await fireEvent.input(screen.getByLabelText('Word or phrase'), { target: { value: 'hello' } });
		await fireEvent.click(screen.getByRole('button', { name: 'Look up' }));
		await waitFor(() => expect(call).toHaveBeenCalledWith('Lookup', 'hello'));
		const fetched = '2026-10-01T12:00:00Z';
		deliver({ SearchID: 'search-1', Language: 'en', Result: { Status: 'found', Entries: [entry], Origin: 'disk_cache', Stale: true, FetchedAt: fetched } });
		deliver({ SearchID: 'search-1', Language: 'es', Result: { Status: 'not_found' } });
		await screen.findByText('a greeting');
		expect(screen.getByText(new RegExp(new Date(fetched).toLocaleDateString('en').replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))).toBeTruthy();
		// when
		await chooseLanguage('es');
		// then
		await screen.findByRole('heading', { name: 'Buscar una palabra' });
		expect(screen.getByText('a greeting')).toBeTruthy();
		expect(screen.getByText(/Example contributors/)).toBeTruthy();
		expect(screen.getByText(new RegExp(new Date(fetched).toLocaleDateString('es').replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))).toBeTruthy();
	});

	it('given_dictionaries_view_and_metadata_failure_when_switching_then_error_and_retry_retranslate', async () => {
		// given
		call.mockImplementation(async (method) => {
			if (method === 'Dictionaries') throw Error('unavailable');
			return {};
		});
		render(Page);
		await fireEvent.click(screen.getByRole('button', { name: 'Offline content' }));
		const original = (await screen.findByText('Unable to load dictionary information.')).textContent;
		// when
		await chooseLanguage('es');
		// then
		await screen.findByRole('heading', { name: 'Contenido sin conexión' });
		expect(screen.queryByText(original)).toBeNull();
		expect(screen.getByRole('alert').textContent).toBe('No se pudo cargar la información de los diccionarios.');
		call.mockImplementation(async (method) => method === 'Dictionaries' ? [dictionary] : {});
		await userEvent.click(screen.getByRole('button', { name: 'Reintentar cargar la información de los diccionarios' }));
		await screen.findByText('English dictionary');
		expect(screen.queryByRole('alert')).toBeNull();
	});

	it('given_translation_catalogs_when_compared_then_keys_and_interpolation_parameters_match', () => {
		// given
		const english = resources.en.translation;
		const spanish = resources.es.translation;
		const parameters = (message) => [...message.matchAll(/{{\\s*([^{}]+?)\\s*}}/g)].map((match) => match[1]).sort();
		// when
		const keys = Object.keys(english).sort();
		// then
		expect(Object.keys(spanish).sort()).toEqual(keys);
		for (const key of keys) {
			expect(spanish[key]).toBeTruthy();
			expect(parameters(spanish[key])).toEqual(parameters(english[key]));
		}
	});

	it('given_spanish_interface_when_lookup_returns_both_languages_then_both_definitions_remain_original', async () => {
		// given
		render(Page);
		await chooseLanguage('es');
		await fireEvent.input(screen.getByLabelText('Palabra o expresión'), { target: { value: 'sol' } });
		// when
		await userEvent.click(screen.getByRole('button', { name: 'Buscar' }));
		await waitFor(() => expect(call).toHaveBeenCalledWith('Lookup', 'sol'));
		deliver({ SearchID: 'search-1', Language: 'en', Result: { Status: 'found', Entries: [entry], Origin: 'online' } });
		deliver({ SearchID: 'search-1', Language: 'es', Result: { Status: 'found', Entries: [{ ...entry, Headword: 'sol', Meanings: [{ Definitions: ['estrella luminosa'] }] }], Origin: 'offline' } });
		// then
		await screen.findByText('a greeting');
		expect(screen.getByText('estrella luminosa')).toBeTruthy();
		expect(screen.getByRole('heading', { name: 'Inglés · hello' })).toBeTruthy();
		expect(screen.getByRole('heading', { name: 'Español · sol' })).toBeTruthy();
	});

	it('given_dictionaries_view_when_switching_to_spanish_and_back_then_the_same_view_remains_open', async () => {
		// given
		render(Page);
		await userEvent.click(screen.getByRole('button', { name: 'Offline content' }));
		await screen.findByText('English dictionary');
		// when
		await chooseLanguage('es');
		await screen.findByRole('heading', { name: 'Contenido sin conexión' });
		await chooseLanguage('en');
		// then
		await screen.findByRole('heading', { name: 'Offline content' });
		expect(screen.getByText('English dictionary')).toBeTruthy();
	});

	it('given_unsafe_result_links_when_language_changes_then_links_remain_blocked', async () => {
		// given
		render(Page);
		await fireEvent.click(screen.getByRole('button', { name: 'Look up' }));
		await waitFor(() => expect(call).toHaveBeenCalledWith('Lookup', ''));
		const unsafe = { ...entry, Source: { ...entry.Source, URL: 'javascript:alert(1)', LicenseURL: 'javascript:alert(2)', AdditionalURLs: ['javascript:alert(3)'] } };
		deliver({ SearchID: 'search-1', Language: 'en', Result: { Status: 'found', Entries: [unsafe], Origin: 'online' } });
		deliver({ SearchID: 'search-1', Language: 'es', Result: { Status: 'not_found' } });
		await screen.findByText('a greeting');
		// when
		await chooseLanguage('es');
		// then
		await screen.findByText('No se dispone de información sobre la fuente o la licencia de esta entrada guardada.');
		expect(screen.queryByRole('link', { name: 'Fuente' })).toBeNull();
		expect(screen.queryByRole('link', { name: 'Licencia' })).toBeNull();
		expect(screen.queryByRole('link', { name: 'Más información' })).toBeNull();
		expect(screen.getByText('a greeting')).toBeTruthy();
	});

	it('given_lookup_unavailable_when_language_changes_then_error_and_retry_are_translated_and_work', async () => {
		// given
		render(Page);
		await fireEvent.click(screen.getByRole('button', { name: 'Look up' }));
		await waitFor(() => expect(call).toHaveBeenCalledWith('Lookup', ''));
		deliver({ SearchID: 'search-1', Language: 'en', Result: { Status: 'source_unavailable' } });
		deliver({ SearchID: 'search-1', Language: 'es', Result: { Status: 'not_found' } });
		deliver({ SearchID: 'search-1', Language: 'fr', Result: { Status: 'not_found' } });
		await screen.findByText('The dictionary is unavailable. Check your connection and try again.');
		// when
		await chooseLanguage('es');
		// then
		await screen.findByText('El diccionario no está disponible. Comprueba tu conexión y vuelve a intentarlo.');
		expect(screen.getByText('La búsqueda en línea en Inglés no está disponible.')).toBeTruthy();
		await userEvent.click(screen.getByRole('button', { name: 'Reintentar en Inglés' }));
		expect(call).toHaveBeenCalledWith('RetryLanguage', 'search-1', '', 'en');
	});

	it('given_failed_download_when_language_changes_then_error_and_retry_are_translated_and_work', async () => {
		// given
		let attempts = 0;
		call.mockImplementation(async (method) => {
			if (method === 'Dictionaries') return [dictionary];
			if (method === 'Progress') return {};
			if (method === 'Download' && ++attempts === 1) throw Error('network');
			return null;
		});
		render(Page);
		await userEvent.click(screen.getByRole('button', { name: 'Offline content' }));
		await screen.findByText('English dictionary');
		await userEvent.click(screen.getByRole('button', { name: 'Download' }));
		await screen.findByText('Couldn’t install the dictionary. Check your connection and available storage, then try again.');
		// when
		await chooseLanguage('es');
		// then
		await screen.findByText('No se pudo instalar el diccionario. Comprueba tu conexión y el espacio disponible y vuelve a intentarlo.');
		await userEvent.click(screen.getByRole('button', { name: 'Volver a intentar' }));
		await waitFor(() => expect(attempts).toBe(2));
		expect(screen.queryByRole('alert')).toBeNull();
	});

	it('given_invalid_saved_language_when_started_then_english_is_used', () => {
		// given
		vi.spyOn(Storage.prototype, 'getItem').mockReturnValue('unsupported');
		// when
		render(Page);
		// then
		expect(screen.getByRole('button', { name: /^Language: English/ })).toBeTruthy();
		expect(document.documentElement.lang).toBe('en');
	});

	it('given_download_in_progress_when_language_changes_then_download_remains_active', async () => {
		// given
		let finish;
		call.mockImplementation((method) => {
			if (method === 'Dictionaries') return Promise.resolve([dictionary]);
			if (method === 'Progress') return Promise.resolve({});
			if (method === 'Download') return new Promise((resolve) => { finish = resolve; });
			return Promise.resolve(null);
		});
		render(Page);
		await fireEvent.click(screen.getByRole('button', { name: 'Offline content' }));
		await screen.findByText('English dictionary');
		await fireEvent.click(screen.getByRole('button', { name: 'Download' }));
		expect(screen.getByRole('progressbar', { name: 'Dictionary download' })).toBeTruthy();
		const downloadCalls = call.mock.calls.filter(([method]) => method === 'Download').length;
		// when
		await chooseLanguage('es');
		// then
		await screen.findByRole('heading', { name: 'Contenido sin conexión' });
		expect(screen.getByRole('progressbar')).toBeTruthy();
		expect(call.mock.calls.filter(([method]) => method === 'Download')).toHaveLength(downloadCalls);
		finish();
	});

	it('given_spanish_interface_when_appearance_keyboard_changes_then_focus_and_selection_work', async () => {
		// given
		const user = userEvent.setup();
		render(Page);
		await chooseLanguage('es', user);
		const trigger = await screen.findByRole('button', { name: 'Apariencia: Sistema' });
		// when
		trigger.focus();
		await user.keyboard('{ArrowDown}');
		const system = screen.getByRole('menuitemradio', { name: 'Sistema' });
		expect(document.activeElement).toBe(system);
		await user.keyboard('{ArrowUp}{Enter}');
		// then
		expect(screen.queryByRole('menu')).toBeNull();
		expect(document.activeElement).toBe(screen.getByRole('button', { name: 'Apariencia: Oscuro' }));
		expect(document.documentElement.dataset.theme).toBe('dark');
	});
});

describe('footer popup menus', () => {
	const languageTrigger = () => screen.getByRole('button', { name: /^Language:/ });

	it('given_closed_menu_when_trigger_clicked_then_popup_opens_with_current_option_marked_and_focused', async () => {
		// given
		const user = userEvent.setup();
		render(Page);
		expect(languageTrigger().getAttribute('aria-haspopup')).toBe('menu');
		expect(languageTrigger().getAttribute('aria-expanded')).toBe('false');
		// when
		await user.click(languageTrigger());
		// then
		expect(languageTrigger().getAttribute('aria-expanded')).toBe('true');
		const options = screen.getAllByRole('menuitemradio').map((o) => o.textContent);
		expect(options).toEqual(['English', 'Español', 'Français']);
		const current = screen.getByRole('menuitemradio', { name: 'English' });
		expect(current.getAttribute('aria-checked')).toBe('true');
		expect(document.activeElement).toBe(current);
	});

	it('given_open_menu_when_option_selected_then_language_applies_popup_closes_and_focus_returns', async () => {
		// given
		const user = userEvent.setup();
		render(Page);
		await user.click(languageTrigger());
		// when
		await user.click(screen.getByRole('menuitemradio', { name: 'Français' }));
		// then
		await screen.findByRole('button', { name: /^Langue: Français/ });
		expect(screen.queryByRole('menu')).toBeNull();
		expect(document.documentElement.lang).toBe('fr');
		expect(document.activeElement).toBe(screen.getByRole('button', { name: /^Langue:/ }));
	});

	it('given_open_menu_when_escape_pressed_then_popup_closes_and_focus_returns_without_change', async () => {
		// given
		const user = userEvent.setup();
		render(Page);
		await user.click(languageTrigger());
		await user.keyboard('{ArrowDown}');
		// when
		await user.keyboard('{Escape}');
		// then
		expect(screen.queryByRole('menu')).toBeNull();
		expect(document.activeElement).toBe(languageTrigger());
		expect(languageTrigger().textContent).toBe('English');
	});

	it('given_open_menu_when_outside_clicked_then_popup_closes', async () => {
		// given
		const user = userEvent.setup();
		render(Page);
		await user.click(languageTrigger());
		// when
		await user.click(document.body);
		// then
		expect(screen.queryByRole('menu')).toBeNull();
		expect(languageTrigger().getAttribute('aria-expanded')).toBe('false');
	});

	it('given_one_menu_open_when_other_trigger_clicked_then_only_one_popup_is_open', async () => {
		// given
		const user = userEvent.setup();
		render(Page);
		await user.click(languageTrigger());
		// when
		await openAppearance(user);
		// then
		expect(screen.getAllByRole('menu')).toHaveLength(1);
		expect(screen.getByRole('menu', { name: 'Appearance' })).toBeTruthy();
		expect(languageTrigger().getAttribute('aria-expanded')).toBe('false');
	});

	it('given_open_menu_when_navigating_with_keys_then_focus_wraps_and_tab_closes', async () => {
		// given
		const user = userEvent.setup();
		render(Page);
		languageTrigger().focus();
		// when
		await user.keyboard('{ArrowDown}');
		await user.keyboard('{End}');
		expect(document.activeElement).toBe(screen.getByRole('menuitemradio', { name: 'Français' }));
		await user.keyboard('{ArrowDown}');
		expect(document.activeElement).toBe(screen.getByRole('menuitemradio', { name: 'English' }));
		await user.keyboard('{ArrowUp}');
		expect(document.activeElement).toBe(screen.getByRole('menuitemradio', { name: 'Français' }));
		await user.keyboard('{Home}');
		expect(document.activeElement).toBe(screen.getByRole('menuitemradio', { name: 'English' }));
		await user.keyboard('{Tab}');
		// then
		expect(screen.queryByRole('menu')).toBeNull();
	});

	it('given_trigger_focused_when_space_pressed_then_popup_opens_and_trigger_shows_current_mode', async () => {
		// given
		const user = userEvent.setup();
		render(Page);
		const trigger = screen.getByRole('button', { name: 'Appearance: System' });
		trigger.focus();
		// when
		await user.keyboard(' ');
		// then
		expect(screen.getByRole('menuitemradio', { name: 'System' }).getAttribute('aria-checked')).toBe('true');
		expect(document.activeElement).toBe(screen.getByRole('menuitemradio', { name: 'System' }));
	});
});
