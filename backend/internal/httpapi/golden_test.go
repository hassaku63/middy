package httpapi

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the golden files in testdata/golden")

// TestGoldenPatternsResponse fixes GET /patterns byte for byte (BL-176,
// #207): moving the catalogue to data must not change what the API says.
// Refresh on purpose with go test ./internal/httpapi -run Golden -update.
func TestGoldenPatternsResponse(t *testing.T) {
	srv := newTestServer(t)
	resp := get(t, srv.URL+"/api/v1/patterns")
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		t.Fatal(err)
	}
	pretty.WriteByte('\n')
	path := filepath.Join("testdata", "golden", "patterns.json")
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, pretty.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path) //nolint:gosec // the repository's own golden files
	if err != nil {
		t.Fatalf("%v (run with -update to create it)", err)
	}
	if !bytes.Equal(pretty.Bytes(), want) {
		t.Fatalf("GET /patterns differs from %s (run with -update if intended)", path)
	}
}
