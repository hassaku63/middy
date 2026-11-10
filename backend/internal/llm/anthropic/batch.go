package anthropic

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	sdk "github.com/anthropics/anthropic-sdk-go/v2"

	"github.com/hassaku63/middy/backend/internal/llm"
)

// ErrBatch means a request sent in a Message Batch came back without an
// answer: it errored, or the batch was canceled or expired.
var ErrBatch = errors.New("anthropic: batch request failed")

// Batcher is an llm.Client that sends the requests made at about the same
// time as one Message Batch, at half the price of single requests, for
// work that can wait for the answers, such as a full evaluation (#334,
// ADR-0039). A request waits until no other has come for the gather time,
// then the gathered requests go as one batch; the batch is polled until it
// ends and each caller gets its own answer. A request made while a batch
// is out (a retry after a first answer) goes in the next batch.
//
// The callers must make their requests concurrently: one caller at a time
// would send batches of one.
type Batcher struct {
	c      *Client
	ctx    context.Context
	gather time.Duration
	poll   time.Duration

	mu      sync.Mutex
	pending []*batchCall
	timer   *time.Timer
}

type batchCall struct {
	params sdk.MessageNewParams
	done   chan batchResult
}

type batchResult struct {
	resp llm.Response
	err  error
}

// NewBatcher sends c's requests in batches. ctx bounds the batches, gather
// is how long to wait for more requests and poll how often to ask whether
// a batch has ended.
func NewBatcher(ctx context.Context, c *Client, gather, poll time.Duration) *Batcher {
	return &Batcher{c: c, ctx: ctx, gather: gather, poll: poll}
}

// Complete puts the request in the next batch and waits for its answer.
func (b *Batcher) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	params, err := b.c.params(req)
	if err != nil {
		return llm.Response{}, err
	}
	call := &batchCall{params: params, done: make(chan batchResult, 1)}
	b.mu.Lock()
	b.pending = append(b.pending, call)
	if b.timer == nil {
		b.timer = time.AfterFunc(b.gather, b.flush)
	} else {
		b.timer.Reset(b.gather)
	}
	b.mu.Unlock()
	select {
	case r := <-call.done:
		return r.resp, r.err
	case <-ctx.Done():
		return llm.Response{}, ctx.Err()
	}
}

// flush sends the gathered requests as one batch and hands out the answers.
func (b *Batcher) flush() {
	b.mu.Lock()
	calls := b.pending
	b.pending, b.timer = nil, nil
	b.mu.Unlock()
	if len(calls) == 0 {
		return
	}
	results, err := b.send(calls)
	for i, call := range calls {
		r, ok := results[i]
		switch {
		case err != nil:
			r = batchResult{err: err}
		case !ok:
			r = batchResult{err: fmt.Errorf("%w: no result for request %d", ErrBatch, i)}
		}
		call.done <- r
	}
}

// send runs one batch to its end and reads its results by request index.
func (b *Batcher) send(calls []*batchCall) (map[int]batchResult, error) {
	reqs := make([]sdk.MessageBatchNewParamsRequest, len(calls))
	for i, call := range calls {
		p := call.params
		// A batch can take longer than the cache's 5 minutes, so the system
		// prompt is cached for an hour: the docs advise it for batches
		// sharing context, where hits are best-effort (ADR-0039).
		system := make([]sdk.TextBlockParam, len(p.System))
		for j, block := range p.System {
			block.CacheControl = sdk.CacheControlEphemeralParam{TTL: sdk.CacheControlEphemeralTTLTTL1h}
			system[j] = block
		}
		reqs[i] = sdk.MessageBatchNewParamsRequest{
			CustomID: strconv.Itoa(i),
			Params: sdk.MessageBatchNewParamsRequestParams{
				Model:        p.Model,
				MaxTokens:    p.MaxTokens,
				Messages:     p.Messages,
				System:       system,
				OutputConfig: p.OutputConfig,
			},
		}
	}
	batch, err := b.c.api.Messages.Batches.New(b.ctx, sdk.MessageBatchNewParams{Requests: reqs})
	if err != nil {
		return nil, classify(err)
	}
	for batch.ProcessingStatus != sdk.MessageBatchProcessingStatusEnded {
		select {
		case <-b.ctx.Done():
			return nil, b.ctx.Err()
		case <-time.After(b.poll):
		}
		if batch, err = b.c.api.Messages.Batches.Get(b.ctx, batch.ID, sdk.MessageBatchGetParams{}); err != nil {
			return nil, classify(err)
		}
	}
	stream := b.c.api.Messages.Batches.ResultsStreaming(b.ctx, batch.ID, sdk.MessageBatchResultsParams{})
	defer func() { _ = stream.Close() }()
	out := map[int]batchResult{}
	for stream.Next() {
		r := stream.Current()
		i, err := strconv.Atoi(r.CustomID)
		if err != nil || i < 0 || i >= len(calls) {
			continue // not one of ours
		}
		switch r.Result.Type {
		case "succeeded":
			msg := r.Result.Message
			resp, err := b.c.response(b.ctx, &msg)
			out[i] = batchResult{resp: resp, err: err}
		case "errored":
			out[i] = batchResult{err: fmt.Errorf("%w: %s: %s", ErrBatch, r.Result.Error.Error.Type, r.Result.Error.Error.Message)}
		default: // canceled, expired
			out[i] = batchResult{err: fmt.Errorf("%w: %s", ErrBatch, r.Result.Type)}
		}
	}
	if err := stream.Err(); err != nil {
		return nil, classify(err)
	}
	return out, nil
}
