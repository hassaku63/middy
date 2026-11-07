// Package llm defines the boundary between middy and large language models.
//
// Design rule (ADR-0003): the LLM interprets a user's image/mood into
// structured musical parameters and talks with the user. It never emits raw
// MIDI notes. Everything deterministic lives in internal/music.
//
// The interface, provider-neutral errors and a fake live here; provider
// adapters live in subpackages (internal/llm/anthropic).
package llm

import (
	"context"
	"encoding/json"
	"errors"
)

// ErrRefused means the provider's safety layer declined the request.
// Adapters wrap it so callers can tell a refusal from an outage without
// importing the adapter. Safety classifiers are probabilistic, so a
// refusal of a harmless request may not repeat.
var ErrRefused = errors.New("llm: request declined by the provider's safety layer")

// Role identifies the author of a chat message.
type Role string

// Chat roles.
const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message is one turn of a conversation.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// Request describes a single completion call.
type Request struct {
	// System is the operator prompt. Keep it stable across calls so that
	// prompt caching stays effective.
	System string
	// Messages is the conversation so far, oldest first.
	Messages []Message
	// OutputSchema, when non-nil, asks the provider for a JSON object that
	// validates against this JSON Schema (structured output). The response
	// Text then contains that JSON document.
	OutputSchema json.RawMessage
	// MaxTokens caps the response length. Zero means provider default.
	MaxTokens int
}

// Response is the provider's answer.
type Response struct {
	// Text is the assistant's reply, or a JSON document when OutputSchema was set.
	Text string
	// Usage reports token counts when the provider exposes them.
	Usage Usage
}

// Usage is a provider-agnostic token accounting record.
type Usage struct {
	InputTokens  int
	OutputTokens int
}

// Client is implemented by every LLM provider adapter.
type Client interface {
	Complete(ctx context.Context, req Request) (Response, error)
}
