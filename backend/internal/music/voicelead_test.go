package music

import "testing"

// TestLeadVoicingsStayAboveTheBass (review of BL-086): in every key the
// led chords start at C4 or above, so the bass in octave 3 (a slash
// chord's, the patterns' bass voice, the walk) is always the lowest note.
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
		for i, h := range tl.Harmony {
			want := h.Chord.Root
			if h.Bass != nil {
				want = *h.Bass
			} else if lowest[h.Start] < 60 {
				t.Errorf("%v chord %d: lowest %d is under C4 without a slash bass", key, i, lowest[h.Start])
			}
			if lowest[h.Start].Class() != want && h.Bass != nil {
				t.Errorf("%v chord %d: the slash bass %v is not the lowest note (%d)", key, i, want, lowest[h.Start])
			}
		}
	}
}
