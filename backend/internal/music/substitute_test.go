package music

import (
	"strings"
	"testing"
)

// TestTritoneSubstitutes (BL-128, ADR-0023): subV7/x is the dominant
// seventh a semitone above x, spelled on the letter a second above, and may
// aim at the tonic.
func TestTritoneSubstitutes(t *testing.T) {
	tests := []struct {
		key    Key
		label  string
		symbol string
		root   PitchClass
	}{
		{Key{C, Major}, "subV7/I", "Db7", Cs},
		{Key{C, Major}, "subV7/V", "Ab7", Gs},
		{Key{C, Major}, "subV7/ii", "Eb7", Ds},
		{Key{C, Major}, "subV7/vi", "Bb7", As},
		{Key{E, Major}, "subV7/I", "F7", F},
		{Key{D, Major}, "subV7/vi", "C7", C},
		{Key{A, Minor}, "subV7/i", "Bb7", As},
		{Key{A, Minor}, "subV7/III", "Db7", Cs},
	}
	for _, tt := range tests {
		t.Run(tt.key.String()+" "+tt.label, func(t *testing.T) {
			step, err := StepFromLabel(tt.key, tt.label, 4)
			if err != nil {
				t.Fatal(err)
			}
			chord, numeral, symbol, err := stepChord(tt.key, step)
			if err != nil {
				t.Fatal(err)
			}
			if chord.Root != tt.root || chord.Quality != Dominant7 || numeral != tt.label || symbol != tt.symbol {
				t.Errorf("got %v %q %q, want %v dom7 %q %q", chord, numeral, symbol, tt.root, tt.label, tt.symbol)
			}
			if canon, err := CanonicalLabel(tt.key, tt.label); err != nil || canon != tt.label {
				t.Errorf("CanonicalLabel = %q, %v", canon, err)
			}
		})
	}
}

func TestTritoneSubstitutesAreLimitedToRealOnes(t *testing.T) {
	tests := []struct {
		key   Key
		label string
		want  string
	}{
		{Key{C, Major}, "subV7/vii°", "not a major or minor chord"},
		{Key{A, Minor}, "subV7/ii°", "not a major or minor chord"},
		{Key{C, Major}, "subV/I", "only V, V7 and subV7"},
		{Key{C, Major}, "subV7/II", "write the target as \"ii\""},
	}
	for _, tt := range tests {
		if _, err := StepFromLabel(tt.key, tt.label, 4); err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s %q: error %v, want it to mention %q", tt.key, tt.label, err, tt.want)
		}
	}
}

// TestTritoneSubstitutesAreAnalysedTowardsTheirTarget: a substitute is a
// dominant that pulls to its target; the explanation names the V7 it
// replaces, and resolves or deflects like V7/x. The label decides between a
// substitute and a borrowed chord of the same sound (Bb7 in C major).
func TestTritoneSubstitutesAreAnalysedTowardsTheirTarget(t *testing.T) {
	c := Key{C, Major}
	tl := analysedLabels(t, c, "ii7", "subV7/I", "Imaj7", "subV7/V", "vi")
	sub := tl.Harmony[1].Analysis
	if sub.AppliedTo != 1 || !sub.Substitute || sub.Function != Dominant {
		t.Fatalf("subV7/I analysis = %+v", sub)
	}
	for _, want := range []string{"裏コード", "V7（G7）の代わり", "次の I へ解決する"} {
		if !strings.Contains(sub.Explanation, want) {
			t.Errorf("subV7/I explanation lacks %q: %s", want, sub.Explanation)
		}
	}
	if e := tl.Harmony[3].Analysis.Explanation; !strings.Contains(e, "V7/V（D7）の代わり") || !strings.Contains(e, "かわす") {
		t.Errorf("subV7/V before vi should name V7/V and deflect: %s", e)
	}

	borrowed := analysedLabels(t, c, "I", "bVII7", "I")
	if a := borrowed.Harmony[1].Analysis; !a.Borrowed || a.Substitute {
		t.Errorf("bVII7 is the borrowed chord: %+v", a)
	}
	substitute := analysedLabels(t, c, "I", "subV7/vi", "vi")
	if a := substitute.Harmony[1].Analysis; a.Borrowed || !a.Substitute || a.AppliedTo != 6 {
		t.Errorf("subV7/vi is the substitute: %+v", a)
	}
}
