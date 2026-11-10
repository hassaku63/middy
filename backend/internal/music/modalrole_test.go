package music

import (
	"strings"
	"testing"
)

// TestModalRoles (#366, ADR-0047): every triad of a mode gets its role by
// the rule alone. The characteristic tone without the tritone marks the
// mode; the tritone blurs it towards the parallel major.
func TestModalRoles(t *testing.T) {
	const (
		tn = RoleTonic
		ch = RoleCharacteristic
		tt = RoleTritone
		ot = RoleOther
	)
	cases := []struct {
		key  Key
		want [8]ModalRole // by degree, 1..7
	}{
		// D dorian, B the characteristic major sixth, F–B the tritone:
		// i Dm, ii Em (B), III F, IV G (B), v Am, vi° Bdim (F, B), VII C.
		{Key{D, Dorian}, [8]ModalRole{"", tn, ch, ot, ch, ot, tt, ot}},
		// G mixolydian, F the minor seventh, B–F the tritone:
		// I G, ii Am, iii° Bdim (B, F), IV C, v Dm (F), vi Em, ♭VII F (F).
		{Key{G, Mixolydian}, [8]ModalRole{"", tn, ot, tt, ot, ch, ot, ch}},
		// E phrygian, F the minor second, F–B the tritone:
		// i Em, ♭II F (F), ♭III G, iv Am, v° Bdim (F, B), ♭VI C, ♭vii Dm (F).
		{Key{E, Phrygian}, [8]ModalRole{"", tn, ch, ot, ot, tt, ot, ch}},
		// A mode in another key follows the same rule: A dorian.
		{Key{A, Dorian}, [8]ModalRole{"", tn, ch, ot, ch, ot, tt, ot}},
	}
	for _, c := range cases {
		for d := 1; d <= 7; d++ {
			if got := modalRole(c.key, d); got != c.want[d] {
				t.Errorf("%v degree %d: got %q, want %q", c.key, d, got, c.want[d])
			}
		}
	}
	for _, k := range []Key{{C, Major}, {A, Minor}} {
		for d := 1; d <= 7; d++ {
			if got := modalRole(k, d); got != "" {
				t.Errorf("%v degree %d has a function, not a role: got %q", k, d, got)
			}
		}
	}
	if got := (Key{D, Dorian}).parallelMajor(); got != (Key{C, Major}) {
		t.Errorf("D dorian's notes are C major's, got %v", got)
	}
	if got := (Key{A, Dorian}).parallelMajor(); got != (Key{G, Major}) {
		t.Errorf("A dorian's notes are G major's, got %v", got)
	}
}

// TestModalRolesInTheAnalysis: the roles reach the analysis and the degree
// map, and the text says what each role does.
func TestModalRolesInTheAnalysis(t *testing.T) {
	analyze := func(k Key, labels ...string) Timeline {
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
		return tl
	}

	mixo := analyze(Key{G, Mixolydian}, "I", "VII", "v", "I")
	vii, v := mixo.Harmony[1].Analysis, mixo.Harmony[2].Analysis
	if vii.Role != RoleCharacteristic || !strings.Contains(vii.Explanation, "三全音を含まない特徴和音") ||
		!strings.Contains(vii.Explanation, "特徴音が根音にあるので、主和音へ戻る終わり方によく使われる") {
		t.Errorf("♭VII carries the minor seventh as its root: %+v", vii)
	}
	if v.Role != RoleCharacteristic || strings.Contains(v.Explanation, "根音にあるので") {
		t.Errorf("v carries the minor seventh as its third, not its root: %+v", v)
	}
	if mixo.Harmony[0].Analysis.Role != RoleTonic {
		t.Errorf("I is the tonic: %+v", mixo.Harmony[0].Analysis)
	}

	dorian := analyze(Key{D, Dorian}, "i", "vi°", "VII", "i")
	dim, sub := dorian.Harmony[1].Analysis, dorian.Harmony[2].Analysis
	want := "三全音（F と B）を含むので、同じ音でできた C major に引き寄せられやすく、ドリアンの中心がぼやける。"
	if dim.Role != RoleTritone || !strings.Contains(dim.Explanation, want) || strings.Contains(dim.Explanation, "明るく聞こえる理由") {
		t.Errorf("vi° blurs dorian towards C major:\n%s", dim.Explanation)
	}
	if sub.Role != RoleOther || strings.Contains(sub.Explanation, "特徴和音") {
		t.Errorf("VII has no characteristic tone: %+v", sub)
	}

	roles := map[string]ModalRole{}
	for _, d := range dorian.Overview.Degrees {
		roles[d.RomanNumeral] = d.Role
	}
	if roles["i"] != RoleTonic || roles["IV"] != RoleCharacteristic || roles["vi°"] != RoleTritone || roles["VII"] != RoleOther {
		t.Errorf("the degree map carries the roles: %v", roles)
	}

	// A seventh can add the tritone a triad lacks (#371): dorian IV7 (G7)
	// and phrygian III7 (G7) have F and B. The analysis says so; the
	// degree map keeps the triad's role.
	vamp := analyze(Key{D, Dorian}, "i7", "IV7", "IV", "i7")
	iv7, iv := vamp.Harmony[1].Analysis, vamp.Harmony[2].Analysis
	want = "7 度の音を足すと三全音（F と B）ができるので、同じ音でできた C major に引き寄せられやすく、ドリアンの中心がぼやける。三和音の IV（G）のほうがドリアンらしく聞こえる。"
	if iv7.Role != RoleTritone || !strings.Contains(iv7.Explanation, want) ||
		strings.Contains(iv7.Explanation, "三全音を含まない") || strings.Contains(iv7.Explanation, "明るく聞こえる理由") {
		t.Errorf("IV7 adds the tritone:\n%s", iv7.Explanation)
	}
	if iv.Role != RoleCharacteristic || !strings.Contains(iv.Explanation, "三全音を含まない特徴和音") {
		t.Errorf("the triad IV stays characteristic: %+v", iv)
	}
	for _, d := range vamp.Overview.Degrees {
		if d.Degree == 4 && d.Role != RoleCharacteristic {
			t.Errorf("the map keeps the triad's role: %+v", d)
		}
	}
	if vamp.Harmony[0].Analysis.Role != RoleTonic {
		t.Errorf("i7 is the tonic: %+v", vamp.Harmony[0].Analysis)
	}
	phr := analyze(Key{E, Phrygian}, "i", "III7", "VImaj7", "i")
	if a := phr.Harmony[1].Analysis; a.Role != RoleTritone || !strings.Contains(a.Explanation, "三和音の III（G）のほうがフリジアンらしく聞こえる") ||
		strings.Contains(a.Explanation, "張りつめて聞こえる理由") {
		t.Errorf("phrygian III7 adds the tritone:\n%s", a.Explanation)
	}

	major := analyze(Key{C, Major}, "I", "IV", "V", "I")
	for _, h := range major.Harmony {
		if h.Analysis.Role != "" {
			t.Errorf("major has functions, not roles: %+v", h.Analysis)
		}
	}
}
