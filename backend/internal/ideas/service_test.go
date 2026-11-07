package ideas

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/llm"
	"github.com/hassaku63/middy/backend/internal/llm/fake"
)

const goodIntent = `{
  "summary": "雨の日曜の午後。静かで少し切ない。",
  "mood": ["静か", "切ない"],
  "candidates": [
    {"title": "窓辺の雨", "key": {"tonic": 9, "mode": "minor"}, "tempoBpm": 72,
     "progression": ["i", "VI", "III", "VII"], "beatsPerChord": 4, "reason": "短調で落ち着いた切なさ。"},
    {"title": "午後の光", "key": {"tonic": 5, "mode": "major"}, "tempoBpm": 80,
     "progression": ["Imaj7", "vi7", "ii7", "V7"], "beatsPerChord": 2, "reason": "七の和音で柔らかく。"}
  ]
}`

const badIntent = `{"summary": "x", "mood": [], "candidates": [
  {"title": "t", "key": {"tonic": 0, "mode": "major"}, "tempoBpm": 500, "progression": ["bVII"], "beatsPerChord": 4, "reason": "r"}
]}`

func user(text string) Message { return Message{Role: llm.RoleUser, Content: text} }

func TestImagineRendersEveryCandidate(t *testing.T) {
	client := fake.New(llm.Response{Text: goodIntent})
	res, err := New(client).Imagine(context.Background(), []Message{user("雨の日曜の午後")})
	if err != nil {
		t.Fatalf("Imagine: %v", err)
	}
	if res.Reply != "雨の日曜の午後。静かで少し切ない。" || len(res.Mood) != 2 {
		t.Fatalf("unexpected reply/mood: %+v", res)
	}
	if len(res.Candidates) != 2 {
		t.Fatalf("got %d candidates", len(res.Candidates))
	}
	first := res.Candidates[0]
	if first.Title != "窓辺の雨" || first.Timeline.TempoBPM != 72 || len(first.Timeline.Harmony) != 4 {
		t.Fatalf("first candidate not rendered: %+v", first)
	}
	if first.Timeline.Harmony[0].RomanNumeral != "i" || first.Timeline.Harmony[0].Analysis == nil {
		t.Fatalf("timeline not analysed: %+v", first.Timeline.Harmony[0])
	}
	// 8 chords of 2 beats = 16 beats.
	if got := res.Candidates[1].Timeline.LengthTicks(); got != 8*480 {
		t.Fatalf("second candidate length = %d", got)
	}

	// The request carried the system prompt, the schema and the user text.
	reqs := client.Requests()
	if len(reqs) != 1 || reqs[0].System == "" || reqs[0].OutputSchema == nil || reqs[0].Messages[0].Content != "雨の日曜の午後" {
		t.Fatalf("unexpected request: %+v", reqs)
	}
}

func TestImagineRetriesOnceWithTheValidationErrors(t *testing.T) {
	client := fake.New(llm.Response{Text: badIntent}, llm.Response{Text: goodIntent})
	res, err := New(client).Imagine(context.Background(), []Message{user("雨")})
	if err != nil {
		t.Fatalf("Imagine: %v", err)
	}
	if len(res.Candidates) != 2 {
		t.Fatalf("retry result not used: %+v", res)
	}
	reqs := client.Requests()
	if len(reqs) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(reqs))
	}
	retry := reqs[1].Messages
	if len(retry) != 3 || retry[1].Role != llm.RoleAssistant || retry[1].Content != badIntent {
		t.Fatalf("retry must quote the bad answer as the assistant turn: %+v", retry)
	}
	if !strings.Contains(retry[2].Content, "tempoBpm") || !strings.Contains(retry[2].Content, "bVII") {
		t.Fatalf("retry must carry the validation errors: %q", retry[2].Content)
	}
}

func TestImagineGivesUpAfterTheRetry(t *testing.T) {
	client := fake.New(llm.Response{Text: badIntent}, llm.Response{Text: "not json at all"})
	_, err := New(client).Imagine(context.Background(), []Message{user("雨")})
	if !errors.Is(err, ErrInvalidOutput) {
		t.Fatalf("want ErrInvalidOutput, got %v", err)
	}
	if len(client.Requests()) != 2 {
		t.Fatalf("exactly one retry expected, got %d requests", len(client.Requests()))
	}
}

func TestImagineReportsProviderFailures(t *testing.T) {
	// Exhaust the script so the next call fails with ErrNoScriptedResponse.
	client := fake.New(llm.Response{Text: goodIntent})
	_, _ = client.Complete(context.Background(), llm.Request{})
	_, err := New(client).Imagine(context.Background(), []Message{user("x")})
	if !errors.Is(err, ErrLLMUnavailable) {
		t.Fatalf("want ErrLLMUnavailable, got %v", err)
	}
}

func TestImagineValidatesTheConversation(t *testing.T) {
	svc := New(fake.New(llm.Response{Text: goodIntent}))
	tests := []struct {
		name string
		msgs []Message
		want string
	}{
		{"empty", nil, "must not be empty"},
		{"ends with assistant", []Message{user("a"), {Role: llm.RoleAssistant, Content: "b"}}, "last message must be from the user"},
		{"bad role", []Message{{Role: "system", Content: "x"}}, "role must be user or assistant"},
		{"blank", []Message{user("   ")}, "content is empty"},
		{"too long", []Message{user(strings.Repeat("あ", MaxMessageRunes+1))}, "too long"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Imagine(context.Background(), tt.msgs)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("want %q, got %v", tt.want, err)
			}
		})
	}
}

func TestUnscriptedFakeProducesRenderableSamples(t *testing.T) {
	svc := New(fake.New())
	for _, text := range []string{"雨の日曜の午後", "出発の朝、晴れ"} {
		res, err := svc.Imagine(context.Background(), []Message{user(text)})
		if err != nil {
			t.Fatalf("%q: %v", text, err)
		}
		if len(res.Candidates) < 1 {
			t.Fatalf("%q: no candidates", text)
		}
	}
	rainy, _ := svc.Imagine(context.Background(), []Message{user("雨")})
	sunny, _ := svc.Imagine(context.Background(), []Message{user("晴れ")})
	if rainy.Candidates[0].Key.Mode == sunny.Candidates[0].Key.Mode {
		t.Fatal("samples should differ by mood keywords")
	}
}

func TestResultJSONShape(t *testing.T) {
	res, err := New(fake.New(llm.Response{Text: goodIntent})).Imagine(context.Background(), []Message{user("雨")})
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"reply"`, `"mood"`, `"candidates"`, `"title"`, `"timeline"`, `"romanNumeral"`} {
		if !strings.Contains(string(b), key) {
			t.Errorf("JSON should contain %s", key)
		}
	}
}
