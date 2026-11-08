// Package evals holds the evaluation set for the intent prompt and the
// judge that scores a result against a case. It has no LLM dependency:
// the eval runner (ideas/eval_test.go, build tag "eval") plugs in the
// real adapter; ordinary tests use the fake.
package evals

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"slices"
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
	// Mode is "major", "minor", "dorian", "mixolydian", "either", or several
	// of those joined with "/" when more than one would fit.
	Mode     string  `json:"mode"`
	TempoMin float64 `json:"tempoMin,omitempty"`
	TempoMax float64 `json:"tempoMax,omitempty"`
	// Feel, when set, is the feel the first candidate must have ("straight"
	// also matches an omitted feel).
	Feel string `json:"feel,omitempty"`
	// ModeInAny, when set, is a mode (or several joined with "/") that at
	// least one candidate must use. The first candidate is the plainest by
	// design, so a case that asks whether a mode is offered at all checks
	// every candidate.
	ModeInAny string `json:"modeInAny,omitempty"`
	// Wants are chords of which at least one should sound in the first
	// candidate, written as music.LiteralChord reads them: against the major
	// scale of the candidate's tonic, so "bVII" is the same sound in major,
	// mixolydian or minor. A miss is reported, not failed: some wants are not
	// yet expressible, and the report measures exactly that (ADR-0013).
	Wants []string `json:"wants,omitempty"`
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
	Title       string
	Key         string
	TempoBPM    float64
	Progression []string
	// Labels of every candidate, for the vocabulary metric.
	Labels []string
	// WantsHit is true when a wanted label appears; meaningful only when
	// the case has Wants.
	WantsHit bool
	// Unconstrained is the model's free-text wish (diagnostic runs only).
	Unconstrained string
	// Retry is why the first answer was not used, when there was a retry.
	Retry string
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
	v.Progression = first.Progression
	v.Unconstrained = res.Unconstrained
	v.Retry = res.Retry
	modes := make([]music.Mode, 0, len(res.Candidates))
	for _, cand := range res.Candidates {
		v.Labels = append(v.Labels, cand.Progression...)
		modes = append(modes, cand.Key.Mode)
	}
	v.WantsHit = wantsHit(c.Wants, first.Key.Tonic, chordsOf(first.Timeline))

	if problem := judgeMode(c.Mode, first.Key.Mode); problem != "" {
		v.Problems = append(v.Problems, problem)
	}
	if problem := judgeModeInAny(c.ModeInAny, modes); problem != "" {
		v.Problems = append(v.Problems, problem)
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
	b.WriteString("| case | first candidate | progression | result |\n| --- | --- | --- | --- |\n")
	for _, v := range verdicts {
		status := "✅"
		if !v.Pass {
			status = "❌ " + strings.Join(v.Problems, "; ")
		} else {
			pass++
		}
		if len(v.Case.Wants) > 0 {
			if v.WantsHit {
				status += " · wants ✓"
			} else {
				status += " · wants ✗ (" + strings.Join(v.Case.Wants, "/") + ")"
			}
		}
		if v.Retry != "" {
			status += " · retry: " + shorten(oneLine(v.Retry), 120)
		}
		fmt.Fprintf(&b, "| %s | %s (%s, %.0f BPM) | %s | %s |\n", v.Case.ID, v.Title, v.Key, v.TempoBPM, strings.Join(v.Progression, " "), status)
	}
	fmt.Fprintf(&b, "\n%d / %d passed\n", pass, len(verdicts))
	b.WriteString(breadthReport(verdicts))
	return b.String()
}

// wantsHit reports whether any wanted chord sounds in the progression.
// Wants are read literally against the tonic (music.LiteralChord) and
// compared with the chords the engine rendered, so a want names a sound,
// not a spelling: "bVII" matches mixolydian's VII, and "IV" in a minor key
// (the dorian colour) does not match "iv".
func wantsHit(wants []string, tonic music.PitchClass, chords []music.Chord) bool {
	for _, w := range wants {
		want, err := music.LiteralChord(tonic, w)
		if err != nil {
			continue // Load-time tests keep cases.json readable
		}
		if slices.Contains(chords, want) {
			return true
		}
	}
	return false
}

// chordsOf lists the chords a timeline sounds, in order.
func chordsOf(tl music.Timeline) []music.Chord {
	out := make([]music.Chord, 0, len(tl.Harmony))
	for _, h := range tl.Harmony {
		out = append(out, h.Chord)
	}
	return out
}

