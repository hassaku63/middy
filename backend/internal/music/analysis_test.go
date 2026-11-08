package music

import (
	"strings"
	"testing"
)

func harmonyOf(key Key, degrees ...int) Timeline {
	prog := make(Progression, 0, len(degrees))
	for _, d := range degrees {
		prog = append(prog, Step{Degree: d, Beats: 4})
	}
	tl, err := FromProgression(key, prog, 120)
	if err != nil {
		panic(err)
	}
	return tl
}

func TestAnalyzeFunctionsCoverEveryDiatonicDegree(t *testing.T) {
	want := []Function{Tonic, Subdominant, Tonic, Subdominant, Dominant, Tonic, Dominant}
	for _, key := range []Key{{C, Major}, {A, Minor}, {Fs, Major}} {
		t.Run(key.String(), func(t *testing.T) {
			tl, err := Analyze(harmonyOf(key, 1, 2, 3, 4, 5, 6, 7))
			if err != nil {
				t.Fatalf("Analyze: %v", err)
			}
			for i, h := range tl.Harmony {
				if h.Analysis == nil {
					t.Fatalf("harmony[%d] has no analysis", i)
				}
				if h.Analysis.Degree != i+1 {
					t.Errorf("harmony[%d]: degree = %d, want %d", i, h.Analysis.Degree, i+1)
				}
				if h.Analysis.Function != want[i] {
					t.Errorf("harmony[%d] (%s): function = %q, want %q", i, h.RomanNumeral, h.Analysis.Function, want[i])
				}
				if h.Analysis.Explanation == "" {
					t.Errorf("harmony[%d]: empty explanation", i)
				}
			}
		})
	}
}

func TestAnalyzeDetectsCadences(t *testing.T) {
	tests := []struct {
		name    string
		key     Key
		degrees []int
		want    []Cadence // per harmony event; first is always NoCadence
	}{
		{"authentic V-I", Key{C, Major}, []int{5, 1}, []Cadence{NoCadence, Authentic}},
		{"plagal IV-I", Key{C, Major}, []int{4, 1}, []Cadence{NoCadence, Plagal}},
		{"half ends on V", Key{C, Major}, []int{1, 4, 5}, []Cadence{NoCadence, NoCadence, Half}},
		{"deceptive V-vi", Key{C, Major}, []int{5, 6}, []Cadence{NoCadence, Deceptive}},
		// A minor v has no leading tone, so it makes no deceptive, authentic
		// or half cadence (ADR-0012, decision 2); the harmonic-minor V does.
		{"minor v-VI is not deceptive", Key{A, Minor}, []int{5, 6}, []Cadence{NoCadence, NoCadence}},
		{"minor v-i is not authentic", Key{A, Minor}, []int{5, 1}, []Cadence{NoCadence, NoCadence}},
		{"minor i-v is not a half cadence", Key{A, Minor}, []int{1, 5}, []Cadence{NoCadence, NoCadence}},
		{"dorian v-i is not authentic", Key{D, Dorian}, []int{5, 1}, []Cadence{NoCadence, NoCadence}},
		{"I-V-vi-IV pop", Key{C, Major}, []int{1, 5, 6, 4}, []Cadence{NoCadence, Half, Deceptive, NoCadence}},
		{"V-V is not a half cadence twice", Key{C, Major}, []int{5, 5, 1}, []Cadence{NoCadence, NoCadence, Authentic}},
		{"ii-V-I jazz", Key{C, Major}, []int{2, 5, 1}, []Cadence{NoCadence, Half, Authentic}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tl, err := Analyze(harmonyOf(tt.key, tt.degrees...))
			if err != nil {
				t.Fatalf("Analyze: %v", err)
			}
			for i, h := range tl.Harmony {
				if h.Analysis.Cadence != tt.want[i] {
					t.Errorf("harmony[%d] (%s): cadence = %q, want %q", i, h.RomanNumeral, h.Analysis.Cadence, tt.want[i])
				}
			}
		})
	}
}

func TestAnalyzeExplanationMentionsFunctionAndCadence(t *testing.T) {
	tl, err := Analyze(harmonyOf(Key{C, Major}, 5, 1))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	first, second := tl.Harmony[0].Analysis.Explanation, tl.Harmony[1].Analysis.Explanation
	for _, want := range []string{"V", "G", "ドミナント"} {
		if !strings.Contains(first, want) {
			t.Errorf("explanation %q should mention %q", first, want)
		}
	}
	for _, want := range []string{"I", "C", "トニック", "正格終止"} {
		if !strings.Contains(second, want) {
			t.Errorf("explanation %q should mention %q", second, want)
		}
	}
}

func TestAnalyzeDoesNotMutateInput(t *testing.T) {
	in := harmonyOf(Key{C, Major}, 1, 5)
	if _, err := Analyze(in); err != nil {
		t.Fatal(err)
	}
	for i, h := range in.Harmony {
		if h.Analysis != nil {
			t.Fatalf("input harmony[%d] was mutated", i)
		}
	}
}

func TestAnalyzeRejectsNonDiatonicChord(t *testing.T) {
	tl := harmonyOf(Key{C, Major}, 1)
	tl.Harmony[0].Chord = Chord{Root: Cs, Quality: MajorTriad}
	_, err := Analyze(tl)
	if err == nil || !strings.Contains(err.Error(), "harmony[0]") || !strings.Contains(err.Error(), "not diatonic") {
		t.Fatalf("want non-diatonic error, got %v", err)
	}
}

func TestAnalyzeEmptyHarmony(t *testing.T) {
	tl := NewTimeline(Key{C, Major}, 100)
	out, err := Analyze(tl)
	if err != nil || len(out.Harmony) != 0 {
		t.Fatalf("empty harmony should analyse to empty, got %v %v", out.Harmony, err)
	}
}

func TestHarmonicMinorDominantStillCadences(t *testing.T) {
	tl := analysed(t, Key{A, Minor}, 4, "i", "V", "VI", "iv", "V7", "i")
	got := []Cadence{}
	for _, h := range tl.Harmony {
		got = append(got, h.Analysis.Cadence)
	}
	want := []Cadence{NoCadence, Half, Deceptive, NoCadence, Half, Authentic}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("harmony[%d] %s: cadence %q, want %q", i, tl.Harmony[i].RomanNumeral, got[i], want[i])
		}
	}
}

func TestMinorSubtonicTakesItsFunctionFromTheNextChord(t *testing.T) {
	tests := []struct {
		labels []string
		at     int
		want   string
	}{
		{[]string{"i", "VII", "III", "VI"}, 1, "V/III の働き"},
		{[]string{"i", "iv", "VII", "i"}, 2, "弱いドミナント"},
		{[]string{"i", "VII", "VI", "V"}, 1, "下りていく線の途中（経過）"},
	}
	for _, tt := range tests {
		tl := analysed(t, Key{A, Minor}, 4, tt.labels...)
		if e := tl.Harmony[tt.at].Analysis.Explanation; !strings.Contains(e, tt.want) {
			t.Errorf("%v: VII explanation lacks %q: %s", tt.labels, tt.want, e)
		}
	}
	// Only in minor: dorian's VII keeps its modal reading.
	tl := analysed(t, Key{D, Dorian}, 4, "i", "VII", "i")
	if strings.Contains(tl.Harmony[1].Analysis.Explanation, "弱いドミナント") {
		t.Errorf("dorian VII: %s", tl.Harmony[1].Analysis.Explanation)
	}
}
