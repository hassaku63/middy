package music

import (
	"fmt"
	"slices"
)

// ChordQuality is the interval structure of a chord above its root.
type ChordQuality int

// Supported chord qualities: four triads, five sevenths and the minor sixth.
// The triads and the first four sevenths cover every diatonic chord in
// major and natural minor keys; the rest serve borrowed and applied chords.
const (
	MajorTriad ChordQuality = iota
	MinorTriad
	DiminishedTriad
	AugmentedTriad
	Major7
	Minor7
	Dominant7
	HalfDiminished7
	MinorSixth  // minor triad with a major sixth; the borrowed ivm6 (BL-094)
	Diminished7 // stacked minor thirds; the passing vii°7/x (BL-130)
	Sus4        // the fourth in place of the third (BL-132)
	Dominant7Sus4
	Power // root and fifth only, no third (BL-132)
)

type qualityInfo struct {
	name      string // stable identifier used in JSON
	symbol    string // suffix in chord symbols, e.g. "m7"
	intervals []int
	major     bool // uppercase roman numeral when true
}

var qualities = map[ChordQuality]qualityInfo{
	MajorTriad:      {"maj", "", []int{0, 4, 7}, true},
	MinorTriad:      {"min", "m", []int{0, 3, 7}, false},
	DiminishedTriad: {"dim", "dim", []int{0, 3, 6}, false},
	AugmentedTriad:  {"aug", "aug", []int{0, 4, 8}, true},
	Major7:          {"maj7", "maj7", []int{0, 4, 7, 11}, true},
	Minor7:          {"min7", "m7", []int{0, 3, 7, 10}, false},
	Dominant7:       {"dom7", "7", []int{0, 4, 7, 10}, true},
	HalfDiminished7: {"m7b5", "m7b5", []int{0, 3, 6, 10}, false},
	MinorSixth:      {"min6", "m6", []int{0, 3, 7, 9}, false},
	Diminished7:     {"dim7", "dim7", []int{0, 3, 6, 9}, false},
	Sus4:            {"sus4", "sus4", []int{0, 5, 7}, true},
	Dominant7Sus4:   {"dom7sus4", "7sus4", []int{0, 5, 7, 10}, true},
	Power:           {"power", "5", []int{0, 7}, true},
}

// Qualities lists every chord quality the engine knows, in declaration
// order, so mirrors of the table (the frontend fixtures) cannot miss one.
func Qualities() []ChordQuality {
	out := make([]ChordQuality, 0, len(qualities))
	for q := range qualities {
		out = append(out, q)
	}
	slices.Sort(out)
	return out
}

func (q ChordQuality) info() (qualityInfo, error) {
	i, ok := qualities[q]
	if !ok {
		return qualityInfo{}, fmt.Errorf("music: unknown chord quality %d", int(q))
	}
	return i, nil
}

// String returns the stable identifier ("maj", "min7", ...).
func (q ChordQuality) String() string {
	i, err := q.info()
	if err != nil {
		return fmt.Sprintf("ChordQuality(%d)", int(q))
	}
	return i.name
}

// Symbol returns the chord-symbol suffix ("", "m", "7", ...).
func (q ChordQuality) Symbol() string {
	i, _ := q.info()
	return i.symbol
}

// Intervals returns semitone offsets from the root, root first.
func (q ChordQuality) Intervals() []int {
	i, _ := q.info()
	out := make([]int, len(i.intervals))
	copy(out, i.intervals)
	return out
}

