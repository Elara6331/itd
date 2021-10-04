package translit

import (
	"strings"
)

// Maps stores transliteration maps as slices to preserve order.
// Some of these maps were sourced from https://codeberg.org/Freeyourgadget/Gadgetbridge
var Maps = map[string][]string{
	"eASCII": {
		"œ", "oe",
		"ª", "a",
		"°", "o",
		"«", `"`,
		"»", `"`,
	},
	"Scandinavian": {
		"Æ", "Ae",
		"æ", "ae",
		"Ø", "Oe",
		"ø", "oe",
		"Å", "Aa",
		"å", "aa",
	},
	"German": {
		"ä", "ae",
		"ö", "oe",
		"ü", "ue",
		"Ä", "Ae",
		"Ö", "Oe",
		"Ü", "Ue",
		"ß", "ss",
		"ẞ", "SS",
	},
	"Hebrew": {
		"א", "a",
		"ב", "b",
		"ג", "g",
		"ד", "d",
		"ה", "h",
		"ו", "u",
		"ז", "z",
		"ח", "kh",
		"ט", "t",
		"י", "y",
		"כ", "c",
		"ל", "l",
		"מ", "m",
		"נ", "n",
		"ס", "s",
		"ע", "'",
		"פ", "p",
		"צ", "ts",
		"ק", "k",
		"ר", "r",
		"ש", "sh",
		"ת", "th",
		"ף", "f",
		"ץ", "ts",
		"ך", "ch",
		"ם", "m",
		"ן", "n",
	},
	"Greek": {
		"α", "a",
		"ά", "a",
		"β", "v",
		"γ", "g",
		"δ", "d",
		"ε", "e",
		"έ", "e",
		"ζ", "z",
		"η", "i",
		"ή", "i",
		"θ", "th",
		"ι", "i",
		"ί", "i",
		"ϊ", "i",
		"ΐ", "i",
		"κ", "k",
		"λ", "l",
		"μ", "m",
		"ν", "n",
		"ξ", "ks",
		"ο", "o",
		"ό", "o",
		"π", "p",
		"ρ", "r",
		"σ", "s",
		"ς", "s",
		"τ", "t",
		"υ", "y",
		"ύ", "y",
		"ϋ", "y",
		"ΰ", "y",
		"φ", "f",
		"χ", "ch",
		"ψ", "ps",
		"ω", "o",
		"ώ", "o",
		"Α", "A",
		"Ά", "A",
		"Β", "B",
		"Γ", "G",
		"Δ", "D",
		"Ε", "E",
		"Έ", "E",
		"Ζ", "Z",
		"Η", "I",
		"Ή", "I",
		"Θ", "TH",
		"Ι", "I",
		"Ί", "I",
		"Ϊ", "I",
		"Κ", "K",
		"Λ", "L",
		"Μ", "M",
		"Ν", "N",
		"Ξ", "KS",
		"Ο", "O",
		"Ό", "O",
		"Π", "P",
		"Ρ", "R",
		"Σ", "S",
		"Τ", "T",
		"Υ", "Y",
		"Ύ", "Y",
		"Ϋ", "Y",
		"Φ", "F",
		"Χ", "CH",
		"Ψ", "PS",
		"Ω", "O",
		"Ώ", "O",
	},
	"Russian": {
		"Ё", "Йo",
		"ё", "йo",
	},
	"Ukranian": {
		"ґ", "gh",
		"є", "je",
		"і", "i",
		"ї", "ji",
		"Ґ", "GH",
		"Є", "JE",
		"І", "I",
		"Ї", "JI",
	},
	"Arabic": {
		"ا", "a",
		"ب", "b",
		"ت", "t",
		"ث", "th",
		"ج", "j",
		"ح", "7",
		"خ", "5",
		"د", "d",
		"ذ", "th",
		"ر", "r",
		"ز", "z",
		"س", "s",
		"ش", "sh",
		"ص", "9",
		"ض", "9'",
		"ط", "6",
		"ظ", "6'",
		"ع", "3",
		"غ", "3'",
		"ف", "f",
		"ق", "q",
		"ك", "k",
		"ل", "l",
		"م", "m",
		"ن", "n",
		"ه", "h",
		"و", "w",
		"ي", "y",
		"ى", "a",
		"ﺓ", "",
		"آ", "2",
		"ئ", "2",
		"إ", "2",
		"ؤ", "2",
		"أ", "2",
		"ء", "2",
		"٠", "0",
		"١", "1",
		"٢", "2",
		"٣", "3",
		"٤", "4",
		"٥", "5",
		"٦", "6",
		"٧", "7",
		"٨", "8",
		"٩", "9",
	},
	"Farsi": {
		"پ", "p",
		"چ", "ch",
		"ژ", "zh",
		"ک", "k",
		"گ", "g",
		"ی", "y",
		"\u200c", " ",
		"؟", "?",
		"٪", "%",
		"؛", ";",
		"،", ":",
		"۱", "1",
		"۲", "2",
		"۳", "3",
		"۴", "4",
		"۵", "5",
		"۶", "6",
		"۷", "7",
		"۸", "8",
		"۹", "9",
		"۰", "0",
		"»", "<",
		"«", ">",
		"ِ", "e",
		"َ", "a",
		"ُ", "o",
		"ّ", "",
	},
	"Polish": {
		"Ł", "L",
		"ł", "l",
	},
	"Lithuanian": {
		"ą", "a",
		"č", "c",
		"ę", "e",
		"ė", "e",
		"į", "i",
		"š", "s",
		"ų", "u",
		"ū", "u",
		"ž", "z",
	},
	"Estonian": {
		"ä", "a",
		"Ä", "A",
		"ö", "o",
		"õ", "o",
		"Ö", "O",
		"Õ", "O",
		"ü", "u",
		"Ü", "U",
	},
	"Icelandic": {
		"Þ", "Th",
		"þ", "th",
		"Ð", "D",
		"ð", "d",
	},
	"Czeck": {
		"ř", "r",
		"ě", "e",
		"ý", "y",
		"á", "a",
		"í", "i",
		"é", "e",
		"ó", "o",
		"ú", "u",
		"ů", "u",
		"ď", "d",
		"ť", "t",
		"ň", "n",
	},
	"French": {
		"à", "a",
		"â", "a",
		"é", "e",
		"è", "e",
		"ê", "e",
		"ë", "e",
		"ù", "u",
		"ü", "u",
		"ÿ", "y",
		"ç", "c",
	},
	"Armenian": {
		"աու", "au",
		"բու", "bu",
		"գու", "gu",
		"դու", "du",
		"եու", "eu",
		"զու", "zu",
		"էու", "eu",
		"ըու", "yu",
		"թու", "tu",
		"ժու", "ju",
		"իու", "iu",
		"լու", "lu",
		"խու", "xu",
		"ծու", "cu",
		"կու", "ku",
		"հու", "hu",
		"ձու", "dzu",
		"ղու", "xu",
		"ճու", "cu",
		"մու", "mu",
		"յու", "yu",
		"նու", "nu",
		"շու", "shu",
		"չու", "chu",
		"պու", "pu",
		"ջու", "ju",
		"ռու", "ru",
		"սու", "su",
		"վու", "vu",
		"տու", "tu",
		"րու", "ru",
		"ցու", "cu",
		"փու", "pu",
		"քու", "qu",
		"օու", "ou",
		"ևու", "eu",
		"ֆու", "fu",
		"ոու", "vou",
		"ու", "u",
		"բո", "bo",
		"գո", "go",
		"դո", "do",
		"զո", "zo",
		"թո", "to",
		"ժո", "jo",
		"լո", "lo",
		"խո", "xo",
		"ծո", "co",
		"կո", "ko",
		"հո", "ho",
		"ձո", "dzo",
		"ղո", "xo",
		"ճո", "co",
		"մո", "mo",
		"յո", "yo",
		"նո", "no",
		"շո", "so",
		"չո", "co",
		"պո", "po",
		"ջո", "jo",
		"ռո", "ro",
		"սո", "so",
		"վո", "vo",
		"տո", "to",
		"րո", "ro",
		"ցո", "co",
		"փո", "po",
		"քո", "qo",
		"ևո", "eo",
		"ֆո", "fo",
		"ո", "vo",
		"եւ", "ev",
		"եվ", "ev",
		"ա", "a",
		"բ", "b",
		"գ", "g",
		"դ", "d",
		"ե", "e",
		"զ", "z",
		"է", "e",
		"ը", "y",
		"թ", "t",
		"ժ", "j",
		"ի", "i",
		"լ", "l",
		"խ", "x",
		"ծ", "c",
		"կ", "k",
		"հ", "h",
		"ձ", "dz",
		"ղ", "x",
		"ճ", "c",
		"մ", "m",
		"յ", "y",
		"ն", "n",
		"շ", "sh",
		"չ", "ch",
		"պ", "p",
		"ջ", "j",
		"ռ", "r",
		"ս", "s",
		"վ", "v",
		"տ", "t",
		"ր", "r",
		"ց", "c",
		"փ", "p",
		"ք", "q",
		"օ", "o",
		"և", "ev",
		"ֆ", "f",
		"ւ", "",
	},
	"Emoji": {
		"😂", ":')",
		"😊", ":)",
		"😃", ":)",
		"😩", "-_-",
		"😏", ":‑J",
		"💜", "<3",
		"💖", "<3",
		"💗", "<3",
		"❤️", "<3",
		"💕", "<3",
		"💞", "<3",
		"💘", "<3",
		"💓", "<3",
		"💚", "<3",
		"💙", "<3",
		"💔", "</3",
		"😱", "D:",
		"😮", ":O",
		"😝", ":P",
		"😍", ":x",
		"😢", ":(",
		"💯", ":100:",
		"🔥", ":fire:",
		"😉", ";)",
		"😴", ":zzz:",
		"💤", ":zzz:",
	},
}

func NewReplacer(useMaps ...string) *strings.Replacer {
	var replace []string
	if customMap, ok := Maps["custom"]; ok {
		replace = append(replace, customMap...)
	}
	for _, useMap := range useMaps {
		if useMap == "custom" {
			continue
		}
		translitMap, ok := Maps[useMap]
		if !ok {
			continue
		}
		replace = append(replace, translitMap...)
	}
	return strings.NewReplacer(replace...)
}
