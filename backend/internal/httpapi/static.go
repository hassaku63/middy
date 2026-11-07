package httpapi

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// Cache policy for the embedded frontend (ADR-0008). Vite emits hashed
// file names under /assets, which are safe to cache forever; everything
// else (index.html, the service worker, the manifest) must be revalidated
// on every load or PWA updates would never reach the user.
const (
	cacheImmutable = "public, max-age=31536000, immutable"
	cacheNever     = "no-cache"
)

// staticHandler serves the built frontend with SPA fallback: any path that
// is not a file and not under /api/ gets index.html. A nil fsys means the
// binary was built without the frontend; the UI then answers 503 while the
// API keeps working.
func staticHandler(fsys fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/") {
			writeError(w, http.StatusNotFound, "not_found", "no such endpoint")
			return
		}
		if fsys == nil {
			writeError(w, http.StatusServiceUnavailable, "frontend_not_built",
				"this server was built without the frontend; run `make build` or use `make dev-frontend`")
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "static files are read-only")
			return
		}

		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if name == "" || !isFile(fsys, name) {
			name = "index.html"
		}
		w.Header().Set("Cache-Control", cachePolicy(name))
		http.ServeFileFS(w, r, fsys, name)
	})
}

func isFile(fsys fs.FS, name string) bool {
	info, err := fs.Stat(fsys, name)
	return err == nil && !info.IsDir()
}

func cachePolicy(name string) string {
	if strings.HasPrefix(name, "assets/") {
		return cacheImmutable
	}
	return cacheNever
}
