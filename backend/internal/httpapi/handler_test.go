package httpapi

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/llm/fake"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	h := NewHandler(Deps{
		LLM:    fake.New(),
		Logger: slog.New(slog.DiscardHandler),
	})
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

// get performs a GET and closes the body when the test ends.
func get(t *testing.T, url string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func TestHealth(t *testing.T) {
	srv := newTestServer(t)

	resp := get(t, srv.URL+"/api/v1/health")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("content-type = %q", ct)
	}
	var body healthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Status != "ok" || body.Version == "" {
		t.Fatalf("unexpected body %+v", body)
	}
}

// TestUnknownRouteIs404: the server answers the API only (ADR-0027). Any
// other path, a screen's included, is the JSON 404, whatever the method.
func TestUnknownRouteIs404(t *testing.T) {
	srv := newTestServer(t)

	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/api/v1/nope"},
		{http.MethodGet, "/api"},
		{http.MethodGet, "/"},
		{http.MethodGet, "/projects/42"},
		{http.MethodGet, "/assets/index.js"},
		{http.MethodPost, "/some/screen"},
	} {
		req, err := http.NewRequestWithContext(t.Context(), tc.method, srv.URL+tc.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		var e errorResponse
		decodeErr := json.NewDecoder(resp.Body).Decode(&e)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound || decodeErr != nil || e.Error.Code != "not_found" {
			t.Errorf("%s %s = %d %+v (%v), want 404 not_found", tc.method, tc.path, resp.StatusCode, e, decodeErr)
		}
	}
}

func TestWriteErrorEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, http.StatusBadRequest, "invalid_request", "bad input")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	var body errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error.Code != "invalid_request" || body.Error.Message != "bad input" {
		t.Fatalf("unexpected body %+v", body)
	}
}

func TestRequestLoggerTurnsAPanicIntoTheErrorEnvelope(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	h := requestLogger(logger, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("engine bug")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/ideas", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	var body errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Error.Code != "internal_error" {
		t.Fatalf("body = %s (%v), want the internal_error envelope", rec.Body.String(), err)
	}
	for _, want := range []string{`"panic":"engine bug"`, `"status":500`, `"duration_ms"`} {
		if !strings.Contains(logs.String(), want) {
			t.Errorf("logs lack %s:\n%s", want, logs.String())
		}
	}
}

func TestRequestLoggerLeavesAStartedResponseAlone(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	h := requestLogger(logger, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		panic("late")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusAccepted || rec.Body.Len() != 0 {
		t.Fatalf("got %d %q, want the handler's own 202 with nothing appended", rec.Code, rec.Body.String())
	}
}

func TestRequestLoggerReachesTheRealWriter(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	h := requestLogger(logger, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if err := http.NewResponseController(w).Flush(); err != nil {
			t.Errorf("Flush through the logger: %v", err)
		}
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
}

// TestKnownPathWithAnotherMethodIs405 (review): the catch-all "/" must not
// turn a wrong method on a real endpoint into "no such endpoint".
func TestKnownPathWithAnotherMethodIs405(t *testing.T) {
	srv := newTestServer(t)
	resp := get(t, srv.URL+"/api/v1/ideas")
	if resp.StatusCode != http.StatusMethodNotAllowed || resp.Header.Get("Allow") != http.MethodPost {
		t.Fatalf("GET /ideas = %d (Allow %q), want 405 with Allow POST", resp.StatusCode, resp.Header.Get("Allow"))
	}
	var body errorResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || body.Error.Code != "method_not_allowed" {
		t.Fatalf("body = %+v (%v), want the method_not_allowed envelope", body, err)
	}
	if resp := get(t, srv.URL+"/api/v1/nope"); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown path = %d, want 404", resp.StatusCode)
	}
}

// TestTrailingClosersAreRejected (review): json.Decoder.More is false before
// a stray "}" or "]", so those must still count as trailing data.
func TestTrailingClosersAreRejected(t *testing.T) {
	for _, tail := range []string{"}", "]", " {}", ""} {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"a":1}`+tail))
		var v map[string]int
		err := decodeJSON(r, &v)
		if tail == "" && err != nil {
			t.Errorf("a clean body: %v", err)
		}
		if tail != "" && err == nil {
			t.Errorf("trailing %q accepted", tail)
		}
	}
}
