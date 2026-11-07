// Package httpapi exposes the application over HTTP.
//
// Routing uses net/http method+path patterns (Go 1.22+). Handlers stay thin:
// decode, call a use case, encode. The request/response shapes are the
// contract documented in docs/api/openapi.yaml; keep both in sync.
package httpapi

import (
	"io/fs"
	"log/slog"
	"net/http"

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
		{http.MethodPost, APIPrefix + "/ideas", handleIdeas(ideas.New(d.LLM))},
	}
}

// NewHandler builds the root http.Handler with all routes registered.
func NewHandler(d Deps) http.Handler {
	if d.Logger == nil {
		d.Logger = slog.Default()
	}
	mux := http.NewServeMux()
	for _, r := range routes(d) {
		mux.HandleFunc(r.Method+" "+r.Path, r.Handler)
	}
	// Everything that is not an API route is the frontend (ADR-0008).
	mux.Handle("/", staticHandler(d.Static))
	return requestLogger(d.Logger, mux)
}

// requestLogger is the only middleware for now. Add more (recover, CORS)
// here rather than in main so tests exercise the same stack.
func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logger.Info("http", "method", r.Method, "path", r.URL.Path, "status", rec.status)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}
