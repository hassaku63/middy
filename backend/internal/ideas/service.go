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
	"github.com/hassaku63/middy/backend/internal/music/pattern"
)

// Sentinel errors let the HTTP layer map failures to status codes.
var (
	// ErrLLMUnavailable wraps transport or provider failures.
	ErrLLMUnavailable = errors.New("ideas: llm unavailable")
	// ErrInvalidOutput means the model did not produce a usable intent even
	// after one retry with the validation errors attached.
	ErrInvalidOutput = errors.New("ideas: llm returned invalid intent")
	// ErrDeclined means the provider's safety layer refused the request
	// twice. Musical images are harmless, so this is almost always a false
	// positive the user can get past by rephrasing.
	ErrDeclined = errors.New("ideas: llm declined the request")
	// ErrInvalidConversation means the messages themselves are not a
	// conversation the service takes (empty, too many, wrong roles).
	ErrInvalidConversation = errors.New("ideas: invalid conversation")
)

// Limits on what we send to the model.
const (
	MaxMessages     = 20
	MaxMessageRunes = 1000
	// Thinking counts toward max_tokens and is always on for the target
	// model, so the ceiling must leave room for it beyond the short JSON.
	maxOutputTokens  = 16000
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
	// Unconstrained is filled only by a diagnostic Service; it never
	// reaches the API.
	Unconstrained string `json:"-"`
	// Retry is why the first answer was not used, when the turn needed the
	// one retry: the validation error, or "refused". Empty otherwise. The
	// evaluation report shows it; it never reaches the API.
	Retry string `json:"-"`
}

// Service turns conversations into candidates.
type Service struct {
	llm llm.Client
	// diagnostics asks the model for intent.DiagnosticSchema's extra
	// free-text field; the evaluation set turns it on (ADR-0013).
	diagnostics bool
}

// New wires the service to an LLM client (real or fake).
func New(client llm.Client) *Service {
	return &Service{llm: client}
}

// NewDiagnostic is New with the evaluation-only "unconstrained" field
// requested from the model. Not for the HTTP server.
func NewDiagnostic(client llm.Client) *Service {
	return &Service{llm: client, diagnostics: true}
}

// Imagine runs one conversational turn. messages must end with a user turn.
func (s *Service) Imagine(ctx context.Context, messages []Message) (Result, error) {
	if err := validateMessages(messages); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrInvalidConversation, err)
	}
	schema := intent.Schema
	if s.diagnostics {
		schema = intent.DiagnosticSchema
	}
	req := llm.Request{
		System:       SystemPrompt(),
		Messages:     toLLM(messages),
		OutputSchema: schema,
		MaxTokens:    maxOutputTokens,
	}

	res, err := s.ask(ctx, req)
	retried := ""
	if err != nil {
		retried = retryReason(err)
		res, err = s.retry(ctx, req, err)
	}
	if err != nil {
		return Result{}, err
	}
	res.Retry = retried + res.Retry
	return res, nil
}

// retryReason names the first failure for Result.Retry.
func retryReason(err error) string {
	var invalid invalidOutputError
	if errors.As(err, &invalid) {
		return invalid.reason
	}
	if errors.Is(err, llm.ErrRefused) {
		return "refused"
	}
	return err.Error()
}

// retry gives the model exactly one more chance after a failure it may get
// past: an unusable answer, which is shown back with the validation errors,
// or a safety refusal, which is resent unchanged because classifiers are
// probabilistic. Any other failure, and any second failure, is final.
func (s *Service) retry(ctx context.Context, req llm.Request, first error) (Result, error) {
	var invalid invalidOutputError
	switch {
	case errors.As(first, &invalid):
		req.Messages = append(append([]llm.Message{}, req.Messages...),
			llm.Message{Role: llm.RoleAssistant, Content: invalid.raw},
			llm.Message{Role: llm.RoleUser, Content: retryInstruction + invalid.reason},
		)
	case errors.Is(first, llm.ErrRefused):
	default:
		return Result{}, first
	}
	res, err := s.ask(ctx, req)
	var contradiction contradictionError
	switch {
	case err == nil:
		return res, nil
	case errors.As(err, &contradiction):
		kept := withoutContradictions(contradiction.res, contradiction.problems)
		kept.Retry = "\nthe retried answer still contradicted its chords; dropped: " + contradiction.reason
		return kept, nil
	case errors.As(err, &invalid):
		return Result{}, fmt.Errorf("%w: %s", ErrInvalidOutput, invalid.reason)
	case errors.Is(err, llm.ErrRefused):
		return Result{}, fmt.Errorf("%w: %v", ErrDeclined, err)
	}
	return Result{}, err
}

