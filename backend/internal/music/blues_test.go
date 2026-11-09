package music

import (
	"strings"
	"testing"
)

// The blues writes I7 and IV7 in a major key: dominant sevenths that keep
// their functions (BL-135, ADR-0026).
func TestBluesDominantsSoundAsDominantSevenths(t *testing.T) {
	c := Key{C, Major}
	tests := []struct {
		label string
		want  Chord
	}{
		{"I7", Chord{Root: C, Quality: Dominant7}},
		{"IV7", Chord{Root: F, Quality: Dominant7}},
		{"Imaj7", Chord{Root: C, Quality: Major7}}, // the key's own seventh, as before
		{"V7", Chord{Root: G, Quality: Dominant7}},
	}
	for _, tt := range tests {
		got, err := CanonicalLabel(c, tt.label)
		if err != nil || got != tt.label {
			t.Errorf("CanonicalLabel(%q) = %q, %v; want %q", tt.label, got, err, tt.label)
		}
		step, err := StepFromLabel(c, tt.label, 4)
		if err != nil {
			t.Fatal(err)
		}
		tl, err := FromProgression(c, Progression{step}, 90)
		if err != nil {
			t.Fatal(err)
		}
		if tl.Harmony[0].Chord != tt.want {
			t.Errorf("%q sounds as %+v, want %+v", tt.label, tl.Harmony[0].Chord, tt.want)
		}
	}
}

func TestBluesDominantsAreNotSecondaryDominants(t *testing.T) {
	c := Key{C, Major}
	tl := analysed(t, c, 4, "I7", "IV7", "I7", "V7", "IV7", "I7")
	for i, want := range []Function{Tonic, Subdominant, Tonic, Dominant, Subdominant, Tonic} {
		a := tl.Harmony[i].Analysis
		if a.Function != want || a.AppliedTo != 0 {
			t.Errorf("harmony[%d] %s: function %s, applied to %d; want %s, not applied", i, tl.Harmony[i].RomanNumeral, a.Function, a.AppliedTo, want)
		}
	}
	first, second := tl.Harmony[0].Analysis.Explanation, tl.Harmony[1].Analysis.Explanation
	if !strings.Contains(first, "ブルースの属七") || !strings.Contains(first, "トニックのまま") || !strings.Contains(first, "Bb") {
		t.Errorf("I7: %q", first)
	}
	if !strings.Contains(second, "ブルースの属七") || !strings.Contains(second, "ブルー・ノート") || !strings.Contains(second, "Eb") {
		t.Errorf("IV7: %q", second)
	}
	if got := tl.Harmony[5].Analysis.Cadence; got != Plagal {
		t.Errorf("IV7 → I7: cadence %q, want plagal", got)
	}
	if got := tl.Harmony[4].Analysis.Cadence; got != NoCadence {
		t.Errorf("V7 → IV7: cadence %q, want none", got)
	}
}

// An I7 must not be said to stay put when the next chord is IV: its
// seventh falls to IV's third, as in the Marunouchi Sadistic loop
// (IVmaj7 – V7/vi – vi7 – I7 – IVmaj7, #272, ADR-0034). Elsewhere it is
// still the blues dominant that does not resolve.
func TestBluesI7SaysWhatItDoesBeforeIV(t *testing.T) {
	c := Key{C, Major}
	tests := []struct {
		name    string
		labels  []string
		at      int
		want    []string
		notWant []string
	}{
		{"into IVmaj7", []string{"IVmaj7", "V7/vi", "vi7", "I7", "IVmaj7"}, 3,
			[]string{"ブルースの属七", "トニックのまま", "次が IV", "Bb が IV の 3 度の A へ半音で下り"}, []string{"IV へ向かう働きではない"}},
		{"into the blues IV7", []string{"I7", "IV7", "I7"}, 0,
			[]string{"次が IV", "IV の 3 度の A"}, []string{"IV へ向かう働きではない"}},
		{"into V7", []string{"I7", "IV7", "I7", "V7"}, 2,
			[]string{"解決を求めずトニックのまま鳴る", "IV へ向かう働きではない"}, []string{"次が IV"}},
		{"last chord", []string{"IV7", "V7", "I7"}, 2,
			[]string{"IV へ向かう働きではない"}, []string{"次が IV"}},
		{"into the borrowed iv", []string{"I7", "iv", "I"}, 0,
			[]string{"IV へ向かう働きではない"}, []string{"次が IV"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tl := analysed(t, c, 4, tt.labels...)
			a := tl.Harmony[tt.at].Analysis
			if a.Function != Tonic || a.AppliedTo != 0 {
				t.Errorf("%s: function %s, applied to %d; want tonic, not applied", tt.labels[tt.at], a.Function, a.AppliedTo)
			}
			for _, w := range tt.want {
				if !strings.Contains(a.Explanation, w) {
					t.Errorf("explanation lacks %q: %q", w, a.Explanation)
				}
			}
			for _, w := range tt.notWant {
				if strings.Contains(a.Explanation, w) {
					t.Errorf("explanation says %q: %q", w, a.Explanation)
				}
			}
		})
	}
}

// The same sound written V7/IV is still the dominant of IV.
func TestV7OfIVIsStillASecondaryDominant(t *testing.T) {
	c := Key{C, Major}
	tl := analysed(t, c, 4, "I", "V7/IV", "IV")
	a := tl.Harmony[1].Analysis
	if a.AppliedTo != 4 || a.Function != Dominant || strings.Contains(a.Explanation, "ブルース") {
		t.Errorf("V7/IV: %+v", a)
	}
	if tl.Harmony[1].Chord != (Chord{Root: C, Quality: Dominant7}) {
		t.Errorf("V7/IV sounds as %+v, want C7", tl.Harmony[1].Chord)
	}
}

// Only I7 and IV7 in major: other degrees and other modes still refuse a
// seventh the key would change.
func TestBluesDominantsOnlyOnIAndIVInMajor(t *testing.T) {
	refused := []struct {
		key   Key
		label string
	}{
		{Key{C, Major}, "II7"},   // the dominant of V: write V7/V
		{Key{C, Major}, "Vmaj7"}, // the key's V7 is a dominant seventh
		{Key{C, Minor}, "I7"},    // minor keys are unchanged
		{Key{C, Minor}, "IV7"},   // the dorian IV7 is not borrowed into minor
		{Key{C, Dorian}, "I7"},   // nor into the modes
		{Key{C, Major}, "i7"},    // lowercase is the minor seventh, borrowed or not, not the blues chord
	}
	for _, tt := range refused {
		if got, err := CanonicalLabel(tt.key, tt.label); err == nil {
			t.Errorf("%q in %s was accepted as %q", tt.label, tt.key, got)
		}
	}
}
