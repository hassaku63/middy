// Package ideas is the application service behind scenario S2: a user's
// image, refined in conversation, becomes two or three playable, analysed
// candidate timelines. The LLM only produces intent.Intent (ADR-0003);
// everything musical is rendered by internal/music.
package ideas

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hassaku63/middy/backend/internal/intent"
	"github.com/hassaku63/middy/backend/internal/llm"
	"github.com/hassaku63/middy/backend/internal/music"
)

// Sentinel errors let the HTTP layer map failures to status codes.
var (
	// ErrLLMUnavailable wraps transport or provider failures.
	ErrLLMUnavailable = errors.New("ideas: llm unavailable")
	// ErrInvalidOutput means the model did not produce a usable intent even
	// after one retry with the validation errors attached.
	ErrInvalidOutput = errors.New("ideas: llm returned invalid intent")
)

// Limits on what we send to the model.
const (
	MaxMessages      = 20
	MaxMessageRunes  = 1000
	maxOutputTokens  = 2048
	retryInstruction = "前回の答えは次の理由で使えませんでした。同じ JSON Schema で、問題を直して答え直してください:\n"
)

// Message is one turn of the conversation as the client keeps it: the
// user's texts and, for assistant turns, the Summary of the earlier answer.
type Message struct {
	Role    llm.Role `json:"role"`
	Content string   `json:"content"`
}

// Candidate pairs the model's proposal with its rendered, analysed timeline.
type Candidate struct {
	intent.Candidate
	Timeline music.Timeline `json:"timeline"`
}

// Result is what the user sees for one turn.
type Result struct {
	Reply      string      `json:"reply"`
	Mood       []string    `json:"mood"`
	Candidates []Candidate `json:"candidates"`
}

// Service turns conversations into candidates.
type Service struct {
	llm llm.Client
}

// New wires the service to an LLM client (real or fake).
func New(client llm.Client) *Service {
	return &Service{llm: client}
}

// Imagine runs one conversational turn. messages must end with a user turn.
func (s *Service) Imagine(ctx context.Context, messages []Message) (Result, error) {
	if err := validateMessages(messages); err != nil {
		return Result{}, err
	}
	req := llm.Request{
		System:       systemPrompt,
		Messages:     toLLM(messages),
		OutputSchema: intent.Schema,
		MaxTokens:    maxOutputTokens,
	}

	in, err := s.ask(ctx, req)
	if err != nil {
		var invalid invalidOutputError
		if !errors.As(err, &invalid) {
			return Result{}, err
		}
		// One retry: show the model its own answer and what was wrong with it.
		retry := req
		retry.Messages = append(append([]llm.Message{}, req.Messages...),
			llm.Message{Role: llm.RoleAssistant, Content: invalid.raw},
			llm.Message{Role: llm.RoleUser, Content: retryInstruction + invalid.reason},
		)
		in, err = s.ask(ctx, retry)
		if err != nil {
			if errors.As(err, &invalid) {
				return Result{}, fmt.Errorf("%w: %s", ErrInvalidOutput, invalid.reason)
			}
			return Result{}, err
		}
	}
	return render(in)
}

// invalidOutputError carries the raw answer and the validation message so
// the retry can quote both.
type invalidOutputError struct {
	raw    string
	reason string
}

func (e invalidOutputError) Error() string { return "invalid intent: " + e.reason }

func (s *Service) ask(ctx context.Context, req llm.Request) (intent.Intent, error) {
	resp, err := s.llm.Complete(ctx, req)
	if err != nil {
		return intent.Intent{}, fmt.Errorf("%w: %v", ErrLLMUnavailable, err)
	}
	in, err := intent.Parse([]byte(resp.Text))
	if err != nil {
		return intent.Intent{}, invalidOutputError{raw: resp.Text, reason: err.Error()}
	}
	return in, nil
}

// render turns every candidate into an analysed timeline. A candidate the
// engine cannot render is a bug in intent.Validate, not user error.
func render(in intent.Intent) (Result, error) {
	out := Result{Reply: in.Summary, Mood: in.Mood}
	if out.Mood == nil {
		out.Mood = []string{}
	}
	for i, c := range in.Candidates {
		prog, err := c.ToProgression()
		if err != nil {
			return Result{}, fmt.Errorf("ideas: candidates[%d]: %w", i, err)
		}
		tl, err := music.FromProgression(c.Key, prog, c.TempoBPM)
		if err != nil {
			return Result{}, fmt.Errorf("ideas: candidates[%d]: %w", i, err)
		}
		tl, err = music.Analyze(tl)
		if err != nil {
			return Result{}, fmt.Errorf("ideas: candidates[%d]: %w", i, err)
		}
		out.Candidates = append(out.Candidates, Candidate{Candidate: c, Timeline: tl})
	}
	return out, nil
}

func validateMessages(messages []Message) error {
	if len(messages) == 0 {
		return errors.New("ideas: messages must not be empty")
	}
	if len(messages) > MaxMessages {
		return fmt.Errorf("ideas: at most %d messages, got %d", MaxMessages, len(messages))
	}
	for i, m := range messages {
		if m.Role != llm.RoleUser && m.Role != llm.RoleAssistant {
			return fmt.Errorf("ideas: messages[%d]: role must be user or assistant", i)
		}
		if strings.TrimSpace(m.Content) == "" {
			return fmt.Errorf("ideas: messages[%d]: content is empty", i)
		}
		if n := len([]rune(m.Content)); n > MaxMessageRunes {
			return fmt.Errorf("ideas: messages[%d]: content is too long (%d > %d characters)", i, n, MaxMessageRunes)
		}
	}
	if messages[len(messages)-1].Role != llm.RoleUser {
		return errors.New("ideas: the last message must be from the user")
	}
	return nil
}

func toLLM(messages []Message) []llm.Message {
	out := make([]llm.Message, len(messages))
	for i, m := range messages {
		out[i] = llm.Message{Role: m.Role, Content: m.Content}
	}
	return out
}
