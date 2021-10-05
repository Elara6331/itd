package translit

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// https://en.wikipedia.org/wiki/Hangul_Jamo_%28Unicode_block%29
var jamoBlock = &unicode.RangeTable{
	R16: []unicode.Range16{{
		Lo:     0x1100,
		Hi:     0x11FF,
		Stride: 1,
	}},
}

// https://en.wikipedia.org/wiki/Hangul_Syllables
var syllablesBlock = &unicode.RangeTable{
	R16: []unicode.Range16{{
		Lo:     0xAC00,
		Hi:     0xD7A3,
		Stride: 1,
	}},
}

// https://en.wikipedia.org/wiki/Hangul_Compatibility_Jamo
var compatJamoBlock = &unicode.RangeTable{
	R16: []unicode.Range16{{
		Lo:     0x3131,
		Hi:     0x318E,
		Stride: 1,
	}},
}

// KoreanTranslit implements transliteration for Korean.
//
// This was translated to Go from the code in https://codeberg.org/Freeyourgadget/Gadgetbridge
type KoreanTranslit struct{}

// User input consisting of isolated jamo is usually mapped to the KS X 1001 compatibility
// block, but jamo resulting from decomposed syllables are mapped to the modern one. This
// function maps compat jamo to modern ones where possible and returns all other characters
// unmodified.
//
// https://en.wikipedia.org/wiki/Hangul_Compatibility_Jamo
// https://en.wikipedia.org/wiki/Hangul_Jamo_%28Unicode_block%29
func decompatJamo(jamo rune) rune {
	// KS X 1001 Hangul filler, not used in modern Unicode. A useful landmark in the
	// compatibility jamo block.
	// https://en.wikipedia.org/wiki/KS_X_1001#Hangul_Filler
	var hangulFiller rune = 0x3164

	// Ignore characters outside compatibility jamo block
	if !unicode.In(jamo, compatJamoBlock) {
		return jamo
	}

	// Vowels are contiguous, in the same order, and unambiguous so it's a simple offset.
	if jamo >= 0x314F && jamo < hangulFiller {
		return jamo - 0x1FEE
	}

	// Consonants are organized differently. No clean way to do this.
	// The compatibility jamo block doesn't distinguish between Choseong (leading) and Jongseong
	// (final) positions, but the modern block does. We map to Choseong here.
	switch jamo {
	case 0x3131:
		return 0x1100 // ㄱ
	case 0x3132:
		return 0x1101 // ㄲ
	case 0x3134:
		return 0x1102 // ㄴ
	case 0x3137:
		return 0x1103 // ㄷ
	case 0x3138:
		return 0x1104 // ㄸ
	case 0x3139:
		return 0x1105 // ㄹ
	case 0x3141:
		return 0x1106 // ㅁ
	case 0x3142:
		return 0x1107 // ㅂ
	case 0x3143:
		return 0x1108 // ㅃ
	case 0x3145:
		return 0x1109 // ㅅ
	case 0x3146:
		return 0x110A // ㅆ
	case 0x3147:
		return 0x110B // ㅇ
	case 0x3148:
		return 0x110C // ㅈ
	case 0x3149:
		return 0x110D // ㅉ
	case 0x314A:
		return 0x110E // ㅊ
	case 0x314B:
		return 0x110F // ㅋ
	case 0x314C:
		return 0x1110 // ㅌ
	case 0x314D:
		return 0x1111 // ㅍ
	case 0x314E:
		return 0x1112 // ㅎ
	}

	// The rest of the compatibility block consists of archaic compounds that are
	// unlikely to be encountered in modern systems. Just leave them alone.
	return jamo
}

