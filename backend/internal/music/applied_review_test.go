package music

import (
	"strings"
	"testing"
)

func analysedLabels(t *testing.T, key Key, labels ...string) Timeline {
	t.Helper()
	var prog Progression
	for _, l := range labels {
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
	an, err := Analyze(tl)
	if err != nil {
		t.Fatal(err)
	}
	return an
}

// TestAppliedChordsAreNotWhatTheirRootSays (review of BL-074/BL-098): a
// secondary dominant on the tonic's root is not the tonic, one on degree 4
// or 7 makes no plagal or modal cadence, and the modal comparison leaves
// applied chords out.
func TestAppliedChordsAreNotWhatTheirRootSays(t *testing.T) {
	c := Key{Tonic: C, Mode: Major}
	end := analysedLabels(t, c, "ii", "V", "I", "V7/IV")
	if flow := end.Overview.Flow; strings.Contains(flow, "主和音に戻って") {
		t.Errorf("C7 as V7/IV is no return home: %s", flow)
	}
	start := analysedLabels(t, c, "V7/IV", "IV", "V", "I")
	if flow := start.Overview.Flow; strings.Contains(flow, "主和音から始まり") {
		t.Errorf("C7 as V7/IV is no tonic start: %s", flow)
	}
	cm := Key{Tonic: C, Mode: Minor}
	plagal := analysedLabels(t, cm, "i", "V/VII", "i")
	if got := plagal.Harmony[2].Analysis.Cadence; got == Plagal {
		t.Errorf("V/VII – i is no plagal cadence")
	}
	d := Key{Tonic: D, Mode: Dorian}
	dorian := analysedLabels(t, d, "i", "V/ii", "ii", "i")
	if s := dorian.Overview.Whole; strings.Contains(s, "VI（Bb）") {
		t.Errorf("an applied chord is no stand-in for the minor VI: %s", s)
	}
}
