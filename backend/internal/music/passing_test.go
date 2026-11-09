package music

import (
	"strings"
	"testing"
)

// TestLeadingToneDiminishedSevenths (BL-130, ADR-0024): vii°7/x is the
// diminished seventh a semitone below x, spelled on the letter a second
// below.
func TestLeadingToneDiminishedSevenths(t *testing.T) {
	tests := []struct {
		key    Key
		label  string
		symbol string
		root   PitchClass
	}{
		{Key{C, Major}, "vii°7/ii", "C#dim7", Cs},
		{Key{C, Major}, "vii°7/V", "F#dim7", Fs},
		{Key{C, Major}, "vii°7/vi", "G#dim7", Gs},
		{Key{F, Major}, "vii°7/ii", "F#dim7", Fs},
		{Key{A, Minor}, "vii°7/iv", "C#dim7", Cs},
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
			if chord.Root != tt.root || chord.Quality != Diminished7 || numeral != tt.label || symbol != tt.symbol {
				t.Errorf("got %v %q %q, want %v dim7 %q %q", chord, numeral, symbol, tt.root, tt.label, tt.symbol)
			}
			if canon, err := CanonicalLabel(tt.key, strings.Replace(tt.label, "°", "o", 1)); err != nil || canon != tt.label {
				t.Errorf("CanonicalLabel of the o spelling = %q, %v", canon, err)
			}
		})
	}
	for _, bad := range []string{"vii°7/I", "vii°7/vii°"} {
		if _, err := StepFromLabel(Key{C, Major}, bad, 4); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}

// TestPassingDiminishedIsAnalysed: I – vii°7/ii – ii7 climbs a semitone at a
// time, so the diminished seventh passes (so does IV – vii°7/V – V); from
// elsewhere it only leads.
func TestPassingDiminishedIsAnalysed(t *testing.T) {
	c := Key{C, Major}
	tl := analysedLabels(t, c, "I", "vii°7/ii", "ii7", "V7")
	a := tl.Harmony[1].Analysis
	if !a.LeadingTone || a.AppliedTo != 2 || a.Function != Dominant {
		t.Fatalf("vii°7/ii analysis = %+v", a)
	}
	for _, want := range []string{"導音の上に積んだ減七", "経過の減七（パッシング・ディミニッシュ）", "次の ii へ解決する"} {
		if !strings.Contains(a.Explanation, want) {
			t.Errorf("explanation lacks %q: %s", want, a.Explanation)
		}
	}
	// IV – #IV°7 – V climbs F, F#, G: it passes too.
	if e := analysedLabels(t, c, "IV", "vii°7/V", "V7", "I").Harmony[1].Analysis.Explanation; !strings.Contains(e, "パッシング") {
		t.Errorf("IV – vii°7/V – V passes: %s", e)
	}
	// From I (C), F#°7 does not climb into place: it only leads to V.
	if e := analysedLabels(t, c, "I", "vii°7/V", "V7", "I").Harmony[1].Analysis.Explanation; strings.Contains(e, "パッシング") || !strings.Contains(e, "次の V へ解決する") {
		t.Errorf("I – vii°7/V – V leads but does not pass: %s", e)
	}
}

// TestMinorTargetsReadTheHarmonicV (BL-131): in a minor key the uppercase V
// is the harmonic-minor dominant on v's root, so V7/V, subV7/V and
// vii°7/V name the same chords as V7/v, subV7/v and vii°7/v, and are
// spelled that way.
func TestMinorTargetsReadTheHarmonicV(t *testing.T) {
	am := Key{A, Minor}
	for label, want := range map[string]string{"V7/V": "V7/v", "subV7/V": "subV7/v", "vii°7/V": "vii°7/v", "V/V": "V/v"} {
		canon, err := CanonicalLabel(am, label)
		if err != nil || canon != want {
			t.Errorf("%s: CanonicalLabel = %q, %v; want %q", label, canon, err, want)
		}
	}
	if _, err := CanonicalLabel(Key{C, Major}, "V7/v"); err == nil {
		t.Error("a major key's v is still the wrong case")
	}
}
