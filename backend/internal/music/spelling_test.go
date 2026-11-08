package music

import "testing"

func TestKeyNamesFollowTheKeySignature(t *testing.T) {
	tests := []struct {
		key  Key
		want string
	}{
		{Key{C, Major}, "C major"},
		{Key{Cs, Major}, "Db major"}, // 5 flats beat 7 sharps
		{Key{Fs, Major}, "Gb major"}, // a tie (6 each) goes to flats
		{Key{Gs, Major}, "Ab major"},
		{Key{As, Major}, "Bb major"},
		{Key{B, Major}, "B major"},
		{Key{A, Minor}, "A minor"},
		{Key{As, Minor}, "Bb minor"},
		{Key{Cs, Minor}, "C# minor"},
		{Key{Ds, Minor}, "Eb minor"}, // parent Gb major
		{Key{Gs, Minor}, "G# minor"}, // parent B major
		{Key{F, Minor}, "F minor"},
	}
	for _, tt := range tests {
		if got := tt.key.String(); got != tt.want {
			t.Errorf("Key{%d,%v}.String() = %q, want %q", tt.key.Tonic, tt.key.Mode, got, tt.want)
		}
	}
}

func TestChordSymbolsUseTheKeysLetters(t *testing.T) {
	tests := []struct {
		key    Key
		degree int
		sev    bool
		want   string
	}{
		{Key{As, Major}, 1, true, "Bbmaj7"},
		{Key{As, Major}, 6, true, "Gm7"},
		{Key{As, Major}, 4, false, "Eb"},
		{Key{Ds, Major}, 7, false, "Ddim"},
		{Key{Cs, Major}, 2, true, "Ebm7"},
		{Key{Gs, Minor}, 3, false, "B"},
		{Key{Gs, Minor}, 6, false, "E"},
		{Key{Ds, Minor}, 6, false, "Cb"}, // the VI of Eb minor is Cb, not B
	}
	for _, tt := range tests {
		c, err := tt.key.Diatonic(tt.degree, tt.sev)
		if err != nil {
			t.Fatal(err)
		}
		if got := tt.key.ChordSymbol(c); got != tt.want {
			t.Errorf("%s degree %d: %q, want %q", tt.key, tt.degree, got, tt.want)
		}
	}
}

func TestHarmonicMinorDominantIsSpelledOnTheFifthDegree(t *testing.T) {
	// D minor's raised dominant is A7 (with C#), F minor's is C7 (with E).
	for _, tt := range []struct {
		key  Key
		want string
	}{{Key{D, Minor}, "A7"}, {Key{F, Minor}, "C7"}, {Key{Ds, Minor}, "Bb7"}} {
		step, err := StepFromLabel(tt.key, "V7", 4)
		if err != nil {
			t.Fatal(err)
		}
		tl, err := FromProgression(tt.key, Progression{step}, 100)
		if err != nil {
			t.Fatal(err)
		}
		if tl.Harmony[0].Symbol != tt.want {
			t.Errorf("%s V7 symbol = %q, want %q", tt.key, tl.Harmony[0].Symbol, tt.want)
		}
	}
}

func TestChromaticRootsFallBackToTheKeysSide(t *testing.T) {
	// Not diatonic to the key: spelled with the key's own accidental.
	flat := Key{F, Major}
	if got := flat.ChordSymbol(Chord{Root: Cs, Quality: MajorTriad}); got != "Db" {
		t.Errorf("F major: %q, want Db", got)
	}
	sharp := Key{D, Major}
	if got := sharp.ChordSymbol(Chord{Root: Ds, Quality: MinorTriad}); got != "D#m" {
		t.Errorf("D major: %q, want D#m", got)
	}
}
