package music

import "testing"

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
	for _, in := range []string{"", "VIII", "C", "1", "I-V", "7", "iiii", "V9", "bVII"} {
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
