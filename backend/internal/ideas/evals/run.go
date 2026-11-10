package evals

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/hassaku63/middy/backend/internal/ideas"
	"github.com/hassaku63/middy/backend/internal/llm"
)

// Concurrency is how many cases are in flight at once after the first. Low
// enough to stay well inside the request-per-minute limit of a small API
// tier; the SDK retries a 429 if it is reached anyway.
const Concurrency = 4

// Run sends every case as a single-turn conversation and judges the result.
// Provider errors count as failures with the error text as the problem.
//
// The first case runs alone: the system prompt is cached, and a cache entry
// can be read only once the response that writes it has started. Sending
// every case at once would write the entry for each of them; sending the
// first alone lets the rest read it.
//
// When the first case cannot reach the model, the rest are not sent: they
// would fail the same way, after the same retries. When ctx ends, the cases
// not yet started are not sent either, and every verdict still comes back,
// so the report shows how far the run got instead of nothing.
func Run(ctx context.Context, svc *ideas.Service, cases []Case) []Verdict {
	return runAll(ctx, cases, Concurrency, true, imagine(ctx, svc))
}

// RunTogether sends every case at once, for a client that batches the
// requests made together (anthropic.Batcher, ADR-0039): sending the first
// alone, or a few at a time, would make batches of one or a few.
func RunTogether(ctx context.Context, svc *ideas.Service, cases []Case) []Verdict {
	return runAll(ctx, cases, len(cases), false, imagine(ctx, svc))
}

// imagine sends one case as a single-turn conversation and judges it.
// Provider errors count as failures with the error text as the problem.
func imagine(ctx context.Context, svc *ideas.Service) func(Case) Verdict {
	return func(c Case) Verdict {
		res, err := svc.Imagine(ctx, []ideas.Message{{Role: llm.RoleUser, Content: c.Image}})
		if err != nil {
			// Only a model that could not be reached stops the run; a
			// refused or invalid answer is a result to report.
			unreachable := errors.Is(err, ideas.ErrLLMUnavailable) || ctx.Err() != nil
			return Verdict{Case: c, Problems: []string{err.Error()}, Errored: unreachable}
		}
		return Judge(c, res)
	}
}

// runAll runs the first case (alone when firstAlone), then the rest with at
// most limit in flight, and returns the verdicts in the order of the cases.
// A first case run alone that errored, or an ended ctx, leaves the cases
// not yet started unsent.
func runAll(ctx context.Context, cases []Case, limit int, firstAlone bool, one func(Case) Verdict) []Verdict {
	out := make([]Verdict, len(cases))
	if len(cases) == 0 {
		return out
	}
	start := 0
	if firstAlone {
		start = 1
		out[0] = one(cases[0])
	}
	if firstAlone && out[0].Errored {
		why := "the first case could not reach the model; the rest would fail the same way"
		if ctx.Err() != nil { // the budget ran out, which says nothing about the model
			why = "the evaluation ran out of time during the first case"
		}
		for i := 1; i < len(cases); i++ {
			out[i] = notRun(cases[i], why)
		}
		return out
	}

	slots := make(chan struct{}, max(limit, 1)) // 0 would block every case after the first
	var wg sync.WaitGroup
	for i := start; i < len(cases); i++ {
		slots <- struct{}{}
		if ctx.Err() != nil {
			<-slots
			out[i] = notRun(cases[i], "the evaluation ran out of time")
			continue
		}
		wg.Go(func() {
			defer func() { <-slots }()
			out[i] = one(cases[i])
		})
	}
	wg.Wait()
	return out
}

func notRun(c Case, why string) Verdict {
	return Verdict{Case: c, Problems: []string{"not run: " + why}, Errored: true}
}

// DefaultBatchMin is how many cases a run needs before EVAL_BATCH=auto
// sends it as Message Batches: a handful costs little either way and is
// quicker one request at a time (#334, ADR-0039).
const DefaultBatchMin = 20

// BatchWanted reads EVAL_BATCH for a run of n cases: "1" batches, "" or
// "0" does not, and "auto" batches when n reaches minimum (EVAL_BATCH_MIN,
// DefaultBatchMin when empty).
func BatchWanted(env, minimum string, n int) (bool, error) {
	switch env {
	case "", "0":
		return false, nil
	case "1":
		return true, nil
	case "auto":
		least := DefaultBatchMin
		if minimum != "" {
			v, err := strconv.Atoi(minimum)
			if err != nil || v < 1 {
				return false, fmt.Errorf("evals: EVAL_BATCH_MIN must be a positive number, got %q", minimum)
			}
			least = v
		}
		return n >= least, nil
	}
	return false, fmt.Errorf("evals: EVAL_BATCH must be 1, 0 or auto, got %q", env)
}
