package pattern

import (
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

func barsTimeline(t *testing.T, ts music.TimeSignature, bars int) music.Timeline {
	t.Helper()
	var prog music.Progression
	for i := range bars {
		prog = append(prog, music.Step{Degree: []int{1, 6, 4, 5}[i%4], Beats: ts.BeatsPerBar()})
	}
	tl, err := music.FromProgressionIn(ts, music.Key{Tonic: music.C, Mode: music.Major}, prog, 100)
	if err != nil {
		t.Fatal(err)
	}
	return tl
}

// TestPhraseFills (BL-126, ADR-0012 decision 5): the drums mark each
// four-bar phrase with a fill in the back half of its last bar and a crash
// on the next phrase's first beat; other bars keep their figure.
func TestPhraseFills(t *testing.T) {
	out, err := Render(barsTimeline(t, music.FourFour, 8), Choice{Drums: "eight-beat"})
	if err != nil {
		t.Fatal(err)
	}
	bar := music.Tick(4 * music.DefaultPPQ)
	half := bar / 2
	toms, crashes := map[int]int{}, map[int]int{}
	for _, n := range out.Tracks[1].Notes {
		b := int(n.Start / bar)
		switch n.Pitch {
		case HighTom, MidTom, LowTom:
			if n.Start%bar < half {
				t.Errorf("a tom before the back half of bar %d", b+1)
			}
			toms[b+1]++
		case Crash:
			if n.Start%bar != 0 {
				t.Errorf("a crash off the downbeat at %d", n.Start)
			}
			crashes[b+1]++
		}
	}
	if toms[4] == 0 || toms[8] == 0 || toms[3] != 0 || toms[7] != 0 {
		t.Errorf("fills belong to bars 4 and 8: %v", toms)
	}
	if crashes[5] != 1 || len(crashes) != 1 {
		t.Errorf("one crash, on bar 5 (the last bar has nothing after it): %v", crashes)
	}
	if e := out.Rhythm.Drums.Explanation; !strings.Contains(e, "4 小節ごと") || !strings.Contains(e, "フィル") {
		t.Errorf("the drums explain the fill: %s", e)
	}

	// Shorter than a phrase: no fill, and nothing said about one.
	short, err := Render(barsTimeline(t, music.FourFour, 2), Choice{Drums: "eight-beat"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(short.Rhythm.Drums.Explanation, "フィル") {
		t.Errorf("two bars have no phrase to mark: %s", short.Rhythm.Drums.Explanation)
	}

	// 6/8 fills its last dotted beat.
	jig, err := Render(barsTimeline(t, music.SixEight, 4), Choice{Drums: "jig"})
	if err != nil {
		t.Fatal(err)
	}
	sixBar := music.BarTicks(music.SixEight, music.DefaultPPQ)
	for _, n := range jig.Tracks[1].Notes {
		if (n.Pitch == MidTom || n.Pitch == LowTom) && (n.Start < 3*sixBar+sixBar/2) {
			t.Errorf("the 6/8 fill sits in the last dotted beat, got a tom at %d", n.Start)
		}
	}
	celtic, err := Render(barsTimeline(t, music.FourFour, 8), Choice{Drums: "bodhran"})
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range celtic.Tracks[1].Notes {
		if n.Pitch == Crash || n.Pitch == Snare {
			t.Errorf("a bodhrán fill stays on its drum: %+v", n)
		}
	}
}

// TestFillsFollowTheKit (review of BL-126): the jazz ride links with soft
// snare into a crash, the bodhrán rolls its toms with no crash, a grid
// plays as written, and a short last bar whose fill would start after the
// end is not named.
func TestFillsFollowTheKit(t *testing.T) {
	bar := music.Tick(4 * music.DefaultPPQ)
	jazz, err := Render(barsTimeline(t, music.FourFour, 8), Choice{Drums: "jazz-ride"})
	if err != nil {
		t.Fatal(err)
	}
	crash := false
	for _, n := range jazz.Tracks[1].Notes {
		switch n.Pitch {
		case HighTom, MidTom, LowTom:
			t.Errorf("a jazz ride fill has no tom descent: %+v", n)
		case Snare:
			if n.Velocity > 80 {
				t.Errorf("the jazz fill's snare is light, got %d", n.Velocity)
			}
		case Crash:
			crash = true
		}
	}
	if !crash || !strings.Contains(jazz.Rhythm.Drums.Explanation, "スネアで軽くつなぐ") {
		t.Errorf("jazz: crash %v, %s", crash, jazz.Rhythm.Drums.Explanation)
	}

	g := grid(2, map[string]string{"kick": "x... x...", "snare": "..x. ..x."})
	gridded, err := Render(barsTimeline(t, music.FourFour, 8), Choice{Grid: &g})
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range gridded.Tracks[1].Notes {
		if n.Pitch != Kick && n.Pitch != Snare {
			t.Errorf("a grid plays as written, got %+v", n)
		}
	}

	// Three bars and two beats: bar 4 ends before its fill would start.
	var prog music.Progression
	for _, d := range []int{1, 6, 4} {
		prog = append(prog, music.Step{Degree: d, Beats: 4})
	}
	prog = append(prog, music.Step{Degree: 5, Beats: 2})
	tl, err := music.FromProgression(music.Key{Tonic: music.C, Mode: music.Major}, prog, 100)
	if err != nil {
		t.Fatal(err)
	}
	short, err := Render(tl, Choice{Drums: "eight-beat"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(short.Rhythm.Drums.Explanation, "フィル") {
		t.Errorf("no fill sounds, none is named: %s", short.Rhythm.Drums.Explanation)
	}
	for _, n := range short.Tracks[1].Notes {
		if n.Start >= 3*bar+bar/2 {
			t.Errorf("nothing after the end: %+v", n)
		}
	}
}
