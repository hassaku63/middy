package pattern

import (
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

// TestDroneReview: the drone stays under the chords in every key, holds
// through chord changes instead of being struck again, and is described
// as the one attack per bar it is.
func TestDroneReview(t *testing.T) {
	for tonic := range music.PitchClass(12) {
		key := music.Key{Tonic: tonic, Mode: music.Major}
		prog := music.Progression{{Degree: 1, Beats: 2}, {Degree: 4, Beats: 2}, {Degree: 5, Beats: 4}, {Degree: 1, Beats: 4}}
		tl, err := music.FromProgression(key, prog, 90)
		if err != nil {
			t.Fatal(err)
		}
		out, err := Render(tl, Choice{Accompaniment: "drone"})
		if err != nil {
			t.Fatal(err)
		}
		tonicNotes := 0
		seen := map[[2]int]bool{}
		for _, n := range out.Tracks[0].Notes {
			k := [2]int{int(n.Start), int(n.Pitch)}
			if seen[k] {
				t.Errorf("%v: %v struck twice at %d", key, n.Pitch, n.Start)
			}
			seen[k] = true
			if n.Pitch.Class() == tonic && n.Pitch < 60 {
				tonicNotes++
				if n.Start != 0 || n.Start+n.Duration != tl.LengthTicks() {
					t.Errorf("%v: the drone's tonic should hold from 0 to the end, got %d+%d", key, n.Start, n.Duration)
				}
			}
		}
		if tonicNotes != 1 {
			t.Errorf("%v: %d drone tonic notes, want one held note", key, tonicNotes)
		}
		if e := out.Rhythm.Accompaniment.Explanation; !strings.Contains(e, "伴奏が 1 回") {
			t.Errorf("%v: the drone sounds once a bar: %s", key, e)
		}
	}
}
