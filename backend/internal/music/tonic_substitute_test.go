package music

import (
	"strings"
	"testing"
)

// TestTonicSubstitutesAndWeakDominants (#371 M1, M2): only I is the
// centre of the key, and minor's v and VII have no leading tone.
func TestTonicSubstitutesAndWeakDominants(t *testing.T) {
	analyze := func(k Key, labels ...string) []string {
		t.Helper()
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
		var texts []string
		for _, h := range tl.Harmony {
			texts = append(texts, h.Analysis.Explanation)
		}
		return texts
	}
	cases := []struct {
		name    string
		key     Key
		labels  []string
		at      int
		want    []string
		notWant []string
	}{
		{"I is the centre", Key{C, Major}, []string{"I", "vi", "IV", "V", "I"}, 0,
			[]string{"トニック（T）。調の中心になる主和音で、落ち着いて聞こえる。"}, nil},
		{"vi stands in for I", Key{C, Major}, []string{"I", "vi", "IV", "V", "I"}, 1,
			[]string{"トニック（T）の仲間。I と 2 つの音を共有する代わりの和音で"}, []string{"調の中心"}},
		{"iii stands in for I", Key{C, Major}, []string{"I", "iii", "IV", "I"}, 1,
			[]string{"トニック（T）の仲間"}, []string{"調の中心"}},
		{"minor VI stands in for i", Key{A, Minor}, []string{"i", "VI", "iv", "i"}, 1,
			[]string{"i と 2 つの音を共有する"}, []string{"調の中心"}},
		{"minor v has no leading tone", Key{A, Minor}, []string{"i", "iv", "v", "i"}, 2,
			[]string{"ただし導音（主音の半音下の音）を持たないので、V（長三和音）ほど強くは i へ引き寄せない。"}, nil},
		{"harmonic V has it", Key{A, Minor}, []string{"i", "iv", "V", "i"}, 2,
			nil, []string{"導音を持たないので"}},
		{"VII with no context", Key{A, Minor}, []string{"i", "VI", "III", "VII"}, 3,
			[]string{"導音（主音の半音下の音）を持たないので、V ほど強くは i へ引き寄せない。"}, nil},
	}
	for _, c := range cases {
		got := analyze(c.key, c.labels...)[c.at]
		for _, w := range c.want {
			if !strings.Contains(got, w) {
				t.Errorf("%s: want %q in\n%s", c.name, w, got)
			}
		}
		for _, w := range c.notWant {
			if strings.Contains(got, w) {
				t.Errorf("%s: do not want %q in\n%s", c.name, w, got)
			}
		}
	}
}
