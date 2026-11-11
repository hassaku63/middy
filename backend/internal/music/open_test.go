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
		{c, "Isus2", "Isus2", "Csus2", Sus2},
		{c, "vsus2", "Vsus2", "Gsus2", Sus2},
		{Key{A, Minor}, "isus2", "Isus2", "Asus2", Sus2},
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
	// C major's IIIsus2 needs F#, its VIIsus2 C# and F#.
	for _, bad := range []string{"IVsus4", "vii°5", "bVIIsus4", "#IV5", "IIIsus2", "VIIsus2"} {
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
	// A sus2 that moves to the same root's triad says the second rises to
	// the third; one that does not only names the second (BL-309).
	open := analysedLabels(t, c, "Isus2", "I", "IVsus2", "V")
	if e := open.Harmony[0].Analysis.Explanation; !strings.Contains(e, "2 度を鳴らす（sus2）") || !strings.Contains(e, "2 度が 3 度へ上がり") {
		t.Errorf("Isus2 – I: %s", e)
	}
	if e := open.Harmony[2].Analysis.Explanation; !strings.Contains(e, "（sus2）") || strings.Contains(e, "3 度へ上がり") {
		t.Errorf("IVsus2 – V: %s", e)
	}
	if got := open.Harmony[0].Symbol; got != "Csus2" {
		t.Errorf("Isus2 symbol = %q", got)
	}
	rock := analysedLabels(t, c, "I5", "bVII5", "IV5", "I5")
	if a := rock.Harmony[1].Analysis; !a.Borrowed || a.Degree != 7 {
		t.Errorf("bVII5 is borrowed from the parallel minor: %+v", a)
	}
}

// A flat that names the key's own degree is that degree (#109): D
// mixolydian's bVII5 is its VII5 (C5), A minor's bVI5 its VI5 (F5).
func TestOpenChordsOnAFlatThatIsTheKeysOwnDegree(t *testing.T) {
	tests := []struct {
		key   Key
		label string
		want  string
		root  PitchClass
	}{
		{Key{D, Mixolydian}, "bVII5", "VII5", C},
		{Key{A, Minor}, "bVI5", "VI5", F},
		{Key{A, Minor}, "bVIIsus4", "VIIsus4", G},
	}
	for _, tt := range tests {
		got, err := CanonicalLabel(tt.key, tt.label)
		if err != nil || got != tt.want {
			t.Errorf("CanonicalLabel(%s, %q) = %q, %v; want %q", tt.key, tt.label, got, err, tt.want)
			continue
		}
		step, err := StepFromLabel(tt.key, tt.label, 4)
		if err != nil {
			t.Fatal(err)
		}
		tl, err := FromProgression(tt.key, Progression{step}, 90)
		if err != nil {
			t.Fatal(err)
		}
		if tl.Harmony[0].Chord.Root != tt.root {
			t.Errorf("%q in %s has root %v, want %v", tt.label, tt.key, tl.Harmony[0].Chord.Root, tt.root)
		}
	}
	if _, err := CanonicalLabel(Key{D, Mixolydian}, "bII5"); err == nil {
		t.Error("bII5 is not D mixolydian's own degree and must be refused")
	}
}

// TestOpenChordSlashBasses: a bass on a chord without a third is named by
// its interval. A sus chord's second or fourth stands in for the third and
// is no first inversion; a power chord's fifth is a second inversion
// (BL-309).
func TestOpenChordSlashBasses(t *testing.T) {
	tests := []struct{ label, want, not string }{
		{"Isus2/2", "3 度の代わりに鳴らす 2 度の D がベース", "転回形"},
		{"Isus4/4", "3 度の代わりに鳴らす 4 度の F がベース", "転回形"},
		{"I5/5", "5 度の G がベース（第 2 転回形）", "3 度の G"},
		{"V7sus4/4", "7 度の F がベース（第 3 転回形）", "3 度の F"},
		{"Isus4/2", "ベースの D はこの和音にない音", "転回形"},
	}
	for _, tt := range tests {
		e := analysedLabels(t, Key{C, Major}, "I", tt.label, "I").Harmony[1].Analysis.Explanation
		if !strings.Contains(e, tt.want) || strings.Contains(e, tt.not) {
			t.Errorf("%s: %s", tt.label, e)
		}
	}
}