// invalidOutputError carries the raw answer and the validation message so
// the retry can quote both.
type invalidOutputError struct {
	raw    string
	reason string
}

func (e invalidOutputError) Error() string { return "invalid intent: " + e.reason }

// ask gets one answer and turns it into a result: parsed, validated,
// rendered and analysed, with the harmony its own words claim checked
// against its chords (claims.go, ADR-0034).
func (s *Service) ask(ctx context.Context, req llm.Request) (Result, error) {
	resp, err := s.llm.Complete(ctx, req)
	if errors.Is(err, llm.ErrRefused) {
		return Result{}, fmt.Errorf("ideas: %w", err)
	}
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrLLMUnavailable, err)
	}
	in, err := intent.Parse([]byte(resp.Text))
	if err != nil {
		return Result{}, invalidOutputError{raw: resp.Text, reason: err.Error()}
	}
	res, err := render(in)
	if err != nil {
		return Result{}, err
	}
	if problems := checkClaims(res); len(problems) > 0 {
		reasons := make([]string, len(problems))
		for i, p := range problems {
			reasons[i] = p.String()
		}
		return Result{}, contradictionError{
			invalidOutputError: invalidOutputError{raw: resp.Text, reason: strings.Join(reasons, "\n")},
			res:                res,
			problems:           problems,
		}
	}
	return res, nil
}

// contradictionError is an answer whose music is sound but whose words
// claim harmony its chords contradict. A retry is asked for as for any
// invalid answer; if the retried answer still contradicts itself, the
// contradicted sentences are dropped instead of failing the turn.
type contradictionError struct {
	invalidOutputError
	res      Result
	problems []claimProblem
}

// Unwrap lets the retry treat it as the invalid answer it is.
func (e contradictionError) Unwrap() error { return e.invalidOutputError }

// render turns every candidate into an analysed timeline. A candidate the
// engine cannot render is a bug in intent.Validate, not user error.
func render(in intent.Intent) (Result, error) {
	out := Result{Reply: in.Summary, Mood: in.Mood, Unconstrained: in.Unconstrained}
	if out.Mood == nil {
		out.Mood = []string{}
	}
	for i, c := range in.Candidates {
		prog, err := c.ToProgression()
		if err != nil {
			return Result{}, fmt.Errorf("ideas: candidates[%d]: %w", i, err)
		}
		tl, err := music.FromProgressionIn(c.TimeSignature(), c.Key, prog, c.TempoBPM)
		if err != nil {
			return Result{}, fmt.Errorf("ideas: candidates[%d]: %w", i, err)
		}
		tl, err = applyPatterns(tl, c)
		if err != nil {
			return Result{}, fmt.Errorf("ideas: candidates[%d]: %w", i, err)
		}
		tl, err = music.Analyze(tl)
		if err != nil {
			return Result{}, fmt.Errorf("ideas: candidates[%d]: %w", i, err)
		}
		if tl.Rhythm == nil || tl.Rhythm.Feel == nil {
			c.Feel = "" // a swing with no pattern to act on is not heard; do not show it
		}
		out.Candidates = append(out.Candidates, Candidate{Candidate: c, Timeline: tl})
	}
	return out, nil
}

// applyPatterns layers the candidate's patterns and feel, if any, on the
// block chords.
func applyPatterns(tl music.Timeline, c intent.Candidate) (music.Timeline, error) {
	return pattern.Render(tl, pattern.Choice{Accompaniment: c.Accompaniment, Drums: c.Drums, Grid: c.Grid(), Feel: pattern.Feel(c.Feel), Style: c.Style, Tensions: c.Tensions})
}

// validateMessages says what is wrong with a conversation. Imagine wraps
// it in ErrInvalidConversation, which carries the package prefix, so the
// messages here have none (#228).
func validateMessages(messages []Message) error {
	if len(messages) == 0 {
		return errors.New("messages must not be empty")
	}
	if len(messages) > MaxMessages {
		return fmt.Errorf("at most %d messages, got %d", MaxMessages, len(messages))
	}
	for i, m := range messages {
		if m.Role != llm.RoleUser && m.Role != llm.RoleAssistant {
			return fmt.Errorf("messages[%d]: role must be user or assistant", i)
		}
		if strings.TrimSpace(m.Content) == "" {
			return fmt.Errorf("messages[%d]: content is empty", i)
		}
		if n := len([]rune(m.Content)); n > MaxMessageRunes {
			return fmt.Errorf("messages[%d]: content is too long (%d > %d characters)", i, n, MaxMessageRunes)
		}
	}
	if messages[len(messages)-1].Role != llm.RoleUser {
		return errors.New("the last message must be from the user")
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
