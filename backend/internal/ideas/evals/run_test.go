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

	got := runAll(context.Background(), cases, limit, true, one)

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
	if got := runAll(context.Background(), nil, 4, true, func(c Case) Verdict { return Verdict{Case: c} }); len(got) != 0 {
		t.Errorf("no cases should give no verdicts, got %v", got)
	}
}

// TestRunAllStopsWhenTheFirstCaseErrors: a first case that cannot reach
// the model leaves the rest unsent, each with a verdict that says why.
func TestRunAllStopsWhenTheFirstCaseErrors(t *testing.T) {
	cases := []Case{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	sent := 0
	got := runAll(context.Background(), cases, 2, true, func(c Case) Verdict {
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
	got := runAll(ctx, cases, 1, true, func(c Case) Verdict {
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

// TestRunGoesOnAfterAModelFailure (review): only a model that cannot be
// reached stops the run; an answer that failed validation is a result
// the report must show, and the rest are still sent.
func TestRunGoesOnAfterAModelFailure(t *testing.T) {
	cases := []Case{{ID: "a"}, {ID: "b"}}
	sent := 0
	runAll(context.Background(), cases, 1, true, func(c Case) Verdict {
		sent++
		return Verdict{Case: c, Problems: []string{"invalid intent"}}
	})
	if sent != 2 {
		t.Errorf("a failed answer must not stop the run: sent %d", sent)
	}
}

// TestRunAllBlamesTheBudgetNotTheModel (review): when the first case fails
// because the time ran out, the rest say so, not that the model is
// unreachable.
func TestRunAllBlamesTheBudgetNotTheModel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cases := []Case{{ID: "a"}, {ID: "b"}}
	got := runAll(ctx, cases, 2, true, func(c Case) Verdict {
		cancel()
		return Verdict{Case: c, Problems: []string{"context canceled"}, Errored: true}
	})
	if p := got[1].Problems[0]; !strings.Contains(p, "ran out of time") {
		t.Errorf("second case: %q, want the time budget named", p)
	}
}

// TestRunTogetherSendsEveryCaseAtOnce (ADR-0039): for a batching client,
// every case is in flight at once, the first included, so the requests
// gather into one batch.
func TestRunTogetherSendsEveryCaseAtOnce(t *testing.T) {
	cases := []Case{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}}
	var mu sync.Mutex
	inFlight, peak := 0, 0
	release := make(chan struct{})
	go func() {
		for {
			mu.Lock()
			n := inFlight
			mu.Unlock()
			if n == len(cases) {
				close(release)
				return
			}
			time.Sleep(time.Millisecond)
		}
	}()
	got := runAll(context.Background(), cases, len(cases), false, func(c Case) Verdict {
		mu.Lock()
		inFlight++
		peak = max(peak, inFlight)
		mu.Unlock()
		<-release
		return Verdict{Case: c, Pass: true}
	})
	if peak != len(cases) || len(got) != len(cases) || got[0].Case.ID != "a" || got[3].Case.ID != "d" {
		t.Fatalf("want all %d in flight and the verdicts in order: peak %d, %+v", len(cases), peak, got)
	}
}

// TestBatchWanted (#334): EVAL_BATCH=auto batches only a run big enough
// for the saving to matter; 1 and 0 force it; a bad value is an error
// before anything is sent.
func TestBatchWanted(t *testing.T) {
	for _, c := range []struct {
		env, minimum string
		n            int
		want         bool
	}{
		{"", "", 44, false},
		{"0", "", 44, false},
		{"1", "", 3, true},
		{"auto", "", 44, true},
		{"auto", "", 9, false},
		{"auto", "", DefaultBatchMin, true},
		{"auto", "5", 9, true},
	} {
		if got, err := BatchWanted(c.env, c.minimum, c.n); err != nil || got != c.want {
			t.Errorf("EVAL_BATCH=%q EVAL_BATCH_MIN=%q with %d cases: %v %v, want %v", c.env, c.minimum, c.n, got, err, c.want)
		}
	}
	for _, bad := range [][2]string{{"yes", ""}, {"auto", "many"}, {"auto", "0"}} {
		if _, err := BatchWanted(bad[0], bad[1], 10); err == nil {
			t.Errorf("EVAL_BATCH=%q EVAL_BATCH_MIN=%q must be an error", bad[0], bad[1])
		}
	}
}
