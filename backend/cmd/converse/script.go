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
// the answer should then take up (#473, BL-314). An expectation is a
// direction to check, not a pass mark: the report counts, per answer, how
// many of its examples take it up (ADR-0065).
type Step struct {
	Say    string  `json:"say"`
	Expect *Expect `json:"expect,omitempty"`
}

// Expect is what a turn asks of its answer. Each field is optional; an
// empty Expect checks nothing. An answer takes a request up when any of
// its examples does; how far it goes (some examples, some of the words) is
// within the tolerance of ADR-0065 and shows as △, not ✗.
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
	// Has are words the answer should carry in its examples' labels or
	// patterns (progression, style, accompaniment, drums, feel, tensions):
	// "sus4" kept, "bossa" kept, "swing" asked for. Each word is counted on
	// its own. A word matches inside a label ("V7" in "V7/vi").
	Has []string `json:"has,omitempty"`
	// HasAny are words of which any one will do: an example that carries
	// one of them takes the request up ("ii7", "iiø7" or "V7" for ii–V–I).
	HasAny []string `json:"hasAny,omitempty"`
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

// Tally is how one expectation fared in one answer: how many of its
// examples took it up (Met of Of, and which), and the mark ✓ (all), △
// (some: within the tolerance) or ✗ (none).
type Tally struct {
	What  string
	Met   int
	Of    int
	Which []int // 1-based examples that took it up, when not all did
}

// Mark is ✓ when every example took the request up, △ when some did, ✗
// when none did.
func (t Tally) Mark() string {
	switch {
	case t.Of > 0 && t.Met == t.Of:
		return "✓"
	case t.Met > 0:
		return "△"
	}
	return "✗"
}

// tally counts, for each expectation of a turn, the examples of its answer
// that take it up, in a fixed order. prevTempo is the previous answer's
// first example's tempo, or 0 when there is none (no tempo to compare).
func tally(e *Expect, cs []ideas.Candidate, prevTempo float64) []Tally {
	if e == nil {
		return nil
	}
	var out []Tally
	count := func(what string, ok func(ideas.Candidate) bool) {
		t := Tally{What: what, Of: len(cs)}
		for i, c := range cs {
			if ok(c) {
				t.Met++
				t.Which = append(t.Which, i+1)
			}
		}
		if t.Met == t.Of {
			t.Which = nil
		}
		out = append(out, t)
	}
	if e.Meter != "" {
		count("meter "+e.Meter, func(c ideas.Candidate) bool { return meterOf(c) == e.Meter })
	}
	if e.Bars > 0 {
		count(strconv.Itoa(e.Bars)+" bars", func(c ideas.Candidate) bool { return measure(c, "").Bars == strconv.Itoa(e.Bars) })
	}
	if len(e.Modes) > 0 {
		count(strings.Join(e.Modes, "/"), func(c ideas.Candidate) bool { return slices.Contains(e.Modes, c.Key.Mode.String()) })
	}
	if e.Tempo != "" && prevTempo > 0 {
		count("tempo "+e.Tempo, func(c ideas.Candidate) bool {
			switch e.Tempo {
			case "up":
				return c.TempoBPM > prevTempo
			case "down":
				return c.TempoBPM < prevTempo
			}
			return c.TempoBPM == prevTempo
		})
	}
	for _, w := range e.Has {
		count(w, func(c ideas.Candidate) bool { return strings.Contains(fieldsOf(c), w) })
	}
	if len(e.HasAny) > 0 {
		count(strings.Join(e.HasAny, " or "), func(c ideas.Candidate) bool {
			return slices.ContainsFunc(e.HasAny, func(w string) bool { return strings.Contains(fieldsOf(c), w) })
		})
	}
	return out
}

// fieldsOf is an example's labels and patterns, for the words of has.
func fieldsOf(c ideas.Candidate) string {
	return strings.Join(append(slices.Clone(c.Progression), c.Style, c.Accompaniment, c.Drums, c.Feel, c.Tensions), " ")
}

func meterOf(c ideas.Candidate) string {
	ts := c.Timeline.TimeSignature
	return fmt.Sprintf("%d/%d", ts.Numerator, ts.Denominator)
}
