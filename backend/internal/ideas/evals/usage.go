package evals

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/hassaku63/middy/backend/internal/llm"
)

// Meter is an llm.Client that adds up the calls and tokens that go through
// it (#334): the report states what a run cost, so ways to spend less can
// be compared. Safe for the concurrent cases of Run and ReviewAll.
type Meter struct {
	next llm.Client

	mu     sync.Mutex
	calls  int
	failed int
	usage  llm.Usage
}

// NewMeter wraps the client.
func NewMeter(next llm.Client) *Meter { return &Meter{next: next} }

// Complete forwards the request and counts it; a call that fails is
// counted with whatever usage it reported.
func (m *Meter) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	resp, err := m.next.Complete(ctx, req)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	if err != nil {
		m.failed++
	}
	u := resp.Usage
	m.usage.InputTokens += u.InputTokens
	m.usage.OutputTokens += u.OutputTokens
	m.usage.CacheReadInputTokens += u.CacheReadInputTokens
	m.usage.CacheCreationInputTokens += u.CacheCreationInputTokens
	return resp, err
}

// Total is the calls, failed calls and summed usage so far.
func (m *Meter) Total() (calls, failed int, usage llm.Usage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls, m.failed, m.usage
}

// UsageReport tables the calls and tokens of each part of a run: the
// answers (with their retries) and the independent reviewer.
func UsageReport(parts ...UsagePart) string {
	var b strings.Builder
	b.WriteString("\n### API の使用量\n\n| 部分 | 呼び出し | 失敗 | 入力 | キャッシュ読み | キャッシュ書き | 出力 |\n| --- | --- | --- | --- | --- | --- | --- |\n")
	var all llm.Usage
	calls, failed := 0, 0
	for _, p := range parts {
		c, f, u := p.Meter.Total()
		fmt.Fprintf(&b, "| %s | %d | %d | %d | %d | %d | %d |\n", p.Name, c, f, u.InputTokens, u.CacheReadInputTokens, u.CacheCreationInputTokens, u.OutputTokens)
		calls, failed = calls+c, failed+f
		all.InputTokens += u.InputTokens
		all.CacheReadInputTokens += u.CacheReadInputTokens
		all.CacheCreationInputTokens += u.CacheCreationInputTokens
		all.OutputTokens += u.OutputTokens
	}
	fmt.Fprintf(&b, "| 計 | %d | %d | %d | %d | %d | %d |\n", calls, failed, all.InputTokens, all.CacheReadInputTokens, all.CacheCreationInputTokens, all.OutputTokens)
	return b.String()
}

// UsagePart names a meter in the report.
type UsagePart struct {
	Name  string
	Meter *Meter
}

// sendNote says how the run's requests went, so the usage above is read
// at the right price: batched requests cost half (ADR-0039).
func sendNote(batch bool) string {
	if batch {
		return "送り方: Message Batches（料金は 1 件ずつの半分。キャッシュの当たりはベストエフォート）\n"
	}
	return "送り方: 1 件ずつ\n"
}
