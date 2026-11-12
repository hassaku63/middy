package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/hassaku63/middy/backend/internal/config"
	"github.com/hassaku63/middy/backend/internal/ideas/evals"
	"github.com/hassaku63/middy/backend/internal/llm"
	"github.com/hassaku63/middy/backend/internal/llm/anthropic"
	"github.com/hassaku63/middy/backend/internal/llm/fake"
)

// The conversation reviewer (#473, BL-315, ADR-0064): a model call that
// reads one whole conversation (what the user said each turn, and what
// came back) and counts where the answers did not follow the requests.
// It sees the conversation as the learner did, not the scripts'
// expectations, and it is the only part of cmd/converse that needs the
// key. Its findings are reported, never failed on.

// Kinds of the tolerance of ADR-0065: shown, never counted.
var withinKinds = []string{
	1: "程度が小さい",
	2: "一部だけ",
	3: "「全部」と言われたが一部",
	4: "初心者向けの単純化",
}

// Within is a place where the answer used the tolerance (ADR-0065).
type Within struct {
	Kind  int    `json:"kind"`
	Turn  int    `json:"turn"`
	Quote string `json:"quote"`
	Why   string `json:"why"`
}

// Categories of ADR-0064, as the reviewer numbers them.
var judgeCategories = []string{
	1: "反映されていない",
	2: "頼んでいないのに変わった",
	3: "指した例の取り違え",
	4: "返事が変更点を誤って言う",
}

const judgeSystem = `あなたは、作曲を学ぶアプリとの会話を読む査読者です。ユーザーはイメージを言葉で伝え、アプリは毎回、例（候補）を返します。ユーザーは例を聴いて、言い直して近づけていきます。会話全体を読み、アプリの答えがユーザーの言葉に従わなかったところだけを数えます。

渡されるもの: 回ごとのユーザーの言葉と、アプリの返事、候補ごとの調・拍子・テンポ・小節数・和音の並び・スタイルなどの型・理由の文。候補の中身は実際に鳴るものです。

数えるのは次の 4 つだけです。
1 反映されていない: その回でユーザーが頼んだ変更（拍子、長さ、明るさ、ジャンル、テンポ、和音など）が、候補のどれにもまったく入っていない。一部の候補にだけ、または頼んだものの一部だけ入っていれば数えない。
2 頼んでいないのに変わった: 前の回までに決まっていたこと（ユーザーが頼んだ拍子・長さ・ジャンル・和音など）が、ユーザーが変えるよう頼んでいないのに、すべての候補で失われた。
3 指した例の取り違え: ユーザーが「2 番目の例をもとに」のように特定の例を指したのに、別の例をもとにした。
4 返事が変更点を誤って言う: 返事や理由の文が「〜にしました」「〜を残しました」と言うのに、候補の中身がそうなっていない。

このアプリは作曲の初心者のためのものです。頼まれたものをどの程度採り入れるかには、幅を認めます。次の 4 つは数えず、findings ではなく within に書いてください（数えないが、どこで幅を使ったかを残す）。
1 程度が小さい: 「テンションを使って」に 1 か所だけ入れる、一部の候補にだけ入れる、など。
2 一部だけ: 複数挙げたもののうち一部だけ採る（「sus と add9 を使って」に sus だけ）。
3 「全部」と言われたが一部: 「全部使って」と明示されたのに一部だけ採る。
4 初心者向けの単純化: 和音や進行を、意図してやさしくする。

次のものは、どこにも書きません。
- 同じ依頼から違う調・テンポ・進行が出ること、アレンジ、好みの違い。
- 理論の細かな誤り（別の査読が見る）。
- 表せない語彙を、近い表現に置き換えたことを返事で断っている場合。
確信が持てないものは書かない。

区分 4（返事が変更点を誤って言う）は幅の外です。「sus と add9 を入れました」と言って sus だけなら、findings に数えます。

findings には、見つけたものごとに category（1〜4）、turn（何回目か）、candidate（候補の番号。返事なら 0）、quote（根拠になる言葉の短い引用）、why（なぜそう数えるか、1 文）を返してください。within には、kind（1〜4）、turn、quote、why を返してください。どちらも、なければ空の配列です。`

var judgeSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "findings": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "category": {"type": "integer", "enum": [1, 2, 3, 4]},
          "turn": {"type": "integer"},
          "candidate": {"type": "integer"},
          "quote": {"type": "string"},
          "why": {"type": "string"}
        },
        "required": ["category", "turn", "candidate", "quote", "why"],
        "additionalProperties": false
      }
    },
    "within": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "kind": {"type": "integer", "enum": [1, 2, 3, 4]},
          "turn": {"type": "integer"},
          "quote": {"type": "string"},
          "why": {"type": "string"}
        },
        "required": ["kind", "turn", "quote", "why"],
        "additionalProperties": false
      }
    }
  },
  "required": ["findings", "within"],
  "additionalProperties": false
}`)

// JudgeFinding is one place where an answer did not follow the request.
type JudgeFinding struct {
	Category  int    `json:"category"`
	Turn      int    `json:"turn"`
	Candidate int    `json:"candidate"` // 1-based; 0 for the reply
	Quote     string `json:"quote"`
	Why       string `json:"why"`
}

// Verdict is the reviewer's reading of one run. Err is set when there is
// none (the call failed, or its answer did not parse).
type Verdict struct {
	Run      string
	Findings []JudgeFinding
	Within   []Within
	Err      string
}

// judgeInput writes what the reviewer sees of one conversation.
func judgeInput(turns []Turn) string {
	var b strings.Builder
	for i, t := range turns {
		fmt.Fprintf(&b, "## %d 回目\nユーザー: %s\n", i+1, t.Sent)
		if t.Answer == nil {
			b.WriteString("（答えなし）\n\n")
			continue
		}
		fmt.Fprintf(&b, "返事: %s\n", t.Answer.Reply)
		for n, c := range t.Answer.Candidates {
			m := measure(c, "")
			fmt.Fprintf(&b, "候補 %d: %s（%s、%s、%s BPM、%s 小節）\n", n+1, c.Title, c.Key.String(), m.Meter,
				fmt.Sprint(c.TempoBPM), m.Bars)
			chords := make([]string, 0, len(c.Timeline.Harmony))
			for _, h := range c.Timeline.Harmony {
				chords = append(chords, fmt.Sprintf("%s（%s）", h.RomanNumeral, h.Symbol))
			}
			fmt.Fprintf(&b, "  和音: %s\n", strings.Join(chords, " → "))
			var kinds []string
			for _, kv := range [][2]string{{"スタイル", c.Style}, {"伴奏", c.Accompaniment}, {"ドラム", c.Drums}, {"ノリ", c.Feel}, {"テンション", c.Tensions}} {
				if kv[1] != "" {
					kinds = append(kinds, kv[0]+" "+kv[1])
				}
			}
			if len(kinds) > 0 {
				fmt.Fprintf(&b, "  型: %s\n", strings.Join(kinds, "、"))
			}
			fmt.Fprintf(&b, "  理由: %s\n", c.Reason)
		}
		b.WriteString("\n")
	}
	return b.String()
}

// judge asks the reviewer about one conversation.
func judge(ctx context.Context, client llm.Client, run string, turns []Turn) Verdict {
	resp, err := client.Complete(ctx, llm.Request{
		System:       judgeSystem,
		Messages:     []llm.Message{{Role: llm.RoleUser, Content: judgeInput(turns)}},
		OutputSchema: judgeSchema,
		MaxTokens:    4096,
	})
	if err != nil {
		return Verdict{Run: run, Err: err.Error()}
	}
	var out struct {
		Findings []JudgeFinding `json:"findings"`
		Within   []Within       `json:"within"`
	}
	if err := json.Unmarshal([]byte(resp.Text), &out); err != nil {
		return Verdict{Run: run, Err: "unreadable review: " + err.Error()}
	}
	kept := out.Findings[:0]
	for _, f := range out.Findings {
		if f.Category >= 1 && f.Category < len(judgeCategories) {
			kept = append(kept, f)
		}
	}
	within := out.Within[:0]
	for _, w := range out.Within {
		if w.Kind >= 1 && w.Kind < len(withinKinds) {
			within = append(within, w)
		}
	}
	return Verdict{Run: run, Findings: kept, Within: within}
}

// judgeAll reviews every run, a few at once, in the order given.
func judgeAll(ctx context.Context, client llm.Client, names []string, runs [][]Turn) []Verdict {
	out := make([]Verdict, len(runs))
	sem := make(chan struct{}, 4)
	var wg sync.WaitGroup
	for i := range runs {
		sem <- struct{}{}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()
			out[i] = judge(ctx, client, names[i], runs[i])
		}(i)
	}
	wg.Wait()
	return out
}

// judgeReport counts the findings by category and lists them, for the
// owner to judge whether the reviewer is right.
func judgeReport(verdicts []Verdict) string {
	reviewed, notRun := 0, 0
	counts := make([]int, len(judgeCategories))
	var lines []string
	for _, v := range verdicts {
		if v.Err != "" {
			notRun++
			lines = append(lines, fmt.Sprintf("| %s | | | 査読できず | %s |", v.Run, oneLine(v.Err)))
			continue
		}
		reviewed++
		for _, f := range v.Findings {
			counts[f.Category]++
			where := "返事"
			if f.Candidate > 0 {
				where = fmt.Sprintf("候補 %d", f.Candidate)
			}
			lines = append(lines, fmt.Sprintf("| %s | %d | %s | %s | %s：%s |", v.Run, f.Turn, where, judgeCategories[f.Category], oneLine(f.Quote), oneLine(f.Why)))
		}
	}
	var b strings.Builder
	b.WriteString("\n### 会話の査読（ADR-0064。報告だけで、合否には使わない）\n\n")
	var sums []string
	for c := 1; c < len(judgeCategories); c++ {
		sums = append(sums, fmt.Sprintf("%s %d", judgeCategories[c], counts[c]))
	}
	fmt.Fprintf(&b, "- 査読した会話: %d（査読できなかったもの: %d）\n- 区分ごとの件数: %s\n", reviewed, notRun, strings.Join(sums, "、"))
	if len(lines) > 0 {
		b.WriteString("\n| run | turn | どこ | 区分 | 引用：理由 |\n| --- | --- | --- | --- | --- |\n")
		b.WriteString(strings.Join(lines, "\n"))
		b.WriteString("\n")
	}
	b.WriteString(withinReport(verdicts))
	return b.String()
}

// withinReport lists where the answers used the tolerance (ADR-0065):
// shown so the owner can read what it let through, never counted above.
func withinReport(verdicts []Verdict) string {
	counts := make([]int, len(withinKinds))
	var lines []string
	for _, v := range verdicts {
		for _, w := range v.Within {
			counts[w.Kind]++
			lines = append(lines, fmt.Sprintf("| %s | %d | %s | %s：%s |", v.Run, w.Turn, withinKinds[w.Kind], oneLine(w.Quote), oneLine(w.Why)))
		}
	}
	var b strings.Builder
	b.WriteString("\n#### 幅の内（ADR-0065。数えない）\n\n")
	var sums []string
	for k := 1; k < len(withinKinds); k++ {
		sums = append(sums, fmt.Sprintf("%s %d", withinKinds[k], counts[k]))
	}
	fmt.Fprintf(&b, "- 種類ごとの件数: %s\n", strings.Join(sums, "、"))
	if len(lines) > 0 {
		b.WriteString("\n| run | turn | 種類 | 引用：理由 |\n| --- | --- | --- | --- |\n")
		b.WriteString(strings.Join(lines, "\n"))
		b.WriteString("\n")
	}
	return b.String()
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "|", "／")), " ")
}

// judgeCmd reads collected runs, reviews each conversation, and writes the
// findings and the API usage. It reads the same settings as the server
// (LLM_PROVIDER, ANTHROPIC_API_KEY, ANTHROPIC_MODEL); with the fake
// provider every review finds nothing, which is enough to try the path
// without a key.
func judgeCmd(w io.Writer, paths []string) error {
	if len(paths) == 0 {
		return errors.New("judge: give the collected run files")
	}
	runs, _, err := readRuns(paths)
	if err != nil {
		return err
	}
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}
	var client llm.Client
	switch cfg.LLMProvider {
	case config.LLMProviderAnthropic:
		client, err = anthropic.New(anthropic.Config{APIKey: cfg.AnthropicAPIKey, Model: cfg.AnthropicModel, Logger: slog.New(slog.DiscardHandler)})
		if err != nil {
			return err
		}
	default:
		client = fake.New()
	}
	meter := evals.NewMeter(client)
	names := make([]string, len(paths))
	for i := range paths {
		names[i] = runName(paths, i)
	}
	verdicts := judgeAll(context.Background(), meter, names, runs)
	_, err = io.WriteString(w, judgeReport(verdicts)+evals.UsageReport(evals.UsagePart{Name: "会話の査読", Meter: meter}))
	return err
}
