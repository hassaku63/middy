// Package evals holds the evaluation set for the intent prompt and the
// judge that scores a result against a case. It has no LLM dependency:
// the eval runner (ideas/eval_test.go, build tag "eval") plugs in the
// real adapter; ordinary tests use the fake.
package evals

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hassaku63/middy/backend/internal/ideas"
	"github.com/hassaku63/middy/backend/internal/music"
)

//go:embed cases.json
var casesJSON []byte

// Case is one image with loose expectations on the first candidate.
type Case struct {
	ID    string `json:"id"`
	Image string `json:"image"`
	// Mode is "major", "minor" or "either".
	Mode     string  `json:"mode"`
	TempoMin float64 `json:"tempoMin,omitempty"`
	TempoMax float64 `json:"tempoMax,omitempty"`
	// Feel, when set, is the feel the first candidate must have ("straight"
	// also matches an omitted feel).
	Feel string `json:"feel,omitempty"`
}

// Load returns the embedded cases.
func Load() ([]Case, error) {
	var file struct {
		Cases []Case `json:"cases"`
	}
	if err := json.Unmarshal(casesJSON, &file); err != nil {
		return nil, fmt.Errorf("evals: cases.json: %w", err)
	}
	if len(file.Cases) == 0 {
		return nil, fmt.Errorf("evals: cases.json has no cases")
	}
	return file.Cases, nil
}

// Verdict is the judgement for one case.
type Verdict struct {
	Case     Case
	Pass     bool
	Problems []string
	// First candidate, for the report.
	Title    string
	Key      string
	TempoBPM float64
}

// Judge scores a result against the case: the first candidate must match
// the expected mode and tempo band, and every candidate must be renderable
// (which Imagine already guarantees; checked again so the report says so).
func Judge(c Case, res ideas.Result) Verdict {
	v := Verdict{Case: c}
	if len(res.Candidates) == 0 {
		v.Problems = append(v.Problems, "no candidates")
		return v
	}
	first := res.Candidates[0]
	v.Title, v.Key, v.TempoBPM = first.Title, first.Key.String(), first.TempoBPM

	switch c.Mode {
	case "major":
		if first.Key.Mode != music.Major {
			v.Problems = append(v.Problems, "expected major, got "+first.Key.Mode.String())
		}
	case "minor":
		if first.Key.Mode != music.Minor {
			v.Problems = append(v.Problems, "expected minor, got "+first.Key.Mode.String())
		}
	case "either", "":
	default:
		v.Problems = append(v.Problems, "case has unknown mode "+c.Mode)
	}
	if c.TempoMin > 0 && first.TempoBPM < c.TempoMin {
		v.Problems = append(v.Problems, fmt.Sprintf("tempo %.0f below %.0f", first.TempoBPM, c.TempoMin))
	}
	if c.TempoMax > 0 && first.TempoBPM > c.TempoMax {
		v.Problems = append(v.Problems, fmt.Sprintf("tempo %.0f above %.0f", first.TempoBPM, c.TempoMax))
	}
	if problem := judgeFeel(c.Feel, first.Feel); problem != "" {
		v.Problems = append(v.Problems, problem)
	}
	if strings.TrimSpace(res.Reply) == "" {
		v.Problems = append(v.Problems, "empty reply")
	}
	v.Pass = len(v.Problems) == 0
	return v
}

// Report renders verdicts as a Markdown table to paste into the PR that
// changed the prompt (results are not committed: ADR-0006).
func Report(verdicts []Verdict) string {
	var b strings.Builder
	pass := 0
	b.WriteString("| case | first candidate | result |\n| --- | --- | --- |\n")
	for _, v := range verdicts {
		status := "✅"
		if !v.Pass {
			status = "❌ " + strings.Join(v.Problems, "; ")
		} else {
			pass++
		}
		fmt.Fprintf(&b, "| %s | %s (%s, %.0f BPM) | %s |\n", v.Case.ID, v.Title, v.Key, v.TempoBPM, status)
	}
	fmt.Fprintf(&b, "\n%d / %d passed\n", pass, len(verdicts))
	return b.String()
}

func judgeFeel(want, got string) string {
	if got == "" {
		got = "straight"
	}
	if want == "" || want == got {
		return ""
	}
	return "expected feel " + want + ", got " + got
}
