// Package httpapi exposes the application over HTTP.
//
// Routing uses net/http method+path patterns (Go 1.22+). Handlers stay thin:
// decode, call a use case, encode. The request/response shapes are the
// contract documented in docs/api/openapi.yaml; keep both in sync.
package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/hassaku63/middy/backend/internal/llm"
)

// Deps are the collaborators handlers need. Everything is an interface or a
// value so tests can construct a Handler without network or secrets.
type Deps struct {
	LLM    llm.Client
	Logger *slog.Logger
}

// NewHandler builds the root http.Handler with all routes registered.
func NewHandler(d Deps) http.Handler {
	if d.Logger == nil {
		d.Logger = slog.Default()
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", handleHealth)
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
