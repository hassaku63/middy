//go:build eval

package evals

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hassaku63/middy/backend/internal/ideas"
	"github.com/hassaku63/middy/backend/internal/llm/anthropic"
)

// TestLiveEval runs the evaluation set against the real API. It is behind
// the "eval" build tag and needs ANTHROPIC_API_KEY. Locally run `make eval`;
// in CI it runs when a pull request gets the `needs-eval` label (eval.yml,
// ADR-0018), which posts the printed table as a comment.
func TestLiveEval(t *testing.T) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}
	client, err := anthropic.New(anthropic.Config{APIKey: key, Model: os.Getenv("ANTHROPIC_MODEL"), Timeout: 60 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	all, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	// EVAL_TAGS="meter,smoke" runs only the cases with those tags (#334).
	var tags []string
	for _, tag := range strings.Split(os.Getenv("EVAL_TAGS"), ",") {
		if tag = strings.TrimSpace(tag); tag != "" {
			tags = append(tags, tag)
		}
	}
	cases, err := Select(all, tags)
	if err != nil {
		t.Fatal(err)
	}
	// The reviewer runs on full runs; a narrowed run skips it (#334).
	// Checked before anything is sent, like the tags.
	review, err := ReviewWanted(tags, os.Getenv("EVAL_REVIEW"))
	if err != nil {
		t.Fatal(err)
	}
	// End a minute before `go test` would kill the run, so the report is
	// printed with whatever finished (the default -timeout is 10 minutes).
	budget := 9 * time.Minute
	if dl, ok := t.Deadline(); ok {
		budget = time.Until(dl) - time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), budget)
	defer cancel()
	// The answers get two thirds of the budget; the independent reviewer
	// (BL-172) the rest. A review cut short is reported as not run.
	runCtx, cancelRun := context.WithTimeout(ctx, budget*2/3)
	defer cancelRun()
	answers, reviewer := NewMeter(client), NewMeter(client)
	verdicts := Run(runCtx, ideas.NewDiagnostic(answers), cases)
	reviewNote := ""
	if review {
		ReviewAll(ctx, reviewer, verdicts)
	} else {
		reviewNote = SkipReview(verdicts)
	}
	fmt.Println(Report(verdicts) + reviewNote + UsageReport(UsagePart{"答え（出し直しを含む）", answers}, UsagePart{"レビュアー", reviewer}) + scopeNote(tags, len(cases), len(all)))
	failed := 0
	for _, v := range verdicts {
		if !v.Pass {
			failed++
		}
	}
	// Taste varies; a regression is when many cases fail, not one.
	if failed > len(verdicts)/4 {
		t.Fatalf("%d of %d cases failed (more than a quarter)", failed, len(verdicts))
	}
}
