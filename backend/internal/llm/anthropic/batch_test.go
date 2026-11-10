package anthropic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hassaku63/middy/backend/internal/llm"
)

// batchServer is a fake Message Batches API: a batch is in progress on the
// first poll and ended on the next, and each request is answered with its
// own words, or errored when it says "fail".
type batchServer struct {
	mu      sync.Mutex
	batches [][]string // the user words of each batch's requests, by custom_id
	polls   map[string]int
}

func (s *batchServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/v1/messages/batches":
		var body struct {
			Requests []struct {
				CustomID string `json:"custom_id"`
				Params   struct {
					Messages []struct {
						Content []struct {
							Text string `json:"text"`
						} `json:"content"`
					} `json:"messages"`
				} `json:"params"`
			} `json:"requests"`
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		words := make([]string, len(body.Requests))
		for _, req := range body.Requests {
			var i int
			_, _ = fmt.Sscan(req.CustomID, &i)
			words[i] = req.Params.Messages[0].Content[0].Text
		}
		s.batches = append(s.batches, words)
		writeBatch(w, len(s.batches), "in_progress")
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/results"):
		var n int
		_, _ = fmt.Sscanf(r.URL.Path, "/v1/messages/batches/b%d/results", &n)
		w.Header().Set("Content-Type", "application/x-jsonl")
		for i, word := range s.batches[n-1] {
			var line map[string]any
			if word == "fail" {
				line = map[string]any{"custom_id": strconv.Itoa(i), "result": map[string]any{
					"type": "errored", "error": map[string]any{"type": "error", "error": map[string]any{"type": "overloaded_error", "message": "busy"}},
				}}
			} else {
				var msg map[string]any
				_ = json.Unmarshal([]byte(messagesResponse("end_turn", "re: "+word)), &msg)
				line = map[string]any{"custom_id": strconv.Itoa(i), "result": map[string]any{"type": "succeeded", "message": msg}}
			}
			b, _ := json.Marshal(line)
			_, _ = w.Write(append(b, '\n'))
		}
	case r.Method == http.MethodGet:
		var n int
		_, _ = fmt.Sscanf(r.URL.Path, "/v1/messages/batches/b%d", &n)
		s.polls[r.URL.Path]++
		writeBatch(w, n, "ended")
	default:
		http.NotFound(w, r)
	}
}

func writeBatch(w http.ResponseWriter, n int, status string) {
	b, _ := json.Marshal(map[string]any{
		"id": fmt.Sprintf("b%d", n), "type": "message_batch", "processing_status": status,
		"request_counts": map[string]int{}, "created_at": "2026-11-10T00:00:00Z", "expires_at": "2026-11-11T00:00:00Z",
	})
	_, _ = w.Write(b)
}

// TestBatcherSendsConcurrentRequestsAsOneBatch (#334): requests made
// together go as one batch, each caller gets its own answer or error, and
// a request made after the batch went out goes in the next one.
func TestBatcherSendsConcurrentRequestsAsOneBatch(t *testing.T) {
	fake := &batchServer{polls: map[string]int{}}
	srv := httptest.NewServer(fake)
	t.Cleanup(srv.Close)
	b := NewBatcher(context.Background(), newClient(t, srv.URL), 50*time.Millisecond, 10*time.Millisecond)

	words := []string{"雨", "fail", "夜"}
	got := make([]string, len(words))
	errs := make([]error, len(words))
	var wg sync.WaitGroup
	for i, word := range words {
		wg.Go(func() {
			resp, err := b.Complete(context.Background(), llm.Request{Messages: []llm.Message{{Role: llm.RoleUser, Content: word}}})
			got[i], errs[i] = resp.Text, err
		})
	}
	wg.Wait()

	if len(fake.batches) != 1 || len(fake.batches[0]) != 3 {
		t.Fatalf("want one batch of 3, got %v", fake.batches)
	}
	if got[0] != "re: 雨" || got[2] != "re: 夜" || errs[0] != nil || errs[2] != nil {
		t.Errorf("answers went to the wrong callers: %q %v", got, errs)
	}
	if !errors.Is(errs[1], ErrBatch) || !strings.Contains(errs[1].Error(), "overloaded_error") {
		t.Errorf("an errored request must be an ErrBatch with its reason, got %v", errs[1])
	}
	if fake.polls["/v1/messages/batches/b1"] == 0 {
		t.Error("the batch was not polled until it ended")
	}

	resp, err := b.Complete(context.Background(), llm.Request{Messages: []llm.Message{{Role: llm.RoleUser, Content: "もう一度"}}})
	if err != nil || resp.Text != "re: もう一度" || len(fake.batches) != 2 {
		t.Fatalf("a later request goes in a second batch: %q %v, %d batches", resp.Text, err, len(fake.batches))
	}
	if resp.Usage.InputTokens != 42 {
		t.Errorf("usage must be read from the batched message: %+v", resp.Usage)
	}
}

// TestBatcherGivesUpWithItsContext: a batch that does not end before the
// batcher's context does fails its requests rather than hanging.
func TestBatcherGivesUpWithItsContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		writeBatch(w, 1, "in_progress") // never ends
	}))
	t.Cleanup(srv.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	b := NewBatcher(ctx, newClient(t, srv.URL), 10*time.Millisecond, 20*time.Millisecond)
	_, err := b.Complete(context.Background(), llm.Request{Messages: []llm.Message{{Role: llm.RoleUser, Content: "雨"}}})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("want the batcher's deadline, got %v", err)
	}
}
