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

// TestAppliedChordsNextToVAndVII (review): V/V – V arrives on V, a half
// cadence; a minor key's VII before V7/VI does not head to III, and VII
// after V/iv does not come from i, whatever their roots' degrees say.
func TestAppliedChordsNextToVAndVII(t *testing.T) {
	c := Key{Tonic: C, Mode: Major}
	half := analysedLabels(t, c, "I", "IV", "V/V", "V")
	if got := half.Harmony[3].Analysis.Cadence; got != Half {
		t.Errorf("V/V – V cadence = %v, want half", got)
	}
	cm := Key{Tonic: C, Mode: Minor}
	toVI := analysedLabels(t, cm, "i", "VII", "V7/VI", "VI")
	if e := toVI.Harmony[1].Analysis.Explanation; strings.Contains(e, "III へ向かう") {
		t.Errorf("VII before V7/VI read as heading to III: %s", e)
	}
	fromIV := analysedLabels(t, cm, "V/iv", "VII", "VI", "V")
	if e := fromIV.Harmony[1].Analysis.Explanation; strings.Contains(e, "i から VI") {
		t.Errorf("VII after V/iv read as coming from i: %s", e)
	}
}
