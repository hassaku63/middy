package music

import (
	"fmt"
	"strings"
)

var romanDegrees = map[string]int{"i": 1, "ii": 2, "iii": 3, "iv": 4, "v": 5, "vi": 6, "vii": 7}

// ParseRomanNumeral reads a roman-numeral chord label as users and LLMs
// write it ("I", "ii", "V7", "vii°", "IVmaj7", "iiø7") into a scale degree
// and whether a seventh is requested.
//
// Only the degree and the seventh matter: the chord quality is always
// derived from the key by Key.Diatonic, so case and decorations such as
// "°" or "+" are accepted but do not change the result. Anything that is
// not a diatonic label is an error.
func ParseRomanNumeral(s string) (degree int, seventh bool, err error) {
	label := strings.TrimSpace(s)
	lower := strings.ToLower(label)

	// Longest suffixes first so that "maj7" is not read as "7".
	for _, suf := range []string{"maj7", "ø7", "m7b5", "m7", "7", "°", "dim", "+", "aug"} {
		if strings.HasSuffix(lower, suf) {
			lower = strings.TrimSuffix(lower, suf)
			seventh = strings.Contains(suf, "7")
			break
		}
	}
	degree, ok := romanDegrees[lower]
	if !ok {
		return 0, false, fmt.Errorf("music: %q is not a roman numeral (expected I..VII, optionally with 7)", s)
	}
	return degree, seventh, nil
}
