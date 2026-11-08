package evals

import (
	"context"
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/ideas"
	"github.com/hassaku63/middy/backend/internal/llm"
	"github.com/hassaku63/middy/backend/internal/llm/fake"
	"github.com/hassaku63/middy/backend/internal/music"
)

func TestCasesLoadAndAreWellFormed(t *testing.T) {
	cases, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) < 20 {
		t.Fatalf("want at least 20 cases, got %d", len(cases))
	}
	seen := map[string]bool{}
	for _, c := range cases {
		if c.ID == "" || c.Image == "" || seen[c.ID] {
			t.Errorf("bad or duplicate case %+v", c)
		}
		seen[c.ID] = true
		if problem := judgeMode(c.Mode, music.Major); strings.Contains(problem, "unknown") {
			t.Errorf("%s: mode %q", c.ID, c.Mode)
		}
		if c.ModeInAny != "" && strings.Contains(judgeMode(c.ModeInAny, music.Major), "unknown") {
			t.Errorf("%s: modeInAny %q", c.ID, c.ModeInAny)
		}
		for _, w := range c.Wants {
			if _, err := music.LiteralChord(music.C, w); err != nil {
				t.Errorf("%s: want %q does not read as a chord: %v", c.ID, w, err)
			}
		}
		if c.TempoMin > 0 && c.TempoMax > 0 && c.TempoMin > c.TempoMax {
			t.Errorf("%s: tempo band inverted", c.ID)
		}
	}
}

func result() ideas.Result {
	return ideas.Result{Reply: "ok", Candidates: []ideas.Candidate{{
		Timeline: music.Timeline{},
	}}}
}

func TestJudge(t *testing.T) {
	mk := func(mode music.Mode, tempo float64) ideas.Result {
		r := result()
		r.Candidates[0].Title = "t"
		r.Candidates[0].Key = music.Key{Tonic: music.C, Mode: mode}
		r.Candidates[0].TempoBPM = tempo
		return r
	}
	tests := []struct {
		name string
		c    Case
		res  ideas.Result
		pass bool
		want string
	}{
		{"minor slow ok", Case{Mode: "minor", TempoMax: 100}, mk(music.Minor, 72), true, ""},
		{"wrong mode", Case{Mode: "minor"}, mk(music.Major, 72), false, "expected minor"},
		{"too fast", Case{Mode: "either", TempoMax: 90}, mk(music.Minor, 120), false, "above 90"},
		{"too slow", Case{Mode: "major", TempoMin: 110}, mk(music.Major, 80), false, "below 110"},
		{"either accepts both", Case{Mode: "either"}, mk(music.Major, 80), true, ""},
		{"no candidates", Case{Mode: "major"}, ideas.Result{Reply: "x"}, false, "no candidates"},
		{"empty reply", Case{Mode: "major"}, func() ideas.Result { r := mk(music.Major, 80); r.Reply = " "; return r }(), false, "empty reply"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Judge(tt.c, tt.res)
			if v.Pass != tt.pass {
				t.Fatalf("pass = %v, problems = %v", v.Pass, v.Problems)
			}
			if tt.want != "" && !strings.Contains(strings.Join(v.Problems, ";"), tt.want) {
				t.Fatalf("problems %v should mention %q", v.Problems, tt.want)
			}
		})
	}
}

// TestRunnerAgainstTheFake exercises the whole runner path with the
// unscripted fake, so the harness itself cannot rot unnoticed. The fake
// only knows a few mood words, so most cases are expected to fail; the
// rainy and sunny ones must pass.
func TestRunnerAgainstTheFake(t *testing.T) {
	cases, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	verdicts := Run(context.Background(), ideas.New(fake.New()), cases)
	byID := map[string]Verdict{}
	for _, v := range verdicts {
		byID[v.Case.ID] = v
	}
	for _, id := range []string{"rainy-sunday", "night-walk"} {
		if !byID[id].Pass {
			t.Errorf("%s should pass with the fake: %v", id, byID[id].Problems)
		}
	}
	report := Report(verdicts)
	if !strings.Contains(report, "| rainy-sunday |") || !strings.Contains(report, "passed") {
		t.Fatalf("report malformed:\n%s", report)
	}
}

var _ llm.Client = (*fake.Client)(nil)

func TestJudgeFeel(t *testing.T) {
	tests := []struct{ want, got, problem string }{
		{"", "", ""},
		{"swing", "swing", ""},
		{"straight", "", ""},
		{"swing", "", "expected feel swing, got straight"},
	}
	for _, tt := range tests {
		if got := judgeFeel(tt.want, tt.got); got != tt.problem {
			t.Errorf("judgeFeel(%q, %q) = %q, want %q", tt.want, tt.got, got, tt.problem)
		}
	}
}

