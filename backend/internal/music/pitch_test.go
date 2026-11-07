package music

import "testing"

func TestPitchConstruction(t *testing.T) {
	tests := []struct {
		pc     PitchClass
		octave int
		want   Pitch
		name   string
	}{
		{C, 4, 60, "C4"},
		{A, 4, 69, "A4"},
		{C, -1, 0, "C-1"},
		{G, 9, 127, "G9"},
		{Fs, 3, 54, "F#3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPitch(tt.pc, tt.octave)
			if p != tt.want {
				t.Fatalf("NewPitch(%v,%d) = %d, want %d", tt.pc, tt.octave, p, tt.want)
			}
			if !p.Valid() {
				t.Fatalf("%d should be valid", p)
			}
			if p.Class() != tt.pc || p.Octave() != tt.octave {
				t.Fatalf("round trip failed: class=%v octave=%d", p.Class(), p.Octave())
			}
			if p.String() != tt.name {
				t.Fatalf("String() = %q, want %q", p.String(), tt.name)
			}
		})
	}
}

func TestPitchRange(t *testing.T) {
	if Pitch(-1).Valid() || Pitch(128).Valid() {
		t.Fatal("out-of-range pitches must be invalid")
	}
	if got := Pitch(127).Transpose(1); got.Valid() {
		t.Fatalf("transposing above range must be invalid, got %d", got)
	}
}

func TestPitchClassArithmetic(t *testing.T) {
	if got := B.Add(1); got != C {
		t.Fatalf("B+1 = %v, want C", got)
	}
	if got := C.Add(-1); got != B {
		t.Fatalf("C-1 = %v, want B", got)
	}
	if got := PitchClass(-13).Normalize(); got != B {
		t.Fatalf("Normalize(-13) = %v, want B", got)
	}
	if Cs.String() != "C#" || Cs.FlatName() != "Db" {
		t.Fatalf("names: %q %q", Cs.String(), Cs.FlatName())
	}
}
