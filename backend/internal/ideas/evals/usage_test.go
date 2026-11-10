package evals

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/hassaku63/middy/backend/internal/llm"
)

type usageClient struct {
	usage llm.Usage
	err   error
}

func (c usageClient) Complete(context.Context, llm.Request) (llm.Response, error) {
	return llm.Response{Text: "{}", Usage: c.usage}, c.err
}

// TestMeterAddsUpEveryCall (#334): concurrent calls are all counted, with
// failed ones, and the report tables each part and the total.
func TestMeterAddsUpEveryCall(t *testing.T) {
	answers := NewMeter(usageClient{usage: llm.Usage{InputTokens: 10, OutputTokens: 100, CacheReadInputTokens: 2000, CacheCreationInputTokens: 5}})
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = answers.Complete(context.Background(), llm.Request{})
		}()
	}
	wg.Wait()
	review := NewMeter(usageClient{usage: llm.Usage{InputTokens: 300, OutputTokens: 50}, err: errors.New("boom")})
	_, _ = review.Complete(context.Background(), llm.Request{})

	if c, f, u := answers.Total(); c != 8 || f != 0 || u.OutputTokens != 800 || u.CacheReadInputTokens != 16000 {
		t.Fatalf("answers: %d calls, %d failed, %+v", c, f, u)
	}
	got := UsageReport(UsagePart{"答え（出し直しを含む）", answers}, UsagePart{"レビュアー", review})
	for _, want := range []string{
		"### API の使用量",
		"| 答え（出し直しを含む） | 8 | 0 | 80 | 16000 | 40 | 800 |",
		"| レビュアー | 1 | 1 | 300 | 0 | 0 | 50 |",
		"| 計 | 9 | 1 | 380 | 16000 | 40 | 850 |",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("report lacks %q:\n%s", want, got)
		}
	}
}

// TestSendNoteSaysHowTheRunWent (ADR-0039): the price of the usage table
// depends on whether the run was batched, so the report says which.
func TestSendNoteSaysHowTheRunWent(t *testing.T) {
	if !strings.Contains(sendNote(true), "Message Batches") || !strings.Contains(sendNote(false), "1 件ずつ") {
		t.Errorf("send notes: %q %q", sendNote(true), sendNote(false))
	}
}
