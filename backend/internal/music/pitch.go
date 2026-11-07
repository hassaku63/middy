package music

import (
	"fmt"
)

// PitchClass is a pitch modulo octave: 0 = C ... 11 = B.
type PitchClass int

// Pitch classes, named with sharps.
const (
	C PitchClass = iota
	Cs
	D
	Ds
	E
	F
	Fs
	G
	Gs
	A
	As
	B
)

var sharpNames = [12]string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
var flatNames = [12]string{"C", "Db", "D", "Eb", "E", "F", "Gb", "G", "Ab", "A", "Bb", "B"}

// Normalize maps any integer onto 0..11.
func (pc PitchClass) Normalize() PitchClass {
	return PitchClass(((int(pc) % 12) + 12) % 12)
}

// Add transposes the pitch class by semitones, wrapping around the octave.
func (pc PitchClass) Add(semitones int) PitchClass {
	return (pc + PitchClass(semitones)).Normalize()
}

// String renders the pitch class with sharps ("C#").
func (pc PitchClass) String() string {
	return sharpNames[pc.Normalize()]
}

// FlatName renders the pitch class with flats ("Db").
func (pc PitchClass) FlatName() string {
	return flatNames[pc.Normalize()]
}

// Pitch is a MIDI note number in 0..127. Middle C (C4) is 60.
type Pitch int

// MIDI pitch range.
const (
	MinPitch Pitch = 0
	MaxPitch Pitch = 127
)

// NewPitch builds a Pitch from a pitch class and a scientific octave number,
// where C4 is middle C (MIDI 60). The result may be out of range; check Valid.
func NewPitch(pc PitchClass, octave int) Pitch {
	return Pitch((octave+1)*12 + int(pc.Normalize()))
}

// Valid reports whether the pitch is representable in MIDI.
func (p Pitch) Valid() bool {
	return p >= MinPitch && p <= MaxPitch
}

// Class returns the pitch class.
func (p Pitch) Class() PitchClass {
	return PitchClass(p).Normalize()
}

// Octave returns the scientific octave number (C4 = 60 -> 4).
func (p Pitch) Octave() int {
	return int(p)/12 - 1
}

// Transpose shifts the pitch by semitones. The result may be out of range.
func (p Pitch) Transpose(semitones int) Pitch {
	return p + Pitch(semitones)
}

// String renders scientific pitch notation with sharps ("C#4").
func (p Pitch) String() string {
	return fmt.Sprintf("%s%d", p.Class(), p.Octave())
}
