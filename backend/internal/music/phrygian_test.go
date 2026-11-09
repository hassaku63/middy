package music

import (
	"strings"
	"testing"
)

// Phrygian (BL-134, #165 E1): minor with a minor second. E phrygian has the
// signature of C major; its major chord on the second degree (F) marks the
// mode, and moving from it down a semitone to the tonic is the phrygian
// cadence.
func TestPhrygianKeyChordsAndNames(t *testing.T) {
	key := Key{E, Phrygian}
	want := []string{"i", "II", "III", "iv", "v°", "VI", "vii"}
	syms := []string{"Em", "F", "G", "Am", "Bdim", "C", "Dm"}
	for d := 1; d <= 7; d++ {
		c, err := key.Diatonic(d, false)
		if err != nil {
			t.Fatal(err)
		}
		rn, _ := RomanNumeral(d, c.Quality)
		if rn != want[d-1] || key.ChordSymbol(c) != syms[d-1] {
			t.Errorf("degree %d: %s %s, want %s %s", d, rn, key.ChordSymbol(c), want[d-1], syms[d-1])
		}
	}
	for _, tt := range []struct {
		key  Key
		want string
	}{
		{Key{E, Phrygian}, "E phrygian"},
		{Key{Ds, Phrygian}, "D# phrygian"}, // parent B major
		{Key{As, Phrygian}, "Bb phrygian"}, // parent Gb major
	} {
		if got := tt.key.String(); got != tt.want {
			t.Errorf("%v: %q, want %q", tt.key, got, tt.want)
		}
	}
	var m Mode
	if err := m.UnmarshalText([]byte("Phrygian")); err != nil || m != Phrygian {
		t.Fatalf("phrygian must round-trip through text: %v %v", m, err)
	}
}

func TestPhrygianExplainsTheSecond(t *testing.T) {
	key := Key{E, Phrygian}
	var prog Progression
	for _, l := range []string{"i", "II", "III", "II", "i"} {
		s, err := StepFromLabel(key, l, 4)
		if err != nil {
			t.Fatalf("%s: %v", l, err)
		}
		prog = append(prog, s)
	}
	tl, err := FromProgression(key, prog, 100)
	if err != nil {
		t.Fatal(err)
	}
	if tl, err = Analyze(tl); err != nil {
		t.Fatal(err)
	}
	ii := tl.Harmony[1].Analysis.Explanation
	for _, want := range []string{"フリジアンの特徴音（主音から短 2 度の F）", "主音の半音上の長三和音（♭II）", "フリジアン終止"} {
		if !strings.Contains(ii, want) {
			t.Errorf("II lacks %q: %s", want, ii)
		}
	}
	if strings.Contains(tl.Harmony[2].Analysis.Explanation, "フリジアン終止") {
		t.Errorf("III is not the phrygian second: %s", tl.Harmony[2].Analysis.Explanation)
	}
	if strings.Contains(tl.Harmony[0].Analysis.Explanation, "特徴音") {
		t.Errorf("i (E G B) has no F: %s", tl.Harmony[0].Analysis.Explanation)
	}
	if w := tl.Overview.Whole; !strings.Contains(w, "E phrygian") || !strings.Contains(w, "2 度の音が半音低い") {
		t.Errorf("the whole names the mode and its colour: %s", w)
	}
}
