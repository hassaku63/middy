package music

import (
	"strings"
	"testing"
)

func TestAppliedDominantsReadSoundAndSpell(t *testing.T) {
	tests := []struct {
		key     Key
		label   string
		numeral string
		symbol  string
		chord   Chord
	}{
		{Key{C, Major}, "V7/V", "V7/V", "D7", Chord{Root: D, Quality: Dominant7}},
		{Key{C, Major}, "V/V", "V/V", "D", Chord{Root: D, Quality: MajorTriad}},
		{Key{C, Major}, "V7/ii", "V7/ii", "A7", Chord{Root: A, Quality: Dominant7}},
		{Key{C, Major}, "V/iii", "V/iii", "B", Chord{Root: B, Quality: MajorTriad}},
		{Key{C, Major}, "V7/IV", "V7/IV", "C7", Chord{Root: C, Quality: Dominant7}},
		{Key{C, Major}, "V7/vi", "V7/vi", "E7", Chord{Root: E, Quality: Dominant7}},
		{Key{F, Major}, "V7/vi", "V7/vi", "A7", Chord{Root: A, Quality: Dominant7}},
		{Key{As, Major}, "V7/ii", "V7/ii", "G7", Chord{Root: G, Quality: Dominant7}},
		{Key{Cs, Major}, "V/V", "V/V", "D#", Chord{Root: Ds, Quality: MajorTriad}}, // Db major: V of Ab is Eb
		{Key{C, Minor}, "V7/iv", "V7/iv", "C7", Chord{Root: C, Quality: Dominant7}},
		{Key{C, Minor}, "V/v", "V/v", "D", Chord{Root: D, Quality: MajorTriad}},
		{Key{C, Minor}, "V/VII", "V/VII", "F", Chord{Root: F, Quality: MajorTriad}},
		{Key{C, Minor}, "V7/VI", "V7/VI", "Eb7", Chord{Root: Ds, Quality: Dominant7}},
		{Key{Cs, Minor}, "V/v", "V/v", "D#", Chord{Root: Ds, Quality: MajorTriad}},
		{Key{D, Dorian}, "V/v", "V/v", "E", Chord{Root: E, Quality: MajorTriad}},
		{Key{G, Mixolydian}, "V7/ii", "V7/ii", "E7", Chord{Root: E, Quality: Dominant7}},
	}
	for _, tt := range tests {
		step, err := StepFromLabel(tt.key, tt.label, 4)
		if err != nil {
			t.Fatalf("%s %q: %v", tt.key, tt.label, err)
		}
		tl, err := FromProgression(tt.key, Progression{step}, 100)
		if err != nil {
			t.Fatalf("%s %q: %v", tt.key, tt.label, err)
		}
		h := tl.Harmony[0]
		if h.Chord != tt.chord || h.RomanNumeral != tt.numeral {
			t.Errorf("%s %q: got %v %q, want %v %q", tt.key, tt.label, h.Chord, h.RomanNumeral, tt.chord, tt.numeral)
		}
		if tt.key.Tonic == Cs && tt.key.Mode == Major {
			continue // Db major spells Eb; checked below with its flat spelling
		}
		if h.Symbol != tt.symbol {
			t.Errorf("%s %q: symbol %q, want %q", tt.key, tt.label, h.Symbol, tt.symbol)
		}
	}
	// Db major (written with flats): V of Ab is Eb.
	step, _ := StepFromLabel(Key{Cs, Major}, "V/V", 4)
	tl, _ := FromProgression(Key{Cs, Major}, Progression{step}, 100)
	if tl.Harmony[0].Symbol != "Eb" {
		t.Errorf("Db major V/V = %q, want Eb", tl.Harmony[0].Symbol)
	}
}

func TestAppliedDominantsAreLimitedToRealOnes(t *testing.T) {
	tests := []struct {
		key   Key
		label string
		want  string
	}{
		{Key{C, Major}, "V/vii°", "is not a major or minor chord"},
		{Key{C, Minor}, "V/ii°", "is not a major or minor chord"},
		{Key{D, Dorian}, "V/vi°", "is not a major or minor chord"},
		{Key{C, Major}, "V/I", "is the dominant itself"},
		{Key{C, Major}, "V/IV", "C major's own I (C)"},         // the tonic triad
		{Key{C, Minor}, "V/III", "C minor's own VII (Bb)"},     // natural minor's VII
		{Key{C, Minor}, "V7/III", "C minor's own VII7"},        // and its seventh
		{Key{G, Mixolydian}, "V7/IV", "G mixolydian's own I7"}, // the mixolydian tonic seventh
		{Key{C, Major}, "V/II", `write the target as "ii"`},
		{Key{C, Major}, "V/ii7", "write the target as the key's plain triad"},
		{Key{C, Major}, "IV/V", "only V, V7 and subV7 can be applied"},
		{Key{C, Major}, "V/X", "not a roman numeral"},
	}
	for _, tt := range tests {
		_, err := StepFromLabel(tt.key, tt.label, 4)
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s %q: error %v, want it to mention %q", tt.key, tt.label, err, tt.want)
		}
		if _, err := CanonicalLabel(tt.key, tt.label); err == nil {
			t.Errorf("%s %q: CanonicalLabel must reject it too", tt.key, tt.label)
		}
	}
	if got, err := CanonicalLabel(Key{C, Major}, " V7/V "); err != nil || got != "V7/V" {
		t.Errorf("CanonicalLabel(V7/V) = %q, %v", got, err)
	}
}

func TestAppliedDominantsAreAnalysedTowardsTheirTarget(t *testing.T) {
	tl := analysed(t, Key{C, Major}, 4, "I", "V7/V", "V", "I")
	a := tl.Harmony[1].Analysis
	if a.Function != Dominant || a.AppliedTo != 5 || a.Degree != 2 || a.Cadence != NoCadence {
		t.Fatalf("V7/V analysis: %+v", a)
	}
	for _, want := range []string{"V7/V（D7）は、V（G）へ向かうセカンダリードミナント", "次の V へ解決する"} {
		if !strings.Contains(a.Explanation, want) {
			t.Errorf("explanation lacks %q: %s", want, a.Explanation)
		}
	}
	// A deceptive turn: V7/vi going somewhere else.
	tl = analysed(t, Key{C, Major}, 4, "I", "V7/vi", "IV", "I")
	if !strings.Contains(tl.Harmony[1].Analysis.Explanation, "ここでは vi へは進まず") {
		t.Errorf("deflected V7/vi: %s", tl.Harmony[1].Analysis.Explanation)
	}
	// Minor: C7 pulls towards iv.
	tl = analysed(t, Key{C, Minor}, 4, "i", "V7/iv", "iv", "V7", "i")
	if a := tl.Harmony[1].Analysis; a.AppliedTo != 4 || !strings.Contains(a.Explanation, "iv（Fm）へ向かう") {
		t.Errorf("minor V7/iv: %+v", a)
	}
	// The chord a step above keeps its ordinary analysis.
	if tl.Harmony[3].Analysis.AppliedTo != 0 || tl.Harmony[4].Analysis.Cadence != Authentic {
		t.Errorf("V7 - i after the applied chord: %+v %+v", tl.Harmony[3].Analysis, tl.Harmony[4].Analysis)
	}
}
