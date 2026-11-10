package music

import (
	"strings"
	"testing"
)

// TestReviewOfTheNightsTexts: the cadence sentence names the chord the
// cadence came from; the cadential six-four is a V's, named as the lane
// shows it, and in a mode or before minor's v it is a six-four only;
// V7/V through the six-four names the V it goes to.
func TestReviewOfTheNightsTexts(t *testing.T) {
	// I ii V V7: the half cadence comes from ii, past the V.
	half := analysed(t, Key{C, Major}, 4, "I", "ii", "V", "V7").Harmony[3].Analysis
	if half.Cadence != Half || !strings.Contains(half.Explanation, "ii から V7 に着いて止まる半終止") ||
		strings.Contains(half.Explanation, "直前の ii") || strings.Contains(half.Explanation, "V から V7") {
		t.Errorf("I ii V V7:\n%+v", half)
	}
	if e := analysed(t, Key{C, Major}, 4, "I", "IV", "ii", "V").Harmony[3].Analysis.Explanation; !strings.Contains(e, "直前の ii から V に着いて止まる半終止") {
		t.Errorf("I IV ii V:\n%s", e)
	}

	// A minor i/5 before the harmonic V7: named i and V7.
	minor := analysed(t, Key{A, Minor}, 4, "iv", "i/5", "V7", "i").Harmony[1].Analysis
	if minor.Function != Dominant || !strings.Contains(minor.Explanation, "次の V7 を飾る和音（終止の四六。V の前に置く、i の第 2 転回形）") {
		t.Errorf("A minor i/5 – V7:\n%+v", minor)
	}
	// Before minor's v and in a mode: a six-four only, no dominant.
	for _, c := range []struct {
		key    Key
		labels []string
	}{
		{Key{A, Minor}, []string{"iv", "i/5", "v", "i"}},
		{Key{D, Dorian}, []string{"IV", "i/5", "v", "i"}},
	} {
		a := analysed(t, c.key, 4, c.labels...).Harmony[1].Analysis
		if a.Function == Dominant || strings.Contains(a.Explanation, "終止の四六") || strings.Contains(a.Explanation, "ドミナント") {
			t.Errorf("%v %v:\n%+v", c.key, c.labels, a)
		}
	}

	// A minor V7/V – i/5 – V7: it goes to V (E), not the scale's v.
	applied := analysed(t, Key{A, Minor}, 4, "i", "V7/V", "i/5", "V7", "i").Harmony[1].Analysis.Explanation
	if strings.Contains(applied, "v（Em）") || !strings.Contains(applied, "V（E）へ向かう") ||
		!strings.Contains(applied, "次の i/5（V を飾る和音）を経て、V へ解決する") {
		t.Errorf("A minor V7/V – i/5 – V7:\n%s", applied)
	}
}

// TestSecondReviewOfTheNightsTexts: the overview names the chord the
// cadence came from and does not settle on a tonic six-four; a borrowed
// ♭VII before iv names the iv.
func TestSecondReviewOfTheNightsTexts(t *testing.T) {
	sus := analysed(t, Key{C, Major}, 4, "I", "IV", "Vsus4", "V").Overview.Flow
	if !strings.Contains(sus, "最後は IV から V への半終止") {
		t.Errorf("I IV Vsus4 V flow:\n%s", sus)
	}
	v7 := analysed(t, Key{C, Major}, 4, "I", "ii", "V", "V7").Overview.Flow
	if strings.Contains(v7, "V から V7") || !strings.Contains(v7, "最後は ii から V7 への半終止") {
		t.Errorf("I ii V V7 flow:\n%s", v7)
	}
	six := analysed(t, Key{C, Major}, 4, "I", "IV", "V", "I/5").Overview.Flow
	if strings.Contains(six, "主和音に戻って落ち着く") || !strings.Contains(six, "響きが浮いたまま終わる") {
		t.Errorf("ending on I/5:\n%s", six)
	}
	b := analysed(t, Key{C, Major}, 4, "I", "bVII", "iv", "I").Harmony[1].Analysis.Explanation
	if strings.Contains(b, "次の IV") || !strings.Contains(b, "次の iv へ進み") {
		t.Errorf("♭VII – iv:\n%s", b)
	}
}
