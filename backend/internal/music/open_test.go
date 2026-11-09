package music

import (
	"strings"
	"testing"
)

// TestOpenChords (BL-132, ADR-0025): sus4, 7sus4 and the power chord stand
// on a degree whose notes are all in the key; the case does not matter and
// is spelled uppercase; a power chord may also stand on the parallel
// minor's lowered roots in a major key.
func TestOpenChords(t *testing.T) {
	c := Key{C, Major}
	tests := []struct {
		key     Key
		label   string
		canon   string
		symbol  string
		quality ChordQuality
	}{
		{c, "Vsus4", "Vsus4", "Gsus4", Sus4},
		{c, "V7sus4", "V7sus4", "G7sus4", Dominant7Sus4},
		{c, "iisus4", "IIsus4", "Dsus4", Sus4},
		{c, "I5", "I5", "C5", Power},
		{c, "bVII5", "♭VII5", "Bb5", Power},
		{c, "bVI5", "♭VI5", "Ab5", Power},
		{Key{E, Minor}, "i5", "I5", "E5", Power},
	}
	for _, tt := range tests {
		t.Run(tt.key.String()+" "+tt.label, func(t *testing.T) {
			step, err := StepFromLabel(tt.key, tt.label, 4)
			if err != nil {
				t.Fatal(err)
			}
			chord, _, symbol, err := stepChord(tt.key, step)
			if err != nil {
				t.Fatal(err)
			}
			if chord.Quality != tt.quality || symbol != tt.symbol {
				t.Errorf("got %v %q, want %v %q", chord.Quality, symbol, tt.quality, tt.symbol)
			}
			if canon, err := CanonicalLabel(tt.key, tt.label); err != nil || canon != tt.canon {
				t.Errorf("CanonicalLabel = %q, %v; want %q", canon, err, tt.canon)
			}
		})
	}
	for _, bad := range []string{"IVsus4", "vii°5", "bVIIsus4", "#IV5"} {
		if _, err := StepFromLabel(c, bad, 4); err == nil {
			t.Errorf("%q accepted in C major", bad)
		}
	}
}

// TestOpenChordsAreExplained: a sus4 that resolves to the same root says
// so; a power chord explains the missing third and keeps its degree's
// cadences.
func TestOpenChordsAreExplained(t *testing.T) {
	c := Key{C, Major}
	tl := analysedLabels(t, c, "I", "Vsus4", "V", "I5")
	sus := tl.Harmony[1].Analysis
	if sus.Degree != 5 || !strings.Contains(sus.Explanation, "掛留（sus4）") || !strings.Contains(sus.Explanation, "4 度が 3 度へ下がって解決する") {
		t.Errorf("Vsus4 – V: %+v", sus)
	}
	if e := tl.Harmony[3].Analysis.Explanation; !strings.Contains(e, "パワーコード") {
		t.Errorf("I5: %s", e)
	}
	// V – I5 still lands on the tonic's root, the leading tone rising to it.
	if got := tl.Harmony[3].Analysis.Cadence; got != Authentic {
		t.Errorf("V – I5 cadence = %q, want authentic", got)
	}
	rock := analysedLabels(t, c, "I5", "bVII5", "IV5", "I5")
	if a := rock.Harmony[1].Analysis; !a.Borrowed || a.Degree != 7 {
		t.Errorf("bVII5 is borrowed from the parallel minor: %+v", a)
	}
}
