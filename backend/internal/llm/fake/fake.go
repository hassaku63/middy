// Package fake is an in-memory llm.Client for tests and local development
// without API keys. It replays scripted responses in order and records every
// request so tests can assert on prompts.
package fake

import (
	"context"
	"errors"
	"sync"

	"github.com/hassaku63/middy/backend/internal/llm"
)

// ErrNoScriptedResponse is returned when the fake runs out of responses.
var ErrNoScriptedResponse = errors.New("fake llm: no scripted response left")

// Client is a scripted llm.Client.
type Client struct {
	mu        sync.Mutex
	responses []llm.Response
	requests  []llm.Request
}

var _ llm.Client = (*Client)(nil)

// New returns a Client that will answer with the given responses in order.
// With no responses it answers every request with a fixed placeholder text,
// which is convenient for running the server locally.
func New(responses ...llm.Response) *Client {
	return &Client{responses: responses}
}

// Complete returns the next scripted response, or the placeholder when the
// client was created without any script.
func (c *Client) Complete(_ context.Context, req llm.Request) (llm.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.requests = append(c.requests, req)

	if c.responses == nil {
		return llm.Response{Text: "(fake llm) " + lastUserContent(req)}, nil
	}
	if len(c.responses) == 0 {
		return llm.Response{}, ErrNoScriptedResponse
	}
	resp := c.responses[0]
	c.responses = c.responses[1:]
	return resp, nil
}

// Requests returns a copy of every request received so far.
func (c *Client) Requests() []llm.Request {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]llm.Request, len(c.requests))
	copy(out, c.requests)
	return out
}

func lastUserContent(req llm.Request) string {
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == llm.RoleUser {
			return req.Messages[i].Content
		}
	}
	return ""
}
