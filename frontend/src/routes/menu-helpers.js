import userEvent from '@testing-library/user-event';
import { screen } from '@testing-library/svelte';

const languageNames = { en: 'English', es: 'Español', fr: 'Français' };

export async function chooseLanguage(code, user = userEvent) {
	await user.click(screen.getByRole('button', { name: /^(Language|Idioma|Langue):/ }));
	await user.click(screen.getByRole('menuitemradio', { name: languageNames[code] }));
}

export async function openAppearance(user = userEvent) {
	await user.click(screen.getByRole('button', { name: /^(Appearance|Apariencia|Apparence):/ }));
}

export async function chooseAppearance(name, user = userEvent) {
	await openAppearance(user);
	await user.click(screen.getByRole('menuitemradio', { name }));
}
