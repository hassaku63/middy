package evals

import (
	"context"

	"github.com/hassaku63/middy/backend/internal/ideas"
	"github.com/hassaku63/middy/backend/internal/llm"
)

// Run sends every case as a single-turn conversation and judges the result.
// Provider errors count as failures with the error text as the problem.
func Run(ctx context.Context, svc *ideas.Service, cases []Case) []Verdict {
	out := make([]Verdict, 0, len(cases))
	for _, c := range cases {
		res, err := svc.Imagine(ctx, []ideas.Message{{Role: llm.RoleUser, Content: c.Image}})
		if err != nil {
			out = append(out, Verdict{Case: c, Problems: []string{err.Error()}})
			continue
		}
		out = append(out, Judge(c, res))
	}
	return out
}
