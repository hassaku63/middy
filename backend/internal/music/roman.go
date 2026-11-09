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
// result. A leading accidental (b, ♭, #, ♯) is read past, so "bVII" gives
// degree 7; StepFromLabel is what reads it as borrowed. Case is not read
// here: StepFromLabel and CanonicalLabel give it meaning (an uppercase V /
// V7 in a minor key is the harmonic-minor dominant). Any other label that
// is not a roman numeral is an error.
func ParseRomanNumeral(s string) (degree int, seventh bool, err error) {
	degree, seventh, _, err = parseRomanNumeral(s)
	return degree, seventh, err
}

// StepFromLabel builds a progression step for a label in a key. In minor
// keys an uppercase "V" or "V7" means the major dominant of harmonic minor
// (the raised leading tone), which is what learners and textbooks mean by
// "V in minor"; lowercase "v" keeps the natural-minor triad.
func StepFromLabel(key Key, label string, beats int) (Step, error) {
	head, bass, err := splitBass(label)
	if err != nil {
		return Step{}, err
	}
	if bass != nil {
		step, err := StepFromLabel(key, head, beats)
		if err != nil {
			return Step{}, err
		}
		if !onRoot(key, step, *bass) { // "I/1" is root position: no slash
			step.Bass = bass
		}
		return step, nil
	}
	return chordStep(key, label, beats)
}

// onRoot reports whether the bass is the chord's own root, which is root
// position, not a slash chord.
func onRoot(key Key, step Step, bass Bass) bool {
	chord, _, _, err := chordOfStep(key, step)
	return err == nil && chord.Root == bass.pitchClass(key)
}

// chordStep reads a label without a slash bass.
func chordStep(key Key, label string, beats int) (Step, error) {
	if step, ok, err := appliedStep(key, label, beats); ok {
		return step, err
	}
	degree, seventh, upper, suffix, accidental, err := parseLabel(label)
	if err != nil {
		return Step{}, err
	}
	step := Step{Degree: degree, Seventh: seventh, Beats: beats}
	if key.Mode == Minor && degree == 5 && upper && accidental == 0 {
		q := MajorTriad
		if seventh {
			q = Dominant7
		}
		step.Quality = &q
		return step, nil
	}
	if accidental == 0 {
		if dia, err := key.Diatonic(degree, seventh); err == nil && dia.Quality.isUpper() == upper && suffixAllows(suffix, seventh, dia.Quality) {
			return step, nil // plainly diatonic
		}
	}
	return outsideStep(key, label, step, seventh, upper, suffix, accidental)
}

// outsideStep handles a label that is not plainly the key's own chord: a
// chord borrowed from the parallel minor in a major key (BL-094), or a
// flat that names one of the key's own chords by its distance from the
// tonic, as lead sheets do (mixolydian's bVII is its VII, minor's bVI its
// VI; #109). The latter is the same sound, so it is accepted and
// respelled, not changed. Anything else without an accidental is left for
// CanonicalLabel to report; it needs the diatonic step to say what would
// have sounded.
func outsideStep(key Key, label string, step Step, seventh, upper bool, suffix string, accidental int) (Step, error) {
	q, ok := ChordQuality(0), false
	switch key.Mode {
	case Major:
		q, ok = borrowedQuality(key, step.Degree, seventh, suffix, accidental)
	case Minor:
		q, ok = minorBorrowedQuality(step.Degree, seventh, upper, suffix, accidental)
	}
	if ok {
		step.Quality = &q
		step.Accidental = accidental
		step.Borrowed = true
		return step, nil
	}
	if accidental == 0 {
		return step, nil
	}
	if same, ok := literalDiatonic(key, label, step.Beats); ok {
		return same, nil
	}
	return Step{}, fmt.Errorf("music: %q is not available in %s", strings.TrimSpace(label), key)
}

// literalDiatonic finds the key's own chord that a label names when read
// literally against the tonic (LiteralChord).
func literalDiatonic(key Key, label string, beats int) (Step, bool) {
	c, err := LiteralChord(key.Tonic, label)
	if err != nil {
		return Step{}, false
	}
	seventh := len(c.Quality.Intervals()) == 4
	for d := 1; d <= 7; d++ {
		if dia, err := key.Diatonic(d, seventh); err == nil && dia == c {
			return Step{Degree: d, Seventh: seventh, Beats: beats}, true
		}
	}
	return Step{}, false
}

// borrowedQuality resolves a label against the parallel minor: the root
// must be that minor key's degree (a lowered root for ♭VI and ♭VII, the
// same root for iv and ii°) and the written quality must be one the
// borrowable set allows.
func borrowedQuality(key Key, degree int, seventh bool, suffix string, accidental int) (ChordQuality, bool) {
	parallel := Key{Tonic: key.Tonic, Mode: Minor}
	root := key.Scale()[degree-1].Add(accidental)
	if parallel.Scale()[degree-1] != root {
		return 0, false
	}
	if suffix == "m6" {
		return MinorSixth, borrowable(degree, MinorSixth)
	}
	c, err := parallel.Diatonic(degree, seventh)
	if err != nil || !borrowable(degree, c.Quality) || !suffixAllows(suffix, seventh, c.Quality) {
		return 0, false
	}
	return c.Quality, true
}

// minorBorrowedQuality is the minor key's side of borrowing (BL-075, ADR-0012
// decision 3): the major I and IV of the parallel major (the Picardy third,
// and the raised sixth of melodic minor or dorian), and the Neapolitan bII.
// All are major triads; the label is written against the minor key's own
// scale, with a flat only for the Neapolitan root, which is not in it.
func minorBorrowedQuality(degree int, seventh, upper bool, suffix string, accidental int) (ChordQuality, bool) {
	if seventh || !upper || suffix != "" {
		return 0, false
	}
	switch {
	case (degree == 1 || degree == 4) && accidental == 0:
		return MajorTriad, true
	case degree == 2 && accidental == -1:
		return MajorTriad, true
	}
	return 0, false
}

func parseRomanNumeral(s string) (degree int, seventh bool, upper bool, err error) {
	label := strings.TrimSpace(s)
	for _, acc := range []string{"♭", "b", "♯", "#"} {
		if len(label) > 1 && strings.HasPrefix(label, acc) {
			label = strings.TrimPrefix(label, acc)
			break
		}
	}
	lower := strings.ToLower(label)

	// Longest suffixes first so that "maj7" is not read as "7".
	for _, suf := range []string{"maj7", "ø7", "m7b5", "m7", "m6", "7", "°", "dim", "+", "aug"} {
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

// CanonicalLabel checks a label against the key and returns the engine's own
// spelling of the chord it names. The written case and seventh type carry
// meaning (an uppercase IV in minor asks for the dorian major chord; I7
// asks for a dominant seventh), so when they disagree with the chord the
// key gives that degree, the label is rejected with what it would have
// sounded like, instead of being changed in silence (ADR-0013).
func CanonicalLabel(key Key, label string) (string, error) {
	head, bass, err := splitBass(label)
	if err != nil {
		return "", err
	}
	if bass != nil {
		canonical, err := CanonicalLabel(key, head)
		if err != nil {
			return "", err
		}
		if step, err := StepFromLabel(key, head, 1); err == nil && onRoot(key, step, *bass) {
			return canonical, nil
		}
		return canonical + bass.suffix(), nil
	}
	if step, ok, err := appliedStep(key, label, 1); ok {
		if err != nil {
			return "", err
		}
		return key.appliedNumeral(step.Degree, step.Seventh, step.Substitute)
	}
	_, seventh, upper, suffix, _, err := parseLabel(label)
	if err != nil {
		return "", err
	}
	step, err := StepFromLabel(key, label, 1)
	if err != nil {
		return "", err
	}
	quality, err := qualityOf(key, step)
	if err != nil {
		return "", err
	}
	canonical, err := RomanNumeralAlt(step.Degree, step.Accidental, quality)
	if err != nil {
		return "", err
	}
	if upper != quality.isUpper() || !suffixAllows(suffix, seventh, quality) {
		symbol := key.spellDegree(step.Degree, step.Accidental) + quality.Symbol()
		return "", fmt.Errorf("music: %q would sound as %s (%s) in %s; write %q", strings.TrimSpace(label), symbol, canonical, key, canonical)
	}
	return canonical, nil
}

// qualityOf resolves the chord quality a step has in a key.
func qualityOf(key Key, step Step) (ChordQuality, error) {
	if step.Quality != nil {
		return *step.Quality, nil
	}
	c, err := key.Diatonic(step.Degree, step.Seventh)
	if err != nil {
		return 0, err
	}
	return c.Quality, nil
}

// isUpper reports whether roman-numeral convention writes the quality in
// uppercase (major-type chords) rather than lowercase (minor, diminished).
func (q ChordQuality) isUpper() bool {
	switch q {
	case MajorTriad, Major7, Dominant7, AugmentedTriad:
		return true
	}
	return false
}

// suffixAllows reports whether an explicit suffix agrees with the quality.
// A bare numeral agrees with any triad of its case ("vii" for vii°); a
// bare "7" means a dominant seventh in uppercase and a minor seventh in
// lowercase, so it also accepts the half-diminished chord that lowercase
// convention writes ø7.
func suffixAllows(suffix string, seventh bool, q ChordQuality) bool {
	switch suffix {
	case "":
		return !seventh && q != Major7 && q != Dominant7 && q != Minor7 && q != HalfDiminished7
	case "maj7":
		return q == Major7
	case "7":
		return q == Dominant7 || q == Minor7 || q == HalfDiminished7
	case "m7":
		return q == Minor7
	case "ø7", "m7b5":
		return q == HalfDiminished7
	case "m6":
		return q == MinorSixth
	case "°", "dim":
		return q == DiminishedTriad
	case "+", "aug":
		return q == AugmentedTriad
	}
	return false
}

// parseLabel is parseRomanNumeral plus the suffix and the leading
// accidental it consumed (-1 for ♭ or b; sharps are parsed but not
// available anywhere yet, so they fail later as "not available").
func parseLabel(s string) (degree int, seventh bool, upper bool, suffix string, accidental int, err error) {
	trimmed := strings.TrimSpace(s)
	switch {
	case strings.HasPrefix(trimmed, "♭") || (strings.HasPrefix(trimmed, "b") && len(trimmed) > 1):
		accidental = -1
	case strings.HasPrefix(trimmed, "♯") || strings.HasPrefix(trimmed, "#"):
		accidental = 1
	}
	lower := strings.ToLower(trimmed)
	for _, suf := range []string{"maj7", "ø7", "m7b5", "m7", "m6", "7", "°", "dim", "+", "aug"} {
		if strings.HasSuffix(lower, suf) {
			suffix = suf
			break
		}
	}
	degree, seventh, upper, err = parseRomanNumeral(s)
	return degree, seventh, upper, suffix, accidental, err
}

// LiteralChord reads a label as written, against the major scale of tonic,
// whatever key the music is in: the case and suffix give the quality and a
// leading ♭ or b lowers the root ("bVII" is the major chord a whole step
// below the tonic, "iv" the minor chord on the fourth, "I7" the dominant
// seventh on the tonic). It is the reference lead sheets and borrowed-chord
// tables use, and lets an evaluation ask for a sound rather than for a
// spelling that only some keys accept.
func LiteralChord(tonic PitchClass, label string) (Chord, error) {
	degree, _, upper, suffix, accidental, err := parseLabel(label)
	if err != nil {
		return Chord{}, err
	}
	if accidental == 1 {
		return Chord{}, fmt.Errorf("music: %q: sharps are not supported", strings.TrimSpace(label))
	}
	major := Key{Tonic: tonic, Mode: Major}
	return Chord{Root: major.Scale()[degree-1].Add(accidental), Quality: literalQuality(upper, suffix)}, nil
}

// literalQuality is the quality a numeral's case and suffix name.
func literalQuality(upper bool, suffix string) ChordQuality {
	switch suffix {
	case "maj7":
		return Major7
	case "ø7", "m7b5":
		return HalfDiminished7
	case "m7":
		return Minor7
	case "m6":
		return MinorSixth
	case "7":
		if upper {
			return Dominant7
		}
		return Minor7
	case "°", "dim":
		return DiminishedTriad
	case "+", "aug":
		return AugmentedTriad
	}
	if upper {
		return MajorTriad
	}
	return MinorTriad
}
