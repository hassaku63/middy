package music

import "testing"

func TestLiteralChordReadsTheLabelAsWritten(t *testing.T) {
	cases := []struct {
		tonic PitchClass
		label string
		want  Chord
	}{
		{C, "bVII", Chord{Root: As, Quality: MajorTriad}},
		{C, "♭VI", Chord{Root: Gs, Quality: MajorTriad}},
		{C, "iv", Chord{Root: F, Quality: MinorTriad}},
		{C, "ivm6", Chord{Root: F, Quality: MinorSixth}},
		{C, "I7", Chord{Root: C, Quality: Dominant7}},
		{C, "IV7", Chord{Root: F, Quality: Dominant7}},
		{C, "ii7", Chord{Root: D, Quality: Minor7}},
		{C, "Imaj7", Chord{Root: C, Quality: Major7}},
		{C, "iiø7", Chord{Root: D, Quality: HalfDiminished7}},
		{C, "vii°", Chord{Root: B, Quality: DiminishedTriad}},
		{C, "bII", Chord{Root: Cs, Quality: MajorTriad}},
		{E, "bVII", Chord{Root: D, Quality: MajorTriad}},
		// Chords without a third, whatever the case (BL-132).
		{C, "Vsus4", Chord{Root: G, Quality: Sus4}},
		{C, "V7sus4", Chord{Root: G, Quality: Dominant7Sus4}},
		{C, "I5", Chord{Root: C, Quality: Power}},
		{C, "bVII5", Chord{Root: As, Quality: Power}},
		{C, "vi5", Chord{Root: A, Quality: Power}},
		// The same sound in any mode: D mixolydian's VII is D major's bVII.
		{D, "bVII", Chord{Root: C, Quality: MajorTriad}},
	}
	for _, c := range cases {
		got, err := LiteralChord(c.tonic, c.label)
		if err != nil {
			t.Fatalf("%v %q: %v", c.tonic, c.label, err)
		}
		if got != c.want {
			t.Errorf("%v %q = %v, want %v", c.tonic, c.label, got, c.want)
		}
	}
	for _, bad := range []string{"X", "#IV", ""} {
		if _, err := LiteralChord(C, bad); err == nil {
			t.Errorf("%q: want an error", bad)
		}
	}
}
