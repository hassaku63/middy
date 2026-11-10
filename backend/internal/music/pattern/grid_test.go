package pattern

import (
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

func grid(division int, voices map[string]string) music.DrumGrid {
	return music.DrumGrid{Division: division, Voices: voices}
}

// TestGridValidation (BL-122): a grid is checked against the meter and
// the tempo, and every problem is named so a retry can fix it.
func TestGridValidation(t *testing.T) {
	tests := []struct {
		name  string
		ts    music.TimeSignature
		tempo float64
		g     music.DrumGrid
		want  string // "" means valid
	}{
		{"eighths in 4/4", music.FourFour, 120, grid(2, map[string]string{"kick": "x... x...", "snare": "..x. ..x.", "hat": "xxxx xxxx"}), ""},
		{"bars may be marked", music.FourFour, 120, grid(2, map[string]string{"kick": "x.|..|x.|.."}), ""},
		{"triplets in 3/4", music.ThreeFour, 120, grid(3, map[string]string{"ride": "x..x.xx.."}), ""},
		{"eighths in 6/8", music.SixEight, 60, grid(3, map[string]string{"kick": "x..x..", "snare": "..o..o"}), ""},
		{"unknown division", music.FourFour, 120, grid(5, map[string]string{"kick": "x...."}), "division must be one of [2 3 4] in 4/4, got 5"},
		{"6/8 divides in three", music.SixEight, 60, grid(2, map[string]string{"kick": "x.x."}), "division must be one of [3 6] in 6/8, got 2"},
		{"wrong length", music.FourFour, 120, grid(2, map[string]string{"kick": "x..."}), `voices["kick"] has 4 steps, want 8 (4 beats of 2)`},
		{"unknown voice", music.FourFour, 120, grid(2, map[string]string{"cowbell": "x......."}), `unknown voice "cowbell"; use one of`},
		{"unknown step", music.FourFour, 120, grid(2, map[string]string{"kick": "x..-x..."}), `voices["kick"] step 4: "-" is not one of X, x, o or a dot`},
		{"silent", music.FourFour, 120, grid(2, map[string]string{"kick": "........"}), "the grid has no strokes"},
		{"three hands at once", music.FourFour, 120, grid(2, map[string]string{"snare": "x.......", "hat": "x.......", "ride": "x......."}), "step 1 needs 3 hands (snare, hat, ride); two hands play at most 2"},
		{"feet do not count as hands", music.FourFour, 120, grid(2, map[string]string{"kick": "x.......", "pedal-hat": "x.......", "snare": "x.......", "hat": "x......."}), ""},
		{"sixteenths too fast", music.FourFour, 160, grid(4, map[string]string{"hat": "xxxxxxxxxxxxxxxx"}), "steps of 94 ms at 160 BPM are shorter than 100 ms; use a coarser division"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GridDrums(tt.g, tt.ts, tt.tempo)
			switch {
			case tt.want == "" && err != nil:
				t.Fatalf("should be valid: %v", err)
			case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
				t.Fatalf("error %v should mention %q", err, tt.want)
			}
		})
	}
}

// TestGridRenders (BL-122): the grid becomes a bar of GM drum hits, with
// the accent, normal and ghost strokes at their own velocities.
func TestGridRenders(t *testing.T) {
	d, err := GridDrums(grid(2, map[string]string{"kick": "X...x...", "snare": "..x...xo"}), music.FourFour, 120)
	if err != nil {
		t.Fatal(err)
	}
	type hit struct {
		start music.Tick
		note  music.Pitch
		vel   int
	}
	var got []hit
	for _, h := range d.Hits {
		got = append(got, hit{h.Start, h.Note, h.Velocity})
	}
	want := []hit{{0, Kick, 104}, {960, Kick, 80}, {480, Snare, 80}, {1440, Snare, 80}, {1680, Snare, 44}}
	if len(got) != len(want) {
		t.Fatalf("hits %v, want %v", got, want)
	}
	for _, w := range want {
		found := false
		for _, g := range got {
			found = found || g == w
		}
		if !found {
			t.Errorf("missing %v in %v", w, got)
		}
	}
	if d.Meter == nil || *d.Meter != music.FourFour || d.ID != GridID {
		t.Errorf("a grid is written for its meter: %+v", d)
	}
}

// TestBackbeatAndHalfTime (BL-122): the explanation names a snare on 2
// and 4 as a backbeat and a snare on 3 alone as half time, for the grid
// and the catalogue alike, and gives a grid's density.
func TestBackbeatAndHalfTime(t *testing.T) {
	tl := blockTimeline(t, music.FourFour)
	explain := func(c Choice) string {
		out, err := Render(tl, c)
		if err != nil {
			t.Fatal(err)
		}
		return out.Rhythm.Drums.Explanation
	}
	if e := explain(Choice{Drums: "eight-beat"}); !strings.Contains(e, "2・4 拍目のスネアがバックビート") || !strings.Contains(e, "弱拍の 2・4 拍目に重みが来て") {
		t.Errorf("eight-beat: %s", e)
	}
	if e := explain(Choice{Drums: "ballad"}); !strings.Contains(e, "ハーフタイム") {
		t.Errorf("ballad: %s", e)
	}
	g := grid(4, map[string]string{"kick": "x... .... x.x. ....", "snare": ".... x... .... x...", "hat": "x.x. x.x. x.x. x.x."})
	e := explain(Choice{Grid: &g})
	for _, want := range []string{"バックビート", "16 マスのうち 8 マス", "16 分の格子"} {
		if !strings.Contains(e, want) {
			t.Errorf("grid: want %q in %s", want, e)
		}
	}
}

// TestGridExplanationReview: a ride-only grid is counted by what plays,
// and a snare on 1 and 3 is not called half time.
func TestGridExplanationReview(t *testing.T) {
	ride, err := GridDrums(grid(2, map[string]string{"ride": "x.x. x.x.", "pedal-hat": "..x. ..x."}), music.FourFour, 120)
	if err != nil {
		t.Fatal(err)
	}
	if e := describeDrums(ride, music.FourFour).Explanation; strings.Contains(e, "キック") || strings.Contains(e, " 0 回") {
		t.Errorf("a ride grid names a kick or counts nothing: %s", e)
	}
	oneThree, err := GridDrums(grid(2, map[string]string{"kick": "x... x...", "snare": "x... x..."}), music.FourFour, 120)
	if err != nil {
		t.Fatal(err)
	}
	if e := describeDrums(oneThree, music.FourFour).Explanation; strings.Contains(e, "ハーフタイム") {
		t.Errorf("a snare on 1 and 3 is not half time: %s", e)
	}
}
