package pattern

import (
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

// TestStylesBindTheirPatterns (BL-082, ADR-0012): a style lists the
// patterns and feels that suit it; block chords and no drums suit every
// style. Fits names the misfit so a retry can fix it.
func TestStylesBindTheirPatterns(t *testing.T) {
	tests := []struct {
		style, acc, drums string
		feel              Feel
		misfit            string // "" means it fits
	}{
		{"celtic", "reel", "bodhran", "", ""},
		{"celtic", "strum", "bodhran", "", `accompaniment "strum" does not fit the celtic style`},
		{"celtic", "reel", "eight-beat", "", `drums "eight-beat" does not fit the celtic style`},
		{"jazz", "comp", "jazz-ride", FeelSwing, ""},
		{"jazz", "comp", "eight-beat", FeelSwing, `drums "eight-beat" does not fit the jazz style`},
		{"bossa", "arpeggio", "bossa", FeelSwing, `feel "swing" does not fit the bossa style`},
		{"rock", "strum", "eight-beat", "", ""},
		{"ballad", "", "", "", ""}, // block chords, no drums
		{"pop", "block", "", "", ""},
	}
	for _, tt := range tests {
		st, err := FindStyle(tt.style)
		if err != nil {
			t.Fatal(err)
		}
		problems := st.Misfits(tt.acc, tt.drums, tt.feel)
		got := strings.Join(problems, "; ")
		switch {
		case tt.misfit == "" && got != "":
			t.Errorf("%s %s/%s: should fit, got %s", tt.style, tt.acc, tt.drums, got)
		case tt.misfit != "" && !strings.Contains(got, tt.misfit):
			t.Errorf("%s %s/%s: want %q, got %q", tt.style, tt.acc, tt.drums, tt.misfit, got)
		}
	}
	if _, err := FindStyle("polka"); err == nil {
		t.Error("an unknown style is an error")
	}
}

// TestEveryStyledPatternExists keeps the style catalogue honest: every id
// a style lists is in the pattern catalogue, and every pattern belongs to
// at least one style.
func TestEveryStyledPatternExists(t *testing.T) {
	used := map[string]bool{}
	for _, st := range Styles() {
		for _, id := range st.Accompaniments {
			if _, err := FindAccompaniment(id); err != nil {
				t.Errorf("%s lists unknown accompaniment %q", st.ID, id)
			}
			used["acc:"+id] = true
		}
		for _, id := range st.Drums {
			if _, err := FindDrums(id); err != nil {
				t.Errorf("%s lists unknown drums %q", st.ID, id)
			}
			used["drums:"+id] = true
		}
	}
	for _, a := range Accompaniments() {
		if a.ID != "block" && !used["acc:"+a.ID] {
			t.Errorf("accompaniment %q belongs to no style", a.ID)
		}
	}
	for _, d := range DrumPatterns() {
		if !used["drums:"+d.ID] {
			t.Errorf("drums %q belong to no style", d.ID)
		}
	}
}

// TestStyleSetsTheInstrument (BL-110): a style with an instrument plays a
// keyboard-default pattern on it (Celtic arpeggios on the harp, bossa on a
// nylon guitar); a pattern with its own instrument keeps it, a style
// without one changes nothing, and the timeline records the style.
func TestStyleSetsTheInstrument(t *testing.T) {
	tl := progression(t, music.Key{Tonic: music.D, Mode: music.Dorian}, 4, 1, 7)
	tests := []struct {
		style, acc string
		program    int
	}{
		{"celtic", "arpeggio", 46}, // harp
		{"celtic", "reel", 25},     // the reel keeps its steel-string guitar
		{"celtic", "", 46},         // block chords too
		{"bossa", "arpeggio", 24},  // nylon guitar
		{"jazz", "comp", 0},        // jazz stays on the piano
		{"", "arpeggio", 0},
	}
	for _, tt := range tests {
		got, err := Render(tl, Choice{Accompaniment: tt.acc, Style: tt.style})
		if err != nil {
			t.Fatalf("%s/%s: %v", tt.style, tt.acc, err)
		}
		if p := got.Tracks[0].Program; p != tt.program {
			t.Errorf("%s/%s: program %d, want %d", tt.style, tt.acc, p, tt.program)
		}
		if tt.style != "" && (got.Rhythm == nil || got.Rhythm.Style == nil || got.Rhythm.Style.ID != tt.style) {
			t.Errorf("%s/%s: the timeline should record the style: %+v", tt.style, tt.acc, got.Rhythm)
		}
	}
	if _, err := Render(tl, Choice{Accompaniment: "arpeggio", Style: "polka"}); err == nil {
		t.Error("an unknown style is an error")
	}
}

// TestStyleGrooves (BL-087): choosing another style for the same
// progression plays the style's own groove, the first of its patterns and
// feels that the meter takes. Every groove fits its style and renders.
func TestStyleGrooves(t *testing.T) {
	tests := []struct {
		style string
		ts    music.TimeSignature
		want  Groove
	}{
		{"celtic", music.FourFour, Groove{Meter: music.FourFour, Accompaniment: "reel", Drums: "bodhran", Feel: FeelStraight}},
		{"celtic", music.SixEight, Groove{Meter: music.SixEight, Accompaniment: "rocking-6", Drums: "jig", Feel: FeelStraight}},
		{"jazz", music.FourFour, Groove{Meter: music.FourFour, Accompaniment: "comp", Drums: "jazz-ride", Feel: FeelSwing}},
		{"jazz", music.ThreeFour, Groove{Meter: music.ThreeFour, Accompaniment: "waltz", Drums: "jazz-waltz", Feel: FeelSwing}},
		// Swing is for simple meters: jazz in 6/8 falls back to straight.
		{"jazz", music.SixEight, Groove{Meter: music.SixEight, Accompaniment: "block", Feel: FeelStraight}},
		// Nothing of rock's is written for 3/4: block chords, no drums.
		{"rock", music.ThreeFour, Groove{Meter: music.ThreeFour, Accompaniment: "block", Feel: FeelStraight}},
		{"pop", music.FourFour, Groove{Meter: music.FourFour, Accompaniment: "pulse", Drums: "eight-beat", Feel: FeelStraight}},
	}
	for _, tt := range tests {
		st, err := FindStyle(tt.style)
		if err != nil {
			t.Fatal(err)
		}
		if got := st.Groove(tt.ts); got != tt.want {
			t.Errorf("%s in %d/%d: got %+v, want %+v", tt.style, tt.ts.Numerator, tt.ts.Denominator, got, tt.want)
		}
	}

	for _, st := range Styles() {
		if len(st.Grooves) != len(GrooveMeters) {
			t.Errorf("%s lists %d grooves, want one per meter", st.ID, len(st.Grooves))
		}
		for _, g := range st.Grooves {
			if m := st.Misfits(g.Accompaniment, g.Drums, g.Feel); len(m) > 0 {
				t.Errorf("%s groove misfits: %v", st.ID, m)
			}
			tl := blockTimeline(t, g.Meter)
			if _, err := Render(tl, g.Choice(st.ID)); err != nil {
				t.Errorf("%s in %d/%d: %v", st.ID, g.Meter.Numerator, g.Meter.Denominator, err)
			}
		}
	}
}

// blockTimeline is ii–V–I–I in C, a bar each, in the meter.
func blockTimeline(t *testing.T, ts music.TimeSignature) music.Timeline {
	t.Helper()
	var prog music.Progression
	for _, d := range []int{2, 5, 1, 1} {
		prog = append(prog, music.Step{Degree: d, Beats: ts.BeatsPerBar()})
	}
	tl, err := music.FromProgressionIn(ts, music.Key{Tonic: music.C, Mode: music.Major}, prog, 100)
	if err != nil {
		t.Fatal(err)
	}
	return tl
}
