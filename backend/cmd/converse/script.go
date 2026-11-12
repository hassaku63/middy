package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/hassaku63/middy/backend/internal/ideas"
)

// Step is one turn of a script: what the user says, and optionally what
// the answer should then be (#473, BL-314). An expectation is a direction
// to check, not a pass mark: the report counts the examples that meet it.
type Step struct {
	Say    string  `json:"say"`
	Expect *Expect `json:"expect,omitempty"`
}

// Expect is what a turn asks of every example of its answer. Each field is
// optional; an empty Expect checks nothing.
type Expect struct {
	// Meter is the time signature, e.g. "3/4".
	Meter string `json:"meter,omitempty"`
	// Bars is the length in whole bars.
	Bars int `json:"bars,omitempty"`
	// Modes are the modes any of which will do, e.g. ["minor", "dorian"].
	Modes []string `json:"modes,omitempty"`
	// Tempo is "up", "down" or "same", against the previous answer's first
	// example: a turn that asks for faster expects "up".
	Tempo string `json:"tempo,omitempty"`
	// Has are words each example must carry in its labels or patterns
	// (progression, style, accompaniment, drums, feel, tensions): "sus4"
	// kept, "bossa" kept, "swing" asked for. A word matches inside a label
	// ("V7" in "V7/vi").
	Has []string `json:"has,omitempty"`
}

// Script is a named conversation: a kind of user and the turns they say.
type Script struct {
	Name  string `json:"name"`
	Steps []Step `json:"turns"`
}

// readScript reads either a script object or, as turns-447.json is, a bare
// array of the user's messages (no expectations).
func readScript(data []byte) (Script, error) {
	var plain []string
	if err := json.Unmarshal(data, &plain); err == nil {
		s := Script{}
		for _, say := range plain {
			s.Steps = append(s.Steps, Step{Say: say})
		}
		return s, nil
	}
	var s Script
	if err := json.Unmarshal(data, &s); err != nil {
		return Script{}, err
	}
	if len(s.Steps) == 0 {
		return Script{}, errors.New("a script needs turns")
	}
	for i, st := range s.Steps {
		if strings.TrimSpace(st.Say) == "" {
			return Script{}, fmt.Errorf("turn %d says nothing", i+1)
		}
		if e := st.Expect; e != nil && e.Tempo != "" && !slices.Contains([]string{"up", "down", "same"}, e.Tempo) {
			return Script{}, fmt.Errorf("turn %d: tempo must be up, down or same, got %q", i+1, e.Tempo)
		}
	}
	return s, nil
}

// check says, for one example, which expectations it meets ("✓ meter")
// and which it does not ("✗ tempo up"), in a fixed order. prevTempo is the
// previous answer's first example's tempo, or 0 when there is none.
func check(e *Expect, c ideas.Candidate, prevTempo float64) []string {
	if e == nil {
		return nil
	}
	var out []string
	mark := func(ok bool, what string) {
		if ok {
			out = append(out, "✓ "+what)
		} else {
			out = append(out, "✗ "+what)
		}
	}
	if e.Meter != "" {
		mark(meterOf(c) == e.Meter, "meter "+e.Meter)
	}
	if e.Bars > 0 {
		mark(measure(c, "").Bars == strconv.Itoa(e.Bars), strconv.Itoa(e.Bars)+" bars")
	}
	if len(e.Modes) > 0 {
		mark(slices.Contains(e.Modes, c.Key.Mode.String()), strings.Join(e.Modes, "/"))
	}
	if e.Tempo != "" && prevTempo > 0 {
		var ok bool
		switch e.Tempo {
		case "up":
			ok = c.TempoBPM > prevTempo
		case "down":
			ok = c.TempoBPM < prevTempo
		default:
			ok = c.TempoBPM == prevTempo
		}
		mark(ok, "tempo "+e.Tempo)
	}
	fields := strings.Join(append(slices.Clone(c.Progression), c.Style, c.Accompaniment, c.Drums, c.Feel, c.Tensions), " ")
	for _, w := range e.Has {
		mark(strings.Contains(fields, w), w)
	}
	return out
}

func meterOf(c ideas.Candidate) string {
	ts := c.Timeline.TimeSignature
	return fmt.Sprintf("%d/%d", ts.Numerator, ts.Denominator)
}
