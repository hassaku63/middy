package pattern

import (
	"slices"

	"github.com/hassaku63/middy/backend/internal/music"
)

// walkLow is the lowest root of the walking bass (E2): its roots lie in
// E2..D#3 and its line stays under the chords, which start at C4.
const walkLow music.Pitch = 40

// walkLine is the walking bass under chord i for n walk hits (BL-086): the
// root first, then the chord's tones upward, and on the last hit the
// semitone below the next chord's root. The last chord leads back to the
// first. A slash chord's bass stands in for its root.
func walkLine(harmony []music.HarmonyEvent, i, n int) []music.Pitch {
	root := func(h music.HarmonyEvent) music.Pitch {
		pc := h.Chord.Root
		if h.Bass != nil {
			pc = *h.Bass
		}
		return walkLow + music.Pitch((int(pc)-int(walkLow.Class())+12)%12)
	}
	h := harmony[i]
	// The root (or the slash bass), then the chord's other tones upward
	// above it.
	low := root(h)
	tones := []music.Pitch{low}
	for _, pc := range h.Chord.PitchClasses() {
		if pc != low.Class() {
			tones = append(tones, low+music.Pitch((int(pc)-int(low.Class())+12)%12))
		}
	}
	slices.Sort(tones)
	line := make([]music.Pitch, n)
	for k := range line {
		switch k {
		case 0:
			line[k] = root(h)
		case n - 1:
			line[k] = root(harmony[(i+1)%len(harmony)]) - 1
		default:
			line[k] = tones[min(k, len(tones)-1)]
		}
	}
	return line
}
