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
	// Drums, when set, lists the drums the first candidate may use, joined
	// with "/"; "none" allows no drums (BL-081).
	Drums string `json:"drums,omitempty"`
	// Meter, when set, is the meter the first candidate must have ("4/4"
	// also matches an omitted meter; ADR-0017).
	Meter string `json:"meter,omitempty"`
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
	Meter       string
	Patterns    string // "accompaniment/drums", "" for block chords and no drums
	Progression []string
	// Labels of every candidate, for the vocabulary metric.
	Labels []string
	// WantsHit is true when a wanted chord sounds in the first candidate;
	// WantsIn is the 1-based candidate where one first sounds (0: none).
	// Both are meaningful only when the case has Wants. They separate "the
	// model never offers it" from "it offers it, but not first".
	WantsHit bool
	WantsIn  int
	// Unconstrained is the model's free-text wish (diagnostic runs only).
	Unconstrained string
	// Retry is why the first answer was not used, when there was a retry.
	Retry string
	// Silent lists the words in the candidates' reasons that name what
	// middy never plays (#128); reported, not failed.
	Silent []string
}

// silentVocabulary names what the examples never sound: instruments beyond
// the piano, guitar and drums of the pattern catalogue, and a melody.
var silentVocabulary = []string{
	"フィドル", "バイオリン", "ヴァイオリン", "ホイッスル", "フルート", "笛", "ストリングス", "弦楽",
	"バグパイプ", "ハープ", "アコーディオン", "サックス", "トランペット", "ブラス", "メロディ", "旋律",
}

// silentWords returns, sorted and once each, the silent words the reasons use.
func silentWords(reasons []string) []string {
	var out []string
	for _, w := range silentVocabulary {
		for _, r := range reasons {
			if strings.Contains(r, w) {
				out = append(out, w)
				break
			}
		}
	}
	slices.Sort(out)
	return out
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
	v.Title, v.Key, v.TempoBPM, v.Meter = first.Title, first.Key.String(), first.TempoBPM, first.Meter
	v.Progression = withBeats(first.Progression, first.Beats)
	v.Patterns = patternsOf(first.Accompaniment, first.Drums)
	v.Unconstrained = res.Unconstrained
	v.Retry = res.Retry
	modes := make([]music.Mode, 0, len(res.Candidates))
	reasons := make([]string, 0, len(res.Candidates))
	for i, cand := range res.Candidates {
		reasons = append(reasons, cand.Reason)
		v.Labels = append(v.Labels, cand.Progression...)
		modes = append(modes, cand.Key.Mode)
		if v.WantsIn == 0 && wantsHit(c.Wants, cand.Key.Tonic, chordsOf(cand.Timeline)) {
			v.WantsIn = i + 1
		}
	}
	v.WantsHit = v.WantsIn == 1
	v.Silent = silentWords(reasons)

	v.Problems = append(v.Problems, judgeFirst(c, first)...)
	if problem := judgeModeInAny(c.ModeInAny, modes); problem != "" {
		v.Problems = append(v.Problems, problem)
	}
	if strings.TrimSpace(res.Reply) == "" {
		v.Problems = append(v.Problems, "empty reply")
	}
	v.Pass = len(v.Problems) == 0
	return v
}

// judgeFirst checks the first candidate against the case: mode, tempo
// band, and, when the case names them, feel, drums and meter.
func judgeFirst(c Case, first ideas.Candidate) []string {
	var problems []string
	add := func(p string) {
		if p != "" {
			problems = append(problems, p)
		}
	}
	add(judgeMode(c.Mode, first.Key.Mode))
	if c.TempoMin > 0 && first.TempoBPM < c.TempoMin {
		add(fmt.Sprintf("tempo %.0f below %.0f", first.TempoBPM, c.TempoMin))
	}
	if c.TempoMax > 0 && first.TempoBPM > c.TempoMax {
		add(fmt.Sprintf("tempo %.0f above %.0f", first.TempoBPM, c.TempoMax))
	}
	add(judgeFeel(c.Feel, first.Feel))
	add(judgeDrums(c.Drums, first.Drums))
	if c.Meter != "" && c.Meter != meterOf(first.Meter) {
		add("expected meter " + c.Meter + ", got " + meterOf(first.Meter))
	}
	return problems
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
			switch {
			case v.WantsHit:
				status += " · wants ✓"
			case v.WantsIn > 0:
				status += fmt.Sprintf(" · wants ✓ in candidate %d", v.WantsIn)
			default:
				status += " · wants ✗ (" + strings.Join(v.Case.Wants, "/") + ")"
			}
		}
		if len(v.Silent) > 0 {
			status += " · reason names " + strings.Join(v.Silent, "/")
		}
		if v.Retry != "" {
			status += " · retry: " + shorten(oneLine(v.Retry), 120)
		}
		fmt.Fprintf(&b, "| %s | %s (%s, %.0f BPM%s) | %s | %s |\n", v.Case.ID, v.Title, v.Key, v.TempoBPM, meterNote(v.Meter)+patternsNote(v.Patterns), strings.Join(v.Progression, " "), status)
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
	wantsAny     int
	retries      int
	silent       int
}

