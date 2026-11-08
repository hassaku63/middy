package music

import (
	"fmt"
	"strings"
)

// Secondary (applied) dominants (BL-074): V/x and V7/x are the major triad
// or dominant seventh a fifth above the chord on degree x, which they pull
// towards as if x were a temporary tonic (V7/V in C major is D7, going to
// G). The target must be a major or minor triad other than the tonic: a
// diminished chord is not approached by its own dominant, and V/I is V.
// The applied chord must also lie outside the key, otherwise it is one of
// the key's own chords under another name (V/IV in C major is C, the
// tonic; only V7/IV, C7, is applied).

// appliedChord returns the secondary dominant of target in k.
func (k Key) appliedChord(target int, seventh bool) (Chord, error) {
	t, err := k.Diatonic(target, false)
	if err != nil {
		return Chord{}, err
	}
	q := MajorTriad
	if seventh {
		q = Dominant7
	}
	return Chord{Root: t.Root.Add(7), Quality: q}, nil
}

// appliedNumeral writes "V/ii", "V7/IV": the target as the key's triad.
func (k Key) appliedNumeral(target int, seventh bool) (string, error) {
	t, err := k.Diatonic(target, false)
	if err != nil {
		return "", err
	}
	rn, err := RomanNumeral(target, t.Quality)
	if err != nil {
		return "", err
	}
	if seventh {
		return "V7/" + rn, nil
	}
	return "V/" + rn, nil
}

// spellApplied names the applied root on the letter a fifth above the
// target's letter, so C# minor's V/v is D#, not Eb.
func (k Key) spellApplied(target int) string {
	name := k.scaleNames()[target-1]
	start := strings.IndexByte(string(letters), name[0])
	root := k.Scale()[target-1].Add(7)
	return spell(letters[(start+4)%7], root)
}

// appliedStep reads "V/x" or "V7/x". ok is false when the label has no
// slash and is not an applied chord at all.
func appliedStep(key Key, label string, beats int) (step Step, ok bool, err error) {
	trimmed := strings.TrimSpace(label)
	head, tail, found := strings.Cut(trimmed, "/")
	if !found {
		return Step{}, false, nil
	}
	var seventh bool
	switch strings.TrimSpace(head) {
	case "V":
	case "V7":
		seventh = true
	default:
		return Step{}, true, fmt.Errorf("music: %q: only V and V7 can be applied to another chord (V/x, V7/x)", trimmed)
	}
	target, _, upper, suffix, accidental, err := parseLabel(tail)
	if err != nil {
		return Step{}, true, err
	}
	if err := checkTarget(key, trimmed, target, upper, suffix, accidental); err != nil {
		return Step{}, true, err
	}
	chord, err := key.appliedChord(target, seventh)
	if err != nil {
		return Step{}, true, err
	}
	if d, ok := degreeOf(key, chord.Root); ok {
		if dia, err := key.Diatonic(d, seventh); err == nil && dia.Quality == chord.Quality {
			own, _ := RomanNumeral(d, dia.Quality)
			return Step{}, true, fmt.Errorf("music: %q is %s's own %s (%s), not a secondary dominant; write %q",
				trimmed, key, own, key.ChordSymbol(chord), own)
		}
	}
	return Step{Degree: target, Seventh: seventh, Beats: beats, Applied: true}, true, nil
}

// checkTarget rejects targets that are not a major or minor triad other
// than the tonic, and targets written in the wrong case.
func checkTarget(key Key, label string, target int, upper bool, suffix string, accidental int) error {
	if accidental != 0 || (suffix != "" && suffix != "°" && suffix != "+" && suffix != "dim" && suffix != "aug") {
		return fmt.Errorf("music: %q: write the target as the key's plain triad (V/ii, V7/IV)", label)
	}
	t, err := key.Diatonic(target, false)
	if err != nil {
		return err
	}
	rn, _ := RomanNumeral(target, t.Quality)
	switch {
	case target == 1:
		return fmt.Errorf("music: %q is the dominant itself; write \"V\" or \"V7\"", label)
	case t.Quality != MajorTriad && t.Quality != MinorTriad:
		return fmt.Errorf("music: %q is not available in %s: %s (%s) is not a major or minor chord, and is not approached by its own dominant",
			label, key, rn, key.ChordSymbol(t))
	case upper != t.Quality.isUpper() || suffix != "":
		return fmt.Errorf("music: %q: the target is %s (%s) in %s; write the target as %q", label, rn, key.ChordSymbol(t), key, rn)
	}
	return nil
}

// appliedTargetOf recognises a chord as the secondary dominant of some
// degree: a major triad or dominant seventh, not the key's own, a fifth
// above a major or minor triad other than the tonic.
func appliedTargetOf(k Key, c Chord) (int, bool) {
	if c.Quality != MajorTriad && c.Quality != Dominant7 {
		return 0, false
	}
	target, ok := degreeOf(k, c.Root.Add(-7))
	if !ok || target == 1 {
		return 0, false
	}
	t, err := k.Diatonic(target, false)
	if err != nil || (t.Quality != MajorTriad && t.Quality != MinorTriad) {
		return 0, false
	}
	return target, true
}
