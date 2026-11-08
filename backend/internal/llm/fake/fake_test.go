package fake

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
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

func TestUnscriptedStructuredOutputIsDeterministicAndMoodAware(t *testing.T) {
	c := New()
	ask := func(text string) string {
		resp, err := c.Complete(context.Background(), llm.Request{
			Messages:     []llm.Message{{Role: llm.RoleUser, Content: text}},
			OutputSchema: []byte(`{}`),
		})
		if err != nil {
			t.Fatal(err)
		}
		return resp.Text
	}
	rainy, rainyAgain, sunny := ask("雨の午後"), ask("雨の午後"), ask("晴れた朝")
	if rainy != rainyAgain {
		t.Fatal("samples must be deterministic")
	}
	if !strings.Contains(rainy, `"minor"`) || !strings.Contains(sunny, `"major"`) {
		t.Fatalf("mood keywords should pick the mode: %s / %s", rainy, sunny)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(rainy), &parsed); err != nil {
		t.Fatalf("sample is not JSON: %v", err)
	}
}

func TestUnscriptedFakeCanGoLong(t *testing.T) {
	c := New()
	resp, err := c.Complete(context.Background(), llm.Request{
		OutputSchema: json.RawMessage(`{}`),
		Messages:     []llm.Message{{Role: llm.RoleUser, Content: "長い曲を雨の日に"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var in struct {
		Candidates []struct {
			Title       string   `json:"title"`
			Progression []string `json:"progression"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal([]byte(resp.Text), &in); err != nil {
		t.Fatal(err)
	}
	if len(in.Candidates[0].Progression) != 16 || in.Candidates[0].Title != "窓辺の雨（長め）" {
		t.Fatalf("long sample = %+v", in.Candidates[0])
	}
}

// TestUnscriptedKeywordsCombine (review of BL-085/077): keywords that
// change the first candidate stay valid together: pushes match a long
// progression, and a half-time grid is written only where it fits.
func TestUnscriptedKeywordsCombine(t *testing.T) {
	for _, text := range []string{"ロックで長い曲", "ハーフタイムのワルツ", "ハーフタイムのケルト"} {
		resp, err := New().Complete(context.Background(), llm.Request{
			OutputSchema: json.RawMessage(`{}`),
			Messages:     []llm.Message{{Role: llm.RoleUser, Content: text}},
		})
		if err != nil {
			t.Fatal(err)
		}
		var in struct {
			Candidates []struct {
				Meter       string          `json:"meter"`
				Style       string          `json:"style"`
				Progression []string        `json:"progression"`
				Pushes      []string        `json:"pushes"`
				DrumGrid    json.RawMessage `json:"drumGrid"`
			} `json:"candidates"`
		}
		if err := json.Unmarshal([]byte(resp.Text), &in); err != nil {
			t.Fatal(err)
		}
		c := in.Candidates[0]
		if c.Pushes != nil && len(c.Pushes) != len(c.Progression) {
			t.Errorf("%s: %d pushes for %d chords", text, len(c.Pushes), len(c.Progression))
		}
		if c.DrumGrid != nil && (c.Meter != "" || (c.Style != "pop" && c.Style != "rock")) {
			t.Errorf("%s: a 4/4 eight-beat grid on %q %q", text, c.Meter, c.Style)
		}
	}
}