func measureBreadth(verdicts []Verdict) breadth {
	m := breadth{progressions: map[string]int{}, keys: map[string]bool{}, vocab: map[string]bool{}}
	for _, v := range verdicts {
		if len(v.Silent) > 0 {
			m.silent++
		}
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
			if v.WantsIn > 0 {
				m.wantsAny++
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
		fmt.Fprintf(&b, "- 期待した語彙（wants）の命中: 最初の候補で %d / %d、どれかの候補で %d / %d\n", m.wantsHits, m.wantsTotal, m.wantsAny, m.wantsTotal)
	}
	fmt.Fprintf(&b, "- 再試行が要った件数: %d（理由は表の retry）\n", m.retries)
	fmt.Fprintf(&b, "- 説明が鳴らないものに触れた件数: %d（楽器やメロディ。表の reason names。#128）\n", m.silent)
	b.WriteString(wishesReport(verdicts))
	return b.String()
}

// wishesReport lists what the model would have written without the
// engine's limits (diagnostic runs only).
func wishesReport(verdicts []Verdict) string {
	var rows []string
	for _, v := range verdicts {
		if w := strings.TrimSpace(v.Unconstrained); w != "" && w != "なし" {
			rows = append(rows, fmt.Sprintf("| %s | %s |", v.Case.ID, oneLine(v.Unconstrained)))
		}
	}
	if len(rows) == 0 {
		return ""
	}
	return "\n### 表せないために使わなかったもの\n\n| case | unconstrained |\n| --- | --- |\n" + strings.Join(rows, "\n") + "\n"
}

// diatonicVocabulary counts the labels the engine can express in one key:
// seven degrees, each as a triad or a seventh chord.
const diatonicVocabulary = 14

func oneLine(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(s), "\n", " "), "|", "/")
}

// judgeDrums accepts any drums when want is empty, else one of the ids
// joined with "/" ("none" for no drums).
func judgeDrums(want, got string) string {
	if want == "" {
		return ""
	}
	if got == "" {
		got = "none"
	}
	if slices.Contains(strings.Split(want, "/"), got) {
		return ""
	}
	return "expected drums " + want + ", got " + got
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

func meterOf(m string) string {
	if m == "" {
		return "4/4"
	}
	return m
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

// meterNote marks a first candidate that is not in 4/4.
// withBeats marks each label with its length ("i:8") when the candidate
// gives per-chord beats (BL-109); otherwise it returns the labels as they are.
func withBeats(labels []string, beats []int) []string {
	if len(beats) != len(labels) {
		return labels
	}
	out := make([]string, len(labels))
	for i, l := range labels {
		out[i] = fmt.Sprintf("%s:%d", l, beats[i])
	}
	return out
}

// patternsOf names the patterns a candidate chose: "reel/bodhran",
// "strum" (no drums), "block/eight-beat" (drums alone), or "".
func patternsOf(acc, drums string) string {
	switch {
	case acc == "" && drums == "":
		return ""
	case drums == "":
		return acc
	case acc == "":
		return "block/" + drums
	}
	return acc + "/" + drums
}

func patternsNote(p string) string {
	if p == "" {
		return ""
	}
	return ", " + p
}

func meterNote(m string) string {
	if m == "" || m == "4/4" {
		return ""
	}
	return ", " + m
}
