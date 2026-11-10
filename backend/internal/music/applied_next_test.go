package music

import (
	"strings"
	"testing"
)

// TestAppliedAndSubtonicFollowTheNextChord (#378 B2, B5, B6): what a
// secondary dominant or a minor VII is said to do matches the chord that
// actually follows.
func TestAppliedAndSubtonicFollowTheNextChord(t *testing.T) {
	// V/V – I6/4 – V: the six-four delays the V; nothing is dodged.
	six := analysed(t, Key{C, Major}, 4, "ii", "V7/V", "I/5", "V7")
	if e := six.Harmony[1].Analysis.Explanation; strings.Contains(e, "かわす") ||
		!strings.Contains(e, "次の I/5（V を飾る和音）を経て、V へ解決する") {
		t.Errorf("V7/V – I/5 – V7:\n%s", e)
	}
	if e := analysed(t, Key{C, Major}, 4, "I", "V7/V", "vi", "I").Harmony[1].Analysis.Explanation; !strings.Contains(e, "引き寄せをかわす") {
		t.Errorf("V7/V – vi dodges:\n%s", e)
	}

	// A minor V7/V – V7: it goes to E7, so its target is V (E), not v (Em).
	minor := analysed(t, Key{A, Minor}, 4, "i", "V7/V", "V7", "i")
	if e := minor.Harmony[1].Analysis.Explanation; strings.Contains(e, "v（Em）") ||
		!strings.Contains(e, "V（E）へ向かう") || !strings.Contains(e, "次の V へ解決する") {
		t.Errorf("A minor V7/V – V7:\n%s", e)
	}

	// i – VII – VI with no V after: down to VI only.
	noV := analysed(t, Key{A, Minor}, 4, "i", "VII", "VI", "VII")
	if e := noV.Harmony[1].Analysis.Explanation; strings.Contains(e, "V へ下りていく") ||
		!strings.Contains(e, "i から VI へ下りていく線の途中") {
		t.Errorf("i – VII – VI – VII:\n%s", e)
	}
	withV := analysed(t, Key{A, Minor}, 4, "i", "VII", "VI", "V")
	if e := withV.Harmony[1].Analysis.Explanation; !strings.Contains(e, "i から VI、V へ下りていく線の途中") {
		t.Errorf("i – VII – VI – V:\n%s", e)
	}
}
