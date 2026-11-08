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

	// A hand drum rolls its toms and has no crash; 6/8 fills its last beat.
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
