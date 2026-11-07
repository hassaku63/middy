package ideas

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hassaku63/middy/backend/internal/llm"
)

// step is one scripted answer: a response or an error.
type step struct {
	text string
	err  error
}

// scripted is an llm.Client that can also fail, which the fake cannot.
type scripted struct {
	steps []step
	reqs  []llm.Request
}

func (s *scripted) Complete(_ context.Context, req llm.Request) (llm.Response, error) {
	s.reqs = append(s.reqs, req)
	if len(s.steps) == 0 {
		return llm.Response{}, errors.New("scripted: no step left")
	}
	next := s.steps[0]
	s.steps = s.steps[1:]
	return llm.Response{Text: next.text}, next.err
}

// errRefusedByAdapter looks like what an adapter returns: llm.ErrRefused wrapped.
var errRefusedByAdapter = fmt.Errorf("anthropic: %w (reasoning_extraction)", llm.ErrRefused)

func TestImagineRetriesOnceAfterARefusal(t *testing.T) {
	tests := []struct {
		name    string
		steps   []step
		wantErr error
	}{
		{"second try answers", []step{{err: errRefusedByAdapter}, {text: goodIntent}}, nil},
		{"refused twice", []step{{err: errRefusedByAdapter}, {err: errRefusedByAdapter}}, ErrDeclined},
		{"refused, then unusable", []step{{err: errRefusedByAdapter}, {text: "not json"}}, ErrInvalidOutput},
		{"unusable, then refused", []step{{text: badIntent}, {err: errRefusedByAdapter}}, ErrDeclined},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &scripted{steps: tt.steps}
			res, err := New(client).Imagine(context.Background(), []Message{user("深夜に一人で勉強している集中した時間")})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("want %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && len(res.Candidates) != 2 {
				t.Fatalf("second answer not used: %+v", res)
			}
			if len(client.reqs) != 2 {
				t.Fatalf("exactly one retry expected, got %d requests", len(client.reqs))
			}
		})
	}
}

func TestImagineResendsARefusedRequestUnchanged(t *testing.T) {
	client := &scripted{steps: []step{{err: errRefusedByAdapter}, {text: goodIntent}}}
	if _, err := New(client).Imagine(context.Background(), []Message{user("x")}); err != nil {
		t.Fatal(err)
	}
	if len(client.reqs[1].Messages) != 1 || client.reqs[1].Messages[0].Content != "x" {
		t.Fatalf("a refusal has nothing to correct, so the retry must be the same request: %+v", client.reqs[1].Messages)
	}
}

func TestImagineDoesNotRetryOutages(t *testing.T) {
	client := &scripted{steps: []step{{err: errors.New("connection reset")}}}
	_, err := New(client).Imagine(context.Background(), []Message{user("x")})
	if !errors.Is(err, ErrLLMUnavailable) || errors.Is(err, ErrDeclined) {
		t.Fatalf("want ErrLLMUnavailable only, got %v", err)
	}
	if len(client.reqs) != 1 {
		t.Fatalf("outages are not retried here (the SDK already retries), got %d requests", len(client.reqs))
	}
}
