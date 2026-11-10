package evals

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/hassaku63/middy/backend/internal/ideas"
	"github.com/hassaku63/middy/backend/internal/intent"
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
			if _, _, err := music.LiteralSlash(music.C, w); err != nil {
				t.Errorf("%s: want %q does not read as a chord: %v", c.ID, w, err)
			}
		}
		if c.TempoMin > 0 && c.TempoMax > 0 && c.TempoMin > c.TempoMax {
			t.Errorf("%s: tempo band inverted", c.ID)
		}
		if len(c.Tags) == 0 {
			t.Errorf("%s: no tags", c.ID)
		}
		for _, tag := range c.Tags {
			if !slices.Contains(Tags, tag) {
				t.Errorf("%s: unknown tag %q", c.ID, tag)
			}
		}
	}
	for _, tag := range Tags {
		if picked, _ := Select(cases, []string{tag}); len(picked) == 0 {
			t.Errorf("tag %q names no case", tag)
		}
	}
}

// TestSelectNarrowsTheRun (#334): a run can be narrowed to the cases with
// some tags, in their order; no tags is every case; a misspelt tag is an
// error rather than an empty run.
func TestSelectNarrowsTheRun(t *testing.T) {
	cases, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	all, err := Select(cases, nil)
	if err != nil || len(all) != len(cases) {
		t.Fatalf("no tags keeps every case: %d of %d, %v", len(all), len(cases), err)
	}
	picked, err := Select(cases, []string{"meter", "slash"})
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, c := range picked {
		ids = append(ids, c.ID)
	}
	if got := strings.Join(ids, " "); got != "graduation-bass minor-cliche waltz-ball jazz-waltz rocking-lullaby irish-jig" {
		t.Errorf("meter and slash picked %s", got)
	}
	if got := scopeNote([]string{"meter", "slash"}, 6, 44); !strings.Contains(got, "タグ meter, slash の 6 / 44 件") {
		t.Errorf("scope note: %q", got)
	}
	if got := scopeNote(nil, 44, 44); !strings.Contains(got, "全件（44 件）") {
		t.Errorf("scope note: %q", got)
	}
	if _, err := Select(cases, []string{"metre"}); err == nil || !strings.Contains(err.Error(), `unknown tag "metre"`) {
		t.Errorf("a misspelt tag must be an error, got %v", err)
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

// TestFakeAnswersEveryImageWithAValidIntent (review): the fake keeps
// keyless runs working, so every eval image (and its "long" variant) gets
// an answer the service can render, and asking again gives the same
// answer: a keyword's edit must not leak into the shared samples.
func TestFakeAnswersEveryImageWithAValidIntent(t *testing.T) {
	cases, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	svc := ideas.New(fake.New())
	ask := func(text string) ideas.Result {
		t.Helper()
		res, err := svc.Imagine(context.Background(), []ideas.Message{{Role: llm.RoleUser, Content: text}})
		if err != nil {
			t.Fatalf("%q: %v", text, err)
		}
		return res
	}
	for _, c := range cases {
		for _, text := range []string{c.Image, "長い" + c.Image} {
			first := ask(text)
			for range 2 {
				again := ask(text)
				if fmt.Sprint(again.Candidates[0].Progression, again.Candidates[0].Feel, again.Candidates[0].Title) !=
					fmt.Sprint(first.Candidates[0].Progression, first.Candidates[0].Feel, first.Candidates[0].Title) {
					t.Errorf("%q: answer changed between requests: %v then %v", text, first.Candidates[0].Progression, again.Candidates[0].Progression)
				}
			}
		}
	}
}

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
		harmony := make([]music.HarmonyEvent, len(tt.chords))
		for i, c := range tt.chords {
			harmony[i] = music.HarmonyEvent{Chord: c}
		}
		if got := wantsHit(tt.wants, tt.tonic, harmony); got != tt.hit {
			t.Errorf("%s: wantsHit = %v, want %v", tt.name, got, tt.hit)
		}
	}
	// A wanted slash chord must sound over its bass (BL-115); a want
	// without a slash matches the chord over any bass.
	b, gs := music.B, music.Gs
	cOverB := []music.HarmonyEvent{{Chord: maj(music.C)}, {Chord: maj(music.C), Bass: &b}}
	if !wantsHit([]string{"I/7"}, music.C, cOverB) || !wantsHit([]string{"I"}, music.C, cOverB[1:]) {
		t.Error("I/7 should match C/B, and I should match C over any bass")
	}
	if wantsHit([]string{"I/7"}, music.C, cOverB[:1]) {
		t.Error("I/7 should not match a plain C")
	}
	amOverGs := []music.HarmonyEvent{{Chord: min(music.A), Bass: &gs}}
	if !wantsHit([]string{"i/7"}, music.A, amOverGs) || wantsHit([]string{"i/b7"}, music.A, amOverGs) {
		t.Error("in A, i/7 is Am/G# and i/b7 is Am/G")
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
		return Verdict{Case: Case{ID: id}, Pass: true, Key: "A minor", Progression: prog, Chords: prog, Labels: prog}
	}
	verdicts := []Verdict{
		v("a", "i", "VI", "III", "VII"),
		v("b", "i", "VI", "III", "VII"),
		v("c", "i", "iv", "V7"),
	}
	verdicts[2].Key = "D minor"
	verdicts[2].Case.Wants = []string{"IV"}
	verdicts[0].Case.Wants = []string{"bVII"}
	verdicts[0].WantsIn = 2
	verdicts[2].Unconstrained = "ドリアンの IV を\n使いたい | 本当は"
	verdicts[1].Retry = "music: \"IV\" would sound as Gm (iv) in D minor; write \"iv\""
	report := Report(verdicts)
	for _, want := range []string{
		"| a | ", // every row carries its progression (BL-090)
		"| i VI III VII |",
		"進行の種類: 2（答えのあった 3 件中）",
		"最も多い進行: `i VI III VII`（2 回）",
		"調の種類: 2",
		"使った語彙（度数 × 七の和音の有無）: 6 / 14", // i VI III VII iv V7 → 1,6,3,7,4 triads + 5 seventh
		"期待した語彙（wants）の命中: 最初の候補で 0 / 2、どれかの候補で 1 / 2",
		"wants ✓ in candidate 2",
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

func TestJudgeFindsWantsInLaterCandidates(t *testing.T) {
	render := func(key music.Key, labels ...string) ideas.Candidate {
		var prog music.Progression
		for _, l := range labels {
			step, err := music.StepFromLabel(key, l, 4)
			if err != nil {
				t.Fatal(err)
			}
			prog = append(prog, step)
		}
		tl, err := music.FromProgression(key, prog, 100)
		if err != nil {
			t.Fatal(err)
		}
		var c ideas.Candidate
		c.Key, c.Progression, c.Timeline = key, labels, tl
		return c
	}
	e := music.Key{Tonic: music.E, Mode: music.Major}
	res := ideas.Result{Reply: "ok", Candidates: []ideas.Candidate{
		render(e, "I", "V", "vi", "IV"),
		render(e, "I", "bVII", "IV", "I"),
	}}
	v := Judge(Case{Mode: "major", Wants: []string{"bVII"}}, res)
	if v.WantsHit || v.WantsIn != 2 {
		t.Fatalf("bVII sounds in the second candidate: hit=%v in=%d", v.WantsHit, v.WantsIn)
	}
}

func TestJudgeMeter(t *testing.T) {
	res := result()
	res.Candidates[0].Key = music.Key{Tonic: music.F, Mode: music.Major}
	if v := Judge(Case{Mode: "either", Meter: "3/4"}, res); v.Pass {
		t.Fatal("an omitted meter is 4/4, not 3/4")
	}
	res.Candidates[0].Meter = "3/4"
	if v := Judge(Case{Mode: "either", Meter: "3/4"}, res); !v.Pass {
		t.Fatalf("3/4 expected and given: %v", v.Problems)
	}
	if !strings.Contains(Report([]Verdict{Judge(Case{ID: "w", Mode: "either"}, res)}), "BPM, 3/4)") {
		t.Error("the report marks a 3/4 first candidate")
	}
}

// TestSilentWordsInReasons (#128): a reason that names an instrument middy
// never plays, or a melody, is reported (not failed) with the words found.
func TestSilentWordsInReasons(t *testing.T) {
	tests := []struct {
		reasons []string
		want    string
	}{
		{[]string{"VII が導音のない明るさを生み、ホイッスルやフィドルが似合う響きに。"}, "フィドル/ホイッスル"},
		{[]string{"ストロークの 8 ビートで軽快に。", "メロディが映える穏やかな I–IV。"}, "メロディ"},
		{[]string{"i と IV の往復でドリアンの明るさを出す。"}, ""},
	}
	for _, tt := range tests {
		if got := strings.Join(silentWords(tt.reasons), "/"); got != tt.want {
			t.Errorf("silentWords(%q) = %q, want %q", tt.reasons, got, tt.want)
		}
	}
	report := Report([]Verdict{{Case: Case{ID: "muji"}, Pass: true, Silent: []string{"フィドル"}}})
	for _, want := range []string{"· reason names フィドル", "説明が鳴らないものに触れた件数: 1"} {
		if !strings.Contains(report, want) {
			t.Errorf("report lacks %q:\n%s", want, report)
		}
	}
}

// TestReportNamesTheFirstPatterns (#131): the report shows the first
// candidate's accompaniment and drums, so a run can be read for whether a
// Celtic image got the Celtic patterns.
func TestReportNamesTheFirstPatterns(t *testing.T) {
	res := result()
	res.Candidates[0].Accompaniment, res.Candidates[0].Drums = "reel", "bodhran"
	if got := Report([]Verdict{Judge(Case{ID: "c", Mode: "either"}, res)}); !strings.Contains(got, "BPM, reel/bodhran)") {
		t.Errorf("report should name the patterns:\n%s", got)
	}
	res.Candidates[0].Feel = "shuffle"
	if got := Report([]Verdict{Judge(Case{ID: "c", Mode: "either"}, res)}); !strings.Contains(got, "BPM, reel/bodhran, shuffle)") {
		t.Errorf("report should name a feel that is not straight (#321):\n%s", got)
	}
	res.Candidates[0].Feel = "straight"
	res.Candidates[0].Accompaniment, res.Candidates[0].Drums = "", ""
	if got := Report([]Verdict{Judge(Case{ID: "c", Mode: "either"}, res)}); strings.Contains(got, "BPM, ") {
		t.Errorf("block chords without drums add nothing:\n%s", got)
	}
}

// TestReportListsRequestedTensions (BL-202): candidates that ask for
// tensions in place of their style's are listed with their position.
func TestReportListsRequestedTensions(t *testing.T) {
	res := result()
	if got := Report([]Verdict{Judge(Case{ID: "c", Mode: "either"}, res)}); strings.Contains(got, "テンションを頼んだ候補") {
		t.Errorf("no section without requested tensions:\n%s", got)
	}
	res.Candidates[0].Tensions, res.Candidates[0].Style = "9-13", "jazz"
	if got := Report([]Verdict{Judge(Case{ID: "c", Mode: "either"}, res)}); !strings.Contains(got, "| c | 1: 9-13 (jazz) |") {
		t.Errorf("report should list the candidate that asked for tensions:\n%s", got)
	}
}

// TestReportShowsPerChordBeats (BL-109): when the first candidate gives
// each chord a length, the progression column shows it ("i:8").
func TestReportShowsPerChordBeats(t *testing.T) {
	res := result()
	res.Candidates[0].Progression = []string{"i", "VI", "III", "VII"}
	res.Candidates[0].Beats = []int{8, 4, 2, 2}
	v := Judge(Case{ID: "c", Mode: "either"}, res)
	if !strings.HasSuffix(v.Progression[0], ":8") || !strings.HasSuffix(v.Progression[1], ":4") {
		t.Errorf("progression should carry the lengths: %v", v.Progression)
	}
}

// TestJudgeDrums (BL-081): a case can name the drums its first candidate
// must use; "none" stands for no drums.
func TestJudgeDrums(t *testing.T) {
	tests := []struct{ want, got, problem string }{
		{"", "eight-beat", ""},
		{"jazz-ride", "jazz-ride", ""},
		{"jazz-ride", "eight-beat", "expected drums jazz-ride, got eight-beat"},
		{"jazz-ride/none", "", ""},
		{"jazz-ride", "", "expected drums jazz-ride, got none"},
	}
	for _, tt := range tests {
		if got := judgeDrums(tt.want, tt.got); got != tt.problem {
			t.Errorf("judgeDrums(%q, %q) = %q, want %q", tt.want, tt.got, got, tt.problem)
		}
	}
}

// TestJudgeStyle (BL-082, #129): a case can name the style its first
// candidate must have, and the report shows the style with the patterns.
func TestJudgeStyle(t *testing.T) {
	res := result()
	res.Candidates[0].Style, res.Candidates[0].Accompaniment, res.Candidates[0].Drums = "celtic", "reel", "bodhran"
	if v := Judge(Case{ID: "c", Mode: "either", Style: "celtic"}, res); !v.Pass {
		t.Errorf("celtic expected and given: %v", v.Problems)
	}
	if v := Judge(Case{ID: "c", Mode: "either", Style: "jazz"}, res); v.Pass || !strings.Contains(strings.Join(v.Problems, ";"), "expected style jazz, got celtic") {
		t.Errorf("jazz expected, celtic given: %v", v.Problems)
	}
	if got := Report([]Verdict{Judge(Case{ID: "c", Mode: "either"}, res)}); !strings.Contains(got, "BPM, celtic: reel/bodhran)") {
		t.Errorf("report should name the style and patterns:\n%s", got)
	}
}

// TestReportShowsPushes (BL-085): a pushed chord is marked in the
// progression column, ^8 for an eighth and ^16 for a sixteenth.
func TestReportShowsPushes(t *testing.T) {
	res := result()
	res.Candidates[0].Progression = []string{"i", "VI", "III", "VII"}
	res.Candidates[0].Pushes = []string{"none", "none", "sixteenth", "eighth"}
	v := Judge(Case{ID: "c", Mode: "either"}, res)
	if got := strings.Join(v.Progression, " "); got != "i VI III^16 VII^8" {
		t.Errorf("progression should mark the pushes: %q", got)
	}
}

// TestReportNamesGrids (BL-077): a grid is named in the patterns column
// with the drums it is based on, and a case's drums are judged by that.
func TestReportNamesGrids(t *testing.T) {
	res := result()
	res.Candidates[0].Accompaniment = "strum"
	res.Candidates[0].DrumGrid = &intent.GridSpec{Division: 2, Voices: []intent.GridLine{{Voice: "kick", Steps: "x......."}}, BasedOn: "eight-beat"}
	v := Judge(Case{ID: "c", Mode: "either", Drums: "eight-beat"}, res)
	if v.Patterns != "strum/grid<eight-beat" || len(v.Problems) > 0 {
		t.Errorf("patterns %q, problems %v", v.Patterns, v.Problems)
	}
}

// TestBreadthCountsChordsNotLengths (review): the same chords with other
// lengths or pushes are the same progression for the breadth metric.
func TestBreadthCountsChordsNotLengths(t *testing.T) {
	a, b := result(), result()
	a.Candidates[0].Progression = []string{"i", "VI", "III", "VII"}
	b.Candidates[0].Progression = []string{"i", "VI", "III", "VII"}
	b.Candidates[0].Beats = []int{8, 4, 2, 2}
	m := measureBreadth([]Verdict{Judge(Case{ID: "a", Mode: "either"}, a), Judge(Case{ID: "b", Mode: "either"}, b)})
	if len(m.progressions) != 1 {
		t.Errorf("one progression, counted %d: %v", len(m.progressions), m.progressions)
	}
}

// TestJudgingAGridOfItsOwn (review): a grid with no basedOn plays drums,
// so it does not pass a case that wants none.
func TestJudgingAGridOfItsOwn(t *testing.T) {
	own := intent.Candidate{DrumGrid: &intent.GridSpec{Division: 2}}
	if got := judgeDrums("none", gridBase(own)); got == "" {
		t.Error("a grid of its own passed a case that wants no drums")
	}
	based := intent.Candidate{DrumGrid: &intent.GridSpec{Division: 2, BasedOn: "eight-beat"}}
	if got := judgeDrums("eight-beat", gridBase(based)); got != "" {
		t.Errorf("a grid based on eight-beat: %s", got)
	}
}

// TestReportKeepsEachCaseOnOneRow (review): a problem or title with a
// newline or "|" must not split the Markdown table.
func TestReportKeepsEachCaseOnOneRow(t *testing.T) {
	v := Verdict{Case: Case{ID: "x"}, Title: "a|b", Problems: []string{"line one\nline two"}}
	report := Report([]Verdict{v})
	for _, line := range strings.Split(report, "\n") {
		if strings.HasPrefix(line, "line two") || strings.Contains(line, "a|b") {
			t.Fatalf("row split or unescaped:\n%s", report)
		}
	}
}

// TestRunAllWithNoLimitStillRuns (review): a limit of 0 must not deadlock.
func TestRunAllWithNoLimitStillRuns(t *testing.T) {
	cases := []Case{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	done := make(chan []Verdict, 1)
	go func() {
		done <- runAll(context.Background(), cases, 0, func(c Case) Verdict { return Verdict{Case: c, Pass: true} })
	}()
	select {
	case out := <-done:
		if len(out) != 3 || !out[2].Pass {
			t.Fatalf("verdicts = %+v", out)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runAll with limit 0 blocked")
	}
}

// TestUnjudgedCasesStayOutOfWantsAndBreadth (review): a case that never got
// an answer shows no "wants ✗" and is not counted in the breadth.
func TestUnjudgedCasesStayOutOfWantsAndBreadth(t *testing.T) {
	answered := Verdict{Case: Case{ID: "a", Wants: []string{"IV"}}, Pass: true, Progression: []string{"I", "IV"}, Chords: []string{"I", "IV"}, WantsHit: true}
	missing := Verdict{Case: Case{ID: "b", Wants: []string{"V"}}, Errored: true, Problems: []string{"not run"}}
	report := Report([]Verdict{answered, missing})
	for _, line := range strings.Split(report, "\n") {
		if strings.HasPrefix(line, "| b |") && strings.Contains(line, "wants") {
			t.Errorf("an unanswered case was judged on its wants: %s", line)
		}
	}
	if !strings.Contains(report, "答えのあった 1 件中") {
		t.Errorf("the breadth counts unanswered cases:\n%s", report)
	}
}
