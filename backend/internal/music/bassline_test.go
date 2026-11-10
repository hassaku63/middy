package music

import "testing"

// TestBassLineIsOneLine (#385 F1): with a slash chord in the example, every
// chord sounds its bass under it, and the line the explanation names is
// the line of lowest notes, in one register.
func TestBassLineIsOneLine(t *testing.T) {
	cases := []struct {
		key    Key
		labels []string
		want   []Pitch
	}{
		// The descending bass: C – B – A – G – F – E – D, then G.
		{Key{C, Major}, []string{"I", "I/7", "vi", "vi/5", "IV", "I/3", "ii7", "V7"},
			[]Pitch{48, 47, 45, 43, 41, 40, 38, 43}},
		// The line cliché: A – G# – G – F# under Am.
		{Key{A, Minor}, []string{"i", "i/#7", "i/7", "i/#6"}, []Pitch{57, 56, 55, 54}},
	}
	for _, c := range cases {
		var prog Progression
		for _, l := range c.labels {
			s, err := StepFromLabel(c.key, l, 4)
			if err != nil {
				t.Fatal(err)
			}
			prog = append(prog, s)
		}
		tl, err := FromProgression(c.key, prog, 100)
		if err != nil {
			t.Fatal(err)
		}
		for i, h := range tl.Harmony {
			lowest := Pitch(127)
			for _, n := range tl.Tracks[0].Notes {
				if n.Start == h.Start {
					lowest = min(lowest, n.Pitch)
				}
			}
			if lowest != c.want[i] {
				t.Errorf("%v %s: lowest %v, want %v", c.key, h.RomanNumeral, lowest, c.want[i])
			}
		}
	}

	// Without a slash chord nothing changes: no bass under the chords.
	if line := BassLine([]Chord{{C, MajorTriad}, {G, MajorTriad}}, []*PitchClass{nil, nil}, [][]Pitch{{60}, {62}}); line != nil {
		t.Errorf("no slash chord, no line: %v", line)
	}
}
