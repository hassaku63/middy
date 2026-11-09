package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/hassaku63/middy/backend/internal/llm/fake"
)

func builtFrontend() fstest.MapFS {
	return fstest.MapFS{
		"index.html":             {Data: []byte("<!doctype html><title>middy</title>")},
		"sw.js":                  {Data: []byte("// service worker")},
		"manifest.webmanifest":   {Data: []byte(`{"name":"middy"}`)},
		"assets/index-abc123.js": {Data: []byte("console.log('app')")},
		"assets/sub/dir.txt":     {Data: []byte("x")},
	}
}

func newStaticServer(t *testing.T, static fstest.MapFS) *httptest.Server {
	t.Helper()
	deps := Deps{LLM: fake.New(), Logger: slog.New(slog.DiscardHandler)}
	if static != nil {
		deps.Static = static
	}
	srv := httptest.NewServer(NewHandler(deps))
	t.Cleanup(srv.Close)
	return srv
}

func TestStaticServesFilesWithCachePolicy(t *testing.T) {
	srv := newStaticServer(t, builtFrontend())
	tests := []struct {
		path      string
		wantBody  string
		wantCache string
		wantType  string
	}{
		{"/", "<!doctype html><title>middy</title>", "no-cache", "text/html"},
		{"/index.html", "<!doctype html><title>middy</title>", "no-cache", "text/html"},
		{"/sw.js", "// service worker", "no-cache", "javascript"},
		{"/manifest.webmanifest", `{"name":"middy"}`, "no-cache", ""},
		{"/assets/index-abc123.js", "console.log('app')", "public, max-age=31536000, immutable", "javascript"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			resp := get(t, srv.URL+tt.path)
			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status = %d", resp.StatusCode)
			}
			if string(body) != tt.wantBody {
				t.Errorf("body = %q", body)
			}
			if cc := resp.Header.Get("Cache-Control"); cc != tt.wantCache {
				t.Errorf("cache-control = %q, want %q", cc, tt.wantCache)
			}
			if tt.wantType != "" && !contains(resp.Header.Get("Content-Type"), tt.wantType) {
				t.Errorf("content-type = %q, want it to contain %q", resp.Header.Get("Content-Type"), tt.wantType)
			}
		})
	}
}

func TestStaticSPAFallback(t *testing.T) {
	srv := newStaticServer(t, builtFrontend())
	for _, p := range []string{"/projects/42", "/deep/../index.html", "/"} {
		t.Run(p, func(t *testing.T) {
			resp := get(t, srv.URL+p)
			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != http.StatusOK || string(body) != "<!doctype html><title>middy</title>" {
				t.Fatalf("status = %d body = %q, want index.html", resp.StatusCode, body)
			}
			if cc := resp.Header.Get("Cache-Control"); cc != "no-cache" {
				t.Errorf("fallback must not be cached, got %q", cc)
			}
		})
	}
}

// TestStaticMissingFileIs404 (ADR-0022): a build file that does not exist
// (an old tab after a deploy) is missing, not the app's HTML.
func TestStaticMissingFileIs404(t *testing.T) {
	srv := newStaticServer(t, builtFrontend())
	for _, p := range []string{"/assets/missing.js", "/assets/sub", "/assets", "/missing.css", "/icons/x.png"} {
		t.Run(p, func(t *testing.T) {
			resp := get(t, srv.URL+p)
			if resp.StatusCode != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", resp.StatusCode)
			}
			if ct := resp.Header.Get("Content-Type"); contains(ct, "text/html") {
				t.Errorf("a missing file must not be answered with HTML (%s)", ct)
			}
			if cc := resp.Header.Get("Cache-Control"); cc != "no-cache" {
				t.Errorf("a 404 must not be cached, got %q", cc)
			}
		})
	}
}

func TestStaticNeverShadowsTheAPI(t *testing.T) {
	srv := newStaticServer(t, builtFrontend())

	resp := get(t, srv.URL+"/api/v1/health")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health status = %d", resp.StatusCode)
	}
	for _, p := range []string{"/api/v1/nope", "/api/v1", "/api"} {
		resp := get(t, srv.URL+p)
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("%s: status = %d, want 404 (not index.html)", p, resp.StatusCode)
		}
		var e errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&e); err != nil || e.Error.Code != "not_found" {
			t.Errorf("%s: want JSON not_found envelope, got %+v (%v)", p, e, err)
		}
	}
}

func TestStaticWithoutBuildIs503(t *testing.T) {
	srv := newStaticServer(t, nil)

	resp := get(t, srv.URL+"/")
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", resp.StatusCode)
	}
	var e errorResponse
	if err := json.NewDecoder(resp.Body).Decode(&e); err != nil || e.Error.Code != "frontend_not_built" {
		t.Fatalf("want frontend_not_built envelope, got %+v (%v)", e, err)
	}
	if health := get(t, srv.URL+"/api/v1/health"); health.StatusCode != http.StatusOK {
		t.Fatalf("API must keep working without the frontend, got %d", health.StatusCode)
	}
}

func TestStaticRejectsWrites(t *testing.T) {
	srv := newStaticServer(t, builtFrontend())
	resp := postJSON(t, srv.URL+"/index.html", "{}")
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", resp.StatusCode)
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
