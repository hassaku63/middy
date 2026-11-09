package intent

import (
	"encoding/json"
	"fmt"
	"testing"
)

// supportedKeywords is the JSON Schema subset the Messages API accepts for
// structured outputs (output_config.format). Anything else is a 400 at
// request time, which no fake-backed test can see; this test is the guard.
// Source: platform.claude.com, build-with-claude/structured-outputs,
// "JSON Schema limitations".
var supportedKeywords = map[string]bool{
	"type": true, "properties": true, "required": true, "additionalProperties": true,
	"items": true, "enum": true, "const": true, "anyOf": true, "allOf": true,
	"$ref": true, "$defs": true, "definitions": true, "default": true,
	"format": true, "pattern": true, "description": true, "minItems": true,
}

func TestSchemaUsesOnlySupportedKeywords(t *testing.T) {
	var root map[string]any
	if err := json.Unmarshal(Schema, &root); err != nil {
		t.Fatalf("schema is not JSON: %v", err)
	}
	for _, problem := range checkSubset("$", root) {
		t.Error(problem)
	}
}

func TestSubsetCheckerCatchesUnsupportedKeywords(t *testing.T) {
	bad := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"tags":  map[string]any{"type": "array", "maxItems": 3.0, "items": map[string]any{"type": "string"}},
			"n":     map[string]any{"type": "integer", "minimum": 0.0},
			"list":  map[string]any{"type": "array", "minItems": 2.0},
			"inner": map[string]any{"type": "object", "properties": map[string]any{}},
		},
		"additionalProperties": false,
	}
	if got := len(checkSubset("$", bad)); got != 4 {
		t.Fatalf("want 4 problems (maxItems, minimum, minItems 2, missing additionalProperties), got %d: %v", got, checkSubset("$", bad))
	}
}

// checkSubset walks a schema node and reports unsupported keywords,
// minItems other than 0 or 1, and objects without additionalProperties false.
func checkSubset(path string, node map[string]any) []string {
	var problems []string
	for key := range node {
		if !supportedKeywords[key] {
			problems = append(problems, fmt.Sprintf("%s: keyword %q is not supported by structured outputs", path, key))
		}
	}
	if v, ok := node["minItems"]; ok && v != 0.0 && v != 1.0 {
		problems = append(problems, fmt.Sprintf("%s: minItems must be 0 or 1, got %v", path, v))
	}
	if node["type"] == "object" && node["additionalProperties"] != false {
		problems = append(problems, fmt.Sprintf("%s: objects need additionalProperties: false", path))
	}
	if props, ok := node["properties"].(map[string]any); ok {
		for name, child := range props {
			if c, ok := child.(map[string]any); ok {
				problems = append(problems, checkSubset(path+"."+name, c)...)
			}
		}
	}
	if items, ok := node["items"].(map[string]any); ok {
		problems = append(problems, checkSubset(path+"[]", items)...)
	}
	return problems
}

func TestDiagnosticSchemaStaysWithinTheSubset(t *testing.T) {
	var root map[string]any
	if err := json.Unmarshal(DiagnosticSchema, &root); err != nil {
		t.Fatal(err)
	}
	for _, problem := range checkSubset("$", root) {
		t.Error(problem)
	}
	props := as[map[string]any](t, root["properties"])
	if _, ok := props["unconstrained"]; !ok {
		t.Fatal("the diagnostic schema must ask for the unconstrained field")
	}
	for _, req := range as[[]any](t, root["required"]) {
		if req == "unconstrained" {
			t.Fatal("unconstrained must stay optional")
		}
	}
	var prod map[string]any
	if err := json.Unmarshal(Schema, &prod); err != nil {
		t.Fatal(err)
	}
	if _, ok := prod["properties"].(map[string]any)["unconstrained"]; ok {
		t.Fatal("the production schema must not ask for it")
	}
}

// as asserts v's dynamic type and fails the test with what it found.
func as[T any](t *testing.T, v any) T {
	t.Helper()
	got, ok := v.(T)
	if !ok {
		t.Fatalf("got %T, want %T", v, got)
	}
	return got
}
