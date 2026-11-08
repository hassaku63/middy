package music

import "slices"

// The range the led chords keep their lowest tone in (C4..B4, where the
// first chord's root position already lies), so a long progression does
// not drift up or down the keyboard (BL-086), and the bass in octave 3 (a
// slash chord's, a pattern's) stays under the chord.
const (
	leadLow  Pitch = 60
	leadHigh Pitch = 71
)

// LeadVoicings voices each chord from the one before with the least
// movement (BL-086): the first chord is the close root position at
// ChordOctave, and each next one is the inversion, kept within
// leadLow..leadHigh, whose voices move least. Ties keep the lowest tone
// nearer middle C.
func LeadVoicings(chords []Chord) [][]Pitch {
	out := make([][]Pitch, len(chords))
	for i, ch := range chords {
		close := ch.Voicing(ChordOctave)
		if i == 0 {
			out[i] = close
			continue
		}
		prev := out[i-1]
		best, bestCost := close, -1
		for _, v := range inversions(close) {
			c := movement(prev, v)
			if bestCost < 0 || c < bestCost || (c == bestCost && dist(v[0], 60) < dist(best[0], 60)) {
				best, bestCost = v, c
			}
		}
		out[i] = best
	}
	return out
}

// inversions lists the chord's close voicings on each of its tones, with
// the lowest tone in leadLow..leadHigh.
func inversions(close []Pitch) [][]Pitch {
	var out [][]Pitch
	for k := range close {
		pcs := append(slices.Clone(close[k:]), close[:k]...)
		for low := leadLow; low <= leadHigh; low++ {
			if (int(low)-int(pcs[0]))%12 != 0 {
				continue
			}
			v := []Pitch{low}
			for _, p := range pcs[1:] {
				next := v[len(v)-1] + 1
				for (int(next)-int(p))%12 != 0 {
					next++
				}
				v = append(v, next)
			}
			out = append(out, v)
		}
	}
	return out
}

// movement is how far the voices travel from a to b: voice by voice when
// the chords have as many tones, else each new tone from its nearest old one.
func movement(a, b []Pitch) int {
	total := 0
	if len(a) == len(b) {
		for i := range a {
			total += dist(a[i], b[i])
		}
		return total
	}
	for _, p := range b {
		nearest := -1
		for _, q := range a {
			if d := dist(p, q); nearest < 0 || d < nearest {
				nearest = d
			}
		}
		total += nearest
	}
	return total
}

func dist(a, b Pitch) int { return abs(int(a) - int(b)) }
