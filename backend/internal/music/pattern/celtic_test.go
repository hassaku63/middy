package pattern

import (
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

// TestReelPlaysBassOnTheBeatsAndChordsOffThem (BL-107): bass on 1 and 3,
// a short chord on every offbeat, on a steel-string guitar.
func TestReelPlaysBassOnTheBeatsAndChordsOffThem(t *testing.T) {
	tl := progression(t, music.Key{Tonic: music.D, Mode: music.Dorian}, 4, 1)
	got, err := Apply(tl, mustAcc(t, "reel"), nil)
	if err != nil {
		t.Fatal(err)
	}
	starts := map[music.Tick]int{}
	for _, n := range notesIn(got, music.ChordTrackName) {
		starts[n.Start]++
		if n.Start%480 != 0 && n.Duration > 240 {
			t.Errorf("offbeat chords should be short: %+v", n)
		}
	}
	want := map[music.Tick]int{0: 1, 240: 3, 720: 3, 960: 1, 1200: 3, 1680: 3}
	for s, n := range want {
		if starts[s] != n {
			t.Errorf("notes at %d = %d, want %d (all: %v)", s, starts[s], n, starts)
		}
	}
	if got.Tracks[0].Program != 25 {
		t.Errorf("reel should use a steel-string guitar (25), got %d", got.Tracks[0].Program)
	}
}

// TestBodhranKeepsSnareAndHatsOut (BL-107): low tom on 1 and 3, a mid tom
// on every eighth, a side stick on 2 and 4; no kick, snare or hi-hat, and
// the description counts the low tom and side stick as the groove.
func TestBodhranKeepsSnareAndHatsOut(t *testing.T) {
	tl := progression(t, music.Key{Tonic: music.D, Mode: music.Dorian}, 4, 1)
	d, err := FindDrums("bodhran")
	if err != nil {
		t.Fatal(err)
	}
	got, err := Apply(tl, mustAcc(t, "reel"), &d)
	if err != nil {
		t.Fatal(err)
	}
	count := map[music.Pitch]int{}
	for _, n := range notesIn(got, "drums") {
		count[n.Pitch]++
	}
	if count[Kick]+count[Snare]+count[ClosedHat]+count[OpenHat] != 0 {
		t.Errorf("a bodhrán has no kick, snare or hi-hat: %v", count)
	}
	if count[LowTom] != 2 || count[MidTom] != 6 || count[SideStick] != 2 {
		t.Errorf("low tom on 1 and 3, mid tom on the other eighths, side stick on 2 and 4: %v", count)
	}
	part := got.Rhythm.Drums
	if part.Hits != 4 || part.OnBeat != 4 {
		t.Errorf("the groove is the low tom and the side stick, all on the beats: %+v", part)
	}
	if !strings.Contains(part.Explanation, "低いタムとサイドスティック") {
		t.Errorf("explanation should name the bodhrán's strokes: %s", part.Explanation)
	}
	kit, _ := FindDrums("eight-beat")
	eight, err := Apply(tl, mustAcc(t, "block"), &kit)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(eight.Rhythm.Drums.Explanation, "キックとスネア") {
		t.Errorf("a kit keeps its wording: %s", eight.Rhythm.Drums.Explanation)
	}
}
