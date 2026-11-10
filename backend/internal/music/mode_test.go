package music

import (
	"strings"
	"testing"
)

func TestModalDiatonicChords(t *testing.T) {
	tests := []struct {
		key  Key
		want []string // roman numerals of the seven triads
		syms []string // their symbols, spelled for the key
	}{
		{Key{D, Dorian}, []string{"i", "ii", "III", "IV", "v", "vi°", "VII"}, []string{"Dm", "Em", "F", "G", "Am", "Bdim", "C"}},
		{Key{G, Mixolydian}, []string{"I", "ii", "iii°", "IV", "v", "vi", "VII"}, []string{"G", "Am", "Bdim", "C", "Dm", "Em", "F"}},
		{Key{Ds, Dorian}, []string{"i", "ii", "III", "IV", "v", "vi°", "VII"}, []string{"Ebm", "Fm", "Gb", "Ab", "Bbm", "Cdim", "Db"}},
	}
	for _, tt := range tests {
		for d := 1; d <= 7; d++ {
			c, err := tt.key.Diatonic(d, false)
			if err != nil {
				t.Fatal(err)
			}
			rn, _ := RomanNumeral(d, c.Quality)
			if rn != tt.want[d-1] || tt.key.ChordSymbol(c) != tt.syms[d-1] {
				t.Errorf("%s degree %d: %s %s, want %s %s", tt.key, d, rn, tt.key.ChordSymbol(c), tt.want[d-1], tt.syms[d-1])
			}
		}
	}
}

func TestModalKeyNamesAndJSON(t *testing.T) {
	for _, tt := range []struct {
		key  Key
		want string
	}{
		{Key{D, Dorian}, "D dorian"},
		{Key{Ds, Dorian}, "Eb dorian"}, // parent Db major
		{Key{G, Mixolydian}, "G mixolydian"},
		{Key{Cs, Mixolydian}, "Db mixolydian"}, // parent Gb major
	} {
		if got := tt.key.String(); got != tt.want {
			t.Errorf("%v: %q, want %q", tt.key, got, tt.want)
		}
	}
	var m Mode
	if err := m.UnmarshalText([]byte("Dorian")); err != nil || m != Dorian {
		t.Fatalf("dorian must round-trip through text: %v %v", m, err)
	}
}

func TestModalLabelsAreCheckedAgainstTheMode(t *testing.T) {
	dorian := Key{D, Dorian}
	if got, err := CanonicalLabel(dorian, "IV"); err != nil || got != "IV" {
		t.Fatalf("the dorian IV is a major chord here: %q %v", got, err)
	}
	if _, err := CanonicalLabel(dorian, "V"); err == nil || !strings.Contains(err.Error(), `"V" would sound as Am (v) in D dorian`) {
		t.Fatalf("an uppercase V is not available in dorian (borrowing comes later): %v", err)
	}
	mixo := Key{G, Mixolydian}
	if got, err := CanonicalLabel(mixo, "VII"); err != nil || got != "VII" {
		t.Fatalf("the mixolydian VII is a major chord here: %q %v", got, err)
	}
	if _, err := CanonicalLabel(mixo, "vii°"); err == nil {
		t.Fatal("vii° does not exist in mixolydian")
	}
}

