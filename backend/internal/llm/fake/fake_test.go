package fake

import (
	"context"
	"errors"
	"testing"

	"github.com/hassaku63/middy/backend/internal/llm"
)

func TestScriptedResponses(t *testing.T) {
	c := New(llm.Response{Text: "one"}, llm.Response{Text: "two"})
	ctx := context.Background()

	for _, want := range []string{"one", "two"} {
		got, err := c.Complete(ctx, llm.Request{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Text != want {
			t.Fatalf("got %q, want %q", got.Text, want)
		}
	}

	if _, err := c.Complete(ctx, llm.Request{}); !errors.Is(err, ErrNoScriptedResponse) {
		t.Fatalf("want ErrNoScriptedResponse, got %v", err)
	}
	if n := len(c.Requests()); n != 3 {
		t.Fatalf("recorded %d requests, want 3", n)
	}
}

func TestPlaceholderEchoesLastUserMessage(t *testing.T) {
	c := New()
	got, err := c.Complete(context.Background(), llm.Request{Messages: []llm.Message{
		{Role: llm.RoleUser, Content: "first"},
		{Role: llm.RoleAssistant, Content: "reply"},
		{Role: llm.RoleUser, Content: "rainy sunday"},
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Text != "(fake llm) rainy sunday" {
		t.Fatalf("got %q", got.Text)
	}
}
