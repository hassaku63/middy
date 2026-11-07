package httpapi

import (
	"bufio"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/llm/fake"
)

// openAPIPath is where the API contract lives relative to this package.
const openAPIPath = "../../../docs/api/openapi.yaml"

// TestRoutesMatchOpenAPI enforces ADR-0004: every registered route is
// documented in docs/api/openapi.yaml and every documented operation is
// implemented. It parses the YAML line-by-line, relying on the document's
// two-space indentation (`  /path:` under `paths:`, `    get:` under a path),
// so that no YAML dependency is needed.
func TestRoutesMatchOpenAPI(t *testing.T) {
	documented := operationsFromOpenAPI(t)

	var implemented []string
	for _, r := range routes(Deps{LLM: fake.New()}) {
		implemented = append(implemented, strings.ToLower(r.Method)+" "+strings.TrimPrefix(r.Path, APIPrefix))
	}
	sort.Strings(implemented)
	sort.Strings(documented)

	if strings.Join(implemented, "\n") != strings.Join(documented, "\n") {
		t.Fatalf("routes and %s disagree\nimplemented:\n  %s\ndocumented:\n  %s",
			openAPIPath, strings.Join(implemented, "\n  "), strings.Join(documented, "\n  "))
	}
}

func operationsFromOpenAPI(t *testing.T) []string {
	t.Helper()
	f, err := os.Open(openAPIPath)
	if err != nil {
		t.Fatalf("open OpenAPI document: %v", err)
	}
	defer func() { _ = f.Close() }()

	methods := map[string]bool{"get": true, "post": true, "put": true, "patch": true, "delete": true}
	var ops []string
	inPaths := false
	currentPath := ""
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		switch {
		case line == "paths:":
			inPaths = true
		case inPaths && len(line) > 0 && line[0] != ' ':
			inPaths = false // next top-level key (components:, ...)
		case inPaths && strings.HasPrefix(line, "  /") && strings.HasSuffix(line, ":"):
			currentPath = strings.TrimSuffix(strings.TrimSpace(line), ":")
		case inPaths && strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "     "):
			method := strings.TrimSuffix(strings.TrimSpace(line), ":")
			if methods[method] && currentPath != "" {
				ops = append(ops, method+" "+currentPath)
			}
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("read OpenAPI document: %v", err)
	}
	if len(ops) == 0 {
		t.Fatalf("no operations found in %s; parser assumptions may be stale", openAPIPath)
	}
	return ops
}
