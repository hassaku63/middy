package music

import "testing"

// TestLeadVoicingsStayAboveTheBass (review of BL-086): in every key the
// led chords start at C4 or above, so the bass under them (the bass line,
// the patterns' bass voice, the walk) is always the lowest note.
func TestLeadVoicingsStayAboveTheBass(t *testing.T) {
	for tonic := range PitchClass(12) {
		key := Key{Tonic: tonic, Mode: Major}
		prog := Progression{
			{Degree: 6, Beats: 4},
			{Degree: 5, Beats: 4, Bass: &Bass{Degree: 7}},
			{Degree: 1, Beats: 4},
			{Degree: 4, Seventh: true, Beats: 4},
			{Degree: 2, Seventh: true, Beats: 4},
		}
		tl, err := FromProgression(key, prog, 100)
		if err != nil {
			t.Fatal(err)
		}
		lowest := map[Tick]Pitch{}
		for _, n := range tl.Tracks[0].Notes {
			if l, ok := lowest[n.Start]; !ok || n.Pitch < l {
				lowest[n.Start] = n.Pitch
			}
		}
		// With a slash chord in the example, every chord sounds its bass
		// under it, on one line (#385); the led chord stays at C4 or above.
		for i, h := range tl.Harmony {
			want := h.Chord.Root
			if h.Bass != nil {
				want = *h.Bass
			}
			if lowest[h.Start].Class() != want || lowest[h.Start] >= 60 {
				t.Errorf("%v chord %d: the bass %v is not the lowest note under C4 (%d)", key, i, want, lowest[h.Start])
			}
			for _, n := range tl.Tracks[0].Notes {
				if n.Start == h.Start && n.Pitch != lowest[h.Start] && n.Pitch < 60 {
					t.Errorf("%v chord %d: the chord's note %d is under C4", key, i, n.Pitch)
				}
			}
		}

	}
}