// Transliterates one jamo at a time.
// Does nothing if it isn't in the modern jamo block.
func translitSingleJamo(jamo rune) string {
	jamo = decompatJamo(jamo)

	switch jamo {
	// Choseong (leading position consonants)
	case 0x1100:
		return "g" // ㄱ
	case 0x1101:
		return "kk" // ㄲ
	case 0x1102:
		return "n" // ㄴ
	case 0x1103:
		return "d" // ㄷ
	case 0x1104:
		return "tt" // ㄸ
	case 0x1105:
		return "r" // ㄹ
	case 0x1106:
		return "m" // ㅁ
	case 0x1107:
		return "b" // ㅂ
	case 0x1108:
		return "pp" // ㅃ
	case 0x1109:
		return "s" // ㅅ
	case 0x110A:
		return "ss" // ㅆ
	case 0x110B:
		return "" // ㅇ
	case 0x110C:
		return "j" // ㅈ
	case 0x110D:
		return "jj" // ㅉ
	case 0x110E:
		return "ch" // ㅊ
	case 0x110F:
		return "k" // ㅋ
	case 0x1110:
		return "t" // ㅌ
	case 0x1111:
		return "p" // ㅍ
	case 0x1112:
		return "h" // ㅎ
	// Jungseong (vowels)
	case 0x1161:
		return "a" // ㅏ
	case 0x1162:
		return "ae" // ㅐ
	case 0x1163:
		return "ya" // ㅑ
	case 0x1164:
		return "yae" // ㅒ
	case 0x1165:
		return "eo" // ㅓ
	case 0x1166:
		return "e" // ㅔ
	case 0x1167:
		return "yeo" // ㅕ
	case 0x1168:
		return "ye" // ㅖ
	case 0x1169:
		return "o" // ㅗ
	case 0x116A:
		return "wa" // ㅘ
	case 0x116B:
		return "wae" // ㅙ
	case 0x116C:
		return "oe" // ㅚ
	case 0x116D:
		return "yo" // ㅛ
	case 0x116E:
		return "u" // ㅜ
	case 0x116F:
		return "wo" // ㅝ
	case 0x1170:
		return "we" // ㅞ
	case 0x1171:
		return "wi" // ㅟ
	case 0x1172:
		return "yu" // ㅠ
	case 0x1173:
		return "eu" // ㅡ
	case 0x1174:
		return "ui" // ㅢ
	case 0x1175:
		return "i" // ㅣ
	// Jongseong (final position consonants)
	case 0x11A8:
		return "k" // ㄱ
	case 0x11A9:
		return "k" // ㄲ
	case 0x11AB:
		return "n" // ㄴ
	case 0x11AE:
		return "t" // ㄷ
	case 0x11AF:
		return "l" // ㄹ
	case 0x11B7:
		return "m" // ㅁ
	case 0x11B8:
		return "p" // ㅂ
	case 0x11BA:
		return "t" // ㅅ
	case 0x11BB:
		return "t" // ㅆ
	case 0x11BC:
		return "ng" // ㅇ
	case 0x11BD:
		return "t" // ㅈ
	case 0x11BE:
		return "t" // ㅊ
	case 0x11BF:
		return "k" // ㅋ
	case 0x11C0:
		return "t" // ㅌ
	case 0x11C1:
		return "p" // ㅍ
	case 0x11C2:
		return "t" // ㅎ
	}

	return string(jamo)
}

