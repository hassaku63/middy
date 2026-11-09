package httpapi

import (
	"os"
	"strings"
	"testing"
)

const problemsDocPath = "../../../docs/api/problems.md"

// TestProblemTypesAreDocumented: every problem type the API answers with
// has a heading in docs/api/problems.md, which its type URI points to
// (ADR-0030).
func TestProblemTypesAreDocumented(t *testing.T) {
	doc, err := os.ReadFile(problemsDocPath)
	if err != nil {
		t.Fatal(err)
	}
	headings := map[string]bool{}
	for line := range strings.Lines(string(doc)) {
		if h, ok := strings.CutPrefix(strings.TrimRight(line, "\n"), "## "); ok {
			headings[h] = true
		}
	}
	for code, title := range problemTitles {
		if !headings[code] {
			t.Errorf("problem type %q has no heading in %s", code, problemsDocPath)
		}
		if title == "" {
			t.Errorf("problem type %q has no title", code)
		}
	}
}
