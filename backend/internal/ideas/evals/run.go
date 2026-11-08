package evals

import (
	"context"
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
func Run(ctx context.Context, svc *ideas.Service, cases []Case) []Verdict {
	return runAll(cases, Concurrency, func(c Case) Verdict {
		res, err := svc.Imagine(ctx, []ideas.Message{{Role: llm.RoleUser, Content: c.Image}})
		if err != nil {
			return Verdict{Case: c, Problems: []string{err.Error()}}
		}
		return Judge(c, res)
	})
}

// runAll runs the first case, then the rest with at most limit in flight,
// and returns the verdicts in the order of the cases.
func runAll(cases []Case, limit int, one func(Case) Verdict) []Verdict {
	out := make([]Verdict, len(cases))
	if len(cases) == 0 {
		return out
	}
	out[0] = one(cases[0])

	slots := make(chan struct{}, limit)
	var wg sync.WaitGroup
	for i := 1; i < len(cases); i++ {
		slots <- struct{}{}
		wg.Go(func() {
			defer func() { <-slots }()
			out[i] = one(cases[i])
		})
	}
	wg.Wait()
	return out
}
