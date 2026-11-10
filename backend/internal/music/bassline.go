package music

import "slices"

// Bass line range: C2 to B3, under the chords (C4 up, BL-086). A line
// that starts on C3 can step down a seventh to D2.
const (
	bassLow  Pitch = 36
	bassHigh Pitch = 59
)

// BassLine is the bass of every chord in a progression with a slash chord
// (#385), or nil in one without: each chord's bass (its slash bass, else
// its root), at the octave nearest the bass before and below the chord's
// lowest note, so a line that steps C – B – A – G, or a line cliché
// A – G# – G – F#, sounds and is drawn as one line in one register. basses
// holds each chord's slash bass (nil for none); led is LeadVoicings.
func BassLine(chords []Chord, basses []*PitchClass, led [][]Pitch) []Pitch {
	if !slices.ContainsFunc(basses, func(b *PitchClass) bool { return b != nil }) {
		return nil
	}
	bassAt := func(i int) PitchClass {
		if basses[i] != nil {
			return *basses[i]
		}
		return chords[i].Root
	}
	line := make([]Pitch, len(chords))
	prev := NewPitch(bassAt(0), ChordOctave-1) // the line starts in octave 3
	for i := range chords {
		pc := bassAt(i)
		ceiling := bassHigh
		if len(led[i]) > 0 && slices.Min(led[i])-1 < ceiling {
			ceiling = slices.Min(led[i]) - 1
		}
		best := Pitch(-1)
		for p := LowestFrom(bassLow, pc); p <= ceiling; p += 12 {
			if best < 0 || abs(int(p)-int(prev)) < abs(int(best)-int(prev)) {
				best = p
			}
		}
		if best < 0 { // no room under the chord in range: the octave below it
			best = LowestFrom(ceiling-11, pc)
		}
		line[i], prev = best, best
	}
	return line
}

// LowestFrom is the lowest pitch of class pc at or above p.
func LowestFrom(p Pitch, pc PitchClass) Pitch {
	for p.Class() != pc.Normalize() {
		p++
	}
	return p
}
