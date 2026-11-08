package pattern

import (
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

func waltzTimeline(t *testing.T, beats int, degrees ...int) music.Timeline {
	t.Helper()
	var prog music.Progression
	for _, d := range degrees {
		prog = append(prog, music.Step{Degree: d, Beats: beats})
	}
	tl, err := music.FromProgressionIn(music.ThreeFour, music.Key{Tonic: music.C, Mode: music.Major}, prog, 150)
	if err != nil {
		t.Fatal(err)
	}
	return tl
}

func TestCatalogueIsSplitByMeter(t *testing.T) {
	ids := func(as []Accompaniment) string {
		var s []string
		for _, a := range as {
			s = append(s, a.ID)
		}
		return strings.Join(s, ",")
	}
	if got := ids(AccompanimentsFor(music.ThreeFour)); got != "block,drone,waltz,arpeggio-3" {
		t.Errorf("3/4 accompaniments = %s", got)
	}
	if got := ids(AccompanimentsFor(music.FourFour)); strings.Contains(got, "waltz") || !strings.Contains(got, "block") {
		t.Errorf("4/4 accompaniments = %s", got)
	}
	if ds := DrumPatternsFor(music.ThreeFour); len(ds) != 1 || ds[0].ID != "waltz" {
		t.Errorf("3/4 drums = %+v", ds)
	}
}

func TestWaltzPlaysOomPahPahInEveryBar(t *testing.T) {
	tl := waltzTimeline(t, 3, 1, 5) // two bars of 3/4
	got, err := Apply(tl, mustAcc(t, "waltz"), nil)
	if err != nil {
		t.Fatal(err)
	}
	starts := map[music.Tick]int{}
	for _, n := range notesIn(got, music.ChordTrackName) {
		starts[n.Start]++
	}
	// bass alone on 1, the chord (3 notes) on 2 and 3, in each of two bars of 1440 ticks.
	want := map[music.Tick]int{0: 1, 480: 3, 960: 3, 1440: 1, 1920: 3, 2400: 3}
	for s, n := range want {
		if starts[s] != n {
			t.Errorf("notes at %d = %d, want %d (all: %v)", s, starts[s], n, starts)
		}
	}
	if !strings.Contains(got.Rhythm.Accompaniment.Explanation, "すべて拍の頭（1・2・3）") {
		t.Errorf("explanation should name three beats: %s", got.Rhythm.Accompaniment.Explanation)
	}
}

func TestBlockLastsOneBarInAnyMeter(t *testing.T) {
	tl := waltzTimeline(t, 6, 1) // one chord over two bars of 3/4
	got, err := Apply(tl, mustAcc(t, "block"), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range notesIn(got, music.ChordTrackName) {
		if n.Duration != 1440 || (n.Start != 0 && n.Start != 1440) {
			t.Fatalf("block in 3/4 must restrike each 1440-tick bar without overlap: %+v", n)
		}
	}
}

func TestWaltzDrumsAndMeterMismatch(t *testing.T) {
	tl := waltzTimeline(t, 3, 1, 4)
	d, err := FindDrums("waltz")
	if err != nil {
		t.Fatal(err)
	}
	got, err := Apply(tl, mustAcc(t, "waltz"), &d)
	if err != nil {
		t.Fatal(err)
	}
	var kicks []music.Tick
	for _, n := range notesIn(got, "drums") {
		if n.Pitch == Kick {
			kicks = append(kicks, n.Start)
		}
	}
	if len(kicks) != 2 || kicks[0] != 0 || kicks[1] != 1440 {
		t.Errorf("one kick on beat 1 of each bar: %v", kicks)
	}
	if _, err := Apply(tl, mustAcc(t, "strum"), nil); err == nil || !strings.Contains(err.Error(), "written for 4/4") {
		t.Errorf("a 4/4 template in 3/4 must be rejected: %v", err)
	}
	four := progression(t, music.Key{Tonic: music.C, Mode: music.Major}, 4, 1)
	if _, err := Apply(four, mustAcc(t, "waltz"), nil); err == nil || !strings.Contains(err.Error(), "written for 3/4") {
		t.Errorf("a 3/4 template in 4/4 must be rejected: %v", err)
	}
}
