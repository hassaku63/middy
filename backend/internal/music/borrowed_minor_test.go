package music

import (
	"strings"
	"testing"
)

// BL-075: a minor key borrows the major I and IV of its parallel major and
// the Neapolitan bII, written against its own scale (ADR-0012, decision 3).
func TestMinorKeysBorrowPicardyIVAndNeapolitan(t *testing.T) {
	cm := Key{C, Minor}
	tests := []struct {
		label, numeral, symbol string
		chord                  Chord
		fn                     Function
		says                   string
	}{
		{"I", "I", "C", Chord{Root: C, Quality: MajorTriad}, Tonic, "同主長調（C major）から借りた和音"},
		{"IV", "IV", "F", Chord{Root: F, Quality: MajorTriad}, Subdominant, "短調の中に一瞬明るさを差す"},
		{"bII", "♭II", "Db", Chord{Root: Cs, Quality: MajorTriad}, Subdominant, "ナポリの和音（♭II）"},
		{"♭II", "♭II", "Db", Chord{Root: Cs, Quality: MajorTriad}, Subdominant, "V へ向かう前によく置かれる"},
	}
	for _, tt := range tests {
		tl := analysed(t, cm, 4, "i", tt.label, "V7", "i")
		h := tl.Harmony[1]
		if h.Chord != tt.chord || h.RomanNumeral != tt.numeral || h.Symbol != tt.symbol {
			t.Errorf("%q: %v %q %q, want %v %q %q", tt.label, h.Chord, h.RomanNumeral, h.Symbol, tt.chord, tt.numeral, tt.symbol)
		}
		a := h.Analysis
		if !a.Borrowed || a.Function != tt.fn || !strings.Contains(a.Explanation, tt.says) {
			t.Errorf("%q: analysis %+v", tt.label, a)
		}
	}
	// The Picardy third: a minor piece closing on the major I.
	tl := analysed(t, cm, 4, "i", "iv", "V7", "I")
	last := tl.Harmony[3].Analysis
	if last.Cadence != Authentic || !strings.Contains(last.Explanation, "ピカルディの 3 度") {
		t.Errorf("Picardy close: %+v", last)
	}
	// Only the set: no major II, no I7, no bVII as a borrowing (it is VII).
	for _, label := range []string{"II", "I7", "IV7", "IVmaj7", "bIImaj7"} {
		if _, err := CanonicalLabel(cm, label); err == nil {
			t.Errorf("%q must not be accepted in C minor", label)
		}
	}
}

// C major in C minor is two chords by function; the label says which.
func TestTheLabelTellsVOfIvFromTheBorrowedI(t *testing.T) {
	cm := Key{C, Minor}
	applied := analysed(t, cm, 4, "i", "V/iv", "iv", "i")
	if a := applied.Harmony[1].Analysis; a.AppliedTo != 4 || a.Borrowed || !strings.Contains(a.Explanation, "iv（Fm）へ向かう") {
		t.Errorf("V/iv: %+v", a)
	}
	borrowed := analysed(t, cm, 4, "i", "I", "iv", "i")
	if a := borrowed.Harmony[1].Analysis; a.AppliedTo != 0 || !a.Borrowed {
		t.Errorf("I: %+v", a)
	}
	if applied.Harmony[1].Chord != borrowed.Harmony[1].Chord {
		t.Error("both are the same C major chord")
	}
}
