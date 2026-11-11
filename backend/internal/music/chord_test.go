package music

import (
	"encoding/json"
	"testing"
)

func TestChordTones(t *testing.T) {
	tests := []struct {
		chord  Chord
		symbol string
		tones  []PitchClass
	}{
		{Chord{C, MajorTriad}, "C", []PitchClass{C, E, G}},
		{Chord{A, MinorTriad}, "Am", []PitchClass{A, C, E}},
		{Chord{B, DiminishedTriad}, "Bdim", []PitchClass{B, D, F}},
		{Chord{C, AugmentedTriad}, "Caug", []PitchClass{C, E, Gs}},
		{Chord{C, Major7}, "Cmaj7", []PitchClass{C, E, G, B}},
		{Chord{D, Minor7}, "Dm7", []PitchClass{D, F, A, C}},
		{Chord{G, Dominant7}, "G7", []PitchClass{G, B, D, F}},
		{Chord{B, HalfDiminished7}, "Bm7b5", []PitchClass{B, D, F, A}},
	}
	for _, tt := range tests {
		t.Run(tt.symbol, func(t *testing.T) {
			if got := tt.chord.String(); got != tt.symbol {
				t.Fatalf("String() = %q, want %q", got, tt.symbol)
			}
			got := tt.chord.PitchClasses()
			if len(got) != len(tt.tones) {
				t.Fatalf("PitchClasses() = %v, want %v", got, tt.tones)
			}
			for i := range got {
				if got[i] != tt.tones[i] {
					t.Fatalf("PitchClasses() = %v, want %v", got, tt.tones)
				}
			}
		})
	}
}

func TestChordVoicing(t *testing.T) {
	got := Chord{C, Major7}.Voicing(4)
	want := []Pitch{60, 64, 67, 71}
	if len(got) != len(want) {
		t.Fatalf("Voicing = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Voicing = %v, want %v", got, want)
		}
	}
}

func TestDiatonicChords(t *testing.T) {
	tests := []struct {
		key      Key
		seventh  bool
		symbols  []string
		numerals []string
	}{
		{
			Key{C, Major}, false,
			[]string{"C", "Dm", "Em", "F", "G", "Am", "Bdim"},
			[]string{"I", "ii", "iii", "IV", "V", "vi", "vii°"},
		},
		{
			Key{C, Major}, true,
			[]string{"Cmaj7", "Dm7", "Em7", "Fmaj7", "G7", "Am7", "Bm7b5"},
			[]string{"Imaj7", "ii7", "iii7", "IVmaj7", "V7", "vi7", "viiø7"},
		},
		{
			Key{A, Minor}, false,
			[]string{"Am", "Bdim", "C", "Dm", "Em", "F", "G"},
			[]string{"i", "ii°", "III", "iv", "v", "VI", "VII"},
		},
		{
			Key{A, Minor}, true,
			[]string{"Am7", "Bm7b5", "Cmaj7", "Dm7", "Em7", "Fmaj7", "G7"},
			[]string{"i7", "iiø7", "IIImaj7", "iv7", "v7", "VImaj7", "VII7"},
		},
		{
			Key{G, Major}, false,
			[]string{"G", "Am", "Bm", "C", "D", "Em", "F#dim"},
			[]string{"I", "ii", "iii", "IV", "V", "vi", "vii°"},
		},
	}
	for _, tt := range tests {
		name := tt.key.String()
		if tt.seventh {
			name += " sevenths"
		}
		t.Run(name, func(t *testing.T) {
			for d := 1; d <= 7; d++ {
				ch, err := tt.key.Diatonic(d, tt.seventh)
				if err != nil {
					t.Fatalf("Diatonic(%d): %v", d, err)
				}
				if got := ch.String(); got != tt.symbols[d-1] {
					t.Errorf("degree %d: chord = %q, want %q", d, got, tt.symbols[d-1])
				}
				rn, err := RomanNumeral(d, ch.Quality)
				if err != nil {
					t.Fatalf("RomanNumeral(%d): %v", d, err)
				}
				if rn != tt.numerals[d-1] {
					t.Errorf("degree %d: numeral = %q, want %q", d, rn, tt.numerals[d-1])
				}
			}
		})
	}
}

func TestDiatonicBounds(t *testing.T) {
	if _, err := (Key{C, Major}).Diatonic(0, false); err == nil {
		t.Fatal("degree 0 must fail")
	}
	if _, err := RomanNumeral(8, MajorTriad); err == nil {
		t.Fatal("degree 8 must fail")
	}
	if _, err := RomanNumeral(1, ChordQuality(99)); err == nil {
		t.Fatal("unknown quality must fail")
	}
}

func TestChordJSON(t *testing.T) {
	c := Chord{Root: G, Quality: Dominant7}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"root":7,"quality":"dom7"}` {
		t.Fatalf("unexpected JSON %s", b)
	}
	var back Chord
	if err := json.Unmarshal(b, &back); err != nil || back != c {
		t.Fatalf("round trip = %+v, %v", back, err)
	}
	if err := json.Unmarshal([]byte(`{"root":0,"quality":"add11"}`), &back); err == nil {
		t.Fatal("unknown quality must fail to unmarshal")
	}
}
