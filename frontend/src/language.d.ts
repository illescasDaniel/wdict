export type InterfaceLanguage = 'en' | 'es';

export interface InterfaceLanguageController {
	subscribe(run: (language: InterfaceLanguage) => void): () => void;
	setLanguage(language: InterfaceLanguage): Promise<void>;
}

export interface LanguagePreferenceStorage {
	getItem(key: string): string | null;
	setItem(key: string, value: string): void;
}

export type LookupMessage =
	| 'searching'
	| 'invalidQuery'
	| 'sourceUnavailable'
	| 'silentUnavailable'
	| 'notFound'
	| 'lookupFailed';

export type DictionaryMessage = 'metadataUnavailable' | 'installationFailed';
