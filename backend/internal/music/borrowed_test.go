package music

import (
	"strings"
	"testing"
)

// Borrowed chords in major keys (BL-094, ADR-0012 §4): the parallel minor
// lends iv, ivm6, ii°, ♭VI and ♭VII. Each is spelled by its degree's letter
// and explained as borrowed.
func TestBorrowedChordsInMajor(t *testing.T) {
	c := Key{C, Major}
	e := Key{E, Major}
	tests := []struct {
		key       Key
		label     string
		canonical string
		symbol    string
		quality   ChordQuality
	}{
		{c, "iv", "iv", "Fm", MinorTriad},
		{c, "iv7", "iv7", "Fm7", Minor7},
		{c, "ivm6", "ivm6", "Fm6", MinorSixth},
		{c, "ii°", "ii°", "Ddim", DiminishedTriad},
		{c, "iiø7", "iiø7", "Dm7b5", HalfDiminished7},
		{c, "bVI", "♭VI", "Ab", MajorTriad},
		{c, "♭VI", "♭VI", "Ab", MajorTriad},
		{c, "bVImaj7", "♭VImaj7", "Abmaj7", Major7},
		{c, "bVII", "♭VII", "Bb", MajorTriad},
		{c, "bVII7", "♭VII7", "Bb7", Dominant7},
		{e, "bVI", "♭VI", "C", MajorTriad},   // E major's sixth is C#; flattened it is C, not B#
		{e, "bVII", "♭VII", "D", MajorTriad}, // D, not C##
		{e, "iv", "iv", "Am", MinorTriad},
	}
	for _, tt := range tests {
		got, err := CanonicalLabel(tt.key, tt.label)
		if err != nil || got != tt.canonical {
			t.Errorf("%s %q: canonical %q, %v; want %q", tt.key, tt.label, got, err, tt.canonical)
			continue
		}
		step, err := StepFromLabel(tt.key, tt.label, 4)
		if err != nil {
			t.Fatal(err)
		}
		tl, err := FromProgression(tt.key, Progression{step}, 100)
		if err != nil {
			t.Fatalf("%s %q: %v", tt.key, tt.label, err)
		}
		h := tl.Harmony[0]
		if h.Symbol != tt.symbol || h.Chord.Quality != tt.quality || h.RomanNumeral != tt.canonical {
			t.Errorf("%s %q: %s %s %v", tt.key, tt.label, h.RomanNumeral, h.Symbol, h.Chord.Quality)
		}
	}
}

func TestBorrowingIsLimitedToTheApprovedSet(t *testing.T) {
	c := Key{C, Major}
	for _, label := range []string{"bIII", "bII", "i", "v", "#iv°", "bVI7", "ivmaj7"} {
		if _, err := CanonicalLabel(c, label); err == nil {
			t.Errorf("%q must not be accepted in C major yet", label)
		}
	}
	// Minor keys borrow only their own set (BL-075): a major II is refused.
	if _, err := CanonicalLabel(Key{A, Minor}, "II"); err == nil || !strings.Contains(err.Error(), "would sound as Bdim (ii°)") {
		t.Errorf("minor keys refuse a major II: %v", err)
	}
}

func TestBorrowedChordsAreAnalysedAsBorrowed(t *testing.T) {
	key := Key{C, Major}
	var prog Progression
	for _, l := range []string{"I", "IV", "iv", "I", "bVII", "I"} {
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
	tl, err = Analyze(tl)
	if err != nil {
		t.Fatal(err)
	}
	iv := tl.Harmony[2].Analysis
	if iv.Degree != 4 || iv.Function != Subdominant || !iv.Borrowed {
		t.Errorf("iv: %+v", iv)
	}
	if !strings.Contains(iv.Explanation, "iv（Fm）") || !strings.Contains(iv.Explanation, "同主短調（C minor）から借りた") {
		t.Errorf("iv explanation: %s", iv.Explanation)
	}
	if tl.Harmony[3].Analysis.Cadence != Plagal {
		t.Errorf("iv -> I is the minor plagal cadence: %+v", tl.Harmony[3].Analysis)
	}
	bvii := tl.Harmony[4].Analysis
	if bvii.Degree != 7 || !bvii.Borrowed || !strings.Contains(bvii.Explanation, "♭VII（Bb）") {
		t.Errorf("♭VII: %+v", bvii)
	}
	if tl.Harmony[1].Analysis.Borrowed {
		t.Error("the diatonic IV is not borrowed")
	}
}