func TestModalAnalysisNamesTheCharacteristicToneAndCadence(t *testing.T) {
	key := Key{G, Mixolydian}
	var prog Progression
	for _, l := range []string{"I", "VII", "IV", "I"} {
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
	tl, err = Analyze(tl)
	if err != nil {
		t.Fatal(err)
	}
	vii := tl.Harmony[1].Analysis
	if !strings.Contains(vii.Explanation, "ミクソリディアンの特徴音（主音から短 7 度の F）") {
		t.Errorf("VII must be explained by the characteristic tone: %s", vii.Explanation)
	}
	if tl.Harmony[0].Analysis.Explanation != tl.Harmony[3].Analysis.Explanation &&
		!strings.Contains(tl.Harmony[3].Analysis.Explanation, "主和音") {
		t.Errorf("I is still the tonic: %s", tl.Harmony[3].Analysis.Explanation)
	}
	if strings.Contains(tl.Harmony[0].Analysis.Explanation, "特徴音") {
		t.Errorf("I (G B D) has no F and must not claim the characteristic tone: %s", tl.Harmony[0].Analysis.Explanation)
	}
	// VII -> I is the modal cadence; in the major key it would be vii° -> I with no cadence label.
	prog2, _ := StepFromLabel(key, "VII", 4)
	tonic, _ := StepFromLabel(key, "I", 4)
	tl2, _ := FromProgression(key, Progression{prog2, tonic}, 100)
	tl2, _ = Analyze(tl2)
	if tl2.Harmony[1].Analysis.Cadence != Modal || !strings.Contains(tl2.Harmony[1].Analysis.Explanation, "モーダル終止") {
		t.Errorf("VII -> I in mixolydian is the modal cadence: %+v", tl2.Harmony[1].Analysis)
	}
	minor := Key{A, Minor}
	m7, _ := StepFromLabel(minor, "VII", 4)
	mi, _ := StepFromLabel(minor, "i", 4)
	tl3, _ := FromProgression(minor, Progression{m7, mi}, 100)
	tl3, _ = Analyze(tl3)
	if tl3.Harmony[1].Analysis.Cadence == Modal {
		t.Error("VII -> i in natural minor keeps today's labelling until ADR-0012's question is answered")
	}
}

// TestModalExplanationsDoNotSpeakOfFunctions (#366, ADR-0046): in a mode
// no chord is explained as tonic, subdominant or dominant, since no
// leading tone pulls back to the tonic; the tonic is the mode's centre
// and any other chord is named by its degree, then by what it does. The
// function stays in Analysis.Function (ADR-0015); major and minor are as
// before.
func TestModalExplanationsDoNotSpeakOfFunctions(t *testing.T) {
	cases := []struct {
		key    Key
		labels []string
		want   []string
	}{
		{Key{D, Dorian}, []string{"i", "IV", "i", "VII"}, []string{
			"i（Dm）はドリアンの主和音。調の中心で、落ち着いて聞こえる。",
			"IV（G）はドリアンの 4 度の和音。 この和音はドリアンの特徴音（主音から長 6 度の B）を含む。",
			"",
			"VII（C）はドリアンの 7 度の和音。",
		}},
		{Key{E, Phrygian}, []string{"i", "II", "i"}, []string{
			"",
			"II（F）はフリジアンの 2 度の和音。",
			"",
		}},
		{Key{G, Mixolydian}, []string{"I", "v", "I"}, []string{
			"",
			"v（Dm）はミクソリディアンの 5 度の和音。",
			"",
		}},
	}
	for _, c := range cases {
		var prog Progression
		for _, l := range c.labels {
			s, err := StepFromLabel(c.key, l, 4)
			if err != nil {
				t.Fatal(err)
			}
			prog = append(prog, s)
		}
		tl, err := FromProgression(c.key, prog, 100)
		if err != nil {
			t.Fatal(err)
		}
		if tl, err = Analyze(tl); err != nil {
			t.Fatal(err)
		}
		for i, h := range tl.Harmony {
			text := h.Analysis.Explanation
			for _, fn := range []string{"トニック（T）", "サブドミナント（S）", "ドミナント（D）"} {
				if strings.Contains(text, fn) {
					t.Errorf("%v %s: a mode must not speak of %s: %s", c.key, h.RomanNumeral, fn, text)
				}
			}
			if c.want[i] != "" && !strings.HasPrefix(text, c.want[i]) {
				t.Errorf("%v %s:\n got %s\nwant %s…", c.key, h.RomanNumeral, text, c.want[i])
			}
		}
	}
	// Major keeps the function.
	major := Key{C, Major}
	iv, _ := StepFromLabel(major, "IV", 4)
	tl, _ := FromProgression(major, Progression{iv}, 100)
	tl, _ = Analyze(tl)
	if !strings.Contains(tl.Harmony[0].Analysis.Explanation, "サブドミナント（S）") {
		t.Errorf("major keeps the function: %s", tl.Harmony[0].Analysis.Explanation)
	}
}
