package pattern

import (
	"strings"
	"testing"
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
