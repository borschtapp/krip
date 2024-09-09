package utils

import "strings"

var domainLangMap = map[string]string{
	"ad": "ca-ad",
	"af": "ps-af",
	"ai": "es-ai",
	"al": "sq-al",
	"am": "hy-am",
	"ao": "pt-ao",
	"ar": "es-ar",
	"at": "de-at",
	"au": "en-au",
	"az": "az-az",
	"bb": "es-bb",
	"bd": "bn-bd",
	"bg": "bg-bg",
	"bh": "ar-bh",
	"bm": "es-bm",
	"bn": "ms-bn",
	"br": "pt-BR",
	"bs": "es-bs",
	"bt": "dz-bt",
	"bw": "tn-bw",
	"bz": "es-bz",
	"cf": "fr-cf",
	"cl": "es-cl",
	"co": "es-co",
	"cu": "es-cu",
	"cy": "el-cy",
	"cz": "cs-cz",
	"de": "de-de",
	"dm": "es-dm",
	"ec": "es-ec",
	"eg": "ar-eg",
	"er": "ti-er",
	"et": "am-et",
	"fi": "fi-fi",
	"fr": "fr-fr",
	"ga": "fr-ga",
	"gd": "es-gd",
	"ge": "ka-ge",
	"gl": "kl-gl",
	"gm": "ff-gm",
	"gn": "fr-gn",
	"gp": "fr-gp",
	"gr": "el-gr",
	"gt": "es-gt",
	"gy": "es-gy",
	"hn": "es-hn",
	"hu": "hu-hu",
	"ie": "ga-ie",
	"il": "he-il",
	"iq": "ar-iq",
	"ir": "fa-ir",
	"it": "it-it",
	"jo": "ar-jo",
	"jp": "ja-jp",
	"kg": "ky-kg",
	"kw": "ar-kw",
	"lb": "ar-lb",
	"li": "de-li",
	"lr": "ff-lr",
	"ls": "st-ls",
	"lt": "lt-lt",
	"lv": "lv-lv",
	"ly": "ar-ly",
	"mc": "fr-mc",
	"me": "sr-me",
	"ml": "fr-ml",
	"mn": "mn-mn",
	"ms": "es-ms",
	"mt": "mt-mt",
	"mu": "fr-mu",
	"mv": "dv-mv",
	"mw": "ny-mw",
	"mx": "es-mx",
	"my": "ms-my",
	"mz": "pt-mz",
	"ne": "fr-ne",
	"ni": "es-ni",
	"nl": "nl-nl",
	"no": "no-no",
	"np": "ne-np",
	"om": "ar-om",
	"pa": "es-pa",
	"ph": "es-ph",
	"pl": "pl-pl",
	"pt": "pt-pt",
	"qa": "ar-qa",
	"re": "fr-re",
	"ro": "ro-ro",
	"ru": "ru-ru",
	"sc": "fr-sc",
	"sd": "ar-sd",
	"se": "sv-se",
	"si": "sl-si",
	"tl": "pt-tl",
	"to": "to-to",
	"tr": "tr-tr",
	"tw": "zh-tw",
	"tz": "sw-tz",
	"ua": "uk-ua",
	"ug": "lg-ug",
	"uk": "en-gb",
	"uy": "es-uy",
	"ve": "es-ve",
	"vn": "vi-vn",
	"vu": "fr-vu",
	"yt": "fr-yt",
}

func LanguageByDomain(tld string) (string, bool) {
	if len(tld) == 2 {
		if lang, ok := domainLangMap[tld]; ok {
			return lang, true
		}
	}

	return "", false
}

func CleanupLang(lang string) string {
	lang = strings.ReplaceAll(lang, "_", "-")
	return lang
}