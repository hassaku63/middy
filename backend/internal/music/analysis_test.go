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
		{"minor deceptive v-VI", Key{A, Minor}, []int{5, 6}, []Cadence{NoCadence, Deceptive}},
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
