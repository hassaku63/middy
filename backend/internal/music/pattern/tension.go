package pattern

import (
	"slices"

	"github.com/hassaku63/middy/backend/internal/music"
)

// Tensions is how a style colours its chords (BL-125, ADR-0012 decision 4):
// the engine adds the tones, the model does not choose them, and the
// roman numerals and analysis stay at the seventh.
type Tensions struct {
	// Ninths adds the 9th to major-seventh, minor-seventh and dominant chords.
	Ninths bool `json:"ninths,omitempty"`
	// Thirteenths adds the 13th to dominant chords as well.
	Thirteenths bool `json:"thirteenths,omitempty"`
	// Add9 adds the 9th to plain major triads.
	Add9 bool `json:"add9,omitempty"`
	// Text says what the style adds, for the rhythm notes.
	Text string `json:"text"`
}

// tensionsOf lists the tones the rule adds to a chord in the key: only
// tones of the key's scale, and never a semitone above a chord tone or
// the slash bass (a minor ninth clashes).
func tensionsOf(t *Tensions, key music.Key, h music.HarmonyEvent) []music.PitchClass {
	if t == nil {
		return nil
	}
	ch := h.Chord
	var adds []int // semitones above the root
	switch ch.Quality {
	case music.Major7, music.Minor7:
		if t.Ninths {
			adds = []int{2}
		}
	case music.Dominant7:
		if t.Ninths {
			adds = append(adds, 2)
		}
		if t.Thirteenths {
			adds = append(adds, 9)
		}
	case music.MajorTriad:
		if t.Add9 {
			adds = []int{2}
		}
	}
	tones := ch.PitchClasses()
	if h.Bass != nil {
		tones = append(tones, *h.Bass)
	}
	var out []music.PitchClass
	for _, a := range adds {
		pc := music.PitchClass((int(ch.Root) + a) % 12)
		clash := slices.ContainsFunc(tones, func(c music.PitchClass) bool { return (int(c)+1)%12 == int(pc) })
		if key.Contains(pc) && !clash && !slices.Contains(tones, pc) {
			out = append(out, pc)
		}
	}
	return out
}

// colour puts a chord's tensions above its voicing, each at the nearest
// octave over the top tone.
func colour(led []music.Pitch, adds []music.PitchClass) []music.Pitch {
	if len(adds) == 0 {
		return led
	}
	out := slices.Clone(led)
	top := led[len(led)-1]
	for _, pc := range adds {
		p := top + 1
		for p.Class() != pc {
			p++
		}
		out = append(out, p)
	}
	return out
}
