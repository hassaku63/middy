package music

import (
	"strings"
	"testing"
)

// TestCadencesEndPhrases (#371 C1, ADR-0051): the cadential six-four is
// no plagal cadence, an inverted V – I is the weaker authentic cadence,
// and an IV – I inside a phrase is no "Amen cadence".
func TestCadencesEndPhrases(t *testing.T) {
	// IV – I/5 – V – I: the I6/4 decorates the V.
	six := analysed(t, Key{C, Major}, 4, "IV", "I/5", "V", "I")
	if a := six.Harmony[1].Analysis; a.Cadence != NoCadence || a.Function != Dominant ||
		strings.Contains(a.Explanation, "変格終止") || strings.Contains(a.Explanation, "落ち着いて") ||
		!strings.Contains(a.Explanation, "次の V を飾る和音（終止の四六）") {
		t.Errorf("I6/4 before V:\n%+v", a)
	}
	if a := six.Harmony[3].Analysis; a.Cadence != Authentic || !strings.Contains(a.Explanation, "もっとも強い区切り") {
		t.Errorf("root-position V – I closes strongly:\n%+v", a)
	}

	// V7/7 – I/3 (G7/B – C/E): still authentic, but weaker.
	inv := analysed(t, Key{C, Major}, 4, "I", "IV", "V7/7", "I/3")
	if a := inv.Harmony[3].Analysis; a.Cadence != Authentic || strings.Contains(a.Explanation, "もっとも強い") ||
		!strings.Contains(a.Explanation, "基本形どうしより区切りは弱い") {
		t.Errorf("inverted V – I:\n%+v", a)
	}

	// IV – I in the second bar of a phrase is a motion, not a cadence.
	mid := analysed(t, Key{C, Major}, 4, "IV", "I", "ii", "V")
	if a := mid.Harmony[1].Analysis; a.Cadence != NoCadence || strings.Contains(a.Explanation, "アーメン終止") {
		t.Errorf("IV – I mid-phrase:\n%+v", a)
	}
	if a := mid.Harmony[3].Analysis; a.Cadence != Half {
		t.Errorf("the phrase ends on V: %+v", a)
	}
}
