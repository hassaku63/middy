package music

// Spelling by key (backlog BL-079, ADR-0013). A pitch class has several
// names (A# and Bb); the key decides which one is right. The rule here is
// the one notation uses: find the key's parent major, spell that scale on
// seven consecutive letters starting from its tonic, and name every
// diatonic pitch by its letter. Minor keys and, later, other modes inherit
// the signature of their parent major, so D dorian spells like C major.

// parentOffset is where each mode's tonic sits inside its parent major
// scale, in semitones (ionian 0, dorian 2, ..., aeolian 9).
var parentOffset = map[Mode]int{Major: 0, Minor: 9, Dorian: 2, Mixolydian: 7}

// majorTonicNames spells the tonic of every major key with the fewer
// accidentals; a tie (F#/Gb, six each) goes to flats, as does Db over C#.
var majorTonicNames = [12]string{"C", "Db", "D", "Eb", "E", "F", "Gb", "G", "Ab", "A", "Bb", "B"}

var letters = []byte("CDEFGAB")
var letterPitch = map[byte]PitchClass{'C': C, 'D': D, 'E': E, 'F': F, 'G': G, 'A': A, 'B': B}

// parentMajor returns the major key whose signature this key uses.
func (k Key) parentMajor() PitchClass {
	return k.Tonic.Add(-parentOffset[k.Mode])
}

// usesFlats reports whether the key's signature is written with flats.
func (k Key) usesFlats() bool {
	name := majorTonicNames[k.parentMajor().Normalize()]
	return len(name) == 2 && name[1] == 'b' || name == "F"
}

// scaleNames spells the key's seven degrees, tonic first.
func (k Key) scaleNames() [7]string {
	parent := majorTonicNames[k.parentMajor().Normalize()]
	// Letters of the parent major scale, then rotate to this key's tonic.
	start := 0
	for i, l := range letters {
		if l == parent[0] {
			start = i
		}
	}
	var parentScale [7]string
	parentPCs := Key{Tonic: k.parentMajor(), Mode: Major}.Scale()
	for i := range 7 {
		letter := letters[(start+i)%7]
		parentScale[i] = spell(letter, parentPCs[i])
	}
	// This key's tonic is some degree of the parent; rotate so it comes first.
	shift := 0
	for i, pc := range parentPCs {
		if pc == k.Tonic.Normalize() {
			shift = i
		}
	}
	var out [7]string
	for i := range 7 {
		out[i] = parentScale[(shift+i)%7]
	}
	return out
}

// spell names pc on the given letter, adding the accidentals needed.
func spell(letter byte, pc PitchClass) string {
	diff := int(pc.Normalize()) - int(letterPitch[letter])
	diff = ((diff+6)%12+12)%12 - 6 // into -6..5
	name := string(letter)
	switch {
	case diff > 0:
		for range diff {
			name += "#"
		}
	case diff < 0:
		for range -diff {
			name += "b"
		}
	}
	return name
}

// TonicName spells the tonic as the key's signature wants it: "Bb" for
// Bb major, "C#" for C# minor.
func (k Key) TonicName() string {
	return k.scaleNames()[0]
}

// PitchClassName spells a pitch class in this key: by its letter when it
// is diatonic, otherwise on the key's side (flats in flat keys).
func (k Key) PitchClassName(pc PitchClass) string {
	pc = pc.Normalize()
	for i, spc := range k.Scale() {
		if spc == pc {
			return k.scaleNames()[i]
		}
	}
	if k.usesFlats() {
		return pc.FlatName()
	}
	return pc.String()
}

// ChordSymbol renders a chord as the key spells it: "Bbmaj7" in Bb major
// where Chord.String would say "A#maj7".
func (k Key) ChordSymbol(c Chord) string {
	return k.PitchClassName(c.Root) + c.Quality.Symbol()
}
