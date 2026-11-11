package pattern

import (
	"slices"

	"github.com/hassaku63/middy/backend/internal/music"
)

// walkLow is the lowest root of the walking bass (E2): its roots lie in
// E2..D#3 and its line stays under the chords, which start at C4.
const walkLow music.Pitch = 40

// walkLine is the walking bass under chord i for n walk hits (BL-086): the
// root first, then the chord's tones up and back down again, so a long
// chord keeps walking (#404 S7), and on the last hit the semitone below
// the next chord's root. The last chord leads back to the first. A slash
// chord's bass stands in for its root. Tones that would reach the chords
// (C4) are left out, so the line stays under them.
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
		if p := low + music.Pitch((int(pc)-int(low.Class())+12)%12); pc != low.Class() && p < music.NewPitch(music.C, music.ChordOctave) {
			tones = append(tones, p)
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
			line[k] = tones[bounce(k, len(tones))]
		}
	}
	return line
}

// bounce is the k-th step of a walk up and down m tones from the first:
// 0, 1, …, m-1, m-2, …, 1, 0, 1, … (always 0 for one tone).
func bounce(k, m int) int {
	if m < 2 {
		return 0
	}
	period := 2 * (m - 1)
	p := k % period
	if p < m {
		return p
	}
	return period - p
}
