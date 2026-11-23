import { createInstance } from "i18next";

export const resources = {
  en: {
    translation: {
      offlineDictionaries: "Offline content",
      backLookup: "← Look up",
      heading: "Look up a word",
      intro: "Find English, Spanish and French definitions for words and phrases.",
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
      french: "French",
      englishInline: "English",
      spanishInline: "Spanish",
      frenchInline: "French",
      staleResult: "Stale saved result",
      offlineResult: "Offline result",
      onlineResult: "Online result",
      savedResult: "Saved result",
      source: "Source",
      pronunciation: "Pronunciation",
      pronunciationUnavailable: "Pronunciation unavailable in this source",
      readWordAloud: "Read word aloud",
      spellingSynthesized:
        "Synthesized from spelling; source pronunciation unavailable",
      spellingNoSource:
        "eSpeak-NG is not available. Install it or use the AppImage to read this word.",
      spellingNoSynthesis: "Could not synthesize this word. Try again.",
      playPronunciation: "Play pronunciation",
      recordingSource: "Recording",
      playbackFailed: "Pronunciation playback failed. Try again.",
      playbackNoSource:
        "No recording, and eSpeak-NG is not installed. Install eSpeak-NG (or use the AppImage build) to hear this transcription.",
      playbackNoSynthesis: "Could not synthesise this pronunciation.",
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
      alternativesNeedDictionary:
        "Install the offline dictionary to see suggestions.",
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
      voiceBlockedPythonMissing:
        "Python 3 is required for natural voices. Install Python 3 and retry.",
      voiceBlockedRuntimeFirst: "Download the Piper runtime first.",
    },
  },
  es: {
    translation: {
      offlineDictionaries: "Contenido sin conexión",
      backLookup: "← Buscar",
      heading: "Buscar una palabra",
      intro:
        "Encuentra definiciones en inglés, español y francés de palabras y expresiones.",
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
      french: "Francés",
      englishInline: "Inglés",
      spanishInline: "Español",
      frenchInline: "Francés",
      staleResult: "Resultado guardado desactualizado",
      offlineResult: "Resultado sin conexión",
      onlineResult: "Resultado en línea",
      savedResult: "Resultado guardado",
      source: "Fuente",
      pronunciation: "Pronunciación",
      pronunciationUnavailable: "Pronunciación no disponible en esta fuente",
      readWordAloud: "Leer palabra en voz alta",
      spellingSynthesized:
        "Voz sintetizada a partir de la escritura; pronunciación de la fuente no disponible",
      spellingNoSource:
        "eSpeak-NG no está disponible. Instálalo o usa la AppImage para leer esta palabra.",
      spellingNoSynthesis:
        "No se pudo sintetizar esta palabra. Vuelve a intentarlo.",
      playPronunciation: "Reproducir la pronunciación",
      recordingSource: "Grabación",
      playbackFailed:
        "No se pudo reproducir la pronunciación. Vuelve a intentarlo.",
      playbackNoSource:
        "No hay grabación y eSpeak-NG no está instalado. Instala eSpeak-NG (o usa la AppImage) para escuchar esta transcripción.",
      playbackNoSynthesis: "No se pudo sintetizar esta pronunciación.",
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
      alternativesNeedDictionary:
        "Instala el diccionario sin conexión para ver sugerencias.",
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
      voiceBlockedPythonMissing:
        "Se necesita Python 3 para las voces naturales. Instala Python 3 y vuelve a intentarlo.",
      voiceBlockedRuntimeFirst: "Descarga primero el runtime de Piper.",
    },
  },
  fr: {
    translation: {
      offlineDictionaries: "Contenu hors ligne",
      backLookup: "← Rechercher",
      heading: "Rechercher un mot",
      intro: "Trouvez des définitions en anglais, espagnol et français pour des mots et expressions.",
      query: "Mot ou expression",
      queryPlaceholder: "Saisissez un mot ou une expression",
      lookup: "Rechercher",
      privacy: "Si un dictionnaire hors ligne ou une entrée enregistrée ne peut pas répondre à votre recherche, celle-ci peut être envoyée au dictionnaire en ligne.",
      searching: "Recherche…",
      invalidQuery: "Saisissez un mot ou une expression de 256 caractères maximum.",
      sourceUnavailable: "Le dictionnaire est indisponible. Vérifiez votre connexion et réessayez.",
      silentUnavailable: "Aucune définition disponible pour le moment.",
      notFound: "Aucune entrée trouvée. Essayez un autre mot ou une autre expression.",
      lookupFailed: "Le dictionnaire est indisponible. Veuillez réessayer.",
      english: "Anglais",
      spanish: "Espagnol",
      french: "Français",
      englishInline: "anglais",
      spanishInline: "espagnol",
      frenchInline: "français",
      staleResult: "Résultat enregistré périmé",
      offlineResult: "Résultat hors ligne",
      onlineResult: "Résultat en ligne",
      savedResult: "Résultat enregistré",
      source: "Source",
      pronunciation: "Prononciation",
      pronunciationUnavailable: "Prononciation indisponible dans cette source",
      readWordAloud: "Lire le mot à voix haute",
      spellingSynthesized: "Voix synthétisée à partir de l’orthographe ; prononciation de la source indisponible",
      spellingNoSource: "eSpeak-NG est indisponible. Installez-le ou utilisez l’AppImage pour lire ce mot.",
      spellingNoSynthesis: "Impossible de synthétiser ce mot. Réessayez.",
      playPronunciation: "Écouter la prononciation",
      recordingSource: "Enregistrement",
      playbackFailed: "Échec de la lecture de la prononciation. Réessayez.",
      playbackNoSource: "eSpeak-NG n’est pas installé. Installez-le ou utilisez l’AppImage pour écouter cette transcription.",
      playbackNoSynthesis: "Impossible de synthétiser cette prononciation.",
      license: "Licence",
      moreInformation: "Plus d’informations",
      missingAttribution: "Les informations de source ou de licence sont indisponibles pour cette entrée enregistrée.",
      languageUnavailable: "Recherche en ligne en {{language}} indisponible.",
      retryLanguage: "Réessayer en {{language}}",
      searchingLanguage: "Recherche en {{language}}…",
      noDefinition: "Aucune définition trouvée.",
      checkingAlternatives: "Recherche d’autres possibilités…",
      alternativesUnavailable: "Suggestions indisponibles pour le moment.",
      alternativesTimedOut: "Le délai de recherche des suggestions est dépassé.",
      alternativesNeedDictionary: "Installez le dictionnaire hors ligne pour voir les suggestions.",
      noAlternatives: "Aucune suggestion trouvée.",
      possibleBaseForms: "Formes de base possibles",
      didYouMean: "Vouliez-vous dire ?",
      previousResult: "Résultat précédent pour « {{query}} »",
      offlineHeading: "Rechercher hors ligne",
      offlineIntro: "Téléchargez des dictionnaires pour obtenir des définitions hors ligne.",
      viewDictionaries: "Voir les dictionnaires hors ligne →",
      dictionariesIntro: "Téléchargez un dictionnaire pour rechercher des mots sans connexion.",
      metadataUnavailable: "Impossible de charger les informations des dictionnaires.",
      retryMetadata: "Réessayer de charger les informations des dictionnaires",
      installed: "Installé",
      installing: "Installation",
      downloading: "Téléchargement",
      notInstalled: "Non installé",
      dictionaryLanguage: "Langue : définitions en {{language}}",
      downloadLabel: "Télécharger",
      dictionaryDownload: "Téléchargement du dictionnaire",
      progress: "{{read}} Mo sur {{total}} Mo",
      checking: "Vérification et installation du dictionnaire…",
      tryAgain: "Réessayer",
      ready: "Prêt à être utilisé hors ligne.",
      installationFailed: "Impossible d’installer le dictionnaire. Vérifiez votre connexion et l’espace disponible, puis réessayez.",
      voiceInstallationFailed: "Impossible d’installer la voix. Vérifiez votre connexion et l’espace disponible, puis réessayez.",
      continueLookup: "Vous pouvez continuer à rechercher des mots pendant le téléchargement.",
      loadingMetadata: "Chargement des informations des dictionnaires…",
      language: "Langue",
      appearance: "Apparence",
      system: "Système",
      light: "Clair",
      dark: "Sombre",
      naturalVoices: "Voix naturelles",
      piperOptional: "Piper · facultatif",
      voiceDownload: "Téléchargement des voix naturelles",
      voiceBlockedPythonMissing: "Python 3 est nécessaire pour les voix naturelles. Installez Python 3 et réessayez.",
      voiceBlockedRuntimeFirst: "Téléchargez d’abord le moteur Piper.",
    },
  },
};

// Messages that embed a language name inside a sentence. The page passes the
// standalone (title-case) label as `language`; these messages use the
// `<name>Inline` form instead, which French needs in lowercase.
const inlineLanguageMessages = new Set([
  "languageUnavailable",
  "retryLanguage",
  "searchingLanguage",
  "dictionaryLanguage",
]);
const languageLabelKeys = ["english", "spanish", "french"];

export function createLanguageController(storage) {
  const i18n = createInstance();
  let language = "en";
  try {
    const saved = storage?.getItem("wdict-language");
    if (["en", "es", "fr"].includes(saved)) language = saved;
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
      if (inlineLanguageMessages.has(key) && typeof options?.language === "string") {
        const labelKey = languageLabelKeys.find((name) => i18n.t(name) === options.language);
        if (labelKey) options = { ...options, language: i18n.t(`${labelKey}Inline`) };
      }
      return i18n.t(key, options);
    },
    async setLanguage(next) {
      if (!["en", "es", "fr"].includes(next)) return;
      await i18n.changeLanguage(next);
      language = next;
      for (const subscriber of subscribers) subscriber(language);
      try {
        storage?.setItem("wdict-language", next);
      } catch {}
    },
  };
}
