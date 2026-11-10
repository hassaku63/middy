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
		!strings.Contains(a.Explanation, "次の V を飾る和音（終止の四六。V の前に置く、I の第 2 転回形）") {
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

// TestCadenceSentencesNameTheChords (#378 B4): the cadence sentence names
// the two chords as the chord lane shows them, case and all.
func TestCadenceSentencesNameTheChords(t *testing.T) {
	cases := []struct {
		key    Key
		labels []string
		want   string
	}{
		{Key{A, Minor}, []string{"i", "iv", "V", "VI"}, "V から主和音ではなく VI へ進む偽終止"},
		{Key{A, Minor}, []string{"i", "VI", "iv", "i"}, "直前の iv から i へ落ち着く変格終止"},
		{Key{A, Minor}, []string{"i", "iv", "V7", "i"}, "直前の V7 から i へ解決する正格終止"},
		{Key{D, Dorian}, []string{"i", "v", "IV", "i"}, "直前の IV から i へ落ち着く変格終止"},
		{Key{G, Mixolydian}, []string{"I", "IV", "v", "VII", "I"}, "直前の VII から I へ進むモーダル終止"},
		{Key{C, Major}, []string{"I", "vi", "ii", "V"}, "直前の ii から V に着いて止まる半終止"},
		{Key{C, Major}, []string{"I", "IV", "V", "I"}, "直前の V から I へ解決する正格終止。もっとも強い区切り。"},
	}
	for _, c := range cases {
		tl := analysed(t, c.key, 4, c.labels...)
		last := tl.Harmony[len(tl.Harmony)-1].Analysis.Explanation
		if !strings.Contains(last, c.want) {
			t.Errorf("%v %v: want %q in\n%s", c.key, c.labels, c.want, last)
		}
	}
}

// TestCadenceEdgeCases (#378 B3, B7): no cadence comes to rest on a tonic
// six-four, and a V prepared by its own sus4 or triad is a half cadence.
func TestCadenceEdgeCases(t *testing.T) {
	// The I/5 ends bar 4 and goes on to V: no authentic cadence on it.
	six := analysed(t, Key{C, Major}, 4, "I", "IV", "V", "I/5", "V", "I")
	if a := six.Harmony[3].Analysis; a.Cadence != NoCadence || strings.Contains(a.Explanation, "正格終止") {
		t.Errorf("V – I/5 at the phrase's end:\n%+v", a)
	}
	end := analysed(t, Key{C, Major}, 4, "I", "ii", "IV", "I/5")
	if a := end.Harmony[3].Analysis; a.Cadence != NoCadence || strings.Contains(a.Explanation, "変格終止") {
		t.Errorf("IV – I/5 at the end:\n%+v", a)
	}
	for _, labels := range [][]string{{"I", "IV", "Vsus4", "V"}, {"I", "ii", "V", "V7"}} {
		tl := analysed(t, Key{C, Major}, 4, labels...)
		if a := tl.Harmony[3].Analysis; a.Cadence != Half || !strings.Contains(a.Explanation, "半終止") {
			t.Errorf("%v ends on V: %+v", labels, a)
		}
	}
	// V – V – I is still one authentic cadence at the end.
	if a := analysed(t, Key{C, Major}, 4, "V", "V7", "I").Harmony[2].Analysis; a.Cadence != Authentic {
		t.Errorf("V – V7 – I: %+v", a)
	}
}
