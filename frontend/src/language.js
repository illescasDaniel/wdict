import { createInstance } from "i18next";

export const resources = {
  en: {
    translation: {
      offlineDictionaries: "Offline content",
      backLookup: "← Look up",
      heading: "Look up a word",
      intro: "Find English and Spanish definitions for words and phrases.",
      query: "Word or phrase",
      queryPlaceholder: "Enter a word or phrase",
      lookup: "Look up",
      privacy:
        "If an offline dictionary or saved entry cannot answer your search, it may be sent to the online dictionary.",
      searching: "Searching…",
      invalidQuery: "Enter a word or phrase of up to 256 characters.",
      sourceUnavailable:
        "The dictionary is unavailable. Check your connection and try again.",
      silentUnavailable: "No definitions available right now.",
      notFound: "No entry found. Try another word or phrase.",
      lookupFailed: "The dictionary is unavailable. Please try again.",
      english: "English",
      spanish: "Spanish",
      staleResult: "Stale saved result",
      offlineResult: "Offline result",
      onlineResult: "Online result",
      savedResult: "Saved result",
      source: "Source",
      pronunciation: "Pronunciation",
      playPronunciation: "Play pronunciation",
      recordingSource: "Recording",
      playbackFailed: "Pronunciation playback failed. Try again.",
      license: "License",
      moreInformation: "More information",
      missingAttribution:
        "Source or license information unavailable for this cached entry.",
      languageUnavailable: "{{language}} online lookup unavailable.",
      retryLanguage: "Retry {{language}}",
      searchingLanguage: "Searching {{language}}…",
      noDefinition: "No definition found.",
      checkingAlternatives: "Checking possible alternatives…",
      alternativesUnavailable: "Suggestions are unavailable right now.",
      alternativesTimedOut: "Suggestions timed out.",
      alternativesNeedDictionary: "Install the offline dictionary to see suggestions.",
      noAlternatives: "No suggestions found.",
      possibleBaseForms: "Possible base forms",
      didYouMean: "Did you mean?",
      previousResult: "Previous result for “{{query}}”",
      offlineHeading: "Look up words offline",
      offlineIntro: "Download dictionaries for offline definitions.",
      viewDictionaries: "View offline dictionaries →",
      dictionariesIntro:
        "Download a dictionary to look up words without a connection.",
      metadataUnavailable: "Unable to load dictionary information.",
      retryMetadata: "Retry loading dictionary information",
      installed: "Installed",
      installing: "Installing",
      downloading: "Downloading",
      notInstalled: "Not installed",
      dictionaryLanguage: "Language: {{language}} definitions",
      downloadLabel: "Download",
      dictionaryDownload: "Dictionary download",
      progress: "{{read}} MB of {{total}} MB",
      checking: "Checking and installing the dictionary…",
      tryAgain: "Try again",
      ready: "Ready to use offline.",
      installationFailed:
        "Couldn’t install the dictionary. Check your connection and available storage, then try again.",
      voiceInstallationFailed:
        "Couldn’t install the voice pack. Check your connection and available storage, then try again.",
      continueLookup: "You can continue looking up words while it downloads.",
      loadingMetadata: "Loading dictionary information…",
      language: "Language",
      appearance: "Appearance",
      system: "System",
      light: "Light",
      dark: "Dark",
      naturalVoices: "Natural voices",
      piperOptional: "Piper · optional",
      voiceDownload: "Natural voices download",
    },
  },
  es: {
    translation: {
      offlineDictionaries: "Contenido sin conexión",
      backLookup: "← Buscar",
      heading: "Buscar una palabra",
      intro:
        "Encuentra definiciones en inglés y español de palabras y expresiones.",
      query: "Palabra o expresión",
      queryPlaceholder: "Introduce una palabra o expresión",
      lookup: "Buscar",
      privacy:
        "Si un diccionario sin conexión o una entrada guardada no puede responder a tu búsqueda, esta puede enviarse al diccionario en línea.",
      searching: "Buscando…",
      invalidQuery:
        "Introduce una palabra o expresión de hasta 256 caracteres.",
      sourceUnavailable:
        "El diccionario no está disponible. Comprueba tu conexión y vuelve a intentarlo.",
      silentUnavailable: "No hay definiciones disponibles en este momento.",
      notFound:
        "No se encontró ninguna entrada. Prueba otra palabra o expresión.",
      lookupFailed: "El diccionario no está disponible. Vuelve a intentarlo.",
      english: "Inglés",
      spanish: "Español",
      staleResult: "Resultado guardado desactualizado",
      offlineResult: "Resultado sin conexión",
      onlineResult: "Resultado en línea",
      savedResult: "Resultado guardado",
      source: "Fuente",
      pronunciation: "Pronunciación",
      playPronunciation: "Reproducir la pronunciación",
      recordingSource: "Grabación",
      playbackFailed:
        "No se pudo reproducir la pronunciación. Vuelve a intentarlo.",
      license: "Licencia",
      moreInformation: "Más información",
      missingAttribution:
        "No se dispone de información sobre la fuente o la licencia de esta entrada guardada.",
      languageUnavailable:
        "La búsqueda en línea en {{language}} no está disponible.",
      retryLanguage: "Reintentar en {{language}}",
      searchingLanguage: "Buscando en {{language}}…",
      noDefinition: "No se encontró ninguna definición.",
      checkingAlternatives: "Buscando posibles alternativas…",
      alternativesUnavailable: "Las sugerencias no están disponibles ahora.",
      alternativesTimedOut: "Las sugerencias superaron el tiempo límite.",
      alternativesNeedDictionary: "Instala el diccionario sin conexión para ver sugerencias.",
      noAlternatives: "No se encontraron sugerencias.",
      possibleBaseForms: "Posibles formas base",
      didYouMean: "¿Quizás quisiste decir?",
      previousResult: "Resultado anterior para «{{query}}»",
      offlineHeading: "Buscar palabras sin conexión",
      offlineIntro:
        "Descarga diccionarios para obtener definiciones sin conexión.",
      viewDictionaries: "Ver diccionarios sin conexión →",
      dictionariesIntro:
        "Descarga un diccionario para buscar palabras sin conexión.",
      metadataUnavailable:
        "No se pudo cargar la información de los diccionarios.",
      retryMetadata: "Reintentar cargar la información de los diccionarios",
      installed: "Instalado",
      installing: "Instalando",
      downloading: "Descargando",
      notInstalled: "No instalado",
      dictionaryLanguage: "Idioma: definiciones en {{language}}",
      downloadLabel: "Descargar",
      dictionaryDownload: "Descarga del diccionario",
      progress: "{{read}} MB de {{total}} MB",
      checking: "Comprobando e instalando el diccionario…",
      tryAgain: "Volver a intentar",
      ready: "Listo para usar sin conexión.",
      installationFailed:
        "No se pudo instalar el diccionario. Comprueba tu conexión y el espacio disponible y vuelve a intentarlo.",
      voiceInstallationFailed:
        "No se pudo instalar el paquete de voz. Comprueba tu conexión y el espacio disponible y vuelve a intentarlo.",
      continueLookup: "Puedes seguir buscando palabras mientras se descarga.",
      loadingMetadata: "Cargando información de los diccionarios…",
      language: "Idioma",
      appearance: "Apariencia",
      system: "Sistema",
      light: "Claro",
      dark: "Oscuro",
      naturalVoices: "Voces naturales",
      piperOptional: "Piper · opcional",
      voiceDownload: "Descarga de voces naturales",
    },
  },
};

export function createLanguageController(storage) {
  const i18n = createInstance();
  let language = "en";
  try {
    const saved = storage?.getItem("wdict-language");
    if (saved === "es" || saved === "en") language = saved;
  } catch {}
  i18n.init({
    resources,
    lng: language,
    fallbackLng: "en",
    initAsync: false,
    interpolation: { escapeValue: false },
  });
  const subscribers = new Set();
  return {
    subscribe(run) {
      subscribers.add(run);
      run(language);
      return () => subscribers.delete(run);
    },
    get language() {
      return language;
    },
    translate(key, options) {
      return i18n.t(key, options);
    },
    async setLanguage(next) {
      if (next !== "en" && next !== "es") return;
      await i18n.changeLanguage(next);
      language = next;
      for (const subscriber of subscribers) subscriber(language);
      try {
        storage?.setItem("wdict-language", next);
      } catch {}
    },
  };
}
