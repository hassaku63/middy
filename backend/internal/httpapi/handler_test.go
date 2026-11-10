package httpapi

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
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
		var e problem
		decodeErr := json.NewDecoder(resp.Body).Decode(&e)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound || decodeErr != nil || e.Code != "not_found" {
			t.Errorf("%s %s = %d %+v (%v), want 404 not_found", tc.method, tc.path, resp.StatusCode, e, decodeErr)
		}
	}
}

// TestWriteErrorIsAProblemDetail: an RFC 9457 problem detail, with the
// type's code as an extension member (ADR-0030).
func TestWriteErrorIsAProblemDetail(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, http.StatusBadRequest, "invalid_request", "bad input")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", ct)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := map[string]any{
		"type":   "https://github.com/hassaku63/middy/blob/main/docs/api/problems.md#invalid_request",
		"title":  "Invalid request",
		"status": float64(http.StatusBadRequest),
		"detail": "bad input",
		"code":   "invalid_request",
	}
	if !reflect.DeepEqual(body, want) {
		t.Fatalf("body = %v, want %v", body, want)
	}
}

func TestRequestLoggerTurnsAPanicIntoTheErrorEnvelope(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	h := requestLogger(logger, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("engine bug")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/ideas", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	var body problem
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Code != "internal_error" {
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
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))
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
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))
}

// TestKnownPathWithAnotherMethodIs405 (review): the catch-all "/" must not
// turn a wrong method on a real endpoint into "no such endpoint".
func TestKnownPathWithAnotherMethodIs405(t *testing.T) {
	srv := newTestServer(t)
	resp := get(t, srv.URL+"/api/v1/ideas")
	if resp.StatusCode != http.StatusMethodNotAllowed || resp.Header.Get("Allow") != http.MethodPost {
		t.Fatalf("GET /ideas = %d (Allow %q), want 405 with Allow POST", resp.StatusCode, resp.Header.Get("Allow"))
	}
	var body problem
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || body.Code != "method_not_allowed" {
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
		r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"a":1}`+tail))
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

// TestEveryResponseSaysNoSniff (#280, #206 A-4): answers and errors alike
// tell the browser not to guess their type.
func TestEveryResponseSaysNoSniff(t *testing.T) {
	srv := newTestServer(t)
	for _, path := range []string{"/api/v1/health", "/api/v1/nope", "/"} {
		resp := get(t, srv.URL+path)
		if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
			t.Errorf("%s: X-Content-Type-Options = %q, want nosniff", path, got)
		}
	}
}
