package music

import (
	"strings"
	"testing"
)

func TestParseRomanNumeral(t *testing.T) {
	tests := []struct {
		in      string
		degree  int
		seventh bool
	}{
		{"I", 1, false}, {"ii", 2, false}, {"III", 3, false}, {"iv", 4, false},
		{"V", 5, false}, {"vi", 6, false}, {"VII", 7, false},
		{"V7", 5, true}, {"ii7", 2, true}, {"IVmaj7", 4, true}, {"Imaj7", 1, true},
		{"viiø7", 7, true}, {"viim7b5", 7, true}, {"vii°", 7, false}, {"viidim", 7, false},
		{"III+", 3, false}, {"IIIaug", 3, false},
		{" v ", 5, false}, {"vIi", 7, false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			d, s, err := ParseRomanNumeral(tt.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d != tt.degree || s != tt.seventh {
				t.Fatalf("got degree=%d seventh=%v, want degree=%d seventh=%v", d, s, tt.degree, tt.seventh)
			}
		})
	}
}

func TestParseRomanNumeralRejectsGarbage(t *testing.T) {
	for _, in := range []string{"", "VIII", "C", "1", "I-V", "7", "iiii", "V9", "bb", "b"} {
		t.Run(in, func(t *testing.T) {
			if _, _, err := ParseRomanNumeral(in); err == nil {
				t.Fatalf("%q should not parse", in)
			}
		})
	}
}

func TestParseRomanNumeralRoundTripsWithRomanNumeral(t *testing.T) {
	// Every label the engine prints must parse back to the same degree/seventh.
	for _, key := range []Key{{C, Major}, {A, Minor}} {
		for d := 1; d <= 7; d++ {
			for _, seventh := range []bool{false, true} {
				ch, err := key.Diatonic(d, seventh)
				if err != nil {
					t.Fatal(err)
				}
				label, err := RomanNumeral(d, ch.Quality)
				if err != nil {
					t.Fatal(err)
				}
				gotD, gotS, err := ParseRomanNumeral(label)
				if err != nil || gotD != d || gotS != seventh {
					t.Errorf("%s: ParseRomanNumeral(%q) = %d,%v,%v; want %d,%v", key, label, gotD, gotS, err, d, seventh)
				}
			}
		}
	}
}

func TestStepFromLabelHarmonicMinorDominant(t *testing.T) {
	minor := Key{A, Minor}
	major := Key{C, Major}
	tests := []struct {
		key     Key
		label   string
		wantQ   *ChordQuality
		wantDeg int
	}{
		{minor, "V", ptr(MajorTriad), 5},
		{minor, "V7", ptr(Dominant7), 5},
		{minor, "v", nil, 5},
		{minor, "v7", nil, 5},
		{minor, "IV", ptr(MajorTriad), 4}, // borrowed from the parallel major (BL-075)
		{major, "V", nil, 5},              // already major; no override needed
		{major, "v", nil, 5},              // case is ignored outside minor V
	}
	for _, tt := range tests {
		t.Run(tt.key.String()+" "+tt.label, func(t *testing.T) {
			step, err := StepFromLabel(tt.key, tt.label, 4)
			if err != nil {
				t.Fatal(err)
			}
			if step.Degree != tt.wantDeg || step.Beats != 4 {
				t.Fatalf("step = %+v", step)
			}
			if (step.Quality == nil) != (tt.wantQ == nil) || (tt.wantQ != nil && *step.Quality != *tt.wantQ) {
				t.Fatalf("quality = %v, want %v", step.Quality, tt.wantQ)
			}
		})
	}
	if _, err := StepFromLabel(minor, "X", 4); err == nil {
		t.Fatal("bad label must fail")
	}
}