// MarshalText implements encoding.TextMarshaler.
func (q ChordQuality) MarshalText() ([]byte, error) {
	i, err := q.info()
	if err != nil {
		return nil, err
	}
	return []byte(i.name), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (q *ChordQuality) UnmarshalText(text []byte) error {
	for quality, i := range qualities {
		if i.name == string(text) {
			*q = quality
			return nil
		}
	}
	return fmt.Errorf("music: unknown chord quality %q", string(text))
}

// Chord is a root pitch class with a quality. Inversions and slash chords
// are not modelled yet (backlog).
type Chord struct {
	Root    PitchClass   `json:"root"`
	Quality ChordQuality `json:"quality"`
}

// String renders a chord symbol such as "Am7" or "G7".
func (c Chord) String() string {
	return c.Root.String() + c.Quality.Symbol()
}

// PitchClasses returns the chord tones, root first.
func (c Chord) PitchClasses() []PitchClass {
	ivs := c.Quality.Intervals()
	out := make([]PitchClass, len(ivs))
	for i, iv := range ivs {
		out[i] = c.Root.Add(iv)
	}
	return out
}

// Voicing returns a close root-position voicing with the root in the given
// octave (C4 = octave 4).
func (c Chord) Voicing(octave int) []Pitch {
	root := NewPitch(c.Root, octave)
	ivs := c.Quality.Intervals()
	out := make([]Pitch, len(ivs))
	for i, iv := range ivs {
		out[i] = root.Transpose(iv)
	}
	return out
}

// triadQualities maps (third, fifth) semitone intervals to a quality.
var triadQualities = map[[2]int]ChordQuality{
	{4, 7}: MajorTriad,
	{3, 7}: MinorTriad,
	{3, 6}: DiminishedTriad,
	{4, 8}: AugmentedTriad,
}

// seventhQualities maps (triad quality, seventh interval) to a seventh chord.
var seventhQualities = map[[2]int]ChordQuality{
	{int(MajorTriad), 11}:      Major7,
	{int(MajorTriad), 10}:      Dominant7,
	{int(MinorTriad), 10}:      Minor7,
	{int(DiminishedTriad), 10}: HalfDiminished7,
}

// Diatonic builds the chord stacked in thirds on a 1-based scale degree of
// the key, as a triad or a seventh chord. The quality is derived from the
// scale itself, so it is correct for any mode added to Mode later.
func (k Key) Diatonic(degree int, seventh bool) (Chord, error) {
	root, err := k.Degree(degree)
	if err != nil {
		return Chord{}, err
	}
	scale := k.Scale()
	interval := func(offset int) int { // semitones from root to the scale step `offset` above it
		return int(scale[(degree-1+offset)%7].Add(-int(root)).Normalize())
	}

	q, ok := triadQualities[[2]int{interval(2), interval(4)}]
	if !ok {
		return Chord{}, fmt.Errorf("music: no triad quality for intervals %d,%d", interval(2), interval(4))
	}
	if !seventh {
		return Chord{Root: root, Quality: q}, nil
	}
	q7, ok := seventhQualities[[2]int{int(q), interval(6)}]
	if !ok {
		return Chord{}, fmt.Errorf("music: unsupported seventh chord on degree %d of %s", degree, k)
	}
	return Chord{Root: root, Quality: q7}, nil
}

var romanBase = [7]string{"I", "II", "III", "IV", "V", "VI", "VII"}

// RomanNumeral renders the conventional roman-numeral label for a chord
// quality on a 1-based degree: uppercase for major/augmented, lowercase for
// minor/diminished, with "°", "+", "7", "ø7" decorations.
func RomanNumeral(degree int, q ChordQuality) (string, error) {
	return RomanNumeralAlt(degree, 0, q)
}

// RomanNumeralAlt is RomanNumeral for a root altered by accidental
// semitones relative to the scale degree: -1 writes "♭" before the numeral
// (the borrowed ♭VI, ♭VII). Only -1 and 0 are supported.
func RomanNumeralAlt(degree, accidental int, q ChordQuality) (string, error) {
	if degree < 1 || degree > 7 {
		return "", fmt.Errorf("music: scale degree must be 1..7, got %d", degree)
	}
	if accidental != 0 && accidental != -1 {
		return "", fmt.Errorf("music: unsupported accidental %d", accidental)
	}
	i, err := q.info()
	if err != nil {
		return "", err
	}
	base := romanBase[degree-1]
	if !i.major {
		base = toLowerASCII(base)
	}
	if accidental == -1 {
		base = "♭" + base
	}
	return base + numeralSuffix[q], nil
}

// numeralSuffix is what a roman numeral adds for a quality; triads of
// either case add nothing.
var numeralSuffix = map[ChordQuality]string{
	DiminishedTriad: "°",
	AugmentedTriad:  "+",
	Major7:          "maj7",
	Minor7:          "7",
	Dominant7:       "7",
	HalfDiminished7: "ø7",
	MinorSixth:      "m6",
	Diminished7:     "°7",
	Sus4:            "sus4",
	Dominant7Sus4:   "7sus4",
	Power:           "5",
}

func toLowerASCII(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		}
	}
	return string(b)
}
