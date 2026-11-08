package evals

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestRunAllSendsTheFirstCaseAlone checks the order the runner promises: the
// first case finishes before any other starts (so the rest can read the
// cached system prompt it wrote), the rest run up to the limit at once, and
// the verdicts come back in the order of the cases.
func TestRunAllSendsTheFirstCaseAlone(t *testing.T) {
	const limit = 3
	cases := make([]Case, 1+2*limit)
	for i := range cases {
		cases[i] = Case{ID: fmt.Sprintf("c%d", i)}
	}

	var (
		mu           sync.Mutex
		inFlight     int
		maxFlight    int
		firstDone    bool
		startedEarly []string
		release      = make(chan struct{})
		once         sync.Once
	)
	one := func(c Case) Verdict {
		mu.Lock()
		if c.ID != "c0" && !firstDone {
			startedEarly = append(startedEarly, c.ID)
		}
		inFlight++
		maxFlight = max(maxFlight, inFlight)
		if inFlight == limit {
			once.Do(func() { close(release) })
		}
		mu.Unlock()

		if c.ID != "c0" {
			// Hold every later case until `limit` of them are in flight,
			// which a sequential runner would never reach.
			select {
			case <-release:
			case <-time.After(5 * time.Second):
				t.Errorf("%s: never saw %d cases in flight at once", c.ID, limit)
			}
		}

		mu.Lock()
		inFlight--
		if c.ID == "c0" {
			firstDone = true
		}
		mu.Unlock()
		return Verdict{Case: c}
	}

	got := runAll(cases, limit, one)

	if len(startedEarly) > 0 {
		t.Errorf("started before the first case finished: %v", startedEarly)
	}
	if maxFlight != limit {
		t.Errorf("at most %d cases should be in flight, saw %d", limit, maxFlight)
	}
	if len(got) != len(cases) {
		t.Fatalf("got %d verdicts for %d cases", len(got), len(cases))
	}
	for i, v := range got {
		if v.Case.ID != cases[i].ID {
			t.Errorf("verdict %d is for %s, want %s", i, v.Case.ID, cases[i].ID)
		}
	}
}

func TestRunAllWithoutCases(t *testing.T) {
	if got := runAll(nil, 4, func(c Case) Verdict { return Verdict{Case: c} }); len(got) != 0 {
		t.Errorf("no cases should give no verdicts, got %v", got)
	}
}