// Some combinations of ending jamo in one syllable and initial jamo in the next are romanized
// irregularly. These exceptions are called "special provisions". In cases where multiple
// romanizations are permitted, we use the one that's least commonly used elsewhere.
//
// Returns empty strring and false if either character is not in the modern jamo block,
// or if there is no special provision for that pair of jamo.
func translitSpecialProvisions(previousEnding rune, nextInitial rune) (string, bool) {
	// Return false if previousEnding not in modern jamo block
	if !unicode.In(previousEnding, jamoBlock) {
		return "", false
	}
	// Return false if nextInitial not in modern jamo block
	if !unicode.In(nextInitial, jamoBlock) {
		return "", false
	}

	// Jongseong (final position) ㅎ has a number of special provisions.
	if previousEnding == 0x11C2 {
		switch nextInitial {
		case 0x110B:
			return "h", true // ㅇ
		case 0x1100:
			return "k", true // ㄱ
		case 0x1102:
			return "nn", true // ㄴ
		case 0x1103:
			return "t", true // ㄷ
		case 0x1105:
			return "nn", true // ㄹ
		case 0x1106:
			return "nm", true // ㅁ
		case 0x1107:
			return "p", true // ㅂ
		case 0x1109:
			return "hs", true // ㅅ
		case 0x110C:
			return "ch", true // ㅈ
		case 0x1112:
			return "t", true // ㅎ
		default:
			return "", false
		}
	}

	// Otherwise, special provisions are denser when grouped by the second jamo.
	switch nextInitial {
	case 0x1100: // ㄱ
		switch previousEnding {
		case 0x11AB:
			return "n-g", true // ㄴ
		default:
			return "", false
		}
	case 0x1102: // ㄴ
		switch previousEnding {
		case 0x11A8:
			return "ngn", true // ㄱ
		case 0x11AE:
			fallthrough // ㄷ
		case 0x11BA:
			fallthrough // ㅅ
		case 0x11BD:
			fallthrough // ㅈ
		case 0x11BE:
			fallthrough // ㅊ
		case 0x11C0: // ㅌ
			return "nn", true
		case 0x11AF:
			return "ll", true // ㄹ
		case 0x11B8:
			return "mn", true // ㅂ
		default:
			return "", false
		}
	case 0x1105: // ㄹ
		switch previousEnding {
		case 0x11A8:
			fallthrough // ㄱ
		case 0x11AB:
			fallthrough // ㄴ
		case 0x11AF: // ㄹ
			return "ll", true
		case 0x11AE:
			fallthrough // ㄷ
		case 0x11BA:
			fallthrough // ㅅ
		case 0x11BD:
			fallthrough // ㅈ
		case 0x11BE:
			fallthrough // ㅊ
		case 0x11C0: // ㅌ
			return "nn", true
		case 0x11B7:
			fallthrough // ㅁ
		case 0x11B8: // ㅂ
			return "mn", true
		case 0x11BC:
			return "ngn", true // ㅇ
		default:
			return "", false
		}
	case 0x1106: // ㅁ
		switch previousEnding {
		case 0x11A8:
			return "ngm", true // ㄱ
		case 0x11AE:
			fallthrough // ㄷ
		case 0x11BA:
			fallthrough // ㅅ
		case 0x11BD:
			fallthrough // ㅈ
		case 0x11BE:
			fallthrough // ㅊ
		case 0x11C0: // ㅌ
			return "nm", true
		case 0x11B8:
			return "mm", true // ㅂ
		default:
			return "", false
		}
	case 0x110B: // ㅇ
		switch previousEnding {
		case 0x11A8:
			return "g", true // ㄱ
		case 0x11AE:
			return "d", true // ㄷ
		case 0x11AF:
			return "r", true // ㄹ
		case 0x11B8:
			return "b", true // ㅂ
		case 0x11BA:
			return "s", true // ㅅ
		case 0x11BC:
			return "ng-", true // ㅇ
		case 0x11BD:
			return "j", true // ㅈ
		case 0x11BE:
			return "ch", true // ㅊ
		default:
			return "", false
		}
	case 0x110F: // ㅋ
		switch previousEnding {
		case 0x11A8:
			return "k-k", true // ㄱ
		default:
			return "", false
		}
	case 0x1110: // ㅌ
		switch previousEnding {
		case 0x11AE:
			fallthrough // ㄷ
		case 0x11BA:
			fallthrough // ㅅ
		case 0x11BD:
			fallthrough // ㅈ
		case 0x11BE:
			fallthrough // ㅊ
		case 0x11C0: // ㅌ
			return "t-t", true
		default:
			return "", false
		}
	case 0x1111: // ㅍ
		switch previousEnding {
		case 0x11B8:
			return "p-p", true // ㅂ
		default:
			return "", false
		}
	default:
		return "", false
	}
}

// Decompose a syllable into several jamo. Does nothing if that isn't possible.
func decompose(syllable rune) string {
	return norm.NFD.String(string(syllable))
}

// Transliterate any Hangul in the given string.
// Leaves any non-Hangul characters unmodified.
func (kt *KoreanTranslit) Transliterate(s string) string {
	if len(s) == 0 {
		return s
	}

	builder := &strings.Builder{}

	nextInitialJamoConsumed := false

	for i, syllable := range s {
		// If character not in blocks, leave it unmodified
		if !unicode.In(syllable, jamoBlock, syllablesBlock, compatJamoBlock) {
			builder.WriteRune(syllable)
			continue
		}

		jamo := decompose(syllable)
		for j, char := range jamo {
			// If we already transliterated the first jamo of this syllable as part of a special
			// provision, skip it. Otherwise, handle it in the unconditional else branch.
			if j == 0 && nextInitialJamoConsumed {
				nextInitialJamoConsumed = false
				continue
			}

			// If this is the last jamo of this syllable and not the last syllable of the
			// string, check for special provisions. If the next char is whitespace or not
			// Hangul, run translitSpecialProvisions() should return no value.
			if j == len(jamo)-1 && i < len(s)-1 {
				nextSyllable := s[i+1]
				nextJamo := decompose(rune(nextSyllable))[0]

				// Attempt to handle special provision
				specialProvision, ok := translitSpecialProvisions(char, rune(nextJamo))
				if ok {
					builder.WriteString(specialProvision)
					nextInitialJamoConsumed = true
				} else {
					// Not a special provision, transliterate normally
					builder.WriteString(translitSingleJamo(char))
				}
				continue
			}
			// Transliterate normally
			builder.WriteString(translitSingleJamo(char))
		}
	}
	return builder.String()
}
