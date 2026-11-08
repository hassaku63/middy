package pattern

import "github.com/hassaku63/middy/backend/internal/music"

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
		return music.NewPitch(pc, music.ChordOctave-1)
	}
	h := harmony[i]
	tones := h.Chord.Voicing(music.ChordOctave - 1)
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
