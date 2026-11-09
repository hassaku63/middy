package evals

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/hassaku63/middy/backend/internal/ideas"
	"github.com/hassaku63/middy/backend/internal/llm"
)

// The independent reviewer (BL-172, ADR-0034 decision 4): a second model
// call that sees only what the learner would see of one answer (the
// request, the reply, each candidate's chords, the engine's explanations
// and the model's reasons) and none of the conversation that produced it.
// It counts the errors of ADR-0034 decision 3 that the deterministic
// check (ideas/claims.go) cannot read: misused progression names and
// wrong variant relations above all. It runs only in the evaluation; its
// findings are reported, never failed on.

// Error categories of ADR-0034, decision 3, as the reviewer numbers them.
var reviewCategories = []string{
	1: "和音名",
	2: "前後の和音",
	3: "呼び名",
	4: "派生関係",
	5: "食い違い",
}

// Finding is one error the reviewer reports.
type Finding struct {
	Category  int    `json:"category"`
	Candidate int    `json:"candidate"` // 1-based; 0 for the reply
	Quote     string `json:"quote"`
	Why       string `json:"why"`
}

// Review is the reviewer's verdict on one answer. Err is set when the
// review could not be had (the model failed, or its answer did not parse).
type Review struct {
	Findings []Finding
	Err      string
}

const reviewerSystem = `あなたは音楽理論の査読者です。作曲と音楽理論を学ぶアプリが出した 1 つの答えを読み、学ぶ人を誤らせる理論の誤りだけを数えます。答えを作った会話の文脈は持っていません。渡されたものだけで判断してください。

渡されるもの: ユーザーの依頼、アプリの返事、候補ごとの調・和音の並び（ローマ数字と実際に鳴る和音名）・アプリの理由の文・エンジンの解説。和音の並びとエンジンの解説は、実際に鳴るものです。

誤りとして数えるのは次の 5 つだけです。
1 和音名: 文に書いた和音名・度数が、その調で実際に鳴る和音と違う（C major の subV7/vi を F7 と書く、など）。
2 前後の和音: 「X へ解決する・進む・戻る」などの主張が、実際の並びと合わない（次にない和音へ進むと書く。最後の和音の次は最初の和音）。
3 呼び名: 名の知れた進行の呼び名（王道進行、丸サ進行、カノン進行、小室進行、循環など）を、その呼び名を決めている和音の並びを含まない進行に使う。並びを変えた版を「〜をもとに」と断って呼ぶのは誤りではない。呼び名の定義に揺れがある場合は、広く使われる定義のどれかに合えば誤りとしない。
4 派生関係: 候補どうしの関係（「候補 1 の V7 を V7/vi に替えた版」など）が、実際の差と合わない。
5 食い違い: 1 つの答えの中で、文どうし、または文とエンジンの解説が矛盾する。

誤りとして数えないもの: 同じ依頼から違う調・テンポ・進行が出ること、アレンジ、慣例的ではないが音楽的に成り立つ選択、印象の言葉（「都会的」「切ない」など）、好みの違い。確信が持てないものは数えない。

見つけた誤りごとに、category（1〜5）、candidate（候補の番号。返事なら 0）、quote（問題の文の短い引用）、why（なぜ誤りか、1 文）を返してください。誤りがなければ findings は空の配列です。`

var reviewSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "findings": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "category": {"type": "integer", "enum": [1, 2, 3, 4, 5]},
          "candidate": {"type": "integer"},
          "quote": {"type": "string"},
          "why": {"type": "string"}
        },
        "required": ["category", "candidate", "quote", "why"],
        "additionalProperties": false
      }
    }
  },
  "required": ["findings"],
  "additionalProperties": false
}`)

// reviewInput writes what the reviewer sees of one answer.
func reviewInput(image string, res ideas.Result) string {
	var b strings.Builder
	fmt.Fprintf(&b, "依頼: %s\n返事: %s\n", image, res.Reply)
	for i, c := range res.Candidates {
		fmt.Fprintf(&b, "\n候補 %d: %s（%s", i+1, c.Title, c.Key.String())
		if c.Meter != "" {
			fmt.Fprintf(&b, "、%s", c.Meter)
		}
		fmt.Fprintf(&b, "）\n理由: %s\n和音: ", c.Reason)
		chords := make([]string, 0, len(c.Timeline.Harmony))
		for _, h := range c.Timeline.Harmony {
			chords = append(chords, fmt.Sprintf("%s（%s）", h.RomanNumeral, h.Symbol))
		}
		b.WriteString(strings.Join(chords, " → "))
		b.WriteString("\nエンジンの解説:\n")
		for j, h := range c.Timeline.Harmony {
			if h.Analysis != nil {
				fmt.Fprintf(&b, "  %d. %s\n", j+1, h.Analysis.Explanation)
			}
		}
		if o := c.Timeline.Overview; o != nil {
			fmt.Fprintf(&b, "  全体: %s\n  流れ: %s\n", o.Whole, o.Flow)
		}
	}
	return b.String()
}

// ReviewAll asks the reviewer about every answered verdict, at most
// Concurrency at once, and stores each review on its verdict. Verdicts
// without an answer are skipped; when ctx ends, the rest stay unreviewed.
func ReviewAll(ctx context.Context, client llm.Client, verdicts []Verdict) {
	reviewAll(ctx, client, verdicts, Concurrency)
}

func reviewAll(ctx context.Context, client llm.Client, verdicts []Verdict, limit int) {
	sem := make(chan struct{}, limit)
	var wg sync.WaitGroup
	for i := range verdicts {
		if verdicts[i].ReviewInput == "" {
			continue
		}
		if ctx.Err() != nil {
			return
		}
		sem <- struct{}{}
		wg.Add(1)
		go func(v *Verdict) {
			defer wg.Done()
			defer func() { <-sem }()
			r := review(ctx, client, v.ReviewInput)
			v.Review = &r
		}(&verdicts[i])
	}
	wg.Wait()
}

func review(ctx context.Context, client llm.Client, input string) Review {
	resp, err := client.Complete(ctx, llm.Request{
		System:       reviewerSystem,
		Messages:     []llm.Message{{Role: llm.RoleUser, Content: input}},
		OutputSchema: reviewSchema,
		MaxTokens:    2048,
	})
	if err != nil {
		return Review{Err: err.Error()}
	}
	var out struct {
		Findings []Finding `json:"findings"`
	}
	if err := json.Unmarshal([]byte(resp.Text), &out); err != nil {
		return Review{Err: "unreadable review: " + err.Error()}
	}
	kept := out.Findings[:0]
	for _, f := range out.Findings {
		if f.Category >= 1 && f.Category < len(reviewCategories) {
			kept = append(kept, f)
		}
	}
	return Review{Findings: kept}
}

// reviewNote is the table's note for one verdict: the findings counted by
// category, or why there is no review.
func reviewNote(r *Review) string {
	switch {
	case r == nil:
		return ""
	case r.Err != "":
		return " · review: not run (" + shorten(oneLine(r.Err), 80) + ")"
	case len(r.Findings) == 0:
		return ""
	}
	counts := make([]int, len(reviewCategories))
	for _, f := range r.Findings {
		counts[f.Category]++
	}
	var parts []string
	for c := 1; c < len(reviewCategories); c++ {
		if counts[c] > 0 {
			parts = append(parts, fmt.Sprintf("%s %d", reviewCategories[c], counts[c]))
		}
	}
	return " · review: " + strings.Join(parts, ", ")
}

// reviewReport sums the reviews and lists every finding, for the owner
// to judge whether the reviewer is right (ADR-0034: its findings are
// hypotheses, not failures).
func reviewReport(verdicts []Verdict) string {
	reviewed, notRun := 0, 0
	counts := make([]int, len(reviewCategories))
	var lines []string
	for _, v := range verdicts {
		switch {
		case v.Review == nil:
			if v.ReviewInput != "" {
				notRun++
			}
			continue
		case v.Review.Err != "":
			notRun++
			continue
		}
		reviewed++
		for _, f := range v.Review.Findings {
			counts[f.Category]++
			where := "返事"
			if f.Candidate > 0 {
				where = fmt.Sprintf("候補 %d", f.Candidate)
			}
			lines = append(lines, fmt.Sprintf("| %s | %s | %s | %s | %s |", v.Case.ID, where, reviewCategories[f.Category], shorten(oneLine(f.Quote), 80), shorten(oneLine(f.Why), 120)))
		}
	}
	if reviewed == 0 && notRun == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n### 独立したレビュアーの指摘（ADR-0034 の誤りの区分）\n\n")
	var sums []string
	for c := 1; c < len(reviewCategories); c++ {
		sums = append(sums, fmt.Sprintf("%s %d", reviewCategories[c], counts[c]))
	}
	fmt.Fprintf(&b, "- レビューしたケース: %d（レビューできなかったもの: %d）\n- 区分ごとの件数: %s\n", reviewed, notRun, strings.Join(sums, "、"))
	if len(lines) > 0 {
		b.WriteString("\n| case | どこ | 区分 | 引用 | 理由 |\n| --- | --- | --- | --- | --- |\n")
		b.WriteString(strings.Join(lines, "\n"))
		b.WriteString("\n")
	}
	return b.String()
}
