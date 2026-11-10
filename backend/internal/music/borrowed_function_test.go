package music

import (
	"strings"
	"testing"
)

// TestBorrowedFunctions (#371 C2, ADR-0050): a major key's chords from the
// parallel minor take the function they do there, not their degree's.
func TestBorrowedFunctions(t *testing.T) {
	analyze := func(labels ...string) Timeline {
		t.Helper()
		k := Key{C, Major}
		var prog Progression
		for _, l := range labels {
			s, err := StepFromLabel(k, l, 4)
			if err != nil {
				t.Fatal(err)
			}
			prog = append(prog, s)
		}
		tl, err := FromProgression(k, prog, 100)
		if err != nil {
			t.Fatal(err)
		}
		if tl, err = Analyze(tl); err != nil {
			t.Fatal(err)
		}
		return tl
	}
	cases := []struct {
		name    string
		labels  []string
		at      int
		fn      Function
		want    []string
		notWant []string
	}{
		{"♭VI is subdominant minor", []string{"I", "bVI", "bVII", "I"}, 1, Subdominant,
			[]string{"サブドミナント（S）", "長調の中では iv と同じく", "サブドミナント・マイナー", "短調の中の VI は、トニックの代わりになる"},
			[]string{"調の中心"}},
		{"♭VII before I is a weak dominant", []string{"I", "bVI", "bVII", "I"}, 2, Dominant,
			[]string{"を持たないので、V ほど強くは I へ引き寄せない、弱いドミナント"}, nil},
		{"♭VII before IV is subdominant", []string{"I", "bVII", "IV", "I"}, 1, Subdominant,
			[]string{"サブドミナント（S）", "次の IV へ進み"}, []string{"ドミナント（D）"}},
		{"♭III5 is the parallel minor's tonic side", []string{"I", "bIII5", "IV", "I"}, 1, Tonic,
			[]string{"同主短調ではトニック（T）の仲間"}, []string{"調の中心", "I と 2 つの音を共有"}},
		{"iv stays subdominant", []string{"I", "iv", "I"}, 1, Subdominant,
			[]string{"サブドミナント（S）"}, []string{"サブドミナント・マイナーの仲間"}},
	}
	for _, c := range cases {
		a := analyze(c.labels...).Harmony[c.at].Analysis
		if a.Function != c.fn {
			t.Errorf("%s: function %s, want %s", c.name, a.Function, c.fn)
		}
		for _, w := range c.want {
			if !strings.Contains(a.Explanation, w) {
				t.Errorf("%s: want %q in\n%s", c.name, w, a.Explanation)
			}
		}
		for _, w := range c.notWant {
			if strings.Contains(a.Explanation, w) {
				t.Errorf("%s: do not want %q in\n%s", c.name, w, a.Explanation)
			}
		}
	}
}
