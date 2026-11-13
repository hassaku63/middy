package anthropic

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hassaku63/middy/backend/internal/llm"
)

// messagesResponse is a recorded Messages API answer, trimmed to the
// fields the adapter reads.
func messagesResponse(stopReason, text string) string {
	body := map[string]any{
		"id": "msg_test", "type": "message", "role": "assistant", "model": "claude-opus-5-5",
		"content":     []map[string]any{{"type": "text", "text": text}},
		"stop_reason": stopReason, "stop_sequence": nil,
		"usage": map[string]any{"input_tokens": 42, "output_tokens": 7, "cache_read_input_tokens": 900, "cache_creation_input_tokens": 30},
	}
	if stopReason == "refusal" {
		body["stop_details"] = map[string]any{"type": "refusal", "category": "general_harms", "explanation": "no"}
	}
	b, _ := json.Marshal(body)
	return string(b)
}

type recorded struct {
	path    string
	headers http.Header
	body    map[string]any
}

// newServer returns a fake API that answers with the given status/body and
// records the last request.
func newServer(t *testing.T, status int, body string) (*httptest.Server, *recorded) {
	t.Helper()
	rec := &recorded{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		rec.path = r.URL.Path
		rec.headers = r.Header.Clone()
		_ = json.Unmarshal(raw, &rec.body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, rec
}

func newClient(t *testing.T, url string) *Client {
	t.Helper()
	zero := 0
	c, err := New(Config{APIKey: "test-key", BaseURL: url, MaxRetries: &zero})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCompleteBuildsTheRequestAndReadsTheAnswer(t *testing.T) {
	srv, rec := newServer(t, 200, messagesResponse("end_turn", `{"summary":"ok"}`))
	c := newClient(t, srv.URL)

	resp, err := c.Complete(context.Background(), llm.Request{
		System: "you are a tutor",
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "雨"},
			{Role: llm.RoleAssistant, Content: "静か"},
			{Role: llm.RoleUser, Content: "もっと明るく"},
		},
		OutputSchema: []byte(`{"type":"object","properties":{"summary":{"type":"string"}}}`),
		MaxTokens:    512,
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if resp.Text != `{"summary":"ok"}` || resp.Usage.InputTokens != 42 || resp.Usage.OutputTokens != 7 ||
		resp.Usage.CacheReadInputTokens != 900 || resp.Usage.CacheCreationInputTokens != 30 {
		t.Fatalf("unexpected response %+v", resp)
	}

	if rec.path != "/v1/messages" {
		t.Errorf("path = %q", rec.path)
	}
	// Current SDKs send the key as a bearer token; older ones used x-api-key.
	if rec.headers.Get("Authorization") != "Bearer test-key" && rec.headers.Get("x-api-key") != "test-key" {
		t.Errorf("api key header missing: %v", rec.headers)
	}
	if rec.body["model"] != DefaultModel || rec.body["max_tokens"] != float64(512) {
		t.Errorf("model/max_tokens = %v / %v", rec.body["model"], rec.body["max_tokens"])
	}
	system := as[map[string]any](t, as[[]any](t, rec.body["system"])[0])
	if system["text"] != "you are a tutor" || system["cache_control"] == nil {
		t.Errorf("system block should carry the prompt and a cache breakpoint: %v", system)
	}
	msgs := as[[]any](t, rec.body["messages"])
	if len(msgs) != 3 || as[map[string]any](t, msgs[1])["role"] != "assistant" {
		t.Errorf("messages not mapped: %v", msgs)
	}
	oc := as[map[string]any](t, rec.body["output_config"])
	if oc["effort"] != "low" {
		t.Errorf("effort = %v, want low", oc["effort"])
	}
	format := as[map[string]any](t, oc["format"])
	if format["type"] != "json_schema" || as[map[string]any](t, format["schema"])["type"] != "object" {
		t.Errorf("structured output format not sent: %v", format)
	}
	if _, has := rec.body["thinking"]; has {
		t.Errorf("thinking must be left to the model's default (adaptive): %v", rec.body["thinking"])
	}
}

func TestCompleteWithoutSchemaOrSystem(t *testing.T) {
	srv, rec := newServer(t, 200, messagesResponse("end_turn", "hello"))
	c := newClient(t, srv.URL)
	resp, err := c.Complete(context.Background(), llm.Request{Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}}})
	if err != nil || resp.Text != "hello" {
		t.Fatalf("got %+v, %v", resp, err)
	}
	if _, has := rec.body["system"]; has {
		t.Errorf("no system prompt should be sent when empty")
	}
	if _, has := rec.body["output_config"].(map[string]any)["format"]; has {
		t.Errorf("no format should be sent without a schema")
	}
	if rec.body["max_tokens"] != float64(DefaultMaxTokens) {
		t.Errorf("default max_tokens not applied: %v", rec.body["max_tokens"])
	}
}

func TestCompleteStopReasons(t *testing.T) {
	tests := []struct {
		stop string
		want error
	}{
		{"refusal", ErrRefused},
		{"max_tokens", ErrTruncated},
	}
	for _, tt := range tests {
		t.Run(tt.stop, func(t *testing.T) {
			srv, _ := newServer(t, 200, messagesResponse(tt.stop, "partial"))
			_, err := newClient(t, srv.URL).Complete(context.Background(), llm.Request{Messages: []llm.Message{{Role: llm.RoleUser, Content: "x"}}})
			if !errors.Is(err, tt.want) {
				t.Fatalf("want %v, got %v", tt.want, err)
			}
		})
	}
}

func TestCompleteAPIErrorsKeepTheStatus(t *testing.T) {
	srv, _ := newServer(t, 429, `{"type":"error","error":{"type":"rate_limit_error","message":"slow down"}}`)
	_, err := newClient(t, srv.URL).Complete(context.Background(), llm.Request{Messages: []llm.Message{{Role: llm.RoleUser, Content: "x"}}})
	if err == nil || !strings.Contains(err.Error(), "status 429") {
		t.Fatalf("want a 429 in the error, got %v", err)
	}
}

func TestCompleteTimeoutCoversTheRetries(t *testing.T) {
	// The API keeps failing with a retryable status: the SDK backs off and
	// tries again (about half a second and more between attempts), so two
	// retries alone take over a second. The client's timeout must end the
	// whole call, not each attempt (#499).
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"type":"error","error":{"type":"api_error","message":"down"}}`))
	}))
	t.Cleanup(srv.Close)
	two := 2
	c, err := New(Config{APIKey: "test-key", BaseURL: srv.URL, MaxRetries: &two, Timeout: 300 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	_, err = c.Complete(context.Background(), llm.Request{Messages: []llm.Message{{Role: llm.RoleUser, Content: "x"}}})
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("a failing API must fail the call")
	}
	if elapsed > 900*time.Millisecond {
		t.Fatalf("the call took %v (%d attempts); the 300ms timeout must cover the retries", elapsed, calls.Load())
	}
}

func TestCompleteTransportErrors(t *testing.T) {
	c := newClient(t, "http://127.0.0.1:1") // nothing listens here
	_, err := c.Complete(context.Background(), llm.Request{Messages: []llm.Message{{Role: llm.RoleUser, Content: "x"}}})
	if err == nil || !strings.Contains(err.Error(), "transport") {
		t.Fatalf("want a transport error, got %v", err)
	}
}

func TestNewValidatesConfig(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("missing API key must be rejected")
	}
	c, err := New(Config{APIKey: "k", Model: "claude-sonnet-5-5", Effort: "medium"})
	if err != nil || c.model != "claude-sonnet-5-5" || string(c.effort) != "medium" {
		t.Fatalf("config not applied: %+v %v", c, err)
	}
}

func TestParamsRejectsBadInput(t *testing.T) {
	c := newClient(t, "http://127.0.0.1:1")
	if _, err := c.params(llm.Request{Messages: []llm.Message{{Role: "system", Content: "x"}}}); err == nil {
		t.Error("unsupported role must be rejected before the call")
	}
	if _, err := c.params(llm.Request{OutputSchema: []byte(`[1]`)}); err == nil {
		t.Error("non-object schema must be rejected before the call")
	}
}

// TestErrorsDoNotCarryTheKey: the server logs upstream errors (BL-144), so
// an API or transport error must never include the API key.
func TestErrorsDoNotCarryTheKey(t *testing.T) {
	const key = "sk-ant-test-SENTINEL-do-not-log"
	zero := 0
	req := llm.Request{Messages: []llm.Message{{Role: llm.RoleUser, Content: "x"}}}
	srv, _ := newServer(t, 401, `{"type":"error","error":{"type":"authentication_error","message":"invalid x-api-key"}}`)
	for name, url := range map[string]string{"api error": srv.URL, "transport error": "http://127.0.0.1:1"} {
		c, err := New(Config{APIKey: key, BaseURL: url, MaxRetries: &zero})
		if err != nil {
			t.Fatal(err)
		}
		_, err = c.Complete(context.Background(), req)
		if err == nil {
			t.Fatalf("%s: want an error", name)
		}
		if strings.Contains(err.Error(), key) || strings.Contains(err.Error(), "SENTINEL") {
			t.Errorf("%s: the error carries the key: %v", name, err)
		}
	}
}

// as asserts v's dynamic type and fails the test with what it found.
func as[T any](t *testing.T, v any) T {
	t.Helper()
	got, ok := v.(T)
	if !ok {
		t.Fatalf("got %T, want %T", v, got)
	}
	return got
}