// shorten cuts s to at most n runes, marking the cut.
func shorten(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// breadth is what breadthReport counts across a whole run.
type breadth struct {
	progressions map[string]int
	keys         map[string]bool
	vocab        map[string]bool
	wantsTotal   int
	wantsHits    int
	retries      int
}

func measureBreadth(verdicts []Verdict) breadth {
	m := breadth{progressions: map[string]int{}, keys: map[string]bool{}, vocab: map[string]bool{}}
	for _, v := range verdicts {
		if len(v.Progression) == 0 {
			continue
		}
		m.progressions[strings.Join(v.Progression, " ")]++
		m.keys[v.Key] = true
		for _, label := range v.Labels {
			if degree, seventh, err := music.ParseRomanNumeral(label); err == nil {
				m.vocab[fmt.Sprintf("%d/%t", degree, seventh)] = true
			}
		}
		if len(v.Case.Wants) > 0 {
			m.wantsTotal++
			if v.WantsHit {
				m.wantsHits++
			}
		}
		if v.Retry != "" {
			m.retries++
		}
	}
	return m
}

// mostRepeated returns the progression seen most often and how often;
// ties go to the alphabetically first so the report is stable.
func (m breadth) mostRepeated() (string, int) {
	top, topN := "", 0
	for p, n := range m.progressions {
		if n > topN || (n == topN && p < top) {
			top, topN = p, n
		}
	}
	return top, topN
}

// breadthReport measures how varied the proposals were across the whole
// run (ADR-0013): distinct first progressions, the most repeated one,
// distinct keys, the share of the diatonic vocabulary used, wants, and
// what the model said it wanted to write.
func breadthReport(verdicts []Verdict) string {
	m := measureBreadth(verdicts)
	var b strings.Builder
	b.WriteString("\n### 提案の幅\n\n")
	fmt.Fprintf(&b, "- 最初の候補の進行の種類: %d（%d 件中）\n", len(m.progressions), len(verdicts))
	if top, n := m.mostRepeated(); n > 1 {
		fmt.Fprintf(&b, "- 最も多い進行: `%s`（%d 回）\n", top, n)
	}
	fmt.Fprintf(&b, "- 調の種類: %d\n", len(m.keys))
	fmt.Fprintf(&b, "- 使った語彙（度数 × 七の和音の有無）: %d / %d\n", len(m.vocab), diatonicVocabulary)
	if m.wantsTotal > 0 {
		fmt.Fprintf(&b, "- 期待した語彙（wants）の命中: %d / %d\n", m.wantsHits, m.wantsTotal)
	}
	fmt.Fprintf(&b, "- 再試行が要った件数: %d（理由は表の retry）\n", m.retries)
	b.WriteString(wishesReport(verdicts))
	return b.String()
}

// wishesReport lists what the model would have written without the
// engine's limits (diagnostic runs only).
func wishesReport(verdicts []Verdict) string {
	var rows []string
	for _, v := range verdicts {
		if v.Unconstrained != "" {
			rows = append(rows, fmt.Sprintf("| %s | %s |", v.Case.ID, oneLine(v.Unconstrained)))
		}
	}
	if len(rows) == 0 {
		return ""
	}
	return "\n### 制約がなければ書きたかったこと\n\n| case | unconstrained |\n| --- | --- |\n" + strings.Join(rows, "\n") + "\n"
}

// diatonicVocabulary counts the labels the engine can express in one key:
// seven degrees, each as a triad or a seventh chord.
const diatonicVocabulary = 14

func oneLine(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(s), "\n", " "), "|", "/")
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

// judgeModeInAny checks that some candidate uses one of the wanted modes.
func judgeModeInAny(want string, got []music.Mode) string {
	if want == "" {
		return ""
	}
	for _, m := range got {
		if judgeMode(want, m) == "" {
			return ""
		}
	}
	return "no candidate in " + want
}

// judgeMode accepts "either", one mode name, or several joined with "/".
func judgeMode(want string, got music.Mode) string {
	if want == "" || want == "either" {
		return ""
	}
	options := strings.Split(want, "/")
	for _, o := range options {
		var m music.Mode
		if err := m.UnmarshalText([]byte(o)); err != nil {
			return "case has unknown mode " + o
		}
		if m == got {
			return ""
		}
	}
	return "expected " + want + ", got " + got.String()
}
