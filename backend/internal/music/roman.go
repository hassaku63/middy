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
// The chord quality is derived from the key by Key.Diatonic, so
// decorations such as "°" or "+" are accepted but do not change the
// result. Case matters in exactly one place: StepFromLabel treats an
// uppercase V / V7 in a minor key as the harmonic-minor dominant. Anything
// that is not a diatonic label is an error.
func ParseRomanNumeral(s string) (degree int, seventh bool, err error) {
	degree, seventh, _, err = parseRomanNumeral(s)
	return degree, seventh, err
}

// StepFromLabel builds a progression step for a label in a key. In minor
// keys an uppercase "V" or "V7" means the major dominant of harmonic minor
// (the raised leading tone), which is what learners and textbooks mean by
// "V in minor"; lowercase "v" keeps the natural-minor triad.
func StepFromLabel(key Key, label string, beats int) (Step, error) {
	degree, seventh, upper, err := parseRomanNumeral(label)
	if err != nil {
		return Step{}, err
	}
	step := Step{Degree: degree, Seventh: seventh, Beats: beats}
	if key.Mode == Minor && degree == 5 && upper {
		q := MajorTriad
		if seventh {
			q = Dominant7
		}
		step.Quality = &q
	}
	return step, nil
}

func parseRomanNumeral(s string) (degree int, seventh bool, upper bool, err error) {
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
		return 0, false, false, fmt.Errorf("music: %q is not a roman numeral (expected I..VII, optionally with 7)", s)
	}
	upper = strings.HasPrefix(label, strings.ToUpper(lower[:1]))
	return degree, seventh, upper, nil
}
