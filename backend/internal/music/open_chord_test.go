package music

import (
	"strings"
	"testing"
)

// TestOpenChordsInTheAnalysis (#378 A1, A2): a power or sus4 chord is
// explained by the notes it has, not by the triad on its degree.
func TestOpenChordsInTheAnalysis(t *testing.T) {
	// D dorian IV5 (G, D) and IVsus4 (G, C, D) lack B, the characteristic
	// tone: no characteristic chord, though the map keeps IV's role.
	dorian := analysed(t, Key{D, Dorian}, 4, "i5", "IV5", "IVsus4", "i5")
	for _, i := range []int{1, 2} {
		a := dorian.Harmony[i].Analysis
		if a.Role != RoleOther || strings.Contains(a.Explanation, "特徴和音") {
			t.Errorf("%s has no B: %+v", dorian.Harmony[i].RomanNumeral, a)
		}
	}
	for _, d := range dorian.Overview.Degrees {
		if d.Degree == 4 && d.Role != RoleCharacteristic {
			t.Errorf("the map keeps the triad's role: %+v", d)
		}
	}
	if a := analysed(t, Key{D, Dorian}, 4, "i", "IV", "i").Harmony[1].Analysis; a.Role != RoleCharacteristic {
		t.Errorf("the triad IV keeps its role: %+v", a)
	}
	if a := analysed(t, Key{G, Mixolydian}, 4, "I5", "v5", "I5").Harmony[1].Analysis; a.Role != RoleOther {
		t.Errorf("mixolydian v5 (D, A) has no F: %+v", a)
	}

	// vi5 (A, E) shares one note with C major's I, ♭III5 (E♭, B♭) one
	// with C minor's i: no count of shared notes for them.
	major := analysed(t, Key{C, Major}, 4, "I5", "vi5", "bIII5", "I5")
	vi, iii := major.Harmony[1].Analysis.Explanation, major.Harmony[2].Analysis.Explanation
	if !strings.Contains(vi, "トニック（T）の仲間の度数。三和音なら I と 2 つの音を共有する") {
		t.Errorf("vi5:\n%s", vi)
	}
	if !strings.Contains(iii, "同主短調ではトニック（T）の仲間の度数（三和音なら短調の i と 2 つの音を共有する）") {
		t.Errorf("♭III5:\n%s", iii)
	}
	if a := analysed(t, Key{C, Major}, 4, "I", "vi", "I").Harmony[1].Analysis; !strings.Contains(a.Explanation, "トニック（T）の仲間。I と 2 つの音を共有する") {
		t.Errorf("the triad vi keeps its sentence:\n%s", a.Explanation)
	}
}
