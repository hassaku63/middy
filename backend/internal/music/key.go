package music

import (
	"fmt"
	"slices"
	"strings"
)

// Mode is a scale type: which note of a diatonic scale is heard as the
// centre. Major (ionian) and Minor (aeolian) came first; Dorian and
// Mixolydian were added for Celtic, rock and modal jazz (ADR-0015).
// Phrygian and Lydian slot in with three lines: intervals, parent offset,
// characteristic degree.
type Mode int

// Supported modes.
const (
	Major      Mode = iota
	Minor           // natural minor (aeolian)
	Dorian          // minor with a major sixth
	Mixolydian      // major with a minor seventh
)

var modeNames = map[Mode]string{Major: "major", Minor: "minor", Dorian: "dorian", Mixolydian: "mixolydian"}

// Modes lists every supported mode in declaration order.
func Modes() []Mode {
	out := make([]Mode, 0, len(modeNames))
	for m := range modeNames {
		out = append(out, m)
	}
	slices.Sort(out)
	return out
}

var modeIntervals = map[Mode][7]int{
	Major:      {0, 2, 4, 5, 7, 9, 11},
	Minor:      {0, 2, 3, 5, 7, 8, 10},
	Dorian:     {0, 2, 3, 5, 7, 9, 10},
	Mixolydian: {0, 2, 4, 5, 7, 9, 10},
}

// characteristicDegree is the scale degree that tells the mode apart from
// the major or minor key it resembles: dorian's major sixth, mixolydian's
// minor seventh. Zero for major and minor.
var characteristicDegree = map[Mode]int{Dorian: 6, Mixolydian: 7}

// Characteristic returns the mode's characteristic scale degree, or 0.
func (m Mode) Characteristic() int {
	return characteristicDegree[m]
}

// String returns the lowercase mode name used in JSON and display.
func (m Mode) String() string {
	if n, ok := modeNames[m]; ok {
		return n
	}
	return fmt.Sprintf("Mode(%d)", int(m))
}

// MarshalText implements encoding.TextMarshaler.
func (m Mode) MarshalText() ([]byte, error) {
	n, ok := modeNames[m]
	if !ok {
		return nil, fmt.Errorf("music: unknown mode %d", int(m))
	}
	return []byte(n), nil
}

// UnmarshalText implements encoding.TextUnmarshaler. Case is ignored
// because LLM structured output does not guarantee the case of enum
// values; MarshalText always writes the lowercase name.
func (m *Mode) UnmarshalText(text []byte) error {
	for mode, name := range modeNames {
		if strings.EqualFold(name, string(text)) {
			*m = mode
			return nil
		}
	}
	names := make([]string, 0, len(modeNames))
	for _, m := range Modes() {
		names = append(names, modeNames[m])
	}
	// The list lets a caller (or the model's retry) fix the value at once.
	return fmt.Errorf("music: unknown mode %q; use one of %v", string(text), names)
}

// Intervals returns the semitone offsets of the seven scale degrees.
func (m Mode) Intervals() [7]int {
	return modeIntervals[m]
}

// Key is a tonic plus a mode, e.g. C major or A minor.
type Key struct {
	Tonic PitchClass `json:"tonic"`
	Mode  Mode       `json:"mode"`
}

// String renders the key as notation spells it: "Bb major", "C# minor".
func (k Key) String() string {
	return fmt.Sprintf("%s %s", k.TonicName(), k.Mode)
}

// Scale returns the seven pitch classes of the key, tonic first.
func (k Key) Scale() [7]PitchClass {
	var out [7]PitchClass
	for i, iv := range k.Mode.Intervals() {
		out[i] = k.Tonic.Add(iv)
	}
	return out
}

// Degree returns the pitch class of a 1-based scale degree.
func (k Key) Degree(degree int) (PitchClass, error) {
	if degree < 1 || degree > 7 {
		return 0, fmt.Errorf("music: scale degree must be 1..7, got %d", degree)
	}
	return k.Scale()[degree-1], nil
}

// Contains reports whether the pitch class is diatonic to the key.
func (k Key) Contains(pc PitchClass) bool {
	pc = pc.Normalize()
	for _, s := range k.Scale() {
		if s == pc {
			return true
		}
	}
	return false
}