func TestWantsAreMatchedBySound(t *testing.T) {
	maj := func(root music.PitchClass) music.Chord { return music.Chord{Root: root, Quality: music.MajorTriad} }
	min := func(root music.PitchClass) music.Chord { return music.Chord{Root: root, Quality: music.MinorTriad} }
	tests := []struct {
		name   string
		wants  []string
		tonic  music.PitchClass
		chords []music.Chord
		hit    bool
	}{
		{"bVII in E major is D", []string{"bVII"}, music.E, []music.Chord{maj(music.E), maj(music.D)}, true},
		{"G mixolydian's VII is its bVII", []string{"bVII"}, music.G, []music.Chord{maj(music.G), maj(music.F)}, true},
		{"the dorian IV is not the minor iv", []string{"IV"}, music.D, []music.Chord{min(music.D), min(music.G)}, false},
		{"the dorian IV", []string{"IV"}, music.D, []music.Chord{min(music.D), maj(music.G)}, true},
		{"I7 is a dominant seventh, not Imaj7", []string{"I7"}, music.As,
			[]music.Chord{{Root: music.As, Quality: music.Major7}}, false},
		{"no wants, no hit", nil, music.C, []music.Chord{maj(music.C)}, false},
	}
	for _, tt := range tests {
		if got := wantsHit(tt.wants, tt.tonic, tt.chords); got != tt.hit {
			t.Errorf("%s: wantsHit = %v, want %v", tt.name, got, tt.hit)
		}
	}
}

func TestJudgeModeInAnyLooksAtEveryCandidate(t *testing.T) {
	res := ideas.Result{Reply: "ok", Candidates: []ideas.Candidate{{}, {}}}
	res.Candidates[0].Key = music.Key{Tonic: music.D, Mode: music.Major}
	res.Candidates[1].Key = music.Key{Tonic: music.E, Mode: music.Dorian}
	c := Case{Mode: "major/dorian/mixolydian", ModeInAny: "dorian/mixolydian"}
	if v := Judge(c, res); !v.Pass {
		t.Fatalf("the second candidate is dorian: %v", v.Problems)
	}
	res.Candidates[1].Key.Mode = music.Minor
	if v := Judge(c, res); v.Pass || !strings.Contains(strings.Join(v.Problems, ";"), "no candidate in dorian/mixolydian") {
		t.Fatalf("no candidate is modal, want a problem: %v", v.Problems)
	}
}

func TestBreadthReportCountsVarietyAndWishes(t *testing.T) {
	v := func(id string, prog ...string) Verdict {
		return Verdict{Case: Case{ID: id}, Pass: true, Key: "A minor", Progression: prog, Labels: prog}
	}
	verdicts := []Verdict{
		v("a", "i", "VI", "III", "VII"),
		v("b", "i", "VI", "III", "VII"),
		v("c", "i", "iv", "V7"),
	}
	verdicts[2].Key = "D minor"
	verdicts[2].Case.Wants = []string{"IV"}
	verdicts[2].Unconstrained = "ドリアンの IV を\n使いたい | 本当は"
	verdicts[1].Retry = "music: \"IV\" would sound as Gm (iv) in D minor; write \"iv\""
	report := Report(verdicts)
	for _, want := range []string{
		"| a | ", // every row carries its progression (BL-090)
		"| i VI III VII |",
		"進行の種類: 2（3 件中）",
		"最も多い進行: `i VI III VII`（2 回）",
		"調の種類: 2",
		"使った語彙（度数 × 七の和音の有無）: 6 / 14", // i VI III VII iv V7 → 1,6,3,7,4 triads + 5 seventh
		"期待した語彙（wants）の命中: 0 / 1",
		"wants ✗ (IV)",
		"| c | ドリアンの IV を 使いたい / 本当は |",
		"retry: music: \"IV\" would sound as Gm (iv) in D minor",
		"再試行が要った件数: 1",
	} {
		if !strings.Contains(report, want) {
			t.Errorf("report lacks %q:\n%s", want, report)
		}
	}
}

func TestJudgeModeAcceptsAlternatives(t *testing.T) {
	if judgeMode("dorian/mixolydian", music.Mixolydian) != "" {
		t.Error("mixolydian satisfies dorian/mixolydian")
	}
	if judgeMode("dorian/mixolydian", music.Minor) != "expected dorian/mixolydian, got minor" {
		t.Error("minor does not")
	}
	if !strings.Contains(judgeMode("lydian", music.Major), "unknown mode") {
		t.Error("unsupported modes are reported")
	}
}