func TestHarmonicMinorDominantRendersAndAnalyses(t *testing.T) {
	key := Key{A, Minor}
	var prog Progression
	for _, l := range []string{"i", "iv", "V7", "i"} {
		s, err := StepFromLabel(key, l, 4)
		if err != nil {
			t.Fatal(err)
		}
		prog = append(prog, s)
	}
	tl, err := FromProgression(key, prog, 100)
	if err != nil {
		t.Fatal(err)
	}
	if got := tl.Harmony[2].Chord.String(); got != "E7" {
		t.Fatalf("V7 in A minor = %q, want E7", got)
	}
	if got := tl.Harmony[2].RomanNumeral; got != "V7" {
		t.Fatalf("numeral = %q, want V7", got)
	}
	// G# (the raised leading tone) must actually sound.
	var hasGs bool
	for _, n := range tl.Tracks[0].Notes {
		if n.Pitch.Class() == Gs {
			hasGs = true
		}
	}
	if !hasGs {
		t.Fatal("expected the raised leading tone G# in the voicing")
	}
	analysed, err := Analyze(tl)
	if err != nil {
		t.Fatalf("Analyze must accept the harmonic-minor dominant: %v", err)
	}
	a := analysed.Harmony[2].Analysis
	if a.Function != Dominant || analysed.Harmony[3].Analysis.Cadence != Authentic {
		t.Fatalf("V7 -> i should be a dominant resolving in an authentic cadence: %+v / %+v", a, analysed.Harmony[3].Analysis)
	}
	if !strings.Contains(a.Explanation, "和声的短音階") {
		t.Fatalf("explanation should mention harmonic minor: %q", a.Explanation)
	}
	if strings.Contains(analysed.Harmony[1].Analysis.Explanation, "和声的短音階") {
		t.Fatal("iv must not get the harmonic-minor note")
	}
}

func ptr(q ChordQuality) *ChordQuality { return &q }

func TestCanonicalLabelRejectsWhatTheKeyWouldSilentlyChange(t *testing.T) {
	dMinor := Key{Tonic: D, Mode: Minor}
	dMajor := Key{Tonic: D, Mode: Major}
	bbMajor := Key{Tonic: As, Mode: Major}
	tests := []struct {
		key     Key
		label   string
		want    string // canonical label when accepted
		wantErr string // substring of the error when rejected
	}{
		{dMinor, "i", "i", ""},
		{dMinor, "V", "V", ""},   // harmonic-minor dominant, as the prompt allows
		{dMinor, "V7", "V7", ""}, // its seventh
		{dMinor, "v", "v", ""},   // natural-minor triad
		{dMinor, "iiø7", "iiø7", ""},
		{dMinor, "iim7b5", "iiø7", ""}, // same chord, the engine's spelling wins
		{dMinor, "VII7", "VII7", ""},
		{dMinor, "II", "", `"II" would sound as Edim (ii°) in D minor`}, // no major II in minor
		{dMinor, "VIImaj7", "", `"VIImaj7" would sound as C7 (VII7) in D minor`},
		{dMajor, "VII", "", `"VII" would sound as C#dim (vii°) in D major`}, // the mixolydian VII
		{dMajor, "vii°", "vii°", ""},
		{dMajor, "vii", "vii°", ""}, // lowercase alone is close enough for a diminished degree
		{dMajor, "v", "", `"v" would sound as A (V) in D major`},
		{bbMajor, "I7", "I7", ""}, // the blues dominant (BL-135)
		{bbMajor, "Vmaj7", "", `"Vmaj7" would sound as F7 (V7) in Bb major`},
		{bbMajor, "Imaj7", "Imaj7", ""},
		{bbMajor, "ii7", "ii7", ""},
		{bbMajor, "bVII", "♭VII", ""}, // borrowed from Bb minor (BL-094)
		{bbMajor, "#iv", "", "is not available"},
	}
	for _, tt := range tests {
		got, err := CanonicalLabel(tt.key, tt.label)
		switch {
		case tt.wantErr != "":
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("%s %q: want error containing %q, got %q / %v", tt.key, tt.label, tt.wantErr, got, err)
			}
		case err != nil:
			t.Errorf("%s %q: unexpected error %v", tt.key, tt.label, err)
		case got != tt.want:
			t.Errorf("%s %q: canonical %q, want %q", tt.key, tt.label, got, tt.want)
		}
	}
}
