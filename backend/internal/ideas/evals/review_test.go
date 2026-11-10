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

// marusaResult is the answer of #272 as the reviewer would see it: the
// royal road called the Marunouchi Sadistic skeleton.
func marusaResult(t *testing.T) ideas.Result {
	t.Helper()
	c := music.Key{Tonic: music.C, Mode: music.Major}
	var prog music.Progression
	for _, l := range []string{"IVmaj7", "V7", "iii7", "vi7"} {
		s, err := music.StepFromLabel(c, l, 4)
		if err != nil {
			t.Fatal(err)
		}
		prog = append(prog, s)
	}
	tl, err := music.FromProgression(c, prog, 96)
	if err != nil {
		t.Fatal(err)
	}
	if tl, err = music.Analyze(tl); err != nil {
		t.Fatal(err)
	}
	cand := ideas.Candidate{Timeline: tl}
	cand.Title, cand.Key, cand.Reason = "王道の丸サ進行", c, "C major で IVmaj7→V7→iii7→vi7 と回す、いわゆる「丸サ進行」の骨格です。"
	return ideas.Result{Reply: "丸サ進行ですね。", Candidates: []ideas.Candidate{cand}}
}

// The reviewer sees the request, the reply, each candidate's chords as
// they sound, the reason and the engine's explanations, and nothing of
// the conversation that made them.
func TestReviewInputShowsWhatTheLearnerSees(t *testing.T) {
	in := reviewInput("丸の内サディスティック風の進行", marusaResult(t))
	for _, want := range []string{"依頼: 丸の内サディスティック風の進行", "返事: 丸サ進行ですね。", "候補 1: 王道の丸サ進行（C major）", "理由: C major で", "IVmaj7（Fmaj7） → V7（G7） → iii7（Em7） → vi7（Am7）", "エンジンの解説:", "サブドミナント", "流れ:"} {
		if !strings.Contains(in, want) {
			t.Errorf("review input lacks %q:\n%s", want, in)
		}
	}
}

func TestReviewCountsFindingsByCategoryAndKeepsFailuresApart(t *testing.T) {
	verdicts := []Verdict{
		{Case: Case{ID: "marusa"}, ReviewInput: "a"},
		{Case: Case{ID: "calm"}, ReviewInput: "b"},
		{Case: Case{ID: "broken"}, ReviewInput: "c"},
		{Case: Case{ID: "unanswered"}}, // no answer: not reviewed
	}
	client := fake.New(
		llm.Response{Text: `{"findings":[{"category":3,"candidate":1,"quote":"丸サ進行の骨格","why":"IV–V–iii–vi は王道進行"},{"category":5,"candidate":0,"quote":"丸サ進行ですね","why":"候補は丸サを含まない"},{"category":9,"candidate":1,"quote":"x","why":"no such category"}]}`},
		llm.Response{Text: `{"findings":[]}`},
		llm.Response{Text: `not json`},
	)
	reviewAll(context.Background(), client, verdicts, 1)

	for i, req := range client.Requests() {
		if req.System != reviewerSystem || req.OutputSchema == nil || len(req.Messages) != 1 {
			t.Errorf("request %d is not a fresh, structured review: %+v", i, req)
		}
	}
	if got := len(verdicts[0].Review.Findings); got != 2 {
		t.Errorf("an unknown category must be dropped: %d findings", got)
	}
	if verdicts[3].Review != nil {
		t.Error("an unanswered case was reviewed")
	}
	if !strings.Contains(verdicts[2].Review.Err, "unreadable") {
		t.Errorf("an unreadable review is not an error: %+v", verdicts[2].Review)
	}

	if got := reviewNote(verdicts[0].Review); got != " · review: 呼び名 1, 食い違い 1" {
		t.Errorf("note = %q", got)
	}
	if got := reviewNote(verdicts[1].Review); got != "" {
		t.Errorf("a clean review needs no note: %q", got)
	}
	report := reviewReport(verdicts)
	for _, want := range []string{"レビューしたケース: 2（レビューできなかったもの: 1）", "呼び名 1", "食い違い 1", "和音名 0", "| marusa | 候補 1 | 呼び名 | 丸サ進行の骨格 |", "| marusa | 返事 | 食い違い |"} {
		if !strings.Contains(report, want) {
			t.Errorf("report lacks %q:\n%s", want, report)
		}
	}
}

func TestNoReviewNoSection(t *testing.T) {
	if got := reviewReport([]Verdict{{Case: Case{ID: "x"}}}); got != "" {
		t.Errorf("a run without reviews printed %q", got)
	}
}

// TestReviewRunsOnFullRuns (#334): the reviewer runs on a full run, not on
// one narrowed by tags, unless EVAL_REVIEW says otherwise; a skipped
// review is said once and not counted as a review that failed.
func TestReviewRunsOnFullRuns(t *testing.T) {
	for _, c := range []struct {
		tags []string
		env  string
		want bool
	}{
		{nil, "", true},
		{[]string{"meter"}, "", false},
		{[]string{"meter"}, "1", true},
		{nil, "0", false},
	} {
		if got, err := ReviewWanted(c.tags, c.env); err != nil || got != c.want {
			t.Errorf("tags %v, EVAL_REVIEW %q: %v %v, want %v", c.tags, c.env, got, err, c.want)
		}
	}
	if _, err := ReviewWanted(nil, "yes"); err == nil {
		t.Error("EVAL_REVIEW=yes must be an error")
	}
	verdicts := []Verdict{{Case: Case{ID: "a"}, Pass: true, ReviewInput: "依頼: …"}}
	note := SkipReview(verdicts)
	if report := Report(verdicts) + note; strings.Contains(report, "レビューできなかったもの") || !strings.Contains(report, "独立したレビュアーは走らせていない") {
		t.Errorf("a skipped review must be said, not counted as failed:\n%s", report)
	}
}
