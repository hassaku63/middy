package pattern

import (
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

// sixEightTimeline renders one chord per degree, each `beats` dotted quarters long.
func sixEightTimeline(t *testing.T, beats int, degrees ...int) music.Timeline {
	t.Helper()
	var prog music.Progression
	for _, d := range degrees {
		prog = append(prog, music.Step{Degree: d, Beats: beats})
	}
	tl, err := music.FromProgressionIn(music.SixEight, music.Key{Tonic: music.C, Mode: music.Major}, prog, 60)
	if err != nil {
		t.Fatal(err)
	}
	return tl
}

func TestSixEightCatalogue(t *testing.T) {
	var accs, drs []string
	for _, a := range AccompanimentsFor(music.SixEight) {
		accs = append(accs, a.ID)
	}
	for _, d := range DrumPatternsFor(music.SixEight) {
		drs = append(drs, d.ID)
	}
	if got := strings.Join(accs, ","); got != "block,rocking-6,arpeggio-6" {
		t.Errorf("6/8 accompaniments = %s", got)
	}
	if got := strings.Join(drs, ","); got != "six-eight,jig" {
		t.Errorf("6/8 drums = %s", got)
	}
}

// TestRockingSplitsEachBeatInThree checks the compound division: bass on
// the head of each dotted-quarter beat, the chord on the other two eighths.
func TestRockingSplitsEachBeatInThree(t *testing.T) {
	got, err := Apply(sixEightTimeline(t, 2, 1), mustAcc(t, "rocking-6"), nil)
	if err != nil {
		t.Fatal(err)
	}
	starts := map[music.Tick]int{}
	for _, n := range notesIn(got, music.ChordTrackName) {
		starts[n.Start]++
	}
	want := map[music.Tick]int{0: 1, 240: 3, 480: 3, 720: 1, 960: 3, 1200: 3}
	for s, n := range want {
		if starts[s] != n {
			t.Errorf("notes at %d = %d, want %d (all: %v)", s, starts[s], n, starts)
		}
	}
	part := got.Rhythm.Accompaniment
	if part.OnBeat != 2 || part.OffBeat != 4 || part.Syncopated {
		t.Errorf("two beats, four divisions, no syncopation: %+v", part)
	}
	for _, want := range []string{"拍の頭（1・2）", "3 つに分"} {
		if !strings.Contains(part.Explanation, want) {
			t.Errorf("explanation lacks %q: %s", want, part.Explanation)
		}
	}
}

func TestJigDrumsKickOnBothBeats(t *testing.T) {
	d, err := FindDrums("jig")
	if err != nil {
		t.Fatal(err)
	}
	got, err := Apply(sixEightTimeline(t, 2, 1, 5), mustAcc(t, "arpeggio-6"), &d)
	if err != nil {
		t.Fatal(err)
	}
	var kicks []music.Tick
	for _, n := range notesIn(got, "drums") {
		if n.Pitch == Kick {
			kicks = append(kicks, n.Start)
		}
	}
	if len(kicks) != 4 || kicks[0] != 0 || kicks[1] != 720 || kicks[2] != 1440 || kicks[3] != 2160 {
		t.Errorf("a kick on each dotted-quarter beat of two bars: %v", kicks)
	}
	if _, err := Apply(sixEightTimeline(t, 2, 1), mustAcc(t, "waltz"), nil); err == nil || !strings.Contains(err.Error(), "written for 3/4") {
		t.Errorf("a 3/4 template in 6/8 must be rejected: %v", err)
	}
}

// TestSwingIsRefusedInSixEight: swing turns even eighths into a
// long-short pair of triplets; 6/8 already divides each beat in three.
func TestSwingIsRefusedInSixEight(t *testing.T) {
	_, err := Render(sixEightTimeline(t, 2, 1), Choice{Accompaniment: "rocking-6", Feel: FeelSwing})
	if err == nil || !strings.Contains(err.Error(), "6/8") {
		t.Errorf("swing in 6/8 must be refused: %v", err)
	}
	if _, err := Render(sixEightTimeline(t, 2, 1), Choice{Accompaniment: "rocking-6"}); err != nil {
		t.Errorf("straight 6/8 must render: %v", err)
	}
}
