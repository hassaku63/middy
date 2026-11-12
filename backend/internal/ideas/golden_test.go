package ideas

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the golden files in testdata/golden")

// TestGoldenSystemPrompt fixes the system prompt, catalogue included
// (BL-176, #207): moving the catalogue to data must not change what the
// model is told. A prompt change refreshes it on purpose with
// go test ./internal/ideas -run Golden -update, and an eval label as before.
func TestGoldenSystemPrompt(t *testing.T) {
	got := SystemPrompt()
	path := filepath.Join("testdata", "golden", "system-prompt.txt")
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path) //nolint:gosec // the repository's own golden files
	if err != nil {
		t.Fatalf("%v (run with -update to create it)", err)
	}
	if got != string(want) {
		t.Fatalf("the system prompt differs from %s (run with -update if intended)", path)
	}
}
