package evals

import (
	"context"
	"fmt"
	"strings"
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

	got := runAll(context.Background(), cases, limit, one)

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
	if got := runAll(context.Background(), nil, 4, func(c Case) Verdict { return Verdict{Case: c} }); len(got) != 0 {
		t.Errorf("no cases should give no verdicts, got %v", got)
	}
}

// TestRunAllStopsWhenTheFirstCaseErrors: a first case that cannot reach
// the model leaves the rest unsent, each with a verdict that says why.
func TestRunAllStopsWhenTheFirstCaseErrors(t *testing.T) {
	cases := []Case{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	sent := 0
	got := runAll(context.Background(), cases, 2, func(c Case) Verdict {
		sent++
		return Verdict{Case: c, Problems: []string{"timeout"}, Errored: true}
	})
	if sent != 1 || len(got) != 3 || !strings.Contains(got[2].Problems[0], "not run") {
		t.Errorf("sent %d, verdicts %+v", sent, got)
	}
}

// TestRunAllReportsWhatTimeLeft: once the context ends, the cases not yet
// started come back as not run, so the report still has every row.
func TestRunAllReportsWhatTimeLeft(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cases := []Case{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}}
	var mu sync.Mutex
	sent := 0
	got := runAll(ctx, cases, 1, func(c Case) Verdict {
		mu.Lock()
		sent++
		mu.Unlock()
		if c.ID == "b" {
			cancel()
		}
		return Verdict{Case: c, Pass: true}
	})
	if sent != 2 || !strings.Contains(got[3].Problems[0], "ran out of time") || !got[1].Pass {
		t.Errorf("sent %d, verdicts %+v", sent, got)
	}
}
