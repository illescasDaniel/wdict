package adapters

import "github.com/illescasDaniel/wdict/internal/core"

// Pinned natural-voices catalog (C1): the Piper runtime wheel and three
// curated voice packs, with pinned URLs, SHA-256 checksums and exact sizes.
// The sizes were pinned from Content-Length during the 2026-11-13 build
// (Hugging Face API tree + PyPI); the voices sit at the pinned revision
// 6b9cfaacb4fb2e02a9621227e3c1a48bec9710c3.
const (
	piperRuntimeID   = core.VoicePackID("piper-runtime")
	piperRuntimeName = "Piper runtime"

	piperPythonCopy       = "Python 3 is required for natural voices. Install Python 3 and retry."
	piperRuntimeFirstCopy = "Download the Piper runtime first."

	piperVoiceRevision = "6b9cfaacb4fb2e02a9621227e3c1a48bec9710c3"
	piperAttribution   = "rhasspy / OHF-Voice piper voices"
	piperLicenseURL    = "https://github.com/OHF-Voice/piper1-gpl/blob/main/COPYING"

	piperWheelURL  = "https://files.pythonhosted.org/packages/84/81/0112a7d510911f33018dc24023d1655bb772f84a4e95fe7f0180f66bbd17/piper_tts-1.8.0-cp39-abi3-manylinux_2_17_x86_64.manylinux2014_x86_64.manylinux_2_28_x86_64.whl"
	piperWheelSHA  = "25b4d3f31ff70c8fa7151908e00aaa5650cbdf16bca8fcf21299f3941b89a7d3"
	piperWheelSize = int64(34131442)
)

// voiceFileURL builds the pinned voice file URL at the pinned revision.
func voiceFileURL(lang, locale, speaker, quality, file string) string {
	return "https://huggingface.co/rhasspy/piper-voices/resolve/" + piperVoiceRevision + "/" + lang + "/" + locale + "/" + speaker + "/" + quality + "/" + file
}

// voiceModelCardURL points at the voice's model card at the pinned revision.
func voiceModelCardURL(lang, locale, speaker, quality string) string {
	return "https://huggingface.co/rhasspy/piper-voices/blob/" + piperVoiceRevision + "/" + lang + "/" + locale + "/" + speaker + "/" + quality + "/MODEL_CARD"
}

// voicePackEntry is one pinned catalog record. Voice entries carry two
// checksummed files: the .onnx model (size/sha) and its .onnx.json
// configuration (extraSize/extraSHA, pinned during the 2026-11-13 build).
type voicePackEntry struct {
	id        core.VoicePackID
	kind      core.VoicePackKind
	name      string // display name, e.g. "en_GB · alba (medium)"
	short     string // short voice id for the Piper result label, e.g. "alba"
	language  core.Language
	size      int64
	sha       string
	url       string
	extraURL  string
	extraSHA  string
	extraSize int64
	source    core.Source
}

// voiceCatalog is the code-only trusted list, in display order: the runtime
// first, then one curated voice per language (alba, amy, sharvard).
var voiceCatalog = []voicePackEntry{
	{
		id: piperRuntimeID, kind: core.VoicePackRuntime, name: piperRuntimeName,
		size: piperWheelSize, sha: piperWheelSHA, url: piperWheelURL,
		source: core.Source{ID: "piper-tts-1.8.0", Name: "Piper 1.8.0 (PyPI wheel)", URL: piperWheelURL, Attribution: piperAttribution, LicenseURL: piperLicenseURL},
	},
	{
		id: "en_GB-alba-medium", kind: core.VoicePackVoice, name: "en_GB · alba (medium)", short: "alba",
		language: core.English, size: 63221984, sha: "1f9da809b6da1ef4190bb5dd6cce4f4af803274539e9cf9e0693085150bac3d7",
		url:      voiceFileURL("en", "en_GB", "alba", "medium", "en_GB-alba-medium.onnx"),
		extraURL: voiceFileURL("en", "en_GB", "alba", "medium", "en_GB-alba-medium.onnx.json"), extraSHA: "6214603a0dd60e9dcd05502e23ef07fbe3d9af36d3189c8cc3eb129e1a099d6f", extraSize: 5471,
		source: core.Source{ID: "en_GB-alba-medium", Name: "Piper voice en_GB-alba-medium", URL: voiceFileURL("en", "en_GB", "alba", "medium", "en_GB-alba-medium.onnx"), Attribution: piperAttribution, LicenseURL: voiceModelCardURL("en", "en_GB", "alba", "medium")},
	},
	{
		id: "en_US-amy-medium", kind: core.VoicePackVoice, name: "en_US · amy (medium)", short: "amy",
		language: core.English, size: 63221984, sha: "cf9ccdec59f05adee47173f6fb4d89e8b0c5f30c178748d2572fe4efd1d2724c",
		url:      voiceFileURL("en", "en_US", "amy", "medium", "en_US-amy-medium.onnx"),
		extraURL: voiceFileURL("en", "en_US", "amy", "medium", "en_US-amy-medium.onnx.json"), extraSHA: "8c9b8aa3ee122c6a63678bd68e586a067c6775116d32e13b7ec919520d8948bf", extraSize: 5465,
		source: core.Source{ID: "en_US-amy-medium", Name: "Piper voice en_US-amy-medium", URL: voiceFileURL("en", "en_US", "amy", "medium", "en_US-amy-medium.onnx"), Attribution: piperAttribution, LicenseURL: voiceModelCardURL("en", "en_US", "amy", "medium")},
	},
	{
		id: "es_ES-sharvard-medium", kind: core.VoicePackVoice, name: "es_ES · sharvard (medium)", short: "sharvard",
		language: core.Spanish, size: 76753841, sha: "234dc4cf0fdb53ccb76b9b2c4d7ef1330364762674c1b78c1b62d8608c528a54",
		url:      voiceFileURL("es", "es_ES", "sharvard", "medium", "es_ES-sharvard-medium.onnx"),
		extraURL: voiceFileURL("es", "es_ES", "sharvard", "medium", "es_ES-sharvard-medium.onnx.json"), extraSHA: "1910940c6b480e0f7cbce5e0da45c8cbe2557ccbfd86a5d22cda141a3b65f521", extraSize: 5486,
		source: core.Source{ID: "es_ES-sharvard-medium", Name: "Piper voice es_ES-sharvard-medium", URL: voiceFileURL("es", "es_ES", "sharvard", "medium", "es_ES-sharvard-medium.onnx"), Attribution: piperAttribution, LicenseURL: voiceModelCardURL("es", "es_ES", "sharvard", "medium")},
	},
}
