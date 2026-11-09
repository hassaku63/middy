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

// Tritone substitutes (BL-128, ADR-0023): subV7/x is the dominant seventh a
// semitone above x, in place of V7/x a fifth above it. The two share their
// tritone (the third and seventh swap), so the substitute pulls to x the
// same way, its root falling a semitone onto x's. Unlike V7/x it may aim at
// the tonic: subV7/I (Db7 in C major) is the commonest.

// Leading-tone diminished sevenths (BL-130, ADR-0024): vii°7/x is the
// diminished seventh on x's leading tone, a semitone below x. Between two
// chords a whole step apart it passes (I – vii°7/ii – ii: the root climbs
// C, C#, D).

// appliedKind is which applied chord a step is.
type appliedKind int

const (
	dominantOf   appliedKind = iota // V/x, V7/x
	substituteOf                    // subV7/x
	leadingOf                       // vii°7/x
)

func appliedKindOf(s Step) appliedKind {
	switch {
	case s.Substitute:
		return substituteOf
	case s.LeadingTone:
		return leadingOf
	}
	return dominantOf
}

// appliedChord returns the applied chord of target in k: its secondary
// dominant, the tritone substitute of that, or its leading-tone diminished
// seventh.
func (k Key) appliedChord(target int, seventh bool, kind appliedKind) (Chord, error) {
	t, err := k.Diatonic(target, false)
	if err != nil {
		return Chord{}, err
	}
	switch kind {
	case substituteOf:
		return Chord{Root: t.Root.Add(1), Quality: Dominant7}, nil
	case leadingOf:
		return Chord{Root: t.Root.Add(-1), Quality: Diminished7}, nil
	}
	q := MajorTriad
	if seventh {
		q = Dominant7
	}
	return Chord{Root: t.Root.Add(7), Quality: q}, nil
}

// appliedNumeral writes "V/ii", "V7/IV", "subV7/I", "vii°7/ii": the target
// as the key's triad.
func (k Key) appliedNumeral(target int, seventh bool, kind appliedKind) (string, error) {
	t, err := k.Diatonic(target, false)
	if err != nil {
		return "", err
	}
	rn, err := RomanNumeral(target, t.Quality)
	if err != nil {
		return "", err
	}
	switch {
	case kind == substituteOf:
		return "subV7/" + rn, nil
	case kind == leadingOf:
		return "vii°7/" + rn, nil
	case seventh:
		return "V7/" + rn, nil
	}
	return "V/" + rn, nil
}

// spellApplied names the applied root on the letter a fifth above the
// target's letter, so C# minor's V/v is D#, not Eb; a substitute's root on
// the letter a second above, so C major's subV7/I is Db, not C#; a
// leading-tone chord's on the letter a seventh above (a second below), so
// C major's vii°7/ii is C#, not Db.
func (k Key) spellApplied(target int, kind appliedKind) string {
	name := k.scaleNames()[target-1]
	start := strings.IndexByte(string(letters), name[0])
	switch kind {
	case substituteOf:
		return spell(letters[(start+1)%7], k.Scale()[target-1].Add(1))
	case leadingOf:
		return spell(letters[(start+6)%7], k.Scale()[target-1].Add(-1))
	}
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
	kind := dominantOf
	switch strings.TrimSpace(head) {
	case "V":
	case "V7":
		seventh = true
	case "subV7":
		seventh, kind = true, substituteOf
	case "vii°7", "viio7", "viidim7":
		seventh, kind = true, leadingOf
	default:
		return Step{}, true, fmt.Errorf("music: %q: only V, V7, subV7 and vii°7 can be applied to another chord (V/x, V7/x, subV7/x, vii°7/x)", trimmed)
	}
	sub := kind == substituteOf
	target, _, upper, suffix, accidental, err := parseLabel(tail)
	if err != nil {
		return Step{}, true, err
	}
	if err := checkTarget(key, trimmed, target, upper, suffix, accidental, sub); err != nil {
		return Step{}, true, err
	}
	chord, err := key.appliedChord(target, seventh, kind)
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
	return Step{Degree: target, Seventh: seventh, Beats: beats, Applied: true, Substitute: sub, LeadingTone: kind == leadingOf}, true, nil
}

// checkTarget rejects targets that are not a major or minor triad (other
// than the tonic, unless sub: a substitute may aim at it), and targets
// written in the wrong case.
func checkTarget(key Key, label string, target int, upper bool, suffix string, accidental int, sub bool) error {
	if !plainTarget(suffix, accidental) {
		return fmt.Errorf("music: %q: write the target as the key's plain triad (V/ii, V7/IV)", label)
	}
	t, err := key.Diatonic(target, false)
	if err != nil {
		return err
	}
	rn, _ := RomanNumeral(target, t.Quality)
	// A minor key's uppercase V is its harmonic-minor dominant (StepFromLabel),
	// on the same root as v, so V7/V and V7/v name the same chord: take the
	// first and spell it as the second (BL-131).
	harmonicV := key.Mode == Minor && target == 5 && upper && suffix == ""
	switch {
	case target == 1 && !sub:
		return fmt.Errorf("music: %q is the dominant itself; write \"V\" or \"V7\"", label)
	case t.Quality != MajorTriad && t.Quality != MinorTriad:
		return fmt.Errorf("music: %q is not available in %s: %s (%s) is not a major or minor chord, and is not approached by its own dominant",
			label, key, rn, key.ChordSymbol(t))
	case !harmonicV && (upper != t.Quality.isUpper() || suffix != ""):
		return fmt.Errorf("music: %q: the target is %s (%s) in %s; write the target as %q", label, rn, key.ChordSymbol(t), key, rn)
	}
	return nil
}

// plainTarget reports a target written as a triad: no accidental, no
// seventh (a triad's own decoration is checked against the key later).
func plainTarget(suffix string, accidental int) bool {
	switch suffix {
	case "", "°", "+", "dim", "aug":
		return accidental == 0
	}
	return false
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

// substituteTargetOf recognises a chord as the tritone substitute of some
// degree's dominant: a dominant seventh a semitone above a major or minor
// triad (the tonic included).
func substituteTargetOf(k Key, c Chord) (int, bool) {
	if c.Quality != Dominant7 {
		return 0, false
	}
	target, ok := degreeOf(k, c.Root.Add(-1))
	if !ok {
		return 0, false
	}
	t, err := k.Diatonic(target, false)
	if err != nil || (t.Quality != MajorTriad && t.Quality != MinorTriad) {
		return 0, false
	}
	return target, true
}

// leadingTargetOf recognises a chord as the leading-tone diminished seventh
// of some degree: a diminished seventh a semitone below a major or minor
// triad other than the tonic.
func leadingTargetOf(k Key, c Chord) (int, bool) {
	if c.Quality != Diminished7 {
		return 0, false
	}
	target, ok := degreeOf(k, c.Root.Add(1))
	if !ok || target == 1 {
		return 0, false
	}
	t, err := k.Diatonic(target, false)
	if err != nil || (t.Quality != MajorTriad && t.Quality != MinorTriad) {
		return 0, false
	}
	return target, true
}
