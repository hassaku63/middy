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
