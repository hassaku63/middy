package music

import "fmt"

// ModalRole is what a diatonic chord does in a mode, where functions
// (T, S, D) do not hold (#366, ADR-0047). It is found from the mode's
// scale alone, by the rule modal harmony is taught with: the
// characteristic tone sets the mode apart, so chords that carry it sound
// of the mode; the scale's tritone pulls towards the parallel major, so
// chords that carry it blur the mode's centre.
type ModalRole string

const (
	// RoleTonic is the mode's centre, the chord on its first degree.
	RoleTonic ModalRole = "tonic"
	// RoleCharacteristic carries the characteristic tone and not the
	// tritone (dorian IV and ii, mixolydian ♭VII and v, phrygian ♭II and
	// ♭vii).
	RoleCharacteristic ModalRole = "characteristic"
	// RoleTritone carries both tones of the scale's tritone (dorian vi°,
	// mixolydian iii°, phrygian v°).
	RoleTritone ModalRole = "tritone"
	// RoleOther is any other chord of the mode.
	RoleOther ModalRole = "other"
)

// modalRole returns the role of the triad on a degree of a mode, and ""
// in major and minor, where the chord has a function instead.
func modalRole(k Key, degree int) ModalRole {
	if k.Mode == Major || k.Mode == Minor || degree < 1 || degree > 7 {
		return ""
	}
	if degree == 1 {
		return RoleTonic
	}
	c, err := k.Diatonic(degree, false)
	if err != nil {
		return RoleOther
	}
	has := func(pc PitchClass) bool {
		for _, p := range c.PitchClasses() {
			if p.Normalize() == pc.Normalize() {
				return true
			}
		}
		return false
	}
	low, high := k.tritone()
	if has(low) && has(high) {
		return RoleTritone
	}
	if tone, err := k.Degree(k.Mode.Characteristic()); err == nil && has(tone) {
		return RoleCharacteristic
	}
	return RoleOther
}

// tritone returns the two tones of the scale a tritone apart, the lower
// one first as the parallel major hears it: its fourth, then its leading
// tone (F and B in the white-key modes).
func (k Key) tritone() (PitchClass, PitchClass) {
	s := k.Scale()
	for _, a := range s {
		for _, b := range s {
			// b is the leading tone of the parallel major: its tonic, a
			// semitone above, is in the scale too.
			if (b-a+12)%12 == 6 && k.Contains(b.Add(1)) {
				return a, b
			}
		}
	}
	return s[0], s[0] // every diatonic mode has one; unreachable
}

// parallelMajor is the major key with the mode's notes (C major for D
// dorian, G mixolydian and E phrygian): its tonic is a semitone above the
// tritone's upper tone, the leading tone.
func (k Key) parallelMajor() Key {
	_, lead := k.tritone()
	return Key{Tonic: lead.Add(1).Normalize(), Mode: Major}
}

const (
	characteristicRoleText = "三全音を含まない特徴和音で、%sの響きをはっきりさせる。"
	characteristicRootText = "特徴音が根音にあるので、主和音へ戻る力がいちばん強い。"
	tritoneRoleText        = "三全音（%s と %s）を含むので、同じ音でできた %s に引き寄せられやすく、%sの中心がぼやける。"
)

// roleText says what a chord does in a mode, after its opening sentence
// and the characteristic tone's (#366). Empty for the tonic, for any
// other chord, and in major and minor.
func roleText(k Key, h HarmonyEvent, role ModalRole) string {
	switch role {
	case RoleCharacteristic:
		s := " " + fmt.Sprintf(characteristicRoleText, modeName[k.Mode])
		if tone, err := k.Degree(k.Mode.Characteristic()); err == nil && h.Chord.Root.Normalize() == tone.Normalize() {
			s += characteristicRootText
		}
		return s
	case RoleTritone:
		low, high := k.tritone()
		return " " + fmt.Sprintf(tritoneRoleText, k.PitchClassName(low), k.PitchClassName(high), k.parallelMajor(), modeName[k.Mode])
	case RoleTonic, RoleOther:
		// The opening sentence says the tonic is the centre; any other
		// chord is said by its characteristic tone or its cadence.
	}
	return ""
}
