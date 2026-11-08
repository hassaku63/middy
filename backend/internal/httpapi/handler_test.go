package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
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
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

// get performs a GET and closes the body when the test ends.
func get(t *testing.T, url string) *http.Response {
	t.Helper()
	resp, err := http.Get(url) //nolint:gosec // test-local URL
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

func TestUnknownRouteIs404(t *testing.T) {
	srv := newTestServer(t)

	resp := get(t, srv.URL+"/api/v1/nope")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
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
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
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
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	h := requestLogger(logger, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if err := http.NewResponseController(w).Flush(); err != nil {
			t.Errorf("Flush through the logger: %v", err)
		}
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
}
