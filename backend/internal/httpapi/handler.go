// Package httpapi exposes the application over HTTP.
//
// Routing uses net/http method+path patterns (Go 1.22+). Handlers stay thin:
// decode, call a use case, encode. The request/response shapes are the
// contract documented in docs/api/openapi.yaml; keep both in sync.
package httpapi

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/hassaku63/middy/backend/internal/ideas"
	"github.com/hassaku63/middy/backend/internal/llm"
)

// Deps are the collaborators handlers need. Everything is an interface or a
// value so tests can construct a Handler without network or secrets.
type Deps struct {
	LLM    llm.Client
	Logger *slog.Logger
	// Static is the built frontend (see internal/web). nil means "not built":
	// the API still works and the UI answers 503.
	Static fs.FS
}

// route is one registered endpoint. Keeping routes as data lets a test
// compare them against docs/api/openapi.yaml (ADR-0004).
type route struct {
	Method  string
	Path    string // full path including the /api/v1 prefix
	Handler http.HandlerFunc
}

// APIPrefix is the base path of every versioned endpoint; it matches the
// `servers[0].url` of the OpenAPI document.
const APIPrefix = "/api/v1"

// routes returns every endpoint the server exposes.
func routes(d Deps) []route {
	return []route{
		{http.MethodGet, APIPrefix + "/health", handleHealth},
		{http.MethodPost, APIPrefix + "/timelines/from-progression", handleFromProgression},
		{http.MethodGet, APIPrefix + "/patterns", handlePatterns},
		{http.MethodPost, APIPrefix + "/export/midi", handleExportMIDI},
		{http.MethodPost, APIPrefix + "/ideas", handleIdeas(ideas.New(d.LLM), d.Logger)},
	}
}

// NewHandler builds the root http.Handler with all routes registered.
func NewHandler(d Deps) http.Handler {
	if d.Logger == nil {
		d.Logger = slog.Default()
	}
	mux := http.NewServeMux()
	allowed := map[string][]string{}
	for _, r := range routes(d) {
		mux.HandleFunc(r.Method+" "+r.Path, r.Handler)
		allowed[r.Path] = append(allowed[r.Path], r.Method)
	}
	// A known path with another method is 405, not the catch-all's 404:
	// "/" below matches every method, so the mux never says so itself.
	for p, methods := range allowed {
		mux.HandleFunc(p, methodNotAllowed(methods))
	}
	// Everything that is not an API route is the frontend (ADR-0008).
	mux.Handle("/", staticHandler(d.Static))
	return requestLogger(d.Logger, mux)
}

// methodNotAllowed answers a known path called with a method it does not take.
func methodNotAllowed(methods []string) http.HandlerFunc {
	allow := strings.Join(methods, ", ")
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Allow", allow)
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "this endpoint takes "+allow)
	}
}

// requestLogger is the only middleware: it logs every request with its
// status and duration, and turns a handler's panic into the JSON error
// envelope (500 internal_error) instead of a dropped connection. Add more
// (CORS) here rather than in main so tests exercise the same stack.
func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		defer func() {
			if v := recover(); v != nil {
				if v == http.ErrAbortHandler { //nolint:errorlint // sentinel value, compared as net/http does
					panic(v)
				}
				logger.Error("http handler panicked", "method", r.Method, "path", r.URL.Path, "panic", fmt.Sprint(v), "stack", string(debug.Stack()))
				if !rec.wrote {
					writeError(rec, http.StatusInternalServerError, "internal_error", "the server failed on this request")
				}
			}
			logger.Info("http", "method", r.Method, "path", r.URL.Path, "status", rec.status, "duration_ms", time.Since(start).Milliseconds())
		}()
		next.ServeHTTP(rec, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.wrote = true
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	s.wrote = true
	return s.ResponseWriter.Write(b)
}

// Unwrap lets http.ResponseController and optional interfaces (Flusher,
// ReaderFrom) reach the real writer.
func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }
