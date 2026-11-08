//go:build eval

package evals

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hassaku63/middy/backend/internal/ideas"
	"github.com/hassaku63/middy/backend/internal/llm/anthropic"
)

// TestLiveEval runs the evaluation set against the real API. It is behind
// the "eval" build tag and needs ANTHROPIC_API_KEY. Locally run `make eval`;
// in CI it runs when a pull request gets the `eval` label (eval.yml,
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
	cases, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	verdicts := Run(ctx, ideas.NewDiagnostic(client), cases)
	fmt.Println(Report(verdicts))
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
