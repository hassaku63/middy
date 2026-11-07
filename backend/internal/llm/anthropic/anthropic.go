// Package anthropic is the llm.Client adapter for the Claude Messages API
// via the official SDK. It is the only package that knows about the
// provider; everything else talks to llm.Client.
//
// Implementation notes live in docs/architecture/llm-integration.md.
package anthropic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	sdk "github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/anthropics/anthropic-sdk-go/v2/option"

	"github.com/hassaku63/middy/backend/internal/llm"
)

// Defaults when Config leaves a field empty.
const (
	DefaultModel     = "claude-opus-5-5"
	DefaultTimeout   = 30 * time.Second
	DefaultMaxTokens = 16000
)

// Sentinel errors callers can branch on.
var (
	// ErrRefused means the provider's safety layer declined the request.
	ErrRefused = errors.New("anthropic: request refused by the model")
	// ErrTruncated means the answer hit MaxTokens; for structured output the
	// JSON is then unusable, so it is an error rather than a partial result.
	ErrTruncated = errors.New("anthropic: answer truncated at max tokens")
)

// Config configures the adapter. Only APIKey is required.
type Config struct {
	APIKey string
	// Model is the model id, e.g. "claude-opus-5-5".
	Model string
	// Effort steers thinking depth; "low" is plenty for intent extraction.
	Effort string
	// Timeout bounds one request including the SDK's retries.
	Timeout time.Duration
	// BaseURL overrides the API endpoint (tests point it at httptest).
	BaseURL string
	// MaxRetries overrides the SDK default (2); tests set 0.
	MaxRetries *int
}

// Client calls the Messages API.
type Client struct {
	api    sdk.Client
	model  string
	effort sdk.OutputConfigEffort
}

var _ llm.Client = (*Client)(nil)

// New builds a Client from Config.
func New(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, errors.New("anthropic: API key is required")
	}
	if cfg.Model == "" {
		cfg.Model = DefaultModel
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultTimeout
	}
	opts := []option.RequestOption{
		option.WithAPIKey(cfg.APIKey),
		option.WithRequestTimeout(cfg.Timeout),
	}
	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}
	if cfg.MaxRetries != nil {
		opts = append(opts, option.WithMaxRetries(*cfg.MaxRetries))
	}
	effort := sdk.OutputConfigEffortLow
	if cfg.Effort != "" {
		effort = sdk.OutputConfigEffort(cfg.Effort)
	}
	return &Client{api: sdk.NewClient(opts...), model: cfg.Model, effort: effort}, nil
}

// Complete sends one request. When req.OutputSchema is set the model is
// constrained to JSON matching it (structured output) and Response.Text is
// that JSON document.
func (c *Client) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	params, err := c.params(req)
	if err != nil {
		return llm.Response{}, err
	}
	msg, err := c.api.Messages.New(ctx, params)
	if err != nil {
		return llm.Response{}, classify(err)
	}
	switch msg.StopReason {
	case sdk.StopReasonRefusal:
		return llm.Response{}, fmt.Errorf("%w (%s)", ErrRefused, msg.StopDetails.Category)
	case sdk.StopReasonMaxTokens:
		return llm.Response{}, ErrTruncated
	}
	// Token accounting per call: the prerequisite for judging cost and for
	// seeing whether the cached system prompt is actually read from cache.
	slog.InfoContext(ctx, "llm usage",
		"model", c.model,
		"input_tokens", msg.Usage.InputTokens,
		"output_tokens", msg.Usage.OutputTokens,
		"cache_read_input_tokens", msg.Usage.CacheReadInputTokens,
		"cache_creation_input_tokens", msg.Usage.CacheCreationInputTokens,
	)
	var text strings.Builder
	for _, block := range msg.Content {
		if tb, ok := block.AsAny().(sdk.TextBlock); ok {
			text.WriteString(tb.Text)
		}
	}
	return llm.Response{
		Text: text.String(),
		Usage: llm.Usage{
			InputTokens:  int(msg.Usage.InputTokens),
			OutputTokens: int(msg.Usage.OutputTokens),
		},
	}, nil
}

func (c *Client) params(req llm.Request) (sdk.MessageNewParams, error) {
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = DefaultMaxTokens
	}
	p := sdk.MessageNewParams{
		Model:        c.model,
		MaxTokens:    int64(maxTokens),
		OutputConfig: sdk.OutputConfigParam{Effort: c.effort},
	}
	if req.System != "" {
		// One stable system block with a cache breakpoint: the prompt is
		// constant across calls, so later requests read it from cache.
		p.System = []sdk.TextBlockParam{{
			Text:         req.System,
			CacheControl: sdk.NewCacheControlEphemeralParam(),
		}}
	}
	for i, m := range req.Messages {
		switch m.Role {
		case llm.RoleUser:
			p.Messages = append(p.Messages, sdk.NewUserMessage(sdk.NewTextBlock(m.Content)))
		case llm.RoleAssistant:
			p.Messages = append(p.Messages, sdk.NewAssistantMessage(sdk.NewTextBlock(m.Content)))
		default:
			return sdk.MessageNewParams{}, fmt.Errorf("anthropic: messages[%d]: unsupported role %q", i, m.Role)
		}
	}
	if req.OutputSchema != nil {
		var schema map[string]any
		if err := json.Unmarshal(req.OutputSchema, &schema); err != nil {
			return sdk.MessageNewParams{}, fmt.Errorf("anthropic: output schema is not a JSON object: %w", err)
		}
		p.OutputConfig.Format = sdk.JSONOutputFormatParam{Schema: schema}
	}
	return p, nil
}

// classify keeps provider status codes visible in the error chain without
// leaking SDK types to callers.
func classify(err error) error {
	var apiErr *sdk.Error
	if errors.As(err, &apiErr) {
		return fmt.Errorf("anthropic: api status %d (request %s): %w", apiErr.StatusCode, apiErr.RequestID, err)
	}
	return fmt.Errorf("anthropic: transport: %w", err)
}
