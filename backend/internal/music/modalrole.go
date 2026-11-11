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
	if k.holdsTritone(c) {
		return RoleTritone
	}
	if tone, err := k.Degree(k.Mode.Characteristic()); err == nil && holds(c, tone) {
		return RoleCharacteristic
	}
	return RoleOther
}

// soundingRole is the role of the chord as it sounds, for the analysis
// (#371): the degree's role, except that a seventh that adds the scale's
// tritone to a triad without it (dorian IV7, G7 with F and B) blurs the
// mode like the tritone chord does, and a chord without the characteristic
// tone (dorian IV5) is no characteristic chord. The degree map keeps the
// triad's role.
// The tonic stays the tonic: it is the centre whatever it adds.
func soundingRole(k Key, degree int, c Chord) ModalRole {
	role := modalRole(k, degree)
	if role == "" || role == RoleTonic {
		return role
	}
	if k.holdsTritone(c) {
		return RoleTritone
	}
	// A power or sus chord can leave out the characteristic tone the
	// triad has (dorian IV5, G and D without B): it does not mark the mode
	// (#378).
	if tone, err := k.Degree(k.Mode.Characteristic()); role == RoleCharacteristic && err == nil && !holds(c, tone) {
		return RoleOther
	}
	return role
}

// holds reports whether the chord has the pitch class.
func holds(c Chord, pc PitchClass) bool {
	for _, p := range c.PitchClasses() {
		if p.Normalize() == pc.Normalize() {
			return true
		}
	}
	return false
}

// holdsTritone reports whether the chord has both tones of the scale's
// tritone.
func (k Key) holdsTritone(c Chord) bool {
	low, high := k.tritone()
	return holds(c, low) && holds(c, high)
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
	characteristicRootText = "特徴音が根音にあるので、主和音へ戻る終わり方によく使われる。"
	tritoneRoleText        = "%s三全音（全音 3 つ分離れた %s と %s）を含むので、同じ音でできた %s に引き寄せられやすく、%sの中心がぼやける。"
	// seventhTritoneText: the triad has no tritone, the seventh adds it.
	seventhTritoneText = "和音に 7 度の音を足すと三全音（全音 3 つ分離れた %s と %s）ができるので、同じ音でできた %s に引き寄せられやすく、%sの中心がぼやける。三和音の %s（%s）のほうが%sらしく聞こえる。"
)

// roleText says what a chord does in a mode, after its opening sentence
// and the characteristic tone's (#366). Empty for the tonic, for any
// other chord, and in major and minor. The role is the sounding chord's
// (soundingRole); the degree tells a seventh that adds the tritone from a
// triad that has it.
func roleText(k Key, h HarmonyEvent, degree int, role ModalRole) string {
	switch role {
	case RoleCharacteristic:
		s := " " + fmt.Sprintf(characteristicRoleText, modeName[k.Mode])
		if tone, err := k.Degree(k.Mode.Characteristic()); err == nil && h.Chord.Root.Normalize() == tone.Normalize() {
			s += characteristicRootText
		}
		return s
	case RoleTritone:
		low, high := k.tritone()
		if modalRole(k, degree) != RoleTritone {
			triad, _ := k.Diatonic(degree, false)
			numeral, _ := RomanNumeral(degree, triad.Quality)
			name := modeName[k.Mode]
			return " " + fmt.Sprintf(seventhTritoneText, k.PitchClassName(low), k.PitchClassName(high), k.parallelMajor(), name, numeral, k.ChordSymbol(triad), name)
		}
		// The tritone chord holds the characteristic tone too (dorian vi°
		// has B): say so, so that "特徴音を含む" in the overview and "特徴和音
		// ではない" here do not seem to clash (#390 T6).
		also := ""
		if tone, err := k.Degree(k.Mode.Characteristic()); err == nil && holds(h.Chord, tone) {
			also = "特徴音も含むが、"
		}
		return " " + fmt.Sprintf(tritoneRoleText, also, k.PitchClassName(low), k.PitchClassName(high), k.parallelMajor(), modeName[k.Mode])
	case RoleTonic, RoleOther:
		// The opening sentence says the tonic is the centre; any other
		// chord is said by its characteristic tone or its cadence.
	}
	return ""
}
