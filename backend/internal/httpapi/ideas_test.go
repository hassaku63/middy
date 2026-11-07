package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/ideas"
	"github.com/hassaku63/middy/backend/internal/llm"
	"github.com/hassaku63/middy/backend/internal/llm/fake"
)

const sampleIntent = `{"summary":"静かな雨。","mood":["静か"],"candidates":[
  {"title":"窓辺の雨","key":{"tonic":9,"mode":"minor"},"tempoBpm":72,"progression":["i","VI","III","VII"],"beatsPerChord":4,"reason":"短調で。"}
]}`

func newIdeasServer(t *testing.T, client llm.Client) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(NewHandler(Deps{
		LLM:    client,
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestIdeasHappyPath(t *testing.T) {
	srv := newIdeasServer(t, fake.New(llm.Response{Text: sampleIntent}))
	resp := postJSON(t, srv.URL+"/api/v1/ideas", `{"messages":[{"role":"user","content":"雨の日曜の午後"}]}`)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}
	var res ideas.Result
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if res.Reply != "静かな雨。" || len(res.Candidates) != 1 || res.Candidates[0].Title != "窓辺の雨" {
		t.Fatalf("unexpected result %+v", res)
	}
	if err := res.Candidates[0].Timeline.Validate(); err != nil || res.Candidates[0].Timeline.Harmony[0].Analysis == nil {
		t.Fatalf("candidate timeline not rendered/analysed: %v", err)
	}
}

func TestIdeasWorksWithTheUnscriptedFake(t *testing.T) {
	srv := newTestServer(t) // Deps.LLM = fake.New(): the local-dev configuration
	resp := postJSON(t, srv.URL+"/api/v1/ideas", `{"messages":[{"role":"user","content":"晴れた朝の出発"}]}`)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}
}

func TestIdeasErrors(t *testing.T) {
	tests := []struct {
		name       string
		client     llm.Client
		body       string
		wantStatus int
		wantCode   string
	}{
		{"malformed json", fake.New(), `{"messages":`, 400, "invalid_json"},
		{"unknown field", fake.New(), `{"messages":[{"role":"user","content":"x"}],"model":"gpt"}`, 400, "invalid_json"},
		{"empty conversation", fake.New(), `{"messages":[]}`, 400, "invalid_request"},
		{"ends with assistant", fake.New(), `{"messages":[{"role":"user","content":"a"},{"role":"assistant","content":"b"}]}`, 400, "invalid_request"},
		{"provider down", exhausted(), `{"messages":[{"role":"user","content":"x"}]}`, 502, "llm_unavailable"},
		{"unusable twice", fake.New(llm.Response{Text: "nope"}, llm.Response{Text: "still nope"}), `{"messages":[{"role":"user","content":"x"}]}`, 502, "llm_invalid_output"},
		{"declined twice", refusing{}, `{"messages":[{"role":"user","content":"x"}]}`, 422, "llm_declined"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newIdeasServer(t, tt.client)
			resp := postJSON(t, srv.URL+"/api/v1/ideas", tt.body)
			if resp.StatusCode != tt.wantStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("status = %d, want %d: %s", resp.StatusCode, tt.wantStatus, body)
			}
			var e errorResponse
			if err := json.NewDecoder(resp.Body).Decode(&e); err != nil {
				t.Fatal(err)
			}
			if e.Error.Code != tt.wantCode || strings.TrimSpace(e.Error.Message) == "" {
				t.Fatalf("error = %+v, want code %q", e.Error, tt.wantCode)
			}
		})
	}
}

// exhausted returns a scripted fake with nothing left, which fails every call.
func exhausted() llm.Client {
	c := fake.New(llm.Response{Text: "used"})
	_, _ = c.Complete(nil, llm.Request{}) //nolint:staticcheck // nil ctx is fine for the fake
	return c
}

// refusing is an llm.Client whose safety layer declines every request.
type refusing struct{}

func (refusing) Complete(context.Context, llm.Request) (llm.Response, error) {
	return llm.Response{}, llm.ErrRefused
}
