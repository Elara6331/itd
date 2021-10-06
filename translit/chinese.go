package translit

import (
	"bytes"
	"strings"
	"unicode"

	"github.com/mozillazg/go-pinyin"
)

// ChineseTranslit implements Transliterator using a pinyin
// conversion library.
type ChineseTranslit struct{}

func (ChineseTranslit) Init() {}

func (ct *ChineseTranslit) Transliterate(s string) string {
	// Create buffer for final output
	outBuf := &bytes.Buffer{}
	// Create buffer to temporarily store chinese characters
	tmpBuf := &bytes.Buffer{}
	// For every character in string
	for _, char := range s {
		// If character in Han range
		if unicode.Is(unicode.Han, char) {
			// Write character to temporary buffer
			tmpBuf.WriteRune(char)
		} else {
			// If buffer contains characters
			if tmpBuf.Len() > 0 {
				// Convert to pinyin (without tones)
				out := pinyin.LazyConvert(tmpBuf.String(), nil)
				// Write space-separated string to output
				outBuf.WriteString(strings.Join(out, " "))
				// Reset temporary buffer
				tmpBuf.Reset()
			}
			// Write character to output
			outBuf.WriteRune(char)
		}
	}
	// Return output string
	return outBuf.String()
}
