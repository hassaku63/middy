package music

import (
	"strings"
	"testing"
)

// TestBassTextReview: a slash bass on the chord's own root is root
// position, not a note outside the chord; the bass of ivm6 on its sixth
// is the added sixth, not a seventh.
func TestBassTextReview(t *testing.T) {
	c := Key{Tonic: C, Mode: Major}
	explain := func(labels ...string) []string {
		t.Helper()
		var prog Progression
		for _, l := range labels {
			s, err := StepFromLabel(c, l, 4)
			if err != nil {
				t.Fatal(err)
			}
			prog = append(prog, s)
		}
		tl, err := FromProgression(c, prog, 100)
		if err != nil {
			t.Fatal(err)
		}
		an, err := Analyze(tl)
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, h := range an.Harmony {
			out = append(out, h.Analysis.Explanation)
		}
		return out
	}
	root := explain("IV", "I/1", "V")[1]
	if strings.Contains(root, "和音にない音") {
		t.Errorf("C under C is the chord's root: %s", root)
	}
	sixth := explain("I", "ivm6/2", "V")[1]
	if strings.Contains(sixth, "7 度") || !strings.Contains(sixth, "6 度") {
		t.Errorf("D under Fm6 is its added sixth: %s", sixth)
	}
}
