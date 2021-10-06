package translit

import (
	"strings"
)

type ArmenianTranslit struct {
	initComplete bool
}

var armenianMap = []string{
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
}

func (at *ArmenianTranslit) Init() {
	if !at.initComplete {
		// Copy map as original will be changed
		lower := armenianMap
		// For every value in copied map
		for i, val := range lower {
			// If index is odd, skip
			if i%2 == 1 {
				continue
			}
			// Capitalize first letter
			capital := strings.Title(val)
			// If capital is not the same as lowercase
			if capital != val {
				// Add capital to map
				armenianMap = append(armenianMap, capital, strings.Title(armenianMap[i+1]))
			}
		}
		// Set init complete to true so it is not run again
		at.initComplete = true
	}
}

func (at *ArmenianTranslit) Transliterate(s string) string {
	return strings.NewReplacer(armenianMap...).Replace(s)
}
